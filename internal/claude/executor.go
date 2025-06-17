package claude

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"claude-web-go/internal/auth"
	"claude-web-go/internal/models"
	"github.com/google/uuid"
)

type Executor struct {
	tmpDir    string
	awsConfig *auth.AWSConfig
}

func NewExecutor() (*Executor, error) {
	awsConfig, err := auth.GetAWSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get AWS config: %w", err)
	}

	if err := auth.SetupEnvironment(awsConfig); err != nil {
		return nil, fmt.Errorf("failed to setup AWS environment: %w", err)
	}

	return &Executor{
		tmpDir:    "/tmp",
		awsConfig: awsConfig,
	}, nil
}

func (e *Executor) Execute(prompt string, contextWindow []models.Message) (string, []models.File, error) {
	sessionDir := filepath.Join(e.tmpDir, uuid.New().String())
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return "", nil, fmt.Errorf("failed to create session directory: %w", err)
	}
	defer os.RemoveAll(sessionDir)

	fullPrompt := e.buildPromptWithContext(prompt, contextWindow)

	cmd := exec.Command("claude", "-p", "--print")
	cmd.Stdin = strings.NewReader(fullPrompt)
	cmd.Dir = sessionDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", nil, fmt.Errorf("claude execution failed: %w, stderr: %s", err, stderr.String())
	}

	files, err := e.scanForFiles(sessionDir)
	if err != nil {
		return stdout.String(), nil, err
	}

	return stdout.String(), files, nil
}

func (e *Executor) buildPromptWithContext(prompt string, contextWindow []models.Message) string {
	var contextBuilder strings.Builder
	
	if len(contextWindow) > 0 {
		contextBuilder.WriteString("Previous conversation context:\n")
		for _, msg := range contextWindow {
			contextBuilder.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
		}
		contextBuilder.WriteString("\n---\n\n")
	}
	
	contextBuilder.WriteString(prompt)
	return contextBuilder.String()
}

func (e *Executor) scanForFiles(dir string) ([]models.File, error) {
	var files []models.File
	
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		file := models.File{
			Name:     entry.Name(),
			Path:     filepath.Join(dir, entry.Name()),
			Size:     info.Size(),
			MimeType: getMimeType(entry.Name()),
		}

		files = append(files, file)
	}

	return files, nil
}

func getMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".png":
		return "image/png"
	case ".svg":
		return "image/svg+xml"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".txt":
		return "text/plain"
	case ".json":
		return "application/json"
	case ".html":
		return "text/html"
	default:
		return "application/octet-stream"
	}
}