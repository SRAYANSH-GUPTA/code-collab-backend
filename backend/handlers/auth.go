package handlers

import (
	"fmt"
	"strings"
	"time"

	"codecollab/config"
	"codecollab/utils"

	"github.com/golang-jwt/jwt/v5"
)

var logger = utils.NewLogger("auth")

type SupabaseClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func VerifyToken(token string, cfg *config.Config) (string, error) {
	if cfg.UseMockAuth {
		logger.Info("Using mock auth - accepting token: %s", token[:min(10, len(token))])
		return "mock-user-" + token, nil
	}

	if token == "" {
		return "", fmt.Errorf("missing token")
	}

	if cfg.SupabaseJWTSecret == "" {
		return "", fmt.Errorf("SUPABASE_JWT_SECRET is missing")
	}

	claims := &SupabaseClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(parsedToken *jwt.Token) (any, error) {
		if _, ok := parsedToken.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %s", parsedToken.Method.Alg())
		}
		return []byte(cfg.SupabaseJWTSecret), nil
	}, jwt.WithExpirationRequired(), jwt.WithIssuedAt(), jwt.WithLeeway(30*time.Second))
	if err != nil {
		return "", fmt.Errorf("invalid auth token: %w", err)
	}

	if !parsedToken.Valid {
		return "", fmt.Errorf("invalid auth token")
	}

	expectedIssuer := strings.TrimRight(cfg.SupabaseURL, "/") + "/auth/v1"
	if cfg.SupabaseURL != "" && claims.Issuer != expectedIssuer {
		return "", fmt.Errorf("invalid token issuer")
	}

	if claims.Subject == "" {
		return "", fmt.Errorf("token subject missing")
	}

	logger.Info("JWT verified for user: %s", claims.Subject)
	return claims.Subject, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
