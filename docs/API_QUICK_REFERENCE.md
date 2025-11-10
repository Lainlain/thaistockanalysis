# API Quick Reference - Thai Stock Analysis

## Base URL
```
http://localhost:7777
```

---

## 1️⃣ Update Opening Data

**Endpoint:** `POST /api/market-data-analysis`

### Morning Opening
```bash
curl -X POST http://localhost:7777/api/market-data-analysis \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2025-11-06",
    "morning_open": {
      "index": 1305.23,
      "change": 2.56,
      "highlights": "<strong>Banking +4.2%</strong>"
    }
  }'
```

### Afternoon Opening
```bash
curl -X POST http://localhost:7777/api/market-data-analysis \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2025-11-06",
    "afternoon_open": {
      "index": 1287.01,
      "change": -4.47,
      "highlights": "<strong>Energy -2.8%</strong>"
    }
  }'
```

---

## 2️⃣ Update Closing Data

**Endpoint:** `POST /api/market-data-close`

### Morning Closing
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

### Afternoon Closing
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

## 📋 Field Reference

| Field | Type | Required | Example |
|-------|------|----------|---------|
| date | string | ✅ | "2025-11-06" |
| index | float | ✅ | 1305.23 |
| change | float | ✅ | 2.56 or -4.47 |
| highlights | string | ❌ | "<strong>Banking +4.2%</strong>" |

---

## 🔄 Typical Workflow

```bash
# 1. Morning Open
POST /api/market-data-analysis
{
  "date": "2025-11-06",
  "morning_open": { "index": 1305.23, "change": 2.56, "highlights": "..." }
}

# 2. Morning Close
POST /api/market-data-close
{
  "date": "2025-11-06",
  "morning_close": { "index": 1308.45, "change": 5.78 }
}

# 3. Afternoon Open
POST /api/market-data-analysis
{
  "date": "2025-11-06",
  "afternoon_open": { "index": 1287.01, "change": -4.47, "highlights": "..." }
}

# 4. Afternoon Close
POST /api/market-data-close
{
  "date": "2025-11-06",
  "afternoon_close": { "index": 1291.23, "change": -0.32 }
}
```

---

## ✅ Success Response
```json
{
  "status": "success",
  "message": "Analysis generated and saved successfully",
  "date": "2025-11-06"
}
```

---

## ❌ Error Codes

| Code | Meaning |
|------|---------|
| 400 | Invalid JSON format |
| 405 | Wrong HTTP method (use POST) |
| 500 | Server error (check logs) |

---

## 📝 Highlights Format

```html
<strong>Sector Name: +X.X%</strong> <br><br> <strong>Another Sector: -X.X%</strong>
```

**Example:**
```html
<strong>Banking sector gains +4.2%</strong> <br><br> <strong>Technology stocks up +3.1%</strong>
```

---

## 🚀 Full Day Update Script

```bash
#!/bin/bash
DATE="2025-11-06"
API="http://localhost:7777"

# Morning Open
curl -X POST $API/api/market-data-analysis -H "Content-Type: application/json" \
-d '{"date":"'$DATE'","morning_open":{"index":1305.23,"change":2.56,"highlights":"<strong>Banking +4.2%</strong>"}}'

# Morning Close
curl -X POST $API/api/market-data-close -H "Content-Type: application/json" \
-d '{"date":"'$DATE'","morning_close":{"index":1308.45,"change":5.78}}'

# Afternoon Open
curl -X POST $API/api/market-data-analysis -H "Content-Type: application/json" \
-d '{"date":"'$DATE'","afternoon_open":{"index":1287.01,"change":-4.47,"highlights":"<strong>Energy -2.8%</strong>"}}'

# Afternoon Close
curl -X POST $API/api/market-data-close -H "Content-Type: application/json" \
-d '{"date":"'$DATE'","afternoon_close":{"index":1291.23,"change":-0.32}}'

echo "✅ Full day updated for $DATE"
```

---

## 🔍 Check Results

**View Article:**
```
http://localhost:7777/article/2025-11-06
```

**Admin Dashboard:**
```
http://localhost:7777/admin
```

**Check Logs:**
```bash
tail -f server.log
```

---

## 💡 Pro Tips

1. ✅ Use YYYY-MM-DD date format
2. ✅ Update sessions in chronological order
3. ✅ Keep highlights concise (2-3 sectors)
4. ✅ Check response status before proceeding
5. ✅ Test with curl before automation

---

For detailed documentation, see: `docs/API_USAGE.md`
