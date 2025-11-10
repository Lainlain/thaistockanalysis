# Thai Stock Analysis - API Usage Documentation

## Overview

This document provides comprehensive documentation for the Thai Stock Analysis API endpoints used to update market data. The API uses AI-powered analysis (Google Gemini) to generate professional market insights.

---

## Base URL

```
http://localhost:7777
```

For production, replace with your actual domain.

---

## Authentication

Currently, the API does not require authentication. Consider implementing API keys or OAuth for production use.

---

## API Endpoints

### 1. Market Opening Analysis

**Endpoint:** `/api/market-data-analysis`

**Method:** `POST`

**Description:** Creates or updates market analysis for opening sessions (morning or afternoon). Uses AI to generate professional analysis based on index data and sector highlights.

#### Request Headers

```
Content-Type: application/json
```

#### Request Body Schema

```json
{
  "date": "string (YYYY-MM-DD format, required)",
  "morning_open": {
    "index": "float (required)",
    "change": "float (required)",
    "highlights": "string (optional)"
  },
  "afternoon_open": {
    "index": "float (required)",
    "change": "float (required)",
    "highlights": "string (optional)"
  }
}
```

**Note:** You can send either `morning_open` OR `afternoon_open`, or both in a single request.

#### Request Examples

##### Morning Session Opening

```bash
curl -X POST http://localhost:7777/api/market-data-analysis \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2025-11-06",
    "morning_open": {
      "index": 1305.23,
      "change": 2.56,
      "highlights": "<strong>Banking sector gains +4.2%</strong> <br><br> <strong>Technology stocks up +3.1%</strong>"
    }
  }'
```

##### Afternoon Session Opening

```bash
curl -X POST http://localhost:7777/api/market-data-analysis \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2025-11-06",
    "afternoon_open": {
      "index": 1287.01,
      "change": -4.47,
      "highlights": "<strong>Energy sector leads decline -2.8%</strong> <br><br> <strong>Property stocks down -1.9%</strong>"
    }
  }'
```

##### Both Sessions (Full Day Update)

```bash
curl -X POST http://localhost:7777/api/market-data-analysis \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2025-11-06",
    "morning_open": {
      "index": 1305.23,
      "change": 2.56,
      "highlights": "<strong>Banking sector gains +4.2%</strong>"
    },
    "afternoon_open": {
      "index": 1287.01,
      "change": -4.47,
      "highlights": "<strong>Energy sector decline -2.8%</strong>"
    }
  }'
```

#### Response

**Success (200 OK):**

```json
{
  "status": "success",
  "message": "Analysis generated and saved successfully",
  "date": "2025-11-06"
}
```

**Error Responses:**

```json
// 400 Bad Request - Invalid JSON
{
  "error": "Invalid JSON"
}

// 405 Method Not Allowed
{
  "error": "Method not allowed"
}

// 500 Internal Server Error
{
  "error": "Error saving analysis"
}
```

---

### 2. Market Closing Summary

**Endpoint:** `/api/market-data-close`

**Method:** `POST`

**Description:** Updates market closing data for morning and/or afternoon sessions. Generates AI-powered summary comparing opening vs closing performance.

#### Request Headers

```
Content-Type: application/json
```

#### Request Body Schema

```json
{
  "date": "string (YYYY-MM-DD format, required)",
  "morning_close": {
    "index": "float (required)",
    "change": "float (required)"
  },
  "afternoon_close": {
    "index": "float (required)",
    "change": "float (required)"
  }
}
```

**Note:** You can send either `morning_close` OR `afternoon_close`, or both.

#### Request Examples

##### Morning Session Closing

```bash
curl -X POST http://localhost:7777/api/market-data-close \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2025-11-06",
    "morning_close": {
      "index": 1308.45,
      "change": 5.78
    }
  }'
```

##### Afternoon Session Closing (End of Day)

```bash
curl -X POST http://localhost:7777/api/market-data-close \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2025-11-06",
    "afternoon_close": {
      "index": 1291.23,
      "change": -0.32
    }
  }'
```

##### Both Sessions

```bash
curl -X POST http://localhost:7777/api/market-data-close \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2025-11-06",
    "morning_close": {
      "index": 1308.45,
      "change": 5.78
    },
    "afternoon_close": {
      "index": 1291.23,
      "change": -0.32
    }
  }'
```

#### Response

**Success (200 OK):**

```json
{
  "status": "success",
  "message": "Summary generated and saved successfully",
  "date": "2025-11-06"
}
```

**Error Responses:**

```json
// 400 Bad Request - Invalid JSON
{
  "error": "Invalid JSON"
}

// 405 Method Not Allowed
{
  "error": "Method not allowed"
}

// 500 Internal Server Error
{
  "error": "Error saving summary"
}
```

---

## Complete Workflow Example

### Full Day Market Data Update

Here's a complete workflow for updating a full trading day:

#### Step 1: Morning Opening
```bash
curl -X POST http://localhost:7777/api/market-data-analysis \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2025-11-06",
    "morning_open": {
      "index": 1305.23,
      "change": 2.56,
      "highlights": "<strong>Banking sector +4.2%</strong> <br><br> <strong>Tech stocks +3.1%</strong>"
    }
  }'
```

#### Step 2: Morning Closing
```bash
curl -X POST http://localhost:7777/api/market-data-close \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2025-11-06",
    "morning_close": {
      "index": 1308.45,
      "change": 5.78
    }
  }'
```

#### Step 3: Afternoon Opening
```bash
curl -X POST http://localhost:7777/api/market-data-analysis \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2025-11-06",
    "afternoon_open": {
      "index": 1287.01,
      "change": -4.47,
      "highlights": "<strong>Energy sector -2.8%</strong> <br><br> <strong>Property -1.9%</strong>"
    }
  }'
```

#### Step 4: Afternoon Closing (End of Day)
```bash
curl -X POST http://localhost:7777/api/market-data-close \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2025-11-06",
    "afternoon_close": {
      "index": 1291.23,
      "change": -0.32
    }
  }'
```

---

## Field Descriptions

### Common Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `date` | string | Yes | Trading date in YYYY-MM-DD format (e.g., "2025-11-06") |

### Market Session Fields (Opening)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `index` | float | Yes | Current SET index value (e.g., 1305.23) |
| `change` | float | Yes | Point change from previous close (use + or - prefix, e.g., 2.56, -4.47) |
| `highlights` | string | Optional | HTML-formatted sector highlights. Use `<strong>` and `<br>` tags for formatting |

### Market Session Fields (Closing)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `index` | float | Yes | Closing SET index value (e.g., 1308.45) |
| `change` | float | Yes | Total point change for the day (use + or -, e.g., 5.78, -0.32) |

---

## Highlights Formatting Guidelines

The `highlights` field supports HTML formatting for rich content display:

### Recommended Format

```html
<strong>Sector Name: Performance</strong> <br><br> <strong>Another Sector: Performance</strong>
```

### Examples

**Single Sector:**
```html
<strong>Banking sector gains +4.2%</strong>
```

**Multiple Sectors:**
```html
<strong>Banking sector +4.2%</strong> <br><br> <strong>Technology stocks +3.1%</strong> <br><br> <strong>Energy sector -1.5%</strong>
```

**With Additional Details:**
```html
<strong>Banking sector leads with +4.2% on strong loan growth</strong> <br><br> <strong>Technology stocks rally +3.1% on AI momentum</strong>
```

---

## Generated Markdown Structure

After calling the APIs, a markdown file is created/updated at `articles/{date}.md` with this structure:

```markdown
# Stock Market Analysis - DD Month YYYY

## Morning Session

### Open Set
* Open Index: 1305.23 (+2.56)
* Highlights: **Banking sector +4.2%** **Tech stocks +3.1%**

### Open Analysis
<p>[AI-generated professional analysis]</p>

### Close Set
* Close Index: 1308.45 (+5.78)

### Close Summary
<p>[AI-generated session summary comparing open vs close]</p>

## Afternoon Session

### Open Set
* Open Index: 1287.01 (-4.47)
* Highlights: **Energy sector -2.8%** **Property -1.9%**

### Open Analysis
<p>[AI-generated professional analysis]</p>

### Close Set
* Close Index: 1291.23 (-0.32)

### Close Summary
<p>[AI-generated session summary]</p>

## Key Takeaways
- [Daily summary point 1]
- [Daily summary point 2]
- [Daily summary point 3]
```

---

## Database Storage

All articles are automatically stored in the SQLite database (`data/admin.db`) with metadata:

- **slug**: Date-based identifier (e.g., "2025-11-06")
- **title**: Auto-generated title (e.g., "Stock Market Analysis - 6 November 2025")
- **created_at**: Article creation timestamp
- **updated_at**: Last modification timestamp

---

## AI Integration

The API uses **Google Gemini AI** to generate:

1. **Opening Analysis**: Professional market analysis based on index movement and sector highlights
2. **Closing Summary**: Comparative analysis of session performance (open vs close)
3. **Key Takeaways**: End-of-day summary with actionable insights (afternoon close only)

### AI Prompt Templates

Templates are loaded from:
- `getanalysis_prompt_human.txt` - Opening analysis
- `getanalysis_prompt_close.txt` - Closing summary

---

## Error Handling

### Common Errors and Solutions

| Error | Cause | Solution |
|-------|-------|----------|
| Invalid JSON | Malformed request body | Validate JSON syntax before sending |
| Method not allowed | Using GET instead of POST | Use POST method |
| Error saving analysis | File system or database issue | Check server logs, verify write permissions |
| No session data | Missing morning_open/afternoon_open | Include at least one session in request |

---

## Rate Limiting

**Current Status:** No rate limiting implemented

**Recommendation for Production:**
- Implement API key authentication
- Add rate limiting (e.g., 100 requests/hour per key)
- Monitor Gemini AI API quota

---

## Best Practices

1. **Sequential Updates**: Update sessions in chronological order (morning open → morning close → afternoon open → afternoon close)

2. **Date Format**: Always use YYYY-MM-DD format (ISO 8601)

3. **Highlights**: Keep highlights concise and focused on major sectors

4. **Error Checking**: Always check response status before proceeding

5. **Idempotency**: Calling the same endpoint with the same data is safe - it will update the existing article

---

## Testing

### Quick Test Script (Bash)

```bash
#!/bin/bash

DATE="2025-11-06"
API_URL="http://localhost:7777"

# Morning Opening
echo "📊 Updating Morning Open..."
curl -X POST $API_URL/api/market-data-analysis \
  -H "Content-Type: application/json" \
  -d "{
    \"date\": \"$DATE\",
    \"morning_open\": {
      \"index\": 1305.23,
      \"change\": 2.56,
      \"highlights\": \"<strong>Banking +4.2%</strong> <br><br> <strong>Tech +3.1%</strong>\"
    }
  }"

echo -e "\n\n✅ Morning Open Updated\n"

# Morning Close
echo "📊 Updating Morning Close..."
curl -X POST $API_URL/api/market-data-close \
  -H "Content-Type: application/json" \
  -d "{
    \"date\": \"$DATE\",
    \"morning_close\": {
      \"index\": 1308.45,
      \"change\": 5.78
    }
  }"

echo -e "\n\n✅ Full Morning Session Complete\n"
```

---

## Monitoring and Logs

Server logs will show:

```
📊 Market Analysis Request for 2025-11-06
✅ Analysis generated and saved successfully
📊 Market Close Request for 2025-11-06
✅ Summary generated and saved successfully
```

Check logs for debugging:
```bash
tail -f server.log
```

---

## Support

For issues or questions:
- **Email**: thaistockanalysis@lainlain.online
- **GitHub**: https://github.com/Lainlain/thaistockanalysis

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2025-11-06 | Initial API documentation |

---

## License

This API is part of the Thai Stock Analysis project.
All rights reserved © 2025 Thai Stock Analysis
