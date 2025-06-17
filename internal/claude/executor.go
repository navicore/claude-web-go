package claude

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"claude-web-go/internal/auth"
	"claude-web-go/internal/logger"
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

	// Check if /tmp exists and is writable
	tmpDir := "/tmp"
	if stat, err := os.Stat(tmpDir); err != nil {
		logger.Log.WithError(err).Error("Cannot access /tmp directory")
		return nil, fmt.Errorf("cannot access /tmp directory: %w", err)
	} else {
		logger.Log.WithFields(map[string]interface{}{
			"tmpDir": tmpDir,
			"isDir":  stat.IsDir(),
			"mode":   stat.Mode(),
		}).Debug("Using tmp directory")
	}

	// Get current working directory for debugging
	if cwd, err := os.Getwd(); err == nil {
		logger.Log.WithField("cwd", cwd).Debug("Current working directory")
	}

	// Check if claude is in PATH
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		logger.Log.WithError(err).Error("Claude CLI not found in PATH")
		return nil, fmt.Errorf("claude CLI not found in PATH: %w", err)
	}
	logger.Log.WithField("claudePath", claudePath).Info("Claude CLI found")

	// Test claude version
	versionCmd := exec.Command("claude", "--version")
	if output, err := versionCmd.CombinedOutput(); err != nil {
		logger.Log.WithError(err).WithField("output", string(output)).Warn("Failed to get claude version")
	} else {
		logger.Log.WithField("version", strings.TrimSpace(string(output))).Info("Claude CLI version")
	}

	// Test gamecode-mcp2 if MCP is configured
	if mcpConfig := os.Getenv("CLAUDE_MCP_CONFIG"); mcpConfig != "" {
		logger.Log.Info("MCP configuration detected, checking gamecode-mcp2...")
		mcpCmd := exec.Command("gamecode-mcp2", "--version")
		if output, err := mcpCmd.CombinedOutput(); err != nil {
			logger.Log.WithError(err).WithField("output", string(output)).Error("Failed to get gamecode-mcp2 version - MCP may not work")
		} else {
			logger.Log.WithField("version", strings.TrimSpace(string(output))).Info("gamecode-mcp2 version")
		}
		
		// Check if tools.yaml exists
		toolsPath := "/app/mcp/tools.yaml"
		if stat, err := os.Stat(toolsPath); err != nil {
			logger.Log.WithError(err).Error("Cannot access tools.yaml file")
		} else {
			logger.Log.WithFields(map[string]interface{}{
				"path": toolsPath,
				"size": stat.Size(),
				"mode": stat.Mode(),
			}).Debug("tools.yaml file found")
		}
	}

	// Log environment variables that claude might need
	logger.Log.WithFields(map[string]interface{}{
		"AWS_REGION": os.Getenv("AWS_REGION"),
		"AWS_ACCESS_KEY_ID_exists": os.Getenv("AWS_ACCESS_KEY_ID") != "",
		"AWS_SECRET_ACCESS_KEY_exists": os.Getenv("AWS_SECRET_ACCESS_KEY") != "",
		"AWS_SESSION_TOKEN_exists": os.Getenv("AWS_SESSION_TOKEN") != "",
		"CLAUDE_CODE_USE_BEDROCK": os.Getenv("CLAUDE_CODE_USE_BEDROCK"),
		"ANTHROPIC_MODEL": os.Getenv("ANTHROPIC_MODEL"),
		"ANTHROPIC_SMALL_FAST_MODEL": os.Getenv("ANTHROPIC_SMALL_FAST_MODEL"),
		"ANTHROPIC_API_KEY": os.Getenv("ANTHROPIC_API_KEY"),
		"ANTHROPIC_ENDPOINT_URL": os.Getenv("ANTHROPIC_ENDPOINT_URL"),
		"HOME": os.Getenv("HOME"),
	}).Info("Environment variables before claude test")

	// First test without AWS credentials to see if it still tries to connect
	logger.Log.Info("Testing if claude uses Bedrock...")
	testEnv := []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + os.Getenv("HOME"),
		"CLAUDE_CODE_USE_BEDROCK=1",
		"ANTHROPIC_MODEL=" + os.Getenv("ANTHROPIC_MODEL"),
	}
	noCredsCmd := exec.Command("claude", "--version")
	noCredsCmd.Env = testEnv
	if output, err := noCredsCmd.CombinedOutput(); err != nil {
		logger.Log.WithError(err).WithField("output", string(output)).Debug("Claude version without AWS creds")
	}

	// Now test with a prompt but no AWS credentials - should fail if using Bedrock
	noCredsTestCmd := exec.Command("claude", "-p", "Hi")
	noCredsTestCmd.Env = testEnv
	if output, err := noCredsTestCmd.CombinedOutput(); err != nil {
		outputStr := string(output)
		if strings.Contains(outputStr, "AWS") || strings.Contains(outputStr, "credentials") {
			logger.Log.Info("Good: Claude appears to need AWS credentials (using Bedrock)")
		} else if strings.Contains(outputStr, "429") || strings.Contains(outputStr, "API key") {
			logger.Log.Error("BAD: Claude still trying to use Anthropic API without AWS credentials!")
		}
		logger.Log.WithField("output", outputStr).Debug("Claude test without AWS creds")
	}

	// Test a simple claude command with timeout
	logger.Log.Info("Testing claude command with full credentials...")
	
	// First test without MCP to isolate issues
	logger.Log.Info("Testing without MCP config first...")
	simpleCtx, simpleCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer simpleCancel()
	
	simpleArgs := []string{}
	if model := os.Getenv("ANTHROPIC_MODEL"); model != "" {
		simpleArgs = append(simpleArgs, "--model", model)
	}
	simpleArgs = append(simpleArgs, "-p", "Say hello")
	
	simpleCmd := exec.CommandContext(simpleCtx, "claude", simpleArgs...)
	simpleCmd.Env = os.Environ()
	
	// Log the exact command being run
	logger.Log.WithFields(map[string]interface{}{
		"command": "claude",
		"args": simpleArgs,
		"envCount": len(simpleCmd.Env),
	}).Debug("Running simple test command")
	
	var simpleStdout, simpleStderr bytes.Buffer
	simpleCmd.Stdout = &simpleStdout
	simpleCmd.Stderr = &simpleStderr
	
	if err := simpleCmd.Run(); err != nil {
		logger.Log.WithError(err).WithFields(map[string]interface{}{
			"stderr": simpleStderr.String(),
			"stdout": simpleStdout.String(),
		}).Warn("Simple test failed")
	} else {
		logger.Log.WithField("output", simpleStdout.String()).Info("Simple test succeeded - MCP might be the issue")
	}
	
	// Now test with full configuration
	testCtx, testCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer testCancel()

	// Build test args
	testArgs := []string{}
	if os.Getenv("LOG_LEVEL") == "debug" {
		testArgs = append(testArgs, "--debug")
	}
	if model := os.Getenv("ANTHROPIC_MODEL"); model != "" {
		testArgs = append(testArgs, "--model", model)
	}
	if allowedTools := os.Getenv("CLAUDE_ALLOWED_TOOLS"); allowedTools != "" {
		testArgs = append(testArgs, "--allowedTools", allowedTools)
	}
	// Disallowed tools with default if not set
	disallowedTools := os.Getenv("CLAUDE_DISALLOWED_TOOLS")
	if disallowedTools == "" {
		disallowedTools = "Bash,Glob,Grep,LS,Read,Edit,MultiEdit,Write,NotebookRead,NotebookEdit,WebFetch,TodoRead,TodoWrite,Task"
	}
	testArgs = append(testArgs, "--disallowedTools", disallowedTools)
	// MCP config support for test
	if mcpConfig := os.Getenv("CLAUDE_MCP_CONFIG"); mcpConfig != "" {
		logger.Log.WithField("mcp_config_test", mcpConfig).Debug("Adding MCP config to test command")
		testArgs = append(testArgs, "--mcp-config", mcpConfig)
	}
	testArgs = append(testArgs, "-p", "Say hello")

	logger.Log.WithField("testArgs", testArgs).Debug("Running test command")

	testCmd := exec.CommandContext(testCtx, "claude", testArgs...)
	testCmd.Env = os.Environ()

	// Capture stdout and stderr separately for better debugging
	var testStdout, testStderr bytes.Buffer
	testCmd.Stdout = &testStdout
	testCmd.Stderr = &testStderr

	err = testCmd.Run()

	// Always log any output we got
	if testStderr.Len() > 0 {
		logger.Log.WithField("stderr", testStderr.String()).Warn("Test command stderr")
	}
	if testStdout.Len() > 0 {
		logger.Log.WithField("stdout", testStdout.String()).Info("Test command stdout")
	}

	if testCtx.Err() == context.DeadlineExceeded {
		logger.Log.Error("Claude test command timed out after 10 seconds")
		logger.Log.Warn("Claude may not be working properly - continuing anyway")
	} else if err != nil {
		logger.Log.WithError(err).Error("Claude test command failed")
		logger.Log.Warn("Claude may not be working properly - continuing anyway")
	} else {
		logger.Log.Info("Claude test command succeeded")
	}

	return &Executor{
		tmpDir:    tmpDir,
		awsConfig: awsConfig,
	}, nil
}

func (e *Executor) Execute(prompt string, contextWindow []models.Message) (string, []models.File, error) {
	sessionID := uuid.New().String()
	sessionDir := filepath.Join(e.tmpDir, sessionID)

	log := logger.Log.WithField("sessionID", sessionID)

	log.WithField("sessionDir", sessionDir).Debug("Creating session directory")
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return "", nil, fmt.Errorf("failed to create session directory: %w", err)
	}

	// Verify directory was created
	if stat, err := os.Stat(sessionDir); err != nil {
		log.WithError(err).Error("Failed to stat session directory after creation")
		return "", nil, fmt.Errorf("session directory not found after creation: %w", err)
	} else {
		log.WithFields(map[string]interface{}{
			"sessionDir": sessionDir,
			"isDir":      stat.IsDir(),
			"mode":       stat.Mode(),
		}).Debug("Session directory created successfully")
	}

	defer func() {
		log.WithField("sessionDir", sessionDir).Debug("Cleaning up session directory")
		if err := os.RemoveAll(sessionDir); err != nil {
			log.WithError(err).Warn("Failed to remove session directory")
		}
	}()

	fullPrompt := e.buildPromptWithContext(prompt, contextWindow)

	// Log the command we're about to run
	awsKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	awsKeyPrefix := ""
	if len(awsKeyID) > 10 {
		awsKeyPrefix = awsKeyID[:10] + "..."
	}

	log.WithFields(map[string]interface{}{
		"directory":                sessionDir,
		"CLAUDE_CODE_USE_BEDROCK":  os.Getenv("CLAUDE_CODE_USE_BEDROCK"),
		"ANTHROPIC_MODEL":          os.Getenv("ANTHROPIC_MODEL"),
		"AWS_REGION":               os.Getenv("AWS_REGION"),
		"AWS_ACCESS_KEY_ID_prefix": awsKeyPrefix,
		"AWS_SESSION_TOKEN_exists": os.Getenv("AWS_SESSION_TOKEN") != "",
		"promptLength":             len(fullPrompt),
	}).Info("Executing claude command")

	// Build command args
	args := []string{}
	if os.Getenv("LOG_LEVEL") == "debug" {
		args = append(args, "--debug")
	}
	if model := os.Getenv("ANTHROPIC_MODEL"); model != "" {
		args = append(args, "--model", model)
	}
	if allowedTools := os.Getenv("CLAUDE_ALLOWED_TOOLS"); allowedTools != "" {
		args = append(args, "--allowedTools", allowedTools)
	}
	// Disallowed tools with default if not set
	disallowedTools := os.Getenv("CLAUDE_DISALLOWED_TOOLS")
	if disallowedTools == "" {
		disallowedTools = "Bash,Glob,Grep,LS,Read,Edit,MultiEdit,Write,NotebookRead,NotebookEdit,WebFetch,TodoRead,TodoWrite,Task"
	}
	args = append(args, "--disallowedTools", disallowedTools)
	// MCP config support
	if mcpConfig := os.Getenv("CLAUDE_MCP_CONFIG"); mcpConfig != "" {
		logger.Log.WithField("mcp_config", mcpConfig).Debug("Adding MCP config to command")
		args = append(args, "--mcp-config", mcpConfig)
	}
	args = append(args, "-p", fullPrompt)

	cmd := exec.Command("claude", args...)
	cmd.Dir = sessionDir
	cmd.Env = os.Environ() // Explicitly pass environment

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Log the exact command for debugging
	log.WithFields(map[string]interface{}{
		"command": "claude",
		"args":    args,
		"dir":     sessionDir,
	}).Debug("Executing command")

	// Create a context with timeout - reduced to 30 seconds since claude should respond quickly
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use CommandContext for timeout support
	cmd = exec.CommandContext(ctx, "claude", args...)
	cmd.Dir = sessionDir
	cmd.Env = os.Environ()
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = nil // Explicitly set stdin to nil

	log.Debug("Running claude command with 30 second timeout")

	// Start the command
	if err := cmd.Start(); err != nil {
		log.WithError(err).Error("Failed to start claude command")
		return "", nil, fmt.Errorf("failed to start claude: %w", err)
	}

	// Wait for completion or timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	var err error
	select {
	case <-ctx.Done():
		// Timeout occurred
		log.Error("Claude command timed out after 30 seconds")
		if err := cmd.Process.Kill(); err != nil {
			log.WithError(err).Warn("Failed to kill claude process")
		}
		// Get any partial output
		if stderr.Len() > 0 {
			log.WithField("stderr", stderr.String()).Error("Claude stderr before timeout")
		}
		if stdout.Len() > 0 {
			log.WithField("stdout", stdout.String()).Error("Claude stdout before timeout")
		}
		return "", nil, fmt.Errorf("claude command timed out")
	case cmdErr := <-done:
		// Command completed
		err = cmdErr
		if err != nil {
			log.WithError(err).Debug("Claude command completed with error")
		}
		// Continue to process output below
	}

	// Log output at appropriate levels
	if stderr.Len() > 0 {
		log.WithField("stderr", stderr.String()).Warn("Claude stderr output")
	}
	if stdout.Len() > 0 {
		log.WithField("outputLength", stdout.Len()).Debug("Claude stdout received")
		log.Debugf("Claude output preview: %s...", stdout.String()[:min(200, stdout.Len())])
	}

	if err != nil {
		log.WithError(err).WithField("stderr", stderr.String()).Error("Claude execution failed")
		return "", nil, fmt.Errorf("claude execution failed: %w, stderr: %s", err, stderr.String())
	}

	log.Debug("Scanning for output files")
	files, err := e.scanForFiles(sessionDir)
	if err != nil {
		log.WithError(err).Warn("Failed to scan for files")
		return stdout.String(), nil, err
	}

	log.WithField("fileCount", len(files)).Info("Claude execution completed successfully")

	return stdout.String(), files, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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

