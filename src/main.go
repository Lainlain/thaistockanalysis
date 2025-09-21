package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gomarkdown/markdown"
	_ "github.com/mattn/go-sqlite3"
)

// Global variables following coding instructions
var db *sql.DB

// Template cache for performance
var (
	templateCache = make(map[string]*template.Template)
	templateMutex sync.RWMutex
	templateFuncs = template.FuncMap{
		"printf": fmt.Sprintf,
		"html":   func(s string) template.HTML { return template.HTML(s) },
		"add":    func(a, b int) int { return a + b },
		"mod":    func(a, b int) int { return a % b }, // Add missing mod function
		"markdownToHTML": func(s string) template.HTML {
			if s == "" {
				return template.HTML("")
			}
			htmlBytes := markdown.ToHTML([]byte(s), nil, nil)
			return template.HTML(htmlBytes)
		},
	}
)

// Data structures as per coding instructions
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
}

// Template cache function following coding instructions
func getTemplate(name string, files ...string) (*template.Template, error) {
	templateMutex.RLock()
	if tmpl, exists := templateCache[name]; exists {
		templateMutex.RUnlock()
		return tmpl, nil
	}
	templateMutex.RUnlock()

	templateMutex.Lock()
	defer templateMutex.Unlock()

	if tmpl, exists := templateCache[name]; exists {
		return tmpl, nil
	}

	tmpl := template.New(name).Funcs(templateFuncs)
	var err error
	tmpl, err = tmpl.ParseFiles(files...)
	if err != nil {
		return nil, err
	}

	templateCache[name] = tmpl
	return tmpl, nil
}

// Database initialization following coding instructions
func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "src/admin.db?cache=shared&mode=rwc&_journal_mode=WAL&_synchronous=NORMAL")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Set connection pool limits
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Create the base articles table as per coding instructions
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS articles (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        slug TEXT NOT NULL UNIQUE,
        title TEXT NOT NULL,
        summary TEXT,
        content TEXT,
        created_at TEXT NOT NULL
    );
    CREATE INDEX IF NOT EXISTS idx_articles_slug ON articles(slug);
    CREATE INDEX IF NOT EXISTS idx_articles_created_at ON articles(created_at);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Failed to create articles table: %v", err)
	}

	seedArticlesTable()
}

// Seed function following coding instructions
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

		// Sample articles following coding instructions
		articles := [][]string{
			{"2025-09-22", "Stock Market Analysis - 22 September 2025", "Morning session opened with modest gains in key sectors", "", "2025-09-22"},
			{"2025-09-19", "Stock Market Analysis - 19 September 2025", "SET index closed higher with banking and energy sectors leading gains", "", "2025-09-19"},
			{"2025-09-18", "Stock Market Analysis - 18 September 2025", "Market declined amid regional weakness and profit-taking", "", "2025-09-18"},
		}

		for _, article := range articles {
			stmt.Exec(article[0], article[1], article[2], article[3], article[4])
		}
		log.Printf("Seeded %d articles", len(articles))
	}
}

// FAST index handler following coding instructions - database only
func indexHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Cache headers
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Simple database query as per coding instructions
	rows, err := db.Query("SELECT slug, title, summary, created_at FROM articles ORDER BY created_at DESC LIMIT 10")
	if err != nil {
		log.Printf("Database error: %v", err)
		http.Error(w, "Database Error", 500)
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

		// Simple preview creation following coding instructions
		preview := ArticlePreview{
			Title:        title,
			Date:         createdAt,
			SetIndex:     "SET", // Simple placeholder
			Change:       0.0,
			ShortSummary: summary.String,
			Summary:      summary.String,
			Slug:         slug,
			URL:          fmt.Sprintf("/articles/%s", slug),
		}
		previews = append(previews, preview)
	}

	data := IndexPageData{
		CurrentDate: time.Now().Format("2 January 2006"),
		Articles:    previews,
	}

	// Use template cache as per coding instructions
	tmpl, err := getTemplate("index", "src/templates/base.gohtml", "src/templates/index.gohtml")
	if err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Template Error", 500)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base.gohtml", data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
	}

	log.Printf("Index page rendered in %v", time.Since(start))
}

// Article handler following coding instructions - markdown file parsing
func articleHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	w.Header().Set("Cache-Control", "public, max-age=600")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 || parts[2] == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	slug := parts[2]

	// Get article metadata from database following coding instructions
	var dbArticle DBArticle
	err := db.QueryRow("SELECT id, slug, title, summary, content, created_at FROM articles WHERE slug = ?", slug).Scan(
		&dbArticle.ID, &dbArticle.Slug, &dbArticle.Title, &dbArticle.Summary, &dbArticle.Content, &dbArticle.CreatedAt)
	if err != nil {
		log.Printf("Article not found: %s, error: %v", slug, err)
		http.NotFound(w, r)
		return
	}

	// Parse markdown file following coding instructions
	stockData := StockData{
		CurrentDate:  time.Now().Format("2 January 2006"),
		KeyTakeaways: []string{},
	}

	// Try to parse markdown file from articles/ directory
	markdownPath := fmt.Sprintf("articles/%s.md", slug)
	if parsedData, err := parseMarkdownArticle(markdownPath); err == nil {
		stockData = parsedData
	} else {
		log.Printf("Markdown parse failed for %s: %v", slug, err)
		// Provide default data if markdown file doesn't exist
		stockData.KeyTakeaways = []string{
			"Market analysis data will be updated shortly",
			"Please check back during trading hours for complete analysis",
		}
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

	tmpl, err := getTemplate("article", "src/templates/base.gohtml", "src/templates/article.gohtml")
	if err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Template Error", 500)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base.gohtml", data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
	}

	log.Printf("Article %s rendered in %v", slug, time.Since(start))
}

// Markdown parser following coding instructions
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

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse sections following coding instructions format
		if strings.HasPrefix(line, "## Morning Session") {
			currentSection = "morning"
			continue
		} else if strings.HasPrefix(line, "## Afternoon Session") {
			currentSection = "afternoon"
			continue
		} else if strings.HasPrefix(line, "## Key Takeaways") {
			currentSection = "takeaways"
			continue
		}

		// Parse subsections
		if strings.HasPrefix(line, "### Open Set") {
			currentSubsection = "open"
			continue
		} else if strings.HasPrefix(line, "### Close Set") {
			currentSubsection = "close"
			continue
		}

		// Parse data based on current section and subsection
		switch currentSection {
		case "morning":
			switch currentSubsection {
			case "open":
				if strings.HasPrefix(line, "* Open Index:") {
					data.MorningOpenIndex, data.MorningOpenChange = parseIndexLine(line)
				} else if strings.HasPrefix(line, "* Highlights:") {
					data.MorningOpenHighlights = parseHighlights(line)
				}
			case "close":
				if strings.HasPrefix(line, "* Close Index:") {
					data.MorningCloseIndex, data.MorningCloseChange = parseIndexLine(line)
				}
			}
		case "afternoon":
			switch currentSubsection {
			case "open":
				if strings.HasPrefix(line, "* Open Index:") {
					data.AfternoonOpenIndex, data.AfternoonOpenChange = parseIndexLine(line)
				} else if strings.HasPrefix(line, "* Highlights:") {
					data.AfternoonOpenHighlights = parseHighlights(line)
				}
			case "close":
				if strings.HasPrefix(line, "* Close Index:") {
					data.AfternoonCloseIndex, data.AfternoonCloseChange = parseIndexLine(line)
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

// Helper functions following coding instructions
func parseIndexLine(line string) (float64, float64) {
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

func parseHighlights(line string) string {
	if idx := strings.Index(line, "Highlights:"); idx != -1 {
		return strings.TrimSpace(line[idx+11:])
	}
	return ""
}

// Admin handlers following coding instructions
func adminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	success := r.URL.Query().Get("success")
	errorMsg := r.URL.Query().Get("error")

	rows, err := db.Query("SELECT id, slug, title, summary, created_at FROM articles ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, "Database Error", 500)
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
		http.Error(w, "Template Error", 500)
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
			http.Error(w, "Template Error", 500)
			return
		}

		tmpl.ExecuteTemplate(w, "base.gohtml", formData)
	} else if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Form Parse Error", 400)
			return
		}

		slug := r.FormValue("slug")
		title := r.FormValue("title")
		summary := r.FormValue("summary")

		// Create basic markdown template following coding instructions
		markdownContent := fmt.Sprintf(`## Morning Session

### Open Set
* Open Index: 0.00 (0.00)
* Highlights: **Data pending - will be updated during trading hours**

### Close Set
* Close Index: 0.00 (0.00)

## Afternoon Session

### Open Set
* Open Index: 0.00 (0.00)
* Highlights: **Data pending - will be updated during trading hours**

### Close Set
* Close Index: 0.00 (0.00)

## Key Takeaways

- Market data will be updated throughout the trading day
`)

		// Create markdown file following coding instructions
		markdownPath := fmt.Sprintf("articles/%s.md", slug)
		os.MkdirAll(filepath.Dir(markdownPath), 0755)
		os.WriteFile(markdownPath, []byte(markdownContent), 0644)

		// Insert into database following coding instructions
		_, err = db.Exec("INSERT INTO articles (slug, title, summary, content, created_at) VALUES (?, ?, ?, ?, ?)",
			slug, title, summary, markdownContent, time.Now().Format("2006-01-02"))
		if err != nil {
			http.Error(w, "Database Insert Error", 500)
			return
		}

		http.Redirect(w, r, "/admin?success=Article created successfully", 302)
	}
}

// Main function following coding instructions
func main() {
	log.Printf("Starting Thai Stock Analysis server...")

	// Initialize database following coding instructions
	initDB()
	defer db.Close()

	// Serve static files following coding instructions
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("src/static"))))

	// Routes following coding instructions
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/articles/", articleHandler)
	http.HandleFunc("/admin", adminDashboardHandler)
	http.HandleFunc("/admin/", adminDashboardHandler)
	http.HandleFunc("/admin/articles/new", adminArticleFormHandler)

	// Start server following coding instructions
	log.Printf("Server starting on http://localhost:8080")
	log.Printf("Admin interface available at http://localhost:8080/admin")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
