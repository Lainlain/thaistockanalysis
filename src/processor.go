package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Session close processor following ThaiStockAnalysis performance patterns
type SessionProcessor struct {
	openData  SessionData
	closeData SessionData
	session   string // "morning" or "afternoon"
}

type SessionData struct {
	Index  float64 `json:"index"`
	Change float64 `json:"change"`
}

// Create close summary prompt file
func createCloseSummaryPrompt() error {
	promptContent := `create short close summary in one mini paragraph for $session session. Open was $open_set $open_change with sectors $open_sectors. Close at $close_set $close_change. Write concise analysis for adsense approve content as middle people understanding`

	return os.WriteFile("getanalysis_prompt_close", []byte(promptContent), 0644)
}

// Process close data and generate summary using open and close data
func processCloseDataWithSummary(openIndex, openChange float64, openHighlights string,
	closeIndex, closeChange float64, sessionType string) (string, error) {

	// Ensure close summary prompt exists
	if _, err := os.Stat("getanalysis_prompt_close"); os.IsNotExist(err) {
		if err := createCloseSummaryPrompt(); err != nil {
			debugLog("Error creating close summary prompt: %v", err)
			return "", err
		}
		debugLog("✅ Created close summary prompt file")
	}

	// Read close summary prompt template
	promptTemplate, err := os.ReadFile("getanalysis_prompt_close")
	if err != nil {
		debugLog("Error reading close summary prompt: %v", err)
		return "", err
	}

	// Replace placeholders for both open and close data - use direct values
	prompt := string(promptTemplate)
	prompt = strings.ReplaceAll(prompt, "$session", sessionType)
	prompt = strings.ReplaceAll(prompt, "$open_set", fmt.Sprintf("%.2f", openIndex))
	prompt = strings.ReplaceAll(prompt, "$open_change", formatChangeValue(openChange)) // Direct value
	prompt = strings.ReplaceAll(prompt, "$open_sectors", openHighlights)
	prompt = strings.ReplaceAll(prompt, "$close_set", fmt.Sprintf("%.2f", closeIndex))
	prompt = strings.ReplaceAll(prompt, "$close_change", formatChangeValue(closeChange)) // Direct value

	debugLog("📝 Final Gemini close summary prompt: %s", prompt)

	// Generate summary using Gemini API
	summary, err := callGeminiAPI(prompt)
	if err != nil {
		debugLog("Gemini API error for %s close summary: %v", sessionType, err)
		// Enhanced fallback summary with professional analysis
		changeDirection := "gained"
		if closeChange < 0 {
			changeDirection = "declined"
		}

		sessionDirection := "consolidation"
		sessionVolatility := "stable"
		if openIndex != closeIndex {
			if closeIndex > openIndex {
				sessionDirection = "upward momentum"
			} else {
				sessionDirection = "downward pressure"
			}

			volatilityRange := openIndex - closeIndex
			if volatilityRange < 0 {
				volatilityRange = -volatilityRange
			}
			if volatilityRange > 10 {
				sessionVolatility = "volatile"
			} else if volatilityRange > 5 {
				sessionVolatility = "moderately active"
			}
		}

		summary = fmt.Sprintf("%s session concluded at %.2f, having %s %s points from the opening level of %.2f (%s). The session demonstrated %s with %s trading patterns, reflecting investor sentiment and market dynamics. Key sector movements including %s shaped the session's trajectory, while institutional activity and market breadth provided insight into underlying market strength and participant confidence throughout the %s trading period.",
			sessionType, closeIndex, changeDirection, formatChangeValue(closeChange),
			openIndex, formatChangeValue(openChange), sessionDirection, sessionVolatility,
			openHighlights, strings.ToLower(sessionType))
	}

	debugLog("✅ Generated %s close summary: %s", sessionType, summary)
	return summary, nil
}

// Update the apiMarketDataWithCloseAnalysisHandler function
// ...existing code...
func apiMarketDataWithCloseAnalysisHandler(w http.ResponseWriter, r *http.Request) {
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

	// CRITICAL FIX: Load existing content FIRST to preserve all data
	var existingContent string
	err = db.QueryRow("SELECT content FROM articles WHERE slug = ?", requestData.Date).Scan(&existingContent)

	var existingData StockData
	if err == nil && existingContent != "" {
		existingData = parseMarkdownContentForAdmin(existingContent)
		debugLog("API: Found existing content for %s, preserving ALL data", requestData.Date)
	} else {
		// Only create new if no existing data found
		existingData = StockData{
			CurrentDate:  time.Now().Format("2 January 2006"),
			KeyTakeaways: []string{},
		}
		debugLog("API: No existing content for %s, creating new", requestData.Date)
	}

	var sessionType string
	var notifications []string

	// CRITICAL FIX: Process ANY open data from request to preserve/update it
	if requestData.MorningOpen.Index > 0 {
		sessionType = "Morning"

		// Generate analysis for morning open data
		analysis, err := generateGeminiAnalysis(requestData.MorningOpen.Index, requestData.MorningOpen.Change, requestData.MorningOpen.Highlights)
		if err != nil {
			debugLog("Gemini API error for morning open: %v", err)
			analysis = fmt.Sprintf("Market opened at %.2f with change of %.2f. %s sector movements observed in morning session.",
				requestData.MorningOpen.Index, requestData.MorningOpen.Change, requestData.MorningOpen.Highlights)
		}

		// PRESERVE/UPDATE morning open data
		existingData.MorningOpenIndex = requestData.MorningOpen.Index
		existingData.MorningOpenChange = requestData.MorningOpen.Change
		if requestData.MorningOpen.Highlights != "" {
			existingData.MorningOpenHighlights = requestData.MorningOpen.Highlights
		}
		existingData.MorningOpenAnalysis = template.HTML(analysis)

		debugLog("API: Updated/preserved morning open data with Gemini analysis")

		// Send Telegram notification for open
		go sendTelegramNotification(requestData.MorningOpen.Index, requestData.MorningOpen.Change,
			requestData.MorningOpen.Highlights, analysis, "Morning Open", getBaseURL(r))
	}

	if requestData.AfternoonOpen.Index > 0 {
		sessionType = "Afternoon"

		// Generate analysis for afternoon open data
		analysis, err := generateGeminiAnalysis(requestData.AfternoonOpen.Index, requestData.AfternoonOpen.Change, requestData.AfternoonOpen.Highlights)
		if err != nil {
			debugLog("Gemini API error for afternoon open: %v", err)
			analysis = fmt.Sprintf("Afternoon session opened at %.2f with change of %.2f. %s sector movements observed.",
				requestData.AfternoonOpen.Index, requestData.AfternoonOpen.Change, requestData.AfternoonOpen.Highlights)
		}

		// PRESERVE/UPDATE afternoon open data
		existingData.AfternoonOpenIndex = requestData.AfternoonOpen.Index
		existingData.AfternoonOpenChange = requestData.AfternoonOpen.Change
		if requestData.AfternoonOpen.Highlights != "" {
			existingData.AfternoonOpenHighlights = requestData.AfternoonOpen.Highlights
		}
		existingData.AfternoonOpenAnalysis = template.HTML(analysis)

		debugLog("API: Updated/preserved afternoon open data with Gemini analysis")

		// Send Telegram notification for open
		go sendTelegramNotification(requestData.AfternoonOpen.Index, requestData.AfternoonOpen.Change,
			requestData.AfternoonOpen.Highlights, analysis, "Afternoon Open", getBaseURL(r))
	}

	// Process morning close data with open data context - PRESERVE existing open data
	if requestData.MorningClose.Index > 0 {
		// Verify we have existing open data before processing close
		if existingData.MorningOpenIndex > 0 {
			sessionType = "Morning"

			// Generate close summary using PRESERVED open data and new close data
			summary, err := processCloseDataWithSummary(
				existingData.MorningOpenIndex, existingData.MorningOpenChange, existingData.MorningOpenHighlights,
				requestData.MorningClose.Index, requestData.MorningClose.Change,
				"Morning")

			if err != nil {
				debugLog("Error generating morning close summary: %v", err)
				summary = fmt.Sprintf("Morning session closed at %.2f with change of %.2f.",
					requestData.MorningClose.Index, requestData.MorningClose.Change)
			}

			// UPDATE close data while PRESERVING open data
			existingData.MorningCloseIndex = requestData.MorningClose.Index
			existingData.MorningCloseChange = requestData.MorningClose.Change
			existingData.MorningCloseSummary = template.HTML("<p>" + summary + "</p>")

			// CRITICAL: All existing open data is PRESERVED automatically
			// existingData.MorningOpenIndex - PRESERVED
			// existingData.MorningOpenChange - PRESERVED
			// existingData.MorningOpenHighlights - PRESERVED
			// existingData.MorningOpenAnalysis - PRESERVED

			debugLog("API: Processed morning close with summary, PRESERVED open data")
			notifications = append(notifications, "Morning Close")

			// Send Telegram notification for close
			go sendTelegramNotificationForClose(
				existingData.MorningOpenIndex, existingData.MorningOpenChange,
				requestData.MorningClose.Index, requestData.MorningClose.Change,
				summary, "Morning Close")
		} else {
			debugLog("WARNING: No existing morning open data found - cannot generate close summary")
			response := MarketDataResponse{
				Success: false,
				Error:   "No existing morning open data found. Please add morning open data first.",
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	// Process afternoon close data with open data context - PRESERVE existing open data
	if requestData.AfternoonClose.Index > 0 {
		if existingData.AfternoonOpenIndex > 0 {
			sessionType = "Afternoon"

			// Generate close summary using PRESERVED open data and new close data
			summary, err := processCloseDataWithSummary(
				existingData.AfternoonOpenIndex, existingData.AfternoonOpenChange, existingData.AfternoonOpenHighlights,
				requestData.AfternoonClose.Index, requestData.AfternoonClose.Change,
				"Afternoon")

			if err != nil {
				debugLog("Error generating afternoon close summary: %v", err)
				summary = fmt.Sprintf("Afternoon session closed at %.2f with change of %.2f.",
					requestData.AfternoonClose.Index, requestData.AfternoonClose.Change)
			}

			// UPDATE close data while PRESERVING open data
			existingData.AfternoonCloseIndex = requestData.AfternoonClose.Index
			existingData.AfternoonCloseChange = requestData.AfternoonClose.Change
			existingData.AfternoonCloseSummary = template.HTML("<p>" + summary + "</p>")

			debugLog("API: Processed afternoon close with summary, PRESERVED open data")
			notifications = append(notifications, "Afternoon Close")

			// Send Telegram notification for close
			go sendTelegramNotificationForClose(
				existingData.AfternoonOpenIndex, existingData.AfternoonOpenChange,
				requestData.AfternoonClose.Index, requestData.AfternoonClose.Change,
				summary, "Afternoon Close")
		} else {
			debugLog("WARNING: No existing afternoon open data found - cannot generate close summary")
			response := MarketDataResponse{
				Success: false,
				Error:   "No existing afternoon open data found. Please add afternoon open data first.",
			}
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	// Generate complete markdown content with ALL preserved data
	processedContent := generateEnhancedMarkdownFromData(existingData)

	// Save to database and files following ThaiStockAnalysis dual storage pattern
	parsedDate, _ := time.Parse("2006-01-02", requestData.Date)
	title := fmt.Sprintf("Stock Market Analysis - %s", parsedDate.Format("2 January 2006"))
	summary := generateSummaryFromAPI(requestData)

	var existingID int
	err = db.QueryRow("SELECT id FROM articles WHERE slug = ?", requestData.Date).Scan(&existingID)

	var responseMessage string
	if len(notifications) > 0 {
		responseMessage = fmt.Sprintf("Article updated with %s summaries and Telegram notifications sent", strings.Join(notifications, ", "))
	} else {
		responseMessage = fmt.Sprintf("Article updated with %s analysis", sessionType)
	}

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

		// Write markdown file for dual storage following ThaiStockAnalysis patterns
		filename := fmt.Sprintf("articles/%s.md", requestData.Date)
		os.WriteFile(filename, []byte(processedContent), 0644)

		// Clear caches for immediate updates following ThaiStockAnalysis performance patterns
		clearMarkdownCache(filename)
		clearTemplateCache()

		response := MarketDataResponse{
			Success: true,
			Message: responseMessage,
			Data: struct {
				ArticleID int    `json:"article_id,omitempty"`
				Slug      string `json:"slug,omitempty"`
				URL       string `json:"url,omitempty"`
			}{
				ArticleID: int(newID),
				Slug:      requestData.Date,
				URL:       fmt.Sprintf("http://localhost:7777/articles/%s", requestData.Date),
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
			Message: responseMessage,
			Data: struct {
				ArticleID int    `json:"article_id,omitempty"`
				Slug      string `json:"slug,omitempty"`
				URL       string `json:"url,omitempty"`
			}{
				ArticleID: existingID,
				Slug:      requestData.Date,
				URL:       fmt.Sprintf("http://localhost:7777/articles/%s", requestData.Date),
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// ...existing code...
// ...existing code...

// Helper function to call Gemini API for close summaries
func callGeminiAPI(prompt string) (string, error) {
	// Prepare Gemini API request following ThaiStockAnalysis performance patterns
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
		return "", fmt.Errorf("error marshaling Gemini request: %v", err)
	}

	// Use gemini-2.5-pro with optimized timeout
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-pro:generateContent?key=%s", GEMINI_API_KEY)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("error calling Gemini API: %v", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading Gemini response: %v", err)
	}

	var geminiResponse GeminiResponse
	err = json.Unmarshal(responseBody, &geminiResponse)
	if err != nil {
		return "", fmt.Errorf("error parsing Gemini response: %v", err)
	}

	if geminiResponse.Error != nil {
		return "", fmt.Errorf("Gemini API error: %s", geminiResponse.Error.Message)
	}

	// Extract generated text
	if len(geminiResponse.Candidates) > 0 && len(geminiResponse.Candidates[0].Content.Parts) > 0 {
		analysis := geminiResponse.Candidates[0].Content.Parts[0].Text
		return strings.TrimSpace(analysis), nil
	}

	return "", fmt.Errorf("no analysis generated by Gemini")
}

// Enhanced Telegram notification specifically for close sessions
func sendTelegramNotificationForClose(openIndex, openChange, closeIndex, closeChange float64, summary, sessionType string) {
	if TELEGRAM_BOT_TOKEN == "YOUR_TELEGRAM_BOT_TOKEN" || TELEGRAM_BOT_TOKEN == "YOUR_ACTUAL_TELEGRAM_BOT_TOKEN_HERE" {
		debugLog("Telegram bot token not configured - skipping notification")
		return
	}

	// Use direct change values without modification
	openChangeStr := formatChangeValue(openChange)
	closeChangeStr := formatChangeValue(closeChange)

	var emoji string
	var timeInfo string

	if strings.Contains(sessionType, "Morning") {
		emoji = "🌅"
		timeInfo = "12:30 PM"
	} else {
		emoji = "🌆"
		timeInfo = "5:00 PM"
	}

	// Truncate summary for Telegram but keep it readable
	truncatedSummary := summary
	if len(summary) > 250 {
		words := strings.Fields(summary)
		truncated := ""
		for _, word := range words {
			if len(truncated+" "+word) > 230 {
				break
			}
			if truncated == "" {
				truncated = word
			} else {
				truncated += " " + word
			}
		}
		truncatedSummary = truncated + "..."
	}

	// Create Telegram message for close session with direct change values
	message := fmt.Sprintf("%s Thai Stock Market - %s\n\n"+
		"📊 Open: %.2f (%s) → Close: %.2f (%s)\n"+
		"⏰ Close Time: %s\n\n"+
		"📝 Session Summary:\n%s\n\n"+
		"🔗 Full analysis: http://localhost:7777",
		emoji, sessionType, openIndex, openChangeStr, closeIndex, closeChangeStr,
		timeInfo, truncatedSummary)

	telegramMsg := TelegramMessage{
		ChatID:    TELEGRAM_CHAT_ID,
		Text:      message,
		ParseMode: "",
	}

	msgBody, err := json.Marshal(telegramMsg)
	if err != nil {
		debugLog("Error marshaling Telegram close message: %v", err)
		return
	}

	telegramURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", TELEGRAM_BOT_TOKEN)

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Post(telegramURL, "application/json", bytes.NewBuffer(msgBody))
	if err != nil {
		debugLog("Error sending Telegram close message: %v", err)
		return
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		debugLog("Error reading Telegram close response: %v", err)
		return
	}

	var telegramResponse TelegramResponse
	err = json.Unmarshal(responseBody, &telegramResponse)
	if err != nil {
		debugLog("Error parsing Telegram close response: %v", err)
		return
	}

	if telegramResponse.Ok {
		debugLog("✅ Telegram close notification sent successfully for %s: Message ID %d", sessionType, telegramResponse.Result.MessageID)
	} else {
		debugLog("❌ Telegram close API error (Code: %d): %s", telegramResponse.ErrorCode, telegramResponse.Description)
	}
}
