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

func GetAWSConfig() (*AWSConfig, error) {
	config := &AWSConfig{
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		SessionToken:    os.Getenv("AWS_SESSION_TOKEN"),
		Region:          os.Getenv("AWS_REGION"),
	}

	if config.Region == "" {
		config.Region = "us-east-1"
	}

	if config.AccessKeyID == "" || config.SecretAccessKey == "" {
		return nil, fmt.Errorf("AWS credentials not found in environment variables")
	}

	return config, nil
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
		os.Setenv("ANTHROPIC_MODEL", "claude-3-5-sonnet-20241022")
	}

	return nil
}