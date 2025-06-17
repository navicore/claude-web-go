package auth

import (
	"fmt"
	"os"
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

	os.Setenv("CLAUDE_CODE_USE_BEDROCK", "true")
	
	// Set model if not already set
	if os.Getenv("ANTHROPIC_MODEL") == "" {
		os.Setenv("ANTHROPIC_MODEL", "us.anthropic.claude-3-5-sonnet-20241022-v2:0")
	}
	
	// Set small/fast model if not already set
	if os.Getenv("ANTHROPIC_SMALL_FAST_MODEL") == "" {
		os.Setenv("ANTHROPIC_SMALL_FAST_MODEL", "us.anthropic.claude-3-5-haiku-20241022-v1:0")
	}

	return nil
}