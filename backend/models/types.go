package models

import "time"


type AnalyzeRequest struct {
	Action   string `json:"action"`
	Language string `json:"language"`
	Code     string `json:"code"`
}


type LintError struct {
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Message  string `json:"message"`
	Severity string `json:"severity"` 
	Length   int    `json:"length"`
}


type AnalyzeResponse struct {
	Type          string      `json:"type"` 
	Errors        []LintError `json:"errors,omitempty"`
	ErrorMessage  string      `json:"message,omitempty"`
	ExecutionTime int         `json:"executionTime,omitempty"` 
}


type Connection struct {
	UserID   string
	LastSeen time.Time
}


type LambdaRequest struct {
	Language string `json:"language"`
	Code     string `json:"code"`
}


type LambdaResponse struct {
	Errors []LintError `json:"errors"`
}
