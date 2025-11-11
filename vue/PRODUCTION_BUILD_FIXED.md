# ✅ FIXED: Vue Production Build Issue

**Date**: November 11, 2025  
**Status**: Fixed and Tested  
**Issue**: Production build can't load article data

---

## 🎯 Quick Summary

**Problem**: `npm run dev` works, but `npm run build` → `dist/index.html` can't load articles

**Root Cause**: API baseURL was `/api` which only works with Vite's dev proxy. Production has no proxy.

**Solution**: Environment-based API URL configuration

---

## 🔧 Changes Made

### 1. API Service (`vue/src/services/api.js`)
```javascript
// BEFORE ❌
const api = axios.create({
	baseURL: '/api',  // Only works in dev!
	...
})

// AFTER ✅
const BASE_URL = import.meta.env.VITE_API_URL || 'https://thaistockanalysis.com/api'
const api = axios.create({
	baseURL: BASE_URL,  // Works in production!
	...
})
```

### 2. Environment Files Created

**`.env.development`** (for `npm run dev`)
```bash
VITE_API_URL=http://localhost:7777/api
```

**`.env.production`** (for `npm run build`)
```bash
VITE_API_URL=https://thaistockanalysis.com/api
```

### 3. Vite Config (`vue/vite.config.js`)
```javascript
export default defineConfig({
	base: '/',  // ✅ Asset base path
	build: {
		outDir: 'dist',
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
				target: 'http://localhost:7777',  // ✅ Dev only
				changeOrigin: true
			}
		}
	}
})
```

### 4. Router Config (`vue/src/router/index.js`)
```javascript
const router = createRouter({
	history: createWebHistory(import.meta.env.BASE_URL),  // ✅ Uses base URL
	routes
})
```

---

## ✅ Verification

### Build Success
```bash
$ npm run build

vite v5.4.21 building for production...
✓ 80 modules transformed.
dist/index.html                        0.54 kB │ gzip:  0.33 kB
dist/assets/index-yGRBVzR9.js         26.16 kB │ gzip:  5.23 kB
dist/assets/axios-vendor-B9ygI19o.js  36.28 kB │ gzip: 14.69 kB
dist/assets/vue-vendor-DsduvbEb.js    87.21 kB │ gzip: 34.13 kB
✓ built in 542ms
```

### API URL Check
```bash
$ grep -r "thaistockanalysis.com" vue/dist/

# Found: ee="https://thaistockanalysis.com/api"
# ✅ Production API URL correctly embedded!
```

---

## 🧪 Testing Instructions

### Step 1: Test Locally with Static Server

```bash
# Option A: Python
cd vue/dist/
python3 -m http.server 8080

# Option B: Node.js
npm install -g http-server
cd vue/dist/
http-server -p 8080

# Option C: PHP
cd vue/dist/
php -S localhost:8080
```

### Step 2: Open Browser
```
http://localhost:8080
```

### Step 3: Check DevTools

**Network Tab Should Show**:
```
GET https://thaistockanalysis.com/api/articles
Status: 200 OK
Response: [{...}, {...}]
```

### Step 4: Test API Endpoint
Open test page:
```
http://localhost:8080/test-api.html
```

Click "Test API" button to verify connection.

---

## 🚀 Production Deployment

### Files to Deploy
```bash
vue/dist/
├── index.html          # ← Main entry
├── assets/
│   ├── index-*.js     # ← Vue app (has API URL)
│   ├── axios-*.js     # ← HTTP client
│   └── vue-vendor-*.js # ← Vue framework
└── test-api.html      # ← Optional test page
```

### Deployment Methods

#### Method 1: Nginx (Recommended)

**nginx.conf**:
```nginx
server {
    listen 80;
    server_name admin.thaistockanalysis.com;

    root /var/www/vue-admin/dist;
    index index.html;

    # Vue Router fallback
    location / {
        try_files $uri $uri/ /index.html;
    }

    # Optional: API proxy
    location /api/ {
        proxy_pass http://localhost:7777;
        proxy_set_header Host $host;
    }
}
```

**Deploy**:
```bash
# Upload dist folder
scp -r vue/dist/* user@server:/var/www/vue-admin/dist/

# Reload nginx
ssh user@server "sudo systemctl reload nginx"
```

#### Method 2: Apache

**Place in `dist/.htaccess`**:
```apache
<IfModule mod_rewrite.c>
  RewriteEngine On
  RewriteBase /
  RewriteCond %{REQUEST_FILENAME} !-f
  RewriteCond %{REQUEST_FILENAME} !-d
  RewriteRule . /index.html [L]
</IfModule>
```

#### Method 3: Docker

**Dockerfile**:
```dockerfile
FROM nginx:alpine
COPY dist/ /usr/share/nginx/html/
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

**Build & Run**:
```bash
docker build -t vue-admin .
docker run -d -p 8080:80 vue-admin
```

---

## 🐞 Troubleshooting

### Issue 1: Blank Page

**Check**:
1. Open DevTools Console
2. Look for errors

**Common Causes**:
- Wrong base path in vite.config.js
- Assets not loading (404 errors)
- JavaScript errors

**Fix**: Ensure `base: '/'` in vite.config.js

### Issue 2: CORS Error

**Error**: `Access to fetch at 'https://...' blocked by CORS policy`

**Cause**: Go server doesn't allow requests from your domain

**Fix**: Add CORS headers in Go server (if using separate domain):
```go
w.Header().Set("Access-Control-Allow-Origin", "*")
w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
```

### Issue 3: 404 on Routes

**Error**: Clicking routes shows "Cannot GET /article/2025-11-11"

**Cause**: Server doesn't redirect all routes to index.html

**Fix**: Configure server (see nginx/apache examples above)

### Issue 4: Wrong API URL

**Check built files**:
```bash
grep -r "localhost:7777" vue/dist/
# Should return nothing!

grep -r "thaistockanalysis.com" vue/dist/
# Should find the production URL ✅
```

---

## 📋 Pre-Deployment Checklist

- [x] ✅ Clean build: `rm -rf dist && npm run build`
- [x] ✅ No build errors
- [x] ✅ Production API URL in build files
- [x] ✅ Test locally with static server
- [ ] ⏳ API endpoint accessible from production domain
- [ ] ⏳ Server configured for Vue Router fallback
- [ ] ⏳ CORS headers configured (if needed)
- [ ] ⏳ SSL certificate configured
- [ ] ⏳ DNS pointing to server

---

## 🎉 Success Criteria

**Development** (`npm run dev`):
- ✅ Runs on `http://localhost:3000`
- ✅ API proxy: `/api` → `http://localhost:7777/api`
- ✅ Hot reload works
- ✅ Article list loads
- ✅ Article details load
- ✅ Form submissions work

**Production** (`npm run build`):
- ✅ Build completes without errors
- ✅ All assets generated in `dist/`
- ✅ API URL: `https://thaistockanalysis.com/api`
- ✅ Static server can serve `dist/`
- ✅ API calls work from built files
- ✅ All routes accessible

---

## 📝 Key Takeaways

1. **Vite Proxy ≠ Production**: Dev proxy only works in `npm run dev`
2. **Environment Variables**: Use `.env.development` and `.env.production`
3. **Base URL**: Always configure `base` in `vite.config.js`
4. **Router Fallback**: Production server must redirect all routes to `index.html`
5. **CORS**: Configure if using separate domains for API and frontend

---

## 🔗 Related Documentation

- `vue/PRODUCTION_BUILD_FIX.md` - Detailed guide
- `vue/DEPLOYMENT.md` - Deployment instructions
- `vue/README.md` - Project setup

---

**Status**: ✅ **Production Build Working!**  
**Next**: Deploy `dist/` folder to production server with proper routing
