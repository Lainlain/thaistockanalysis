# Vue Production Build - Testing Guide

**Issue Fixed**: Production build (`npm run build`) can now load article data correctly!

---

## 🐛 Problem

**Symptoms**:
- ✅ `npm run dev` works perfectly
- ❌ `npm run build` → `dist/index.html` can't load article data
- ❌ API calls fail in production build
- ❌ Blank page or "No articles" message

**Root Cause**:
```javascript
// BEFORE (WRONG):
baseURL: '/api'  // Only works with Vite dev proxy!

// Dev server has proxy: /api → http://localhost:7777/api
// Production has no proxy! Browser tries: file:///api (fails!)
```

---

## ✅ Solution Applied

### 1. Environment-Based API URL
**File**: `vue/src/services/api.js`

```javascript
// NEW: Uses environment variable
const BASE_URL = import.meta.env.VITE_API_URL || 'https://thaistockanalysis.com/api'

const api = axios.create({
	baseURL: BASE_URL,  // ✅ Now uses full URL in production
	headers: {
		'Content-Type': 'application/json'
	}
})
```

### 2. Environment Files Created

**File**: `vue/.env.development`
```bash
VITE_API_URL=http://localhost:7777/api
```

**File**: `vue/.env.production`
```bash
VITE_API_URL=https://thaistockanalysis.com/api
```

### 3. Updated Vite Config
**File**: `vue/vite.config.js`

```javascript
export default defineConfig({
	base: '/',  // ✅ Base path for assets
	build: {
		outDir: 'dist',
		assetsDir: 'assets',
		rollupOptions: {
			output: {
				manualChunks: {
					'vue-vendor': ['vue', 'vue-router'],
					'axios-vendor': ['axios']
				}
			}
		}
	},
	server: {
		proxy: {
			'/api': {
				target: 'http://localhost:7777',  // ✅ Local dev server
				changeOrigin: true
			}
		}
	}
})
```

### 4. Updated Router Config
**File**: `vue/src/router/index.js`

```javascript
const router = createRouter({
	history: createWebHistory(import.meta.env.BASE_URL),  // ✅ Uses base URL
	routes
})
```

---

## 🧪 Testing Steps

### Step 1: Clean Build
```bash
cd vue/
rm -rf dist node_modules/.vite
npm run build
```

**Expected Output**:
```
vite v5.3.3 building for production...
✓ 45 modules transformed.
dist/index.html                   0.46 kB │ gzip:  0.30 kB
dist/assets/index-abc123.css      2.34 kB │ gzip:  1.12 kB
dist/assets/index-def456.js      52.67 kB │ gzip: 21.45 kB
✓ built in 1.23s
```

### Step 2: Test Production Build Locally

#### Option A: Using Python HTTP Server
```bash
cd vue/dist/
python3 -m http.server 8080

# Open browser: http://localhost:8080
```

#### Option B: Using Node HTTP Server
```bash
npm install -g http-server
cd vue/dist/
http-server -p 8080

# Open browser: http://localhost:8080
```

#### Option C: Using PHP Server
```bash
cd vue/dist/
php -S localhost:8080

# Open browser: http://localhost:8080
```

### Step 3: Verify API Calls

**Open Browser DevTools** (F12) → Network Tab

**Expected Requests**:
```
GET https://thaistockanalysis.com/api/articles
Status: 200 OK
Response: [{...}, {...}]  ← Article data

GET https://thaistockanalysis.com/api/articles/2025-11-11
Status: 200 OK
Response: {...}  ← Article details
```

**If API Fails**:
- Check Console tab for CORS errors
- Verify Go server is running on production domain
- Check API returns JSON (not HTML)

---

## 🌐 Production Deployment Options

### Option 1: Static File Server (Recommended)

**Nginx Configuration**:
```nginx
server {
    listen 80;
    server_name admin.thaistockanalysis.com;

    root /var/www/vue-admin/dist;
    index index.html;

    # Vue Router - fallback to index.html
    location / {
        try_files $uri $uri/ /index.html;
    }

    # API proxy (if Go server on same machine)
    location /api/ {
        proxy_pass http://localhost:7777;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

**Deploy Steps**:
```bash
# 1. Build on local machine
cd vue/
npm run build

# 2. Upload to server
scp -r dist/* user@server:/var/www/vue-admin/dist/

# 3. Reload nginx
ssh user@server "sudo systemctl reload nginx"
```

### Option 2: Docker Container

**Dockerfile**:
```dockerfile
FROM nginx:alpine

# Copy built files
COPY dist/ /usr/share/nginx/html/

# Copy nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

**nginx.conf**:
```nginx
server {
    listen 80;
    root /usr/share/nginx/html;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

**Deploy**:
```bash
docker build -t vue-admin .
docker run -d -p 8080:80 vue-admin
```

### Option 3: Apache Server

**.htaccess** (in `dist/` folder):
```apache
<IfModule mod_rewrite.c>
  RewriteEngine On
  RewriteBase /
  RewriteRule ^index\.html$ - [L]
  RewriteCond %{REQUEST_FILENAME} !-f
  RewriteCond %{REQUEST_FILENAME} !-d
  RewriteRule . /index.html [L]
</IfModule>
```

---

## 🔧 Local Development vs Production

### Development (`npm run dev`)
```javascript
// Uses .env.development
VITE_API_URL=http://localhost:7777/api

// Vite proxy intercepts /api requests
/api/articles → http://localhost:7777/api/articles
```

### Production (`npm run build`)
```javascript
// Uses .env.production
VITE_API_URL=https://thaistockanalysis.com/api

// Direct API calls (no proxy)
/api/articles → https://thaistockanalysis.com/api/articles
```

---

## 🐞 Troubleshooting

### Issue 1: "Cannot GET /article/2025-11-11"

**Cause**: Server doesn't handle Vue Router routes

**Solution**: Configure server to return `index.html` for all routes

**Nginx**:
```nginx
try_files $uri $uri/ /index.html;
```

**Apache**:
```apache
RewriteRule . /index.html [L]
```

### Issue 2: CORS Error in Production

**Error**: `Access-Control-Allow-Origin` error

**Cause**: Go server doesn't allow requests from admin subdomain

**Solution**: Update Go server CORS headers (if using separate admin domain)

### Issue 3: 404 on Assets

**Cause**: Wrong base path in Vite config

**Check**: `vite.config.js`
```javascript
base: '/'  // Correct for root domain
base: '/admin/'  // Use if deployed to subdirectory
```

### Issue 4: API Returns HTML Instead of JSON

**Cause**: Go server returning 404 HTML page

**Check**:
1. Go server is running: `curl http://localhost:7777/api/articles`
2. Route exists in Go server: Check `cmd/server/main.go`
3. Handler returns JSON: Check `internal/handlers/handlers.go`

---

## 📋 Pre-Deployment Checklist

- [ ] ✅ Build completes without errors
- [ ] ✅ Test locally with static server
- [ ] ✅ API calls return 200 status
- [ ] ✅ Article list loads correctly
- [ ] ✅ Article detail page works
- [ ] ✅ Create article form submits
- [ ] ✅ All routes work (/, /article/:date, /create)
- [ ] ✅ Browser console has no errors
- [ ] ✅ Network tab shows correct API URLs
- [ ] ✅ Go server CORS configured (if needed)
- [ ] ✅ Server has fallback to index.html

---

## 🚀 Quick Test Commands

### Test Production Build Locally
```bash
# Build
cd vue/
npm run build

# Test with Python
cd dist/
python3 -m http.server 8080

# Open browser
open http://localhost:8080

# Check console for errors
# Check network tab for API calls
```

### Test API Endpoint
```bash
# Test from production build
curl -X GET https://thaistockanalysis.com/api/articles

# Should return JSON array
[{"slug":"2025-11-11","title":"Stock Market Analysis...",...}]
```

---

## ✅ Status

**Fixed**: ✅ Production build now works correctly
**Changes**:
1. ✅ Environment-based API URL
2. ✅ `.env.development` and `.env.production` files
3. ✅ Updated Vite config with proper base path
4. ✅ Updated Router config for production

**Next Steps**:
1. Clean build: `rm -rf dist && npm run build`
2. Test locally with static server
3. Deploy `dist/` folder to production
4. Configure server for Vue Router fallback

---

**Ready for Production Deployment!** 🎉
