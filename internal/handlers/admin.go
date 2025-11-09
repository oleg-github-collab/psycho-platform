package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	db *sql.DB
}

func NewAdminHandler(db *sql.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

func (h *AdminHandler) GetStats(c *gin.Context) {
	var stats struct {
		TotalUsers       int `json:"total_users"`
		TotalTopics      int `json:"total_topics"`
		TotalGroups      int `json:"total_groups"`
		TotalMessages    int `json:"total_messages"`
		TotalSessions    int `json:"total_sessions"`
		TotalPsychologists int `json:"total_psychologists"`
	}

	h.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)
	h.db.QueryRow("SELECT COUNT(*) FROM topics").Scan(&stats.TotalTopics)
	h.db.QueryRow("SELECT COUNT(*) FROM groups").Scan(&stats.TotalGroups)
	h.db.QueryRow("SELECT COUNT(*) FROM messages").Scan(&stats.TotalMessages)
	h.db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&stats.TotalSessions)
	h.db.QueryRow("SELECT COUNT(*) FROM users WHERE is_psychologist = true").Scan(&stats.TotalPsychologists)

	c.JSON(http.StatusOK, stats)
}

func (h *AdminHandler) ToggleUserStatus(c *gin.Context) {
	userID := c.Param("id")
	action := c.Query("action")

	if action != "activate" && action != "deactivate" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action"})
		return
	}

	isActive := action == "activate"
	_, err := h.db.Exec("UPDATE users SET is_active = $1 WHERE id = $2", isActive, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *AdminHandler) SetPsychologist(c *gin.Context) {
	userID := c.Param("id")
	isPsychologist := c.Query("value") == "true"

	_, err := h.db.Exec("UPDATE users SET is_psychologist = $1, role = CASE WHEN $1 = true THEN 'psychologist' ELSE 'user' END WHERE id = $2", isPsychologist, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update psychologist status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *AdminHandler) GetUsers(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT id, username, display_name, avatar_url, role, is_psychologist, is_active, created_at
		FROM users
		ORDER BY created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	defer rows.Close()

	users := []map[string]interface{}{}
	for rows.Next() {
		var id, username, displayName, avatarURL, role string
		var isPsychologist, isActive bool
		var createdAt string
		rows.Scan(&id, &username, &displayName, &avatarURL, &role, &isPsychologist, &isActive, &createdAt)
		users = append(users, map[string]interface{}{
			"id":              id,
			"username":        username,
			"display_name":    displayName,
			"avatar_url":      avatarURL,
			"role":            role,
			"is_psychologist": isPsychologist,
			"is_active":       isActive,
			"created_at":      createdAt,
		})
	}

	c.JSON(http.StatusOK, users)
}
