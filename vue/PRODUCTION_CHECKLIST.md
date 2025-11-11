# Production Deployment Checklist

## 📋 Pre-Deployment

### Code Preparation
- [ ] All features tested in development mode
- [ ] All console.log() statements removed or commented
- [ ] Error handling tested (network failures, API errors)
- [ ] Mobile responsiveness verified on multiple devices (iPhone, Android)
- [ ] All hardcoded localhost URLs replaced with production domain
- [ ] Code committed to version control (git)

### Environment Configuration
- [ ] `.env.example` file exists with correct template
- [ ] Production `.env` file created (if needed for custom domain)
- [ ] `VITE_API_URL` set to production domain: `https://thaistockanalysis.com`
- [ ] Backend environment variables configured (GEMINI_API_KEY, TELEGRAM_BOT_TOKEN)

### Build Process
- [ ] Run `npm install` to ensure dependencies are up-to-date
- [ ] Run `npm run build` successfully
- [ ] No build errors or warnings
- [ ] Run `npm run preview` to test production build locally
- [ ] Verify all routes work in preview mode (test deep links)

---

## 🖥️ Server Setup

### Backend (Go Server)
- [ ] Go server running on production server (port 7777)
- [ ] CORS headers configured to allow Vue admin domain:
  ```go
  w.Header().Set("Access-Control-Allow-Origin", "https://admin.thaistockanalysis.com")
  w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
  w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
  ```
- [ ] API endpoints tested:
  - [ ] `GET /api/articles` returns JSON
  - [ ] `GET /api/articles/{date}` returns full data
  - [ ] `POST /api/market-data-analysis` works
  - [ ] `POST /api/market-data-close` works
- [ ] Database (`data/admin.db`) accessible and populated
- [ ] Markdown files directory (`articles/`) exists with write permissions
- [ ] Gemini API key working (test with curl)
- [ ] Telegram notifications working (if enabled)

### Frontend (Vue App)
- [ ] Web server installed (Nginx recommended)
- [ ] SSL certificate obtained (Let's Encrypt via Certbot)
- [ ] Domain/subdomain configured (e.g., `admin.thaistockanalysis.com`)
- [ ] DNS A record pointing to server IP
- [ ] Firewall rules allow HTTP (80) and HTTPS (443)

---

## 🔧 Nginx Configuration

### Create Server Block
```bash
sudo nano /etc/nginx/sites-available/vue-admin
```

### Minimal Config
```nginx
server {
    listen 443 ssl http2;
    server_name admin.thaistockanalysis.com;

    ssl_certificate /etc/letsencrypt/live/admin.thaistockanalysis.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/admin.thaistockanalysis.com/privkey.pem;

    root /var/www/vue-admin/dist;
    index index.html;

    # SPA fallback (critical!)
    location / {
        try_files $uri $uri/ /index.html;
    }

    # Proxy API to backend
    location /api {
        proxy_pass https://thaistockanalysis.com;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Gzip compression
    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml;
}

# HTTP to HTTPS redirect
server {
    listen 80;
    server_name admin.thaistockanalysis.com;
    return 301 https://$server_name$request_uri;
}
```

### Enable and Test
```bash
# Enable site
sudo ln -s /etc/nginx/sites-available/vue-admin /etc/nginx/sites-enabled/

# Test configuration
sudo nginx -t

# Reload Nginx
sudo systemctl reload nginx
```

- [ ] Nginx configuration file created
- [ ] Configuration syntax tested (`nginx -t`)
- [ ] Nginx reloaded successfully

---

## 📦 Deployment Steps

### 1. Upload Build Files
```bash
# Create deployment directory
ssh user@server 'mkdir -p /var/www/vue-admin'

# Upload dist/ folder
scp -r dist/* user@server:/var/www/vue-admin/dist/

# Set permissions
ssh user@server 'sudo chown -R www-data:www-data /var/www/vue-admin'
```

- [ ] Files uploaded to server
- [ ] Permissions set correctly
- [ ] Nginx user can read files

### 2. SSL Certificate (Let's Encrypt)
```bash
# Install Certbot
sudo apt install certbot python3-certbot-nginx

# Obtain certificate
sudo certbot --nginx -d admin.thaistockanalysis.com

# Test auto-renewal
sudo certbot renew --dry-run
```

- [ ] SSL certificate installed
- [ ] HTTPS working
- [ ] HTTP redirects to HTTPS
- [ ] Auto-renewal configured

### 3. Security Configuration
Choose ONE method:

#### Option A: Basic Authentication (Recommended)
```bash
# Install Apache utilities
sudo apt install apache2-utils

# Create password file
sudo htpasswd -c /etc/nginx/.htpasswd admin

# Add to Nginx config:
location / {
    auth_basic "Admin Area";
    auth_basic_user_file /etc/nginx/.htpasswd;
    try_files $uri $uri/ /index.html;
}
```

- [ ] Basic auth configured
- [ ] Credentials tested

#### Option B: IP Whitelisting
```bash
# Add to Nginx config:
location / {
    allow 203.0.113.0/24;  # Your IP range
    deny all;
    try_files $uri $uri/ /index.html;
}
```

- [ ] Whitelist configured
- [ ] Access from allowed IP works
- [ ] Access from other IPs blocked

---

## ✅ Post-Deployment Testing

### Functionality Tests
- [ ] Homepage loads at `https://admin.thaistockanalysis.com`
- [ ] Navigation works (Home, Create Article buttons)
- [ ] Article list loads and displays data
- [ ] Click on article opens detail page
- [ ] Detail page loads existing data
- [ ] Can submit morning open data
- [ ] Can submit morning close data
- [ ] Can submit afternoon open data
- [ ] Can submit afternoon close data
- [ ] Create new article works
- [ ] Success/error messages display correctly

### Mobile Tests
- [ ] Test on iPhone (Safari)
- [ ] Test on Android (Chrome)
- [ ] Touch targets are easy to tap (44px minimum)
- [ ] No horizontal scrolling
- [ ] Text is readable without zooming
- [ ] Buttons are full-width and tappable
- [ ] Forms are easy to fill on mobile keyboard

### Performance Tests
- [ ] Page loads in < 3 seconds (use Lighthouse in Chrome DevTools)
- [ ] Lighthouse Performance score > 90
- [ ] No console errors in browser
- [ ] API calls respond in < 2 seconds
- [ ] Gzip compression enabled (check Network tab)

### Security Tests
- [ ] HTTPS enforced (HTTP redirects)
- [ ] Authentication working (if configured)
- [ ] CORS headers prevent unauthorized domains
- [ ] No sensitive data in console logs
- [ ] No API keys in frontend code (check dist/assets/*)

---

## 🚨 Rollback Plan

If deployment fails:

### 1. Keep Old Files
```bash
# Before deployment, backup current version
ssh user@server 'cp -r /var/www/vue-admin /var/www/vue-admin.backup'
```

### 2. Rollback Steps
```bash
# Restore previous version
ssh user@server 'rm -rf /var/www/vue-admin'
ssh user@server 'mv /var/www/vue-admin.backup /var/www/vue-admin'

# Reload Nginx
ssh user@server 'sudo systemctl reload nginx'
```

- [ ] Backup created before deployment
- [ ] Rollback procedure tested

---

## 📊 Monitoring

### Log Locations
- **Nginx Access**: `/var/log/nginx/access.log`
- **Nginx Error**: `/var/log/nginx/error.log`
- **Go Backend**: Check systemd journal (`journalctl -u thaistockanalysis`)

### Check Commands
```bash
# Check Nginx status
sudo systemctl status nginx

# Check backend status
sudo systemctl status thaistockanalysis

# Tail logs in real-time
sudo tail -f /var/log/nginx/access.log
sudo journalctl -u thaistockanalysis -f
```

- [ ] Logs accessible
- [ ] No critical errors in logs
- [ ] Monitoring alerts configured (optional)

---

## 🔄 Post-Deployment Tasks

### Documentation
- [ ] Update production URL in README.md
- [ ] Document deployment date and version
- [ ] Share credentials with team (securely)
- [ ] Update API documentation if endpoints changed

### Backup
- [ ] Database backup scheduled (`data/admin.db`)
- [ ] Article files backup scheduled (`articles/*.md`)
- [ ] Frontend build archived (`dist.tar.gz`)

### Communication
- [ ] Notify team of deployment
- [ ] Share admin panel URL: `https://admin.thaistockanalysis.com`
- [ ] Share authentication credentials (if applicable)
- [ ] Schedule training session (if needed)

---

## 🎯 Performance Targets

### Baseline Metrics
Record after first deployment:

- **Load Time**: _____ seconds (target: < 3s)
- **Lighthouse Performance**: _____ (target: > 90)
- **Bundle Size**: _____ KB (current: ~150KB gzipped)
- **API Response Time**: _____ ms (target: < 2000ms)
- **Mobile Performance**: _____ (target: > 80)

### Optimization Checklist
- [ ] Enable gzip compression in Nginx
- [ ] Enable HTTP/2 in Nginx
- [ ] Set cache headers for static assets
- [ ] Use CDN for Tailwind CSS (already configured)
- [ ] Lazy load routes (already configured in Vue Router)

---

## 📝 Notes

### Common Issues & Solutions

**Issue**: 404 on refresh (e.g., `/article/2025-09-30`)
**Fix**: Ensure `try_files $uri /index.html` is in Nginx config

**Issue**: CORS errors in browser console
**Fix**: Add `Access-Control-Allow-Origin` header in Go backend

**Issue**: API calls return 502 Bad Gateway
**Fix**: Check Go backend is running: `sudo systemctl status thaistockanalysis`

**Issue**: SSL certificate error
**Fix**: Verify certificate paths in Nginx config, renew if expired

---

## ✅ Sign-Off

### Deployment Information
- **Date**: _______________
- **Version**: v1.0.0
- **Deployed By**: _______________
- **Production URL**: https://admin.thaistockanalysis.com
- **Backend URL**: https://thaistockanalysis.com

### Approval
- [ ] All tests passed
- [ ] Team notified
- [ ] Documentation updated
- [ ] Monitoring configured
- [ ] Backup scheduled

**Status**: ☐ Ready for Production | ☐ Deployed | ☐ Verified

---

**Next Deployment**: See `DEPLOYMENT.md` for detailed procedures
