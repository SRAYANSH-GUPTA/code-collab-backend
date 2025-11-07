package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"codecollab/config"
	"codecollab/utils"
)

var logger = utils.NewLogger("auth")

type SupabaseUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func VerifyToken(token string, cfg *config.Config) (string, error) {

	if cfg.UseMockAuth {
		logger.Info("Using mock auth - accepting token: %s", token[:min(10, len(token))])

		return "mock-user-" + token[:min(8, len(token))], nil
	}

	if cfg.SupabaseURL == "" || cfg.SupabaseAnonKey == "" {
		return "", fmt.Errorf("Supabase configuration missing")
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("%s/auth/v1/user", cfg.SupabaseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("apikey", cfg.SupabaseAnonKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to verify token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("invalid token: %s", string(body))
	}

	var user SupabaseUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", fmt.Errorf("failed to decode user: %w", err)
	}

	logger.Info("Token verified for user: %s", user.ID)
	return user.ID, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
