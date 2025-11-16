package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {

	SupabaseURL     string
	SupabaseAnonKey string


	AWSRegion          string
	AWSAccessKeyID     string
	AWSSecretAccessKey string


	LambdaARNTypeScript string
	LambdaARNPython     string
	LambdaARNDart       string
	LambdaARNGo         string
	LambdaARNCpp        string


	Port string
	Env  string

	LokiURL string

	UseMockLambda bool
	UseMockAuth   bool

	
	
	STUNServers []string
	
	TURNServers     []string
	TURNUsername    string
	TURNPassword    string
}

func Load() *Config {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	
	defaultSTUN := []string{
		"stun:stun.l.google.com:19302",
		"stun:stun1.l.google.com:19302",
		"stun:stun2.l.google.com:19302",
	}

	return &Config{
		SupabaseURL:         getEnv("SUPABASE_URL", ""),
		SupabaseAnonKey:     getEnv("SUPABASE_ANON_KEY", ""),
		AWSRegion:           getEnv("AWS_REGION", "us-east-1"),
		AWSAccessKeyID:      getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey:  getEnv("AWS_SECRET_ACCESS_KEY", ""),
		LambdaARNTypeScript: getEnv("LAMBDA_ARN_TYPESCRIPT", ""),
		LambdaARNPython:     getEnv("LAMBDA_ARN_PYTHON", ""),
		LambdaARNDart:       getEnv("LAMBDA_ARN_DART", ""),
		LambdaARNGo:         getEnv("LAMBDA_ARN_GO", ""),
		LambdaARNCpp:        getEnv("LAMBDA_ARN_CPP", ""),
		Port:                getEnv("PORT", "8080"),
		Env:                 getEnv("ENV", "development"),
		LokiURL:             getEnv("LOKI_URL", "http:
		UseMockLambda:       getBoolEnv("USE_MOCK_LAMBDA", false),
		UseMockAuth:         getBoolEnv("USE_MOCK_AUTH", false),

		
		STUNServers:      getSliceEnv("STUN_SERVERS", defaultSTUN),
		TURNServers:      getSliceEnv("TURN_SERVERS", []string{}),
		TURNUsername:     getEnv("TURN_USERNAME", ""),
		TURNPassword:     getEnv("TURN_PASSWORD", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getSliceEnv(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		
		
		result := []string{}
		for _, v := range splitAndTrim(value, ",") {
			if v != "" {
				result = append(result, v)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}

func splitAndTrim(s, sep string) []string {
	parts := []string{}
	for _, part := range split(s, sep) {
		trimmed := trimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func split(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
