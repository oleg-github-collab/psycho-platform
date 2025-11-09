package handlers

import (
	"database/sql"
	"net/http"
	"psycho-platform/internal/models"
	"psycho-platform/internal/websocket"

	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	db  *sql.DB
	hub *websocket.Hub
}

func NewMessageHandler(db *sql.DB, hub *websocket.Hub) *MessageHandler {
	return &MessageHandler{db: db, hub: hub}
}

func (h *MessageHandler) CreateMessage(c *gin.Context) {
	userID := c.GetString("user_id")
	var req models.CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var message models.Message
	err := h.db.QueryRow(`
		INSERT INTO messages (content, topic_id, group_id, user_id, parent_id, quoted_message_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, content, topic_id, group_id, user_id, parent_id, quoted_message_id, is_edited, created_at
	`, req.Content, req.TopicID, req.GroupID, userID, req.ParentID, req.QuotedMessageID).Scan(
		&message.ID, &message.Content, &message.TopicID, &message.GroupID,
		&message.UserID, &message.ParentID, &message.QuotedMessageID,
		&message.IsEdited, &message.CreatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
		return
	}

	// Get user info
	var user models.User
	h.db.QueryRow("SELECT id, username, display_name, avatar_url FROM users WHERE id = $1", userID).Scan(
		&user.ID, &user.Username, &user.DisplayName, &user.AvatarURL,
	)
	message.User = &user

	// Update message count
	if req.TopicID != nil {
		h.db.Exec("UPDATE topics SET messages_count = messages_count + 1 WHERE id = $1", *req.TopicID)
	}

	// Broadcast via WebSocket
	roomID := ""
	if req.TopicID != nil {
		roomID = "topic_" + *req.TopicID
	} else if req.GroupID != nil {
		roomID = "group_" + *req.GroupID
	}

	if roomID != "" {
		h.hub.BroadcastToRoom(roomID, map[string]interface{}{
			"type":    "new_message",
			"payload": message,
		})
	}

	c.JSON(http.StatusCreated, message)
}

func (h *MessageHandler) GetMessages(c *gin.Context) {
	topicID := c.Query("topic_id")
	groupID := c.Query("group_id")
	limit := c.DefaultQuery("limit", "50")

	query := `
		SELECT m.id, m.content, m.topic_id, m.group_id, m.user_id, m.parent_id,
		       m.quoted_message_id, m.is_edited, m.edited_at, m.created_at,
		       u.username, u.display_name, u.avatar_url
		FROM messages m
		JOIN users u ON m.user_id = u.id
		WHERE ($1 = '' OR m.topic_id = $1::uuid)
		  AND ($2 = '' OR m.group_id = $2::uuid)
		ORDER BY m.created_at DESC
		LIMIT $3
	`

	rows, err := h.db.Query(query, topicID, groupID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}
	defer rows.Close()

	messages := []models.Message{}
	for rows.Next() {
		var msg models.Message
		var user models.User
		err := rows.Scan(
			&msg.ID, &msg.Content, &msg.TopicID, &msg.GroupID, &msg.UserID,
			&msg.ParentID, &msg.QuotedMessageID, &msg.IsEdited, &msg.EditedAt,
			&msg.CreatedAt, &user.Username, &user.DisplayName, &user.AvatarURL,
		)
		if err != nil {
			continue
		}
		user.ID = msg.UserID
		msg.User = &user

		// Get reactions
		reactRows, _ := h.db.Query(`
			SELECT r.id, r.emoji, r.user_id, u.username, u.display_name
			FROM reactions r
			JOIN users u ON r.user_id = u.id
			WHERE r.message_id = $1
		`, msg.ID)

		reactions := []models.Reaction{}
		for reactRows.Next() {
			var reaction models.Reaction
			var reactUser models.User
			reactRows.Scan(&reaction.ID, &reaction.Emoji, &reaction.UserID, &reactUser.Username, &reactUser.DisplayName)
			reactUser.ID = reaction.UserID
			reaction.User = &reactUser
			reactions = append(reactions, reaction)
		}
		reactRows.Close()
		msg.Reactions = reactions

		messages = append(messages, msg)
	}

	c.JSON(http.StatusOK, messages)
}

func (h *MessageHandler) AddReaction(c *gin.Context) {
	userID := c.GetString("user_id")
	messageID := c.Param("id")
	var req models.AddReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var reactionID string
	err := h.db.QueryRow(`
		INSERT INTO reactions (message_id, user_id, emoji)
		VALUES ($1, $2, $3)
		ON CONFLICT (message_id, user_id, emoji) DO NOTHING
		RETURNING id
	`, messageID, userID, req.Emoji).Scan(&reactionID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add reaction"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": reactionID})
}

func (h *MessageHandler) RemoveReaction(c *gin.Context) {
	userID := c.GetString("user_id")
	messageID := c.Param("id")
	emoji := c.Query("emoji")

	_, err := h.db.Exec("DELETE FROM reactions WHERE message_id = $1 AND user_id = $2 AND emoji = $3", messageID, userID, emoji)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove reaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *MessageHandler) EditMessage(c *gin.Context) {
	userID := c.GetString("user_id")
	messageID := c.Param("id")

	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify ownership
	var ownerID string
	err := h.db.QueryRow("SELECT user_id FROM messages WHERE id = $1", messageID).Scan(&ownerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot edit other user's message"})
		return
	}

	_, err = h.db.Exec(`
		UPDATE messages
		SET content = $1, is_edited = true, edited_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, req.Content, messageID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to edit message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *MessageHandler) DeleteMessage(c *gin.Context) {
	userID := c.GetString("user_id")
	messageID := c.Param("id")

	// Verify ownership
	var ownerID string
	err := h.db.QueryRow("SELECT user_id FROM messages WHERE id = $1", messageID).Scan(&ownerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	if ownerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete other user's message"})
		return
	}

	_, err = h.db.Exec(`
		UPDATE messages
		SET is_deleted = true, deleted_at = CURRENT_TIMESTAMP, content = '[Видалено]'
		WHERE id = $1
	`, messageID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *MessageHandler) MarkAsRead(c *gin.Context) {
	userID := c.GetString("user_id")
	messageID := c.Param("id")

	_, err := h.db.Exec(`
		INSERT INTO message_read_receipts (message_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (message_id, user_id) DO NOTHING
	`, messageID, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *MessageHandler) StartTyping(c *gin.Context) {
	userID := c.GetString("user_id")
	roomID := c.Query("room")

	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room parameter required"})
		return
	}

	_, err := h.db.Exec(`
		INSERT INTO typing_indicators (user_id, room_id, started_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, room_id)
		DO UPDATE SET started_at = CURRENT_TIMESTAMP
	`, userID, roomID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update typing status"})
		return
	}

	// Broadcast typing indicator
	h.hub.BroadcastToRoom(roomID, map[string]interface{}{
		"type": "typing",
		"payload": map[string]interface{}{
			"user_id":  userID,
			"is_typing": true,
		},
	})

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *MessageHandler) StopTyping(c *gin.Context) {
	userID := c.GetString("user_id")
	roomID := c.Query("room")

	if roomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room parameter required"})
		return
	}

	h.db.Exec(`
		DELETE FROM typing_indicators
		WHERE user_id = $1 AND room_id = $2
	`, userID, roomID)

	// Broadcast stop typing
	h.hub.BroadcastToRoom(roomID, map[string]interface{}{
		"type": "typing",
		"payload": map[string]interface{}{
			"user_id":  userID,
			"is_typing": false,
		},
	})

	c.JSON(http.StatusOK, gin.H{"success": true})
}
