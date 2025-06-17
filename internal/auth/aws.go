package auth

import (
	"fmt"
	"os"
	
	"claude-web-go/internal/logger"
)

type AWSConfig struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Region          string
}

var sessionManager *SessionManager

func GetAWSConfig() (*AWSConfig, error) {
	baseAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	baseSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	region := os.Getenv("AWS_REGION")

	if region == "" {
		region = "us-east-1"
	}

	if baseAccessKey == "" || baseSecretKey == "" {
		return nil, fmt.Errorf("AWS credentials not found in environment variables")
	}

	// Initialize session manager if not already done
	if sessionManager == nil {
		sessionManager = NewSessionManager(baseAccessKey, baseSecretKey, region)
	}

	// Get session credentials (will refresh if needed)
	sessionCreds, err := sessionManager.GetSessionCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to get session credentials: %w", err)
	}

	return &AWSConfig{
		AccessKeyID:     sessionCreds.AccessKeyID,
		SecretAccessKey: sessionCreds.SecretAccessKey,
		SessionToken:    sessionCreds.SessionToken,
		Region:          region,
	}, nil
}

func SetupEnvironment(config *AWSConfig) error {
	env := map[string]string{
		"AWS_ACCESS_KEY_ID":     config.AccessKeyID,
		"AWS_SECRET_ACCESS_KEY": config.SecretAccessKey,
		"AWS_REGION":            config.Region,
	}

	if config.SessionToken != "" {
		env["AWS_SESSION_TOKEN"] = config.SessionToken
	}

	for key, value := range env {
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("failed to set environment variable %s: %w", key, err)
		}
	}

	os.Setenv("CLAUDE_CODE_USE_BEDROCK", "1")
	
	// Make sure ANTHROPIC_API_KEY is not set, as it might conflict
	os.Unsetenv("ANTHROPIC_API_KEY")
	
	// Set model if not already set
	if os.Getenv("ANTHROPIC_MODEL") == "" {
		os.Setenv("ANTHROPIC_MODEL", "us.anthropic.claude-sonnet-4-20250514-v1:0")
	}
	
	// Set small/fast model if not already set
	if os.Getenv("ANTHROPIC_SMALL_FAST_MODEL") == "" {
		os.Setenv("ANTHROPIC_SMALL_FAST_MODEL", "anthropic.claude-3-5-haiku-20241022-v1:0")
	}

	// Log the configuration for debugging
	logger.Log.WithFields(map[string]interface{}{
		"AWS_REGION": config.Region,
		"AWS_ACCESS_KEY_ID_prefix": config.AccessKeyID[:min(10, len(config.AccessKeyID))],
		"AWS_SESSION_TOKEN_exists": config.SessionToken != "",
		"AWS_SESSION_TOKEN_prefix": getTokenPrefix(config.SessionToken),
		"ANTHROPIC_MODEL": os.Getenv("ANTHROPIC_MODEL"),
		"CLAUDE_CODE_USE_BEDROCK": os.Getenv("CLAUDE_CODE_USE_BEDROCK"),
	}).Info("AWS environment configured")

	return nil
}

func getTokenPrefix(token string) string {
	if token == "" {
		return "none"
	}
	if len(token) > 20 {
		return token[:20] + "..."
	}
	return token
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}