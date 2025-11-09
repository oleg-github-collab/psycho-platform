package models

import "time"

type Group struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	AvatarURL    string    `json:"avatar_url"`
	IsPrivate    bool      `json:"is_private"`
	CreatedBy    string    `json:"created_by"`
	MembersCount int       `json:"members_count"`
	IsMember     bool      `json:"is_member,omitempty"`
	Role         string    `json:"role,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateGroupRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=100"`
	Description string `json:"description"`
	IsPrivate   bool   `json:"is_private"`
}

type GroupMember struct {
	ID       string    `json:"id"`
	GroupID  string    `json:"group_id"`
	UserID   string    `json:"user_id"`
	User     *User     `json:"user,omitempty"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}
