package handlers

import (
	"database/sql"
	"net/http"
	"psycho-platform/internal/config"
	"psycho-platform/internal/models"

	"github.com/gin-gonic/gin"
)

type SessionHandler struct {
	db  *sql.DB
	cfg *config.Config
}

func NewSessionHandler(db *sql.DB, cfg *config.Config) *SessionHandler {
	return &SessionHandler{db: db, cfg: cfg}
}

func (h *SessionHandler) CreateSession(c *gin.Context) {
	userID := c.GetString("user_id")
	var req models.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sessionType := req.SessionType
	if sessionType == "" {
		sessionType = "webinar"
	}

	maxParticipants := req.MaxParticipants
	if maxParticipants == 0 {
		maxParticipants = 50
	}

	durationMinutes := req.DurationMinutes
	if durationMinutes == 0 {
		durationMinutes = 60
	}

	var session models.Session
	err := h.db.QueryRow(`
		INSERT INTO sessions (title, description, session_type, psychologist_id, max_participants, scheduled_at, duration_minutes, is_private, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'scheduled')
		RETURNING id, title, description, session_type, hms_room_id, hms_room_code, psychologist_id, max_participants, scheduled_at, duration_minutes, is_private, status, created_at, updated_at
	`, req.Title, req.Description, sessionType, userID, maxParticipants, req.ScheduledAt, durationMinutes, req.IsPrivate).Scan(
		&session.ID, &session.Title, &session.Description, &session.SessionType,
		&session.HMSRoomID, &session.HMSRoomCode, &session.PsychologistID,
		&session.MaxParticipants, &session.ScheduledAt, &session.DurationMinutes,
		&session.IsPrivate, &session.Status, &session.CreatedAt, &session.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	// TODO: Integrate with 100ms API to create room
	// For now, generate placeholder room code
	h.db.Exec("UPDATE sessions SET hms_room_code = $1 WHERE id = $2", "room_"+session.ID[:8], session.ID)

	c.JSON(http.StatusCreated, session)
}

func (h *SessionHandler) GetSessions(c *gin.Context) {
	query := `
		SELECT s.id, s.title, s.description, s.session_type, s.hms_room_id, s.hms_room_code,
		       s.psychologist_id, s.max_participants, s.scheduled_at, s.duration_minutes,
		       s.is_private, s.status, s.created_at, s.updated_at,
		       u.username, u.display_name, u.avatar_url
		FROM sessions s
		JOIN users u ON s.psychologist_id = u.id
		WHERE s.status != 'cancelled' AND (s.is_private = false OR s.psychologist_id = $1)
		ORDER BY s.scheduled_at ASC
	`

	userID := c.GetString("user_id")
	rows, err := h.db.Query(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sessions"})
		return
	}
	defer rows.Close()

	sessions := []models.Session{}
	for rows.Next() {
		var session models.Session
		var psychologist models.User
		err := rows.Scan(
			&session.ID, &session.Title, &session.Description, &session.SessionType,
			&session.HMSRoomID, &session.HMSRoomCode, &session.PsychologistID,
			&session.MaxParticipants, &session.ScheduledAt, &session.DurationMinutes,
			&session.IsPrivate, &session.Status, &session.CreatedAt, &session.UpdatedAt,
			&psychologist.Username, &psychologist.DisplayName, &psychologist.AvatarURL,
		)
		if err != nil {
			continue
		}
		psychologist.ID = session.PsychologistID
		session.Psychologist = &psychologist
		sessions = append(sessions, session)
	}

	c.JSON(http.StatusOK, sessions)
}

func (h *SessionHandler) GetRoomToken(c *gin.Context) {
	sessionID := c.Param("id")
	userID := c.GetString("user_id")

	var roomCode string
	err := h.db.QueryRow("SELECT hms_room_code FROM sessions WHERE id = $1", sessionID).Scan(&roomCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	// TODO: Generate actual 100ms token
	// For now, return placeholder
	c.JSON(http.StatusOK, gin.H{
		"room_code": roomCode,
		"token":     "placeholder_token_" + userID,
	})
}
