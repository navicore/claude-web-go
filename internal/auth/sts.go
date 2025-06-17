package auth

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type STSCredentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Expiration      time.Time
}

// GetSTSCredentials uses AWS CLI to get temporary credentials
// This works with IAM roles, SSO, or any AWS CLI authentication method
func GetSTSCredentials() (*STSCredentials, error) {
	// Use aws sts get-session-token for regular IAM users
	// For assumed roles or SSO, use aws sts get-caller-identity first
	cmd := exec.Command("aws", "sts", "get-caller-identity")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to verify AWS credentials: %w", err)
	}

	// If get-caller-identity works, we have valid credentials
	// Now check if we need session token (for MFA or assumed role)
	var callerIdentity map[string]interface{}
	if err := json.Unmarshal(output, &callerIdentity); err != nil {
		return nil, fmt.Errorf("failed to parse caller identity: %w", err)
	}

	// Try to get credentials from the credential process
	cmd = exec.Command("aws", "configure", "export-credentials")
	output, err = cmd.Output()
	if err != nil {
		// Fallback to environment variables if export-credentials fails
		return nil, fmt.Errorf("failed to export credentials: %w", err)
	}

	var exportedCreds struct {
		AccessKeyID     string `json:"AccessKeyId"`
		SecretAccessKey string `json:"SecretAccessKey"`
		SessionToken    string `json:"SessionToken"`
		Expiration      string `json:"Expiration"`
	}

	if err := json.Unmarshal(output, &exportedCreds); err != nil {
		return nil, fmt.Errorf("failed to parse exported credentials: %w", err)
	}

	creds := &STSCredentials{
		AccessKeyID:     exportedCreds.AccessKeyID,
		SecretAccessKey: exportedCreds.SecretAccessKey,
		SessionToken:    exportedCreds.SessionToken,
	}

	if exportedCreds.Expiration != "" {
		if exp, err := time.Parse(time.RFC3339, exportedCreds.Expiration); err == nil {
			creds.Expiration = exp
		}
	}

	return creds, nil
}

// RefreshSTSCredentials checks if credentials are expired and refreshes them
func RefreshSTSCredentials(current *STSCredentials) (*STSCredentials, error) {
	// If no expiration or still valid for more than 5 minutes, keep current
	if current.Expiration.IsZero() || time.Until(current.Expiration) > 5*time.Minute {
		return current, nil
	}

	// Otherwise, get new credentials
	return GetSTSCredentials()
}