package enterprise

import (
	"context"
	"fmt"
	"strings"
	"time"

	"panoptic/internal/logger"
)

// APIManagement handles API key operations
type APIManagement struct {
	Manager *EnterpriseManager
	Logger   logger.Logger
}

// NewAPIManagement creates new API management handler
func NewAPIManagement(manager *EnterpriseManager) *APIManagement {
	return &APIManagement{
		Manager: manager,
		Logger:  manager.Logger,
	}
}

// CreateAPIKey creates a new API key
func (am *APIManagement) CreateAPIKey(ctx context.Context, req CreateAPIKeyRequest) (*APIKey, error) {
	// Validate request
	if err := am.validateCreateAPIKeyRequest(ctx, req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Check API key limits
	if am.Manager.APIKeysExceedLimit() {
		return nil, fmt.Errorf("maximum number of API keys reached")
	}

	// Generate API key and secret
	key := am.generateAPIKey()
	secret := am.generateAPISecret()

	// Create API key
	apiKey := &APIKey{
		ID:          am.Manager.generateID(),
		UserID:      req.UserID,
		Name:        req.Name,
		Key:         key,
		Secret:      secret,
		Permissions: req.Permissions,
		Scopes:      req.Scopes,
		RateLimit:   req.RateLimit,
		Enabled:     req.Enabled,
		CreatedAt:   time.Now(),
		UsageCount:  0,
		Metadata:    req.Metadata,
	}

	if req.ExpiresInHours > 0 {
		expiresAt := time.Now().Add(time.Duration(req.ExpiresInHours) * time.Hour)
		apiKey.ExpiresAt = &expiresAt
	}

	// Store API key
	am.Manager.APIKeys[apiKey.ID] = apiKey

	// Update user's API keys
	am.Manager.updateUserAPIKeys(req.UserID, apiKey.ID)

	// Log audit entry
	am.Manager.logAuditEntry(AuditEntry{
		ID:         am.Manager.generateID(),
		Timestamp:  time.Now(),
		UserID:     req.UserID,
		Username:   am.getUsername(req.UserID),
		Action:     "api_key.create",
		Resource:   "api_key",
		ResourceID: apiKey.ID,
		Details:    map[string]string{"name": req.Name, "key": key},
		Success:    true,
		Severity:   "medium",
		Category:   "access",
	})

	// Save data
	if err := am.Manager.saveData(); err != nil {
		am.Logger.Errorf("Failed to save API key data: %v", err)
	}

	am.Logger.Infof("API key created successfully: %s (user: %s)", req.Name, am.getUsername(req.UserID))
	return apiKey, nil
}

// GetAPIKey retrieves an API key by ID
func (am *APIManagement) GetAPIKey(ctx context.Context, keyID string) (*APIKey, error) {
	apiKey, exists := am.Manager.APIKeys[keyID]
	if !exists {
		return nil, fmt.Errorf("API key not found: %s", keyID)
	}

	return apiKey, nil
}

// GetAPIKeyByKey retrieves an API key by the actual key value
func (am *APIManagement) GetAPIKeyByKey(ctx context.Context, key string) (*APIKey, error) {
	for _, apiKey := range am.Manager.APIKeys {
		if apiKey.Key == key {
			return apiKey, nil
		}
	}
	return nil, fmt.Errorf("API key not found: %s", key)
}

// UpdateAPIKey updates an existing API key
func (am *APIManagement) UpdateAPIKey(ctx context.Context, keyID string, req UpdateAPIKeyRequest) (*APIKey, error) {
	apiKey, err := am.GetAPIKey(ctx, keyID)
	if err != nil {
		return nil, err
	}

	// Check permissions
	if !am.hasAPIKeyPermission(ctx, apiKey, "api_key.update") {
		return nil, fmt.Errorf("insufficient permissions to update API key")
	}

	// Update fields
	if req.Name != "" {
		apiKey.Name = req.Name
	}
	if req.Permissions != nil {
		apiKey.Permissions = req.Permissions
	}
	if req.Scopes != nil {
		apiKey.Scopes = req.Scopes
	}
	if req.RateLimit > 0 {
		apiKey.RateLimit = req.RateLimit
	}
	if req.Enabled {
		apiKey.Enabled = req.Enabled
	}
	if req.ExpiresInHours > 0 {
		expiresAt := time.Now().Add(time.Duration(req.ExpiresInHours) * time.Hour)
		apiKey.ExpiresAt = &expiresAt
	} else if req.ExpiresInHours == 0 {
		apiKey.ExpiresAt = nil
	}
	if req.Metadata != nil {
		apiKey.Metadata = req.Metadata
	}

	// Log audit entry
	am.Manager.logAuditEntry(AuditEntry{
		ID:         am.Manager.generateID(),
		Timestamp:  time.Now(),
		UserID:     apiKey.UserID,
		Username:   am.getUsername(apiKey.UserID),
		Action:     "api_key.update",
		Resource:   "api_key",
		ResourceID: apiKey.ID,
		Details:    map[string]string{"name": apiKey.Name},
		Success:    true,
		Severity:   "medium",
		Category:   "access",
	})

	// Save data
	if err := am.Manager.saveData(); err != nil {
		am.Logger.Errorf("Failed to save API key data: %v", err)
	}

	am.Logger.Infof("API key updated successfully: %s", apiKey.Name)
	return apiKey, nil
}

// DeleteAPIKey deletes an API key
func (am *APIManagement) DeleteAPIKey(ctx context.Context, keyID string) error {
	apiKey, err := am.GetAPIKey(ctx, keyID)
	if err != nil {
		return err
	}

	// Check permissions
	if !am.hasAPIKeyPermission(ctx, apiKey, "api_key.delete") {
		return fmt.Errorf("insufficient permissions to delete API key")
	}

	// Delete API key
	delete(am.Manager.APIKeys, keyID)

	// Update user's API keys
	am.Manager.removeUserAPIKey(apiKey.UserID, keyID)

	// Log audit entry
	am.Manager.logAuditEntry(AuditEntry{
		ID:         am.Manager.generateID(),
		Timestamp:  time.Now(),
		UserID:     apiKey.UserID,
		Username:   am.getUsername(apiKey.UserID),
		Action:     "api_key.delete",
		Resource:   "api_key",
		ResourceID: apiKey.ID,
		Details:    map[string]string{"name": apiKey.Name},
		Success:    true,
		Severity:   "high",
		Category:   "access",
	})

	// Save data
	if err := am.Manager.saveData(); err != nil {
		am.Logger.Errorf("Failed to save API key data: %v", err)
	}

	am.Logger.Infof("API key deleted successfully: %s", apiKey.Name)
	return nil
}

// RegenerateAPIKeySecret regenerates an API key secret
func (am *APIManagement) RegenerateAPIKeySecret(ctx context.Context, keyID string) (*APIKey, error) {
	apiKey, err := am.GetAPIKey(ctx, keyID)
	if err != nil {
		return nil, err
	}

	// Check permissions
	if !am.hasAPIKeyPermission(ctx, apiKey, "api_key.update") {
		return nil, fmt.Errorf("insufficient permissions to update API key")
	}

	// Generate new secret
	newSecret := am.generateAPISecret()
	apiKey.Secret = newSecret

	// Log audit entry (without logging the secret)
	am.Manager.logAuditEntry(AuditEntry{
		ID:         am.Manager.generateID(),
		Timestamp:  time.Now(),
		UserID:     apiKey.UserID,
		Username:   am.getUsername(apiKey.UserID),
		Action:     "api_key.regenerate_secret",
		Resource:   "api_key",
		ResourceID: apiKey.ID,
		Details:    map[string]string{"name": apiKey.Name},
		Success:    true,
		Severity:   "medium",
		Category:   "access",
	})

	// Save data
	if err := am.Manager.saveData(); err != nil {
		am.Logger.Errorf("Failed to save API key data: %v", err)
	}

	am.Logger.Infof("API key secret regenerated successfully: %s", apiKey.Name)
	return apiKey, nil
}

// ValidateAPIKey validates an API key for authentication
func (am *APIManagement) ValidateAPIKey(ctx context.Context, key, secret string) (*APIKey, error) {
	apiKey, err := am.GetAPIKeyByKey(ctx, key)
	if err != nil {
		// Log failed authentication
		am.Manager.logAuditEntry(AuditEntry{
			ID:        am.Manager.generateID(),
			Timestamp: time.Now(),
			Action:    "api_key.validate",
			Resource:  "api_key",
			Details:   map[string]string{"reason": "key_not_found"},
			Success:   false,
			Severity:  "medium",
			Category:  "access",
		})
		return nil, fmt.Errorf("invalid API key")
	}

	if !apiKey.Enabled {
		am.Manager.logAuditEntry(AuditEntry{
			ID:        am.Manager.generateID(),
			Timestamp: time.Now(),
			UserID:    apiKey.UserID,
			Username:  am.getUsername(apiKey.UserID),
			Action:    "api_key.validate",
			Resource:  "api_key",
			ResourceID: apiKey.ID,
			Details:   map[string]string{"reason": "key_disabled"},
			Success:   false,
			Severity:  "medium",
			Category:  "access",
		})
		return nil, fmt.Errorf("API key is disabled")
	}

	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		am.Manager.logAuditEntry(AuditEntry{
			ID:        am.Manager.generateID(),
			Timestamp: time.Now(),
			UserID:    apiKey.UserID,
			Username:  am.getUsername(apiKey.UserID),
			Action:    "api_key.validate",
			Resource:  "api_key",
			ResourceID: apiKey.ID,
			Details:   map[string]string{"reason": "key_expired"},
			Success:   false,
			Severity:  "medium",
			Category:  "access",
		})
		return nil, fmt.Errorf("API key has expired")
	}

	if apiKey.Secret != secret {
		am.Manager.logAuditEntry(AuditEntry{
			ID:        am.Manager.generateID(),
			Timestamp: time.Now(),
			UserID:    apiKey.UserID,
			Username:  am.getUsername(apiKey.UserID),
			Action:    "api_key.validate",
			Resource:  "api_key",
			ResourceID: apiKey.ID,
			Details:   map[string]string{"reason": "invalid_secret"},
			Success:   false,
			Severity:  "medium",
			Category:  "access",
		})
		return nil, fmt.Errorf("invalid API key secret")
	}

	// Update usage statistics
	now := time.Now()
	apiKey.LastUsed = &now
	apiKey.UsageCount++

	// Log successful authentication
	am.Manager.logAuditEntry(AuditEntry{
		ID:        am.Manager.generateID(),
		Timestamp: now,
		UserID:    apiKey.UserID,
		Username:  am.getUsername(apiKey.UserID),
		Action:    "api_key.validate",
		Resource:  "api_key",
		ResourceID: apiKey.ID,
		Details:   map[string]string{"usage_count": fmt.Sprintf("%d", apiKey.UsageCount)},
		Success:   true,
		Severity:  "low",
		Category:  "access",
	})

	return apiKey, nil
}

// ListAPIKeys lists all API keys with filtering
func (am *APIManagement) ListAPIKeys(ctx context.Context, req ListAPIKeysRequest) (*ListAPIKeysResponse, error) {
	var apiKeys []APIKey
	for _, apiKey := range am.Manager.APIKeys {
		// Apply filters
		if req.UserID != "" && apiKey.UserID != req.UserID {
			continue
		}
		if req.Enabled != nil && *req.Enabled != apiKey.Enabled {
			continue
		}
		if req.Permission != "" && !contains(apiKey.Permissions, req.Permission) {
			continue
		}
		if req.Scope != "" && !contains(apiKey.Scopes, req.Scope) {
			continue
		}

		// Remove sensitive data
		safeAPIKey := *apiKey
		safeAPIKey.Secret = "*****"
		apiKeys = append(apiKeys, safeAPIKey)
	}

	// Apply pagination
	total := len(apiKeys)
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize
	if end > total {
		end = total
	}
	if start >= total {
		return &ListAPIKeysResponse{
			APIKeys: []APIKey{},
			Total:   total,
			Page:    req.Page,
			PageSize: req.PageSize,
		}, nil
	}

	pagedAPIKeys := apiKeys[start:end]

	return &ListAPIKeysResponse{
		APIKeys: pagedAPIKeys,
		Total:   total,
		Page:    req.Page,
		PageSize: req.PageSize,
	}, nil
}

// CheckAPIKeyRateLimit checks if an API key has exceeded its rate limit
func (am *APIManagement) CheckAPIKeyRateLimit(ctx context.Context, keyID string, windowMinutes int) (bool, time.Duration, error) {
	apiKey, err := am.GetAPIKey(ctx, keyID)
	if err != nil {
		return false, 0, err
	}

	// For demo purposes, implement simple in-memory rate limiting
	// In a real implementation, use Redis or similar for distributed rate limiting
	if apiKey.RateLimit <= 0 {
		return true, 0, nil // No rate limit
	}

	// Check recent usage (mock implementation)
	// In a real implementation, track usage in a sliding window
	if apiKey.UsageCount > apiKey.RateLimit {
		return false, 1 * time.Hour, nil // Mock cooldown
	}

	return true, 0, nil
}

// GetAPIKeyUsage retrieves API key usage statistics
func (am *APIManagement) GetAPIKeyUsage(ctx context.Context, keyID string, days int) (*APIKeyUsageResponse, error) {
	apiKey, err := am.GetAPIKey(ctx, keyID)
	if err != nil {
		return nil, err
	}

	// Mock usage statistics for demo
	response := &APIKeyUsageResponse{
		KeyID:      keyID,
		KeyName:    apiKey.Name,
		TotalUsage: apiKey.UsageCount,
		DailyUsage: make(map[string]int),
		HourlyUsage: make(map[string]int),
		Endpoints:  make(map[string]int),
		TopEndpoints: []EndpointUsage{
			{Name: "GET /api/v1/tests", Count: 150},
			{Name: "POST /api/v1/tests", Count: 100},
			{Name: "GET /api/v1/results", Count: 75},
		},
		LastUsed:   apiKey.LastUsed,
		CreatedAt:  apiKey.CreatedAt,
	}

	// Generate mock daily usage for the requested period
	endDate := time.Now()
	for i := 0; i < days; i++ {
		date := endDate.AddDate(0, 0, -i).Format("2006-01-02")
		response.DailyUsage[date] = apiKey.UsageCount / days // Even distribution
	}

	// Generate mock hourly usage for today
	for hour := 0; hour < 24; hour++ {
		hourStr := fmt.Sprintf("%02d", hour)
		response.HourlyUsage[hourStr] = apiKey.UsageCount / 24 // Even distribution
	}

	return response, nil
}

// Request types

type CreateAPIKeyRequest struct {
	UserID          string            `json:"user_id"`
	Name            string            `json:"name"`
	Permissions     []string          `json:"permissions"`
	Scopes          []string          `json:"scopes"`
	RateLimit       int               `json:"rate_limit"`
	Enabled         bool              `json:"enabled"`
	ExpiresInHours  int               `json:"expires_in_hours"`
	Metadata        map[string]string `json:"metadata"`
}

type UpdateAPIKeyRequest struct {
	Name           string            `json:"name,omitempty"`
	Permissions    []string          `json:"permissions,omitempty"`
	Scopes         []string          `json:"scopes,omitempty"`
	RateLimit      int               `json:"rate_limit,omitempty"`
	Enabled        bool              `json:"enabled,omitempty"`
	ExpiresInHours int               `json:"expires_in_hours,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

type ListAPIKeysRequest struct {
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	UserID     string `json:"user_id,omitempty"`
	Enabled    *bool  `json:"enabled,omitempty"`
	Permission string `json:"permission,omitempty"`
	Scope      string `json:"scope,omitempty"`
}

type ListAPIKeysResponse struct {
	APIKeys []APIKey `json:"api_keys"`
	Total   int       `json:"total"`
	Page    int       `json:"page"`
	PageSize int       `json:"page_size"`
}

type APIKeyUsageResponse struct {
	KeyID       string          `json:"key_id"`
	KeyName     string          `json:"key_name"`
	TotalUsage  int             `json:"total_usage"`
	DailyUsage  map[string]int  `json:"daily_usage"`
	HourlyUsage map[string]int  `json:"hourly_usage"`
	Endpoints   map[string]int  `json:"endpoints"`
	TopEndpoints []EndpointUsage `json:"top_endpoints"`
	LastUsed    *time.Time      `json:"last_used"`
	CreatedAt   time.Time       `json:"created_at"`
}

type EndpointUsage struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// Helper methods

func (am *APIManagement) validateCreateAPIKeyRequest(ctx context.Context, req CreateAPIKeyRequest) error {
	if req.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	if req.Name == "" {
		return fmt.Errorf("API key name is required")
	}
	if len(req.Permissions) == 0 {
		return fmt.Errorf("at least one permission is required")
	}
	if req.RateLimit < 0 {
		return fmt.Errorf("rate limit must be non-negative")
	}
	return nil
}

func (am *APIManagement) hasAPIKeyPermission(ctx context.Context, apiKey *APIKey, permission string) bool {
	// Check if user has permission to manage this API key
	return contains(apiKey.Permissions, permission)
}

func (am *APIManagement) generateAPIKey() string {
	return strings.ToUpper(fmt.Sprintf("pk_%s", am.Manager.generateID()[:16]))
}

func (am *APIManagement) generateAPISecret() string {
	return strings.ToUpper(fmt.Sprintf("sk_%s", am.Manager.generateID()[:32]))
}

func (am *APIManagement) getUsername(userID string) string {
	for _, user := range am.Manager.Users {
		if user.ID == userID {
			return user.Username
		}
	}
	return "unknown"
}

func (em *EnterpriseManager) updateUserAPIKeys(userID, keyID string) {
	if user, exists := em.Users[userID]; exists {
		user.APIKeys = appendUnique(user.APIKeys, keyID)
		user.UpdatedAt = time.Now()
	}
}

func (em *EnterpriseManager) removeUserAPIKey(userID, keyID string) {
	if user, exists := em.Users[userID]; exists {
		user.APIKeys = remove(user.APIKeys, keyID)
		user.UpdatedAt = time.Now()
	}
}

func (em *EnterpriseManager) APIKeysExceedLimit() bool {
	if em.Config.MaxAPIKeys <= 0 {
		return false
	}

	return len(em.APIKeys) >= em.Config.MaxAPIKeys
}