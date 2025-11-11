# 🔄 Quick Migration Guide - localhost → Production

This guide explains how the Vue Admin Panel now works in both **development** and **production** environments.

---

## 🎯 What Changed?

### Before (Development Only)
```javascript
// vite.config.js - hardcoded localhost
proxy: {
  '/api': {
    target: 'http://localhost:7777',  // ❌ Only works locally
  }
}
```

### After (Development + Production)
```javascript
// vite.config.js - environment-aware
proxy: {
  '/api': {
    target: process.env.VITE_API_URL || 'https://thaistockanalysis.com',  // ✅ Works everywhere
    changeOrigin: true,
    secure: true
  }
}
```

---

## 🚀 How It Works Now

### Development Mode (Local)
```bash
# Option 1: Use default localhost (create .env)
echo "VITE_API_URL=http://localhost:7777" > .env
npm run dev

# Option 2: Set environment variable inline
VITE_API_URL=http://localhost:7777 npm run dev

# Result:
# Vue app: http://localhost:3000
# API calls: /api/* → proxied to → http://localhost:7777/api/*
```

### Production Mode (Deployed)
```bash
# Build with default production domain
npm run build

# Result:
# Vue app: Deployed at https://admin.thaistockanalysis.com
# API calls: /api/* → proxied to → https://thaistockanalysis.com/api/*
```

---

## 📝 Environment Variable Priority

The proxy target resolves in this order:

1. **Environment Variable** (`.env` file or shell export)
   ```bash
   VITE_API_URL=http://custom-domain.com
   ```

2. **Default Production Domain** (if no env var)
   ```javascript
   'https://thaistockanalysis.com'
   ```

---

## 🔧 Common Scenarios

### Scenario 1: Local Development with Local Backend
```bash
# Create .env file
echo "VITE_API_URL=http://localhost:7777" > .env

# Run dev server
npm run dev

# API calls go to: http://localhost:7777
```

### Scenario 2: Local Development with Production Backend
```bash
# Create .env file
echo "VITE_API_URL=https://thaistockanalysis.com" > .env

# Run dev server
npm run dev

# API calls go to: https://thaistockanalysis.com (production data!)
```

### Scenario 3: Production Deployment (Default)
```bash
# No .env needed - uses default
npm run build

# Deploy dist/ folder
scp -r dist/* user@server:/var/www/vue-admin/

# API calls go to: https://thaistockanalysis.com
```

### Scenario 4: Production Deployment (Custom Domain)
```bash
# Set custom domain
export VITE_API_URL=https://api.custom-domain.com

# Build with custom domain
npm run build

# API calls go to: https://api.custom-domain.com
```

---

## 🧪 Testing Both Environments

### Test Local Backend
```bash
# Terminal 1: Start Go backend
cd ..
go run cmd/server/main.go
# Backend running on http://localhost:7777

# Terminal 2: Start Vue dev server
cd vue/
echo "VITE_API_URL=http://localhost:7777" > .env
npm run dev
# Vue running on http://localhost:3000

# Test: http://localhost:3000
# API calls go to localhost:7777 ✅
```

### Test Production Backend
```bash
# Terminal 1: Vue dev server with production API
cd vue/
echo "VITE_API_URL=https://thaistockanalysis.com" > .env
npm run dev

# Test: http://localhost:3000
# API calls go to https://thaistockanalysis.com ✅
```

### Test Production Build Locally
```bash
# Build production bundle
npm run build

# Preview production build
npm run preview
# Runs on http://localhost:4173

# Test: http://localhost:4173
# API calls go to https://thaistockanalysis.com (default) ✅
```

---

## ⚠️ Important Notes

### 1. **Environment Variables are Build-Time**
The `VITE_API_URL` is embedded during build:

```bash
# This sets the API URL permanently in the build
VITE_API_URL=https://production.com npm run build

# dist/ folder now has production.com hardcoded
# You CANNOT change it without rebuilding
```

### 2. **`.env` File is NOT Deployed**
The `.env` file is only for **local development**:

```bash
# .gitignore already has this:
.env

# So .env never goes to production
# Production uses the default: https://thaistockanalysis.com
```

### 3. **CORS Configuration Required**
Production backend must allow Vue admin domain:

```go
// In Go backend handlers
w.Header().Set("Access-Control-Allow-Origin", "https://admin.thaistockanalysis.com")
w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
```

Without this, browser will block API calls with CORS errors.

---

## 🐛 Troubleshooting

### Issue: "Network Error" in Console
**Check**: What's the API target?

```bash
# 1. Check if .env exists
cat .env

# 2. Check Vite config
cat vite.config.js | grep target

# 3. Check environment variable
echo $VITE_API_URL
```

**Fix**: Set correct API URL in `.env`:
```bash
echo "VITE_API_URL=http://localhost:7777" > .env
npm run dev
```

### Issue: CORS Error in Browser
**Symptom**: Console shows "CORS policy blocked"

**Fix**: Add CORS headers in Go backend:
```go
w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")  // Dev
// OR
w.Header().Set("Access-Control-Allow-Origin", "https://admin.thaistockanalysis.com")  // Prod
```

### Issue: 404 on API Calls
**Check**: Is backend running?

```bash
# Test backend directly
curl http://localhost:7777/api/articles

# If 404, backend is not running or route missing
go run cmd/server/main.go
```

---

## 📋 Quick Reference

### Development Commands
```bash
# Install dependencies (first time)
npm install

# Start dev server (localhost backend)
echo "VITE_API_URL=http://localhost:7777" > .env
npm run dev

# Start dev server (production backend)
echo "VITE_API_URL=https://thaistockanalysis.com" > .env
npm run dev
```

### Production Commands
```bash
# Build for production (default domain)
npm run build

# Build for custom domain
VITE_API_URL=https://custom-domain.com npm run build

# Preview production build locally
npm run preview
```

### Environment Files
```bash
# .env (local development only - not in git)
VITE_API_URL=http://localhost:7777

# .env.example (template - committed to git)
VITE_API_URL=https://thaistockanalysis.com
```

---

## ✅ Verification Checklist

After setting up, verify both modes work:

### Local Development
- [ ] Go backend running on port 7777
- [ ] `.env` file exists with `VITE_API_URL=http://localhost:7777`
- [ ] `npm run dev` starts without errors
- [ ] Browser opens to `http://localhost:3000`
- [ ] Article list loads data from local backend
- [ ] Can create/edit articles
- [ ] No CORS errors in console

### Production Build
- [ ] `npm run build` completes successfully
- [ ] `dist/` folder created with assets
- [ ] `npm run preview` works at `http://localhost:4173`
- [ ] API calls go to `https://thaistockanalysis.com`
- [ ] All routes work (/, /create, /article/*)

---

## 🎓 For Team Members

### If You're Just Developing Locally
```bash
# 1. Clone repo and enter vue/ folder
cd vue/

# 2. Install dependencies (first time only)
npm install

# 3. Create .env for local backend
echo "VITE_API_URL=http://localhost:7777" > .env

# 4. Start both servers:
# Terminal 1: Go backend
cd ..
go run cmd/server/main.go

# Terminal 2: Vue dev server
cd vue/
npm run dev

# 5. Open http://localhost:3000
```

### If You're Deploying to Production
```bash
# 1. Build production bundle (no .env needed)
cd vue/
npm run build

# 2. Upload to server
scp -r dist/* user@server:/var/www/vue-admin/

# 3. Done! API calls automatically go to https://thaistockanalysis.com
```

---

## 📞 Need Help?

- **Documentation**: See `DEPLOYMENT.md` for complete production guide
- **Checklist**: See `PRODUCTION_CHECKLIST.md` for step-by-step deployment
- **API Docs**: See `../docs/API_QUICK_REFERENCE.md` for backend API reference

---

**Summary**: The Vue app now intelligently routes API calls based on environment - **localhost for development**, **production domain for deployment**. No code changes needed between environments! 🎉
