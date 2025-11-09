package models

import "time"

type Message struct {
	ID              string     `json:"id"`
	Content         string     `json:"content"`
	TopicID         *string    `json:"topic_id,omitempty"`
	GroupID         *string    `json:"group_id,omitempty"`
	UserID          string     `json:"user_id"`
	User            *User      `json:"user,omitempty"`
	ParentID        *string    `json:"parent_id,omitempty"`
	QuotedMessageID *string    `json:"quoted_message_id,omitempty"`
	QuotedMessage   *Message   `json:"quoted_message,omitempty"`
	IsEdited        bool       `json:"is_edited"`
	EditedAt        *time.Time `json:"edited_at,omitempty"`
	Reactions       []Reaction `json:"reactions,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type CreateMessageRequest struct {
	Content         string  `json:"content" binding:"required"`
	TopicID         *string `json:"topic_id"`
	GroupID         *string `json:"group_id"`
	ParentID        *string `json:"parent_id"`
	QuotedMessageID *string `json:"quoted_message_id"`
}

type Reaction struct {
	ID        string    `json:"id"`
	MessageID string    `json:"message_id"`
	UserID    string    `json:"user_id"`
	User      *User     `json:"user,omitempty"`
	Emoji     string    `json:"emoji"`
	CreatedAt time.Time `json:"created_at"`
}

type AddReactionRequest struct {
	Emoji string `json:"emoji" binding:"required"`
}
