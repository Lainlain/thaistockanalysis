package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"thaistockanalysis/internal/database"
	"thaistockanalysis/internal/models"
	"thaistockanalysis/internal/services"
)

// Handler contains dependencies for HTTP handlers
type Handler struct {
	MarkdownService *services.MarkdownService
	TemplateService *services.TemplateService
	ArticlesDir     string
	TemplateDir     string
}

// NewHandler creates a new handler with dependencies
func NewHandler(articlesDir, templateDir string) *Handler {
	return &Handler{
		MarkdownService: services.NewMarkdownService(),
		TemplateService: services.NewTemplateService(),
		ArticlesDir:     articlesDir,
		TemplateDir:     templateDir,
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
		var shortSummary string = article.Summary.String

		// Parse markdown file to get real data
		if stockData, err := h.MarkdownService.GetCachedStockData(markdownPath); err == nil {
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

		preview := models.ArticlePreview{
			Title:        article.Title,
			Date:         article.CreatedAt,
			SetIndex:     setIndex,
			Change:       change,
			ShortSummary: shortSummary,
			Summary:      shortSummary,
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
		http.Error(w, "Internal Server Error", 500)
		return
	}

	tmpl.ExecuteTemplate(w, "base.gohtml", data)
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
