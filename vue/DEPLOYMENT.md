# Vue Admin Panel - Production Deployment Guide

## 🎯 Overview

This is a **mobile-first Vue 3 admin panel** for managing Thai Stock Analysis articles. It provides a lightweight, responsive interface for creating and editing market data on-the-go.

### Key Features
- **Mobile-Optimized UI**: Touch-friendly interfaces with optimized text sizes and full-width buttons
- **Four Trading Sessions**: Morning/Afternoon Open/Close data management
- **Real-time API Integration**: Direct connection to Go backend with Gemini AI analysis
- **Responsive Design**: Tailwind CSS with vertical stacking for small screens
- **Offline-Friendly**: Minimal dependencies for fast mobile loading

### Tech Stack
- **Vue 3.5.13** - Composition API
- **Vue Router 4.4.0** - SPA routing
- **Axios 1.7.2** - HTTP client
- **Vite 5.4.21** - Build tool
- **Tailwind CSS** - Styling (CDN)

---

## 📦 Installation & Setup

### Prerequisites
- **Node.js**: v18.0.0 or higher
- **npm**: v9.0.0 or higher
- **Go Backend**: Running on port 7777 (or production domain)

### Development Setup

1. **Clone & Navigate**:
```bash
cd vue/
```

2. **Install Dependencies**:
```bash
npm install
```

3. **Configure Environment** (optional for development):
```bash
# Create .env file for local development
echo "VITE_API_URL=http://localhost:7777" > .env
```

4. **Start Development Server**:
```bash
npm run dev
```

Access at: `http://localhost:3000`

---

## 🚀 Production Deployment

### Step 1: Build for Production

```bash
cd vue/
npm run build
```

This creates a `dist/` folder with optimized static assets.

### Step 2: Environment Configuration

For production, the app connects to `https://thaistockanalysis.com` by default. To override:

```bash
# Set custom API URL
export VITE_API_URL=https://your-custom-domain.com
npm run build
```

### Step 3: Deploy Static Assets

#### Option A: Serve with Nginx (Recommended)

```nginx
server {
    listen 443 ssl http2;
    server_name admin.thaistockanalysis.com;

    # SSL certificates
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    root /var/www/vue-admin/dist;
    index index.html;

    # SPA fallback
    location / {
        try_files $uri $uri/ /index.html;
    }

    # Proxy API requests to Go backend
    location /api {
        proxy_pass https://thaistockanalysis.com;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
}
```

**Deploy Commands**:
```bash
# Copy built files to server
scp -r dist/* user@server:/var/www/vue-admin/dist/

# Restart Nginx
sudo systemctl restart nginx
```

#### Option B: Deploy to Vercel/Netlify

**Vercel**:
```bash
npm install -g vercel
vercel --prod
```

**Netlify**:
```bash
npm install -g netlify-cli
netlify deploy --prod --dir=dist
```

⚠️ **Important**: Configure environment variable `VITE_API_URL=https://thaistockanalysis.com` in deployment platform settings.

### Step 4: CORS Configuration (Backend)

Ensure Go backend allows requests from admin domain:

```go
// In cmd/server/main.go or handlers
w.Header().Set("Access-Control-Allow-Origin", "https://admin.thaistockanalysis.com")
w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
```

---

## 🔌 API Endpoints

All endpoints are prefixed with `/api` and proxied to the Go backend.

### 1. **GET /api/articles**
Returns list of all articles with market data.

**Response**:
```json
[
  {
    "slug": "2025-09-30",
    "title": "Stock Market Analysis - 30 September 2025",
    "summary": "Market opens strong with energy sector...",
    "morning_open_index": 1302.75,
    "morning_open_change": 16.49,
    "morning_close_index": 1280.38,
    "morning_close_change": -7.69,
    "afternoon_open_index": 1279.48,
    "afternoon_open_change": -8.59,
    "afternoon_close_index": 1295.80,
    "afternoon_close_change": 5.15
  }
]
```

### 2. **GET /api/articles/{date}**
Returns full article data for specific date.

**Example**: `GET /api/articles/2025-09-30`

**Response**:
```json
{
  "date": "2025-09-30",
  "title": "Stock Market Analysis - 30 September 2025",
  "morning_open": {
    "index": 1302.75,
    "change": 16.49,
    "highlights": "Energy firms rally eight points...",
    "analysis": "<p>Strong opening driven by oil sector...</p>"
  },
  "morning_close": {
    "index": 1280.38,
    "change": -7.69,
    "highlights": "+94 +97 +90 +80 +61 +59 +58 +68 +62",
    "summary": "<p>Morning session ends with profit-taking...</p>"
  },
  "afternoon_open": {
    "index": 1279.48,
    "change": -8.59,
    "highlights": "Energy sector continues pressure...",
    "analysis": "<p>Afternoon opens with continued selling...</p>"
  },
  "afternoon_close": {
    "index": 1295.80,
    "change": 5.15,
    "highlights": "+68 +61 +64 +78 +63 +87 +80 +94",
    "summary": "<p>Market recovers in final hour...</p>"
  }
}
```

### 3. **POST /api/market-data-analysis**
Submit opening data (morning/afternoon).

**Request Body**:
```json
{
  "date": "2025-09-30",
  "session": "morning",
  "index": "1302.75",
  "change": "16.49",
  "highlights": "Energy firms rally eight points as oil prices spike."
}
```

**Response**: `200 OK` with generated AI analysis.

### 4. **POST /api/market-data-close**
Submit closing data with summary.

**Request Body**:
```json
{
  "date": "2025-09-30",
  "session": "morning",
  "index": "1280.38",
  "change": "-7.69",
  "highlights": "+94 +97 +90 +80 +61 +59 +58 +68 +62"
}
```

**Response**: `200 OK` with generated summary.

---

## 📱 Mobile Optimization Features

### Design Principles
- **Vertical Stacking**: All inputs stack vertically (no horizontal grid)
- **Touch-Friendly**: 44px minimum touch targets (iOS/Android standard)
- **Readable Text**: Optimized font sizes (text-xl → text-sm hierarchy)
- **Full-Width Buttons**: Easier tapping with w-full buttons
- **Minimal Padding**: Compact spacing (p-3, p-4) for more content on screen
- **No Horizontal Scroll**: 3-row textareas with `resize-none`

### Component Responsiveness

#### ArticleList.vue
- **Cards**: `p-4` padding, `space-y-3` between cards
- **Text**: `text-xl` titles, `text-sm` summaries, `text-xs` metadata
- **Touch Feedback**: `active:bg-gray-100` for tactile response
- **Loading State**: Centered spinner with "Loading articles..."

#### ArticleDetail.vue
- **Form Inputs**: `space-y-3` vertical spacing, `text-sm` labels
- **Textareas**: `rows="3"` for highlights (no horizontal scroll)
- **Session Sections**: `mb-6` between morning/afternoon sessions
- **Buttons**: `w-full` with `text-base` text, `py-3` padding
- **Data Loading**: Fetches existing data via `loadArticleData()` on mount

#### CreateArticle.vue
- **Date Picker**: Auto-selects current date, `text-sm` styling
- **Textareas**: 3-row highlights inputs with `resize-none`
- **Submit Flow**: Sequential buttons for open → close → open → close
- **Success Feedback**: Alerts on successful submission

---

## 🔧 Configuration Reference

### Vite Config (vite.config.js)

```javascript
export default defineConfig({
	plugins: [vue()],
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
})
```

**Environment Variables**:
- `VITE_API_URL`: Backend API base URL (default: `https://thaistockanalysis.com`)

### API Service (src/services/api.js)

All API calls use Axios with relative URLs (proxied by Vite in dev):

```javascript
// Development: http://localhost:3000/api/* → http://localhost:7777/api/*
// Production: https://admin.domain.com/api/* → https://thaistockanalysis.com/api/*

export default {
  getArticles: () => axios.get('/api/articles'),
  getArticle: (date) => axios.get(`/api/articles/${date}`),
  submitMorningOpen: (data) => axios.post('/api/market-data-analysis', data),
  // ...
}
```

---

## 🐛 Troubleshooting

### Issue: "Failed to load articles"
**Cause**: Backend not running or CORS blocking.
**Solution**:
1. Verify Go server is running on port 7777 (or production domain)
2. Check browser console for CORS errors
3. Add `Access-Control-Allow-Origin` header in Go backend

### Issue: "Cannot GET /article/2025-09-30" (404 on refresh)
**Cause**: SPA routing requires server-side fallback.
**Solution**: Configure server to serve `index.html` for all routes:
```nginx
location / {
    try_files $uri $uri/ /index.html;
}
```

### Issue: API requests fail with "Network Error"
**Cause**: Incorrect `VITE_API_URL` or backend unreachable.
**Solution**:
1. Check environment variable: `echo $VITE_API_URL`
2. Test backend directly: `curl https://thaistockanalysis.com/api/articles`
3. Verify SSL certificates if using HTTPS

### Issue: Textareas show horizontal scroll
**Cause**: Browser default textarea styling.
**Solution**: Already fixed with `resize-none` class. If issue persists:
```css
textarea {
  resize: none !important;
  overflow-x: hidden;
}
```

### Issue: "npm run dev" shows blank page
**Cause**: Missing dependencies or build errors.
**Solution**:
```bash
# Clear cache and reinstall
rm -rf node_modules package-lock.json
npm install
npm run dev
```

---

## 📊 Performance Optimization

### Production Build Size
- **Bundle Size**: ~150KB gzipped (Vue + Router + Axios)
- **Tailwind**: CDN (~20KB) - no build-time compilation
- **Images**: Minimal (only favicon)

### Load Time Targets
- **First Contentful Paint**: < 1.5s
- **Time to Interactive**: < 3s
- **Lighthouse Score**: > 90

### Optimization Tips
1. **Code Splitting**: Already optimized by Vite (automatic)
2. **Lazy Loading**: Routes lazy-loaded via dynamic imports
3. **CDN Assets**: Tailwind served from CDN (cached globally)
4. **Compression**: Enable gzip/brotli in Nginx:
   ```nginx
   gzip on;
   gzip_types text/plain text/css application/json application/javascript;
   ```

---

## 🔐 Security Considerations

### Backend Route Protection
Admin routes on main Go server are redirected to homepage:
```go
// In cmd/server/main.go
mux.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/", http.StatusMovedPermanently)
})
```

### Vue Admin Access Control
Currently, the Vue app has **no authentication**. For production:

**Recommended Solutions**:
1. **Basic Auth (Nginx)**:
   ```nginx
   location / {
       auth_basic "Admin Area";
       auth_basic_user_file /etc/nginx/.htpasswd;
   }
   ```

2. **JWT Authentication**:
   - Add login endpoint to Go backend
   - Store JWT in localStorage
   - Include token in Axios headers
   - Validate on backend for all /api requests

3. **IP Whitelisting**:
   ```nginx
   location / {
       allow 203.0.113.0/24;  # Your IP range
       deny all;
   }
   ```

### Data Validation
All user inputs are submitted to backend - validation handled by Go server.

---

## 📝 Development Workflow

### Adding New Features

1. **Create Component**:
   ```bash
   touch src/views/NewFeature.vue
   ```

2. **Add Route**:
   ```javascript
   // src/router/index.js
   {
     path: '/new-feature',
     name: 'NewFeature',
     component: () => import('../views/NewFeature.vue')
   }
   ```

3. **Add API Method** (if needed):
   ```javascript
   // src/services/api.js
   export default {
     newFeatureAPI: (data) => axios.post('/api/new-endpoint', data)
   }
   ```

4. **Test Locally**:
   ```bash
   npm run dev
   ```

### Code Standards
- **Composition API**: Use `<script setup>` syntax
- **Reactive Data**: `ref()` for primitives, `reactive()` for objects
- **Async**: Use `async/await` with try-catch
- **Mobile-First**: Test on small viewports (375px width)

---

## 🚧 Known Limitations

1. **No Authentication**: Vue app is publicly accessible if deployed
2. **No Offline Mode**: Requires internet connection for API calls
3. **Basic Error Handling**: Shows alert() for errors (can be improved with toast notifications)
4. **No Image Upload**: Article management is text-only
5. **No Real-time Updates**: No WebSocket for live data (refresh required)

---

## 📞 Support & Maintenance

### File Structure
```
vue/
├── src/
│   ├── App.vue              # Root component with navigation
│   ├── main.js              # Vue app initialization
│   ├── router/
│   │   └── index.js         # Route definitions
│   ├── services/
│   │   └── api.js           # Axios API service
│   └── views/
│       ├── ArticleList.vue  # List all articles
│       ├── ArticleDetail.vue # Edit existing article
│       └── CreateArticle.vue # Create new article
├── index.html               # Entry point (Tailwind CDN)
├── vite.config.js           # Vite configuration
├── package.json             # Dependencies
└── DEPLOYMENT.md            # This file
```

### Logs & Debugging
- **Development**: Browser console for Vue/Axios errors
- **Production**: Check Nginx access/error logs
- **Backend**: Go server logs on stdout (check systemd journal)

### Version History
- **v1.0.0** (Current): Mobile-first Vue 3 admin panel with 4 trading sessions
- Backend API: /api/articles, /api/articles/{date}, /api/market-data-analysis, /api/market-data-close

---

## 🎓 Additional Resources

### Documentation Links
- [Vue 3 Docs](https://vuejs.org/guide/introduction.html)
- [Vue Router](https://router.vuejs.org/)
- [Axios](https://axios-http.com/docs/intro)
- [Vite](https://vitejs.dev/guide/)
- [Tailwind CSS](https://tailwindcss.com/docs)

### Related Files
- **Backend API Reference**: `../docs/API_QUICK_REFERENCE.md`
- **Go Server Guide**: `../.github/copilot-instructions.md`
- **HTML Bug Fix**: `../BUGFIX_HTML_RENDERING.md`

---

## ✅ Production Checklist

Before deploying to production:

- [ ] Run `npm run build` successfully
- [ ] Test all routes in production build (`npm run preview`)
- [ ] Verify API endpoints work with HTTPS domain
- [ ] Configure CORS on Go backend
- [ ] Set up SSL certificates (Let's Encrypt recommended)
- [ ] Configure Nginx with SPA fallback
- [ ] Enable gzip compression
- [ ] Set up authentication (Basic Auth/JWT/IP whitelist)
- [ ] Test on actual mobile devices (iOS/Android)
- [ ] Monitor backend logs for API errors
- [ ] Create database backups before first production use

---

**Last Updated**: 2025-01-XX  
**Maintainer**: ThaiStockAnalysis Development Team  
**License**: Private - Internal Use Only
