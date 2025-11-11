# 🚀 Quick Start Guide - Vue Admin Panel

## Step 1: Install Dependencies

```bash
cd vue
npm install
```

## Step 2: Start the Go Backend

In a separate terminal:

```bash
cd ..
go run cmd/server/main.go
```

The Go server should start on **http://localhost:7777**

## Step 3: Start the Vue App

```bash
npm run dev
```

The Vue app will start on **http://localhost:3000**

## Step 4: Create Your First Article

1. Open http://localhost:3000 in your browser
2. Click **"+ New Article"** button
3. Today's date is auto-selected (you can change it)
4. Fill in Morning Opening data:
   - Index: `1287.01`
   - Change: `4.47`
   - Highlights: `7 => +79 , +75 , +78 , +70 , +73 , +76 :: 4 => +49 , +45 , +48 , +40 , +43 , +46`
5. Click **"Submit Morning Open"**
6. Wait for success message ✅
7. Repeat for:
   - Morning Close (Index + Change only)
   - Afternoon Open (Index + Change + Highlights)
   - Afternoon Close (Index + Change only)

## 📝 What Happens Behind the Scenes

When you submit data:

1. **Vue App** → Sends POST request to Go backend
2. **Go Backend** → Receives data + calls Gemini AI
3. **Gemini AI** → Generates professional market analysis
4. **Go Backend** → Saves to `articles/YYYY-MM-DD.md` file
5. **Go Backend** → Sends Telegram notification
6. **Article Created** → View it on the main site at http://localhost:7777

## 🎯 API Format Examples

### Morning Open Request
```json
{
  "date": "2025-11-11",
  "morning_open": {
    "index": 1287.01,
    "change": 4.47,
    "highlights": "7 => +79 , +75 , +78 :: 4 => +49 , +45 , +48"
  }
}
```

### Morning Close Request
```json
{
  "date": "2025-11-11",
  "morning_close": {
    "index": 1281.04,
    "change": -1.50
  }
}
```

## ✅ Verify Everything Works

1. Check Vue app runs: http://localhost:3000
2. Check Go backend runs: http://localhost:7777
3. Submit test data via Vue admin
4. Check `articles/2025-11-11.md` file was created
5. View article on main site: http://localhost:7777/articles/2025-11-11

## 🐛 Common Issues

**Vue app won't start:**
```bash
# Make sure you're in the vue directory
cd vue
npm install
npm run dev
```

**Can't connect to backend:**
```bash
# Make sure Go server is running
cd ..
go run cmd/server/main.go
```

**"Network Error":**
- Check both servers are running (Vue on :3000, Go on :7777)
- Check console for detailed error messages

## 📚 Next Steps

- Read the full README.md for detailed documentation
- Check the Go backend API docs in `/docs/API_QUICK_REFERENCE.md`
- Explore the Gemini prompt template in `/gemini` file

---

**Happy managing! 📈**
