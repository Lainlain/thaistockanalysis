# ThaiStockAnalysis - AI Agent Instructions

This document provides essential knowledge for AI coding agents to be immediately productive in the `ThaiStockAnalysis` Go codebase.

## 1. Big Picture Architecture

The application is a Go-based web server that provides stock market analysis. It serves public-facing article pages and an admin interface for managing these articles.

-   **Language & Framework:** Go, with `net/http` for routing.
-   **Templating:** Uses Go's built-in `html/template` package. Templates are located in `src/templates/`. `base.gohtml` acts as the main layout, with other templates embedding their content into it.
-   **Markdown Processing:** `github.com/gomarkdown/markdown` is used to convert markdown content from article files into HTML for display.
-   **Database:** SQLite, managed via `github.com/mattn/go-sqlite3`. The database file is `src/admin.db`.
-   **Article Storage:** Articles are stored in two ways:
    1.  **Markdown Files:** Detailed stock analysis content is stored as `.md` files in the `articles/` directory (e.g., `articles/2025-09-19.md`). These are parsed by `parseMarkdownArticle` in `src/main.go`.
    2.  **SQLite Database:** Metadata for articles (slug, title, summary, creation date) is stored in the `articles` table within `src/admin.db`. This is used for the admin dashboard and article previews.

## 2. Key Components and Data Flows

-   **`src/main.go`:**
    -   **Entry Point:** Contains the `main` function, which initializes the database (`initDB`), sets up static file serving, defines all HTTP routes, and starts the server.
    -   **Data Structures:** Defines `StockData` (for detailed article content), `ArticlePreview` (for index page listings), `DBArticle` (for database interaction), `IndexPageData`, `AdminDashboardData`, and `AdminArticleFormData` (for template rendering).
    -   **`parseMarkdownArticle(filePath string) (StockData, error)`:** Reads and parses markdown files into a `StockData` struct. It extracts structured data based on markdown headers and list items.
    -   **`initDB()` and `seedArticlesTable()`:** Handle database setup and initial data population.

-   **`articles/` directory:** Contains markdown files (`YYYY-MM-DD.md`) with the detailed stock analysis content.

-   **`src/templates/` directory:**
    -   `base.gohtml`: The main layout template.
    -   `index.gohtml`: Displays a list of `ArticlePreview` items.
    -   `article.gohtml`: Displays a single `StockData` article.
    -   `admin.gohtml`: Admin dashboard, listing `DBArticle` entries from the database.
    -   `admin_article_form.gohtml`: Form for creating/editing articles.

## 3. Critical Developer Workflows

-   **Running the application:**
    ```bash
    go run src/main.go
    ```
    The server will start on `http://localhost:8080`.

-   **Database Interaction:**
    -   The `admin.db` SQLite database is automatically initialized and seeded on application startup if it doesn't exist or is empty.
    -   Database operations (CRUD for articles) are handled directly within the HTTP handlers in `src/main.go` using `database/sql` and `github.com/mattn/go-sqlite3`.

-   **Article Management (Admin Interface):**
    -   Access the admin dashboard at `http://localhost:8080/admin`.
    -   New articles can be added via `http://localhost:8080/admin/articles/new`.
    -   Existing articles can be edited via `http://localhost:8080/admin/articles/edit/{id}`.
    -   Note that while article metadata is managed in the DB, the full content for public viewing is still read from markdown files.

## 4. Project-Specific Conventions and Patterns

-   **Routing:** All routing and handler logic is centralized in `src/main.go` using `http.HandleFunc`. There is no separate `handlers` package.
-   **Template Loading:** Templates are loaded by cloning `base.gohtml` and then parsing specific page templates into the cloned set. This ensures `base.gohtml` is always the parent.
-   **Markdown Parsing:** The `parseMarkdownArticle` function has specific logic to extract data based on predefined markdown headers (e.g., `## Morning Session`, `### Open Index:`, `* Highlights:`) and converts analysis/summary sections to HTML. Any changes to the markdown article structure will require updates to this parsing logic.
-   **Slug Format:** Article slugs are expected to be in `YYYY-MM-DD` format, matching the markdown file names.

## 5. Integration Points and External Dependencies

-   **`go.mod`:** Lists direct dependencies:
    -   `github.com/gomarkdown/markdown`: For markdown to HTML conversion.
    -   `github.com/mattn/go-sqlite3`: SQLite driver for Go.
-   **Static Assets:** Served from the `src/static/` directory under the `/static/` URL path.

---
Please provide feedback on any unclear or incomplete sections to iterate on these instructions.
