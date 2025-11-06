package utils

import (
	"fmt"
	"log"
	"time"
)


type Logger struct {
	prefix string
}


func NewLogger(prefix string) *Logger {
	return &Logger{prefix: prefix}
}


func (l *Logger) Info(message string, args ...interface{}) {
	log.Printf("[INFO] [%s] %s\n", l.prefix, fmt.Sprintf(message, args...))
}


func (l *Logger) Error(message string, args ...interface{}) {
	log.Printf("[ERROR] [%s] %s\n", l.prefix, fmt.Sprintf(message, args...))
}


func (l *Logger) Warn(message string, args ...interface{}) {
	log.Printf("[WARN] [%s] %s\n", l.prefix, fmt.Sprintf(message, args...))
}


func (l *Logger) Debug(message string, args ...interface{}) {
	log.Printf("[DEBUG] [%s] %s\n", l.prefix, fmt.Sprintf(message, args...))
}


func LogRequest(userID, action string, duration time.Duration) {
	log.Printf("[REQUEST] UserID: %s, Action: %s, Duration: %dms\n", userID, action, duration.Milliseconds())
}


func LogConnection(event, userID string) {
	log.Printf("[CONNECTION] Event: %s, UserID: %s\n", event, userID)
}


func LogLambda(language string, duration time.Duration, errorCount int) {
	log.Printf("[LAMBDA] Language: %s, Duration: %dms, Errors: %d\n", language, duration.Milliseconds(), errorCount)
}
