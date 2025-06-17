package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type FileManager struct {
	sessions map[string]*SessionFiles
	mu       sync.RWMutex
	ttl      time.Duration
}

type SessionFiles struct {
	Dir       string
	CreatedAt time.Time
	Files     map[string]string
}

func NewFileManager(ttl time.Duration) *FileManager {
	fm := &FileManager{
		sessions: make(map[string]*SessionFiles),
		ttl:      ttl,
	}
	
	go fm.cleanup()
	return fm
}

func (fm *FileManager) StoreFile(sessionID, originalPath, filename string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	session, exists := fm.sessions[sessionID]
	if !exists {
		sessionDir := filepath.Join("/tmp", "claude-web", sessionID)
		if err := os.MkdirAll(sessionDir, 0755); err != nil {
			return err
		}
		
		session = &SessionFiles{
			Dir:       sessionDir,
			CreatedAt: time.Now(),
			Files:     make(map[string]string),
		}
		fm.sessions[sessionID] = session
	}

	destPath := filepath.Join(session.Dir, filename)
	
	src, err := os.Open(originalPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	session.Files[filename] = destPath
	return nil
}

func (fm *FileManager) GetFile(sessionID, filename string) (string, error) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	session, exists := fm.sessions[sessionID]
	if !exists {
		return "", fmt.Errorf("session not found")
	}

	path, exists := session.Files[filename]
	if !exists {
		return "", fmt.Errorf("file not found")
	}

	return path, nil
}

func (fm *FileManager) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		fm.mu.Lock()
		now := time.Now()
		
		for sessionID, session := range fm.sessions {
			if now.Sub(session.CreatedAt) > fm.ttl {
				os.RemoveAll(session.Dir)
				delete(fm.sessions, sessionID)
			}
		}
		fm.mu.Unlock()
	}
}