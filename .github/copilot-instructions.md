# ThaiStockAnalysis - AI Agent Instructions

This document provides essential knowledge for AI coding agents to be immediately productive in the `ThaiStockAnalysis` Go codebase.

## 1. Big Picture Architecture

The application is a performance-optimized Go web server that provides Thai stock market analysis. It serves public-facing article pages and an admin interface for managing articles.

-   **Language & Framework:** Go 1.24.6, with `net/http` for routing.
-   **Performance Design:** Heavy emphasis on caching - template cache, markdown cache with expiry, and fast database-only index loading.
-   **Templating:** Uses Go's built-in `html/template` package with custom functions (`printf`, `html`, `add`, `markdownToHTML`). Templates in `src/templates/` use `base.gohtml` as the main layout.
-   **Markdown Processing:** `github.com/gomarkdown/markdown` converts markdown to HTML with mutex-protected caching (5-minute expiry).
-   **Database:** SQLite via `github.com/mattn/go-sqlite3`. Database file is `src/admin.db` (note: also exists at root level).
-   **Article Storage:** Dual storage system:
    1.  **Markdown Files:** Detailed content in `articles/YYYY-MM-DD.md` format, parsed by optimized `parseMarkdownArticle` function.
    2.  **SQLite Database:** Metadata (slug, title, summary, content, created_at) in `articles` table for fast dashboard queries.

## 2. Key Components and Data Flows

-   **`src/main.go` (704 lines):**
    -   **Performance-Critical Functions:**
        -   `getTemplate()`: Thread-safe template caching with double-check locking pattern.
        -   `getCachedStockData()`: Markdown parsing cache with 5-minute expiry using mutex protection.
        -   `indexHandler()`: "SUPER FAST" database-only queries (no file system access during page load).
    -   **Data Structures:** `StockData` (complex morning/afternoon session data), `ArticlePreview` (index listings), `DBArticle` (database interaction), template data structs.
    -   **Caching Infrastructure:** Global `templateCache`, `markdownCache`, `cacheExpiry` maps with `sync.RWMutex` for thread safety.
    -   **Debug System:** `DEBUG_MODE` constant and `debugLog()` function for production performance.

-   **Specialized Markdown Parsing:**
    -   Extracts structured data from specific headers: `## Morning Session`, `### Open Set`, `### Open Analysis`, etc.
    -   Parses index values with regex: `(\d+\.?\d*)\s*\(([+-]?\d+\.?\d*)\)` pattern.
    -   Converts analysis sections to `template.HTML` for safe rendering.

-   **Database Patterns:**
    -   Auto-initialization with `initDB()` and `seedArticlesTable()`.
    -   Dynamic column addition (`content` column migration logic).
    -   `addMissingArticlesToDB()` syncs filesystem articles to database on startup.

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
    The server will start on `http://localhost:7777` (NOTE: Port changed from 8080 to 7777).

-   **Using VS Code Tasks:** A "Run Go server" task is configured for background execution.

-   **Performance Monitoring:**
    -   Set `DEBUG_MODE = true` in `src/main.go` for detailed logging.
    -   Template and markdown caches are crucial for performance - clearing them requires restart.
    -   Index page uses database-only queries for maximum speed.

-   **Database Interaction:**
    -   The `admin.db` SQLite database is automatically initialized and seeded on application startup if it doesn't exist or is empty.
    -   Database operations (CRUD for articles) are handled directly within the HTTP handlers in `src/main.go` using `database/sql` and `github.com/mattn/go-sqlite3`.
    -   `addMissingArticlesToDB()` automatically syncs filesystem articles to database on startup.

-   **Article Management (Admin Interface):**
    -   Access the admin dashboard at `http://localhost:7777/admin`.
    -   New articles can be added via `http://localhost:7777/admin/articles/new`.
    -   Existing articles can be edited via `http://localhost:7777/admin/articles/edit/{id}`.
    -   Creating new articles automatically generates markdown files with structured templates.

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
