package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	configpkg "thaistockanalysis/configs"
	"thaistockanalysis/internal/database"
	"thaistockanalysis/internal/handlers"
)

func main() {
	// Load configuration
	cfg := configpkg.LoadConfig()

	// --- TEMPORARY CODE TO LIST MODELS ---
	listModelsURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", cfg.GeminiAPIKey)
	resp, err := http.Get(listModelsURL)
	if err != nil {
		log.Fatalf("Failed to call Gemini ListModels API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read ListModels response: %v", err)
	}
	log.Printf("--- Available Gemini Models ---\n%s\n-----------------------------\n", string(body))
	// --- END TEMPORARY CODE ---

	// Initialize database
	dbPath := filepath.Join(cfg.DatabasePath)
	err = database.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Sync filesystem articles to database
	database.AddMissingArticlesToDB(cfg.ArticlesDir)

	// Initialize handlers
	h := handlers.NewHandler(cfg.ArticlesDir, cfg.TemplateDir, cfg)

	// Create HTTP server
	mux := http.NewServeMux()

	// Serve static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(cfg.StaticDir))))

	// Routes
	mux.HandleFunc("/", h.IndexHandler)
	mux.HandleFunc("/articles/", h.ArticleHandler)
	
	// Redirect admin routes to homepage - use Vue admin panel on port 3000 instead
	mux.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	})
	mux.HandleFunc("/admin/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	})

	// About page
	mux.HandleFunc("/about", h.AboutHandler)

	// Contact page
	mux.HandleFunc("/contact", h.ContactHandler)

	// API endpoints for market data
	mux.HandleFunc("/api/articles", h.ArticlesAPIHandler)
	mux.HandleFunc("/api/articles/", h.ArticleAPIHandler)
	mux.HandleFunc("/api/market-data-analysis", h.MarketDataAnalysisHandler)
	mux.HandleFunc("/api/market-data-close", h.MarketDataCloseHandler)

	// Legal pages
	mux.HandleFunc("/privacy", h.PrivacyHandler)
	mux.HandleFunc("/terms", h.TermsHandler)
	mux.HandleFunc("/disclaimer", h.DisclaimerHandler)

	// Create server with timeouts
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("🚀 ThaiStockAnalysis server starting on http://localhost:%s", cfg.Port)
		log.Printf("📊 Admin dashboard: http://localhost:%s/admin", cfg.Port)
		log.Printf("🏠 Homepage: http://localhost:%s", cfg.Port)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Server is shutting down...")

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited")
}
