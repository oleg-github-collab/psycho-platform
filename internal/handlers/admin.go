package handlers

import (
	"database/sql"
	"net/http"
	"strings"

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
		TotalUsers        int `json:"total_users"`
		TotalTopics       int `json:"total_topics"`
		TotalGroups       int `json:"total_groups"`
		TotalMessages     int `json:"total_messages"`
		TotalSessions     int `json:"total_sessions"`
		TotalPremiumUsers int `json:"total_premium_users"`
		TotalBasicUsers   int `json:"total_basic_users"`
		TotalSuperAdmins  int `json:"total_super_admins"`
	}

	h.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)
	h.db.QueryRow("SELECT COUNT(*) FROM topics").Scan(&stats.TotalTopics)
	h.db.QueryRow("SELECT COUNT(*) FROM groups").Scan(&stats.TotalGroups)
	h.db.QueryRow("SELECT COUNT(*) FROM messages").Scan(&stats.TotalMessages)
	h.db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&stats.TotalSessions)
	h.db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'premium'").Scan(&stats.TotalPremiumUsers)
	h.db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'basic'").Scan(&stats.TotalBasicUsers)
	h.db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'super_admin'").Scan(&stats.TotalSuperAdmins)

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

func (h *AdminHandler) GetUsers(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT id, username, display_name, avatar_url, role, is_active, created_at
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
		var isActive bool
		var createdAt string
		rows.Scan(&id, &username, &displayName, &avatarURL, &role, &isActive, &createdAt)
		users = append(users, map[string]interface{}{
			"id":           id,
			"username":     username,
			"display_name": displayName,
			"avatar_url":   avatarURL,
			"role":         role,
			"is_active":    isActive,
			"created_at":   createdAt,
		})
	}

	c.JSON(http.StatusOK, users)
}

func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := strings.ToLower(req.Role)
	if role != "super_admin" && role != "premium" && role != "basic" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	var currentRole string
	err = tx.QueryRow("SELECT role FROM users WHERE id = $1", userID).Scan(&currentRole)
	if err == sql.ErrNoRows {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load user"})
		return
	}

	if role == "super_admin" {
		_, err = tx.Exec(`
			UPDATE users
			SET role = CASE
				WHEN id = $1 THEN 'super_admin'
				WHEN role = 'super_admin' AND id <> $1 THEN 'basic'
				ELSE role
			END
			WHERE id = $1 OR role = 'super_admin'
		`, userID)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
			return
		}
	} else {
		if currentRole == "super_admin" {
			var superAdmins int
			if err := tx.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'super_admin'").Scan(&superAdmins); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify super admins"})
				return
			}
			if superAdmins <= 1 {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot remove the last super admin"})
				return
			}
		}

		if _, err := tx.Exec("UPDATE users SET role = $1 WHERE id = $2", role, userID); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit changes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
