package enterprise

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"panoptic/internal/logger"
)

// EnterpriseIntegration handles enterprise features integration with Panoptic
type EnterpriseIntegration struct {
	Manager                *EnterpriseManager
	UserManagement         *UserManagement
	ProjectManagement      *ProjectManagement
	TeamManagement         *TeamManagement
	AuditManagement        *AuditManagement
	APIManagement          *APIManagement
	Logger                 logger.Logger
	Initialized           bool
}

// NewEnterpriseIntegration creates new enterprise integration
func NewEnterpriseIntegration(log logger.Logger) *EnterpriseIntegration {
	manager := NewEnterpriseManager(log)
	
	return &EnterpriseIntegration{
		Manager:           manager,
		UserManagement:    NewUserManagement(manager),
		ProjectManagement:  NewProjectManagement(manager),
		TeamManagement:     NewTeamManagement(manager),
		AuditManagement:    NewAuditManagement(manager),
		APIManagement:      NewAPIManagement(manager),
		Logger:            log,
		Initialized:       false,
	}
}

// Initialize initializes enterprise integration
func (ei *EnterpriseIntegration) Initialize(configPath string) error {
	if configPath == "" {
		ei.Logger.Info("Enterprise integration disabled: no configuration provided")
		return nil
	}

	// Load enterprise configuration
	config, err := ei.loadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load enterprise configuration: %w", err)
	}

	// Initialize enterprise manager
	if err := ei.Manager.Initialize(config); err != nil {
		return fmt.Errorf("failed to initialize enterprise manager: %w", err)
	}

	if !ei.Manager.Enabled {
		ei.Logger.Info("Enterprise integration is disabled")
		return nil
	}

	// Create default admin user if none exists
	if err := ei.createDefaultAdmin(); err != nil {
		ei.Logger.Warnf("Failed to create default admin user: %v", err)
	}

	ei.Initialized = true
	ei.Logger.Infof("Enterprise integration initialized for organization: %s", config.OrganizationName)
	return nil
}

// ExecuteEnterpriseAction executes enterprise-specific actions
func (ei *EnterpriseIntegration) ExecuteEnterpriseAction(ctx context.Context, actionType string, params map[string]interface{}) (interface{}, error) {
	if !ei.Initialized {
		return nil, fmt.Errorf("enterprise integration is not initialized")
	}

	switch actionType {
	case "user_create":
		return ei.createUser(ctx, params)
	case "user_authenticate":
		return ei.authenticateUser(ctx, params)
	case "project_create":
		return ei.createProject(ctx, params)
	case "team_create":
		return ei.createTeam(ctx, params)
	case "api_key_create":
		return ei.createAPIKey(ctx, params)
	case "audit_report":
		return ei.getAuditReport(ctx, params)
	case "compliance_check":
		return ei.getComplianceStatus(ctx, params)
	case "enterprise_status":
		return ei.getEnterpriseStatus(ctx, params)
	case "license_info":
		return ei.getLicenseInfo(ctx, params)
	case "backup_data":
		return ei.backupData(ctx, params)
	case "cleanup_data":
		return ei.cleanupData(ctx, params)
	default:
		return nil, fmt.Errorf("unsupported enterprise action: %s", actionType)
	}
}

// Action handlers

func (ei *EnterpriseIntegration) createUser(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	req := CreateUserRequest{
		Username:  getString(params, "username"),
		Email:     getString(params, "email"),
		FirstName: getString(params, "first_name"),
		LastName:  getString(params, "last_name"),
		Password:  getString(params, "password"),
		Role:      getString(params, "role"),
	}

	if teams, ok := params["team_ids"].([]string); ok {
		req.TeamIDs = teams
	}
	if projects, ok := params["project_ids"].([]string); ok {
		req.ProjectIDs = projects
	}

	user, err := ei.UserManagement.CreateUser(ctx, req)
	if err != nil {
		return nil, err
	}

	// Remove sensitive data
	user.PasswordHash = ""
	return map[string]interface{}{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"created_at": user.CreatedAt,
	}, nil
}

func (ei *EnterpriseIntegration) authenticateUser(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	username := getString(params, "username")
	password := getString(params, "password")

	session, err := ei.UserManagement.AuthenticateUser(ctx, username, password)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"session_id":  session.ID,
		"user_id":     session.UserID,
		"expires_at":  session.ExpiresAt,
		"created_at":  session.CreatedAt,
	}, nil
}

func (ei *EnterpriseIntegration) createProject(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	req := CreateProjectRequest{
		Name:        getString(params, "name"),
		Description: getString(params, "description"),
		OwnerID:     getString(params, "owner_id"),
	}

	if teams, ok := params["team_ids"].([]string); ok {
		req.TeamIDs = teams
	}
	if members, ok := params["member_ids"].([]string); ok {
		req.MemberIDs = members
	}

	project, err := ei.ProjectManagement.CreateProject(ctx, req)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"project_id": project.ID,
		"name":       project.Name,
		"owner_id":   project.OwnerID,
		"status":     project.Status,
		"created_at": project.CreatedAt,
	}, nil
}

func (ei *EnterpriseIntegration) createTeam(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	req := CreateTeamRequest{
		Name:        getString(params, "name"),
		Description: getString(params, "description"),
		LeadID:      getString(params, "lead_id"),
	}

	if members, ok := params["member_ids"].([]string); ok {
		req.MemberIDs = members
	}
	if projects, ok := params["project_ids"].([]string); ok {
		req.ProjectIDs = projects
	}

	team, err := ei.TeamManagement.CreateTeam(ctx, req)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"team_id":    team.ID,
		"name":       team.Name,
		"lead_id":    team.LeadID,
		"active":     team.Active,
		"created_at": team.CreatedAt,
	}, nil
}

func (ei *EnterpriseIntegration) createAPIKey(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	req := CreateAPIKeyRequest{
		UserID:      getString(params, "user_id"),
		Name:        getString(params, "name"),
		RateLimit:   getInt(params, "rate_limit", 0),
		Enabled:     getBool(params, "enabled", true),
	}

	if permissions, ok := params["permissions"].([]string); ok {
		req.Permissions = permissions
	}
	if scopes, ok := params["scopes"].([]string); ok {
		req.Scopes = scopes
	}

	apiKey, err := ei.APIManagement.CreateAPIKey(ctx, req)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"api_key_id":  apiKey.ID,
		"name":        apiKey.Name,
		"key":         apiKey.Key,
		"permissions": apiKey.Permissions,
		"scopes":      apiKey.Scopes,
		"rate_limit":  apiKey.RateLimit,
		"expires_at":  apiKey.ExpiresAt,
		"created_at":  apiKey.CreatedAt,
	}, nil
}

func (ei *EnterpriseIntegration) getAuditReport(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	req := GetAuditLogRequest{
		Page:     getInt(params, "page", 1),
		PageSize: getInt(params, "page_size", 50),
		UserID:   getString(params, "user_id"),
		Action:   getString(params, "action"),
		Severity: getString(params, "severity"),
	}

	response, err := ei.AuditManagement.GetAuditLog(ctx, req)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"entries":   response.Entries,
		"total":     response.Total,
		"page":      response.Page,
		"page_size": response.PageSize,
	}, nil
}

func (ei *EnterpriseIntegration) getComplianceStatus(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	req := GetComplianceStatusRequest{
		Standards: getStringSlice(params, "standards"),
	}

	response, err := ei.AuditManagement.GetComplianceStatus(ctx, req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (ei *EnterpriseIntegration) getEnterpriseStatus(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	if !ei.Manager.Enabled {
		return map[string]interface{}{
			"enabled": false,
			"reason":  "enterprise features are disabled",
		}, nil
	}

	status := map[string]interface{}{
		"enabled":            true,
		"organization_name":  ei.Manager.Config.OrganizationName,
		"domain":            ei.Manager.Config.Domain,
		"license_type":       ei.Manager.Config.License.Type,
		"license_expires_at": ei.Manager.Config.License.ExpiresAt,
		"total_users":       len(ei.Manager.Users),
		"active_users":      ei.countActiveUsers(),
		"total_projects":    len(ei.Manager.Projects),
		"active_projects":   ei.countActiveProjects(),
		"total_teams":       len(ei.Manager.Teams),
		"active_teams":      ei.countActiveTeams(),
		"total_api_keys":    len(ei.Manager.APIKeys),
		"active_api_keys":   ei.countActiveAPIKeys(),
		"audit_entries":     len(ei.Manager.AuditLog),
		"storage_path":      ei.Manager.Config.StoragePath,
		"initialized_at":    time.Now(),
	}

	return status, nil
}

func (ei *EnterpriseIntegration) getLicenseInfo(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	license := ei.Manager.Config.License

	info := map[string]interface{}{
		"key":             license.Key,
		"type":            license.Type,
		"max_users":       license.MaxUsers,
		"max_projects":    license.MaxProjects,
		"max_api_keys":    license.MaxAPIKeys,
		"expires_at":      license.ExpiresAt,
		"features":        license.Features,
		"validation_url":  license.ValidationURL,
		"days_until_expiry": time.Until(license.ExpiresAt).Hours() / 24,
	}

	// Add current usage info
	info["current_users"] = len(ei.Manager.Users)
	info["current_projects"] = len(ei.Manager.Projects)
	info["current_api_keys"] = len(ei.Manager.APIKeys)

	return info, nil
}

func (ei *EnterpriseIntegration) backupData(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	if !ei.Manager.Config.BackupConfig.Enabled {
		return nil, fmt.Errorf("backup is not enabled")
	}

	backupType := getString(params, "type")

	// Create backup
	backupPath := filepath.Join(ei.Manager.Config.StoragePath, "backups")
	if err := os.MkdirAll(backupPath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	backupName := fmt.Sprintf("backup_%s_%s.tar.gz", backupType, time.Now().Format("20060102_150405"))
	backupFile := filepath.Join(backupPath, backupName)

	// In a real implementation, create actual backup file
	if err := ei.createBackupFile(backupFile, backupType); err != nil {
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}

	// Log backup creation
	ei.Manager.logAuditEntry(AuditEntry{
		ID:        ei.Manager.generateID(),
		Timestamp: time.Now(),
		Action:    "backup.create",
		Resource:  "backup",
		Details: map[string]string{
			"type":       backupType,
			"file_name":  backupName,
			"file_size":  fmt.Sprintf("%d", ei.getBackupSize(backupFile)),
		},
		Success:  true,
		Severity: "medium",
		Category: "system",
	})

	return map[string]interface{}{
		"backup_name": backupName,
		"backup_type": backupType,
		"backup_path": backupFile,
		"created_at":  time.Now(),
	}, nil
}

func (ei *EnterpriseIntegration) cleanupData(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	dryRun := getBool(params, "dry_run", false)
	includeAudit := getBool(params, "include_audit", true)
	includeData := getBool(params, "include_data", true)

	req := ExecuteCleanupRequest{
		IncludeAudit: includeAudit,
		IncludeData:  includeData,
		DryRun:      dryRun,
	}

	response, err := ei.AuditManagement.ExecuteCleanup(ctx, req)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"started_at":  response.StartedAt,
		"completed_at": response.CompletedAt,
		"duration":    response.Duration.String(),
		"results":     response.Results,
		"dry_run":     dryRun,
	}, nil
}

// Helper methods

func (ei *EnterpriseIntegration) loadConfig(configPath string) (EnterpriseConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return EnterpriseConfig{}, err
	}

	var config EnterpriseConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return EnterpriseConfig{}, err
	}

	// Set default storage path if not provided
	if config.StoragePath == "" {
		config.StoragePath = "./enterprise_data"
	}

	return config, nil
}

func (ei *EnterpriseIntegration) createDefaultAdmin() error {
	// Check if any admin user exists
	for _, user := range ei.Manager.Users {
		if user.Role == "admin" {
			return nil
		}
	}

	// Create default admin user
	req := CreateUserRequest{
		Username:  "admin",
		Email:     "admin@" + ei.Manager.Config.Domain,
		FirstName:  "System",
		LastName:   "Administrator",
		Password:   "admin123", // Should be changed on first login
		Role:       "admin",
	}

	user, err := ei.UserManagement.CreateUser(context.Background(), req)
	if err != nil {
		return err
	}

	ei.Logger.Warnf("Default admin user created: username=admin, password=admin123 (please change password)")
	ei.Logger.Infof("Admin user details: %s (%s)", user.Username, user.Email)
	return nil
}

func (ei *EnterpriseIntegration) countActiveUsers() int {
	count := 0
	for _, user := range ei.Manager.Users {
		if user.Active {
			count++
		}
	}
	return count
}

func (ei *EnterpriseIntegration) countActiveProjects() int {
	count := 0
	for _, project := range ei.Manager.Projects {
		if project.Status == "active" {
			count++
		}
	}
	return count
}

func (ei *EnterpriseIntegration) countActiveTeams() int {
	count := 0
	for _, team := range ei.Manager.Teams {
		if team.Active {
			count++
		}
	}
	return count
}

func (ei *EnterpriseIntegration) countActiveAPIKeys() int {
	count := 0
	for _, apiKey := range ei.Manager.APIKeys {
		if apiKey.Enabled {
			count++
		}
	}
	return count
}

func (ei *EnterpriseIntegration) createBackupFile(backupPath string, backupType string) error {
	// In a real implementation, create actual backup
	// For demo purposes, just create empty file
	file, err := os.Create(backupPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("Backup type: %s\nCreated at: %s\n", backupType, time.Now().Format(time.RFC3339)))
	return err
}

func (ei *EnterpriseIntegration) getBackupSize(backupPath string) int64 {
	info, err := os.Stat(backupPath)
	if err != nil {
		return 0
	}
	return info.Size()
}

// Utility functions

func getString(params map[string]interface{}, key string) string {
	if val, ok := params[key].(string); ok {
		return val
	}
	return ""
}

func getStringSlice(params map[string]interface{}, key string) []string {
	if val, ok := params[key].([]string); ok {
		return val
	}
	return []string{}
}

func getInt(params map[string]interface{}, key string, defaultValue int) int {
	if val, ok := params[key].(int); ok {
		return val
	}
	if val, ok := params[key].(float64); ok {
		return int(val)
	}
	return defaultValue
}

func getBool(params map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := params[key].(bool); ok {
		return val
	}
	return defaultValue
}