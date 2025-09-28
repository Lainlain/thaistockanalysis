package models


import (
	"database/sql"
	"html/template"
)

// StockData represents parsed stock market data from markdown articles
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

// ArticlePreview represents a summary view of an article for listings
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

// DBArticle represents an article as stored in the database
type DBArticle struct {
	ID        int
	Slug      string
	Title     string
	Summary   sql.NullString
	Content   sql.NullString
	CreatedAt string
}

// IndexPageData contains data for the homepage template
type IndexPageData struct {
	CurrentDate string
	Articles    []ArticlePreview
}

// AdminDashboardData contains data for the admin dashboard template
type AdminDashboardData struct {
	CurrentDate string
	Articles    []DBArticle
	Success     string
	Error       string
}

// AdminArticleFormData contains data for the article creation/edit form
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

// ArticleDetail contains complete article data for display
type ArticleDetail struct {
	Title     string
	Slug      string
	Summary   string
	CreatedAt string
	StockData
}