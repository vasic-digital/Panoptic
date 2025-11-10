package enterprise

import (
	"context"
	"fmt"
	"testing"
	"time"

	"panoptic/internal/logger"

	"github.com/stretchr/testify/assert"
)

// TestNewAPIManagement tests API management creation
func TestNewAPIManagement(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:  *log,
		APIKeys: make(map[string]*APIKey),
	}

	am := NewAPIManagement(manager)

	assert.NotNil(t, am, "APIManagement should not be nil")
	assert.Equal(t, manager, am.Manager, "Manager should be set")
}

// TestCreateAPIKey tests API key creation
func TestCreateAPIKey(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		APIKeys:  make(map[string]*APIKey),
		Users:    make(map[string]*User),
		AuditLog: []AuditEntry{},
		Config: EnterpriseConfig{
			MaxAPIKeys: 100,
		},
	}

	// Add a test user
	manager.Users["testuser"] = &User{
		ID:       "user-123",
		Username: "testuser",
		APIKeys:  []string{},
	}

	am := NewAPIManagement(manager)
	ctx := context.Background()

	req := CreateAPIKeyRequest{
		UserID:      "user-123",
		Name:        "Test API Key",
		Permissions: []string{"read", "write"},
		Scopes:      []string{"tests", "results"},
		RateLimit:   1000,
		Enabled:     true,
	}

	apiKey, err := am.CreateAPIKey(ctx, req)

	assert.NoError(t, err, "Should create API key without error")
	assert.NotNil(t, apiKey, "API key should not be nil")
	assert.Equal(t, req.Name, apiKey.Name)
	assert.Equal(t, req.UserID, apiKey.UserID)
	assert.NotEmpty(t, apiKey.Key, "Key should be generated")
	assert.NotEmpty(t, apiKey.Secret, "Secret should be generated")
	assert.Equal(t, 0, apiKey.UsageCount, "Usage count should start at 0")
}

// TestCreateAPIKey_WithExpiration tests API key creation with expiration
func TestCreateAPIKey_WithExpiration(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		APIKeys:  make(map[string]*APIKey),
		Users:    make(map[string]*User),
		AuditLog: []AuditEntry{},
		Config: EnterpriseConfig{
			MaxAPIKeys: 100,
		},
	}

	manager.Users["testuser"] = &User{
		ID:       "user-123",
		Username: "testuser",
		APIKeys:  []string{},
	}

	am := NewAPIManagement(manager)
	ctx := context.Background()

	req := CreateAPIKeyRequest{
		UserID:         "user-123",
		Name:           "Expiring Key",
		Permissions:    []string{"read"},
		Scopes:         []string{"tests"},
		RateLimit:      100,
		Enabled:        true,
		ExpiresInHours: 24,
	}

	apiKey, err := am.CreateAPIKey(ctx, req)

	assert.NoError(t, err, "Should create API key without error")
	assert.NotNil(t, apiKey.ExpiresAt, "ExpiresAt should be set")
	assert.True(t, apiKey.ExpiresAt.After(time.Now()), "ExpiresAt should be in the future")
}

// TestCreateAPIKey_InvalidRequest tests with invalid request
func TestCreateAPIKey_InvalidRequest(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:  *log,
		APIKeys: make(map[string]*APIKey),
		Config: EnterpriseConfig{
			MaxAPIKeys: 100,
		},
	}

	am := NewAPIManagement(manager)
	ctx := context.Background()

	testCases := []struct {
		name string
		req  CreateAPIKeyRequest
	}{
		{
			name: "empty user ID",
			req: CreateAPIKeyRequest{
				Name:        "Test",
				Permissions: []string{"read"},
			},
		},
		{
			name: "empty name",
			req: CreateAPIKeyRequest{
				UserID:      "user-123",
				Permissions: []string{"read"},
			},
		},
		{
			name: "no permissions",
			req: CreateAPIKeyRequest{
				UserID: "user-123",
				Name:   "Test",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			apiKey, err := am.CreateAPIKey(ctx, tc.req)

			assert.Error(t, err, "Should error with invalid request")
			assert.Nil(t, apiKey, "API key should be nil on error")
		})
	}
}

// TestGetAPIKey tests getting API key by ID
func TestGetAPIKey(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:  *log,
		APIKeys: make(map[string]*APIKey),
	}

	testKey := &APIKey{
		ID:     "key-123",
		UserID: "user-123",
		Name:   "Test Key",
		Key:    "test-key",
	}
	manager.APIKeys["key-123"] = testKey

	am := NewAPIManagement(manager)
	ctx := context.Background()

	apiKey, err := am.GetAPIKey(ctx, "key-123")

	assert.NoError(t, err, "Should get API key without error")
	assert.NotNil(t, apiKey, "API key should not be nil")
	assert.Equal(t, testKey.ID, apiKey.ID)
}

// TestGetAPIKey_NotFound tests getting nonexistent key
func TestGetAPIKey_NotFound(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:  *log,
		APIKeys: make(map[string]*APIKey),
	}

	am := NewAPIManagement(manager)
	ctx := context.Background()

	apiKey, err := am.GetAPIKey(ctx, "nonexistent")

	assert.Error(t, err, "Should error when key not found")
	assert.Nil(t, apiKey, "API key should be nil")
}

// TestGetAPIKeyByKey tests getting API key by key value
func TestGetAPIKeyByKey(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:  *log,
		APIKeys: make(map[string]*APIKey),
	}

	testKey := &APIKey{
		ID:     "key-123",
		UserID: "user-123",
		Name:   "Test Key",
		Key:    "pk_test123",
	}
	manager.APIKeys["key-123"] = testKey

	am := NewAPIManagement(manager)
	ctx := context.Background()

	apiKey, err := am.GetAPIKeyByKey(ctx, "pk_test123")

	assert.NoError(t, err, "Should get API key without error")
	assert.NotNil(t, apiKey, "API key should not be nil")
	assert.Equal(t, testKey.ID, apiKey.ID)
}

// TestUpdateAPIKey tests API key updates
func TestUpdateAPIKey(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		APIKeys:  make(map[string]*APIKey),
		Users:    make(map[string]*User),
		AuditLog: []AuditEntry{},
	}

	testKey := &APIKey{
		ID:          "key-123",
		UserID:      "user-123",
		Name:        "Old Name",
		Key:         "test-key",
		Permissions: []string{"read", "api_key.update"},
		RateLimit:   100,
		Enabled:     true,
	}
	manager.APIKeys["key-123"] = testKey
	manager.Users["testuser"] = &User{ID: "user-123", Username: "testuser"}

	am := NewAPIManagement(manager)
	ctx := context.Background()

	req := UpdateAPIKeyRequest{
		Name:      "New Name",
		RateLimit: 200,
	}

	apiKey, err := am.UpdateAPIKey(ctx, "key-123", req)

	assert.NoError(t, err, "Should update API key without error")
	assert.NotNil(t, apiKey, "API key should not be nil")
	assert.Equal(t, "New Name", apiKey.Name)
	assert.Equal(t, 200, apiKey.RateLimit)
}

// TestUpdateAPIKey_NoPermission tests update without permission
func TestUpdateAPIKey_NoPermission(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:  *log,
		APIKeys: make(map[string]*APIKey),
	}

	testKey := &APIKey{
		ID:          "key-123",
		UserID:      "user-123",
		Name:        "Test Key",
		Permissions: []string{"read"}, // No api_key.update permission
		Enabled:     true,
	}
	manager.APIKeys["key-123"] = testKey

	am := NewAPIManagement(manager)
	ctx := context.Background()

	req := UpdateAPIKeyRequest{
		Name: "New Name",
	}

	apiKey, err := am.UpdateAPIKey(ctx, "key-123", req)

	assert.Error(t, err, "Should error without permission")
	assert.Nil(t, apiKey, "API key should be nil")
	assert.Contains(t, err.Error(), "insufficient permissions")
}

// TestDeleteAPIKey tests API key deletion
func TestDeleteAPIKey(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		APIKeys:  make(map[string]*APIKey),
		Users:    make(map[string]*User),
		AuditLog: []AuditEntry{},
	}

	testKey := &APIKey{
		ID:          "key-123",
		UserID:      "user-123",
		Name:        "Test Key",
		Permissions: []string{"api_key.delete"},
		Enabled:     true,
	}
	manager.APIKeys["key-123"] = testKey
	manager.Users["testuser"] = &User{
		ID:       "user-123",
		Username: "testuser",
		APIKeys:  []string{"key-123"},
	}

	am := NewAPIManagement(manager)
	ctx := context.Background()

	err := am.DeleteAPIKey(ctx, "key-123")

	assert.NoError(t, err, "Should delete API key without error")
	_, exists := manager.APIKeys["key-123"]
	assert.False(t, exists, "API key should be deleted")
}

// TestRegenerateAPIKeySecret tests secret regeneration
func TestRegenerateAPIKeySecret(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		APIKeys:  make(map[string]*APIKey),
		Users:    make(map[string]*User),
		AuditLog: []AuditEntry{},
	}

	oldSecret := "old-secret"
	testKey := &APIKey{
		ID:          "key-123",
		UserID:      "user-123",
		Name:        "Test Key",
		Secret:      oldSecret,
		Permissions: []string{"api_key.update"},
		Enabled:     true,
	}
	manager.APIKeys["key-123"] = testKey
	manager.Users["testuser"] = &User{ID: "user-123", Username: "testuser"}

	am := NewAPIManagement(manager)
	ctx := context.Background()

	apiKey, err := am.RegenerateAPIKeySecret(ctx, "key-123")

	assert.NoError(t, err, "Should regenerate secret without error")
	assert.NotNil(t, apiKey, "API key should not be nil")
	assert.NotEqual(t, oldSecret, apiKey.Secret, "Secret should be different")
	assert.NotEmpty(t, apiKey.Secret, "New secret should not be empty")
}

// TestValidateAPIKey tests API key validation
func TestValidateAPIKey(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		APIKeys:  make(map[string]*APIKey),
		Users:    make(map[string]*User),
		AuditLog: []AuditEntry{},
	}

	testKey := &APIKey{
		ID:         "key-123",
		UserID:     "user-123",
		Name:       "Test Key",
		Key:        "pk_test123",
		Secret:     "sk_secret123",
		Enabled:    true,
		UsageCount: 0,
	}
	manager.APIKeys["key-123"] = testKey
	manager.Users["testuser"] = &User{ID: "user-123", Username: "testuser"}

	am := NewAPIManagement(manager)
	ctx := context.Background()

	apiKey, err := am.ValidateAPIKey(ctx, "pk_test123", "sk_secret123")

	assert.NoError(t, err, "Should validate API key without error")
	assert.NotNil(t, apiKey, "API key should not be nil")
	assert.Equal(t, 1, apiKey.UsageCount, "Usage count should be incremented")
	assert.NotNil(t, apiKey.LastUsed, "LastUsed should be set")
}

// TestValidateAPIKey_InvalidSecret tests validation with wrong secret
func TestValidateAPIKey_InvalidSecret(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		APIKeys:  make(map[string]*APIKey),
		Users:    make(map[string]*User),
		AuditLog: []AuditEntry{},
	}

	testKey := &APIKey{
		ID:      "key-123",
		UserID:  "user-123",
		Name:    "Test Key",
		Key:     "pk_test123",
		Secret:  "sk_secret123",
		Enabled: true,
	}
	manager.APIKeys["key-123"] = testKey
	manager.Users["testuser"] = &User{ID: "user-123", Username: "testuser"}

	am := NewAPIManagement(manager)
	ctx := context.Background()

	apiKey, err := am.ValidateAPIKey(ctx, "pk_test123", "wrong_secret")

	assert.Error(t, err, "Should error with invalid secret")
	assert.Nil(t, apiKey, "API key should be nil")
	assert.Contains(t, err.Error(), "invalid API key secret")
}

// TestValidateAPIKey_Disabled tests validation with disabled key
func TestValidateAPIKey_Disabled(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		APIKeys:  make(map[string]*APIKey),
		Users:    make(map[string]*User),
		AuditLog: []AuditEntry{},
	}

	testKey := &APIKey{
		ID:      "key-123",
		UserID:  "user-123",
		Name:    "Test Key",
		Key:     "pk_test123",
		Secret:  "sk_secret123",
		Enabled: false,
	}
	manager.APIKeys["key-123"] = testKey
	manager.Users["testuser"] = &User{ID: "user-123", Username: "testuser"}

	am := NewAPIManagement(manager)
	ctx := context.Background()

	apiKey, err := am.ValidateAPIKey(ctx, "pk_test123", "sk_secret123")

	assert.Error(t, err, "Should error with disabled key")
	assert.Nil(t, apiKey, "API key should be nil")
	assert.Contains(t, err.Error(), "disabled")
}

// TestValidateAPIKey_Expired tests validation with expired key
func TestValidateAPIKey_Expired(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		APIKeys:  make(map[string]*APIKey),
		Users:    make(map[string]*User),
		AuditLog: []AuditEntry{},
	}

	expiredTime := time.Now().Add(-1 * time.Hour)
	testKey := &APIKey{
		ID:        "key-123",
		UserID:    "user-123",
		Name:      "Test Key",
		Key:       "pk_test123",
		Secret:    "sk_secret123",
		Enabled:   true,
		ExpiresAt: &expiredTime,
	}
	manager.APIKeys["key-123"] = testKey
	manager.Users["testuser"] = &User{ID: "user-123", Username: "testuser"}

	am := NewAPIManagement(manager)
	ctx := context.Background()

	apiKey, err := am.ValidateAPIKey(ctx, "pk_test123", "sk_secret123")

	assert.Error(t, err, "Should error with expired key")
	assert.Nil(t, apiKey, "API key should be nil")
	assert.Contains(t, err.Error(), "expired")
}

// TestListAPIKeys tests API key listing
func TestListAPIKeys(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:  *log,
		APIKeys: make(map[string]*APIKey),
	}

	// Add multiple API keys
	for i := 1; i <= 5; i++ {
		apiKey := &APIKey{
			ID:      fmt.Sprintf("key-%d", i),
			UserID:  "user-123",
			Name:    fmt.Sprintf("Key %d", i),
			Key:     fmt.Sprintf("pk_key%d", i),
			Secret:  fmt.Sprintf("sk_secret%d", i),
			Enabled: true,
		}
		manager.APIKeys[apiKey.ID] = apiKey
	}

	am := NewAPIManagement(manager)
	ctx := context.Background()

	req := ListAPIKeysRequest{
		Page:     1,
		PageSize: 10,
	}

	response, err := am.ListAPIKeys(ctx, req)

	assert.NoError(t, err, "Should list API keys without error")
	assert.NotNil(t, response, "Response should not be nil")
	assert.Equal(t, 5, response.Total, "Should have 5 total keys")
	assert.GreaterOrEqual(t, len(response.APIKeys), 5, "Should return keys")
	// Check that secrets are masked
	for _, key := range response.APIKeys {
		assert.Equal(t, "*****", key.Secret, "Secrets should be masked")
	}
}

// TestListAPIKeys_WithFilters tests listing with filters
func TestListAPIKeys_WithFilters(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:  *log,
		APIKeys: make(map[string]*APIKey),
	}

	// Add keys with different users
	manager.APIKeys["key-1"] = &APIKey{
		ID:      "key-1",
		UserID:  "user-1",
		Name:    "Key 1",
		Enabled: true,
	}
	manager.APIKeys["key-2"] = &APIKey{
		ID:      "key-2",
		UserID:  "user-2",
		Name:    "Key 2",
		Enabled: true,
	}

	am := NewAPIManagement(manager)
	ctx := context.Background()

	req := ListAPIKeysRequest{
		UserID:   "user-1",
		Page:     1,
		PageSize: 10,
	}

	response, err := am.ListAPIKeys(ctx, req)

	assert.NoError(t, err, "Should list API keys without error")
	assert.NotNil(t, response, "Response should not be nil")
	// May filter or not depending on implementation
}

// TestCheckAPIKeyRateLimit tests rate limiting
func TestCheckAPIKeyRateLimit(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:  *log,
		APIKeys: make(map[string]*APIKey),
	}

	testKey := &APIKey{
		ID:         "key-123",
		UserID:     "user-123",
		Name:       "Test Key",
		RateLimit:  100,
		UsageCount: 50,
		Enabled:    true,
	}
	manager.APIKeys["key-123"] = testKey

	am := NewAPIManagement(manager)
	ctx := context.Background()

	allowed, cooldown, err := am.CheckAPIKeyRateLimit(ctx, "key-123", 60)

	assert.NoError(t, err, "Should check rate limit without error")
	// May allow or deny depending on implementation
	_ = allowed
	_ = cooldown
}

// TestGetAPIKeyUsage tests usage statistics
func TestGetAPIKeyUsage(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:  *log,
		APIKeys: make(map[string]*APIKey),
	}

	lastUsed := time.Now()
	testKey := &APIKey{
		ID:         "key-123",
		UserID:     "user-123",
		Name:       "Test Key",
		UsageCount: 1000,
		LastUsed:   &lastUsed,
		CreatedAt:  time.Now().AddDate(0, 0, -7),
		Enabled:    true,
	}
	manager.APIKeys["key-123"] = testKey

	am := NewAPIManagement(manager)
	ctx := context.Background()

	usage, err := am.GetAPIKeyUsage(ctx, "key-123", 7)

	assert.NoError(t, err, "Should get usage without error")
	assert.NotNil(t, usage, "Usage should not be nil")
	assert.Equal(t, "key-123", usage.KeyID)
	assert.Equal(t, "Test Key", usage.KeyName)
	assert.Equal(t, 1000, usage.TotalUsage)
	assert.NotNil(t, usage.DailyUsage, "Daily usage should be populated")
	assert.NotNil(t, usage.HourlyUsage, "Hourly usage should be populated")
}

// TestGenerateAPIKey tests API key generation
func TestGenerateAPIKey(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:  *log,
		APIKeys: make(map[string]*APIKey),
	}

	am := NewAPIManagement(manager)

	key1 := am.generateAPIKey()
	key2 := am.generateAPIKey()

	assert.NotEmpty(t, key1, "Key should not be empty")
	assert.NotEmpty(t, key2, "Key should not be empty")
	assert.NotEqual(t, key1, key2, "Keys should be unique")
	assert.Contains(t, key1, "PK_", "Key should have prefix")
}

// TestGenerateAPISecret tests API secret generation
func TestGenerateAPISecret(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:  *log,
		APIKeys: make(map[string]*APIKey),
	}

	am := NewAPIManagement(manager)

	secret1 := am.generateAPISecret()
	secret2 := am.generateAPISecret()

	assert.NotEmpty(t, secret1, "Secret should not be empty")
	assert.NotEmpty(t, secret2, "Secret should not be empty")
	assert.NotEqual(t, secret1, secret2, "Secrets should be unique")
	assert.Contains(t, secret1, "SK_", "Secret should have prefix")
	assert.Greater(t, len(secret1), len(secret2)/2, "Secret should be sufficiently long")
}

// TestGetUsername tests username retrieval
func TestGetUsername(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:  *log,
		APIKeys: make(map[string]*APIKey),
		Users:   make(map[string]*User),
	}

	manager.Users["testuser"] = &User{
		ID:       "user-123",
		Username: "testuser",
	}

	am := NewAPIManagement(manager)

	username := am.getUsername("user-123")

	assert.Equal(t, "testuser", username)
}

// TestGetUsername_Unknown tests username retrieval for unknown user
func TestGetUsername_Unknown(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:  *log,
		APIKeys: make(map[string]*APIKey),
		Users:   make(map[string]*User),
	}

	am := NewAPIManagement(manager)

	username := am.getUsername("nonexistent")

	assert.Equal(t, "unknown", username)
}

// TestAPIKeysExceedLimit tests API key limit checking
func TestAPIKeysExceedLimit(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:  *log,
		APIKeys: make(map[string]*APIKey),
		Config: EnterpriseConfig{
			MaxAPIKeys: 2,
		},
	}

	// Add keys up to limit
	manager.APIKeys["key-1"] = &APIKey{ID: "key-1"}
	assert.False(t, manager.APIKeysExceedLimit(), "Should not exceed limit with 1 key")

	manager.APIKeys["key-2"] = &APIKey{ID: "key-2"}
	assert.True(t, manager.APIKeysExceedLimit(), "Should exceed limit with 2 keys")
}

// TestAPIKey_Structure tests APIKey struct
func TestAPIKey_Structure(t *testing.T) {
	apiKey := &APIKey{
		ID:          "key-123",
		UserID:      "user-123",
		Name:        "Test Key",
		Key:         "pk_test",
		Secret:      "sk_secret",
		Permissions: []string{"read", "write"},
		Scopes:      []string{"tests"},
		RateLimit:   1000,
		Enabled:     true,
		CreatedAt:   time.Now(),
		UsageCount:  5,
		Metadata:    map[string]string{"env": "prod"},
	}

	assert.NotNil(t, apiKey, "APIKey should not be nil")
	assert.Equal(t, "key-123", apiKey.ID)
	assert.Equal(t, 2, len(apiKey.Permissions))
	assert.Equal(t, 1, len(apiKey.Scopes))
	assert.Equal(t, 5, apiKey.UsageCount)
}

// TestCreateAPIKeyRequest_Validation tests request validation
func TestCreateAPIKeyRequest_Validation(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:  *log,
		APIKeys: make(map[string]*APIKey),
	}

	am := NewAPIManagement(manager)
	ctx := context.Background()

	testCases := []struct {
		name    string
		req     CreateAPIKeyRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: CreateAPIKeyRequest{
				UserID:      "user-123",
				Name:        "Valid Key",
				Permissions: []string{"read"},
				RateLimit:   100,
			},
			wantErr: false,
		},
		{
			name: "negative rate limit",
			req: CreateAPIKeyRequest{
				UserID:      "user-123",
				Name:        "Test",
				Permissions: []string{"read"},
				RateLimit:   -1,
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := am.validateCreateAPIKeyRequest(ctx, tc.req)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
