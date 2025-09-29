package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"thaistockanalysis/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// InitDB initializes the database connection and creates necessary tables
func InitDB(dbPath string) error {
	// Create directory structure if it doesn't exist
	if strings.Contains(dbPath, "/") {
		dbDir := dbPath[:strings.LastIndex(dbPath, "/")]
		if _, err := os.Stat(dbDir); os.IsNotExist(err) {
			err := os.MkdirAll(dbDir, 0755)
			if err != nil {
				return fmt.Errorf("failed to create database directory '%s': %v", dbDir, err)
			}
			log.Printf("📁 Created database directory: %s", dbDir)
		}
	}

	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
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
	_, err = DB.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create articles table: %v", err)
	}

	// Check and add content column if it doesn't exist
	var columnName string
	err = DB.QueryRow("SELECT name FROM PRAGMA_TABLE_INFO('articles') WHERE name='content'").Scan(&columnName)
	if err == sql.ErrNoRows {
		_, err = DB.Exec("ALTER TABLE articles ADD COLUMN content TEXT")
		if err != nil {
			return fmt.Errorf("failed to add 'content' column: %v", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check for 'content' column: %v", err)
	}

	seedArticlesTable()
	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// GetArticles retrieves articles from the database with pagination
func GetArticles(limit int) ([]models.DBArticle, error) {
	query := "SELECT id, slug, title, summary, created_at FROM articles ORDER BY created_at DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []models.DBArticle
	for rows.Next() {
		var article models.DBArticle
		err := rows.Scan(&article.ID, &article.Slug, &article.Title, &article.Summary, &article.CreatedAt)
		if err != nil {
			continue
		}
		articles = append(articles, article)
	}

	return articles, nil
}

// GetArticleBySlug retrieves a single article by its slug
func GetArticleBySlug(slug string) (*models.DBArticle, error) {
	var article models.DBArticle
	err := DB.QueryRow("SELECT id, slug, title, summary, content, created_at FROM articles WHERE slug = ?", slug).Scan(
		&article.ID, &article.Slug, &article.Title, &article.Summary, &article.Content, &article.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &article, nil
}

// CreateArticle creates a new article in the database
func CreateArticle(slug, title, summary, content string) error {
	_, err := DB.Exec("INSERT INTO articles (slug, title, summary, content, created_at) VALUES (?, ?, ?, ?, ?)",
		slug, title, summary, content, slug)
	return err
}

// ArticleExists checks if an article with the given slug exists
func ArticleExists(slug string) (bool, error) {
	var exists bool
	err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM articles WHERE slug = ?)", slug).Scan(&exists)
	return exists, err
}

// seedArticlesTable populates the database with initial articles if empty
func seedArticlesTable() {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM articles").Scan(&count)
	if err != nil {
		log.Printf("Error checking article count: %v", err)
		return
	}

	if count == 0 {
		stmt, err := DB.Prepare("INSERT INTO articles(slug, title, summary, content, created_at) VALUES(?, ?, ?, ?, ?)")
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

// AddMissingArticlesToDB syncs filesystem articles to database
func AddMissingArticlesToDB(articlesDir string) {
	files, err := os.ReadDir(articlesDir)
	if err != nil {
		log.Printf("Error reading articles directory: %v", err)
		return
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		slug := strings.TrimSuffix(file.Name(), ".md")

		exists, err := ArticleExists(slug)
		if err != nil {
			log.Printf("Error checking if article exists: %v", err)
			continue
		}

		if !exists {
			title := fmt.Sprintf("Stock Market Analysis - %s", slug)
			summary := "Thai stock market analysis including SET index movements, sector highlights, and key insights."

			err := CreateArticle(slug, title, summary, "")
			if err != nil {
				log.Printf("Error creating article %s: %v", slug, err)
			}
		}
	}
}
