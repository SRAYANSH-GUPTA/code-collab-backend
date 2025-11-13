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
}

func Load() *Config {
	
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
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
		LokiURL:             getEnv("LOKI_URL", "http://loki:3100"),
		UseMockLambda:       getBoolEnv("USE_MOCK_LAMBDA", false),
		UseMockAuth:         getBoolEnv("USE_MOCK_AUTH", false),
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
