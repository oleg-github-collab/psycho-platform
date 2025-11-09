package models

import "time"

type Session struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	SessionType     string    `json:"session_type"`
	HMSRoomID       string    `json:"hms_room_id,omitempty"`
	HMSRoomCode     string    `json:"hms_room_code,omitempty"`
	PsychologistID  string    `json:"psychologist_id"`
	Psychologist    *User     `json:"psychologist,omitempty"`
	MaxParticipants int       `json:"max_participants"`
	ScheduledAt     time.Time `json:"scheduled_at"`
	DurationMinutes int       `json:"duration_minutes"`
	IsPrivate       bool      `json:"is_private"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CreateSessionRequest struct {
	Title           string    `json:"title" binding:"required"`
	Description     string    `json:"description"`
	SessionType     string    `json:"session_type"`
	MaxParticipants int       `json:"max_participants"`
	ScheduledAt     time.Time `json:"scheduled_at" binding:"required"`
	DurationMinutes int       `json:"duration_minutes"`
	IsPrivate       bool      `json:"is_private"`
}

type Appointment struct {
	ID              string     `json:"id"`
	PsychologistID  string     `json:"psychologist_id"`
	Psychologist    *User      `json:"psychologist,omitempty"`
	ClientID        string     `json:"client_id"`
	Client          *User      `json:"client,omitempty"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	ScheduledAt     time.Time  `json:"scheduled_at"`
	DurationMinutes int        `json:"duration_minutes"`
	Status          string     `json:"status"`
	Notes           string     `json:"notes"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type CreateAppointmentRequest struct {
	PsychologistID  string    `json:"psychologist_id" binding:"required"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	ScheduledAt     time.Time `json:"scheduled_at" binding:"required"`
	DurationMinutes int       `json:"duration_minutes"`
}
