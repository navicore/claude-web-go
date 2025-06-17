package models

import "time"

type Message struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Files     []File    `json:"files,omitempty"`
}

type File struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	MimeType string `json:"mimeType"`
	Size     int64  `json:"size"`
}

type ChatRequest struct {
	Message       string    `json:"message"`
	SessionID     string    `json:"sessionId"`
	ContextWindow []Message `json:"contextWindow"`
}

type ChatResponse struct {
	SessionID string    `json:"sessionId"`
	Message   Message   `json:"message"`
	Files     []File    `json:"files"`
	Error     string    `json:"error,omitempty"`
}