package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BookmarkHandler struct {
	db *sql.DB
}

func NewBookmarkHandler(db *sql.DB) *BookmarkHandler {
	return &BookmarkHandler{db: db}
}

func (h *BookmarkHandler) AddBookmark(c *gin.Context) {
	userID := c.GetString("user_id")
	messageID := c.Param("message_id")

	_, err := h.db.Exec(`
		INSERT INTO message_bookmarks (user_id, message_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, message_id) DO NOTHING
	`, userID, messageID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add bookmark"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *BookmarkHandler) RemoveBookmark(c *gin.Context) {
	userID := c.GetString("user_id")
	messageID := c.Param("message_id")

	_, err := h.db.Exec(`
		DELETE FROM message_bookmarks
		WHERE user_id = $1 AND message_id = $2
	`, userID, messageID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove bookmark"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *BookmarkHandler) GetBookmarks(c *gin.Context) {
	userID := c.GetString("user_id")
	limit := c.DefaultQuery("limit", "50")

	rows, err := h.db.Query(`
		SELECT m.id, m.content, m.created_at, m.topic_id, m.group_id,
		       u.username, u.display_name, u.avatar_url,
		       mb.created_at as bookmarked_at
		FROM message_bookmarks mb
		JOIN messages m ON mb.message_id = m.id
		JOIN users u ON m.user_id = u.id
		WHERE mb.user_id = $1
		ORDER BY mb.created_at DESC
		LIMIT $2
	`, userID, limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bookmarks"})
		return
	}
	defer rows.Close()

	bookmarks := []map[string]interface{}{}
	for rows.Next() {
		var (
			id, content, createdAt, topicID, groupID           sql.NullString
			username, displayName, avatarURL, bookmarkedAt string
		)
		rows.Scan(&id, &content, &createdAt, &topicID, &groupID,
			&username, &displayName, &avatarURL, &bookmarkedAt)

		bookmarks = append(bookmarks, map[string]interface{}{
			"id":            id.String,
			"content":       content.String,
			"created_at":    createdAt.String,
			"topic_id":      topicID.String,
			"group_id":      groupID.String,
			"bookmarked_at": bookmarkedAt,
			"user": map[string]string{
				"username":     username,
				"display_name": displayName,
				"avatar_url":   avatarURL,
			},
		})
	}

	c.JSON(http.StatusOK, bookmarks)
}

func (h *BookmarkHandler) IsBookmarked(c *gin.Context) {
	userID := c.GetString("user_id")
	messageID := c.Param("message_id")

	var exists bool
	err := h.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM message_bookmarks
			WHERE user_id = $1 AND message_id = $2
		)
	`, userID, messageID).Scan(&exists)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check bookmark"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"is_bookmarked": exists})
}
