package auth

import (
	"context"
	"fmt"
	"sync"
	"time"

	"claude-web-go/internal/logger"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type SessionManager struct {
	baseAccessKey    string
	baseSecretKey    string
	region           string
	currentSession   *SessionCredentials
	mu               sync.RWMutex
}

type SessionCredentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	Expiration      time.Time
}

func NewSessionManager(accessKey, secretKey, region string) *SessionManager {
	return &SessionManager{
		baseAccessKey: accessKey,
		baseSecretKey: secretKey,
		region:        region,
	}
}

func (sm *SessionManager) GetSessionCredentials() (*SessionCredentials, error) {
	sm.mu.RLock()
	if sm.currentSession != nil {
		timeUntilExpiry := time.Until(sm.currentSession.Expiration)
		logger.Log.WithFields(map[string]interface{}{
			"expiresIn": timeUntilExpiry.String(),
			"expiration": sm.currentSession.Expiration.Format(time.RFC3339),
			"currentTime": time.Now().Format(time.RFC3339),
		}).Debug("Checking cached session token")
		
		if timeUntilExpiry > 5*time.Minute {
			defer sm.mu.RUnlock()
			logger.Log.Info("Using cached session token")
			return sm.currentSession, nil
		}
		logger.Log.Info("Session token expired or expiring soon, will refresh")
	}
	sm.mu.RUnlock()

	// Need to refresh
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Double-check after acquiring write lock
	if sm.currentSession != nil && time.Until(sm.currentSession.Expiration) > 5*time.Minute {
		return sm.currentSession, nil
	}

	logger.Log.WithField("accessKeyPrefix", sm.baseAccessKey[:min(10, len(sm.baseAccessKey))]).Info("Getting new session token")

	// Create STS client with base credentials
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(sm.region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			sm.baseAccessKey,
			sm.baseSecretKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS config: %w", err)
	}

	stsClient := sts.NewFromConfig(cfg)

	// Get session token (valid for 12 hours by default)
	logger.Log.Debug("Calling AWS STS GetSessionToken")
	result, err := stsClient.GetSessionToken(context.TODO(), &sts.GetSessionTokenInput{
		DurationSeconds: aws.Int32(43200), // 12 hours
	})
	if err != nil {
		logger.Log.WithError(err).Error("Failed to get session token")
		return nil, fmt.Errorf("failed to get session token: %w", err)
	}
	
	sm.currentSession = &SessionCredentials{
		AccessKeyID:     *result.Credentials.AccessKeyId,
		SecretAccessKey: *result.Credentials.SecretAccessKey,
		SessionToken:    *result.Credentials.SessionToken,
		Expiration:      *result.Credentials.Expiration,
	}
	
	logger.Log.WithFields(map[string]interface{}{
		"expiration": sm.currentSession.Expiration.Format(time.RFC3339),
		"validFor": time.Until(sm.currentSession.Expiration).String(),
		"accessKeyPrefix": sm.currentSession.AccessKeyID[:10] + "...",
	}).Info("Successfully obtained new session token")

	return sm.currentSession, nil
}

