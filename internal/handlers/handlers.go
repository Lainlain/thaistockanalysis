package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	config "thaistockanalysis/configs"
	"thaistockanalysis/internal/database"
	"thaistockanalysis/internal/models"
	"thaistockanalysis/internal/services"
)

// Handler contains dependencies for HTTP handlers
type Handler struct {
	MarkdownService *services.MarkdownService
	TemplateService *services.TemplateService
	TelegramService *services.TelegramService
	PromptService   *services.PromptService // Added PromptService
	ArticlesDir     string
	TemplateDir     string
	Config          *config.Config
}

// NewHandler creates a new handler with dependencies
func NewHandler(articlesDir, templateDir string, cfg *config.Config) *Handler {
	// Initialize PromptService
	promptService, err := services.NewPromptService("highlights_for_prompt.json")
	if err != nil {
		log.Fatalf("Failed to create PromptService: %v", err)
	}

	return &Handler{
		MarkdownService: services.NewMarkdownService(cfg.CacheExpiry),
		TemplateService: services.NewTemplateService(),
		TelegramService: services.NewTelegramService(cfg.TelegramBotToken, cfg.TelegramChannel),
		PromptService:   promptService, // Use the initialized service
		ArticlesDir:     articlesDir,
		TemplateDir:     templateDir,
		Config:          cfg,
	}
}

// IndexHandler handles the homepage
func (h *Handler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	// Get articles from database
	articles, err := database.GetArticles(20)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	var previews []models.ArticlePreview
	for _, article := range articles {
		// Try to load actual data from markdown file
		markdownPath := fmt.Sprintf("%s/%s.md", h.ArticlesDir, article.Slug)
		var setIndex string = "--"
		var change float64 = 0.0
		var shortSummary string

		// Parse markdown file to get real data
		if stockData, err := h.MarkdownService.GetCachedStockData(markdownPath); err == nil {
			// Determine the summary message first based on afternoon open data
			if stockData.AfternoonOpenIndex > 0 {
				shortSummary = "Daily full analysis available"
			} else {
				shortSummary = "Morning session analysis available."
			}

			// Use the most recent close index available for display
			if stockData.AfternoonCloseIndex > 0 {
				setIndex = fmt.Sprintf("%.2f", stockData.AfternoonCloseIndex)
				change = stockData.AfternoonCloseChange
			} else if stockData.MorningCloseIndex > 0 {
				setIndex = fmt.Sprintf("%.2f", stockData.MorningCloseIndex)
				change = stockData.MorningCloseChange
			} else if stockData.AfternoonOpenIndex > 0 {
				setIndex = fmt.Sprintf("%.2f", stockData.AfternoonOpenIndex)
				change = stockData.AfternoonOpenChange
			} else if stockData.MorningOpenIndex > 0 {
				setIndex = fmt.Sprintf("%.2f", stockData.MorningOpenIndex)
				change = stockData.MorningOpenChange
			}
		} else {
			shortSummary = "Failed to load analysis."
			fmt.Printf("Failed to parse markdown file %s: %v\n", markdownPath, err)
		}

		date, err := time.Parse("2006-01-02", article.CreatedAt)
		if err != nil {
			date = time.Now()
		}

		preview := models.ArticlePreview{
			Title:        article.Title,
			Date:         date.Format("2 Jan 2006"),
			SetIndex:     setIndex,
			Change:       change,
			ShortSummary: shortSummary,
			Slug:         article.Slug,
			URL:          fmt.Sprintf("/articles/%s", article.Slug),
		}
		previews = append(previews, preview)
	}

	data := models.IndexPageData{
		CurrentDate: time.Now().Format("2 January 2006"),
		Articles:    previews,
	}

	// Use cached templates
	tmpl, err := h.TemplateService.GetTemplate("index",
		fmt.Sprintf("%s/base.gohtml", h.TemplateDir),
		fmt.Sprintf("%s/index.gohtml", h.TemplateDir))
	if err != nil {
		fmt.Printf("Template error: %v\n", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "base.gohtml", data); err != nil {
		fmt.Printf("Template execution error: %v\n", err)
		http.Error(w, "Internal Server Error", 500)
		return
	}
}

// ArticleHandler handles individual article pages
func (h *Handler) ArticleHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")

	if len(parts) < 3 || parts[2] == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	slug := parts[2]

	// Fast database lookup
	dbArticle, err := database.GetArticleBySlug(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Fast default data
	stockData := models.StockData{
		CurrentDate:  time.Now().Format("2 January 2006"),
		KeyTakeaways: []string{},
	}

	// Use cached markdown parsing
	markdownPath := fmt.Sprintf("%s/%s.md", h.ArticlesDir, slug)
	if parsedData, err := h.MarkdownService.GetCachedStockData(markdownPath); err == nil {
		stockData = parsedData
	}

	data := models.ArticleDetail{
		Title:     dbArticle.Title,
		Slug:      dbArticle.Slug,
		Summary:   dbArticle.Summary.String,
		CreatedAt: dbArticle.CreatedAt,
		StockData: stockData,
	}

	// Use cached templates
	tmpl, err := h.TemplateService.GetTemplate("article",
		fmt.Sprintf("%s/base.gohtml", h.TemplateDir),
		fmt.Sprintf("%s/article.gohtml", h.TemplateDir))
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	tmpl.ExecuteTemplate(w, "base.gohtml", data)
}

// AdminDashboardHandler handles the admin dashboard
func (h *Handler) AdminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	success := r.URL.Query().Get("success")
	errorMsg := r.URL.Query().Get("error")

	articles, err := database.GetArticles(0) // Get all articles
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	data := models.AdminDashboardData{
		CurrentDate: time.Now().Format("2 January 2006"),
		Articles:    articles,
		Success:     success,
		Error:       errorMsg,
	}

	tmpl, err := h.TemplateService.GetTemplate("admin",
		fmt.Sprintf("%s/base.gohtml", h.TemplateDir),
		fmt.Sprintf("%s/admin.gohtml", h.TemplateDir))
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	tmpl.ExecuteTemplate(w, "base.gohtml", data)
}

// AdminArticleFormHandler handles article creation and editing
func (h *Handler) AdminArticleFormHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		formData := models.AdminArticleFormData{
			CurrentDate: time.Now().Format("2 January 2006"),
			IsEdit:      false,
			Action:      "/admin/articles/new",
		}

		tmpl, err := h.TemplateService.GetTemplate("admin_form",
			fmt.Sprintf("%s/base.gohtml", h.TemplateDir),
			fmt.Sprintf("%s/admin_article_form.gohtml", h.TemplateDir))
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			return
		}

		tmpl.ExecuteTemplate(w, "base.gohtml", formData)
	} else if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Error parsing form", 400)
			return
		}

		slug := r.FormValue("slug")
		title := r.FormValue("title")
		summary := r.FormValue("summary")

		markdownContent := fmt.Sprintf(`## Morning Session

### Open Set
* Open Index: 0.00 (0.00)
* Highlights: %s

### Open Analysis
<p>Morning analysis will be updated.</p>

<hr>

### Close Set
* Close Index: 0.00 (0.00)
* Highlights: %s

### Close Summary
<p>Morning session data will be updated at 12:30 PM.</p>

<hr>

## Afternoon Session

### Open Set
* Open Index: 0.00 (0.00)
* Highlights: %s

### Open Analysis
<p>Afternoon analysis will be updated.</p>

<hr>

### Close Set
* Close Index: 0.00 (0.00)
* Highlights: %s

### Close Summary
<p>Afternoon session data will be updated at 5:00 PM.</p>

<hr>

## Key Takeaways

- Market analysis pending
- Full analysis available after market close
`, summary, summary, summary, summary)

		markdownPath := fmt.Sprintf("%s/%s.md", h.ArticlesDir, slug)
		os.WriteFile(markdownPath, []byte(markdownContent), 0644)

		// Clear cache for this file
		h.MarkdownService.ClearCache(markdownPath)

		err = database.CreateArticle(slug, title, summary, markdownContent)
		if err != nil {
			http.Error(w, "Error creating article", 500)
			return
		}

		http.Redirect(w, r, "/admin?success=Article created successfully", 302)
	}
}

// PrivacyHandler handles the privacy policy page
func (h *Handler) PrivacyHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		LastUpdated string
	}{
		LastUpdated: "September 26, 2025",
	}

	tmpl, err := h.TemplateService.GetTemplate("privacy",
		fmt.Sprintf("%s/base.gohtml", h.TemplateDir),
		fmt.Sprintf("%s/privacy.gohtml", h.TemplateDir))
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
}

// TermsHandler handles the terms of service page
func (h *Handler) TermsHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		LastUpdated string
	}{
		LastUpdated: "September 26, 2025",
	}

	tmpl, err := h.TemplateService.GetTemplate("terms",
		fmt.Sprintf("%s/base.gohtml", h.TemplateDir),
		fmt.Sprintf("%s/terms.gohtml", h.TemplateDir))
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
}

// DisclaimerHandler handles the investment disclaimer page
func (h *Handler) DisclaimerHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		LastUpdated string
	}{
		LastUpdated: "September 26, 2025",
	}

	tmpl, err := h.TemplateService.GetTemplate("disclaimer",
		fmt.Sprintf("%s/base.gohtml", h.TemplateDir),
		fmt.Sprintf("%s/disclaimer.gohtml", h.TemplateDir))
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
}

// AboutHandler handles the about page
func (h *Handler) AboutHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		LastUpdated string
	}{
		LastUpdated: "September 26, 2025",
	}

	tmpl, err := h.TemplateService.GetTemplate("about",
		fmt.Sprintf("%s/base.gohtml", h.TemplateDir),
		fmt.Sprintf("%s/about.gohtml", h.TemplateDir))
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
}

// ContactHandler handles the contact page
func (h *Handler) ContactHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		CurrentDate string
	}{
		CurrentDate: time.Now().Format("2 January 2006"),
	}

	tmpl, err := h.TemplateService.GetTemplate("contact",
		fmt.Sprintf("%s/base.gohtml", h.TemplateDir),
		fmt.Sprintf("%s/contact.gohtml", h.TemplateDir))
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
}

// Gemini AI API structures
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

type GeminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
}

type GeminiCandidate struct {
	Content GeminiContent `json:"content"`
}

// loadHumanStylePrompt loads and formats the human-style prompt template
func (h *Handler) loadHumanStylePrompt(date, sessionType, openOrClose, indexValue, indexChange, highlights string) (string, error) {
	promptFile := "getanalysis_prompt_human.txt"
	content, err := os.ReadFile(promptFile)
	if err != nil {
		log.Printf("Warning: Could not load human prompt template: %v", err)
		// Return basic fallback prompt
		return fmt.Sprintf(`Generate professional Thai stock market %s session analysis for %s:
Index: %s (%s)
Key Highlights: %s

Provide engaging analysis covering market sentiment, technical outlook, and recommendations.
Write in English, keep under 300 words, format as 3-4 paragraphs.`,
			sessionType, date, indexValue, indexChange, highlights), nil
	}

	// Replace placeholders with actual data
	replacer := strings.NewReplacer(
		"{date}", date,
		"{session_type}", sessionType,
		"{open_or_close}", openOrClose,
		"{index_value}", indexValue,
		"{index_change}", indexChange,
		"{highlights}", highlights,
	)

	return replacer.Replace(string(content)), nil
}

// loadHumanStyleClosePrompt loads and formats the human-style closing prompt template
func (h *Handler) loadHumanStyleClosePrompt(date, sessionType, openingIndex, openingChange, closingIndex, closingChange, sessionPerformance string) (string, error) {
	promptFile := "getanalysis_prompt_close_human.txt"
	content, err := os.ReadFile(promptFile)
	if err != nil {
		log.Printf("Warning: Could not load closing prompt template: %v", err)
		// Return basic fallback prompt
		return fmt.Sprintf(`Generate brief Thai stock market %s session summary for %s:
Opening: %s (%s)
Closing: %s (%s)
Session: %s

Provide concise analysis covering session performance, sentiment, technical outlook, and recommendations.
Write in English, keep under 200 words, format as 3-4 paragraphs.`,
			sessionType, date, openingIndex, openingChange, closingIndex, closingChange, sessionPerformance), nil
	}

	// Replace placeholders with actual data
	replacer := strings.NewReplacer(
		"{date}", date,
		"{session_type}", sessionType,
		"{opening_index}", openingIndex,
		"{opening_change}", openingChange,
		"{closing_index}", closingIndex,
		"{closing_change}", closingChange,
		"{session_performance}", sessionPerformance,
	)

	return replacer.Replace(string(content)), nil
}

// convertNumbersToHighlights converts number strings to meaningful sector highlights
func (h *Handler) convertNumbersToHighlights(numberStr string) string {
	// Load highlights mapping from JSON file
	highlightsFile := "highlights_for_prompt.json"
	content, err := os.ReadFile(highlightsFile)
	if err != nil {
		log.Printf("Warning: Could not load highlights mapping: %v", err)
		return numberStr // Return original if can't load mapping
	}

	var highlightsMap map[string][]string
	if err := json.Unmarshal(content, &highlightsMap); err != nil {
		log.Printf("Warning: Could not parse highlights JSON: %v", err)
		return numberStr
	}

	// Split by <br> tags to get two groups
	groups := strings.Split(numberStr, "<br>")

	var highlights []string
	re := regexp.MustCompile(`([+-]?\d+)`)

	// Process each group and take only the FIRST number from each group
	for _, group := range groups {
		if strings.TrimSpace(group) == "" {
			continue
		}

		// Find the first number in this group
		matches := re.FindAllStringSubmatch(group, -1)
		if len(matches) > 0 && len(matches[0]) > 1 {
			originalNumber := matches[0][1]                 // First number with sign
			digit := originalNumber[len(originalNumber)-1:] // Get last digit for mapping

			if phrases, exists := highlightsMap[digit]; exists && len(phrases) > 0 {
				// Randomly select ONE phrase per number
				randomIndex := rand.Intn(len(phrases))
				selectedPhrase := phrases[randomIndex]

				// Format: Only the text, no number display
				highlights = append(highlights, selectedPhrase)
			}
		}
	}

	if len(highlights) > 0 {
		return strings.Join(highlights, "\n\n")
	}

	return numberStr // Fallback to original if no mapping found
}

// callGeminiAI makes a request to Gemini AI API
func (h *Handler) callGeminiAI(prompt string) (string, error) {

	apiKey := h.Config.GeminiAPIKey
	if apiKey == "" {
		log.Printf("GEMINI_API_KEY not set, using mock response")
		return h.generateMockGeminiResponse(prompt), nil
	}

	// The prompt is now pre-formatted with instructions, no need for additional system prompt
	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Role:  "user",
				Parts: []GeminiPart{{Text: prompt}},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	// Make API call with retry logic - using the v1beta gemini-2.5-flash model
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%s", apiKey)

	var resp *http.Response
	var body []byte
	maxRetries := 2

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			waitTime := time.Duration(15+attempt*10) * time.Second // 15s, 25s delays
			log.Printf("Retrying Gemini API call in %v (attempt %d/%d)", waitTime, attempt+1, maxRetries+1)
			time.Sleep(waitTime)
		}

		resp, err = http.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			if attempt == maxRetries {
				log.Printf("Gemini API request failed after %d attempts: %v", maxRetries+1, err)
				return h.generateMockGeminiResponse(prompt), nil
			}
			continue
		}
		defer resp.Body.Close()

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			if attempt == maxRetries {
				log.Printf("Failed to read Gemini API response after %d attempts: %v", maxRetries+1, err)
				return h.generateMockGeminiResponse(prompt), nil
			}
			continue
		}

		// Check for rate limiting (429) or quota exceeded
		if resp.StatusCode == 429 || (resp.StatusCode != http.StatusOK && strings.Contains(string(body), "quota")) {
			if attempt == maxRetries {
				log.Printf("Gemini API quota/rate limit exceeded after %d attempts. Status: %d, Response: %s", maxRetries+1, resp.StatusCode, string(body))
				return h.generateMockGeminiResponse(prompt), nil
			}
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("Gemini API error (attempt %d): Status %d, Response: %s", attempt+1, resp.StatusCode, string(body))
			if attempt == maxRetries {
				return h.generateMockGeminiResponse(prompt), nil
			}
			continue
		}

		// Success - break out of retry loop
		break
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return h.generateMockGeminiResponse(prompt), nil
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

// generateMockGeminiResponse creates a data-driven mock response when API fails
func (h *Handler) generateMockGeminiResponse(prompt string) string {
	// Extract data from prompt to create contextual response
	lines := strings.Split(prompt, "\n")
	var indexValue, highlights string

	for _, line := range lines {
		if strings.Contains(line, "Opening Index:") || strings.Contains(line, "Opening:") {
			indexValue = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.Contains(line, "Highlights:") {
			highlights = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.Contains(line, "Closing:") {
			indexValue = strings.TrimSpace(strings.Split(line, ":")[1])
		}
	}

	if strings.Contains(strings.ToLower(prompt), "takeaway") {
		return `- Market showed mixed performance with selective sector rotation based on fundamental strengths
- Technical analysis indicates consolidation phase with key support and resistance levels being tested
- Trading volume patterns suggest institutional participation remains selective across sectors
- Economic indicators and policy developments continue influencing investor sentiment and positioning
- Risk management remains crucial given current market volatility and global economic uncertainties`
	}

	if highlights != "" {
		// Generate varied content based on session type and data
		if strings.Contains(strings.ToLower(prompt), "morning") && strings.Contains(strings.ToLower(prompt), "opening") {
			if highlights != "" {
				// Parse highlights to create natural narrative
				var themeDescription string
				if strings.Contains(strings.ToLower(highlights), "technology") {
					themeDescription = "Technology sector developments and innovation themes"
				} else if strings.Contains(strings.ToLower(highlights), "banking") || strings.Contains(strings.ToLower(highlights), "financial") {
					themeDescription = "Financial sector dynamics and banking developments"
				} else if strings.Contains(strings.ToLower(highlights), "energy") {
					themeDescription = "Energy sector momentum and commodity-related themes"
				} else if strings.Contains(strings.ToLower(highlights), "foreign") {
					themeDescription = "Foreign investor activity and capital flow patterns"
				} else {
					themeDescription = "Sector-specific developments and thematic investment trends"
				}

				return fmt.Sprintf(`The SET Index opened the morning session at %s, with %s capturing early market attention. Trading patterns show selective institutional interest aligned with these market themes.

Opening momentum reflects measured investor positioning, with participants evaluating both technical levels and fundamental sector developments. The early price action indicates balanced sentiment across different market segments.

Key sector activity demonstrates strategic positioning by institutional investors, particularly in areas showing relative strength. Technical indicators suggest the market is testing important reference levels established in recent sessions.

For the remainder of the morning session, focus should remain on how effectively the market can sustain current levels while managing evolving sector dynamics. Risk management continues to be essential given current market conditions.`, indexValue, themeDescription)
			}
			return fmt.Sprintf(`The morning session opened at %s with measured sentiment reflecting current market conditions. Early trading patterns suggest cautious positioning as participants evaluate overnight developments and regional market cues.

Initial price action indicates balanced institutional participation, with selective interest across different market segments. The opening level establishes important technical reference points for the session ahead.

Market breadth and volume patterns in early trading will be crucial indicators of underlying conviction. Participants are closely watching for follow-through momentum or potential consolidation around current levels.

Trading strategy should emphasize selective opportunities while maintaining disciplined risk management approach given the evolving market environment.`, indexValue)
		}

		if strings.Contains(strings.ToLower(prompt), "afternoon") && strings.Contains(strings.ToLower(prompt), "opening") {
			if highlights != "" {
				// Parse highlights for afternoon context
				var marketPressure string
				if strings.Contains(strings.ToLower(highlights), "outflow") || strings.Contains(strings.ToLower(highlights), "selling") {
					marketPressure = "selling pressure and outflow concerns influencing market sentiment"
				} else if strings.Contains(strings.ToLower(highlights), "profit") {
					marketPressure = "profit-taking activities affecting institutional positioning"
				} else if strings.Contains(strings.ToLower(highlights), "foreign") {
					marketPressure = "foreign investor activity reshaping market dynamics"
				} else {
					marketPressure = "evolving sector themes continuing to guide trading patterns"
				}

				return fmt.Sprintf(`The afternoon session commenced at %s, with %s from earlier developments. Post-lunch trading shows participants reassessing positions based on morning session outcomes and fresh market information.

Afternoon opening levels provide important context for how effectively the market is processing earlier movements. Institutional activity patterns suggest selective approach to current price levels and sector opportunities.

Technical analysis indicates the market is evaluating key support and resistance zones established during morning trading. Volume characteristics and sector rotation patterns will be important indicators for afternoon direction.

Trading focus should emphasize monitoring how well current levels can be sustained while managing evolving market pressures and sector-specific developments.`, indexValue, marketPressure)
			}
			return fmt.Sprintf(`The afternoon session opened at %s, reflecting the market's digestion of morning developments. Post-lunch sentiment appears measured as participants evaluate the sustainability of earlier movements.

Institutional activity shows selective positioning, with afternoon trading patterns often revealing clearer directional biases. The opening level serves as a key reference point for potential afternoon trends.

Technical analysis suggests the market is in a critical phase, testing important support and resistance levels established during morning trading. Volume characteristics will be important for confirming any directional moves.

Afternoon strategy should focus on managing positions established earlier while remaining alert to evolving sector themes and institutional flows.`, indexValue)
		}

		// For close analysis - different content based on session performance
		if strings.Contains(strings.ToLower(prompt), "closing") || strings.Contains(strings.ToLower(prompt), "close") {
			// Extract opening and closing values to provide comparative analysis
			sessionType := "morning"
			if strings.Contains(strings.ToLower(prompt), "afternoon") {
				sessionType = "afternoon"
			}

			return fmt.Sprintf(`The %s session concluded at %s, providing important insights into current market sentiment and institutional positioning. Session performance reflects the ongoing balance between buyer and seller conviction across different market segments.

Closing patterns reveal how effectively the market absorbed trading volumes and sector-specific developments throughout the session. The final level establishes crucial technical reference points for subsequent trading periods.

Institutional participation patterns during the session indicate selective approach to current market conditions, with focus on fundamental value and technical positioning. End-of-session activity often provides insights into positioning ahead of upcoming market catalysts.

The session's price action contributes to broader market themes and technical chart patterns that will influence near-term trading strategies and investment decisions.`, sessionType, indexValue)
		}

		return fmt.Sprintf(`Market analysis for the current session shows index positioning at %s, reflecting evolving investor sentiment and institutional flows. Current price levels indicate ongoing evaluation of economic fundamentals and market technical factors.

Session dynamics reveal balanced participation with selective sector emphasis based on fundamental developments and valuation considerations. Market participants continue navigating between growth opportunities and risk management priorities.

Technical indicators suggest the market remains in a transitional phase, with key support and resistance levels being actively tested. Volume patterns indicate measured institutional involvement across different market segments.

Investment approach should emphasize selective stock picking based on individual company fundamentals while maintaining appropriate risk management protocols given current market conditions.`, indexValue)
	}

	return fmt.Sprintf(`Market analysis for the session shows the index at %s, indicating current investor sentiment and market dynamics. Trading patterns reflect a balanced approach with selective sector participation.

Technical analysis suggests the market is in a consolidation phase, with participants carefully evaluating economic developments and corporate fundamentals. Volume patterns indicate measured institutional involvement.

The current price action reflects cautious positioning as market participants assess various economic indicators and policy developments. Sector rotation continues based on fundamental outlook and valuation considerations.

Investment strategy should emphasize risk management and selective opportunities in sectors with strong fundamentals and favorable technical setups.`, indexValue)
}

// API Request/Response structures for market data endpoints
type MarketSession struct {
	Index      float64 `json:"index"`
	Change     float64 `json:"change"`
	Highlights string  `json:"highlights,omitempty"`
}

// MarketCloseSession for close data (no highlights needed)
type MarketCloseSession struct {
	Index  float64 `json:"index"`
	Change float64 `json:"change"`
}

type MarketDataAnalysisRequest struct {
	Date          string         `json:"date"`
	MorningOpen   *MarketSession `json:"morning_open,omitempty"`
	AfternoonOpen *MarketSession `json:"afternoon_open,omitempty"`
}

type MarketDataCloseRequest struct {
	Date           string         `json:"date"`
	MorningClose   *MarketSession `json:"morning_close,omitempty"`
	AfternoonClose *MarketSession `json:"afternoon_close,omitempty"`
} // MarketDataAnalysisHandler processes market analysis data and generates content with Gemini AI
func (h *Handler) MarketDataAnalysisHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MarketDataAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("📊 Market Analysis Request for %s", req.Date)

	// Generate analysis content with Gemini AI
	analysisContent := h.generateAnalysisWithGemini(req)

	// Save to file and database
	if err := h.saveAnalysisToFile(req.Date, analysisContent); err != nil {
		log.Printf("Error saving analysis to file: %v", err)
		http.Error(w, "Error saving analysis", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Analysis generated and saved successfully",
		"date":    req.Date,
	})
}

// MarketDataCloseHandler processes market close data and generates summary with Gemini AI
func (h *Handler) MarketDataCloseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MarketDataCloseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("📊 Market Close Request for %s", req.Date)

	// Generate summary content with Gemini AI
	summaryContent := h.generateSummaryWithGemini(req)

	// Save to file and database
	if err := h.saveSummaryToFile(req.Date, summaryContent); err != nil {
		log.Printf("Error saving summary to file: %v", err)
		http.Error(w, "Error saving summary", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Summary generated and saved successfully",
		"date":    req.Date,
	})
}

// generateAnalysisWithGemini integrates with Gemini AI to generate market analysis
func (h *Handler) generateAnalysisWithGemini(req MarketDataAnalysisRequest) string {
	sessionType := "morning"
	session := req.MorningOpen
	if req.AfternoonOpen != nil {
		sessionType = "afternoon"
		session = req.AfternoonOpen
	}

	// Check if session data is available
	if session == nil {
		log.Printf("Error: No session data provided for %s", req.Date)
		return "No market data available for analysis."
	}

	// Convert number highlights to meaningful sector text for the AI prompt
	narrativeHighlight := h.convertNumbersToHighlights(session.Highlights)

	// Use human-style prompt for more engaging analysis
	prompt, err := h.loadHumanStylePrompt(
		req.Date,
		sessionType,
		"opening",
		fmt.Sprintf("%.2f", session.Index),
		fmt.Sprintf("%+.2f", session.Change),
		narrativeHighlight,
	)
	if err != nil {
		log.Printf("Error loading prompt template: %v", err)
		return "Market analysis temporarily unavailable."
	}

	// Get market analysis
	aiAnalysis, err := h.callGeminiAI(prompt)
	if err != nil {
		log.Printf("Error generating market analysis: %v", err)
		aiAnalysis = "Market analysis indicates mixed sentiment with selective sector rotation and cautious investor positioning."
	}

	// Send Telegram notification after successful Gemini analysis
	openIndex := fmt.Sprintf("%.2f", session.Index)
	change := fmt.Sprintf("%+.2f", session.Change)
	sessionName := fmt.Sprintf("%s Session Open", strings.Title(sessionType))

	if err := h.TelegramService.SendMarketUpdate(sessionName, openIndex, change, req.Date); err != nil {
		log.Printf("⚠️  Failed to send Telegram notification: %v", err)
	}

	return fmt.Sprintf(`
## %s Session

### Open Set
* Open Index: %.2f (%+.2f)
* Highlights: %s <br><br> %s

### Open Analysis
%s

`, strings.Title(sessionType), session.Index, session.Change, narrativeHighlight, session.Highlights, aiAnalysis)
} // parseSessionOpeningData reads existing markdown file and extracts opening data for specific session
func (h *Handler) parseSessionOpeningData(date, sessionType string) (*MarketSession, error) {
	filename := fmt.Sprintf("%s/%s.md", h.ArticlesDir, date)

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %v", filename, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var targetSection string
	if sessionType == "morning" {
		targetSection = "## Morning Session"
	} else {
		targetSection = "## Afternoon Session"
	}

	inTargetSection := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Check for session headers
		if strings.Contains(line, targetSection) {
			inTargetSection = true
			continue
		}

		// Stop if we hit another level 2 section (##), but not level 3 (###)
		if inTargetSection && strings.HasPrefix(line, "##") && !strings.HasPrefix(line, "###") && !strings.Contains(line, targetSection) {
			break
		}

		// Look for index pattern: "* Index: 1295.80 (+5.15)" or "* Open Index: 1295.80 (+5.15)"
		// But exclude "Close Index:" which is for close data, not open data
		if inTargetSection && (strings.Contains(line, "Index:") && !strings.Contains(line, "Close Index:")) {
			// Extract index and change using regex
			re := regexp.MustCompile(`(\d+\.?\d*)\s*\(([+-]?\d+\.?\d*)\)`)
			matches := re.FindStringSubmatch(line)

			if len(matches) >= 3 {
				indexVal, err1 := strconv.ParseFloat(matches[1], 64)
				changeVal, err2 := strconv.ParseFloat(matches[2], 64)

				if err1 == nil && err2 == nil {
					return &MarketSession{
						Index:  indexVal,
						Change: changeVal,
					}, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("could not find opening data for %s session in file %s", sessionType, filename)
}

// generateSummaryWithGemini generates comprehensive market summary for both sessions
func (h *Handler) generateSummaryWithGemini(req MarketDataCloseRequest) string {
	var content strings.Builder

	// Handle Morning Session Close
	if req.MorningClose != nil {
		morningContent := h.generateSessionClose("morning", req.Date, req.MorningClose)
		content.WriteString(morningContent)
	}

	// Handle Afternoon Session Close
	if req.AfternoonClose != nil {
		afternoonContent := h.generateSessionClose("afternoon", req.Date, req.AfternoonClose)
		content.WriteString(afternoonContent)
	}

	if req.MorningClose == nil && req.AfternoonClose == nil {
		content.WriteString("### Error\nNo closing data provided for any session.\n\n")
	}

	content.WriteString("---\n")
	return content.String()
}

// generateSessionClose generates closing data for a specific session
func (h *Handler) generateSessionClose(sessionType, date string, closeData *MarketSession) string {
	// Get corresponding opening data from file
	openData, err := h.parseSessionOpeningData(date, sessionType)
	if err != nil {
		log.Printf("Warning: Could not parse %s opening data: %v", sessionType, err)
		return fmt.Sprintf(`
### Close Set
* Close Index: %.2f (%+.2f)

### Close Summary
<p>%s session closed at %.2f (%+.2f). Analysis pending opening data confirmation.</p>

`, closeData.Index, closeData.Change, strings.Title(sessionType), closeData.Index, closeData.Change)
	}

	// Calculate session performance
	sessionDiff := closeData.Index - openData.Index
	sessionPerf := "gained"
	if sessionDiff < 0 {
		sessionPerf = "lost"
		sessionDiff = -sessionDiff
	}

	// Use human-style closing prompt for more engaging session summary
	prompt, err := h.loadHumanStyleClosePrompt(
		date,
		sessionType,
		fmt.Sprintf("%.2f", openData.Index),
		fmt.Sprintf("%+.2f", openData.Change),
		fmt.Sprintf("%.2f", closeData.Index),
		fmt.Sprintf("%+.2f", closeData.Change),
		fmt.Sprintf("%s %.2f points", sessionPerf, sessionDiff),
	)
	if err != nil {
		log.Printf("Error loading closing prompt template: %v", err)
		return fmt.Sprintf(`
### Close Set
* Close Index: %.2f (%+.2f)

### Close Summary
<p>Session analysis temporarily unavailable.</p>

`, closeData.Index, closeData.Change)
	}

	// Get AI-generated comparative analysis
	aiAnalysis, err := h.callGeminiAI(prompt)
	if err != nil {
		log.Printf("Error calling Gemini AI: %v", err)
		aiAnalysis = "Professional market analysis temporarily unavailable. Session data suggests mixed market conditions with intraday volatility."
	}

	closeSection := fmt.Sprintf(`
### Close Set
* Close Index: %.2f (%+.2f)

### Close Summary
<p>%s session closed at %.2f (%+.2f) after %s %.2f points from %.2f opening. %s</p>

`, closeData.Index, closeData.Change, strings.Title(sessionType), closeData.Index, closeData.Change, sessionPerf, sessionDiff, openData.Index, aiAnalysis)

	// If this is afternoon close, add Key Takeaways
	if sessionType == "afternoon" {
		keyTakeaways := h.generateKeyTakeaways(date, closeData.Index, closeData.Change)
		closeSection += keyTakeaways
	}

	return closeSection
}

// generateKeyTakeaways generates daily key takeaways for afternoon close
func (h *Handler) generateKeyTakeaways(date string, finalIndex, finalChange float64) string {
	// Create comprehensive prompt for daily summary
	prompt := fmt.Sprintf(`Generate key takeaways for Thai stock market trading day %s:

Final Index: %.2f with total daily change of %.2f points

Please provide 3-5 key takeaways that summarize the entire day's trading including:
1. Overall market performance and sentiment
2. Key sector winners and losers
3. Notable market trends and patterns observed
4. Significant institutional or foreign investor activity
5. Important market-moving events or catalysts
6. Technical analysis insights for tomorrow's outlook

IMPORTANT: Write the analysis in ENGLISH language only.
IMPORTANT: Format response as bullet points starting with dash (-) character.
Example format:
- First key takeaway about market performance
- Second takeaway about sector rotation
- Third takeaway about institutional activity

Each takeaway should be concise but informative, focusing on actionable insights for Thai stock market investors.`,
		date, finalIndex, finalChange)

	// Get AI-generated key takeaways
	aiTakeaways, err := h.callGeminiAI(prompt)
	if err != nil {
		log.Printf("Error generating key takeaways: %v", err)
		aiTakeaways = "- Market performance reflected mixed sentiment with selective sector rotation\n- Trading patterns indicated institutional positioning for upcoming developments\n- Technical indicators suggest continued monitoring of key support and resistance levels"
	}

	return fmt.Sprintf(`
## Key Takeaways

%s

`, aiTakeaways)
}

// saveAnalysisToFile saves generated analysis to markdown file and creates database entry
func (h *Handler) saveAnalysisToFile(date, content string) error {
	filename := fmt.Sprintf("%s/%s.md", h.ArticlesDir, date)

	// Check if file exists
	var err error
	var isNewFile bool
	if _, statErr := os.Stat(filename); os.IsNotExist(statErr) {
		isNewFile = true
		// New file - add main title first
		parsedDate, _ := time.Parse("2006-01-02", date)
		finalContent := fmt.Sprintf("# Stock Market Analysis - %s\n\n%s", parsedDate.Format("2 January 2006"), content)

		// Write new file
		err = os.WriteFile(filename, []byte(finalContent), 0644)
	} else {
		// File exists - append content
		file, openErr := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
		if openErr != nil {
			return fmt.Errorf("failed to open file: %v", openErr)
		}
		defer file.Close()
		_, err = file.WriteString(content)
	}

	if err != nil {
		return fmt.Errorf("failed to write content: %v", err)
	}

	// Create database entry for new files
	if isNewFile {
		parsedDate, _ := time.Parse("2006-01-02", date)
		title := fmt.Sprintf("Stock Market Analysis - %s", parsedDate.Format("2 January 2006"))
		summary := "Thai stock market analysis including SET index movements, sector highlights, and key insights."

		// Check if article already exists in database
		exists, err := database.ArticleExists(date)
		if err != nil {
			log.Printf("Error checking if article exists in database: %v", err)
		} else if !exists {
			// Create database entry
			if err := database.CreateArticle(date, title, summary, ""); err != nil {
				log.Printf("Error creating database entry for %s: %v", date, err)
			} else {
				log.Printf("📊 Database entry created for %s", date)
			}
		}
	}

	log.Printf("📝 Analysis saved to %s", filename)
	return nil
}

// saveSummaryToFile saves generated summary to markdown file and creates database entry
func (h *Handler) saveSummaryToFile(date, content string) error {
	filename := fmt.Sprintf("%s/%s.md", h.ArticlesDir, date)

	// Check if file exists before opening
	var isNewFile bool
	if _, statErr := os.Stat(filename); os.IsNotExist(statErr) {
		isNewFile = true
	}

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write content: %v", err)
	}

	// Create database entry for new files
	if isNewFile {
		parsedDate, _ := time.Parse("2006-01-02", date)
		title := fmt.Sprintf("Stock Market Analysis - %s", parsedDate.Format("2 January 2006"))
		summary := "Thai stock market analysis including SET index movements, sector highlights, and key insights."

		// Check if article already exists in database
		exists, err := database.ArticleExists(date)
		if err != nil {
			log.Printf("Error checking if article exists in database: %v", err)
		} else if !exists {
			// Create database entry
			if err := database.CreateArticle(date, title, summary, ""); err != nil {
				log.Printf("Error creating database entry for %s: %v", date, err)
			} else {
				log.Printf("📊 Database entry created for %s", date)
			}
		}
	}

	log.Printf("📝 Summary saved to %s", filename)
	return nil
}
