package main

import (
	"context"
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

	// Initialize database
	dbPath := filepath.Join(cfg.DatabasePath)
	err := database.InitDB(dbPath)
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
	mux.HandleFunc("/admin", h.AdminDashboardHandler)
	mux.HandleFunc("/admin/", h.AdminDashboardHandler)
	mux.HandleFunc("/admin/articles/new", h.AdminArticleFormHandler)

	// About page
	mux.HandleFunc("/about", h.AboutHandler)

	// Contact page
	mux.HandleFunc("/contact", h.ContactHandler)

	// API endpoints for market data
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
