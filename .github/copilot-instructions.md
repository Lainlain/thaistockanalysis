# ThaiStockAnalysis - AI Agent Instructions

This document provides essential knowledge for AI coding agents to be immediately productive in the `ThaiStockAnalysis` Go codebase.

## 1. Big Picture Architecture

The application is a high-performance Go web server providing Thai stock market analysis with a **dual architecture**:

- **Modern Modular Architecture** (`cmd/server/main.go`): Clean separation of concerns with proper dependency injection
- **Legacy Monolithic Architecture** (`src/main.go`): Original implementation with all logic centralized (2,300+ lines)

**Current Recommended Entry Point**: `cmd/server/main.go` with modular structure

**Core Technologies:**
- **Language**: Go 1.24.6 with `net/http` routing
- **Database**: SQLite via `github.com/mattn/go-sqlite3`, located at `data/admin.db`
- **Templates**: Go's `html/template` with Tailwind CSS in `web/templates/`
- **Caching**: Thread-safe template/markdown caches with configurable expiry (default 5 minutes)
- **Article Storage**: Dual system (markdown files + SQLite metadata)
- **External APIs**: Google Gemini AI integration and Telegram Bot notifications

## 2. Key Components and Data Flow

### Modern Architecture (`cmd/server/main.go`)
- **Entry Point**: `cmd/server/main.go` - Clean server initialization with graceful shutdown
- **Handlers**: `internal/handlers/handlers.go` - HTTP request handling with dependency injection
- **Services**: `internal/services/services.go` - Business logic (MarkdownService, TemplateService)
- **Database**: `internal/database/database.go` - Database operations and migrations
- **Models**: `internal/models/models.go` - Data structures (`StockData`, `ArticlePreview`, `DBArticle`)
- **Config**: `configs/config.go` - Environment-based configuration management

### Specialized Data Structures
- **StockData**: Complex morning/afternoon session data with index values and HTML analysis
- **ArticlePreview**: Summary view for homepage listings with cached SET index data
- **Markdown Parsing**: Extracts structured data from headers (`## Morning Session`, `### Open Set`)
- **Index Parsing**: Regex pattern `(\d+\.?\d*)\s*\(([+-]?\d+\.?\d*)\)` for stock values

## 3. Critical Developer Workflows

### Running the Application
```bash
# Modern architecture (recommended)
go run cmd/server/main.go

# Legacy architecture (still functional)
go run src/main.go
```

### VS Code Tasks (Pre-configured)
- **"Run Go Server"**: Background server execution
- **"Build Binary"**: Creates `bin/thaistockanalysis` executable
- **"Test All"** & **"Test with Coverage"**: Full test suite execution
- **"Docker Build"** & **"Docker Run"**: Containerized deployment

### Environment Configuration
Configure via environment variables or defaults in `configs/config.go`:
- `PORT=7777` (default)
- `DATABASE_PATH=data/admin.db`
- `ARTICLES_DIR=articles`
- `TEMPLATE_DIR=web/templates`
- `STATIC_DIR=web/static`
- `DEBUG_MODE=false`
- `CACHE_EXPIRY=5` (minutes)
- `GEMINI_API_KEY` (for AI analysis, has hardcoded default)
- `TELEGRAM_BOT_TOKEN` (for notifications, has hardcoded default)
- `TELEGRAM_CHANNEL` (target channel ID, has hardcoded default)

**Security Note**: Production deployments should override API keys via environment variables.

### Docker Deployment
```bash
# Multi-stage build with SQLite support
docker build -t thaistockanalysis .
docker-compose up -d
```

## 4. Project-Specific Conventions

### Article Management
- **File Format**: `articles/YYYY-MM-DD.md` (e.g., `2025-09-26.md`)
- **Auto-Sync**: `database.AddMissingArticlesToDB()` syncs filesystem to database on startup
- **Admin Interface**: `/admin` for CRUD operations, auto-generates structured markdown

### Markdown Structure (Critical for Parsing)
```markdown
# Stock Market Analysis - DD Month YYYY
## Morning Session
### Open Set
* Open Index: 1295.80 (+5.15)
* Highlights: **Sector info**
### Open Analysis
<p>HTML analysis content</p>
## Afternoon Session
### Open Set
* Open Index: 1287.01 (-4.47)
```

### Caching Strategy
- **Template Cache**: `sync.RWMutex` protected, global scope
- **Markdown Cache**: Configurable expiry (default 5 minutes) with mutex protection
- **Cache Management**: `ClearCache(filePath)` for targeted invalidation
- **Performance**: Database-only index queries avoid filesystem I/O

### Debugging and Development
- **Debug Files**: `debug_parser.go` and `debug_template.go` for standalone testing
- **Server Logs**: Multiple log files (`server.log`, `server_test.log`, `server_clean.log`) for troubleshooting
- **Market Data Testing**: `test_new_format.go` for validating parsing logic
- **Article Backup**: `articles_backup_*/` directories contain historical data for testing
- **Standalone Testing**: Root-level test files (`test_server.log`, `test_template.go`) for component isolation
- **Article Backup**: `articles_backup_*/` directories contain historical data for testing
- **Standalone Testing**: Root-level test files (`test_server.log`, `test_template.go`) for component isolation

## 5. Integration Points

### Dependencies (`go.mod`)
- `github.com/gomarkdown/markdown`: Markdown to HTML conversion
- `github.com/mattn/go-sqlite3`: SQLite database driver (requires CGO)

### API Endpoints
- `/api/market-data-analysis`: Gemini AI-powered market analysis
- `/api/market-data-close`: Market closing data processing
- `/admin/articles/new`: Structured article creation form

### Static Assets
- **Location**: `web/static/` (modern) or `src/static/` (legacy)
- **Tailwind CSS**: Pre-built styles in `/static/css/`
- **Responsive Design**: Mobile-first approach with card layouts

### Development vs Production
- **Debug Mode**: Environment variable `DEBUG_MODE=true`
- **Database**: Auto-migration and seeding on startup
- **Port**: Default 7777 (changed from 8080)
- **Docker**: Production-ready with health checks and volume mounts

---

**Architecture Migration Note**: When making changes, prefer the modular architecture in `cmd/server/` over the legacy `src/main.go` approach.
