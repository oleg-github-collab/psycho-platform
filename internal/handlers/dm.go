package handlers

import (
	"database/sql"
	"net/http"
	"psycho-platform/internal/websocket"

	"github.com/gin-gonic/gin"
)

type DMHandler struct {
	db  *sql.DB
	hub *websocket.Hub
}

func NewDMHandler(db *sql.DB, hub *websocket.Hub) *DMHandler {
	return &DMHandler{db: db, hub: hub}
}

type CreateDMRequest struct {
	RecipientID string `json:"recipient_id" binding:"required"`
	Content     string `json:"content" binding:"required"`
}

func (h *DMHandler) SendDirectMessage(c *gin.Context) {
	userID := c.GetString("user_id")
	var req CreateDMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user is blocked
	var isBlocked bool
	h.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM user_blocks
			WHERE user_id = $1 AND blocked_user_id = $2
		)
	`, req.RecipientID, userID).Scan(&isBlocked)

	if isBlocked {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are blocked by this user"})
		return
	}

	// Create conversation if not exists
	var conversationID string
	err := h.db.QueryRow(`
		INSERT INTO conversations (user1_id, user2_id)
		SELECT $1, $2
		WHERE NOT EXISTS (
			SELECT 1 FROM conversations
			WHERE (user1_id = $1 AND user2_id = $2)
			   OR (user1_id = $2 AND user2_id = $1)
		)
		RETURNING id
	`, userID, req.RecipientID).Scan(&conversationID)

	if err != nil {
		// Conversation exists, get it
		h.db.QueryRow(`
			SELECT id FROM conversations
			WHERE (user1_id = $1 AND user2_id = $2)
			   OR (user1_id = $2 AND user2_id = $1)
		`, userID, req.RecipientID).Scan(&conversationID)
	}

	// Create message
	var message struct {
		ID             string `json:"id"`
		ConversationID string `json:"conversation_id"`
		SenderID       string `json:"sender_id"`
		Content        string `json:"content"`
		IsRead         bool   `json:"is_read"`
		CreatedAt      string `json:"created_at"`
	}

	err = h.db.QueryRow(`
		INSERT INTO direct_messages (conversation_id, sender_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, conversation_id, sender_id, content, is_read, created_at
	`, conversationID, userID, req.Content).Scan(
		&message.ID, &message.ConversationID, &message.SenderID,
		&message.Content, &message.IsRead, &message.CreatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	// Update conversation timestamp
	h.db.Exec(`
		UPDATE conversations
		SET last_message_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, conversationID)

	// Broadcast via WebSocket
	h.hub.BroadcastToRoom("dm_"+req.RecipientID, map[string]interface{}{
		"type":    "new_dm",
		"payload": message,
	})

	c.JSON(http.StatusCreated, message)
}

func (h *DMHandler) GetConversations(c *gin.Context) {
	userID := c.GetString("user_id")

	rows, err := h.db.Query(`
		SELECT c.id, c.last_message_at,
		       u.id, u.username, u.display_name, u.avatar_url,
		       COALESCE(us.is_online, false) as is_online,
		       COALESCE(dm.content, '') as last_message,
		       (SELECT COUNT(*) FROM direct_messages
		        WHERE conversation_id = c.id
		          AND sender_id != $1
		          AND is_read = false) as unread_count
		FROM conversations c
		JOIN users u ON (CASE
			WHEN c.user1_id = $1 THEN c.user2_id
			ELSE c.user1_id
		END) = u.id
		LEFT JOIN user_status us ON u.id = us.user_id
		LEFT JOIN LATERAL (
			SELECT content FROM direct_messages
			WHERE conversation_id = c.id
			ORDER BY created_at DESC LIMIT 1
		) dm ON true
		WHERE c.user1_id = $1 OR c.user2_id = $1
		ORDER BY c.last_message_at DESC
	`, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get conversations"})
		return
	}
	defer rows.Close()

	conversations := []map[string]interface{}{}
	for rows.Next() {
		var (
			id, lastMessageAt, userIDStr, username, displayName, avatarURL, lastMessage string
			isOnline                                                                    bool
			unreadCount                                                                 int
		)
		rows.Scan(&id, &lastMessageAt, &userIDStr, &username, &displayName,
			&avatarURL, &isOnline, &lastMessage, &unreadCount)

		conversations = append(conversations, map[string]interface{}{
			"id":              id,
			"last_message_at": lastMessageAt,
			"other_user": map[string]interface{}{
				"id":           userIDStr,
				"username":     username,
				"display_name": displayName,
				"avatar_url":   avatarURL,
				"is_online":    isOnline,
			},
			"last_message": lastMessage,
			"unread_count": unreadCount,
		})
	}

	c.JSON(http.StatusOK, conversations)
}

func (h *DMHandler) GetMessages(c *gin.Context) {
	userID := c.GetString("user_id")
	conversationID := c.Param("id")
	limit := c.DefaultQuery("limit", "50")

	// Verify user is part of conversation
	var exists bool
	h.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM conversations
			WHERE id = $1 AND (user1_id = $2 OR user2_id = $2)
		)
	`, conversationID, userID).Scan(&exists)

	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	rows, err := h.db.Query(`
		SELECT dm.id, dm.sender_id, dm.content, dm.is_read, dm.is_edited,
		       dm.created_at, dm.edited_at,
		       u.username, u.display_name, u.avatar_url
		FROM direct_messages dm
		JOIN users u ON dm.sender_id = u.id
		WHERE dm.conversation_id = $1
		ORDER BY dm.created_at DESC
		LIMIT $2
	`, conversationID, limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get messages"})
		return
	}
	defer rows.Close()

	messages := []map[string]interface{}{}
	for rows.Next() {
		var (
			id, senderID, content, createdAt, editedAt, username, displayName, avatarURL string
			isRead, isEdited                                                             bool
		)
		rows.Scan(&id, &senderID, &content, &isRead, &isEdited, &createdAt,
			&editedAt, &username, &displayName, &avatarURL)

		messages = append(messages, map[string]interface{}{
			"id":         id,
			"sender_id":  senderID,
			"content":    content,
			"is_read":    isRead,
			"is_edited":  isEdited,
			"created_at": createdAt,
			"edited_at":  editedAt,
			"sender": map[string]interface{}{
				"id":           senderID,
				"username":     username,
				"display_name": displayName,
				"avatar_url":   avatarURL,
			},
		})
	}

	// Mark as read
	h.db.Exec(`
		UPDATE direct_messages
		SET is_read = true
		WHERE conversation_id = $1 AND sender_id != $2
	`, conversationID, userID)

	c.JSON(http.StatusOK, messages)
}

func (h *DMHandler) MarkAsRead(c *gin.Context) {
	userID := c.GetString("user_id")
	conversationID := c.Param("id")

	_, err := h.db.Exec(`
		UPDATE direct_messages
		SET is_read = true
		WHERE conversation_id = $1 AND sender_id != $2
	`, conversationID, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
