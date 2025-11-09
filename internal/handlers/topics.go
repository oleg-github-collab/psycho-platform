package handlers

import (
	"database/sql"
	"net/http"
	"psycho-platform/internal/models"

	"github.com/gin-gonic/gin"
)

type TopicHandler struct {
	db *sql.DB
}

func NewTopicHandler(db *sql.DB) *TopicHandler {
	return &TopicHandler{db: db}
}

func (h *TopicHandler) CreateTopic(c *gin.Context) {
	userID := c.GetString("user_id")
	var req models.CreateTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var topic models.Topic
	err := h.db.QueryRow(`
		INSERT INTO topics (title, description, is_public, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, title, description, is_public, created_by, votes_count, messages_count, created_at, updated_at
	`, req.Title, req.Description, req.IsPublic, userID).Scan(
		&topic.ID, &topic.Title, &topic.Description, &topic.IsPublic,
		&topic.CreatedBy, &topic.VotesCount, &topic.MessagesCount,
		&topic.CreatedAt, &topic.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create topic"})
		return
	}

	c.JSON(http.StatusCreated, topic)
}

func (h *TopicHandler) GetTopics(c *gin.Context) {
	userID := c.GetString("user_id")
	onlyPublic := c.Query("public") == "true"

	query := `
		SELECT t.id, t.title, t.description, t.is_public, t.created_by, t.votes_count, t.messages_count,
		       t.created_at, t.updated_at, u.username, u.display_name, u.avatar_url,
		       COALESCE(tv.vote_type, '') as user_vote
		FROM topics t
		JOIN users u ON t.created_by = u.id
		LEFT JOIN topic_votes tv ON t.id = tv.topic_id AND tv.user_id = $1
		WHERE ($2 = false OR t.is_public = true)
		ORDER BY t.votes_count DESC, t.created_at DESC
	`

	rows, err := h.db.Query(query, userID, onlyPublic)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch topics"})
		return
	}
	defer rows.Close()

	topics := []models.Topic{}
	for rows.Next() {
		var topic models.Topic
		var user models.User
		err := rows.Scan(
			&topic.ID, &topic.Title, &topic.Description, &topic.IsPublic,
			&topic.CreatedBy, &topic.VotesCount, &topic.MessagesCount,
			&topic.CreatedAt, &topic.UpdatedAt,
			&user.Username, &user.DisplayName, &user.AvatarURL,
			&topic.UserVote,
		)
		if err != nil {
			continue
		}
		user.ID = topic.CreatedBy
		topic.CreatedByUser = &user
		topics = append(topics, topic)
	}

	c.JSON(http.StatusOK, topics)
}

func (h *TopicHandler) VoteTopic(c *gin.Context) {
	userID := c.GetString("user_id")
	topicID := c.Param("id")
	voteType := c.Query("type")

	if voteType != "up" && voteType != "down" {
		voteType = "up"
	}

	// Check if already voted
	var existingVote string
	err := h.db.QueryRow("SELECT vote_type FROM topic_votes WHERE topic_id = $1 AND user_id = $2", topicID, userID).Scan(&existingVote)

	if err == sql.ErrNoRows {
		// New vote
		_, err = h.db.Exec("INSERT INTO topic_votes (topic_id, user_id, vote_type) VALUES ($1, $2, $3)", topicID, userID, voteType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to vote"})
			return
		}
	} else if existingVote == voteType {
		// Remove vote
		_, err = h.db.Exec("DELETE FROM topic_votes WHERE topic_id = $1 AND user_id = $2", topicID, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove vote"})
			return
		}
	} else {
		// Update vote
		_, err = h.db.Exec("UPDATE topic_votes SET vote_type = $1 WHERE topic_id = $2 AND user_id = $3", voteType, topicID, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vote"})
			return
		}
	}

	// Update vote count
	var votesCount int
	h.db.QueryRow(`
		UPDATE topics SET votes_count = (
			SELECT COUNT(*) FROM topic_votes WHERE topic_id = $1 AND vote_type = 'up'
		) - (
			SELECT COUNT(*) FROM topic_votes WHERE topic_id = $1 AND vote_type = 'down'
		)
		WHERE id = $1
		RETURNING votes_count
	`, topicID).Scan(&votesCount)

	c.JSON(http.StatusOK, gin.H{"votes_count": votesCount})
}
