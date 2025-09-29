package config

import (
	"log"
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	Port             string
	DatabasePath     string
	ArticlesDir      string
	TemplateDir      string
	StaticDir        string
	DebugMode        bool
	CacheExpiry      int // minutes
	GeminiAPIKey     string
	TelegramBotToken string
	TelegramChannel  string
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	config := &Config{
		Port:             getEnv("PORT", "7777"),
		DatabasePath:     getEnv("DATABASE_PATH", "data/admin.db"),
		ArticlesDir:      getEnv("ARTICLES_DIR", "articles"),
		TemplateDir:      getEnv("TEMPLATE_DIR", "web/templates"),
		StaticDir:        getEnv("STATIC_DIR", "web/static"),
		DebugMode:        getEnvBool("DEBUG_MODE", false),
		CacheExpiry:      getEnvInt("CACHE_EXPIRY", 5),
		GeminiAPIKey:     getEnv("GEMINI_API_KEY", "AIzaSyBkw_fi16Q39yjZdZ0C3PTw-vuADTR-KAM"),
		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", "7912088515:AAFn3YbnE-84MmMgvhoc6vpJ5HiLPtH5IEg"),
		TelegramChannel:  getEnv("TELEGRAM_CHANNEL", "-1002240874831"),
	}

	log.Printf("Configuration loaded:")
	log.Printf("  Port: %s", config.Port)
	log.Printf("  Database: %s", config.DatabasePath)
	log.Printf("  Articles: %s", config.ArticlesDir)
	log.Printf("  Templates: %s", config.TemplateDir)
	log.Printf("  Static: %s", config.StaticDir)
	log.Printf("  Debug: %t", config.DebugMode)
	log.Printf("  Cache Expiry: %d minutes", config.CacheExpiry)
	log.Printf("  Gemini API Key: %s", config.GeminiAPIKey[:10]+"...")
	log.Printf("  Telegram Bot: %s", config.TelegramBotToken[:10]+"...")
	log.Printf("  Telegram Channel: %s", config.TelegramChannel)

	return config
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool gets boolean environment variable with default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getEnvInt gets integer environment variable with default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
