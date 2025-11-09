package handlers

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FileHandler struct {
	db          *sql.DB
	uploadDir   string
	maxFileSize int64
}

func NewFileHandler(db *sql.DB) *FileHandler {
	uploadDir := "./uploads"
	os.MkdirAll(uploadDir, 0755)
	os.MkdirAll(filepath.Join(uploadDir, "images"), 0755)
	os.MkdirAll(filepath.Join(uploadDir, "documents"), 0755)
	os.MkdirAll(filepath.Join(uploadDir, "audio"), 0755)

	return &FileHandler{
		db:          db,
		uploadDir:   uploadDir,
		maxFileSize: 50 * 1024 * 1024, // 50MB
	}
}

type FileAttachment struct {
	ID        string `json:"id"`
	MessageID string `json:"message_id"`
	Filename  string `json:"filename"`
	FileType  string `json:"file_type"`
	FileSize  int64  `json:"file_size"`
	FileURL   string `json:"file_url"`
	CreatedAt string `json:"created_at"`
}

func (h *FileHandler) UploadFile(c *gin.Context) {
	userID := c.GetString("user_id")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}
	defer file.Close()

	// Check file size
	if header.Size > h.maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large (max 50MB)"})
		return
	}

	// Determine file type
	fileType := getFileType(header.Filename)
	subDir := getSubDirectory(fileType)

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	filename := uuid.New().String() + ext
	filePath := filepath.Join(h.uploadDir, subDir, filename)

	// Create file
	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	defer dst.Close()

	// Copy file
	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Save to database
	fileURL := fmt.Sprintf("/uploads/%s/%s", subDir, filename)
	var fileID string
	err = h.db.QueryRow(`
		INSERT INTO file_attachments (user_id, filename, original_name, file_type, file_size, file_url)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, userID, filename, header.Filename, fileType, header.Size, fileURL).Scan(&fileID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file metadata"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        fileID,
		"filename":  header.Filename,
		"file_type": fileType,
		"file_size": header.Size,
		"file_url":  fileURL,
	})
}

func (h *FileHandler) AttachToMessage(c *gin.Context) {
	messageID := c.Param("message_id")
	var req struct {
		FileID string `json:"file_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := h.db.Exec(`
		UPDATE file_attachments
		SET message_id = $1
		WHERE id = $2
	`, messageID, req.FileID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to attach file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *FileHandler) GetMessageFiles(c *gin.Context) {
	messageID := c.Param("message_id")

	rows, err := h.db.Query(`
		SELECT id, filename, original_name, file_type, file_size, file_url, created_at
		FROM file_attachments
		WHERE message_id = $1
		ORDER BY created_at ASC
	`, messageID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch files"})
		return
	}
	defer rows.Close()

	files := []map[string]interface{}{}
	for rows.Next() {
		var id, filename, originalName, fileType, fileURL, createdAt string
		var fileSize int64
		rows.Scan(&id, &filename, &originalName, &fileType, &fileSize, &fileURL, &createdAt)
		files = append(files, map[string]interface{}{
			"id":            id,
			"filename":      filename,
			"original_name": originalName,
			"file_type":     fileType,
			"file_size":     fileSize,
			"file_url":      fileURL,
			"created_at":    createdAt,
		})
	}

	c.JSON(http.StatusOK, files)
}

func (h *FileHandler) DeleteFile(c *gin.Context) {
	userID := c.GetString("user_id")
	fileID := c.Param("id")

	var filename, fileURL string
	err := h.db.QueryRow(`
		SELECT filename, file_url FROM file_attachments
		WHERE id = $1 AND user_id = $2
	`, fileID, userID).Scan(&filename, &fileURL)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Delete from filesystem
	filePath := filepath.Join(h.uploadDir, strings.TrimPrefix(fileURL, "/uploads/"))
	os.Remove(filePath)

	// Delete from database
	_, err = h.db.Exec("DELETE FROM file_attachments WHERE id = $1", fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func getFileType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg":
		return "image"
	case ".pdf":
		return "pdf"
	case ".doc", ".docx", ".txt", ".rtf":
		return "document"
	case ".mp3", ".wav", ".ogg", ".m4a", ".webm":
		return "audio"
	case ".mp4", ".avi", ".mov", ".webm":
		return "video"
	case ".zip", ".rar", ".7z", ".tar", ".gz":
		return "archive"
	default:
		return "other"
	}
}

func getSubDirectory(fileType string) string {
	switch fileType {
	case "image":
		return "images"
	case "audio":
		return "audio"
	default:
		return "documents"
	}
}
