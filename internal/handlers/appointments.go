package handlers

import (
	"database/sql"
	"net/http"
	"psycho-platform/internal/models"

	"github.com/gin-gonic/gin"
)

type AppointmentHandler struct {
	db *sql.DB
}

func NewAppointmentHandler(db *sql.DB) *AppointmentHandler {
	return &AppointmentHandler{db: db}
}

func (h *AppointmentHandler) CreateAppointment(c *gin.Context) {
	userID := c.GetString("user_id")
	var req models.CreateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	durationMinutes := req.DurationMinutes
	if durationMinutes == 0 {
		durationMinutes = 60
	}

	var appointment models.Appointment
	err := h.db.QueryRow(`
		INSERT INTO appointments (psychologist_id, client_id, title, description, scheduled_at, duration_minutes, status)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending')
		RETURNING id, psychologist_id, client_id, title, description, scheduled_at, duration_minutes, status, notes, created_at, updated_at
	`, req.PsychologistID, userID, req.Title, req.Description, req.ScheduledAt, durationMinutes).Scan(
		&appointment.ID, &appointment.PsychologistID, &appointment.ClientID,
		&appointment.Title, &appointment.Description, &appointment.ScheduledAt,
		&appointment.DurationMinutes, &appointment.Status, &appointment.Notes,
		&appointment.CreatedAt, &appointment.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create appointment"})
		return
	}

	c.JSON(http.StatusCreated, appointment)
}

func (h *AppointmentHandler) GetAppointments(c *gin.Context) {
	userID := c.GetString("user_id")

	query := `
		SELECT a.id, a.psychologist_id, a.client_id, a.title, a.description,
		       a.scheduled_at, a.duration_minutes, a.status, a.notes, a.created_at, a.updated_at,
		       p.username as p_username, p.display_name as p_display_name, p.avatar_url as p_avatar,
		       cl.username as c_username, cl.display_name as c_display_name, cl.avatar_url as c_avatar
		FROM appointments a
		JOIN users p ON a.psychologist_id = p.id
		JOIN users cl ON a.client_id = cl.id
		WHERE a.psychologist_id = $1 OR a.client_id = $1
		ORDER BY a.scheduled_at ASC
	`

	rows, err := h.db.Query(query, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch appointments"})
		return
	}
	defer rows.Close()

	appointments := []models.Appointment{}
	for rows.Next() {
		var apt models.Appointment
		var psychologist, client models.User
		err := rows.Scan(
			&apt.ID, &apt.PsychologistID, &apt.ClientID, &apt.Title, &apt.Description,
			&apt.ScheduledAt, &apt.DurationMinutes, &apt.Status, &apt.Notes,
			&apt.CreatedAt, &apt.UpdatedAt,
			&psychologist.Username, &psychologist.DisplayName, &psychologist.AvatarURL,
			&client.Username, &client.DisplayName, &client.AvatarURL,
		)
		if err != nil {
			continue
		}
		psychologist.ID = apt.PsychologistID
		client.ID = apt.ClientID
		apt.Psychologist = &psychologist
		apt.Client = &client
		appointments = append(appointments, apt)
	}

	c.JSON(http.StatusOK, appointments)
}

func (h *AppointmentHandler) UpdateAppointmentStatus(c *gin.Context) {
	appointmentID := c.Param("id")
	status := c.Query("status")

	if status != "confirmed" && status != "cancelled" && status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		return
	}

	_, err := h.db.Exec("UPDATE appointments SET status = $1 WHERE id = $2", status, appointmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update appointment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
