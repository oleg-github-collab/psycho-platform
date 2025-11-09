package hms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Client struct {
	APIKey    string
	APISecret string
	BaseURL   string
}

type CreateRoomRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateRoomResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type RoomTokenRequest struct {
	RoomID string
	UserID string
	Role   string
}

func NewClient(apiKey, apiSecret string) *Client {
	return &Client{
		APIKey:    apiKey,
		APISecret: apiSecret,
		BaseURL:   "https://api.100ms.live/v2",
	}
}

func (c *Client) CreateRoom(name, description string) (*CreateRoomResponse, error) {
	reqBody := CreateRoomRequest{
		Name:        name,
		Description: description,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/rooms", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.generateManagementToken())

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create room: status %d", resp.StatusCode)
	}

	var roomResp CreateRoomResponse
	if err := json.NewDecoder(resp.Body).Decode(&roomResp); err != nil {
		return nil, err
	}

	return &roomResp, nil
}

func (c *Client) GenerateRoomToken(req RoomTokenRequest) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"access_key": c.APIKey,
		"room_id":    req.RoomID,
		"user_id":    req.UserID,
		"role":       req.Role,
		"type":       "app",
		"version":    2,
		"iat":        now.Unix(),
		"nbf":        now.Unix(),
		"exp":        now.Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(c.APISecret))
}

func (c *Client) generateManagementToken() string {
	now := time.Now()
	claims := jwt.MapClaims{
		"access_key": c.APIKey,
		"type":       "management",
		"version":    2,
		"iat":        now.Unix(),
		"nbf":        now.Unix(),
		"exp":        now.Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(c.APISecret))
	return tokenString
}
