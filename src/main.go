package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gomarkdown/markdown"
	_ "github.com/mattn/go-sqlite3"
)

// Global database connection
var db *sql.DB

type StockData struct {
	CurrentDate              string
	MorningOpenIndex         string
	MorningOpenChange        float64
	MorningOpenHighlights    string
	MorningOpenAnalysis      template.HTML
	MorningCloseIndex        string
	MorningCloseChange       float64
	MorningCloseHighlights   string
	MorningCloseSummary      template.HTML
	AfternoonOpenIndex       string
	AfternoonOpenChange      float64
	AfternoonOpenHighlights  string
	AfternoonOpenAnalysis    template.HTML
	AfternoonCloseIndex      string
	AfternoonCloseChange     float64
	AfternoonCloseHighlights string
	AfternoonCloseSummary    template.HTML
	KeyTakeaways             []string
}

type ArticlePreview struct {
	Date         string
	SetIndex     string
	Change       float64
	ShortSummary string
	Slug         string
}

// DBArticle represents an article stored in the database
type DBArticle struct {
	ID        int
	Slug      string
	Title     string
	Summary   sql.NullString // Can be NULL
	Content   sql.NullString // Can be NULL
	CreatedAt string         // Stored as TEXT in YYYY-MM-DD format
}

type IndexPageData struct {
	CurrentDate string
	Articles    []ArticlePreview
}

// AdminDashboardData for the admin template
type AdminDashboardData struct {
	CurrentDate string
	Articles    []DBArticle
	Success     string // For success messages
	Error       string // For error messages
}

// AdminArticleFormData for the article form template (for new/edit)
type AdminArticleFormData struct {
	CurrentDate string // For default date in form
	Article     DBArticle
	Error       string
	IsEdit      bool   // To differentiate between new and edit mode in the template
	Action      string // Add Action field

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
	KeyTakeaways             string // Will be a comma-separated string from form
}

// PageData for static pages like privacy, terms, etc.
type PageData struct {
	LastUpdated string
}

// initDB initializes the SQLite database and creates tables
func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./admin.db") // Open or create admin.db in project root
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Create articles table if it doesn't exist
	createTableSQL := `
    CREATE TABLE IF NOT EXISTS articles (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        slug TEXT NOT NULL UNIQUE,
        title TEXT NOT NULL,
        summary TEXT,
        content TEXT, -- New column for markdown content
        created_at TEXT NOT NULL
    );`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Failed to create articles table: %v", err)
	}
	log.Println("Database initialized and 'articles' table ensured.")

	// Add 'content' column if it doesn't exist (for backward compatibility)
	var columnName string
	err = db.QueryRow("SELECT name FROM PRAGMA_TABLE_INFO('articles') WHERE name='content'").Scan(&columnName)
	if err == sql.ErrNoRows {
		log.Println("Adding 'content' column to 'articles' table...")
		_, err = db.Exec("ALTER TABLE articles ADD COLUMN content TEXT")
		if err != nil {
			log.Fatalf("Failed to add 'content' column: %v", err)
		}
		log.Println("'content' column added successfully.")
	} else if err != nil {
		log.Fatalf("Failed to check for 'content' column: %v", err)
	}

	// Seed some sample data if the table is empty
	seedArticlesTable()
}

// seedArticlesTable inserts sample articles into the database if it's empty
func seedArticlesTable() {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM articles").Scan(&count)
	if err != nil {
		log.Printf("Error checking article count: %v", err)
		return
	}

	if count == 0 {
		log.Println("Seeding articles table with sample data...")
		// Added 'content' to the INSERT statement
		stmt, err := db.Prepare("INSERT INTO articles(slug, title, summary, content, created_at) VALUES(?, ?, ?, ?, ?)")
		if err != nil {
			log.Printf("Error preparing statement: %v", err)
			return
		}
		defer stmt.Close()

		// Provide an empty string for the content field
		_, err = stmt.Exec("2025-09-19", "Stock Market Analysis - 19 September 2025", "SET index closed higher with banking and energy sectors leading gains", "", "2025-09-19")
		if err != nil {
			log.Printf("Error inserting sample article 1: %v", err)
		}
		_, err = stmt.Exec("2025-09-18", "Stock Market Analysis - 18 September 2025", "Market declined amid regional weakness", "", "2025-09-18")
		if err != nil {
			log.Printf("Error inserting sample article 2: %v", err)
		}
		log.Println("Sample articles seeded.")
	}
}

// parseMarkdownArticle reads a markdown file and extracts data into StockData
func parseMarkdownArticle(filePath string) (StockData, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return StockData{}, fmt.Errorf("failed to read markdown file: %w", err)
	}

	data := StockData{}
	data.CurrentDate = time.Now().Format("2 January 2006") // Default, will be overridden by markdown

	lines := strings.Split(string(content), "\n")

	var currentContentBuilder strings.Builder
	var currentContentTarget string // e.g., "MorningOpenAnalysis", "MorningCloseSummary"

	// Helper function to flush the builder content to the correct field
	flushContent := func() {
		if currentContentBuilder.Len() > 0 && currentContentTarget != "" {
			html := template.HTML(markdown.ToHTML([]byte(currentContentBuilder.String()), nil, nil))
			switch currentContentTarget {
			case "MorningOpenAnalysis":
				data.MorningOpenAnalysis = html
			case "MorningCloseSummary":
				data.MorningCloseSummary = html
			case "AfternoonOpenAnalysis":
				data.AfternoonOpenAnalysis = html
			case "AfternoonCloseSummary":
				data.AfternoonCloseSummary = html
			}
			currentContentBuilder.Reset()
			currentContentTarget = "" // Reset target after flushing
		}
	}

	// State variables to track where we are in the document
	inMorningSession := false
	inAfternoonSession := false
	inKeyTakeawaysSection := false // Renamed to avoid conflict with data.KeyTakeaways

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Session headers
		if strings.HasPrefix(line, "## Morning Session") {
			flushContent()
			inMorningSession = true
			inAfternoonSession = false
			inKeyTakeawaysSection = false
			continue
		}
		if strings.HasPrefix(line, "## Afternoon Session") {
			flushContent()
			inMorningSession = false
			inAfternoonSession = true
			inKeyTakeawaysSection = false
			continue
		}
		if strings.HasPrefix(line, "## Key Takeaways") {
			flushContent()
			inMorningSession = false
			inAfternoonSession = false
			inKeyTakeawaysSection = true
			continue
		}

		// Handle content based on current session
		if inMorningSession {
			if strings.HasPrefix(line, "### Open Set") {
				flushContent()
				currentContentTarget = "MorningOpenSet" // Set context for highlights
				continue
			}
			if strings.HasPrefix(line, "### Open Analysis") {
				flushContent()
				currentContentTarget = "MorningOpenAnalysis"
				continue
			}
			if strings.HasPrefix(line, "### Close Set") {
				flushContent()
				currentContentTarget = "MorningCloseSet" // Set context for highlights
				continue
			}
			if strings.HasPrefix(line, "### Close Summary") {
				flushContent()
				currentContentTarget = "MorningCloseSummary"
				continue
			}

			// List items within Morning Session
			if strings.HasPrefix(line, "* Open Index:") {
				parts := strings.Fields(strings.TrimPrefix(line, "* Open Index:"))
				if len(parts) > 0 {
					data.MorningOpenIndex = parts[0]
				}
				if len(parts) > 1 {
					changeStr := strings.Trim(parts[1], "()")
					if change, err := strconv.ParseFloat(changeStr, 64); err == nil {
						data.MorningOpenChange = change
					}
				}
				continue
			}
			if strings.HasPrefix(line, "* Highlights:") {
				if currentContentTarget == "MorningOpenSet" {
					data.MorningOpenHighlights = strings.TrimSpace(strings.TrimPrefix(line, "* Highlights:"))
				} else if currentContentTarget == "MorningCloseSet" {
					data.MorningCloseHighlights = strings.TrimSpace(strings.TrimPrefix(line, "* Highlights:"))
				}
				continue
			}
			if strings.HasPrefix(line, "* Close Index:") {
				parts := strings.Fields(strings.TrimPrefix(line, "* Close Index:"))
				if len(parts) > 0 {
					data.MorningCloseIndex = parts[0]
				}
				if len(parts) > 1 {
					changeStr := strings.Trim(parts[1], "()")
					if change, err := strconv.ParseFloat(changeStr, 64); err == nil {
						data.MorningCloseChange = change
					}
				}
				continue
			}
		} else if inAfternoonSession {
			if strings.HasPrefix(line, "### Open Set") {
				flushContent()
				currentContentTarget = "AfternoonOpenSet" // Set context for highlights
				continue
			}
			if strings.HasPrefix(line, "### Open Analysis") {
				flushContent()
				currentContentTarget = "AfternoonOpenAnalysis"
				continue
			}
			if strings.HasPrefix(line, "### Close Set") || strings.HasPrefix(line, "### End of Day Set") {
				flushContent()
				currentContentTarget = "AfternoonCloseSet" // Set context for highlights
				continue
			}
			if strings.HasPrefix(line, "### Close Summary") {
				flushContent()
				currentContentTarget = "AfternoonCloseSummary"
				continue
			}

			// List items within Afternoon Session
			if strings.HasPrefix(line, "* Open Index:") {
				parts := strings.Fields(strings.TrimPrefix(line, "* Open Index:"))
				if len(parts) > 0 {
					data.AfternoonOpenIndex = parts[0]
				}
				if len(parts) > 1 {
					changeStr := strings.Trim(parts[1], "()")
					if change, err := strconv.ParseFloat(changeStr, 64); err == nil {
						data.AfternoonOpenChange = change
					}
				}
				continue
			}
			if strings.HasPrefix(line, "* Highlights:") {
				if currentContentTarget == "AfternoonOpenSet" {
					data.AfternoonOpenHighlights = strings.TrimSpace(strings.TrimPrefix(line, "* Highlights:"))
				} else if currentContentTarget == "AfternoonCloseSet" {
					data.AfternoonCloseHighlights = strings.TrimSpace(strings.TrimPrefix(line, "* Highlights:"))
				}
				continue
			}
			if strings.HasPrefix(line, "* Close Index:") {
				parts := strings.Fields(strings.TrimPrefix(line, "* Close Index:"))
				if len(parts) > 0 {
					data.AfternoonCloseIndex = parts[0]
				}
				if len(parts) > 1 {
					changeStr := strings.Trim(parts[1], "()")
					if change, err := strconv.ParseFloat(changeStr, 64); err == nil {
						data.AfternoonCloseChange = change
					}
				}
				continue
			}
		} else if inKeyTakeawaysSection {
			if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
				prefix := "-"
				if strings.HasPrefix(line, "*") {
					prefix = "*"
				}
				data.KeyTakeaways = append(data.KeyTakeaways, strings.TrimSpace(strings.TrimPrefix(line, prefix)))
				continue
			}
		}

		// If we are collecting content for a target field, append the line
		if currentContentTarget != "" {
			currentContentBuilder.WriteString(line + "\n")
		}
	}

	flushContent() // Flush any remaining content after the loop

	log.Printf("DEBUG: parseMarkdownArticle finished for %s. Data: %+v", filePath, data)

	return data, nil
}

// buildMarkdownFromForm constructs a markdown string from the structured form fields
func buildMarkdownFromForm(r *http.Request) string {
	// Ensure form is parsed
	_ = r.ParseForm()

	// Helper to get and trim form value
	get := func(key string) string { return strings.TrimSpace(r.FormValue(key)) }

	// Read fields
	moi := get("morning_open_index")
	moc := get("morning_open_change")
	moh := get("morning_open_highlights")
	moa := get("morning_open_analysis")
	mci := get("morning_close_index")
	mcc := get("morning_close_change")
	mch := get("morning_close_highlights")
	mcs := get("morning_close_summary")

	aoi := get("afternoon_open_index")
	aoc := get("afternoon_open_change")
	aoh := get("afternoon_open_highlights")
	aoa := get("afternoon_open_analysis")
	aci := get("afternoon_close_index")
	acc := get("afternoon_close_change")
	ach := get("afternoon_close_highlights")
	acs := get("afternoon_close_summary")

	kt := get("key_takeaways")

	// Helper for formatting change like (0.23)
	formatChange := func(s string) string {
		if s == "" {
			return ""
		}
		if _, err := strconv.ParseFloat(s, 64); err == nil {
			return fmt.Sprintf(" (%s)", s)
		}
		// If not a number already, just wrap in parentheses if not present
		if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
			return " " + s
		}
		return fmt.Sprintf(" (%s)", s)
	}

	var b strings.Builder

	// Morning Session
	b.WriteString("## Morning Session\n\n")
	// Open
	b.WriteString("### Open Set\n")
	if moi != "" {
		b.WriteString("* Open Index: ")
		b.WriteString(moi)
		b.WriteString(formatChange(moc))
		b.WriteString("\n")
	}
	if moh != "" {
		b.WriteString("* Highlights: ")
		b.WriteString(moh)
		b.WriteString("\n")
	}
	b.WriteString("\n### Open Analysis\n")
	if moa != "" {
		b.WriteString(moa)
		if !strings.HasSuffix(moa, "\n") {
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")
	// Close
	b.WriteString("### Close Set\n")
	if mci != "" {
		b.WriteString("* Close Index: ")
		b.WriteString(mci)
		b.WriteString(formatChange(mcc))
		b.WriteString("\n")
	}
	if mch != "" {
		b.WriteString("* Highlights: ")
		b.WriteString(mch)
		b.WriteString("\n")
	}
	b.WriteString("\n### Close Summary\n")
	if mcs != "" {
		b.WriteString(mcs)
		if !strings.HasSuffix(mcs, "\n") {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	// Afternoon Session
	b.WriteString("## Afternoon Session\n\n")
	// Open
	b.WriteString("### Open Set\n")
	if aoi != "" {
		b.WriteString("* Open Index: ")
		b.WriteString(aoi)
		b.WriteString(formatChange(aoc))
		b.WriteString("\n")
	}
	if aoh != "" {
		b.WriteString("* Highlights: ")
		b.WriteString(aoh)
		b.WriteString("\n")
	}
	b.WriteString("\n### Open Analysis\n")
	if aoa != "" {
		b.WriteString(aoa)
		if !strings.HasSuffix(aoa, "\n") {
			b.WriteString("\n")
		}
	}
	b.WriteString("\n")
	// Close
	b.WriteString("### Close Set\n")
	if aci != "" {
		b.WriteString("* Close Index: ")
		b.WriteString(aci)
		b.WriteString(formatChange(acc))
		b.WriteString("\n")
	}
	if ach != "" {
		b.WriteString("* Highlights: ")
		b.WriteString(ach)
		b.WriteString("\n")
	}
	b.WriteString("\n### Close Summary\n")
	if acs != "" {
		b.WriteString(acs)
		if !strings.HasSuffix(acs, "\n") {
			b.WriteString("\n")
		}
	}

	// Key Takeaways
	ktItems := []string{}
	for _, item := range strings.Split(kt, ",") {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			ktItems = append(ktItems, trimmed)
		}
	}
	if len(ktItems) > 0 {
		b.WriteString("\n## Key Takeaways\n\n")
		for _, item := range ktItems {
			b.WriteString("- ")
			b.WriteString(item)
			b.WriteString("\n")
		}
	}

	return b.String()
}

// formatDate converts "YYYY-MM-DD" to "DD MMM YYYY"
func formatDate(dateStr string) string {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		log.Printf("Error parsing date string '%s': %v", dateStr, err)
		return dateStr // Return original on error
	}
	return t.Format("02 Jan 2006")
}

func main() {
	initDB()
	defer db.Close()

	fs := http.FileServer(http.Dir("src/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/articles/", articleHandler)
	http.HandleFunc("/admin", adminDashboardHandler)
	http.HandleFunc("/admin/articles/new", adminArticleFormHandler)
	http.HandleFunc("/admin/articles/edit/", adminArticleFormHandler)
	http.HandleFunc("/admin/articles/save", adminArticleSaveHandler)
	http.HandleFunc("/admin/articles/delete/", adminArticleDeleteHandler)

	// Static page routes
	http.HandleFunc("/about", staticPageHandler("about.gohtml", "About Us"))
	http.HandleFunc("/contact", staticPageHandler("contact.gohtml", "Contact Us"))
	http.HandleFunc("/privacy", staticPageHandler("privacy.gohtml", "Privacy Policy"))
	http.HandleFunc("/terms", staticPageHandler("terms.gohtml", "Terms of Service"))
	http.HandleFunc("/disclaimer", staticPageHandler("disclaimer.gohtml", "Disclaimer"))

	log.Println("Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	baseTmpl, err := template.ParseFiles("src/templates/base.gohtml")
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	tmpl, err := baseTmpl.Clone()
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	_, err = tmpl.ParseFiles("src/templates/index.gohtml")
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	rows, err := db.Query("SELECT slug, title, summary, created_at FROM articles ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, "Failed to fetch articles", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var articles []ArticlePreview
	for rows.Next() {
		var dbArticle DBArticle
		if err := rows.Scan(&dbArticle.Slug, &dbArticle.Title, &dbArticle.Summary, &dbArticle.CreatedAt); err != nil {
			log.Printf("Error scanning article from DB: %v", err)
			continue // Skip this article on error
		}

		// Parse the corresponding markdown file to get live data
		filePath := fmt.Sprintf("articles/%s.md", dbArticle.Slug)
		stockData, err := parseMarkdownArticle(filePath)
		if err != nil {
			log.Printf("Could not parse markdown file %s: %v. Skipping.", filePath, err)
			continue // Skip article if markdown can't be parsed
		}

		summary := ""
		if dbArticle.Summary.Valid {
			summary = dbArticle.Summary.String
		}

		// Use parsed data. Fallback to "N/A" if not found.
		setIndex := "N/A"
		change := 0.0
		if stockData.AfternoonCloseIndex != "" {
			setIndex = stockData.AfternoonCloseIndex
			change = stockData.AfternoonCloseChange
		}

		article := ArticlePreview{
			Date:         formatDate(dbArticle.CreatedAt),
			SetIndex:     setIndex,
			Change:       change,
			ShortSummary: summary,
			Slug:         dbArticle.Slug,
		}
		articles = append(articles, article)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating over articles", http.StatusInternalServerError)
		return
	}

	data := IndexPageData{
		CurrentDate: time.Now().Format("2 January 2006"),
		Articles:    articles,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Internal Server Error", 500)
	}
}

func articleHandler(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/articles/")
	if slug == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	stockData, err := parseMarkdownArticle("articles/" + slug + ".md")
	if err != nil {
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}

	// Set the date from the slug
	t, err := time.Parse("2006-01-02", slug)
	if err == nil {
		stockData.CurrentDate = t.Format("2 January 2006")
	}

	baseTmpl, err := template.ParseFiles("src/templates/base.gohtml")
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	tmpl, err := baseTmpl.Clone()
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	_, err = tmpl.ParseFiles("src/templates/article.gohtml")
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base.gohtml", stockData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func adminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, slug, title, summary, content, created_at FROM articles ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, "Failed to fetch articles", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var articles []DBArticle
	for rows.Next() {
		var article DBArticle
		if err := rows.Scan(&article.ID, &article.Slug, &article.Title, &article.Summary, &article.Content, &article.CreatedAt); err != nil {
			http.Error(w, "Failed to scan article", http.StatusInternalServerError)
			return
		}
		articles = append(articles, article)
	}

	baseTmpl, err := template.ParseFiles("src/templates/base.gohtml")
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	tmpl, err := baseTmpl.Clone()
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	_, err = tmpl.ParseFiles("src/templates/admin.gohtml")
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base.gohtml", AdminDashboardData{Articles: articles})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func adminArticleFormHandler(w http.ResponseWriter, r *http.Request) {
	baseTmpl, err := template.ParseFiles("src/templates/base.gohtml")
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	tmpl, err := baseTmpl.Clone()
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	_, err = tmpl.ParseFiles("src/templates/admin_article_form.gohtml")
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}

	// Check if it's an edit or new request
	if strings.HasPrefix(r.URL.Path, "/admin/articles/edit/") {
		idStr := strings.TrimPrefix(r.URL.Path, "/admin/articles/edit/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid article ID", http.StatusBadRequest)
			return
		}

		var article DBArticle
		err = db.QueryRow("SELECT id, slug, title, summary, created_at FROM articles WHERE id = ?", id).Scan(&article.ID, &article.Slug, &article.Title, &article.Summary, &article.CreatedAt)
		if err != nil {
			http.Error(w, "Article not found", http.StatusNotFound)
			return
		}

		filePath := "articles/" + article.Slug + ".md"
		formData := AdminArticleFormData{
			Article: article,
			IsEdit:  true,
			Action:  "/admin/articles/save",
		}

		if parsed, perr := parseMarkdownArticle(filePath); perr == nil {
			formData.MorningOpenIndex = parsed.MorningOpenIndex
			formData.MorningOpenChange = parsed.MorningOpenChange
			formData.MorningOpenHighlights = parsed.MorningOpenHighlights
			formData.MorningOpenAnalysis = string(parsed.MorningOpenAnalysis)
			formData.MorningCloseIndex = parsed.MorningCloseIndex
			formData.MorningCloseChange = parsed.MorningCloseChange
			formData.MorningCloseHighlights = parsed.MorningCloseHighlights
			formData.MorningCloseSummary = string(parsed.MorningCloseSummary)
			formData.AfternoonOpenIndex = parsed.AfternoonOpenIndex
			formData.AfternoonOpenChange = parsed.AfternoonOpenChange
			formData.AfternoonOpenHighlights = parsed.AfternoonOpenHighlights
			formData.AfternoonOpenAnalysis = string(parsed.AfternoonOpenAnalysis)
			formData.AfternoonCloseIndex = parsed.AfternoonCloseIndex
			formData.AfternoonCloseChange = parsed.AfternoonCloseChange
			formData.AfternoonCloseHighlights = parsed.AfternoonCloseHighlights
			formData.AfternoonCloseSummary = string(parsed.AfternoonCloseSummary)
			formData.KeyTakeaways = strings.Join(parsed.KeyTakeaways, ", ")
		} else {
			log.Printf("Could not parse markdown for edit form (slug: %s): %v", article.Slug, perr)
		}

		err = tmpl.ExecuteTemplate(w, "base.gohtml", formData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else { // New article
		data := AdminArticleFormData{
			CurrentDate: time.Now().Format("2006-01-02"),
			IsEdit:      false,
			Action:      "/admin/articles/save",
		}
		err = tmpl.ExecuteTemplate(w, "base.gohtml", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func adminArticleSaveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm()

	idStr := r.FormValue("id")
	slug := r.FormValue("slug")
	title := r.FormValue("title")
	summary := r.FormValue("summary")
	createdAt := r.FormValue("created_at")

	if slug == "" || title == "" || createdAt == "" {
		http.Error(w, "Slug, Title, and Created At are required", http.StatusBadRequest)
		return
	}

	content := buildMarkdownFromForm(r)

	// Write the markdown file
	os.MkdirAll("articles", 0755)
	filePath := "articles/" + slug + ".md"
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		log.Printf("Error writing markdown file %s: %v", filePath, err)
		http.Error(w, "Failed to save article content", http.StatusInternalServerError)
		return
	}

	if idStr != "" { // Update existing article
		id, _ := strconv.Atoi(idStr)
		var oldSlug string
		db.QueryRow("SELECT slug FROM articles WHERE id = ?", id).Scan(&oldSlug)

		_, err := db.Exec("UPDATE articles SET slug=?, title=?, summary=?, content=?, created_at=? WHERE id=?",
			slug, title, summary, content, createdAt, id)
		if err != nil {
			http.Error(w, "Failed to update article in database", http.StatusInternalServerError)
			return
		}

		// If slug changed, remove old markdown file
		if oldSlug != slug && oldSlug != "" {
			os.Remove("articles/" + oldSlug + ".md")
		}
	} else { // Insert new article
		_, err := db.Exec("INSERT INTO articles (slug, title, summary, content, created_at) VALUES (?, ?, ?, ?, ?)",
			slug, title, summary, content, createdAt)
		if err != nil {
			http.Error(w, "Failed to create article in database", http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func adminArticleDeleteHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/admin/articles/delete/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	var slug string
	err = db.QueryRow("SELECT slug FROM articles WHERE id = ?", id).Scan(&slug)
	if err != nil {
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}

	_, err = db.Exec("DELETE FROM articles WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Failed to delete article from database", http.StatusInternalServerError)
		return
	}

	// Delete the associated markdown file
	os.Remove("articles/" + slug + ".md")

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// staticPageHandler creates a handler for a static page.
func staticPageHandler(templateName, pageTitle string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		baseTmpl, err := template.ParseFiles("src/templates/base.gohtml")
		if err != nil {
			log.Printf("Error parsing base template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		tmpl, err := baseTmpl.Clone()
		if err != nil {
			log.Printf("Error cloning base template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		_, err = tmpl.ParseFiles(fmt.Sprintf("src/templates/%s", templateName))
		if err != nil {
			log.Printf("Error parsing %s: %v", templateName, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		data := PageData{
			LastUpdated: time.Now().Format("2 January 2006"),
		}

		err = tmpl.ExecuteTemplate(w, "base.gohtml", data)
		if err != nil {
			log.Printf("Error executing template %s: %v", templateName, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}
