package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ActivityHandler struct {
	db *sql.DB
}

func NewActivityHandler(db *sql.DB) *ActivityHandler {
	return &ActivityHandler{db: db}
}

func (h *ActivityHandler) GetActivityFeed(c *gin.Context) {
	userID := c.GetString("user_id")
	limit := c.DefaultQuery("limit", "50")

	// Get user's activity and followed users' activity
	rows, err := h.db.Query(`
		SELECT a.id, a.user_id, a.activity_type, a.entity_type, a.entity_id,
		       a.content, a.metadata, a.created_at,
		       u.username, u.display_name, u.avatar_url
		FROM activity_feed a
		JOIN users u ON a.user_id = u.id
		WHERE a.user_id = $1
		   OR a.user_id IN (
		       -- Users in same groups
		       SELECT DISTINCT gm2.user_id
		       FROM group_members gm1
		       JOIN group_members gm2 ON gm1.group_id = gm2.group_id
		       WHERE gm1.user_id = $1 AND gm2.user_id != $1
		   )
		ORDER BY a.created_at DESC
		LIMIT $2
	`, userID, limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch activity"})
		return
	}
	defer rows.Close()

	activities := []map[string]interface{}{}
	for rows.Next() {
		var (
			id, userIDStr, activityType, entityType, entityID, content, createdAt string
			username, displayName, avatarURL                                      string
			metadata                                                              []byte
		)
		rows.Scan(&id, &userIDStr, &activityType, &entityType, &entityID,
			&content, &metadata, &createdAt, &username, &displayName, &avatarURL)

		var metaMap map[string]interface{}
		json.Unmarshal(metadata, &metaMap)

		activities = append(activities, map[string]interface{}{
			"id":            id,
			"activity_type": activityType,
			"entity_type":   entityType,
			"entity_id":     entityID,
			"content":       content,
			"metadata":      metaMap,
			"created_at":    createdAt,
			"user": map[string]string{
				"id":           userIDStr,
				"username":     username,
				"display_name": displayName,
				"avatar_url":   avatarURL,
			},
		})
	}

	c.JSON(http.StatusOK, activities)
}

func (h *ActivityHandler) GetTrendingTopics(c *gin.Context) {
	limit := c.DefaultQuery("limit", "10")
	timeframe := c.DefaultQuery("timeframe", "24") // hours

	rows, err := h.db.Query(`
		SELECT t.id, t.title, t.description, t.votes_count, t.messages_count,
		       COUNT(DISTINCT m.id) as recent_messages,
		       u.username, u.display_name
		FROM topics t
		JOIN users u ON t.created_by = u.id
		LEFT JOIN messages m ON t.id = m.topic_id
		    AND m.created_at > NOW() - INTERVAL '1 hour' * $1
		WHERE t.is_public = true
		GROUP BY t.id, t.title, t.description, t.votes_count, t.messages_count,
		         u.username, u.display_name
		ORDER BY (t.votes_count + COUNT(DISTINCT m.id) * 2) DESC
		LIMIT $2
	`, timeframe, limit)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch trending topics"})
		return
	}
	defer rows.Close()

	topics := []map[string]interface{}{}
	for rows.Next() {
		var (
			id, title, description, username, displayName                      string
			votesCount, messagesCount, recentMessages int
		)
		rows.Scan(&id, &title, &description, &votesCount, &messagesCount,
			&recentMessages, &username, &displayName)

		topics = append(topics, map[string]interface{}{
			"id":              id,
			"title":           title,
			"description":     description,
			"votes_count":     votesCount,
			"messages_count":  messagesCount,
			"recent_messages": recentMessages,
			"trending_score":  votesCount + recentMessages*2,
			"author": map[string]string{
				"username":     username,
				"display_name": displayName,
			},
		})
	}

	c.JSON(http.StatusOK, topics)
}

// Helper to create activity
func (h *ActivityHandler) CreateActivity(userID, activityType, entityType, entityID, content string, metadata map[string]interface{}) error {
	metaJSON, _ := json.Marshal(metadata)

	_, err := h.db.Exec(`
		INSERT INTO activity_feed (user_id, activity_type, entity_type, entity_id, content, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, activityType, entityType, entityID, content, metaJSON)

	return err
}
