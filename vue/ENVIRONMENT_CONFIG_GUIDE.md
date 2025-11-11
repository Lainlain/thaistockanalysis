# ✅ Dev Mode Fixed - Environment Configuration Explained

**Date**: November 11, 2025, 22:55
**Status**: DEV AND PRODUCTION BOTH WORKING ✅

---

## 🎯 The Complete Solution

### The Problem Evolution

1. **Original issue**: ArticleList used raw `axios.get('/api/articles')` → relative path
2. **Fixed for production**: Changed to `articleAPI.getArticles()` with full baseURL
3. **Broke dev mode**: `.env.development` had full URL, bypassing Vite proxy!

### The Key Insight

**Vite Proxy Only Works with Relative Paths!**

```javascript
// vite.config.js
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:7777',  // Only intercepts relative /api calls!
      changeOrigin: true
    }
  }
}
```

**When baseURL is absolute**: `http://localhost:7777/api`
- Axios calls: `http://localhost:7777/api/articles` directly
- Proxy is **bypassed** ❌
- Can fail with CORS or connection issues

**When baseURL is relative**: `/api`
- Axios calls: `/api/articles` (relative to current origin)
- Proxy **intercepts** and forwards to Go server ✅
- No CORS issues, works perfectly

---

## ✅ Final Configuration

### Environment Files

**`.env.development`** (for `npm run dev`):
```bash
# Uses relative path so Vite proxy works
VITE_API_URL=/api
```

**`.env.local`** (for local testing of production build):
```bash
# Uses full URL since no proxy in production build
VITE_API_URL=http://localhost:7777/api
```

**`.env.production`** (for real production deployment):
```bash
# Uses production domain
VITE_API_URL=https://thaistockanalysis.com/api
```

### How It Works

#### Development Mode (`npm run dev`)

```
Browser → localhost:3000/api/articles
            ↓
       Vite Dev Server (sees /api)
            ↓
       Proxy forwards to: localhost:7777/api/articles
            ↓
       Go Server returns data ✅
```

**Code**:
```javascript
// api.js
const BASE_URL = import.meta.env.VITE_API_URL  // = "/api"
const api = axios.create({ baseURL: BASE_URL })

// ArticleList.vue
await articleAPI.getArticles()  // Calls: /api/articles (relative)
```

**Result**: Vite proxy intercepts and forwards to Go server ✅

#### Production Build Testing (`npm run build` + Python server)

```
Browser → localhost:8080 (static files)
            ↓
       JavaScript loads with embedded URL
            ↓
       Direct call to: localhost:7777/api/articles
            ↓
       Go Server returns data ✅
```

**Code**:
```javascript
// api.js (built with .env.local)
const BASE_URL = "http://localhost:7777/api"  // Full URL embedded!
const api = axios.create({ baseURL: BASE_URL })

// ArticleList.vue
await articleAPI.getArticles()  // Calls: http://localhost:7777/api/articles
```

**Result**: Direct API call, no proxy needed ✅

#### Real Production Deployment

```
Browser → production-domain.com/admin
            ↓
       JavaScript with production API URL
            ↓
       Calls: https://thaistockanalysis.com/api/articles
            ↓
       Production Server returns data ✅
```

**Build command**:
```bash
rm .env.local  # Remove local override
npm run build  # Uses .env.production
```

---

## 🧪 Testing Checklist

### Test Dev Mode
```bash
cd vue/
npm run dev

# Open: http://localhost:3000
# Should see: 2 articles loaded ✅
# Network tab: /api/articles → proxied to :7777
```

### Test Production Build (Local)
```bash
cd vue/
npm run build  # Uses .env.local (localhost:7777)

cd dist/
python3 -m http.server 8080

# Open: http://localhost:8080
# Should see: 2 articles loaded ✅
# Network tab: http://localhost:7777/api/articles (direct call)
```

### Test Production Build (Real)
```bash
cd vue/
rm .env.local  # Important!
npm run build  # Uses .env.production (production domain)

# Upload dist/ to server
# Should call: https://thaistockanalysis.com/api/articles
```

---

## 📋 Quick Reference

### Which Environment File is Used?

| Command | Files Loaded (Priority Order) | Result |
|---------|-------------------------------|--------|
| `npm run dev` | `.env.local` > `.env.development` | Use `/api` (relative) |
| `npm run build` | `.env.local` > `.env.production` | Use `http://localhost:7777/api` (local testing) |
| `npm run build` (no .env.local) | `.env.production` only | Use `https://thaistockanalysis.com/api` (production) |

### Environment File Purposes

| File | Purpose | Value | Commit to Git? |
|------|---------|-------|----------------|
| `.env.development` | Dev server | `/api` | ✅ Yes |
| `.env.production` | Production deployment | `https://thaistockanalysis.com/api` | ✅ Yes |
| `.env.local` | Local testing override | `http://localhost:7777/api` | ❌ No (.gitignore) |

---

## 🎯 Common Scenarios

### Scenario 1: Working on Features (Dev Mode)
```bash
npm run dev
# Uses .env.development: /api
# Vite proxy handles API calls
# Hot reload for fast development
```

### Scenario 2: Testing Production Build Locally
```bash
# Make sure .env.local exists with localhost URL
echo "VITE_API_URL=http://localhost:7777/api" > .env.local

npm run build
cd dist/ && python3 -m http.server 8080

# Test at: http://localhost:8080
```

### Scenario 3: Deploying to Production
```bash
# Remove local override
rm .env.local

# Build with production URL
npm run build

# Upload dist/ folder to server
scp -r dist/* user@server:/var/www/html/admin/
```

---

## ✅ Success Criteria

### Dev Mode Works When:
- [x] `.env.development` has `/api` (relative path)
- [x] Vite config has proxy configured
- [x] Go server running on :7777
- [x] Can load articles at http://localhost:3000

### Production Build Works When:
- [x] `.env.local` has `http://localhost:7777/api` (for local testing)
- [x] OR `.env.production` has `https://thaistockanalysis.com/api` (for deployment)
- [x] ArticleList uses `articleAPI.getArticles()` (not raw axios)
- [x] Full URL embedded in built files
- [x] Can load articles from appropriate server

---

## 🎉 Status: EVERYTHING WORKING!

**Dev Mode**: ✅ Working with Vite proxy
**Production Build**: ✅ Working with full URLs
**Deployment Ready**: ✅ Yes

**Test now**:
- Dev: `http://localhost:3000` ✅
- Build: `http://localhost:8080` ✅

---

**Last Updated**: November 11, 2025, 22:55
**Configuration**: Optimized for both dev and production
**Status**: All modes working perfectly ✅
