# ThaiStockAnalysis - AI Agent Instructions

This document provides essential knowledge for AI coding agents to be immediately productive in the `ThaiStockAnalysis` Go codebase.

## 1. Big Picture Architecture

This is a high-performance Go web server for Thai stock market analysis with a **dual architecture**:

- **Modern Modular Architecture** (`cmd/server/main.go`): Clean separation of concerns with proper dependency injection
- **Legacy Monolithic Architecture** (`src/main.go`): Original 2,342-line implementation (avoid for new features)

**ALWAYS use `cmd/server/main.go` entry point** - the modular structure with handlers, services, database, and models layers.

**Core Tech Stack:**
- Go 1.24.6 with stdlib `net/http` (no external routers)
- SQLite via `github.com/mattn/go-sqlite3` (**CGO required** - build with `CGO_ENABLED=1`)
- Go `html/template` + Tailwind CSS for frontend
- Dual storage: Markdown files (`articles/*.md`) + SQLite metadata (`data/admin.db`)
- Google Gemini AI `gemini-2.0-flash-lite-001` (v1beta endpoint) for market analysis with retry logic
- Telegram Bot for notifications

## 2. Critical Data Structures & Parsing Logic

### StockData Model (4 Sessions)
Each trading day has **four distinct session states** parsed from markdown:
- **Morning Open**: Index, Change, Highlights, Analysis (template.HTML)
- **Morning Close**: Index, Change, Highlights, Summary (template.HTML)
- **Afternoon Open**: Index, Change, Highlights, Analysis (template.HTML)
- **Afternoon Close**: Index, Change, Highlights, Summary (template.HTML)

### Markdown Parsing Convention (internal/services/services.go:80-160)
The parser expects this **exact structure** in `articles/YYYY-MM-DD.md`:
```markdown
# Stock Market Analysis - 30 September 2025

## Morning Session
### Open Set
* Open Index: 1302.75 (+16.49)
* Highlights: Energy firms rally eight points as oil prices spike.
### Open Analysis
<p>HTML analysis content...</p>
### Close Set
* Close Index: 1280.38 (-7.69)
### Close Summary
<p>Summary content...</p>

## Afternoon Session
### Open Set
* Open Index: 1279.48 (-8.59)
* Highlights: **+94 +97 +90...**
### Open Analysis
<p>Analysis...</p>
```

**Parser supports dual formats**: Old (`### Open Set`) and new (`### Market Opening Data`) for backwards compatibility.

### Index Value Extraction
Regex pattern: `(\d+\.?\d*)\s*\(([+-]?\d+\.?\d*)\)` parses "1295.80 (+5.15)" → index: 1295.80, change: +5.15

### Highlight Number-to-Narrative Conversion
`PromptService.GenerateHighlightNarrative()` converts raw numbers like "+68 +61 +64" to human-readable sector insights:
1. Extract **first number** from highlights string
2. Use **last digit** as key in `highlights_for_prompt.json`
3. Randomly select phrase from mapped array
4. Example: "+68" → last digit "8" → "Energy sector rallies eight points on rising oil futures."

## 3. Developer Workflows

### Running the Application
```bash
# Recommended: Modern modular architecture
go run cmd/server/main.go

# Access points:
# - Homepage: http://localhost:7777
# - Admin: http://localhost:7777/admin
# - API: http://localhost:7777/api/market-data-analysis
```

### VS Code Tasks (Use These, Not Manual Commands)
- **"Run Go Server"** (`isBackground: true`) - Use for development server
- **"Build Binary"** - Creates `bin/thaistockanalysis` executable
- **"Test All"** / **"Test with Coverage"** - Run full test suite
- **"Docker Build"** / **"Docker Run"** - Containerized deployment with nginx reverse proxy

### Environment Configuration Pattern
All config in `configs/config.go` with **environment variable overrides**:
```go
config := &Config{
    Port:             getEnv("PORT", "7777"),
    CacheExpiry:      getEnvInt("CACHE_EXPIRY", 0), // 0 = disabled
    GeminiAPIKey:     getEnv("GEMINI_API_KEY", "AIzaSy..."), // ⚠️ Hardcoded default
}
```

**Security Critical**: `GEMINI_API_KEY`, `TELEGRAM_BOT_TOKEN`, `TELEGRAM_CHANNEL` have hardcoded defaults. Production **MUST** override via environment variables.

### Docker Deployment with CGO
Multi-stage Dockerfile handles SQLite's CGO requirement:
```bash
docker build -t thaistockanalysis .     # Uses alpine + gcc/musl-dev
docker-compose up -d                     # Includes nginx, backup service
```

## 4. Project-Specific Patterns

### Article Lifecycle & Filesystem Sync
1. Articles stored as `articles/YYYY-MM-DD.md` (e.g., `2025-09-30.md`)
2. On server startup: `database.AddMissingArticlesToDB(cfg.ArticlesDir)` auto-syncs filesystem → SQLite
3. Database stores metadata only (slug, title, summary) - **not full markdown content**
4. Homepage queries database for performance (avoids parsing all markdown files)
5. Article detail page parses markdown on-demand via `MarkdownService.GetCachedStockData()`

### Caching Strategy (Thread-Safe)
```go
// Global caches in internal/services/services.go
var (
    markdownCache = make(map[string]models.StockData) // Protected by cacheMutex
    templateCache = make(map[string]*template.Template) // Protected by templateMutex
)
```

**Cache Expiry Behavior** (`CACHE_EXPIRY` in minutes):
- `0` (default): Cache **disabled**, always parses fresh markdown
- `>0`: Cache enabled with TTL, check expiry before returning cached data
- Clear cache: `MarkdownService.ClearCache(filePath)` for targeted invalidation

### Gemini AI Integration Pattern (handlers.go:558-646)
**Retry logic with fallback**:
1. Attempt API call with `gemini-2.0-flash-lite-001` model (faster than standard)
2. On failure (network, 429 rate limit, quota): Retry up to 2 times with 15s/25s delays
3. Final fallback: `generateMockGeminiResponse()` creates data-driven mock from prompt content
4. **Never throws errors** - always returns usable content for article generation

API Request Structure:
```go
POST https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash-lite-001:generateContent?key={apiKey}
{
  "contents": [{"role": "user", "parts": [{"text": "{prompt}"}]}]
}
```

### API Endpoints for Market Data
**`POST /api/market-data-analysis`**: Updates opening data (morning/afternoon)
**`POST /api/market-data-close`**: Updates closing data with session summary
Both endpoints:
1. Load existing markdown file for the date
2. Generate AI analysis via Gemini
3. Update specific session section
4. Write back to `articles/YYYY-MM-DD.md`
5. Send Telegram notification

See `docs/API_QUICK_REFERENCE.md` for curl examples.

## 5. Critical Integration Points

### Template Function Map (services.go:290+)
Custom functions available in all templates:
```go
funcMap := template.FuncMap{
    "printf":         fmt.Sprintf,
    "html":           func(s string) template.HTML { return template.HTML(s) },
    "add":            func(a, b int) int { return a + b },
    "markdownToHTML": func(s string) template.HTML {...},
}
```

### Database Auto-Migration (database.go:15-60)
On `InitDB()`:
1. Creates `data/` directory if missing
2. Creates `articles` table with schema
3. Checks for `content` column, adds via `ALTER TABLE` if missing
4. Seeds with sample articles if empty
5. **Does NOT delete existing data**

### HTTP Handler Initialization Pattern
All handlers require full service injection:
```go
h := handlers.NewHandler(cfg.ArticlesDir, cfg.TemplateDir, cfg)
// h.MarkdownService, h.TemplateService, h.PromptService, h.Config all initialized
```

### Static Assets Routing
```go
mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(cfg.StaticDir))))
// Serves: web/static/css/*.css, web/static/js/*.js, web/static/*.png
```

## 6. Debugging & Testing

### Debug Artifacts in Repository
- `debug_parser.go`, `debug_template.go` - Standalone component testing
- `test_new_format.go` - Validates markdown parsing logic
- `server*.log` files - Various troubleshooting logs (git-ignored recommended)
- `articles_backup_*/` - Historical data for regression testing

### Common Build Issues
**Problem**: `undefined reference to sqlite3_*`
**Solution**: Ensure `CGO_ENABLED=1` and gcc installed: `go build -tags=cgo`

**Problem**: Template not updating after changes
**Solution**: Clear template cache or restart server (cache is global variable)

**Problem**: Article parsing fails silently
**Solution**: Check markdown structure matches exact format in section 2 (especially `### Open Set` vs `### Open Analysis`)

---

**Architecture Rule**: Always extend the modular `cmd/server/` structure. The `src/main.go` legacy code is for reference only.
