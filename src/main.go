package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gomarkdown/markdown"
	_ "github.com/mattn/go-sqlite3"
)

// Global database connection
var db *sql.DB

// Template cache for performance
var (
	templateCache = make(map[string]*template.Template)
	templateMutex sync.RWMutex
)

// Cache for parsed markdown files
var (
	markdownCache = make(map[string]StockData)
	cacheMutex    sync.RWMutex
	cacheExpiry   = make(map[string]time.Time)
)

// Disable debug logging for production performance
const DEBUG_MODE = true // Enable debug logging to troubleshoot

func debugLog(format string, args ...interface{}) {
	if DEBUG_MODE {
		log.Printf("DEBUG: "+format, args...)
	}
}

const (
	GEMINI_API_KEY     = "AIzaSyBkw_fi16Q39yjZdZ0C3PTw-vuADTR-KAM"        // Your working API key
	TELEGRAM_BOT_TOKEN = "7912088515:AAFn3YbnE-84MmMgvhoc6vpJ5HiLPtH5IEg" // Your working bot token
	TELEGRAM_CHAT_ID   = "5743904087"                                     // FIX: Use the actual chat ID from debug logs
)

type StockData struct {
	Title                    string // Add this field
	CurrentDate              string
	MorningOpenIndex         float64
	MorningOpenChange        float64
	MorningOpenHighlights    string
	MorningOpenAnalysis      template.HTML
	MorningCloseIndex        float64
	MorningCloseChange       float64
	MorningCloseHighlights   string
	MorningCloseSummary      template.HTML
	AfternoonOpenIndex       float64
	AfternoonOpenChange      float64
	AfternoonOpenHighlights  string
	AfternoonOpenAnalysis    template.HTML
	AfternoonCloseIndex      float64
	AfternoonCloseChange     float64
	AfternoonCloseHighlights string
	AfternoonCloseSummary    template.HTML
	KeyTakeaways             []string
}

type ArticlePreview struct {
	Title        string
	Date         string
	SetIndex     string
	Change       float64
	ShortSummary string
	Summary      string
	Slug         string
	URL          string
}

type DBArticle struct {
	ID        int
	Slug      string
	Title     string
	Summary   sql.NullString
	Content   sql.NullString
	CreatedAt string
}

type IndexPageData struct {
	CurrentDate string
	Articles    []ArticlePreview
}

type AdminDashboardData struct {
	Articles    []DBArticle
	CurrentDate string
	Success     string
	Error       string
}

type AdminArticleFormData struct {
	IsEdit                   bool
	Action                   string
	Article                  DBArticle
	Error                    string
	CurrentDate              string
	MorningOpenIndex         string
	MorningOpenChange        float64
	MorningOpenHighlights    string
	MorningOpenAnalysis      string
	MorningCloseIndex        string
	MorningCloseChange       float64
	MorningCloseHighlights   string
	MorningCloseSummary      string
	AfternoonOpenIndex       string
	AfternoonOpenChange      float64
	AfternoonOpenHighlights  string
	AfternoonOpenAnalysis    string
	AfternoonCloseIndex      string
	AfternoonCloseChange     float64
	AfternoonCloseHighlights string
	AfternoonCloseSummary    string
	KeyTakeaways             string
}

// API request structures for market data
type MarketDataRequest struct {
	Date        string `json:"date"` // YYYY-MM-DD format
	MorningOpen struct {
		Index      float64 `json:"index"`
		Change     float64 `json:"change"`
		Highlights string  `json:"highlights"`
	} `json:"morning_open"`
	MorningClose struct {
		Index  float64 `json:"index"`
		Change float64 `json:"change"`
	} `json:"morning_close"`
	AfternoonOpen struct {
		Index      float64 `json:"index"`
		Change     float64 `json:"change"`
		Highlights string  `json:"highlights"`
	} `json:"afternoon_open"`
	AfternoonClose struct {
		Index  float64 `json:"index"`
		Change float64 `json:"change"`
	} `json:"afternoon_close"`
}

// API response structure
type MarketDataResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		ArticleID int    `json:"article_id,omitempty"`
		Slug      string `json:"slug,omitempty"`
		URL       string `json:"url,omitempty"`
	} `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

// Gemini API request structure
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

// Gemini API response structure
type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
	Error      *GeminiError      `json:"error,omitempty"`
}

type GeminiCandidate struct {
	Content GeminiContent `json:"content"`
}

type GeminiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Telegram API structures
type TelegramMessage struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

type TelegramResponse struct {
	Ok     bool `json:"ok"`
	Result struct {
		MessageID int `json:"message_id"`
	} `json:"result"`
	ErrorCode   int    `json:"error_code,omitempty"`
	Description string `json:"description,omitempty"`
}

// Fast template cache with mutex protection
func getTemplate(name string, files ...string) (*template.Template, error) {
	templateMutex.RLock()
	if tmpl, exists := templateCache[name]; exists {
		templateMutex.RUnlock()
		return tmpl, nil
	}
	templateMutex.RUnlock()

	// Double-check locking for performance
	templateMutex.Lock()
	defer templateMutex.Unlock()

	if tmpl, exists := templateCache[name]; exists {
		return tmpl, nil
	}

	// Custom template functions
	funcMap := template.FuncMap{
		"printf": fmt.Sprintf,
		"html":   func(s string) template.HTML { return template.HTML(s) },
		"add":    func(a, b int) int { return a + b },
		"markdownToHTML": func(md string) template.HTML {
			html := markdown.ToHTML([]byte(md), nil, nil)
			return template.HTML(html)
		},
	}

	// For admin templates, use the provided files
	if len(files) > 0 {
		tmpl, err := template.New("").Funcs(funcMap).ParseFiles(files...)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template files: %v", err)
		}
		templateCache[name] = tmpl
		return tmpl, nil
	}

	// Default behavior for page templates
	templateFile := fmt.Sprintf("src/templates/%s.gohtml", name)
	if _, err := os.Stat(templateFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("template file not found: %s", templateFile)
	}

	baseTemplate, err := template.New("").Funcs(funcMap).ParseFiles("src/templates/base.gohtml")
	if err != nil {
		return nil, fmt.Errorf("failed to parse base template: %v", err)
	}

	tmpl, err := baseTemplate.ParseFiles(templateFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %v", err)
	}

	templateCache[name] = tmpl
	return tmpl, nil
}

// Fast cached markdown parsing
func getCachedStockData(filePath string) (StockData, error) {
	cacheMutex.RLock()
	if data, exists := markdownCache[filePath]; exists {
		if expiry, hasExpiry := cacheExpiry[filePath]; hasExpiry {
			if time.Now().Before(expiry) {
				cacheMutex.RUnlock()
				return data, nil
			}
		}
	}
	cacheMutex.RUnlock()

	// Parse and cache
	data, err := parseMarkdownArticle(filePath)
	if err != nil {
		return data, err
	}

	cacheMutex.Lock()
	markdownCache[filePath] = data
	cacheExpiry[filePath] = time.Now().Add(5 * time.Minute) // Cache for 5 minutes
	cacheMutex.Unlock()

	return data, nil
}

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "src/admin.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	createTableSQL := `
    CREATE TABLE IF NOT EXISTS articles (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        slug TEXT NOT NULL UNIQUE,
        title TEXT NOT NULL,
        summary TEXT,
        content TEXT,
        created_at TEXT NOT NULL
    );`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Failed to create articles table: %v", err)
	}

	var columnName string
	err = db.QueryRow("SELECT name FROM PRAGMA_TABLE_INFO('articles') WHERE name='content'").Scan(&columnName)
	if err == sql.ErrNoRows {
		_, err = db.Exec("ALTER TABLE articles ADD COLUMN content TEXT")
		if err != nil {
			log.Fatalf("Failed to add 'content' column: %v", err)
		}
	} else if err != nil {
		log.Fatalf("Failed to check for 'content' column: %v", err)
	}

	seedArticlesTable()
}

func seedArticlesTable() {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM articles").Scan(&count)
	if err != nil {
		log.Printf("Error checking article count: %v", err)
		return
	}

	if count == 0 {
		stmt, err := db.Prepare("INSERT INTO articles(slug, title, summary, content, created_at) VALUES(?, ?, ?, ?, ?)")
		if err != nil {
			log.Printf("Error preparing statement: %v", err)
			return
		}
		defer stmt.Close()

		articles := [][]string{
			{"2025-09-19", "Stock Market Analysis - 19 September 2025", "SET index closed higher with banking and energy sectors leading gains", "", "2025-09-19"},
			{"2025-09-18", "Stock Market Analysis - 18 September 2025", "Market declined amid regional weakness", "", "2025-09-18"},
			{"2025-09-22", "Stock Market Analysis - 22 September 2025", "Morning session opened with modest gains", "", "2025-09-22"},
		}

		for _, article := range articles {
			stmt.Exec(article[0], article[1], article[2], article[3], article[4])
		}
	}
}

// Optimized markdown parser - removed excessive logging
func parseMarkdownArticle(filePath string) (StockData, error) {
	data := StockData{
		CurrentDate:  time.Now().Format("2 January 2006"),
		KeyTakeaways: []string{},
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return data, err
	}

	lines := strings.Split(string(content), "\n")
	currentSection := ""
	currentSubsection := ""
	analysisContent := ""
	summaryContent := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Main sections
		if strings.HasPrefix(line, "## Morning Session") {
			currentSection = "morning"
			currentSubsection = ""
			continue
		} else if strings.HasPrefix(line, "## Afternoon Session") {
			currentSection = "afternoon"
			currentSubsection = ""
			continue
		}

		// Subsections
		if strings.HasPrefix(line, "### Open Set") {
			currentSubsection = "open"
			analysisContent = ""
			continue
		} else if strings.HasPrefix(line, "### Open Analysis") {
			currentSubsection = "open_analysis"
			analysisContent = ""
			continue
		} else if strings.HasPrefix(line, "### Close Set") {
			currentSubsection = "close"
			summaryContent = ""
			continue
		} else if strings.HasPrefix(line, "### Close Summary") {
			currentSubsection = "close_summary"
			summaryContent = ""
			continue
		} else if strings.HasPrefix(line, "## Key Takeaways") {
			currentSection = "takeaways"
			currentSubsection = ""
			continue
		}

		// Fast parsing without excessive logging
		switch currentSection {
		case "morning":
			switch currentSubsection {
			case "open":
				if strings.HasPrefix(line, "* Open Index:") {
					data.MorningOpenIndex, data.MorningOpenChange = parseIndexLineFromMarkdown(line)
				} else if strings.HasPrefix(line, "* Highlights:") {
					data.MorningOpenHighlights = parseHighlightsFromMarkdown(line)
				}
			case "open_analysis":
				if strings.HasPrefix(line, "<p>") || analysisContent != "" {
					if analysisContent != "" {
						analysisContent += "\n"
					}
					analysisContent += line
					if strings.HasSuffix(line, "</p>") || (!strings.HasPrefix(line, "<") && line != "") {
						data.MorningOpenAnalysis = template.HTML(analysisContent)
					}
				}
			case "close":
				if strings.HasPrefix(line, "* Close Index:") {
					data.MorningCloseIndex, data.MorningCloseChange = parseIndexLineFromMarkdown(line)
				} else if strings.HasPrefix(line, "* Highlights:") {
					data.MorningCloseHighlights = parseHighlightsFromMarkdown(line)
				}
			case "close_summary":
				if strings.HasPrefix(line, "<p>") || summaryContent != "" {
					if summaryContent != "" {
						summaryContent += "\n"
					}
					summaryContent += line
					if strings.HasSuffix(line, "</p>") || (!strings.HasPrefix(line, "<") && line != "") {
						data.MorningCloseSummary = template.HTML(summaryContent)
					}
				}
			}
		case "afternoon":
			switch currentSubsection {
			case "open":
				if strings.HasPrefix(line, "* Open Index:") {
					data.AfternoonOpenIndex, data.AfternoonOpenChange = parseIndexLineFromMarkdown(line)
				} else if strings.HasPrefix(line, "* Highlights:") {
					data.AfternoonOpenHighlights = parseHighlightsFromMarkdown(line)
				}
			case "open_analysis":
				if strings.HasPrefix(line, "<p>") || analysisContent != "" {
					if analysisContent != "" {
						analysisContent += "\n"
					}
					analysisContent += line
					if strings.HasSuffix(line, "</p>") || (!strings.HasPrefix(line, "<") && line != "") {
						data.AfternoonOpenAnalysis = template.HTML(analysisContent)
					}
				}
			case "close":
				if strings.HasPrefix(line, "* Close Index:") {
					data.AfternoonCloseIndex, data.AfternoonCloseChange = parseIndexLineFromMarkdown(line)
				} else if strings.HasPrefix(line, "* Highlights:") {
					data.AfternoonCloseHighlights = parseHighlightsFromMarkdown(line)
				}
			case "close_summary":
				if strings.HasPrefix(line, "<p>") || summaryContent != "" {
					if summaryContent != "" {
						summaryContent += "\n"
					}
					summaryContent += line
					if strings.HasSuffix(line, "</p>") || (!strings.HasPrefix(line, "<") && line != "") {
						data.AfternoonCloseSummary = template.HTML(summaryContent)
					}
				}
			}
		case "takeaways":
			if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
				takeaway := strings.TrimSpace(line[1:])
				if takeaway != "" {
					data.KeyTakeaways = append(data.KeyTakeaways, takeaway)
				}
			}
		}
	}

	return data, nil
}

func parseIndexLineFromMarkdown(line string) (float64, float64) {
	re := regexp.MustCompile(`\*\s*(?:Open|Close) Index:\s*(\d+\.?\d*)\s*\(([+-]?\d+\.?\d*)\)`)
	matches := re.FindStringSubmatch(line)

	if len(matches) >= 3 {
		index, err1 := strconv.ParseFloat(matches[1], 64)
		change, err2 := strconv.ParseFloat(matches[2], 64)
		if err1 == nil && err2 == nil {
			return index, change
		}
	}
	return 0.0, 0.0
}

func parseHighlightsFromMarkdown(line string) string {
	if idx := strings.Index(line, "Highlights:"); idx != -1 {
		highlight := strings.TrimSpace(line[idx+11:])
		// Clean markdown bold formatting (**text**)
		highlight = strings.ReplaceAll(highlight, "**", "")
		return highlight
	}
	return ""
}

// Enhanced formatChangeValue function with auto calculation from previous close
// ...existing code...
// Enhanced formatChangeValue function with auto calculation from previous close
func formatChangeValue(change float64) string {
	if change > 0 {
		return fmt.Sprintf("+%.2f", change)
	} else if change < 0 {
		return fmt.Sprintf("%.2f", change) // Already has minus sign
	} else {
		return "+0.00" // Default to positive zero
	}
}

// Helper function to get base URL from request
func getBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, r.Host)
}

// Helper function to generate summary from API data
func generateSummaryFromAPI(data MarketDataRequest) string {
	highlights := []string{}

	if data.MorningOpen.Highlights != "" {
		highlights = append(highlights, data.MorningOpen.Highlights)
	}

	if data.AfternoonOpen.Highlights != "" {
		highlights = append(highlights, data.AfternoonOpen.Highlights)
	}

	if len(highlights) > 0 {
		return fmt.Sprintf("Market analysis featuring %s", strings.Join(highlights, ", "))
	}

	return "Thai stock market analysis with SET index movements and sector highlights"
}

// Helper function to generate markdown content from API data (basic version)
func generateMarkdownFromAPI(data MarketDataRequest) string {
	var content strings.Builder

	content.WriteString("## Morning Session\n\n")
	content.WriteString("### Open Set\n")

	if data.MorningOpen.Index > 0 {
		changeStr := formatChangeValue(data.MorningOpen.Change)
		content.WriteString(fmt.Sprintf("* Open Index: %.2f (%s)\n", data.MorningOpen.Index, changeStr))
	}

	if data.MorningOpen.Highlights != "" {
		content.WriteString(fmt.Sprintf("* Highlights: **%s**\n", data.MorningOpen.Highlights))
	}

	content.WriteString("\n<hr>\n\n### Close Set\n")

	if data.MorningClose.Index > 0 {
		changeStr := formatChangeValue(data.MorningClose.Change)
		content.WriteString(fmt.Sprintf("* Close Index: %.2f (%s)\n", data.MorningClose.Index, changeStr))
	}

	content.WriteString("\n<hr>\n\n## Afternoon Session\n\n")
	content.WriteString("### Open Set\n")

	if data.AfternoonOpen.Index > 0 {
		changeStr := formatChangeValue(data.AfternoonOpen.Change)
		content.WriteString(fmt.Sprintf("* Open Index: %.2f (%s)\n", data.AfternoonOpen.Index, changeStr))
	}

	if data.AfternoonOpen.Highlights != "" {
		content.WriteString(fmt.Sprintf("* Highlights: **%s**\n", data.AfternoonOpen.Highlights))
	}

	content.WriteString("\n<hr>\n\n### Close Set\n")

	if data.AfternoonClose.Index > 0 {
		changeStr := formatChangeValue(data.AfternoonClose.Change)
		content.WriteString(fmt.Sprintf("* Close Index: %.2f (%s)\n", data.AfternoonClose.Index, changeStr))
	}

	return content.String()
}

func init() {
	// Initialize database
	initDB()

	// Sync filesystem articles to database on startup
	addMissingArticlesToDB()
}

// Update the main() function around line 616
// Update the main() function around line 616

func main() {
	// Test Telegram configuration on startup
	if DEBUG_MODE {
		go func() {
			time.Sleep(2 * time.Second) // Wait for server to start
			debugLog("🔍 Getting Telegram chat information...")
			getTelegramChatID()
		}()
	}

	// Existing routes
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/articles/", articleHandler)
	http.HandleFunc("/admin", adminDashboardHandler)
	http.HandleFunc("/admin/articles/new", adminArticleFormHandler)
	http.HandleFunc("/admin/articles/edit/", adminEditArticleHandler)
	http.HandleFunc("/admin/articles/delete/", adminDeleteArticleHandler)

	// Add the API routes
	http.HandleFunc("/api/market-data", apiMarketDataHandler)
	http.HandleFunc("/api/market-data-analysis", apiMarketDataWithAnalysisHandler)
	http.HandleFunc("/api/market-data-close", apiMarketDataWithCloseAnalysisHandler) // NEW ENDPOINT

	// Static file server
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("src/static/"))))

	// Server configuration following ThaiStockAnalysis performance patterns
	debugLog("ThaiStockAnalysis server starting on http://localhost:7777")
	debugLog("Admin dashboard: http://localhost:7777/admin")
	debugLog("API endpoint (open): http://localhost:7777/api/market-data-analysis")
	debugLog("API endpoint (close): http://localhost:7777/api/market-data-close") // NEW
	debugLog("API endpoint (basic): http://localhost:7777/api/market-data")

	log.Fatal(http.ListenAndServe(":7777", nil))
}

// Replace the sendTelegramNotification function around line 649

// Enhanced Telegram notification with session-specific messaging and better error handling
// Replace the sendTelegramNotification function around line 649

// Enhanced Telegram notification with proper escaping and chat ID detection
// Replace the sendTelegramNotification function around line 649

// Replace the sendTelegramNotification function around line 649

// Enhanced Telegram notification with proper message length handling
// ...existing code...

// Enhanced Telegram notification with dynamic URL support
func sendTelegramNotification(openIndex, openChange float64, highlights, analysis, sessionType, baseURL string) {
	if TELEGRAM_BOT_TOKEN == "YOUR_TELEGRAM_BOT_TOKEN" || TELEGRAM_BOT_TOKEN == "YOUR_ACTUAL_TELEGRAM_BOT_TOKEN_HERE" {
		debugLog("Telegram bot token not configured - skipping notification")
		return
	}

	changeStr := formatChangeValue(openChange)

	var emoji string
	var timeInfo string

	if strings.Contains(sessionType, "Morning") {
		emoji = "🌅"
		timeInfo = "9:30 AM"
	} else {
		emoji = "🌆"
		timeInfo = "2:30 PM"
	}

	// Truncate analysis for Telegram (4096 character limit) but keep it readable
	truncatedAnalysis := analysis
	if len(analysis) > 300 {
		// Find a good breaking point near 300 characters
		words := strings.Fields(analysis)
		truncated := ""
		for _, word := range words {
			if len(truncated+" "+word) > 280 {
				break
			}
			if truncated == "" {
				truncated = word
			} else {
				truncated += " " + word
			}
		}
		truncatedAnalysis = truncated + "... (Full analysis available on website)"
	}

	// Create Telegram message with dynamic URL
	message := fmt.Sprintf("%s Thai Stock Market - %s\n\n"+
		"📊 Index: %.2f (%s)\n"+
		"🎯 Highlights: %s\n"+
		"⏰ Time: %s\n\n"+
		"🤖 AI Analysis:\n%s\n\n"+
		"🔗 Full analysis: %s",
		emoji, sessionType, openIndex, changeStr, highlights, timeInfo, truncatedAnalysis, baseURL)

	telegramMsg := TelegramMessage{
		ChatID:    TELEGRAM_CHAT_ID,
		Text:      message,
		ParseMode: "",
	}

	msgBody, err := json.Marshal(telegramMsg)
	if err != nil {
		debugLog("Error marshaling Telegram message: %v", err)
		return
	}

	telegramURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", TELEGRAM_BOT_TOKEN)

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Post(telegramURL, "application/json", bytes.NewBuffer(msgBody))
	if err != nil {
		debugLog("Error sending Telegram message: %v", err)
		return
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		debugLog("Error reading Telegram response: %v", err)
		return
	}

	var telegramResponse TelegramResponse
	err = json.Unmarshal(responseBody, &telegramResponse)
	if err != nil {
		debugLog("Error parsing Telegram response: %v", err)
		return
	}

	if telegramResponse.Ok {
		debugLog("✅ Telegram notification sent successfully for %s: Message ID %d", sessionType, telegramResponse.Result.MessageID)
	} else {
		debugLog("❌ Telegram API error (Code: %d): %s", telegramResponse.ErrorCode, telegramResponse.Description)
	}
}

// ...existing code...
// Add this function before main()

// Helper function to get your Telegram chat ID
func getTelegramChatID() {
	if TELEGRAM_BOT_TOKEN == "YOUR_TELEGRAM_BOT_TOKEN" || TELEGRAM_BOT_TOKEN == "YOUR_ACTUAL_TELEGRAM_BOT_TOKEN_HERE" {
		debugLog("❌ Telegram bot token not configured")
		return
	}

	// Get updates to find your chat ID
	updatesURL := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates", TELEGRAM_BOT_TOKEN)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(updatesURL)

	if err != nil {
		debugLog("❌ Error getting Telegram updates: %v", err)
		return
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		debugLog("❌ Error reading Telegram updates response: %v", err)
		return
	}

	debugLog("📱 Telegram Updates Response: %s", string(responseBody))
	debugLog("💡 To get your chat ID: Message your bot first, then check the response above for 'chat':{'id':NUMBER}")
}

// Basic API handler for receiving market data without AI analysis
// ...existing code...

// Basic API handler for receiving market data without AI analysis
func apiMarketDataHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers for API access
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only accept POST requests
	if r.Method != "POST" {
		response := MarketDataResponse{
			Success: false,
			Error:   "Method not allowed. Use POST.",
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get dynamic base URL
	baseURL := getBaseURL(r)

	// Parse JSON request body
	var requestData MarketDataRequest
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		debugLog("API JSON parsing error: %v", err)
		response := MarketDataResponse{
			Success: false,
			Error:   "Invalid JSON format",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate required fields
	if requestData.Date == "" {
		response := MarketDataResponse{
			Success: false,
			Error:   "Date field is required (YYYY-MM-DD format)",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Generate markdown content from API data (without AI analysis)
	content := generateMarkdownFromAPI(requestData)

	// Parse date for title generation
	parsedDate, _ := time.Parse("2006-01-02", requestData.Date)
	title := fmt.Sprintf("Stock Market Analysis - %s", parsedDate.Format("2 January 2006"))
	summary := generateSummaryFromAPI(requestData)

	// Check if article already exists
	var existingID int
	err = db.QueryRow("SELECT id FROM articles WHERE slug = ?", requestData.Date).Scan(&existingID)

	if err == sql.ErrNoRows {
		// Create new article
		result, err := db.Exec(`INSERT INTO articles (slug, title, summary, content, created_at) VALUES (?, ?, ?, ?, ?)`,
			requestData.Date, title, summary, content, requestData.Date)

		if err != nil {
			debugLog("API database insert error: %v", err)
			response := MarketDataResponse{
				Success: false,
				Error:   "Database error creating article",
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		newID, _ := result.LastInsertId()

		// Write markdown file for dual storage
		filename := fmt.Sprintf("articles/%s.md", requestData.Date)
		os.WriteFile(filename, []byte(content), 0644)

		// Clear caches for immediate updates
		clearMarkdownCache(filename)
		clearTemplateCache()

		response := MarketDataResponse{
			Success: true,
			Message: "Article created successfully",
			Data: struct {
				ArticleID int    `json:"article_id,omitempty"`
				Slug      string `json:"slug,omitempty"`
				URL       string `json:"url,omitempty"`
			}{
				ArticleID: int(newID),
				Slug:      requestData.Date,
				URL:       fmt.Sprintf("%s/articles/%s", baseURL, requestData.Date),
			},
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)

	} else if err == nil {
		// Update existing article
		_, err = db.Exec(`UPDATE articles SET title = ?, summary = ?, content = ? WHERE id = ?`,
			title, summary, content, existingID)

		if err != nil {
			debugLog("API database update error: %v", err)
			response := MarketDataResponse{
				Success: false,
				Error:   "Database error updating article",
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		filename := fmt.Sprintf("articles/%s.md", requestData.Date)
		os.WriteFile(filename, []byte(content), 0644)

		clearMarkdownCache(filename)
		clearTemplateCache()

		response := MarketDataResponse{
			Success: true,
			Message: "Article updated successfully",
			Data: struct {
				ArticleID int    `json:"article_id,omitempty"`
				Slug      string `json:"slug,omitempty"`
				URL       string `json:"url,omitempty"`
			}{
				ArticleID: existingID,
				Slug:      requestData.Date,
				URL:       fmt.Sprintf("%s/articles/%s", baseURL, requestData.Date),
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// ...existing code...

// Admin delete article handler
func adminDeleteArticleHandler(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	path := r.URL.Path
	idStr := strings.TrimPrefix(path, "/admin/articles/delete/")

	if idStr == "" {
		http.Redirect(w, r, "/admin?error=Article ID required", http.StatusSeeOther)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/admin?error=Invalid article ID", http.StatusSeeOther)
		return
	}

	if r.Method == "POST" {
		// Get article slug before deletion for file cleanup
		var slug string
		err := db.QueryRow("SELECT slug FROM articles WHERE id = ?", id).Scan(&slug)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Redirect(w, r, "/admin?error=Article not found", http.StatusSeeOther)
				return
			}
			debugLog("Error getting article slug: %v", err)
			http.Redirect(w, r, "/admin?error=Database error", http.StatusSeeOther)
			return
		}

		// Delete from database with performance-optimized query
		_, err = db.Exec("DELETE FROM articles WHERE id = ?", id)
		if err != nil {
			debugLog("Database deletion error: %v", err)
			http.Redirect(w, r, "/admin?error=Failed to delete article", http.StatusSeeOther)
			return
		}

		// Delete markdown file following dual storage pattern
		markdownPath := fmt.Sprintf("articles/%s.md", slug)
		if err := os.Remove(markdownPath); err != nil {
			debugLog("Error deleting markdown file: %v", err)
			// Continue - database deletion succeeded
		}

		// Clear caches for immediate updates
		clearMarkdownCache(markdownPath)
		clearTemplateCache()

		// Redirect with success message
		http.Redirect(w, r, "/admin?success=Article deleted successfully", http.StatusSeeOther)

	} else {
		// Only allow POST method for deletion
		http.Redirect(w, r, "/admin?error=Invalid request method", http.StatusSeeOther)
	}
}

// Enhanced API handler with Gemini analysis integration into article sections
// ...existing code...

// Enhanced API handler with Gemini analysis integration into article sections
func apiMarketDataWithAnalysisHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers for API access
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only accept POST requests
	if r.Method != "POST" {
		response := MarketDataResponse{
			Success: false,
			Error:   "Method not allowed. Use POST.",
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get dynamic base URL
	baseURL := getBaseURL(r)

	// Parse JSON request body
	var requestData MarketDataRequest
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		debugLog("API JSON parsing error: %v", err)
		response := MarketDataResponse{
			Success: false,
			Error:   "Invalid JSON format",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate required fields
	if requestData.Date == "" {
		response := MarketDataResponse{
			Success: false,
			Error:   "Date field is required (YYYY-MM-DD format)",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Load existing content and merge with new data
	var existingContent string
	err = db.QueryRow("SELECT content FROM articles WHERE slug = ?", requestData.Date).Scan(&existingContent)

	var existingData StockData
	if err == nil && existingContent != "" {
		existingData = parseMarkdownContentForAdmin(existingContent)
		debugLog("API: Found existing content for %s, preserving data", requestData.Date)
	} else {
		existingData = StockData{
			CurrentDate:  time.Now().Format("2 January 2006"),
			KeyTakeaways: []string{},
		}
		debugLog("API: No existing content for %s, creating new", requestData.Date)
	}

	// Generate Gemini analysis and integrate into appropriate sections
	var analysisText string
	var sessionType string

	// Check for morning open data and generate analysis
	if requestData.MorningOpen.Index > 0 {
		sessionType = "Morning"
		analysis, err := generateGeminiAnalysis(requestData.MorningOpen.Index, requestData.MorningOpen.Change, requestData.MorningOpen.Highlights)
		if err != nil {
			debugLog("Gemini API error for morning open: %v", err)
			// Enhanced fallback analysis
			changeDirection := "gained"
			sentiment := "positive"
			if requestData.MorningOpen.Change < 0 {
				changeDirection = "declined"
				sentiment = "cautious"
			}

			analysisText = fmt.Sprintf(`<p>The SET Index opened at %.2f, having %s %.2f points in early trading. This opening reflects %s market sentiment as investors assess current economic conditions and sector-specific developments.</p>

<p>Key sector movements show: %s. These sectoral patterns indicate selective buying interest and suggest investors are focusing on specific themes and opportunities in the current market environment.</p>

<p>The opening price level of %.2f represents an important technical reference point. Market participants will be watching for follow-through buying or selling to confirm the sustainability of this opening momentum. Trading volume and breadth will be crucial indicators of underlying market conviction.</p>

<p>External factors including global market conditions, currency movements, and economic indicators continue to influence investor sentiment. The market's ability to maintain current levels will depend on both domestic economic developments and international market trends.</p>

<p>Moving forward, key levels to watch include immediate support and resistance zones around this opening price. Sustained movement above or below these levels could signal the direction for the remainder of the trading session.</p>`,
				requestData.MorningOpen.Index, changeDirection, math.Abs(requestData.MorningOpen.Change), sentiment,
				requestData.MorningOpen.Highlights, requestData.MorningOpen.Index)
		} else {
			analysisText = analysis
		}

		// Merge morning open data with Gemini analysis
		existingData.MorningOpenIndex = requestData.MorningOpen.Index
		existingData.MorningOpenChange = requestData.MorningOpen.Change
		if requestData.MorningOpen.Highlights != "" {
			existingData.MorningOpenHighlights = requestData.MorningOpen.Highlights
		}
		// Add Gemini analysis to morning open analysis section
		existingData.MorningOpenAnalysis = template.HTML(analysisText)

		debugLog("API: Updated morning open with Gemini analysis")

		// Send Telegram notification with dynamic URL
		go sendTelegramNotification(requestData.MorningOpen.Index, requestData.MorningOpen.Change,
			requestData.MorningOpen.Highlights, analysisText, "Morning Open", baseURL)
	}

	// Check for afternoon open data and generate analysis
	if requestData.AfternoonOpen.Index > 0 {
		sessionType = "Afternoon"
		analysis, err := generateGeminiAnalysis(requestData.AfternoonOpen.Index, requestData.AfternoonOpen.Change, requestData.AfternoonOpen.Highlights)
		if err != nil {
			debugLog("Gemini API error for afternoon open: %v", err)
			// Enhanced fallback analysis for afternoon
			changeDirection := "gained"
			sentiment := "positive"
			if requestData.AfternoonOpen.Change < 0 {
				changeDirection = "declined"
				sentiment = "cautious"
			}

			analysisText = fmt.Sprintf(`<p>The SET Index opened the afternoon session at %.2f, having %s %.2f points. This afternoon opening reflects %s market sentiment as trading resumes with fresh momentum and renewed investor interest.</p>

<p>Sector developments show: %s. These afternoon patterns indicate continued market dynamics and suggest that investors are responding to midday developments and reassessing their positions based on morning session performance.</p>

<p>The afternoon opening at %.2f establishes a crucial reference point for the remainder of the trading day. Market participants will closely monitor trading volumes and price action to gauge whether the afternoon session can build upon or diverge from morning trends.</p>

<p>Afternoon sessions often reflect institutional repositioning and foreign fund activity, making this opening particularly significant for understanding broader market sentiment. The ability to sustain current levels will be key to determining the day's overall market direction.</p>

<p>Key factors to monitor during the afternoon include follow-through on morning themes, any late-breaking news developments, and the market's response to global cues that may have emerged during the midday break. These elements will shape the session's trajectory.</p>`,
				requestData.AfternoonOpen.Index, changeDirection, math.Abs(requestData.AfternoonOpen.Change), sentiment,
				requestData.AfternoonOpen.Highlights, requestData.AfternoonOpen.Index)
		} else {
			analysisText = analysis
		}

		// Merge afternoon open data with Gemini analysis
		existingData.AfternoonOpenIndex = requestData.AfternoonOpen.Index
		existingData.AfternoonOpenChange = requestData.AfternoonOpen.Change
		if requestData.AfternoonOpen.Highlights != "" {
			existingData.AfternoonOpenHighlights = requestData.AfternoonOpen.Highlights
		}
		// Add Gemini analysis to afternoon open analysis section
		existingData.AfternoonOpenAnalysis = template.HTML(analysisText)

		debugLog("API: Updated afternoon open with Gemini analysis")

		// Send Telegram notification with dynamic URL
		go sendTelegramNotification(requestData.AfternoonOpen.Index, requestData.AfternoonOpen.Change,
			requestData.AfternoonOpen.Highlights, analysisText, "Afternoon Open", baseURL)
	}

	// Merge other session data without overwriting existing analysis
	if requestData.MorningClose.Index > 0 {
		existingData.MorningCloseIndex = requestData.MorningClose.Index
		existingData.MorningCloseChange = requestData.MorningClose.Change

		debugLog("API: Updated morning close data")
	}

	if requestData.AfternoonClose.Index > 0 {
		existingData.AfternoonCloseIndex = requestData.AfternoonClose.Index
		existingData.AfternoonCloseChange = requestData.AfternoonClose.Change

		debugLog("API: Updated afternoon close data")
	}

	// Generate complete markdown content with integrated analysis
	processedContent := generateEnhancedMarkdownFromData(existingData)

	// Parse date for title generation
	parsedDate, _ := time.Parse("2006-01-02", requestData.Date)
	title := fmt.Sprintf("Stock Market Analysis - %s", parsedDate.Format("2 January 2006"))
	summary := generateSummaryFromAPI(requestData)

	// Save to database and files
	var existingID int
	err = db.QueryRow("SELECT id FROM articles WHERE slug = ?", requestData.Date).Scan(&existingID)

	if err == sql.ErrNoRows {
		// Create new article
		result, err := db.Exec(`INSERT INTO articles (slug, title, summary, content, created_at) VALUES (?, ?, ?, ?, ?)`,
			requestData.Date, title, summary, processedContent, requestData.Date)

		if err != nil {
			debugLog("API database insert error: %v", err)
			response := MarketDataResponse{
				Success: false,
				Error:   "Database error creating article",
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		newID, _ := result.LastInsertId()

		// Write markdown file for dual storage
		filename := fmt.Sprintf("articles/%s.md", requestData.Date)
		os.WriteFile(filename, []byte(processedContent), 0644)

		// Clear caches for immediate updates
		clearMarkdownCache(filename)
		clearTemplateCache()

		response := MarketDataResponse{
			Success: true,
			Message: fmt.Sprintf("Article created with %s AI analysis and Telegram notification sent", sessionType),
			Data: struct {
				ArticleID int    `json:"article_id,omitempty"`
				Slug      string `json:"slug,omitempty"`
				URL       string `json:"url,omitempty"`
			}{
				ArticleID: int(newID),
				Slug:      requestData.Date,
				URL:       fmt.Sprintf("%s/articles/%s", baseURL, requestData.Date),
			},
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)

	} else if err == nil {
		// Update existing article
		_, err = db.Exec(`UPDATE articles SET title = ?, summary = ?, content = ? WHERE id = ?`,
			title, summary, processedContent, existingID)

		if err != nil {
			debugLog("API database update error: %v", err)
			response := MarketDataResponse{
				Success: false,
				Error:   "Database error updating article",
			}
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		filename := fmt.Sprintf("articles/%s.md", requestData.Date)
		os.WriteFile(filename, []byte(processedContent), 0644)

		clearMarkdownCache(filename)
		clearTemplateCache()

		response := MarketDataResponse{
			Success: true,
			Message: fmt.Sprintf("Article updated with %s AI analysis and Telegram notification sent", sessionType),
			Data: struct {
				ArticleID int    `json:"article_id,omitempty"`
				Slug      string `json:"slug,omitempty"`
				URL       string `json:"url,omitempty"`
			}{
				ArticleID: existingID,
				Slug:      requestData.Date,
				URL:       fmt.Sprintf("%s/articles/%s", baseURL, requestData.Date),
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// ...existing code...
// Function to generate analysis using Gemini API
// Function to generate analysis using Gemini API (around line 996)
// Fix the generateGeminiAnalysis function around line 1260

// Replace the generateGeminiAnalysis function around line 1260

// Enhanced Gemini analysis with better prompt debugging
// Replace the generateGeminiAnalysis function around line 1260

// Enhanced Gemini analysis with full content preservation
func generateGeminiAnalysis(openIndex, openChange float64, highlights string) (string, error) {
	// Read prompt template from getanalysis_prompt file
	promptTemplate, err := os.ReadFile("getanalysis_prompt")
	if err != nil {
		debugLog("Error reading prompt template: %v", err)
		return "", err
	}

	// Replace placeholders following ThaiStockAnalysis conventions
	prompt := string(promptTemplate)
	prompt = strings.ReplaceAll(prompt, "$open_set", fmt.Sprintf("%.2f", openIndex))
	prompt = strings.ReplaceAll(prompt, "$change", formatChangeValue(openChange))
	prompt = strings.ReplaceAll(prompt, "$sectors", highlights)

	debugLog("📝 Final Gemini prompt: %s", prompt)

	// Prepare Gemini API request following performance patterns
	geminiRequest := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{
						Text: prompt,
					},
				},
			},
		},
	}

	requestBody, err := json.Marshal(geminiRequest)
	if err != nil {
		debugLog("Error marshaling Gemini request: %v", err)
		return "", err
	}

	// Use gemini-2.0-flash-lite-001 for faster responses with optimized timeout
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-lite-001:generateContent?key=%s", GEMINI_API_KEY)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		debugLog("Error calling Gemini API: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		debugLog("Error reading Gemini response: %v", err)
		return "", err
	}

	var geminiResponse GeminiResponse
	err = json.Unmarshal(responseBody, &geminiResponse)
	if err != nil {
		debugLog("Error parsing Gemini response: %v", err)
		return "", err
	}

	if geminiResponse.Error != nil {
		debugLog("Gemini API error: %s", geminiResponse.Error.Message)
		return "", fmt.Errorf("Gemini API error: %s", geminiResponse.Error.Message)
	}

	// Extract full generated text without truncation
	if len(geminiResponse.Candidates) > 0 && len(geminiResponse.Candidates[0].Content.Parts) > 0 {
		analysis := geminiResponse.Candidates[0].Content.Parts[0].Text

		// Keep full analysis for article content - no truncation
		debugLog("✅ Generated Gemini analysis: %s", analysis)
		return strings.TrimSpace(analysis), nil
	}

	return "", fmt.Errorf("no analysis generated by Gemini")
}

// Fix the typo in generateEnhancedMarkdownFromData function around line 1301

// Enhanced markdown generator that includes analysis sections
func generateEnhancedMarkdownFromData(data StockData) string {
	var content strings.Builder

	// Write front matter with current date
	content.WriteString(fmt.Sprintf("# Stock Market Analysis - %s\n\n", data.CurrentDate))

	// Morning Session Section
	content.WriteString("## Morning Session\n\n")

	// Morning Open Section - CRITICAL: Always include if data exists
	if data.MorningOpenIndex > 0 {
		content.WriteString("### Open Set\n\n")
		changeStr := formatChangeValue(data.MorningOpenChange)
		content.WriteString(fmt.Sprintf("* Open Index: %.2f (%s)\n", data.MorningOpenIndex, changeStr))

		if data.MorningOpenHighlights != "" {
			content.WriteString(fmt.Sprintf("* Highlights: **%s**\n", data.MorningOpenHighlights))
		}
		content.WriteString("\n")

		// Morning Open Analysis - CRITICAL: Preserve HTML analysis
		if data.MorningOpenAnalysis != "" {
			content.WriteString("### Open Analysis\n\n")
			content.WriteString(string(data.MorningOpenAnalysis))
			content.WriteString("\n\n")
		}
	}

	content.WriteString("<hr>\n\n")

	// Morning Close Section - Add if data exists
	if data.MorningCloseIndex > 0 {
		content.WriteString("### Close Set\n\n")
		changeStr := formatChangeValue(data.MorningCloseChange)
		content.WriteString(fmt.Sprintf("* Close Index: %.2f (%s)\n", data.MorningCloseIndex, changeStr))
		content.WriteString("\n")

		// Morning Close Summary - Add if exists
		if data.MorningCloseSummary != "" {
			content.WriteString("### Close Summary\n\n")
			content.WriteString(string(data.MorningCloseSummary))
			content.WriteString("\n\n")
		}
	}

	content.WriteString("<hr>\n\n")

	// Afternoon Session Section
	content.WriteString("## Afternoon Session\n\n")

	// Afternoon Open Section
	if data.AfternoonOpenIndex > 0 {
		content.WriteString("### Open Set\n\n")
		changeStr := formatChangeValue(data.AfternoonOpenChange)
		content.WriteString(fmt.Sprintf("* Open Index: %.2f (%s)\n", data.AfternoonOpenIndex, changeStr))

		if data.AfternoonOpenHighlights != "" {
			content.WriteString(fmt.Sprintf("* Highlights: **%s**\n", data.AfternoonOpenHighlights))
		}
		content.WriteString("\n")

		// Afternoon Open Analysis
		if data.AfternoonOpenAnalysis != "" {
			content.WriteString("### Open Analysis\n\n")
			content.WriteString(string(data.AfternoonOpenAnalysis))
			content.WriteString("\n\n")
		}
	}

	content.WriteString("<hr>\n\n")

	// Afternoon Close Section
	if data.AfternoonCloseIndex > 0 {
		content.WriteString("### Close Set\n\n")
		changeStr := formatChangeValue(data.AfternoonCloseChange)
		content.WriteString(fmt.Sprintf("* Close Index: %.2f (%s)\n", data.AfternoonCloseIndex, changeStr))
		content.WriteString("\n")

		// Afternoon Close Summary
		if data.AfternoonCloseSummary != "" {
			content.WriteString("### Close Summary\n\n")
			content.WriteString(string(data.AfternoonCloseSummary))
			content.WriteString("\n\n")
		}
	}

	// Key Takeaways section
	if len(data.KeyTakeaways) > 0 {
		content.WriteString("## Key Takeaways\n\n")
		for _, takeaway := range data.KeyTakeaways {
			content.WriteString(fmt.Sprintf("* %s\n", takeaway))
		}
		content.WriteString("\n")
	}

	return content.String()
}
func init() {
	// Initialize database
	initDB()

	// Sync filesystem articles to database on startup
	addMissingArticlesToDB()
}

// Helper function to strip HTML tags for clean form display
func stripHTMLTags(content string) string {
	// Remove common HTML tags for clean admin form display
	content = strings.ReplaceAll(content, "<p>", "")
	content = strings.ReplaceAll(content, "</p>", "")
	content = strings.ReplaceAll(content, "<br>", "\n")
	content = strings.ReplaceAll(content, "<br/>", "\n")
	content = strings.ReplaceAll(content, "<br />", "\n")

	return strings.TrimSpace(content)
}

// Helper function to parse markdown content specifically for admin forms
// Replace the parseMarkdownContentForAdmin function around line 1385

// Enhanced markdown parser following ThaiStockAnalysis structured data extraction

func parseMarkdownContentForAdmin(markdownContent string) StockData {
	debugLog("🔍 Parsing markdown content for admin, length: %d", len(markdownContent))

	data := StockData{
		CurrentDate:  time.Now().Format("2 January 2006"),
		KeyTakeaways: []string{},
	}

	lines := strings.Split(markdownContent, "\n")

	var currentSection string
	var analysisContent strings.Builder
	var summaryContent strings.Builder
	var inAnalysis bool
	var inSummary bool

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Section headers
		if strings.HasPrefix(line, "## Morning Session") {
			currentSection = "morning"
			inAnalysis = false
			inSummary = false
			continue
		} else if strings.HasPrefix(line, "## Afternoon Session") {
			currentSection = "afternoon"
			inAnalysis = false
			inSummary = false
			continue
		}

		// Subsection headers
		if strings.HasPrefix(line, "### Open Set") {
			inAnalysis = false
			inSummary = false
			continue
		} else if strings.HasPrefix(line, "### Close Set") {
			inAnalysis = false
			inSummary = false
			continue
		} else if strings.HasPrefix(line, "### Open Analysis") {
			inAnalysis = true
			inSummary = false
			analysisContent.Reset()
			continue
		} else if strings.HasPrefix(line, "### Close Summary") {
			inAnalysis = false
			inSummary = true
			summaryContent.Reset()
			continue
		}

		// Stop analysis/summary collection on new section
		if strings.HasPrefix(line, "###") || strings.HasPrefix(line, "##") || line == "<hr>" {
			if inAnalysis {
				// Save collected analysis
				if currentSection == "morning" && analysisContent.Len() > 0 {
					data.MorningOpenAnalysis = template.HTML(strings.TrimSpace(analysisContent.String()))
					debugLog("✅ Parsed morning open analysis: %d chars", len(data.MorningOpenAnalysis))
				} else if currentSection == "afternoon" && analysisContent.Len() > 0 {
					data.AfternoonOpenAnalysis = template.HTML(strings.TrimSpace(analysisContent.String()))
					debugLog("✅ Parsed afternoon open analysis: %d chars", len(data.AfternoonOpenAnalysis))
				}
				inAnalysis = false
				analysisContent.Reset()
			}
			if inSummary {
				// Save collected summary
				if currentSection == "morning" && summaryContent.Len() > 0 {
					data.MorningCloseSummary = template.HTML(strings.TrimSpace(summaryContent.String()))
					debugLog("✅ Parsed morning close summary: %d chars", len(data.MorningCloseSummary))
				} else if currentSection == "afternoon" && summaryContent.Len() > 0 {
					data.AfternoonCloseSummary = template.HTML(strings.TrimSpace(summaryContent.String()))
					debugLog("✅ Parsed afternoon close summary: %d chars", len(data.AfternoonCloseSummary))
				}
				inSummary = false
				summaryContent.Reset()
			}
			continue
		}

		// Collect analysis content
		if inAnalysis {
			if analysisContent.Len() > 0 {
				analysisContent.WriteString("\n")
			}
			analysisContent.WriteString(line)
			continue
		}

		// Collect summary content
		if inSummary {
			if summaryContent.Len() > 0 {
				summaryContent.WriteString("\n")
			}
			summaryContent.WriteString(line)
			continue
		}

		// Parse index values and highlights
		if strings.HasPrefix(line, "* Open Index:") || strings.HasPrefix(line, "* Close Index:") {
			// Extract index and change values using regex
			indexRegex := regexp.MustCompile(`(\d+\.?\d*)\s*\(([+-]?\d+\.?\d*)\)`)
			matches := indexRegex.FindStringSubmatch(line)

			if len(matches) == 3 {
				index, _ := strconv.ParseFloat(matches[1], 64)
				change, _ := strconv.ParseFloat(matches[2], 64)

				if strings.Contains(line, "Open Index:") {
					if currentSection == "morning" {
						data.MorningOpenIndex = index
						data.MorningOpenChange = change
						debugLog("✅ Parsed morning open: %.2f (%.2f)", index, change)
					} else if currentSection == "afternoon" {
						data.AfternoonOpenIndex = index
						data.AfternoonOpenChange = change
						debugLog("✅ Parsed afternoon open: %.2f (%.2f)", index, change)
					}
				} else if strings.Contains(line, "Close Index:") {
					if currentSection == "morning" {
						data.MorningCloseIndex = index
						data.MorningCloseChange = change
						debugLog("✅ Parsed morning close: %.2f (%.2f)", index, change)
					} else if currentSection == "afternoon" {
						data.AfternoonCloseIndex = index
						data.AfternoonCloseChange = change
						debugLog("✅ Parsed afternoon close: %.2f (%.2f)", index, change)
					}
				}
			}
		}

		// Parse highlights
		if strings.HasPrefix(line, "* Highlights:") {
			highlights := strings.TrimPrefix(line, "* Highlights:")
			highlights = strings.Trim(highlights, " *")

			if currentSection == "morning" {
				data.MorningOpenHighlights = highlights
				debugLog("✅ Parsed morning highlights: %s", highlights)
			} else if currentSection == "afternoon" {
				data.AfternoonOpenHighlights = highlights
				debugLog("✅ Parsed afternoon highlights: %s", highlights)
			}
		}

		// Parse key takeaways
		if strings.HasPrefix(line, "## Key Takeaways") {
			currentSection = "takeaways"
			continue
		}

		if currentSection == "takeaways" && strings.HasPrefix(line, "* ") {
			takeaway := strings.TrimPrefix(line, "* ")
			data.KeyTakeaways = append(data.KeyTakeaways, takeaway)
		}
	}

	// Handle end-of-content analysis/summary
	if inAnalysis && analysisContent.Len() > 0 {
		if currentSection == "morning" {
			data.MorningOpenAnalysis = template.HTML(strings.TrimSpace(analysisContent.String()))
			debugLog("✅ Final morning open analysis: %d chars", len(data.MorningOpenAnalysis))
		} else if currentSection == "afternoon" {
			data.AfternoonOpenAnalysis = template.HTML(strings.TrimSpace(analysisContent.String()))
			debugLog("✅ Final afternoon open analysis: %d chars", len(data.AfternoonOpenAnalysis))
		}
	}

	if inSummary && summaryContent.Len() > 0 {
		if currentSection == "morning" {
			data.MorningCloseSummary = template.HTML(strings.TrimSpace(summaryContent.String()))
			debugLog("✅ Final morning close summary: %d chars", len(data.MorningCloseSummary))
		} else if currentSection == "afternoon" {
			data.AfternoonCloseSummary = template.HTML(strings.TrimSpace(summaryContent.String()))
			debugLog("✅ Final afternoon close summary: %d chars", len(data.AfternoonCloseSummary))
		}
	}

	debugLog("🔍 Parse complete - Morning Open: %.2f, Analysis: %d chars",
		data.MorningOpenIndex, len(data.MorningOpenAnalysis))

	return data
}

// Helper function to clear markdown cache for specific file
func clearMarkdownCache(filePath string) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// Remove specific file from cache
	delete(markdownCache, filePath)
	delete(cacheExpiry, filePath)
}

// Helper function to clear template cache
func clearTemplateCache() {
	templateMutex.Lock()
	defer templateMutex.Unlock()

	// Clear all cached templates
	templateCache = make(map[string]*template.Template)
}

// Helper function to generate markdown content from form data
func generateMarkdownContent(r *http.Request) string {
	var content strings.Builder

	content.WriteString("## Morning Session\n\n")
	content.WriteString("### Open Set\n")

	morningOpenIndex := r.FormValue("morning_open_index")
	morningOpenChangeInput := r.FormValue("morning_open_change")
	if morningOpenIndex != "" {
		// Use the change value as entered (preserving +/- signs)
		changeStr := "0.00"
		if morningOpenChangeInput != "" {
			// Clean and validate the input but preserve the sign
			cleanInput := strings.TrimSpace(morningOpenChangeInput)
			if cleanInput != "" {
				// Try to parse to validate it's a number
				if _, err := strconv.ParseFloat(strings.TrimPrefix(cleanInput, "+"), 64); err == nil {
					changeStr = cleanInput
					// Ensure positive numbers have + sign
					if !strings.HasPrefix(changeStr, "+") && !strings.HasPrefix(changeStr, "-") {
						changeStr = "+" + changeStr
					}
				}
			}
		}
		content.WriteString(fmt.Sprintf("* Open Index: %s (%s)\n", morningOpenIndex, changeStr))
	}

	morningOpenHighlights := r.FormValue("morning_open_highlights")
	if morningOpenHighlights != "" {
		content.WriteString(fmt.Sprintf("* Highlights: **%s**\n", morningOpenHighlights))
	}

	content.WriteString("\n### Open Analysis\n")
	morningOpenAnalysis := r.FormValue("morning_open_analysis")
	if morningOpenAnalysis != "" {
		content.WriteString(fmt.Sprintf("<p>%s</p>\n", morningOpenAnalysis))
	}

	content.WriteString("\n<hr>\n\n### Close Set\n")

	morningCloseIndex := r.FormValue("morning_close_index")
	morningCloseChangeInput := r.FormValue("morning_close_change")
	if morningCloseIndex != "" {
		changeStr := "0.00"
		if morningCloseChangeInput != "" {
			cleanInput := strings.TrimSpace(morningCloseChangeInput)
			if cleanInput != "" {
				if _, err := strconv.ParseFloat(strings.TrimPrefix(cleanInput, "+"), 64); err == nil {
					changeStr = cleanInput
					if !strings.HasPrefix(changeStr, "+") && !strings.HasPrefix(changeStr, "-") {
						changeStr = "+" + changeStr
					}
				}
			}
		}
		content.WriteString(fmt.Sprintf("* Close Index: %s (%s)\n", morningCloseIndex, changeStr))
	}

	morningCloseHighlights := r.FormValue("morning_close_highlights")
	if morningCloseHighlights != "" {
		content.WriteString(fmt.Sprintf("* Highlights: **%s**\n", morningCloseHighlights))
	}

	content.WriteString("\n### Close Summary\n")
	morningCloseSummary := r.FormValue("morning_close_summary")
	if morningCloseSummary != "" {
		content.WriteString(fmt.Sprintf("<p>%s</p>\n", morningCloseSummary))
	}

	content.WriteString("\n<hr>\n\n## Afternoon Session\n\n")
	content.WriteString("### Open Set\n")

	afternoonOpenIndex := r.FormValue("afternoon_open_index")
	afternoonOpenChangeInput := r.FormValue("afternoon_open_change")
	if afternoonOpenIndex != "" {
		changeStr := "0.00"
		if afternoonOpenChangeInput != "" {
			cleanInput := strings.TrimSpace(afternoonOpenChangeInput)
			if cleanInput != "" {
				if _, err := strconv.ParseFloat(strings.TrimPrefix(cleanInput, "+"), 64); err == nil {
					changeStr = cleanInput
					if !strings.HasPrefix(changeStr, "+") && !strings.HasPrefix(changeStr, "-") {
						changeStr = "+" + changeStr
					}
				}
			}
		}
		content.WriteString(fmt.Sprintf("* Open Index: %s (%s)\n", afternoonOpenIndex, changeStr))
	}

	afternoonOpenHighlights := r.FormValue("afternoon_open_highlights")
	if afternoonOpenHighlights != "" {
		content.WriteString(fmt.Sprintf("* Highlights: **%s**\n", afternoonOpenHighlights))
	}

	content.WriteString("\n### Open Analysis\n")
	afternoonOpenAnalysis := r.FormValue("afternoon_open_analysis")
	if afternoonOpenAnalysis != "" {
		content.WriteString(fmt.Sprintf("<p>%s</p>\n", afternoonOpenAnalysis))
	}

	content.WriteString("\n<hr>\n\n### Close Set\n")

	afternoonCloseIndex := r.FormValue("afternoon_close_index")
	afternoonCloseChangeInput := r.FormValue("afternoon_close_change")
	if afternoonCloseIndex != "" {
		changeStr := "0.00"
		if afternoonCloseChangeInput != "" {
			cleanInput := strings.TrimSpace(afternoonCloseChangeInput)
			if cleanInput != "" {
				if _, err := strconv.ParseFloat(strings.TrimPrefix(cleanInput, "+"), 64); err == nil {
					changeStr = cleanInput
					if !strings.HasPrefix(changeStr, "+") && !strings.HasPrefix(changeStr, "-") {
						changeStr = "+" + changeStr
					}
				}
			}
		}
		content.WriteString(fmt.Sprintf("* Close Index: %s (%s)\n", afternoonCloseIndex, changeStr))
	}

	afternoonCloseHighlights := r.FormValue("afternoon_close_highlights")
	if afternoonCloseHighlights != "" {
		content.WriteString(fmt.Sprintf("* Highlights: **%s**\n", afternoonCloseHighlights))
	}

	content.WriteString("\n### Close Summary\n")
	afternoonCloseSummary := r.FormValue("afternoon_close_summary")
	if afternoonCloseSummary != "" {
		content.WriteString(fmt.Sprintf("<p>%s</p>\n", afternoonCloseSummary))
	}

	content.WriteString("\n<hr>\n\n## Key Takeaways\n\n")
	keyTakeaways := r.FormValue("key_takeaways")
	if keyTakeaways != "" {
		takeaways := strings.Split(keyTakeaways, "\n")
		for _, takeaway := range takeaways {
			takeaway = strings.TrimSpace(takeaway)
			if takeaway != "" {
				content.WriteString(fmt.Sprintf("- %s\n", takeaway))
			}
		}
	}

	return content.String()
}

// addMissingArticlesToDB syncs filesystem articles to database on startup
func addMissingArticlesToDB() {
	// Get all markdown files from articles directory
	files, err := os.ReadDir("articles")
	if err != nil {
		debugLog("Error reading articles directory: %v", err)
		return
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			slug := strings.TrimSuffix(file.Name(), ".md")

			// Check if article already exists in database
			var count int
			err := db.QueryRow("SELECT COUNT(*) FROM articles WHERE slug = ?", slug).Scan(&count)
			if err != nil {
				debugLog("Error checking article existence: %v", err)
				continue
			}

			// If article doesn't exist in database, add it
			if count == 0 {
				content, err := os.ReadFile(fmt.Sprintf("articles/%s", file.Name()))
				if err != nil {
					debugLog("Error reading file %s: %v", file.Name(), err)
					continue
				}

				// Generate title from slug
				title := generateTitleFromSlug(slug)

				// Extract summary from content (first key takeaway or analysis)
				summary := extractSummaryFromContent(string(content))

				// Use slug as created_at date (assuming YYYY-MM-DD format)
				createdAt := slug

				// Insert into database
				_, err = db.Exec(`INSERT INTO articles (slug, title, summary, content, created_at) VALUES (?, ?, ?, ?, ?)`,
					slug, title, summary, string(content), createdAt)

				if err != nil {
					debugLog("Error inserting article %s: %v", slug, err)
				} else {
					debugLog("Added article %s to database", slug)
				}
			}
		}
	}
}

// Helper function to generate title from slug
func generateTitleFromSlug(slug string) string {
	// Parse YYYY-MM-DD format
	if len(slug) == 10 && slug[4] == '-' && slug[7] == '-' {
		if date, err := time.Parse("2006-01-02", slug); err == nil {
			return fmt.Sprintf("Stock Market Analysis - %s", date.Format("2 January 2006"))
		}
	}
	return fmt.Sprintf("Stock Market Analysis - %s", slug)
}

// Helper function to extract summary from markdown content
func extractSummaryFromContent(content string) string {
	lines := strings.Split(content, "\n")

	// Look for key takeaways first
	inTakeaways := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "## Key Takeaways") {
			inTakeaways = true
			continue
		}
		if inTakeaways && (strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*")) {
			takeaway := strings.TrimSpace(line[1:])
			if len(takeaway) > 10 {
				return takeaway
			}
		}
		if inTakeaways && strings.HasPrefix(line, "##") {
			break
		}
	}

	// Fall back to first analysis section
	inAnalysis := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Analysis") && strings.HasPrefix(line, "###") {
			inAnalysis = true
			continue
		}
		if inAnalysis && strings.HasPrefix(line, "<p>") {
			text := strings.TrimPrefix(line, "<p>")
			text = strings.TrimSuffix(text, "</p>")
			if len(text) > 20 {
				return text
			}
		}
		if inAnalysis && strings.HasPrefix(line, "##") {
			break
		}
	}

	return "Market analysis and insights for the trading day"
}

// SUPER FAST index handler - database-only queries for maximum performance
func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := getTemplate("index")
	if err != nil {
		debugLog("Template loading error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Database-only query for SUPER FAST performance - no file system access
	rows, err := db.Query("SELECT slug, title, summary, created_at FROM articles ORDER BY created_at DESC LIMIT 10")
	if err != nil {
		debugLog("Database query error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var articles []ArticlePreview
	for rows.Next() {
		var slug, title, createdAt string
		var summary sql.NullString

		err := rows.Scan(&slug, &title, &summary, &createdAt)
		if err != nil {
			debugLog("Row scan error: %v", err)
			continue
		}

		// Parse date for display
		var displayDate string
		if date, err := time.Parse("2006-01-02", slug); err == nil {
			displayDate = date.Format("2 January 2006")
		} else {
			displayDate = createdAt
		}

		// Use summary from database or generate fallback
		summaryText := "Market analysis and insights for the trading day"
		if summary.Valid && summary.String != "" {
			summaryText = summary.String
		}

		article := ArticlePreview{
			Title:        title,
			Date:         displayDate,
			ShortSummary: summaryText,
			Summary:      summaryText,
			Slug:         slug,
			URL:          fmt.Sprintf("/articles/%s", slug),
		}
		articles = append(articles, article)
	}

	data := IndexPageData{
		CurrentDate: time.Now().Format("2 January 2006"),
		Articles:    articles,
	}

	debugLog("SUPER FAST index page loaded with %d articles from database only", len(articles))

	err = tmpl.ExecuteTemplate(w, "base.gohtml", data)
	if err != nil {
		debugLog("Template execution error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Replace the articleHandler function around line 1720

// Fast article handler with cached markdown parsing following ThaiStockAnalysis patterns
func articleHandler(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/articles/")
	if slug == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	tmpl, err := getTemplate("article")
	if err != nil {
		debugLog("Template loading error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Try to get from markdown file first for detailed content (performance-optimized caching)
	filePath := fmt.Sprintf("articles/%s.md", slug)
	data, err := getCachedStockData(filePath)
	if err != nil {
		// Fallback to database content following dual storage pattern
		var content sql.NullString
		err = db.QueryRow("SELECT content FROM articles WHERE slug = ?", slug).Scan(&content)
		if err != nil {
			debugLog("Article not found: %s", slug)
			http.NotFound(w, r)
			return
		}

		if content.Valid && content.String != "" {
			data = parseMarkdownContentForAdmin(content.String)
		} else {
			debugLog("No content found for article: %s", slug)
			http.NotFound(w, r)
			return
		}
	}

	// Parse date from slug for display and set title (ThaiStockAnalysis convention)
	if date, err := time.Parse("2006-01-02", slug); err == nil {
		data.CurrentDate = date.Format("2 January 2006")
		data.Title = fmt.Sprintf("Thai Stock Market Analysis - %s", date.Format("2 January 2006"))
	} else {
		data.Title = fmt.Sprintf("Thai Stock Market Analysis - %s", slug)
	}

	debugLog("Article page loaded: %s with cached parsing", slug)

	// Execute template with proper base template following ThaiStockAnalysis template patterns
	err = tmpl.ExecuteTemplate(w, "base.gohtml", data)
	if err != nil {
		debugLog("Template execution error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Admin dashboard handler
func adminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := getTemplate("admin")
	if err != nil {
		debugLog("Template loading error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get all articles from database
	rows, err := db.Query("SELECT id, slug, title, summary, created_at FROM articles ORDER BY created_at DESC")
	if err != nil {
		debugLog("Database query error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var articles []DBArticle
	for rows.Next() {
		var article DBArticle
		err := rows.Scan(&article.ID, &article.Slug, &article.Title, &article.Summary, &article.CreatedAt)
		if err != nil {
			debugLog("Row scan error: %v", err)
			continue
		}
		articles = append(articles, article)
	}

	// Get success/error messages from URL parameters
	success := r.URL.Query().Get("success")
	errorMsg := r.URL.Query().Get("error")

	data := AdminDashboardData{
		Articles:    articles,
		CurrentDate: time.Now().Format("2 January 2006"),
		Success:     success,
		Error:       errorMsg,
	}

	debugLog("Admin dashboard loaded with %d articles", len(articles))

	err = tmpl.ExecuteTemplate(w, "admin.gohtml", data)
	if err != nil {
		debugLog("Template execution error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// Admin article form handler (create new article)
func adminArticleFormHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Show the form - use proper template loading for admin forms
		tmpl, err := template.New("").Funcs(template.FuncMap{
			"printf": fmt.Sprintf,
			"html":   func(s string) template.HTML { return template.HTML(s) },
			"add":    func(a, b int) int { return a + b },
		}).ParseFiles("src/templates/admin_article_form.gohtml")

		if err != nil {
			debugLog("Template loading error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		data := AdminArticleFormData{
			IsEdit:      false,
			Action:      "/admin/articles/new",
			CurrentDate: time.Now().Format("2006-01-02"),
		}

		err = tmpl.ExecuteTemplate(w, "admin_article_form.gohtml", data)
		if err != nil {
			debugLog("Template execution error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	} else if r.Method == "POST" {
		// Process the form
		slug := r.FormValue("slug")
		title := r.FormValue("title")
		summary := r.FormValue("summary")

		if slug == "" || title == "" {
			tmpl, _ := template.New("").ParseFiles("src/templates/admin_article_form.gohtml")
			data := AdminArticleFormData{
				IsEdit:      false,
				Action:      "/admin/articles/new",
				Error:       "Slug and title are required",
				CurrentDate: time.Now().Format("2006-01-02"),
			}
			tmpl.ExecuteTemplate(w, "admin_article_form.gohtml", data)
			return
		}

		content := generateMarkdownContent(r)

		_, err := db.Exec(`INSERT INTO articles (slug, title, summary, content, created_at) VALUES (?, ?, ?, ?, ?)`,
			slug, title, summary, content, slug)

		if err != nil {
			debugLog("Database insert error: %v", err)
			tmpl, _ := template.New("").ParseFiles("src/templates/admin_article_form.gohtml")
			data := AdminArticleFormData{
				IsEdit:      false,
				Action:      "/admin/articles/new",
				Error:       "Error creating article: " + err.Error(),
				CurrentDate: time.Now().Format("2006-01-02"),
			}
			tmpl.ExecuteTemplate(w, "admin_article_form.gohtml", data)
			return
		}

		filename := fmt.Sprintf("articles/%s.md", slug)
		err = os.WriteFile(filename, []byte(content), 0644)
		if err != nil {
			debugLog("Error writing markdown file: %v", err)
		}

		clearMarkdownCache(filename)
		clearTemplateCache()

		http.Redirect(w, r, "/admin?success=Article created successfully", http.StatusSeeOther)
	}
}

// Complete the adminEditArticleHandler function (replace the incomplete version)

// Admin edit article handler
func adminEditArticleHandler(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL path
	path := r.URL.Path
	idStr := strings.TrimPrefix(path, "/admin/articles/edit/")

	if idStr == "" {
		http.Redirect(w, r, "/admin?error=Article ID required", http.StatusSeeOther)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Redirect(w, r, "/admin?error=Invalid article ID", http.StatusSeeOther)
		return
	}

	if r.Method == "GET" {
		// Get article from database
		var article DBArticle
		err := db.QueryRow("SELECT id, slug, title, summary, content FROM articles WHERE id = ?", id).Scan(
			&article.ID, &article.Slug, &article.Title, &article.Summary, &article.Content)

		if err != nil {
			if err == sql.ErrNoRows {
				http.Redirect(w, r, "/admin?error=Article not found", http.StatusSeeOther)
				return
			}
			debugLog("Database error: %v", err)
			http.Redirect(w, r, "/admin?error=Database error", http.StatusSeeOther)
			return
		}

		// Parse existing content for form
		var stockData StockData
		if article.Content.Valid && article.Content.String != "" {
			stockData = parseMarkdownContentForAdmin(article.Content.String)
		}

		// Show the form with existing data - use direct template parsing for admin forms
		tmpl, err := template.New("").Funcs(template.FuncMap{
			"printf": fmt.Sprintf,
			"html":   func(s string) template.HTML { return template.HTML(s) },
			"add":    func(a, b int) int { return a + b },
		}).ParseFiles("src/templates/admin_article_form.gohtml")

		if err != nil {
			debugLog("Template loading error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		data := AdminArticleFormData{
			IsEdit:                   true,
			Action:                   fmt.Sprintf("/admin/articles/edit/%d", id),
			Article:                  article,
			CurrentDate:              time.Now().Format("2006-01-02"),
			MorningOpenIndex:         fmt.Sprintf("%.2f", stockData.MorningOpenIndex),
			MorningOpenChange:        stockData.MorningOpenChange,
			MorningOpenHighlights:    stockData.MorningOpenHighlights,
			MorningOpenAnalysis:      stripHTMLTags(string(stockData.MorningOpenAnalysis)),
			MorningCloseIndex:        fmt.Sprintf("%.2f", stockData.MorningCloseIndex),
			MorningCloseChange:       stockData.MorningCloseChange,
			MorningCloseHighlights:   stockData.MorningCloseHighlights,
			MorningCloseSummary:      stripHTMLTags(string(stockData.MorningCloseSummary)),
			AfternoonOpenIndex:       fmt.Sprintf("%.2f", stockData.AfternoonOpenIndex),
			AfternoonOpenChange:      stockData.AfternoonOpenChange,
			AfternoonOpenHighlights:  stockData.AfternoonOpenHighlights,
			AfternoonOpenAnalysis:    stripHTMLTags(string(stockData.AfternoonOpenAnalysis)),
			AfternoonCloseIndex:      fmt.Sprintf("%.2f", stockData.AfternoonCloseIndex),
			AfternoonCloseChange:     stockData.AfternoonCloseChange,
			AfternoonCloseHighlights: stockData.AfternoonCloseHighlights,
			AfternoonCloseSummary:    stripHTMLTags(string(stockData.AfternoonCloseSummary)),
			KeyTakeaways:             strings.Join(stockData.KeyTakeaways, "\n"),
		}

		err = tmpl.ExecuteTemplate(w, "admin_article_form.gohtml", data)
		if err != nil {
			debugLog("Template execution error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}

	} else if r.Method == "POST" {
		// Update article
		title := r.FormValue("title")
		summary := r.FormValue("summary")
		slug := r.FormValue("slug")

		// Validate required fields
		if title == "" || slug == "" {
			http.Redirect(w, r, fmt.Sprintf("/admin/articles/edit/%d?error=Title and slug are required", id), http.StatusSeeOther)
			return
		}

		// Generate markdown content from form
		content := generateMarkdownContent(r)

		// Update in database
		_, err := db.Exec(`UPDATE articles SET title = ?, summary = ?, content = ? WHERE id = ?`,
			title, summary, content, id)

		if err != nil {
			debugLog("Database update error: %v", err)
			http.Redirect(w, r, fmt.Sprintf("/admin/articles/edit/%d?error=Update failed", id), http.StatusSeeOther)
			return
		}

		// Update markdown file for dual storage
		filename := fmt.Sprintf("articles/%s.md", slug)
		err = os.WriteFile(filename, []byte(content), 0644)
		if err != nil {
			debugLog("Error writing markdown file: %v", err)
		}

		// Clear caches for immediate updates
		clearMarkdownCache(filename)
		clearTemplateCache()

		// Redirect to admin dashboard with success message
		http.Redirect(w, r, "/admin?success=Article updated successfully", http.StatusSeeOther)
	}
}
