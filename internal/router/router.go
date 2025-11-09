package router

import (
	"database/sql"
	"net/http"
	"psycho-platform/internal/config"
	"psycho-platform/internal/handlers"
	"psycho-platform/internal/middleware"
	"psycho-platform/internal/websocket"

	"github.com/gin-gonic/gin"
	gorilla "github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var upgrader = gorilla.Upgrader{
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

	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(cfg.FrontendURL))

	// Rate limiting
	rateLimiter := middleware.NewRateLimiter(redis, 60) // 60 requests per minute
	r.Use(rateLimiter.Middleware())

	// Static files
	r.Static("/static", "./web/static")
	r.Static("/uploads", "./uploads")
	r.StaticFile("/", "./web/index.html")
	r.NoRoute(func(c *gin.Context) {
		c.File("./web/index.html")
	})

	// Health checks
	healthHandler := handlers.NewHealthHandler(db, redis)
	r.GET("/health", healthHandler.Check)
	r.GET("/ready", healthHandler.Ready)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, cfg)
	topicHandler := handlers.NewTopicHandler(db)
	messageHandler := handlers.NewMessageHandler(db, hub)
	groupHandler := handlers.NewGroupHandler(db)
	sessionHandler := handlers.NewSessionHandler(db, cfg)
	appointmentHandler := handlers.NewAppointmentHandler(db)
	adminHandler := handlers.NewAdminHandler(db)
	profileHandler := handlers.NewProfileHandler(db)
	dmHandler := handlers.NewDMHandler(db, hub)
	notificationHandler := handlers.NewNotificationHandler(db, hub)
	fileHandler := handlers.NewFileHandler(db)
	searchHandler := handlers.NewSearchHandler(db)
	bookmarkHandler := handlers.NewBookmarkHandler(db)
	activityHandler := handlers.NewActivityHandler(db)

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

		// Profile
		protected.PATCH("/profile", profileHandler.UpdateProfile)
		protected.GET("/profile/:id", profileHandler.GetUserProfile)
		protected.GET("/users/search", profileHandler.SearchUsers)
		protected.POST("/users/:id/block", profileHandler.BlockUser)
		protected.DELETE("/users/:id/block", profileHandler.UnblockUser)
		protected.GET("/users/blocked", profileHandler.GetBlockedUsers)
		protected.POST("/status/online", profileHandler.SetOnlineStatus)

		// Direct Messages
		protected.GET("/conversations", dmHandler.GetConversations)
		protected.POST("/conversations/send", dmHandler.SendDirectMessage)
		protected.GET("/conversations/:id/messages", dmHandler.GetMessages)
		protected.POST("/conversations/:id/read", dmHandler.MarkAsRead)

		// Topics
		protected.GET("/topics", topicHandler.GetTopics)
		protected.POST("/topics", topicHandler.CreateTopic)
		protected.POST("/topics/:id/vote", topicHandler.VoteTopic)

		// Messages
		protected.GET("/messages", messageHandler.GetMessages)
		protected.POST("/messages", messageHandler.CreateMessage)
		protected.PATCH("/messages/:id", messageHandler.EditMessage)
		protected.DELETE("/messages/:id", messageHandler.DeleteMessage)
		protected.POST("/messages/:id/reactions", messageHandler.AddReaction)
		protected.DELETE("/messages/:id/reactions", messageHandler.RemoveReaction)
		protected.POST("/messages/:id/read", messageHandler.MarkAsRead)
		protected.POST("/messages/typing/start", messageHandler.StartTyping)
		protected.POST("/messages/typing/stop", messageHandler.StopTyping)

		// Groups
		protected.GET("/groups", groupHandler.GetGroups)
		protected.POST("/groups", groupHandler.CreateGroup)
		protected.POST("/groups/:id/join", groupHandler.JoinGroup)
		protected.POST("/groups/:id/leave", groupHandler.LeaveGroup)
		protected.POST("/groups/:id/invite", groupHandler.CreateInvitation)
		protected.POST("/groups/join/:code", groupHandler.JoinByInvitation)
		protected.PATCH("/groups/:id/members/:member_id/role", groupHandler.UpdateMemberRole)
		protected.DELETE("/groups/:id/members/:member_id", groupHandler.RemoveMember)

		// Topics enhanced
		protected.POST("/topics/:id/pin", groupHandler.PinTopic)
		protected.DELETE("/topics/:id/pin", groupHandler.UnpinTopic)

		// Sessions
		protected.GET("/sessions", sessionHandler.GetSessions)
		protected.POST("/sessions", sessionHandler.CreateSession)
		protected.GET("/sessions/:id/token", sessionHandler.GetRoomToken)

		// Appointments
		protected.GET("/appointments", appointmentHandler.GetAppointments)
		protected.POST("/appointments", appointmentHandler.CreateAppointment)
		protected.PATCH("/appointments/:id/status", appointmentHandler.UpdateAppointmentStatus)

		// Notifications
		protected.GET("/notifications", notificationHandler.GetNotifications)
		protected.POST("/notifications/:id/read", notificationHandler.MarkAsRead)
		protected.POST("/notifications/read-all", notificationHandler.MarkAllAsRead)
		protected.GET("/notifications/unread-count", notificationHandler.GetUnreadCount)
		protected.DELETE("/notifications/:id", notificationHandler.DeleteNotification)

		// Files
		protected.POST("/upload", fileHandler.UploadFile)
		protected.POST("/messages/:id/attach", fileHandler.AttachToMessage)
		protected.GET("/messages/:id/files", fileHandler.GetMessageFiles)
		protected.DELETE("/files/:id", fileHandler.DeleteFile)

		// Search
		protected.GET("/search", searchHandler.GlobalSearch)
		protected.GET("/search/messages", searchHandler.SearchMessages)

		// Bookmarks
		protected.POST("/messages/:id/bookmark", bookmarkHandler.AddBookmark)
		protected.DELETE("/messages/:id/bookmark", bookmarkHandler.RemoveBookmark)
		protected.GET("/bookmarks", bookmarkHandler.GetBookmarks)
		protected.GET("/messages/:id/is-bookmarked", bookmarkHandler.IsBookmarked)

		// Activity
		protected.GET("/activity", activityHandler.GetActivityFeed)
		protected.GET("/trending", activityHandler.GetTrendingTopics)

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
