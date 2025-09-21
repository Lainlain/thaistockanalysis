package main

import (
	"database/sql"
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
const DEBUG_MODE = false

func debugLog(format string, args ...interface{}) {
	if DEBUG_MODE {
		log.Printf("DEBUG: "+format, args...)
	}
}

type StockData struct {
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
	CurrentDate string
	Articles    []DBArticle
	Success     string
	Error       string
}

type AdminArticleFormData struct {
	CurrentDate string
	Article     DBArticle
	Error       string
	IsEdit      bool
	Action      string

	// Fields for structured markdown content
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

// Fast template cache with mutex protection
func getTemplate(name string, files ...string) (*template.Template, error) {
	templateMutex.RLock()
	if tmpl, exists := templateCache[name]; exists {
		templateMutex.RUnlock()
		return tmpl, nil
	}
	templateMutex.RUnlock()

	templateMutex.Lock()
	defer templateMutex.Unlock()

	// Double-check pattern
	if tmpl, exists := templateCache[name]; exists {
		return tmpl, nil
	}

	// Create template with functions
	tmpl := template.New(name).Funcs(template.FuncMap{
		"printf": fmt.Sprintf,
		"html":   func(s string) template.HTML { return template.HTML(s) },
		"add":    func(a, b int) int { return a + b },
		"markdownToHTML": func(s string) template.HTML {
			if s == "" {
				return template.HTML("")
			}
			htmlBytes := markdown.ToHTML([]byte(s), nil, nil)
			return template.HTML(htmlBytes)
		},
	})

	var err error
	tmpl, err = tmpl.ParseFiles(files...)
	if err != nil {
		return nil, err
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
						data.MorningOpenAnalysis = template.HTML(markdown.ToHTML([]byte(analysisContent), nil, nil))
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
						data.MorningCloseSummary = template.HTML(markdown.ToHTML([]byte(summaryContent), nil, nil))
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
						data.AfternoonOpenAnalysis = template.HTML(markdown.ToHTML([]byte(analysisContent), nil, nil))
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
						data.AfternoonCloseSummary = template.HTML(markdown.ToHTML([]byte(summaryContent), nil, nil))
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
		return strings.TrimSpace(line[idx+11:])
	}
	return ""
}

// Enhanced index handler - loads data from markdown files
func indexHandler(w http.ResponseWriter, r *http.Request) {
	// Get articles from database
	rows, err := db.Query("SELECT slug, title, summary, created_at FROM articles ORDER BY created_at DESC LIMIT 20")
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	defer rows.Close()

	var previews []ArticlePreview
	for rows.Next() {
		var slug, title, createdAt string
		var summary sql.NullString

		err := rows.Scan(&slug, &title, &summary, &createdAt)
		if err != nil {
			continue
		}

		// Try to load actual data from markdown file
		markdownPath := fmt.Sprintf("articles/%s.md", slug)
		var setIndex string = "--"
		var change float64 = 0.0
		var shortSummary string = summary.String

		// Parse markdown file to get real data
		if stockData, err := getCachedStockData(markdownPath); err == nil {
			// Use the most recent close index available
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

			// Create a short summary from highlights and takeaways
			if len(stockData.KeyTakeaways) > 0 {
				shortSummary = stockData.KeyTakeaways[0]
			} else if stockData.MorningOpenHighlights != "" {
				shortSummary = stockData.MorningOpenHighlights
			} else if stockData.AfternoonOpenHighlights != "" {
				shortSummary = stockData.AfternoonOpenHighlights
			}
		}

		preview := ArticlePreview{
			Title:        title,
			Date:         createdAt,
			SetIndex:     setIndex,
			Change:       change,
			ShortSummary: shortSummary,
			Summary:      shortSummary,
			Slug:         slug,
			URL:          fmt.Sprintf("/articles/%s", slug),
		}
		previews = append(previews, preview)
	}

	data := IndexPageData{
		CurrentDate: time.Now().Format("2 January 2006"),
		Articles:    previews,
	}

	// Use cached templates
	tmpl, err := getTemplate("index", "src/templates/base.gohtml", "src/templates/index.gohtml")
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	tmpl.ExecuteTemplate(w, "base.gohtml", data)
}

// Fast article handler with caching
func articleHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")

	if len(parts) < 3 || parts[2] == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	slug := parts[2]

	// Fast database lookup
	var dbArticle DBArticle
	err := db.QueryRow("SELECT id, slug, title, summary, content, created_at FROM articles WHERE slug = ?", slug).Scan(
		&dbArticle.ID, &dbArticle.Slug, &dbArticle.Title, &dbArticle.Summary, &dbArticle.Content, &dbArticle.CreatedAt)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Fast default data
	stockData := StockData{
		CurrentDate:  time.Now().Format("2 January 2006"),
		KeyTakeaways: []string{},
	}

	// Use cached markdown parsing
	markdownPath := fmt.Sprintf("articles/%s.md", slug)
	if parsedData, err := getCachedStockData(markdownPath); err == nil {
		stockData = parsedData
	}

	data := struct {
		Title     string
		Slug      string
		Summary   string
		CreatedAt string
		StockData
	}{
		Title:     dbArticle.Title,
		Slug:      dbArticle.Slug,
		Summary:   dbArticle.Summary.String,
		CreatedAt: dbArticle.CreatedAt,
		StockData: stockData,
	}

	// Use cached templates
	tmpl, err := getTemplate("article", "src/templates/base.gohtml", "src/templates/article.gohtml")
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	tmpl.ExecuteTemplate(w, "base.gohtml", data)
}

// Fast admin handlers
func adminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	success := r.URL.Query().Get("success")
	errorMsg := r.URL.Query().Get("error")

	rows, err := db.Query("SELECT id, slug, title, summary, created_at FROM articles ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	defer rows.Close()

	var articles []DBArticle
	for rows.Next() {
		var article DBArticle
		err := rows.Scan(&article.ID, &article.Slug, &article.Title, &article.Summary, &article.CreatedAt)
		if err != nil {
			continue
		}
		articles = append(articles, article)
	}

	data := AdminDashboardData{
		CurrentDate: time.Now().Format("2 January 2006"),
		Articles:    articles,
		Success:     success,
		Error:       errorMsg,
	}

	tmpl, err := getTemplate("admin", "src/templates/base.gohtml", "src/templates/admin.gohtml")
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	tmpl.ExecuteTemplate(w, "base.gohtml", data)
}

func adminArticleFormHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		formData := AdminArticleFormData{
			CurrentDate: time.Now().Format("2 January 2006"),
			IsEdit:      false,
			Action:      "/admin/articles/new",
		}

		tmpl, err := getTemplate("admin_form", "src/templates/base.gohtml", "src/templates/admin_article_form.gohtml")
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
* Highlights: **Data pending - will be updated during trading hours**

### Open Analysis
<p>Market analysis will be updated when trading begins at 9:30 AM.</p>

<hr>

### Close Set
* Close Index: 0.00 (0.00)

### Close Summary
<p>Morning session summary will be available at 12:30 PM.</p>

<hr>

## Afternoon Session

### Open Set
* Open Index: 0.00 (0.00)
* Highlights: **Data pending - will be updated during trading hours**

### Open Analysis
<p>Afternoon session analysis will be updated when trading resumes at 2:30 PM.</p>

<hr>

### Close Set
* Close Index: 0.00 (0.00)

### Close Summary
<p>Final market summary will be available at 4:30 PM with complete analysis.</p>

<hr>

## Key Takeaways

- Market data will be updated throughout the trading day
- Live updates scheduled at key trading session times
- Full analysis available after market close
`)

		markdownPath := fmt.Sprintf("articles/%s.md", slug)
		os.WriteFile(markdownPath, []byte(markdownContent), 0644)

		// Clear cache for this file
		cacheMutex.Lock()
		delete(markdownCache, markdownPath)
		delete(cacheExpiry, markdownPath)
		cacheMutex.Unlock()

		_, err = db.Exec("INSERT INTO articles (slug, title, summary, content, created_at) VALUES (?, ?, ?, ?, ?)",
			slug, title, summary, markdownContent, time.Now().Format("2006-01-02"))
		if err != nil {
			http.Error(w, "Error creating article", 500)
			return
		}

		http.Redirect(w, r, "/admin?success=Article created successfully", 302)
	}
}

func addMissingArticlesToDB() {
	files, err := os.ReadDir("articles")
	if err != nil {
		return
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		slug := strings.TrimSuffix(file.Name(), ".md")

		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM articles WHERE slug = ?)", slug).Scan(&exists)
		if err != nil {
			continue
		}

		if !exists {
			title := fmt.Sprintf("Stock Market Analysis - %s", slug)
			summary := "Thai stock market analysis including SET index movements, sector highlights, and key insights."

			db.Exec("INSERT INTO articles (slug, title, summary, content, created_at) VALUES (?, ?, ?, ?, ?)",
				slug, title, summary, "", slug)
		}
	}
}

func main() {
	initDB()
	defer db.Close()

	addMissingArticlesToDB()

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("src/static"))))

	// Routes
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/articles/", articleHandler)
	http.HandleFunc("/admin", adminDashboardHandler)
	http.HandleFunc("/admin/", adminDashboardHandler)
	http.HandleFunc("/admin/articles/new", adminArticleFormHandler)

	log.Printf("Server starting on http://localhost:7777")
	if err := http.ListenAndServe(":7777", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
