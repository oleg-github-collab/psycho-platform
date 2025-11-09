package handlers

import (
	"database/sql"
	"net/http"
	"psycho-platform/internal/auth"
	"psycho-platform/internal/config"
	"psycho-platform/internal/models"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	db  *sql.DB
	cfg *config.Config
}

func NewAuthHandler(db *sql.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{db: db, cfg: cfg}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if username exists
	var exists bool
	err := h.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", req.Username).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user
	displayName := req.DisplayName
	if displayName == "" {
		displayName = req.Username
	}

	var user models.User
	err = h.db.QueryRow(`
		INSERT INTO users (username, password_hash, display_name, role)
		VALUES ($1, $2, $3, 'basic')
		RETURNING id,
			username,
			COALESCE(display_name, username),
			COALESCE(avatar_url, ''),
			COALESCE(bio, ''),
			COALESCE(role, 'basic'),
			is_active,
			created_at,
			updated_at
	`, req.Username, hashedPassword, displayName).Scan(
		&user.ID, &user.Username, &user.DisplayName, &user.AvatarURL,
		&user.Bio, &user.Role, &user.IsActive,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate token
	token, err := auth.GenerateToken(user.ID, user.Role, h.cfg.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, models.AuthResponse{
		Token: token,
		User:  &user,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	err := h.db.QueryRow(`
		SELECT id,
			username,
			password_hash,
			COALESCE(display_name, username),
			COALESCE(avatar_url, ''),
			COALESCE(bio, ''),
			COALESCE(role, 'basic'),
			is_active,
			created_at,
			updated_at
		FROM users
		WHERE username = $1
	`, req.Username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.DisplayName,
		&user.AvatarURL, &user.Bio, &user.Role,
		&user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "Account is disabled"})
		return
	}

	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Role, h.cfg.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{
		Token: token,
		User:  &user,
	})
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	userID := c.GetString("user_id")

	var user models.User
	err := h.db.QueryRow(`
		SELECT id,
			username,
			COALESCE(display_name, username),
			COALESCE(avatar_url, ''),
			COALESCE(bio, ''),
			COALESCE(role, 'basic'),
			is_active,
			created_at,
			updated_at
		FROM users
		WHERE id = $1
	`, userID).Scan(
		&user.ID, &user.Username, &user.DisplayName, &user.AvatarURL,
		&user.Bio, &user.Role, &user.IsActive,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}
