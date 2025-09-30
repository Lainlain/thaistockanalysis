package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"thaistockanalysis/internal/models"

	"github.com/gomarkdown/markdown"
)

// Cache for parsed markdown files
var (
	markdownCache = make(map[string]models.StockData)
	cacheMutex    sync.RWMutex
	cacheExpiry   = make(map[string]time.Time)
)

// Template cache for performance
var (
	templateCache = make(map[string]*template.Template)
	templateMutex sync.RWMutex
)

// MardownService handles markdown file parsing and caching
type MarkdownService struct {
	cacheExpiry time.Duration
}

// NewMarkdownService creates a new markdown service
func NewMarkdownService(cacheExpiryMinutes int) *MarkdownService {
	return &MarkdownService{
		cacheExpiry: time.Duration(cacheExpiryMinutes) * time.Minute,
	}
}

// GetCachedStockData retrieves stock data from cache or parses if not cached
func (ms *MarkdownService) GetCachedStockData(filePath string) (models.StockData, error) {
	// If cache is disabled (0 minutes), always parse fresh
	if ms.cacheExpiry == 0 {
		return ms.ParseMarkdownArticle(filePath)
	}

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
	data, err := ms.ParseMarkdownArticle(filePath)
	if err != nil {
		return data, err
	}

	cacheMutex.Lock()
	markdownCache[filePath] = data
	cacheExpiry[filePath] = time.Now().Add(ms.cacheExpiry)
	cacheMutex.Unlock()

	return data, nil
}

// ParseMarkdownArticle parses a markdown file into structured stock data
func (ms *MarkdownService) ParseMarkdownArticle(filePath string) (models.StockData, error) {
	data := models.StockData{
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

		// Subsections - support both old and new formats
		if strings.HasPrefix(line, "### Open Set") || strings.HasPrefix(line, "### Market Opening Data") {
			currentSubsection = "open"
			analysisContent = ""
			continue
		} else if strings.HasPrefix(line, "### Open Analysis") || strings.HasPrefix(line, "### Market Analysis") {
			currentSubsection = "open_analysis"
			analysisContent = ""
			continue
		} else if strings.HasPrefix(line, "### Close Set") || strings.HasPrefix(line, "### Market Closing Data") {
			currentSubsection = "close"
			summaryContent = ""
			continue
		} else if strings.HasPrefix(line, "### Close Summary") || strings.HasPrefix(line, "### Market Summary") {
			currentSubsection = "close_summary"
			summaryContent = ""
			continue
		} else if strings.HasPrefix(line, "## Key Takeaways") {
			currentSection = "takeaways"
			currentSubsection = ""
			continue
		}

		// Parse content based on section and subsection
		switch currentSection {
		case "morning":
			ms.parseMorningSession(line, currentSubsection, &data, &analysisContent, &summaryContent)
		case "afternoon":
			ms.parseAfternoonSession(line, currentSubsection, &data, &analysisContent, &summaryContent)
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

// parseMorningSession handles parsing of morning session data
func (ms *MarkdownService) parseMorningSession(line, subsection string, data *models.StockData, analysisContent, summaryContent *string) {
	switch subsection {
	case "open":
		if strings.HasPrefix(line, "* Open Index:") || strings.HasPrefix(line, "* Index:") {
			data.MorningOpenIndex, data.MorningOpenChange = ms.parseIndexLine(line)
		} else if strings.HasPrefix(line, "* Highlights:") {
			data.MorningOpenHighlights = ms.parseHighlights(line)
		} else if data.MorningOpenHighlights != "" && line != "" && !strings.HasPrefix(line, "###") && !strings.HasPrefix(line, "##") && !strings.HasPrefix(line, "*") {
			// Continue collecting highlights content that spans multiple lines
			if data.MorningOpenHighlights != "" {
				data.MorningOpenHighlights += "\n\n" + line
			}
		}
	case "open_analysis":
		if strings.HasPrefix(line, "<p>") || *analysisContent != "" {
			if *analysisContent != "" {
				*analysisContent += "\n"
			}
			*analysisContent += line
			if strings.HasSuffix(line, "</p>") || (!strings.HasPrefix(line, "<") && line != "") {
				data.MorningOpenAnalysis = template.HTML(markdown.ToHTML([]byte(*analysisContent), nil, nil))
			}
		}
	case "close":
		if strings.HasPrefix(line, "* Close Index:") {
			data.MorningCloseIndex, data.MorningCloseChange = ms.parseIndexLine(line)
		} else if strings.HasPrefix(line, "* Highlights:") {
			data.MorningCloseHighlights = ms.parseHighlights(line)
		}
	case "close_summary":
		if strings.HasPrefix(line, "<p>") || *summaryContent != "" {
			if *summaryContent != "" {
				*summaryContent += "\n"
			}
			*summaryContent += line
			if strings.HasSuffix(line, "</p>") || (!strings.HasPrefix(line, "<") && line != "") {
				data.MorningCloseSummary = template.HTML(markdown.ToHTML([]byte(*summaryContent), nil, nil))
			}
		}
	}
}

// parseAfternoonSession handles parsing of afternoon session data
func (ms *MarkdownService) parseAfternoonSession(line, subsection string, data *models.StockData, analysisContent, summaryContent *string) {
	switch subsection {
	case "open":
		if strings.HasPrefix(line, "* Open Index:") || strings.HasPrefix(line, "* Index:") {
			data.AfternoonOpenIndex, data.AfternoonOpenChange = ms.parseIndexLine(line)
		} else if strings.HasPrefix(line, "* Highlights:") {
			data.AfternoonOpenHighlights = ms.parseHighlights(line)
		} else if data.AfternoonOpenHighlights != "" && line != "" && !strings.HasPrefix(line, "###") && !strings.HasPrefix(line, "##") && !strings.HasPrefix(line, "*") {
			// Continue collecting highlights content that spans multiple lines
			if data.AfternoonOpenHighlights != "" {
				data.AfternoonOpenHighlights += "\n\n" + line
			}
		}
	case "open_analysis":
		if strings.HasPrefix(line, "<p>") || *analysisContent != "" {
			if *analysisContent != "" {
				*analysisContent += "\n"
			}
			*analysisContent += line
			if strings.HasSuffix(line, "</p>") || (!strings.HasPrefix(line, "<") && line != "") {
				data.AfternoonOpenAnalysis = template.HTML(markdown.ToHTML([]byte(*analysisContent), nil, nil))
			}
		}
	case "close":
		if strings.HasPrefix(line, "* Close Index:") || strings.HasPrefix(line, "* Index:") {
			data.AfternoonCloseIndex, data.AfternoonCloseChange = ms.parseIndexLine(line)
		} else if strings.HasPrefix(line, "* Highlights:") {
			data.AfternoonCloseHighlights = ms.parseHighlights(line)
		}
	case "close_summary":
		if strings.HasPrefix(line, "<p>") || *summaryContent != "" {
			if *summaryContent != "" {
				*summaryContent += "\n"
			}
			*summaryContent += line
			if strings.HasSuffix(line, "</p>") || (!strings.HasPrefix(line, "<") && line != "") {
				data.AfternoonCloseSummary = template.HTML(markdown.ToHTML([]byte(*summaryContent), nil, nil))
			}
		}
	}
}

// parseIndexLine extracts index value and change from a line
func (ms *MarkdownService) parseIndexLine(line string) (float64, float64) {
	// Parse "* Open Index: 1270.96 (4.85)" or "* Close Index: 1275.40 (9.29)"
	re := regexp.MustCompile(`(\d+\.?\d*)\s*\(([+-]?\d+\.?\d*)\)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) >= 3 {
		index, _ := strconv.ParseFloat(matches[1], 64)
		change, _ := strconv.ParseFloat(matches[2], 64)
		return index, change
	}
	return 0, 0
}

// parseHighlights extracts highlights from a line
func (ms *MarkdownService) parseHighlights(line string) string {
	// Remove "* Highlights: " prefix
	if strings.HasPrefix(line, "* Highlights: ") {
		content := strings.TrimSpace(line[14:])
		// Replace <br> tags with actual newlines for proper display
		content = strings.ReplaceAll(content, "<br>", "\n")
		content = strings.ReplaceAll(content, "<br/>", "\n")
		content = strings.ReplaceAll(content, "<br />", "\n")
		return content
	}
	return line
}

// ClearCache clears the markdown cache for a specific file
func (ms *MarkdownService) ClearCache(filePath string) {
	cacheMutex.Lock()
	delete(markdownCache, filePath)
	delete(cacheExpiry, filePath)
	cacheMutex.Unlock()
}

// TemplateService handles template caching and rendering
type TemplateService struct{}

// NewTemplateService creates a new template service
func NewTemplateService() *TemplateService {
	return &TemplateService{}
}

// GetTemplate retrieves a cached template or loads and caches it
func (ts *TemplateService) GetTemplate(name string, files ...string) (*template.Template, error) {
	templateMutex.RLock()
	if tmpl, exists := templateCache[name]; exists {
		templateMutex.RUnlock()
		return tmpl, nil
	}
	templateMutex.RUnlock()

	// Create template with custom functions
	funcMap := template.FuncMap{
		"printf":         fmt.Sprintf,
		"html":           func(s string) template.HTML { return template.HTML(s) },
		"add":            func(a, b int) int { return a + b },
		"markdownToHTML": func(s string) template.HTML { return template.HTML(markdown.ToHTML([]byte(s), nil, nil)) },
	}

	// Parse templates
	tmpl := template.New("base.gohtml").Funcs(funcMap)
	tmpl, err := tmpl.ParseFiles(files...)
	if err != nil {
		return nil, err
	}

	// Cache the template
	templateMutex.Lock()
	templateCache[name] = tmpl
	templateMutex.Unlock()

	return tmpl, nil
}

// ClearTemplateCache clears all cached templates
func (ts *TemplateService) ClearTemplateCache() {
	templateMutex.Lock()
	defer templateMutex.Unlock()
	templateCache = make(map[string]*template.Template)
}

// TelegramService handles Telegram bot messaging
type TelegramService struct {
	BotToken string
	Channel  string
}

// NewTelegramService creates a new Telegram service
func NewTelegramService(botToken, channel string) *TelegramService {
	return &TelegramService{
		BotToken: botToken,
		Channel:  channel,
	}
}

// TelegramMessage represents a Telegram bot message
type TelegramMessage struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

// SendMarketUpdate sends a market update message to the Telegram channel
func (ts *TelegramService) SendMarketUpdate(sessionType, openIndex, change, date string) error {
	if ts.BotToken == "" || ts.Channel == "" {
		log.Printf("⚠️  Telegram not configured, skipping notification")
		return nil
	}

	// Determine session time and create Myanmar language message
	var sessionTime, myanmarTitle string
	if strings.Contains(strings.ToLower(sessionType), "morning") {
		sessionTime = "12:01 PM"
		myanmarTitle = fmt.Sprintf("%s(%s) အတွက် Thai Stock Analysis ဂဏန်းများရပါပြီ", date, sessionTime)
	} else {
		sessionTime = "4:30 PM"
		myanmarTitle = fmt.Sprintf("%s(%s) အတွက် Thai Stock Analysis ဂဏန်းများရပါပြီ", date, sessionTime)
	}

	message := fmt.Sprintf("📊 *Thai Stock Market - %s*\n\n", sessionType)
	message += fmt.Sprintf("🔍 *Open Index:* `%s`\n", openIndex)
	message += fmt.Sprintf("📈 *Change:* `%s`\n\n", change)
	message += fmt.Sprintf("📅 *%s*\n\n", myanmarTitle)
	message += "အောက်ကလင့်ခ်ကိုနှိပ်ပြီးကြည့်ပါ\n"
	message += "🌐 https://thaistockanalysis.com"

	telegramMsg := TelegramMessage{
		ChatID:    ts.Channel,
		Text:      message,
		ParseMode: "Markdown",
	}

	jsonData, err := json.Marshal(telegramMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal Telegram message: %v", err)
	}

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", ts.BotToken)

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send Telegram message: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Telegram API returned status code: %d", resp.StatusCode)
	}

	log.Printf("✅ Telegram notification sent: %s - Index: %s, Change: %s", sessionType, openIndex, change)
	return nil
}

// ExtractMarketData extracts Open Index and Change from market data text
func (ts *TelegramService) ExtractMarketData(text string) (openIndex, change string) {
	// Pattern to match "Open Index: 1295.80 (+5.15)" format
	indexPattern := regexp.MustCompile(`(?i)open\s+index[:\s]+([0-9,]+\.?\d*)\s*\(([+-]?[0-9,]+\.?\d*)\)`)

	matches := indexPattern.FindStringSubmatch(text)
	if len(matches) >= 3 {
		openIndex = matches[1]
		change = matches[2]
		return
	}

	// Fallback: try to find any number patterns that might be index values
	numberPattern := regexp.MustCompile(`([0-9,]+\.?\d*)\s*\(([+-]?[0-9,]+\.?\d*)\)`)
	matches = numberPattern.FindStringSubmatch(text)
	if len(matches) >= 3 {
		openIndex = matches[1]
		change = matches[2]
		return
	}

	return "N/A", "N/A"
}
