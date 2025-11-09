package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	db *sql.DB
}

func NewProfileHandler(db *sql.DB) *ProfileHandler {
	return &ProfileHandler{db: db}
}

type UpdateProfileRequest struct {
	DisplayName string `json:"display_name"`
	Bio         string `json:"bio"`
	AvatarURL   string `json:"avatar_url"`
	Status      string `json:"status"`
}

func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.db.Exec(`
		UPDATE users
		SET display_name = COALESCE(NULLIF($1, ''), display_name),
		    bio = $2,
		    avatar_url = COALESCE(NULLIF($3, ''), avatar_url),
		    status = COALESCE(NULLIF($4, ''), status),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
	`, req.DisplayName, req.Bio, req.AvatarURL, req.Status, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *ProfileHandler) GetUserProfile(c *gin.Context) {
	userID := c.Param("id")

	var user struct {
		ID             string `json:"id"`
		Username       string `json:"username"`
		DisplayName    string `json:"display_name"`
		AvatarURL      string `json:"avatar_url"`
		Bio            string `json:"bio"`
		Status         string `json:"status"`
		Role           string `json:"role"`
		IsPsychologist bool   `json:"is_psychologist"`
		IsOnline       bool   `json:"is_online"`
		LastSeen       string `json:"last_seen"`
	}

	err := h.db.QueryRow(`
		SELECT u.id, u.username, u.display_name, u.avatar_url, u.bio,
		       COALESCE(u.status, '') as status, u.role, u.is_psychologist,
		       COALESCE(us.is_online, false) as is_online,
		       COALESCE(us.last_seen::text, '') as last_seen
		FROM users u
		LEFT JOIN user_status us ON u.id = us.user_id
		WHERE u.id = $1 AND u.is_active = true
	`, userID).Scan(
		&user.ID, &user.Username, &user.DisplayName, &user.AvatarURL,
		&user.Bio, &user.Status, &user.Role, &user.IsPsychologist,
		&user.IsOnline, &user.LastSeen,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *ProfileHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	psychologistsOnly := c.Query("psychologists") == "true"
	limit := c.DefaultQuery("limit", "20")

	sqlQuery := `
		SELECT u.id, u.username, u.display_name, u.avatar_url, u.bio,
		       u.role, u.is_psychologist,
		       COALESCE(us.is_online, false) as is_online
		FROM users u
		LEFT JOIN user_status us ON u.id = us.user_id
		WHERE u.is_active = true
		  AND ($1 = '' OR u.username ILIKE '%' || $1 || '%' OR u.display_name ILIKE '%' || $1 || '%')
		  AND ($2 = false OR u.is_psychologist = true)
		ORDER BY us.is_online DESC, u.display_name ASC
		LIMIT $3
	`

	rows, err := h.db.Query(sqlQuery, query, psychologistsOnly, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search users"})
		return
	}
	defer rows.Close()

	users := []map[string]interface{}{}
	for rows.Next() {
		var user struct {
			ID             string
			Username       string
			DisplayName    string
			AvatarURL      string
			Bio            string
			Role           string
			IsPsychologist bool
			IsOnline       bool
		}
		rows.Scan(&user.ID, &user.Username, &user.DisplayName, &user.AvatarURL,
			&user.Bio, &user.Role, &user.IsPsychologist, &user.IsOnline)
		users = append(users, map[string]interface{}{
			"id":              user.ID,
			"username":        user.Username,
			"display_name":    user.DisplayName,
			"avatar_url":      user.AvatarURL,
			"bio":             user.Bio,
			"role":            user.Role,
			"is_psychologist": user.IsPsychologist,
			"is_online":       user.IsOnline,
		})
	}

	c.JSON(http.StatusOK, users)
}

func (h *ProfileHandler) BlockUser(c *gin.Context) {
	userID := c.GetString("user_id")
	blockedUserID := c.Param("id")

	if userID == blockedUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot block yourself"})
		return
	}

	_, err := h.db.Exec(`
		INSERT INTO user_blocks (user_id, blocked_user_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, blocked_user_id) DO NOTHING
	`, userID, blockedUserID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to block user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *ProfileHandler) UnblockUser(c *gin.Context) {
	userID := c.GetString("user_id")
	blockedUserID := c.Param("id")

	_, err := h.db.Exec(`
		DELETE FROM user_blocks
		WHERE user_id = $1 AND blocked_user_id = $2
	`, userID, blockedUserID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unblock user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *ProfileHandler) GetBlockedUsers(c *gin.Context) {
	userID := c.GetString("user_id")

	rows, err := h.db.Query(`
		SELECT u.id, u.username, u.display_name, u.avatar_url
		FROM user_blocks ub
		JOIN users u ON ub.blocked_user_id = u.id
		WHERE ub.user_id = $1
		ORDER BY ub.created_at DESC
	`, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get blocked users"})
		return
	}
	defer rows.Close()

	users := []map[string]interface{}{}
	for rows.Next() {
		var id, username, displayName, avatarURL string
		rows.Scan(&id, &username, &displayName, &avatarURL)
		users = append(users, map[string]interface{}{
			"id":           id,
			"username":     username,
			"display_name": displayName,
			"avatar_url":   avatarURL,
		})
	}

	c.JSON(http.StatusOK, users)
}

func (h *ProfileHandler) SetOnlineStatus(c *gin.Context) {
	userID := c.GetString("user_id")
	isOnline := c.Query("online") == "true"

	if isOnline {
		_, err := h.db.Exec(`
			INSERT INTO user_status (user_id, is_online, last_seen)
			VALUES ($1, true, CURRENT_TIMESTAMP)
			ON CONFLICT (user_id)
			DO UPDATE SET is_online = true, last_seen = CURRENT_TIMESTAMP
		`, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
			return
		}
	} else {
		_, err := h.db.Exec(`
			UPDATE user_status
			SET is_online = false, last_seen = CURRENT_TIMESTAMP
			WHERE user_id = $1
		`, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
