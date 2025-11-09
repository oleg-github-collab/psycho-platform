package handlers

import (
	"database/sql"
	"net/http"
	"psycho-platform/internal/models"

	"github.com/gin-gonic/gin"
)

type GroupHandler struct {
	db *sql.DB
}

func NewGroupHandler(db *sql.DB) *GroupHandler {
	return &GroupHandler{db: db}
}

func (h *GroupHandler) CreateGroup(c *gin.Context) {
	userID := c.GetString("user_id")
	var req models.CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer tx.Rollback()

	var group models.Group
	err = tx.QueryRow(`
		INSERT INTO groups (name, description, is_private, created_by, members_count)
		VALUES ($1, $2, $3, $4, 1)
		RETURNING id, name, description, avatar_url, is_private, created_by, members_count, created_at, updated_at
	`, req.Name, req.Description, req.IsPrivate, userID).Scan(
		&group.ID, &group.Name, &group.Description, &group.AvatarURL,
		&group.IsPrivate, &group.CreatedBy, &group.MembersCount,
		&group.CreatedAt, &group.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create group"})
		return
	}

	// Add creator as admin
	_, err = tx.Exec("INSERT INTO group_members (group_id, user_id, role) VALUES ($1, $2, 'admin')", group.ID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add creator to group"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusCreated, group)
}

func (h *GroupHandler) GetGroups(c *gin.Context) {
	userID := c.GetString("user_id")

	query := `
		SELECT g.id, g.name, g.description, g.avatar_url, g.is_private, g.created_by,
		       g.members_count, g.created_at, g.updated_at,
		       COALESCE(gm.role, '') as user_role,
		       CASE WHEN gm.user_id IS NOT NULL THEN true ELSE false END as is_member
		FROM groups g
		LEFT JOIN group_members gm ON g.id = gm.group_id AND gm.user_id = $1
		WHERE g.is_private = false OR gm.user_id IS NOT NULL
		ORDER BY g.created_at DESC
	`

	rows, err := h.db.Query(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch groups"})
		return
	}
	defer rows.Close()

	groups := []models.Group{}
	for rows.Next() {
		var group models.Group
		err := rows.Scan(
			&group.ID, &group.Name, &group.Description, &group.AvatarURL,
			&group.IsPrivate, &group.CreatedBy, &group.MembersCount,
			&group.CreatedAt, &group.UpdatedAt, &group.Role, &group.IsMember,
		)
		if err != nil {
			continue
		}
		groups = append(groups, group)
	}

	c.JSON(http.StatusOK, groups)
}

func (h *GroupHandler) JoinGroup(c *gin.Context) {
	userID := c.GetString("user_id")
	groupID := c.Param("id")

	// Check if group is private
	var isPrivate bool
	err := h.db.QueryRow("SELECT is_private FROM groups WHERE id = $1", groupID).Scan(&isPrivate)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
		return
	}

	if isPrivate {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot join private group"})
		return
	}

	_, err = h.db.Exec("INSERT INTO group_members (group_id, user_id, role) VALUES ($1, $2, 'member') ON CONFLICT DO NOTHING", groupID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join group"})
		return
	}

	h.db.Exec("UPDATE groups SET members_count = (SELECT COUNT(*) FROM group_members WHERE group_id = $1) WHERE id = $1", groupID)

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *GroupHandler) LeaveGroup(c *gin.Context) {
	userID := c.GetString("user_id")
	groupID := c.Param("id")

	_, err := h.db.Exec("DELETE FROM group_members WHERE group_id = $1 AND user_id = $2", groupID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to leave group"})
		return
	}

	h.db.Exec("UPDATE groups SET members_count = (SELECT COUNT(*) FROM group_members WHERE group_id = $1) WHERE id = $1", groupID)

	c.JSON(http.StatusOK, gin.H{"success": true})
}
