# API Documentation

## Overview

ThaiStockAnalysis provides a RESTful API for accessing Thai stock market data and managing articles.

## Base URL

```
http://localhost:7777
```

## Authentication

Currently, the API does not require authentication for public endpoints. Admin endpoints will require authentication in future versions.

## Rate Limiting

No rate limiting is currently implemented, but it's recommended for production use.

## Response Format

All responses are in HTML format for web browser consumption. JSON API endpoints may be added in future versions.

## Endpoints

### Public Endpoints

#### GET /

**Description**: Homepage with latest stock market articles

**Response**: HTML page with article listings

**Example Response**: Homepage with responsive card layout showing recent articles

---

#### GET /articles/{slug}

**Description**: Individual article view with detailed stock data

**Parameters**:
- `slug` (string): Article slug in YYYY-MM-DD format

**Response**: HTML page with detailed stock analysis

**Example**: `/articles/2025-09-19`

---

#### GET /static/*

**Description**: Static assets (CSS, JS, images)

**Response**: Static file content with appropriate MIME type

---

### Admin Endpoints

#### GET /admin

**Description**: Admin dashboard with article management

**Response**: HTML admin interface

**Features**:
- Article listing with metadata
- Quick actions for editing/deleting
- Creation shortcuts

---

#### GET /admin/articles/new

**Description**: Article creation form

**Response**: HTML form for creating new articles

**Form Fields**:
- `slug`: Unique article identifier
- `title`: Article title
- `summary`: Brief description

---

#### POST /admin/articles/new

**Description**: Create new article

**Content-Type**: `application/x-www-form-urlencoded`

**Parameters**:
- `slug` (string): Unique identifier
- `title` (string): Article title  
- `summary` (string): Article summary

**Response**: Redirect to admin dashboard with success message

---

## Data Models

### StockData

```go
type StockData struct {
    CurrentDate              string        // Current date formatted
    MorningOpenIndex         float64       // Morning open index value
    MorningOpenChange        float64       // Morning open change
    MorningOpenHighlights    string        // Morning highlights
    MorningOpenAnalysis      template.HTML // Morning analysis HTML
    MorningCloseIndex        float64       // Morning close index
    MorningCloseChange       float64       // Morning close change
    MorningCloseHighlights   string        // Morning close highlights
    MorningCloseSummary      template.HTML // Morning summary HTML
    AfternoonOpenIndex       float64       // Afternoon open index
    AfternoonOpenChange      float64       // Afternoon open change
    AfternoonOpenHighlights  string        // Afternoon highlights
    AfternoonOpenAnalysis    template.HTML // Afternoon analysis HTML
    AfternoonCloseIndex      float64       // Afternoon close index
    AfternoonCloseChange     float64       // Afternoon close change
    AfternoonCloseHighlights string        // Afternoon close highlights
    AfternoonCloseSummary    template.HTML // Afternoon summary HTML
    KeyTakeaways             []string      // Important insights
}
```

### ArticlePreview

```go
type ArticlePreview struct {
    Title        string  // Article title
    Date         string  // Publication date
    SetIndex     string  // Latest SET index value
    Change       float64 // Index change
    ShortSummary string  // Brief summary
    Summary      string  // Full summary
    Slug         string  // URL slug
    URL          string  // Full article URL
}
```

### DBArticle

```go
type DBArticle struct {
    ID        int               // Database ID
    Slug      string            // URL slug
    Title     string            // Article title
    Summary   sql.NullString    // Article summary
    Content   sql.NullString    // Article content
    CreatedAt string            // Creation date
}
```

## Error Responses

### 404 Not Found

Returned when an article slug doesn't exist.

### 500 Internal Server Error

Returned when there's a server-side error (database issues, template errors, etc.).

### 400 Bad Request

Returned when form data is malformed or missing required fields.

## Performance Considerations

### Caching

- **Template Caching**: Templates are cached in memory with mutex protection
- **Markdown Caching**: Parsed markdown files are cached for 5 minutes
- **Database Optimization**: Efficient queries with prepared statements

### Response Times

- Homepage: ~10-50ms (cached)
- Article pages: ~20-100ms (cached)
- Admin dashboard: ~30-150ms

### Resource Usage

- Memory: ~10-50MB typical usage
- CPU: Low usage with caching
- Disk: SQLite database + static files

## Future API Enhancements

### Planned JSON API

```
GET /api/v1/articles           # List articles
GET /api/v1/articles/{slug}    # Get article data
POST /api/v1/articles          # Create article (admin)
PUT /api/v1/articles/{slug}    # Update article (admin)
DELETE /api/v1/articles/{slug} # Delete article (admin)
```

### Authentication

- JWT-based authentication for admin endpoints
- Role-based access control
- API key authentication for external integrations

### Webhooks

- Real-time notifications for new articles
- Stock data update notifications
- Admin action logging

## Examples

### Getting Article Data

```bash
# Get homepage
curl http://localhost:7777

# Get specific article
curl http://localhost:7777/articles/2025-09-19

# Get admin dashboard
curl http://localhost:7777/admin
```

### Creating an Article

```bash
# Create new article via form
curl -X POST http://localhost:7777/admin/articles/new \
  -d "slug=2025-09-25" \
  -d "title=Market Analysis September 25" \
  -d "summary=Daily stock market analysis"
```

## SDK/Client Libraries

Currently, no official SDK is available. The API is designed for browser consumption with HTML responses.

Future SDKs planned:
- Go client library
- JavaScript/TypeScript client
- Python client library

## Testing

### Manual Testing

```bash
# Test homepage
curl -I http://localhost:7777

# Test article page
curl -I http://localhost:7777/articles/2025-09-19

# Test static assets
curl -I http://localhost:7777/static/css/style.css
```

### Load Testing

```bash
# Install hey for load testing
go install github.com/rakyll/hey@latest

# Test homepage performance
hey -n 1000 -c 10 http://localhost:7777

# Test article performance
hey -n 1000 -c 10 http://localhost:7777/articles/2025-09-19
```

## Security Considerations

### Input Validation

- All form inputs are validated
- SQL injection protection via parameterized queries
- XSS prevention through template escaping

### HTTPS

For production deployment:
- Enable HTTPS/TLS encryption
- Use secure cookies
- Implement HSTS headers

### Content Security Policy

Recommended CSP headers for production:

```
Content-Security-Policy: default-src 'self'; 
                        style-src 'self' 'unsafe-inline'; 
                        script-src 'self';
                        img-src 'self' data:;
```