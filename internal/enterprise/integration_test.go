package enterprise

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"panoptic/internal/logger"
)

// Test NewEnterpriseIntegration
func TestNewEnterpriseIntegration(t *testing.T) {
	log := logger.NewLogger(false)
	ei := NewEnterpriseIntegration(*log)

	assert.NotNil(t, ei)
	assert.NotNil(t, ei.Manager)
	assert.NotNil(t, ei.UserManagement)
	assert.NotNil(t, ei.ProjectManagement)
	assert.NotNil(t, ei.TeamManagement)
	assert.NotNil(t, ei.AuditManagement)
	assert.NotNil(t, ei.APIManagement)
	assert.NotNil(t, ei.Logger)
	assert.False(t, ei.Initialized)
}

// Test Initialize with empty config path (disabled mode)
func TestInitialize_EmptyConfigPath(t *testing.T) {
	log := logger.NewLogger(false)
	ei := NewEnterpriseIntegration(*log)

	err := ei.Initialize("")
	assert.NoError(t, err)
	assert.False(t, ei.Initialized)
}

// Test Initialize with valid config
func TestInitialize_ValidConfig(t *testing.T) {
	log := logger.NewLogger(false)
	ei := NewEnterpriseIntegration(*log)

	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "enterprise.yaml")

	configData := `
enabled: true
organization_name: "Test Corp"
domain: "testcorp.com"
storage_path: "` + filepath.Join(tmpDir, "data") + `"
license:
  key: "test-license-key"
  type: "enterprise"
  max_users: 100
  max_projects: 50
  max_api_keys: 100
  expires_at: "2030-12-31T23:59:59Z"
  features: ["sso", "audit"]
  validation_url: "https://license.example.com/validate"
sso_config:
  enabled: false
siem_config:
  enabled: false
backup_config:
  enabled: true
  schedule: "0 2 * * *"
  retention_days: 30
  backup_location: "` + tmpDir + `"
`
	err := os.WriteFile(configPath, []byte(configData), 0644)
	assert.NoError(t, err)

	err = ei.Initialize(configPath)
	assert.NoError(t, err)
	assert.True(t, ei.Initialized)
	assert.True(t, ei.Manager.Enabled)

	// Verify default admin was created
	assert.Greater(t, len(ei.Manager.Users), 0)
}

// Test Initialize with invalid config path
func TestInitialize_InvalidConfigPath(t *testing.T) {
	log := logger.NewLogger(false)
	ei := NewEnterpriseIntegration(*log)

	err := ei.Initialize("/nonexistent/config.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load enterprise configuration")
}

// Test ExecuteEnterpriseAction - not initialized
func TestExecuteEnterpriseAction_NotInitialized(t *testing.T) {
	log := logger.NewLogger(false)
	ei := NewEnterpriseIntegration(*log)

	ctx := context.Background()
	result, err := ei.ExecuteEnterpriseAction(ctx, "user_create", map[string]interface{}{})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not initialized")
}

// Test ExecuteEnterpriseAction - unsupported action
func TestExecuteEnterpriseAction_UnsupportedAction(t *testing.T) {
	ei := setupTestIntegration(t)

	ctx := context.Background()
	result, err := ei.ExecuteEnterpriseAction(ctx, "invalid_action", map[string]interface{}{})

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unsupported enterprise action")
}

// Test createUser action
func TestExecuteEnterpriseAction_CreateUser(t *testing.T) {
	ei := setupTestIntegration(t)

	ctx := context.Background()
	params := map[string]interface{}{
		"username":   "testuser",
		"email":      "test@example.com",
		"first_name": "Test",
		"last_name":  "User",
		"password":   "password123",
		"role":       "developer",
	}

	result, err := ei.ExecuteEnterpriseAction(ctx, "user_create", params)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap := result.(map[string]interface{})
	assert.Equal(t, "testuser", resultMap["username"])
	assert.Equal(t, "test@example.com", resultMap["email"])
	assert.Equal(t, "developer", resultMap["role"])
	assert.NotEmpty(t, resultMap["user_id"])
}

// Test authenticateUser action
func TestExecuteEnterpriseAction_AuthenticateUser(t *testing.T) {
	ei := setupTestIntegration(t)

	// Use the default admin user that was created during initialization
	ctx := context.Background()
	authParams := map[string]interface{}{
		"username": "admin",
		"password": "admin123",
	}

	result, err := ei.ExecuteEnterpriseAction(ctx, "user_authenticate", authParams)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap := result.(map[string]interface{})
	assert.NotEmpty(t, resultMap["session_id"])
	assert.NotEmpty(t, resultMap["user_id"])
	assert.NotNil(t, resultMap["expires_at"])
}

// Test createProject action
func TestExecuteEnterpriseAction_CreateProject(t *testing.T) {
	ei := setupTestIntegration(t)

	ctx := context.Background()
	params := map[string]interface{}{
		"name":        "Test Project",
		"description": "A test project",
		"owner_id":    "owner123",
	}

	result, err := ei.ExecuteEnterpriseAction(ctx, "project_create", params)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap := result.(map[string]interface{})
	assert.Equal(t, "Test Project", resultMap["name"])
	assert.Equal(t, "owner123", resultMap["owner_id"])
	assert.NotEmpty(t, resultMap["project_id"])
}

// Test createTeam action
func TestExecuteEnterpriseAction_CreateTeam(t *testing.T) {
	ei := setupTestIntegration(t)

	ctx := context.Background()
	params := map[string]interface{}{
		"name":        "Test Team",
		"description": "A test team",
		"lead_id":     "lead123",
	}

	result, err := ei.ExecuteEnterpriseAction(ctx, "team_create", params)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap := result.(map[string]interface{})
	assert.Equal(t, "Test Team", resultMap["name"])
	assert.Equal(t, "lead123", resultMap["lead_id"])
	assert.NotEmpty(t, resultMap["team_id"])
}

// Test createAPIKey action
func TestExecuteEnterpriseAction_CreateAPIKey(t *testing.T) {
	ei := setupTestIntegration(t)

	ctx := context.Background()
	params := map[string]interface{}{
		"user_id":    "user123",
		"name":       "Test API Key",
		"rate_limit": 1000,
		"enabled":    true,
		"permissions": []string{"read", "write"},
		"scopes":      []string{"tests", "reports"},
	}

	result, err := ei.ExecuteEnterpriseAction(ctx, "api_key_create", params)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap := result.(map[string]interface{})
	assert.Equal(t, "Test API Key", resultMap["name"])
	assert.NotEmpty(t, resultMap["api_key_id"])
	assert.NotEmpty(t, resultMap["key"])
}

// Test getAuditReport action
func TestExecuteEnterpriseAction_GetAuditReport(t *testing.T) {
	ei := setupTestIntegration(t)

	ctx := context.Background()
	params := map[string]interface{}{
		"page":      1,
		"page_size": 50,
	}

	result, err := ei.ExecuteEnterpriseAction(ctx, "audit_report", params)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap := result.(map[string]interface{})
	assert.NotNil(t, resultMap["entries"])
	assert.NotNil(t, resultMap["total"])
}

// Test getComplianceStatus action
func TestExecuteEnterpriseAction_GetComplianceStatus(t *testing.T) {
	ei := setupTestIntegration(t)

	ctx := context.Background()
	params := map[string]interface{}{
		"standards": []string{"SOC2", "GDPR"},
	}

	result, err := ei.ExecuteEnterpriseAction(ctx, "compliance_check", params)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// Test getEnterpriseStatus action
func TestExecuteEnterpriseAction_GetEnterpriseStatus(t *testing.T) {
	ei := setupTestIntegration(t)

	ctx := context.Background()
	params := map[string]interface{}{}

	result, err := ei.ExecuteEnterpriseAction(ctx, "enterprise_status", params)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap := result.(map[string]interface{})
	assert.True(t, resultMap["enabled"].(bool))
	assert.NotEmpty(t, resultMap["organization_name"])
	assert.NotNil(t, resultMap["total_users"])
}

// Test getLicenseInfo action
func TestExecuteEnterpriseAction_GetLicenseInfo(t *testing.T) {
	ei := setupTestIntegration(t)

	ctx := context.Background()
	params := map[string]interface{}{}

	result, err := ei.ExecuteEnterpriseAction(ctx, "license_info", params)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap := result.(map[string]interface{})
	assert.NotEmpty(t, resultMap["key"])
	assert.NotEmpty(t, resultMap["type"])
	assert.NotNil(t, resultMap["max_users"])
}

// Test backupData action
func TestExecuteEnterpriseAction_BackupData(t *testing.T) {
	ei := setupTestIntegration(t)

	ctx := context.Background()
	params := map[string]interface{}{
		"type": "full",
	}

	result, err := ei.ExecuteEnterpriseAction(ctx, "backup_data", params)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap := result.(map[string]interface{})
	assert.NotEmpty(t, resultMap["backup_name"])
	assert.Equal(t, "full", resultMap["backup_type"])
}

// Test cleanupData action
func TestExecuteEnterpriseAction_CleanupData(t *testing.T) {
	ei := setupTestIntegration(t)

	ctx := context.Background()
	params := map[string]interface{}{
		"dry_run":       true,
		"include_audit": true,
		"include_data":  false,
	}

	result, err := ei.ExecuteEnterpriseAction(ctx, "cleanup_data", params)

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap := result.(map[string]interface{})
	assert.True(t, resultMap["dry_run"].(bool))
	assert.NotNil(t, resultMap["results"])
}

// Test loadConfig
func TestLoadConfig(t *testing.T) {
	log := logger.NewLogger(false)
	ei := NewEnterpriseIntegration(*log)

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.yaml")

	configData := `
enabled: true
organization_name: "Test Org"
domain: "test.com"
storage_path: "/tmp/test"
license:
  key: "test-key"
  type: "enterprise"
  max_users: 50
  max_projects: 25
  max_api_keys: 50
  expires_at: "2030-01-01T00:00:00Z"
`
	err := os.WriteFile(configPath, []byte(configData), 0644)
	assert.NoError(t, err)

	config, err := ei.loadConfig(configPath)

	assert.NoError(t, err)
	assert.True(t, config.Enabled)
	assert.Equal(t, "Test Org", config.OrganizationName)
	assert.Equal(t, "test.com", config.Domain)
}

// Test loadConfig with missing file
func TestLoadConfig_MissingFile(t *testing.T) {
	log := logger.NewLogger(false)
	ei := NewEnterpriseIntegration(*log)

	_, err := ei.loadConfig("/nonexistent/config.yaml")

	assert.Error(t, err)
}

// Test loadConfig with invalid YAML
func TestLoadConfig_InvalidYAML(t *testing.T) {
	log := logger.NewLogger(false)
	ei := NewEnterpriseIntegration(*log)

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	err := os.WriteFile(configPath, []byte("invalid: yaml: content:"), 0644)
	assert.NoError(t, err)

	_, err = ei.loadConfig(configPath)

	assert.Error(t, err)
}

// Test createDefaultAdmin
func TestCreateDefaultAdmin(t *testing.T) {
	ei := setupTestIntegration(t)

	// Clear existing users
	ei.Manager.Users = make(map[string]*User)

	err := ei.createDefaultAdmin()

	assert.NoError(t, err)
	assert.Equal(t, 1, len(ei.Manager.Users))

	// Verify admin user exists
	var adminUser *User
	for _, user := range ei.Manager.Users {
		if user.Role == "admin" {
			adminUser = user
			break
		}
	}

	assert.NotNil(t, adminUser)
	assert.Equal(t, "admin", adminUser.Username)
}

// Test createDefaultAdmin with existing admin
func TestCreateDefaultAdmin_AlreadyExists(t *testing.T) {
	ei := setupTestIntegration(t)

	// Verify admin already exists from setup
	initialCount := len(ei.Manager.Users)
	assert.Greater(t, initialCount, 0)

	err := ei.createDefaultAdmin()

	assert.NoError(t, err)
	// Count should not increase
	assert.Equal(t, initialCount, len(ei.Manager.Users))
}

// Test countActiveUsers
func TestCountActiveUsers(t *testing.T) {
	ei := setupTestIntegration(t)

	// Add some test users
	ei.Manager.Users = map[string]*User{
		"user1": {ID: "user1", Username: "user1", Active: true},
		"user2": {ID: "user2", Username: "user2", Active: true},
		"user3": {ID: "user3", Username: "user3", Active: false},
	}

	count := ei.countActiveUsers()

	assert.Equal(t, 2, count)
}

// Test countActiveProjects
func TestCountActiveProjects(t *testing.T) {
	ei := setupTestIntegration(t)

	// Add some test projects
	ei.Manager.Projects = map[string]*Project{
		"proj1": {ID: "proj1", Name: "Project 1", Status: "active"},
		"proj2": {ID: "proj2", Name: "Project 2", Status: "active"},
		"proj3": {ID: "proj3", Name: "Project 3", Status: "archived"},
	}

	count := ei.countActiveProjects()

	assert.Equal(t, 2, count)
}

// Test countActiveTeams
func TestCountActiveTeams(t *testing.T) {
	ei := setupTestIntegration(t)

	// Add some test teams
	ei.Manager.Teams = map[string]*Team{
		"team1": {ID: "team1", Name: "Team 1", Active: true},
		"team2": {ID: "team2", Name: "Team 2", Active: true},
		"team3": {ID: "team3", Name: "Team 3", Active: false},
	}

	count := ei.countActiveTeams()

	assert.Equal(t, 2, count)
}

// Test countActiveAPIKeys
func TestCountActiveAPIKeys(t *testing.T) {
	ei := setupTestIntegration(t)

	// Add some test API keys
	now := time.Now()
	exp1 := now.Add(24 * time.Hour)
	exp2 := now.Add(24 * time.Hour)
	exp3 := now.Add(24 * time.Hour)
	ei.Manager.APIKeys = map[string]*APIKey{
		"key1": {ID: "key1", Name: "Key 1", Enabled: true, ExpiresAt: &exp1},
		"key2": {ID: "key2", Name: "Key 2", Enabled: true, ExpiresAt: &exp2},
		"key3": {ID: "key3", Name: "Key 3", Enabled: false, ExpiresAt: &exp3},
	}

	count := ei.countActiveAPIKeys()

	assert.Equal(t, 2, count)
}

// Test utility functions
func TestGetString(t *testing.T) {
	params := map[string]interface{}{
		"existing": "value",
		"number":   123,
	}

	assert.Equal(t, "value", getString(params, "existing"))
	assert.Equal(t, "", getString(params, "nonexistent"))
	assert.Equal(t, "", getString(params, "number"))
}

func TestGetStringSlice(t *testing.T) {
	params := map[string]interface{}{
		"existing": []string{"a", "b", "c"},
		"number":   123,
	}

	result := getStringSlice(params, "existing")
	assert.Equal(t, []string{"a", "b", "c"}, result)

	result = getStringSlice(params, "nonexistent")
	assert.Equal(t, []string{}, result)

	result = getStringSlice(params, "number")
	assert.Equal(t, []string{}, result)
}

func TestGetInt(t *testing.T) {
	params := map[string]interface{}{
		"int_value":   42,
		"float_value": 3.14,
		"string":      "not a number",
	}

	assert.Equal(t, 42, getInt(params, "int_value", 0))
	assert.Equal(t, 3, getInt(params, "float_value", 0))
	assert.Equal(t, 99, getInt(params, "nonexistent", 99))
	assert.Equal(t, 99, getInt(params, "string", 99))
}

func TestGetBool(t *testing.T) {
	params := map[string]interface{}{
		"true_value":  true,
		"false_value": false,
		"string":      "not a bool",
	}

	assert.True(t, getBool(params, "true_value", false))
	assert.False(t, getBool(params, "false_value", true))
	assert.True(t, getBool(params, "nonexistent", true))
	assert.False(t, getBool(params, "string", false))
}

// Test createBackupFile
func TestCreateBackupFile(t *testing.T) {
	log := logger.NewLogger(false)
	ei := NewEnterpriseIntegration(*log)

	tmpDir := t.TempDir()
	backupPath := filepath.Join(tmpDir, "test_backup.tar.gz")

	err := ei.createBackupFile(backupPath, "full")

	assert.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(backupPath)
	assert.NoError(t, err)
}

// Test getBackupSize
func TestGetBackupSize(t *testing.T) {
	log := logger.NewLogger(false)
	ei := NewEnterpriseIntegration(*log)

	tmpDir := t.TempDir()
	backupPath := filepath.Join(tmpDir, "test_backup.tar.gz")

	// Create test file
	err := os.WriteFile(backupPath, []byte("test content"), 0644)
	assert.NoError(t, err)

	size := ei.getBackupSize(backupPath)

	assert.Greater(t, size, int64(0))
}

// Test getBackupSize with nonexistent file
func TestGetBackupSize_NonexistentFile(t *testing.T) {
	log := logger.NewLogger(false)
	ei := NewEnterpriseIntegration(*log)

	size := ei.getBackupSize("/nonexistent/file.tar.gz")

	assert.Equal(t, int64(0), size)
}

// Test getEnterpriseStatus when disabled
func TestGetEnterpriseStatus_Disabled(t *testing.T) {
	log := logger.NewLogger(false)
	ei := NewEnterpriseIntegration(*log)

	// Initialize with empty config (disabled mode)
	err := ei.Initialize("")
	assert.NoError(t, err)

	// Manually set initialized for testing disabled status
	ei.Initialized = true
	ei.Manager.Enabled = false

	ctx := context.Background()
	result, err := ei.getEnterpriseStatus(ctx, map[string]interface{}{})

	assert.NoError(t, err)
	assert.NotNil(t, result)

	resultMap := result.(map[string]interface{})
	assert.False(t, resultMap["enabled"].(bool))
	assert.Contains(t, resultMap["reason"], "disabled")
}

// Test backupData with backup disabled
func TestBackupData_BackupDisabled(t *testing.T) {
	ei := setupTestIntegration(t)

	// Disable backup
	ei.Manager.Config.BackupConfig.Enabled = false

	ctx := context.Background()
	params := map[string]interface{}{
		"type": "full",
	}

	result, err := ei.backupData(ctx, params)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "backup is not enabled")
}

// Helper function to setup test integration
func setupTestIntegration(t *testing.T) *EnterpriseIntegration {
	log := logger.NewLogger(false)
	ei := NewEnterpriseIntegration(*log)

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "enterprise.yaml")

	configData := `
enabled: true
organization_name: "Test Corp"
domain: "testcorp.com"
storage_path: "` + filepath.Join(tmpDir, "data") + `"
license:
  key: "test-license-key"
  type: "enterprise"
  max_users: 100
  max_projects: 50
  max_api_keys: 100
  expires_at: "2030-12-31T23:59:59Z"
  features: ["sso", "audit", "backup", "compliance"]
  validation_url: "https://license.example.com/validate"
sso_config:
  enabled: false
siem_config:
  enabled: false
compliance:
  enabled: true
  standards: ["SOC2", "GDPR", "HIPAA", "PCI-DSS"]
  data_retention: 90
backup_config:
  enabled: true
  schedule: "0 2 * * *"
  retention_days: 30
  backup_location: "` + tmpDir + `"
`
	err := os.WriteFile(configPath, []byte(configData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	err = ei.Initialize(configPath)
	if err != nil {
		t.Fatalf("Failed to initialize test integration: %v", err)
	}

	return ei
}
