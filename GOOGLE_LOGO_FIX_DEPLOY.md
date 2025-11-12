# 🔧 GOOGLE SEARCH LOGO FIX - DEPLOYMENT REQUIRED

**Date**: November 12, 2025
**Status**: ⚠️ CHANGES MADE LOCALLY - NEEDS PRODUCTION DEPLOYMENT

---

## 🔍 Problem Identified

Your logo shows on the website but NOT in Google Search because:

1. **Production site hasn't been updated** with the new code
2. **Logo file returns 404** on production: `https://thaistockanalysis.com/static/logo.png`
3. **Old JSON-LD schema** - missing Organization with logo
4. **Old og:image meta tag** - pointing to non-existent `/static/images/og-image.jpg`

### What Production Currently Has (OLD)
```html
<!-- ❌ Old og:image (doesn't exist) -->
<meta property="og:image" content="/static/images/og-image.jpg">

<!-- ❌ Old JSON-LD (no Organization, no logo) -->
<script type="application/ld+json">
{
  "@type": "WebSite",
  "name": "Thai Stock Analysis",
  ...
  <!-- NO LOGO PROPERTY -->
}
</script>
```

---

## ✅ Changes Made (IN LOCAL FILES)

### 1. Updated `web/templates/base.gohtml`

**New og:image with absolute URL:**
```html
<meta property="og:image" content="https://thaistockanalysis.com/static/logo.png">
```

**New JSON-LD with Organization schema:**
```json
{
  "@context": "https://schema.org",
  "@graph": [
    {
      "@type": "Organization",
      "name": "Thai Stock Analysis",
      "url": "https://thaistockanalysis.com",
      "logo": "https://thaistockanalysis.com/static/logo.png"  ← GOOGLE LOOKS FOR THIS
    },
    {
      "@type": "WebSite",
      ...
    }
  ]
}
```

**New favicon/icon links:**
```html
<link rel="icon" type="image/png" sizes="192x192" href="/static/logo.png">
<link rel="apple-touch-icon" sizes="180x180" href="/static/logo.png">
<link rel="manifest" href="/static/site.webmanifest">
```

### 2. Created `web/static/site.webmanifest`
```json
{
  "name": "Thai Stock Analysis",
  "icons": [
    {
      "src": "/static/logo.png",
      "sizes": "192x192",
      "type": "image/png"
    }
  ]
}
```

### 3. Created `web/static/robots.txt`
```
User-agent: *
Allow: /
Allow: /static/
Sitemap: https://thaistockanalysis.com/sitemap.xml
```

### 4. Updated `cmd/server/main.go`
Added robots.txt handler to serve the file properly.

---

## 🚀 DEPLOYMENT STEPS (YOU MUST DO THIS)

### Step 1: Test Locally First
```bash
# Start your Go server
go run cmd/server/main.go

# In another terminal, test the changes:
curl -I http://localhost:7777/static/logo.png
# Should return: HTTP/1.1 200 OK

curl -I http://localhost:7777/robots.txt
# Should return: HTTP/1.1 200 OK

curl -s http://localhost:7777/ | grep -A 15 'application/ld+json'
# Should show the NEW JSON-LD with Organization and logo

curl -s http://localhost:7777/ | grep 'og:image'
# Should show: content="https://thaistockanalysis.com/static/logo.png"
```

### Step 2: Build and Deploy to Production

**Option A: Docker Deployment**
```bash
# Build new Docker image
docker build -t thaistockanalysis .

# Stop old container
docker stop thaistockanalysis

# Start new container
docker run -d --name thaistockanalysis \
  -p 7777:7777 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/articles:/app/articles \
  thaistockanalysis
```

**Option B: Direct Build**
```bash
# Build binary
go build -o bin/thaistockanalysis cmd/server/main.go

# Copy to production server (replace with your actual server)
scp bin/thaistockanalysis user@thaistockanalysis.com:/path/to/app/
scp -r web/templates user@thaistockanalysis.com:/path/to/app/web/
scp -r web/static user@thaistockanalysis.com:/path/to/app/web/

# SSH to server and restart
ssh user@thaistockanalysis.com
sudo systemctl restart thaistockanalysis
```

**Option C: Git + Pull on Server**
```bash
# Commit changes locally
git add web/templates/base.gohtml web/static/robots.txt web/static/site.webmanifest cmd/server/main.go
git commit -m "Add Google Search logo with Organization schema and robots.txt"
git push origin main

# On production server
ssh user@thaistockanalysis.com
cd /path/to/app
git pull origin main
go build -o bin/thaistockanalysis cmd/server/main.go
sudo systemctl restart thaistockanalysis
```

### Step 3: Verify Production After Deployment
```bash
# Test logo is accessible
curl -I https://thaistockanalysis.com/static/logo.png
# Must return: HTTP/2 200

# Test robots.txt
curl https://thaistockanalysis.com/robots.txt
# Should show the robots.txt content

# Test JSON-LD in production
curl -s https://thaistockanalysis.com/ | grep -A 20 'application/ld+json'
# Should show Organization with logo

# Test og:image
curl -s https://thaistockanalysis.com/ | grep 'og:image'
# Should show: https://thaistockanalysis.com/static/logo.png
```

---

## 🧪 Google Validation Steps

### After Production Deployment, Test With Google Tools:

### 1. Rich Results Test
1. Go to: https://search.google.com/test/rich-results
2. Enter: `https://thaistockanalysis.com`
3. Click "Test URL"
4. **Expected**: Should detect Organization schema with logo

### 2. Google Search Console (Recommended)
1. Go to: https://search.google.com/search-console
2. Add/verify your property (if not already done)
3. Use **URL Inspection Tool**:
   - Enter: `https://thaistockanalysis.com`
   - Click "Test Live URL"
   - Should show the new JSON-LD with logo
4. Click **"Request Indexing"** to speed up Google's reprocessing
5. Submit/update your **sitemap.xml**

### 3. Schema Markup Validator
1. Go to: https://validator.schema.org/
2. Enter: `https://thaistockanalysis.com`
3. **Expected**: Should validate Organization schema with logo property

---

## 📊 Logo Requirements for Google

Make sure your `logo.png` meets these requirements:

✅ **Minimum size**: 112 × 112 pixels (Google requirement)
✅ **Recommended**: Square format (1:1 aspect ratio)
✅ **Format**: PNG or high-quality JPG
✅ **URL**: Must be absolute and publicly accessible
✅ **No blocking**: Not blocked by robots.txt (we added Allow: /static/)

### Check Your Logo Size
```bash
# Install imagemagick if not already installed
# Ubuntu/Debian: sudo apt install imagemagick
# macOS: brew install imagemagick

# Check logo dimensions
identify web/static/logo.png
# Should show dimensions like: logo.png PNG 192x192 ...
```

If your logo is smaller than 112x112, you'll need a larger version.

---

## ⏱️ Timeline for Google to Show Logo

After deploying and requesting indexing:

- **Rich Results Test**: Immediate (shows what Google sees now)
- **Search Console**: 1-3 days for reprocessing
- **Live Search Results**: 3-7 days typically, can be up to 2 weeks

**Speed it up**:
1. Use Search Console "Request Indexing" ✅
2. Share your site on social media (triggers crawls)
3. Submit sitemap with homepage priority
4. Ensure homepage is linked from other pages

---

## 🎯 Summary Checklist

### Before Deployment ✅ (Done Locally)
- [x] Updated `web/templates/base.gohtml` with Organization JSON-LD
- [x] Changed og:image to absolute URL with logo.png
- [x] Added favicon/icon links
- [x] Created `web/static/site.webmanifest`
- [x] Created `web/static/robots.txt`
- [x] Added robots.txt handler to `cmd/server/main.go`

### You Must Do (Production Deployment)
- [ ] Test changes locally first
- [ ] Deploy updated code to production server
- [ ] Verify logo.png is accessible: `https://thaistockanalysis.com/static/logo.png`
- [ ] Verify robots.txt works: `https://thaistockanalysis.com/robots.txt`
- [ ] Check homepage source has new JSON-LD and og:image
- [ ] Test with Google Rich Results Test
- [ ] Request indexing in Search Console
- [ ] Verify logo dimensions are at least 112×112px

---

## 🆘 Troubleshooting

### Logo Still Returns 404 After Deployment
**Check**:
- Is `web/static/logo.png` in the production server files?
- Is the static file handler configured correctly?
- Check server logs for errors

### Google Still Doesn't Show Logo After 1 Week
**Check**:
1. View page source of `https://thaistockanalysis.com`
   - Confirm Organization JSON-LD is present
   - Confirm logo URL is absolute and correct
2. Test the logo URL directly in browser
   - Must be publicly accessible (not 404)
3. Use Rich Results Test
   - Should detect the Organization schema
4. Check logo image size
   - Must be at least 112×112 pixels
5. Request indexing again in Search Console

### Logo Shows in Rich Results Test But Not in Search
**This is normal!** Google takes time to update search results:
- Test shows what Google CAN see (validates markup)
- Actual search results update in 3-7 days
- Keep requesting indexing in Search Console

---

## 📝 Next Steps

1. **DEPLOY TO PRODUCTION** (most important!)
2. Test with the verification steps above
3. Request indexing in Search Console
4. Monitor Search Console for schema detection
5. Wait 3-7 days for Google to show logo in search

---

**Status**: ⚠️ LOCAL CHANGES READY - WAITING FOR PRODUCTION DEPLOYMENT
**Files Changed**: 4 files (base.gohtml, site.webmanifest, robots.txt, main.go)
**Action Required**: DEPLOY TO PRODUCTION NOW

---

## Quick Deploy Commands

```bash
# 1. Test locally
go run cmd/server/main.go

# 2. Build for production
go build -o bin/thaistockanalysis cmd/server/main.go

# 3. Deploy (choose your method above)

# 4. Verify after deploy
curl -I https://thaistockanalysis.com/static/logo.png
curl -s https://thaistockanalysis.com/ | grep -A 20 'application/ld+json'
```
