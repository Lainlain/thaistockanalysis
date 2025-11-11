# 🎯 Vue Production Build - Local Testing Guide

**Issue**: Getting 404 error when testing production build locally

---

## 🐛 The Problem

When you build with `npm run build` and test the `dist/` folder locally, you get:
```
AxiosError: Request failed with status code 404
```

**Why?**
- Production build uses: `https://thaistockanalysis.com/api`
- Your local machine can't access internal Go server on production domain
- You need the Go server running locally at `localhost:7777`

---

## ✅ Solution: Two Testing Approaches

### Approach 1: Test with Local Go Server (Recommended)

**Step 1: Create `.env.local`**

Already created: `vue/.env.local`
```bash
VITE_API_URL=http://localhost:7777/api
```

**Step 2: Start Go Server**
```bash
# Terminal 1: Start Go server
cd "/home/lainlain/Desktop/Go Lang /ThaiStockAnalysis/ThaiStockAnalysis (copy)"
go run cmd/server/main.go

# Should see:
# 🚀 ThaiStockAnalysis server starting on http://localhost:7777
```

**Step 3: Build Vue App**
```bash
# Terminal 2: Build with local API
cd vue/
rm -rf dist
npm run build
```

**Step 4: Serve and Test**
```bash
# Still in vue/
cd dist/
python3 -m http.server 8080

# Open browser: http://localhost:8080
```

**Now it works!** ✅
- Vue app: `http://localhost:8080`
- API calls: `http://localhost:7777/api`
- Both running on your local machine

---

### Approach 2: Test with Production API (No Local Server)

If you want to test against the real production API:

**Step 1: Remove `.env.local`**
```bash
cd vue/
rm .env.local  # Remove local override
```

**Step 2: Build with Production API**
```bash
npm run build
# Uses .env.production → https://thaistockanalysis.com/api
```

**Step 3: Test (API calls go to production)**
```bash
cd dist/
python3 -m http.server 8080

# Open: http://localhost:8080
# API calls will go to: https://thaistockanalysis.com/api ✅
```

**Note**: This requires:
- Production server is running
- CORS configured to allow `localhost:8080` origin
- You have internet connection

---

## 📋 Environment Files Priority

Vite loads environment files in this order (highest priority first):

1. **`.env.local`** ← Always loaded, git-ignored (LOCAL TESTING)
2. **`.env.production`** ← Loaded during `npm run build`
3. **`.env.development`** ← Loaded during `npm run dev`
4. **`.env`** ← Always loaded (base config)

**Current Setup**:
```
vue/
├── .env.local           # VITE_API_URL=http://localhost:7777/api
├── .env.development     # VITE_API_URL=http://localhost:7777/api
└── .env.production      # VITE_API_URL=https://thaistockanalysis.com/api
```

---

## 🧪 Testing Scenarios

### Scenario 1: Local Development (`npm run dev`)
```bash
npm run dev

# Uses: .env.development
# API: http://localhost:7777/api
# Proxy: Yes (Vite dev server)
# Requires: Go server running locally
```

### Scenario 2: Local Production Test (`npm run build` + `.env.local`)
```bash
npm run build
cd dist/
python3 -m http.server 8080

# Uses: .env.local (overrides .env.production)
# API: http://localhost:7777/api
# Proxy: No (static files)
# Requires: Go server running locally
```

### Scenario 3: Production Build Test with Real API
```bash
rm .env.local  # Remove local override
npm run build
cd dist/
python3 -m http.server 8080

# Uses: .env.production
# API: https://thaistockanalysis.com/api
# Proxy: No
# Requires: Production server running + CORS
```

### Scenario 4: Real Production Deployment
```bash
rm .env.local  # Remove local override
npm run build

# Upload dist/ to production server
# Uses: .env.production
# API: https://thaistockanalysis.com/api
# No local server needed
```

---

## 🔧 Quick Fix Commands

### To Test Locally (with local Go server):
```bash
# 1. Ensure .env.local exists
cat vue/.env.local
# Should see: VITE_API_URL=http://localhost:7777/api

# 2. Start Go server
go run cmd/server/main.go &

# 3. Build and test
cd vue/
rm -rf dist && npm run build
cd dist/
python3 -m http.server 8080

# Open: http://localhost:8080
```

### To Build for Production (no .env.local):
```bash
# 1. Remove local override
rm vue/.env.local

# 2. Build
cd vue/
rm -rf dist && npm run build

# 3. Upload dist/ to server
scp -r dist/* user@server:/var/www/html/
```

---

## 🐞 Troubleshooting

### Error: "404 Not Found"

**Check 1**: Which API URL is in the build?
```bash
cd vue/dist/
grep -r "localhost:7777" assets/*.js
# Found? → Using local API ✅
# Not found? → Using production API

grep -r "thaistockanalysis.com" assets/*.js
# Found? → Using production API ✅
```

**Check 2**: Is Go server running?
```bash
curl http://localhost:7777/api/articles
# Should return JSON array ✅

# If connection refused:
go run cmd/server/main.go
```

**Check 3**: Which `.env` file is being used?
```bash
ls -la vue/.env*

# .env.local exists? → Takes priority over .env.production
# Want production API? → rm vue/.env.local
```

### Error: "CORS Policy Blocked"

If testing locally with production API:

**Add to Go server** (`cmd/server/main.go`):
```go
// Add CORS headers for local testing
w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
```

---

## 📝 Recommended Workflow

### For Development:
```bash
# Use dev server (hot reload)
cd vue/
npm run dev

# API proxied to localhost:7777
# No build needed
```

### For Local Testing:
```bash
# Test production build locally
# Requires: .env.local + local Go server

# Terminal 1:
go run cmd/server/main.go

# Terminal 2:
cd vue/
npm run build
cd dist/
python3 -m http.server 8080
```

### For Production Deployment:
```bash
# Build for production
rm vue/.env.local  # Remove local override
cd vue/
npm run build

# Upload dist/ folder
# API calls will use production domain
```

---

## ✅ Current Status

**Files**:
- ✅ `.env.local` created (uses `localhost:7777`)
- ✅ `.env.development` exists (uses `localhost:7777`)
- ✅ `.env.production` exists (uses `thaistockanalysis.com`)

**To Test Now**:
```bash
# 1. Start Go server
go run cmd/server/main.go

# 2. Already built with .env.local
cd vue/dist/
python3 -m http.server 8080

# 3. Open browser
http://localhost:8080
```

**Should work now!** ✅

---

## 🎯 Summary

**Problem**: Production build can't load articles when tested locally
**Cause**: Build uses production API URL, but local machine can't access it
**Solution**: Use `.env.local` to override API URL for local testing
**Result**: Can test production build with local Go server ✅

---

**Remember**:
- `.env.local` = Local testing (git-ignored)
- `.env.production` = Real production deployment
- Always start Go server when testing locally
- Remove `.env.local` before deploying to production
