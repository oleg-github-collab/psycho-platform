package models

import "time"

type Topic struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	IsPublic      bool      `json:"is_public"`
	CreatedBy     string    `json:"created_by"`
	CreatedByUser *User     `json:"created_by_user,omitempty"`
	VotesCount    int       `json:"votes_count"`
	MessagesCount int       `json:"messages_count"`
	UserVote      string    `json:"user_vote,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateTopicRequest struct {
	Title       string `json:"title" binding:"required,min=3,max=255"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

type TopicVote struct {
	ID        string    `json:"id"`
	TopicID   string    `json:"topic_id"`
	UserID    string    `json:"user_id"`
	VoteType  string    `json:"vote_type"`
	CreatedAt time.Time `json:"created_at"`
}
