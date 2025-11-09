package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	db *sql.DB
}

func NewSearchHandler(db *sql.DB) *SearchHandler {
	return &SearchHandler{db: db}
}

type SearchResults struct {
	Messages []map[string]interface{} `json:"messages"`
	Topics   []map[string]interface{} `json:"topics"`
	Groups   []map[string]interface{} `json:"groups"`
	Users    []map[string]interface{} `json:"users"`
}

func (h *SearchHandler) GlobalSearch(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter required"})
		return
	}

	userID := c.GetString("user_id")
	limit := c.DefaultQuery("limit", "10")
	results := SearchResults{
		Messages: []map[string]interface{}{},
		Topics:   []map[string]interface{}{},
		Groups:   []map[string]interface{}{},
		Users:    []map[string]interface{}{},
	}

	// Search messages
	messageRows, _ := h.db.Query(`
		SELECT m.id, m.content, m.created_at, u.username, u.display_name
		FROM messages m
		JOIN users u ON m.user_id = u.id
		WHERE m.content ILIKE '%' || $1 || '%'
		  AND m.is_deleted = false
		  AND (m.topic_id IN (SELECT id FROM topics WHERE is_public = true)
		       OR m.group_id IN (SELECT group_id FROM group_members WHERE user_id = $2))
		ORDER BY m.created_at DESC
		LIMIT $3
	`, query, userID, limit)

	if messageRows != nil {
		defer messageRows.Close()
		for messageRows.Next() {
			var id, content, createdAt, username, displayName string
			messageRows.Scan(&id, &content, &createdAt, &username, &displayName)
			results.Messages = append(results.Messages, map[string]interface{}{
				"id":         id,
				"content":    content,
				"created_at": createdAt,
				"user":       map[string]string{"username": username, "display_name": displayName},
			})
		}
	}

	// Search topics
	topicRows, _ := h.db.Query(`
		SELECT id, title, description, votes_count, messages_count
		FROM topics
		WHERE (title ILIKE '%' || $1 || '%' OR description ILIKE '%' || $1 || '%')
		  AND is_public = true
		ORDER BY votes_count DESC
		LIMIT $2
	`, query, limit)

	if topicRows != nil {
		defer topicRows.Close()
		for topicRows.Next() {
			var id, title, description string
			var votesCount, messagesCount int
			topicRows.Scan(&id, &title, &description, &votesCount, &messagesCount)
			results.Topics = append(results.Topics, map[string]interface{}{
				"id":             id,
				"title":          title,
				"description":    description,
				"votes_count":    votesCount,
				"messages_count": messagesCount,
			})
		}
	}

	// Search groups
	groupRows, _ := h.db.Query(`
		SELECT id, name, description, members_count
		FROM groups
		WHERE (name ILIKE '%' || $1 || '%' OR description ILIKE '%' || $1 || '%')
		  AND is_private = false
		ORDER BY members_count DESC
		LIMIT $2
	`, query, limit)

	if groupRows != nil {
		defer groupRows.Close()
		for groupRows.Next() {
			var id, name, description string
			var membersCount int
			groupRows.Scan(&id, &name, &description, &membersCount)
			results.Groups = append(results.Groups, map[string]interface{}{
				"id":            id,
				"name":          name,
				"description":   description,
				"members_count": membersCount,
			})
		}
	}

	// Search users
	userRows, _ := h.db.Query(`
		SELECT id, username, display_name, bio, is_psychologist
		FROM users
		WHERE (username ILIKE '%' || $1 || '%' OR display_name ILIKE '%' || $1 || '%' OR bio ILIKE '%' || $1 || '%')
		  AND is_active = true
		ORDER BY is_psychologist DESC
		LIMIT $2
	`, query, limit)

	if userRows != nil {
		defer userRows.Close()
		for userRows.Next() {
			var id, username, displayName, bio string
			var isPsychologist bool
			userRows.Scan(&id, &username, &displayName, &bio, &isPsychologist)
			results.Users = append(results.Users, map[string]interface{}{
				"id":              id,
				"username":        username,
				"display_name":    displayName,
				"bio":             bio,
				"is_psychologist": isPsychologist,
			})
		}
	}

	c.JSON(http.StatusOK, results)
}

func (h *SearchHandler) SearchMessages(c *gin.Context) {
	query := c.Query("q")
	topicID := c.Query("topic_id")
	groupID := c.Query("group_id")
	limit := c.DefaultQuery("limit", "50")

	sqlQuery := `
		SELECT m.id, m.content, m.created_at, u.username, u.display_name, u.avatar_url
		FROM messages m
		JOIN users u ON m.user_id = u.id
		WHERE m.content ILIKE '%' || $1 || '%'
		  AND m.is_deleted = false
		  AND ($2 = '' OR m.topic_id = $2::uuid)
		  AND ($3 = '' OR m.group_id = $3::uuid)
		ORDER BY m.created_at DESC
		LIMIT $4
	`

	rows, err := h.db.Query(sqlQuery, query, topicID, groupID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}
	defer rows.Close()

	messages := []map[string]interface{}{}
	for rows.Next() {
		var id, content, createdAt, username, displayName, avatarURL string
		rows.Scan(&id, &content, &createdAt, &username, &displayName, &avatarURL)
		messages = append(messages, map[string]interface{}{
			"id":         id,
			"content":    content,
			"created_at": createdAt,
			"user": map[string]string{
				"username":     username,
				"display_name": displayName,
				"avatar_url":   avatarURL,
			},
		})
	}

	c.JSON(http.StatusOK, messages)
}
