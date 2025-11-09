package handlers

import (
	"database/sql"
	"net/http"
	"psycho-platform/internal/websocket"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	db  *sql.DB
	hub *websocket.Hub
}

func NewNotificationHandler(db *sql.DB, hub *websocket.Hub) *NotificationHandler {
	return &NotificationHandler{db: db, hub: hub}
}

type Notification struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Link      string `json:"link"`
	IsRead    bool   `json:"is_read"`
	CreatedAt string `json:"created_at"`
}

func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	userID := c.GetString("user_id")
	limit := c.DefaultQuery("limit", "50")

	rows, err := h.db.Query(`
		SELECT id, user_id, type, title, content, link, is_read, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
	}
	defer rows.Close()

	notifications := []Notification{}
	for rows.Next() {
		var n Notification
		rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Content, &n.Link, &n.IsRead, &n.CreatedAt)
		notifications = append(notifications, n)
	}

	c.JSON(http.StatusOK, notifications)
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID := c.GetString("user_id")
	notificationID := c.Param("id")

	_, err := h.db.Exec(`
		UPDATE notifications
		SET is_read = true
		WHERE id = $1 AND user_id = $2
	`, notificationID, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := c.GetString("user_id")

	_, err := h.db.Exec(`
		UPDATE notifications
		SET is_read = true
		WHERE user_id = $1 AND is_read = false
	`, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark all as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID := c.GetString("user_id")

	var count int
	err := h.db.QueryRow(`
		SELECT COUNT(*) FROM notifications
		WHERE user_id = $1 AND is_read = false
	`, userID).Scan(&count)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	userID := c.GetString("user_id")
	notificationID := c.Param("id")

	_, err := h.db.Exec(`
		DELETE FROM notifications
		WHERE id = $1 AND user_id = $2
	`, notificationID, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Helper function to create notification
func (h *NotificationHandler) CreateNotification(userID, notifType, title, content, link string) error {
	var notifID string
	err := h.db.QueryRow(`
		INSERT INTO notifications (user_id, type, title, content, link)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, userID, notifType, title, content, link).Scan(&notifID)

	if err != nil {
		return err
	}

	// Send via WebSocket
	h.hub.BroadcastToRoom("user_"+userID, map[string]interface{}{
		"type": "notification",
		"payload": map[string]interface{}{
			"id":      notifID,
			"type":    notifType,
			"title":   title,
			"content": content,
			"link":    link,
		},
	})

	return nil
}
