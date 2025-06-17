package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"claude-web-go/internal/claude"
	"claude-web-go/internal/models"
	"claude-web-go/internal/storage"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Server struct {
	executor    *claude.Executor
	fileManager *storage.FileManager
	upgrader    websocket.Upgrader
}

func NewServer() (*Server, error) {
	executor, err := claude.NewExecutor()
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}

	return &Server{
		executor:    executor,
		fileManager: storage.NewFileManager(30 * time.Minute),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}, nil
}

func (s *Server) HandleChat(w http.ResponseWriter, r *http.Request) {
	var req models.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		req.SessionID = uuid.New().String()
	}

	output, files, err := s.executor.Execute(req.Message, req.ContextWindow)
	
	response := models.ChatResponse{
		SessionID: req.SessionID,
		Message: models.Message{
			ID:        uuid.New().String(),
			Role:      "assistant",
			Content:   output,
			Timestamp: time.Now(),
		},
	}

	if err != nil {
		response.Error = err.Error()
	}

	if len(files) > 0 {
		for _, file := range files {
			if err := s.fileManager.StoreFile(req.SessionID, file.Path, file.Name); err == nil {
				response.Files = append(response.Files, file)
			}
		}
		response.Message.Files = response.Files
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) HandleFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["sessionId"]
	filename := vars["filename"]

	filePath, err := s.fileManager.GetFile(sessionID, filename)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	mimeType := getMimeType(filename)
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Disposition", "inline; filename=\""+filename+"\"")
	
	io.Copy(w, file)
}

func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		var req models.ChatRequest
		if err := conn.ReadJSON(&req); err != nil {
			break
		}

		output, files, err := s.executor.Execute(req.Message, req.ContextWindow)
		
		response := models.ChatResponse{
			SessionID: req.SessionID,
			Message: models.Message{
				ID:        uuid.New().String(),
				Role:      "assistant",
				Content:   output,
				Timestamp: time.Now(),
			},
		}

		if err != nil {
			response.Error = err.Error()
		}

		if len(files) > 0 {
			for _, file := range files {
				if err := s.fileManager.StoreFile(req.SessionID, file.Path, file.Name); err == nil {
					response.Files = append(response.Files, file)
				}
			}
			response.Message.Files = response.Files
		}

		if err := conn.WriteJSON(response); err != nil {
			break
		}
	}
}

func getMimeType(filename string) string {
	ext := filepath.Ext(filename)
	switch ext {
	case ".png":
		return "image/png"
	case ".svg":
		return "image/svg+xml"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	default:
		return "application/octet-stream"
	}
}