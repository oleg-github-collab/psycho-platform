package main

import (
	"log"
	"os"
	"psycho-platform/internal/config"
	"psycho-platform/internal/database"
	"psycho-platform/internal/router"
	"psycho-platform/internal/websocket"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("=== Starting Psycho Platform API ===")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	log.Println("Loading configuration...")
	cfg := config.Load()
	log.Printf("Environment: %s", cfg.Environment)

	// Initialize database
	log.Println("Connecting to PostgreSQL...")
	db, err := database.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()
	log.Println("✓ PostgreSQL connected")

	// Run migrations
	log.Println("Running migrations...")
	if err := database.RunMigrations(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}
	log.Println("✓ Migrations completed")

	// Initialize Redis
	log.Println("Connecting to Redis...")
	redisClient, err := database.NewRedisClient(cfg.RedisURL)
	if err != nil {
		log.Printf("WARNING: Failed to connect to Redis: %v", err)
		log.Println("Continuing without Redis (rate limiting disabled)")
		redisClient = nil
	} else {
		defer redisClient.Close()
		log.Println("✓ Redis connected")
	}

	// Initialize WebSocket hub
	log.Println("Initializing WebSocket hub...")
	hub := websocket.NewHub()
	go hub.Run()
	log.Println("✓ WebSocket hub running")

	// Setup router
	log.Println("Setting up routes...")
	r := router.Setup(db, redisClient, hub, cfg)
	log.Println("✓ Routes configured")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("✓ Server starting on port %s", port)
	log.Println("=== API Ready ===")
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
