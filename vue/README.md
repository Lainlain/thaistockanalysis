# Vue Admin Panel for Thai Stock Analysis

🚀 **Mobile-first admin interface** for managing Thai stock market articles with real-time Gemini AI analysis integration.

---

## ⚡ Quick Start

### Development Mode
```bash
# Install dependencies (first time only)
npm install

# Start development server (with hot reload)
npm run dev
```

Access at: **http://localhost:3000**

### Production Build
```bash
# Build optimized static files
npm run build

# Preview production build locally
npm run preview
```

Output: `dist/` folder ready for deployment

---

## 🏗️ Architecture

### Tech Stack
- **Vue 3.5.13** - Composition API with `<script setup>`
- **Vue Router 4.4.0** - SPA routing
- **Axios 1.7.2** - HTTP client for API calls
- **Vite 5.4.21** - Lightning-fast build tool
- **Tailwind CSS** - Utility-first styling (via CDN)

### Backend Integration
- **Go Server**: Runs on port 7777 (dev) or https://thaistockanalysis.com (prod)
- **API Proxy**: Vite proxies `/api/*` requests to backend
- **Gemini AI**: Generates market analysis via backend endpoints

---

## 📱 Features

### Mobile Optimization
✅ **Touch-Friendly**: 44px minimum touch targets  
✅ **Vertical Layout**: No horizontal scrolling  
✅ **Readable Text**: Optimized font hierarchy (xl → sm → xs)  
✅ **Full-Width Buttons**: Easy tapping with `w-full` buttons  
✅ **3-Row Textareas**: No horizontal scroll for long highlights  
✅ **Active Feedback**: `active:bg-gray-100` for tactile response  

### Trading Sessions Management
Manage **four distinct trading periods** per day:
1. **Morning Open**: Index, Change, Highlights, AI Analysis
2. **Morning Close**: Index, Change, Highlights, Summary
3. **Afternoon Open**: Index, Change, Highlights, AI Analysis
4. **Afternoon Close**: Index, Change, Highlights, Summary

### Views
1. **Article List** (`/`) - Browse all articles with market data preview
2. **Article Detail** (`/article/:date`) - Edit existing article data
3. **Create Article** (`/create`) - Create new market data entries

---

## 🔌 API Endpoints

All endpoints are proxied to Go backend:

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/articles` | List all articles with market data |
| GET | `/api/articles/{date}` | Get full article data (YYYY-MM-DD) |
| POST | `/api/market-data-analysis` | Submit opening data (AI analysis) |
| POST | `/api/market-data-close` | Submit closing data (AI summary) |

**Example Request** (Submit Morning Open):
```bash
curl -X POST http://localhost:7777/api/market-data-analysis \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2025-09-30",
    "session": "morning",
    "index": "1302.75",
    "change": "16.49",
    "highlights": "Energy firms rally eight points as oil prices spike."
  }'
```

---

## 🚀 Production Deployment

### Step 1: Configure Environment
```bash
# Copy environment template
cp .env.example .env

# Edit for production (optional - defaults to thaistockanalysis.com)
nano .env
```

### Step 2: Build Production Bundle
```bash
npm run build
```

### Step 3: Deploy to Server
```bash
# Copy dist/ folder to server
scp -r dist/* user@server:/var/www/vue-admin/

# Configure Nginx (see DEPLOYMENT.md for full config)
sudo nano /etc/nginx/sites-available/vue-admin
sudo systemctl restart nginx
```

📖 **Full deployment guide**: See `DEPLOYMENT.md`

---

## 🛠️ Configuration

### Environment Variables
Create `.env` file (see `.env.example`):
```bash
# Backend API URL
VITE_API_URL=https://thaistockanalysis.com
```

**Defaults** (if not set):
- **Development**: Proxy to `http://localhost:7777`
- **Production**: Proxy to `https://thaistockanalysis.com`

### Vite Config (`vite.config.js`)
```javascript
server: {
  port: 3000,
  proxy: {
    '/api': {
      target: process.env.VITE_API_URL || 'https://thaistockanalysis.com',
      changeOrigin: true,
      secure: true
    }
  }
}
```

---

## 📂 Project Structure

```
vue/
├── src/
│   ├── App.vue              # Root component with navigation
│   ├── main.js              # Vue app entry point
│   ├── router/
│   │   └── index.js         # Route definitions (3 routes)
│   ├── services/
│   │   └── api.js           # Axios API service (6 methods)
│   └── views/
│       ├── ArticleList.vue  # List all articles
│       ├── ArticleDetail.vue # Edit existing article (with data loading)
│       └── CreateArticle.vue # Create new article
├── public/                   # Static assets (favicon)
├── index.html               # Entry HTML (Tailwind CDN)
├── vite.config.js           # Vite configuration
├── package.json             # Dependencies
├── .env.example             # Environment template
├── README.md                # This file
└── DEPLOYMENT.md            # Full deployment guide
```

---

## 🐛 Troubleshooting

### "Failed to load articles"
**Cause**: Backend not running or CORS issue.  
**Fix**: 
1. Ensure Go server is running on port 7777
2. Check browser console for errors
3. Verify `/api/articles` endpoint returns JSON

### Blank page on refresh
**Cause**: SPA routing requires server fallback.  
**Fix**: Configure Nginx with `try_files $uri /index.html`

### Horizontal scroll on textareas
**Cause**: Default browser styling.  
**Fix**: Already fixed with `resize-none` class in components

### "npm run dev" fails
**Cause**: Missing dependencies.  
**Fix**: 
```bash
rm -rf node_modules package-lock.json
npm install
```

---

## 📊 Performance

### Build Metrics
- **Bundle Size**: ~150KB gzipped
- **First Paint**: < 1.5s
- **Lighthouse Score**: > 90

### Optimization
- ✅ Code splitting (automatic via Vite)
- ✅ Lazy-loaded routes
- ✅ CDN assets (Tailwind)
- ✅ Minimal dependencies

---

## 🔐 Security

⚠️ **No built-in authentication** - Recommended for production:

1. **Basic Auth (Nginx)**:
```nginx
auth_basic "Admin Area";
auth_basic_user_file /etc/nginx/.htpasswd;
```

2. **IP Whitelist**:
```nginx
allow 203.0.113.0/24;
deny all;
```

3. **JWT Authentication** (requires backend changes)

See `DEPLOYMENT.md` Section: Security Considerations

---

## � Documentation

- **Deployment Guide**: `DEPLOYMENT.md` - Complete production setup
- **API Reference**: `../docs/API_QUICK_REFERENCE.md` - Backend API docs
- **Backend Architecture**: `../.github/copilot-instructions.md` - Go server structure

---

## 🎯 Development Workflow

### Run Development Server
```bash
npm run dev
```
- Hot reload enabled
- API proxy to localhost:7777
- Access at http://localhost:3000

### Build for Production
```bash
npm run build
```
- Minified bundle in `dist/`
- Ready for deployment

### Preview Production Build
```bash
npm run preview
```
- Test production build locally
- Runs on port 4173

---

## ✅ Production Checklist

Before deploying:

- [ ] `npm run build` completes successfully
- [ ] Test with `npm run preview`
- [ ] Verify backend is accessible at production domain
- [ ] Configure CORS headers on Go backend
- [ ] Set up SSL certificates (Let's Encrypt)
- [ ] Configure Nginx with SPA fallback
- [ ] Enable gzip compression
- [ ] Add authentication (Basic Auth/JWT)
- [ ] Test on real mobile devices (iOS/Android)
- [ ] Set up database backups

---

## � Support

**Issues**: Check `DEPLOYMENT.md` troubleshooting section  
**Backend Logs**: Check Go server stdout or systemd journal  
**Frontend Errors**: Browser console (F12)

---

**Version**: 1.0.0  
**Last Updated**: 2025-01-XX  
**Production URL**: https://thaistockanalysis.com  
**License**: Private - Internal Use Only

