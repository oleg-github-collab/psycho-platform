package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Enhanced group functionality

func (h *GroupHandler) PinTopic(c *gin.Context) {
	userID := c.GetString("user_id")
	topicID := c.Param("id")

	// Check if user is admin
	role := c.GetString("user_role")
	if role != "super_admin" && role != "premium" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only elevated users can pin topics"})
		return
	}

	_, err := h.db.Exec(`
		UPDATE topics
		SET is_pinned = true, pinned_at = CURRENT_TIMESTAMP, pinned_by = $1
		WHERE id = $2
	`, userID, topicID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to pin topic"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *GroupHandler) UnpinTopic(c *gin.Context) {
	topicID := c.Param("id")

	role := c.GetString("user_role")
	if role != "super_admin" && role != "premium" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only elevated users can unpin topics"})
		return
	}

	_, err := h.db.Exec(`
		UPDATE topics
		SET is_pinned = false, pinned_at = NULL, pinned_by = NULL
		WHERE id = $1
	`, topicID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unpin topic"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *GroupHandler) CreateInvitation(c *gin.Context) {
	userID := c.GetString("user_id")
	groupID := c.Param("id")

	var req struct {
		ExpiresIn int `json:"expires_in"` // hours
		MaxUses   int `json:"max_uses"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		req.ExpiresIn = 24 * 7 // 7 days default
		req.MaxUses = 0        // unlimited
	}

	// Check if user is admin/moderator of the group
	var role string
	err := h.db.QueryRow(`
		SELECT role FROM group_members
		WHERE group_id = $1 AND user_id = $2
	`, groupID, userID).Scan(&role)

	if err != nil || (role != "admin" && role != "moderator") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins/moderators can create invitations"})
		return
	}

	// Generate invitation code
	inviteCode := generateInviteCode()

	var expiresAt *time.Time
	if req.ExpiresIn > 0 {
		exp := time.Now().Add(time.Duration(req.ExpiresIn) * time.Hour)
		expiresAt = &exp
	}

	var inviteID string
	err = h.db.QueryRow(`
		INSERT INTO group_invitations (group_id, invitation_code, created_by, expires_at, max_uses)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, groupID, inviteCode, userID, expiresAt, req.MaxUses).Scan(&inviteID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create invitation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":              inviteID,
		"invitation_code": inviteCode,
		"expires_at":      expiresAt,
		"max_uses":        req.MaxUses,
	})
}

func (h *GroupHandler) JoinByInvitation(c *gin.Context) {
	userID := c.GetString("user_id")
	inviteCode := c.Param("code")

	// Validate invitation
	var groupID, createdBy string
	var expiresAt sql.NullTime
	var maxUses, usesCount int
	var isActive bool

	err := h.db.QueryRow(`
		SELECT group_id, created_by, expires_at, max_uses, uses_count, is_active
		FROM group_invitations
		WHERE invitation_code = $1
	`, inviteCode).Scan(&groupID, &createdBy, &expiresAt, &maxUses, &usesCount, &isActive)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid invitation code"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if !isActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invitation is no longer active"})
		return
	}

	if expiresAt.Valid && expiresAt.Time.Before(time.Now()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invitation has expired"})
		return
	}

	if maxUses > 0 && usesCount >= maxUses {
		c.JSON(http.StatusForbidden, gin.H{"error": "Invitation has reached maximum uses"})
		return
	}

	// Add user to group
	_, err = h.db.Exec(`
		INSERT INTO group_members (group_id, user_id, role)
		VALUES ($1, $2, 'member')
		ON CONFLICT (group_id, user_id) DO NOTHING
	`, groupID, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join group"})
		return
	}

	// Update invitation usage
	h.db.Exec(`
		UPDATE group_invitations
		SET uses_count = uses_count + 1
		WHERE invitation_code = $1
	`, inviteCode)

	// Update group member count
	h.db.Exec(`
		UPDATE groups
		SET members_count = (SELECT COUNT(*) FROM group_members WHERE group_id = $1)
		WHERE id = $1
	`, groupID)

	c.JSON(http.StatusOK, gin.H{"success": true, "group_id": groupID})
}

func (h *GroupHandler) UpdateMemberRole(c *gin.Context) {
	userID := c.GetString("user_id")
	groupID := c.Param("id")
	memberID := c.Param("member_id")

	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate role
	if req.Role != "admin" && req.Role != "moderator" && req.Role != "member" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}

	// Check if user is admin of the group
	var role string
	err := h.db.QueryRow(`
		SELECT role FROM group_members
		WHERE group_id = $1 AND user_id = $2
	`, groupID, userID).Scan(&role)

	if err != nil || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can change member roles"})
		return
	}

	// Update member role
	_, err = h.db.Exec(`
		UPDATE group_members
		SET role = $1
		WHERE group_id = $2 AND user_id = $3
	`, req.Role, groupID, memberID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *GroupHandler) RemoveMember(c *gin.Context) {
	userID := c.GetString("user_id")
	groupID := c.Param("id")
	memberID := c.Param("member_id")

	// Check if user is admin or moderator
	var role string
	err := h.db.QueryRow(`
		SELECT role FROM group_members
		WHERE group_id = $1 AND user_id = $2
	`, groupID, userID).Scan(&role)

	if err != nil || (role != "admin" && role != "moderator") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins/moderators can remove members"})
		return
	}

	// Cannot remove admin
	var targetRole string
	h.db.QueryRow(`
		SELECT role FROM group_members
		WHERE group_id = $1 AND user_id = $2
	`, groupID, memberID).Scan(&targetRole)

	if targetRole == "admin" && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot remove admin"})
		return
	}

	// Remove member
	_, err = h.db.Exec(`
		DELETE FROM group_members
		WHERE group_id = $1 AND user_id = $2
	`, groupID, memberID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member"})
		return
	}

	// Update member count
	h.db.Exec(`
		UPDATE groups
		SET members_count = (SELECT COUNT(*) FROM group_members WHERE group_id = $1)
		WHERE id = $1
	`, groupID)

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func generateInviteCode() string {
	b := make([]byte, 12)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:16]
}
