package router

import (
	"database/sql"
	"net/http"
	"psycho-platform/internal/config"
	"psycho-platform/internal/handlers"
	"psycho-platform/internal/middleware"
	"psycho-platform/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Setup(db *sql.DB, redis *redis.Client, hub *websocket.Hub, cfg *config.Config) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.Use(middleware.CORS(cfg.FrontendURL))

	// Static files
	r.Static("/static", "./web/static")
	r.StaticFile("/", "./web/index.html")
	r.NoRoute(func(c *gin.Context) {
		c.File("./web/index.html")
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, cfg)
	topicHandler := handlers.NewTopicHandler(db)
	messageHandler := handlers.NewMessageHandler(db, hub)
	groupHandler := handlers.NewGroupHandler(db)
	sessionHandler := handlers.NewSessionHandler(db, cfg)
	appointmentHandler := handlers.NewAppointmentHandler(db)
	adminHandler := handlers.NewAdminHandler(db)

	// Public routes
	api := r.Group("/api")
	{
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		// Auth
		protected.GET("/auth/me", authHandler.GetMe)

		// Topics
		protected.GET("/topics", topicHandler.GetTopics)
		protected.POST("/topics", topicHandler.CreateTopic)
		protected.POST("/topics/:id/vote", topicHandler.VoteTopic)

		// Messages
		protected.GET("/messages", messageHandler.GetMessages)
		protected.POST("/messages", messageHandler.CreateMessage)
		protected.POST("/messages/:id/reactions", messageHandler.AddReaction)
		protected.DELETE("/messages/:id/reactions", messageHandler.RemoveReaction)

		// Groups
		protected.GET("/groups", groupHandler.GetGroups)
		protected.POST("/groups", groupHandler.CreateGroup)
		protected.POST("/groups/:id/join", groupHandler.JoinGroup)
		protected.POST("/groups/:id/leave", groupHandler.LeaveGroup)

		// Sessions
		protected.GET("/sessions", sessionHandler.GetSessions)
		protected.POST("/sessions", sessionHandler.CreateSession)
		protected.GET("/sessions/:id/token", sessionHandler.GetRoomToken)

		// Appointments
		protected.GET("/appointments", appointmentHandler.GetAppointments)
		protected.POST("/appointments", appointmentHandler.CreateAppointment)
		protected.PATCH("/appointments/:id/status", appointmentHandler.UpdateAppointmentStatus)

		// WebSocket
		protected.GET("/ws", func(c *gin.Context) {
			userID := c.GetString("user_id")
			conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
			if err != nil {
				return
			}
			websocket.ServeWs(hub, conn, userID)
		})
	}

	// Admin routes
	admin := api.Group("/admin")
	admin.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	admin.Use(middleware.AdminOnly())
	{
		admin.GET("/stats", adminHandler.GetStats)
		admin.GET("/users", adminHandler.GetUsers)
		admin.PATCH("/users/:id/status", adminHandler.ToggleUserStatus)
		admin.PATCH("/users/:id/psychologist", adminHandler.SetPsychologist)
	}

	return r
}
