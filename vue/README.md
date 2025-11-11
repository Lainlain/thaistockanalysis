# Thai Stock Analysis - Vue Admin Panel

A modern Vue 3 admin panel for managing Thai stock market analysis articles. This application provides a clean interface to create and update daily market data that feeds into the Go backend for AI-powered analysis generation.

## 🌟 Features

- **📊 Article List**: View all existing stock market analysis articles
- **✏️ Article Editor**: Update existing articles with new market data
- **➕ Create New Articles**: Create new daily analysis with auto-date selection
- **🕐 Four Trading Sessions**: Manage morning/afternoon opening and closing data separately
- **🤖 AI Integration**: Automatically triggers Gemini AI analysis generation
- **📱 Responsive Design**: Beautiful Tailwind CSS interface that works on all devices

## 🚀 Quick Start

### Prerequisites

- Node.js 18+ and npm
- Thai Stock Analysis Go server running on `http://localhost:7777`

### Installation

1. **Navigate to the vue folder:**
   ```bash
   cd vue
   ```

2. **Install dependencies:**
   ```bash
   npm install
   ```

3. **Start development server:**
   ```bash
   npm run dev
   ```

4. **Open your browser:**
   - Vue Admin Panel: [http://localhost:3000](http://localhost:3000)
   - Go Backend: [http://localhost:7777](http://localhost:7777) (must be running)

## 📖 Usage Guide

### Creating a New Article

1. Click **"+ New Article"** button in the app bar
2. The date picker will auto-select today's date (you can change it if needed)
3. Fill in each section separately:

#### Morning Session
- **Morning Open**: 3 fields (Index, Change, Highlights)
  - Example: Index: `1287.01`, Change: `4.47`, Highlights: `7 => +79, +75, +78 :: 4 => +49, +45, +48`
- **Morning Close**: 2 fields (Index, Change)
  - Example: Index: `1281.04`, Change: `-1.50`

#### Afternoon Session
- **Afternoon Open**: 3 fields (Index, Change, Highlights)
  - Example: Index: `1279.48`, Change: `-8.59`, Highlights: `7 => +79, +75, +78 :: 4 => +49, +45, +48`
- **Afternoon Close**: 2 fields (Index, Change)
  - Example: Index: `1275.20`, Change: `-3.28`

4. Click the submit button for each section independently
5. The Go backend will automatically generate AI analysis using Gemini

### Updating Existing Articles

1. Click on any article from the main list
2. Fill in the fields you want to update
3. Click the respective submit button for that section
4. Changes will be sent to the Go backend immediately

### Highlights Format

The highlights field accepts this format:
```
7 => +79 , +75 , +78 , +70 , +73 , +76 :: 4 => +49 , +45 , +48 , +40 , +43 , +46
```

This represents sector performance indicators that the AI will use to generate market narrative.

## 🏗️ Project Structure

```
vue/
├── index.html              # HTML entry point with Tailwind CDN
├── package.json            # Dependencies and scripts
├── vite.config.js          # Vite configuration with API proxy
└── src/
    ├── main.js             # Vue app initialization
    ├── App.vue             # Root component with navigation bar
    ├── router/
    │   └── index.js        # Vue Router configuration
    ├── services/
    │   └── api.js          # Axios API service layer
    └── views/
        ├── ArticleList.vue     # Homepage showing all articles
        ├── ArticleDetail.vue   # Article detail/edit page
        └── CreateArticle.vue   # New article creation form
```

## 🔌 API Integration

The Vue app communicates with these Go backend endpoints:

### Market Opening Data
```bash
POST /api/market-data-analysis
Content-Type: application/json

{
  "date": "2025-11-11",
  "morning_open": {
    "index": 1287.01,
    "change": 4.47,
    "highlights": "7 => +79, +75, +78 :: 4 => +49, +45, +48"
  }
}
```

### Market Closing Data
```bash
POST /api/market-data-close
Content-Type: application/json

{
  "date": "2025-11-11",
  "morning_close": {
    "index": 1281.04,
    "change": -1.50
  }
}
```

## 🛠️ Development

### Available Scripts

- `npm run dev` - Start development server on port 3000
- `npm run build` - Build for production
- `npm run preview` - Preview production build

### Technology Stack

- **Vue 3** - Progressive JavaScript framework
- **Vue Router 4** - Official routing library
- **Axios** - Promise-based HTTP client
- **Vite** - Next generation frontend tooling
- **Tailwind CSS** - Utility-first CSS framework (CDN)

### Proxy Configuration

The Vite dev server proxies `/api` requests to `http://localhost:7777` to avoid CORS issues during development. This is configured in `vite.config.js`.

## 🐛 Troubleshooting

### "Failed to load articles"
- **Solution**: Make sure the Go server is running on port 7777
  ```bash
  cd ..
  go run cmd/server/main.go
  ```

### "Network Error" when submitting data
- **Solution**: Check that both servers are running:
  - Vue: `http://localhost:3000` (this app)
  - Go: `http://localhost:7777` (backend API)

### Changes not appearing
- **Solution**: The Go backend generates markdown files in `articles/YYYY-MM-DD.md`. Check that directory for the generated content.

## 📝 Notes

- The article list currently shows sample data. In production, you would implement a proper API endpoint to fetch articles from the Go backend's SQLite database.
- Each section (morning_open, morning_close, afternoon_open, afternoon_close) is submitted independently to allow flexibility in updating articles throughout the trading day.
- The Go backend uses Google Gemini AI to generate market analysis automatically based on the data you submit.

## 🔗 Related Documentation

- Go Backend API: See `/docs/API_QUICK_REFERENCE.md` in the parent directory
- Gemini Prompt Instructions: See `/gemini` file for AI analysis prompt template
- Backend Instructions: See `/.github/copilot-instructions.md` for full architecture details

## 📄 License

Part of the Thai Stock Analysis project.

---

**Built with ❤️ for Thai stock market analysis**
