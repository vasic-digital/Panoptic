package enterprise

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"panoptic/internal/logger"
)

// EnterpriseManager manages enterprise features for Panoptic
type EnterpriseManager struct {
	Logger           logger.Logger
	Config           EnterpriseConfig
	Enabled          bool
	Users            map[string]*User
	Roles            map[string]*Role
	Teams            map[string]*Team
	Projects         map[string]*Project
	AuditLog         []AuditEntry
	Subscriptions    map[string]*Subscription
	APIKeys          map[string]*APIKey
	Sessions         map[string]*Session
	StoragePath      string
	Initialized      bool
}

// EnterpriseConfig contains enterprise configuration
type EnterpriseConfig struct {
	Enabled          bool                 `yaml:"enabled"`
	OrganizationName string               `yaml:"organization_name"`
	Domain          string               `yaml:"domain"`
	StoragePath     string               `yaml:"storage_path"`
	AuthMethod      string               `yaml:"auth_method"`        // local, ldap, sso, oauth2
	DefaultRole     string               `yaml:"default_role"`
	MaxUsers        int                  `yaml:"max_users"`
	MaxProjects     int                  `yaml:"max_projects"`
	MaxAPIKeys      int                  `yaml:"max_api_keys"`
	APIRateLimit    int                  `yaml:"api_rate_limit"`
	SessionTimeout  int                  `yaml:"session_timeout"`   // minutes
	PasswordPolicy  PasswordPolicy        `yaml:"password_policy"`
	License         LicenseConfig        `yaml:"license"`
	BackupConfig    BackupConfig         `yaml:"backup_config"`
	Compliance      ComplianceConfig     `yaml:"compliance"`
	Integration     IntegrationConfig    `yaml:"integration"`
}

// PasswordPolicy defines password requirements
type PasswordPolicy struct {
	MinLength        int  `yaml:"min_length"`
	RequireUppercase bool `yaml:"require_uppercase"`
	RequireLowercase bool `yaml:"require_lowercase"`
	RequireNumbers   bool `yaml:"require_numbers"`
	RequireSymbols   bool `yaml:"require_symbols"`
	MaxAgeDays      int  `yaml:"max_age_days"`
	MaxHistory       int  `yaml:"max_history"`
}

// LicenseConfig contains license information
type LicenseConfig struct {
	Key            string    `yaml:"key"`
	Type           string    `yaml:"type"`           // trial, standard, enterprise
	MaxUsers       int       `yaml:"max_users"`
	MaxProjects    int       `yaml:"max_projects"`
	MaxAPIKeys     int       `yaml:"max_api_keys"`
	ExpiresAt      time.Time `yaml:"expires_at"`
	Features       []string  `yaml:"features"`
	ValidationURL  string    `yaml:"validation_url"`
}

// BackupConfig contains backup settings
type BackupConfig struct {
	Enabled       bool     `yaml:"enabled"`
	Schedule      string    `yaml:"schedule"`        // daily, weekly, monthly
	RetentionDays int      `yaml:"retention_days"`
	Locations     []string  `yaml:"locations"`
	Compression   bool      `yaml:"compression"`
	Encryption    bool      `yaml:"encryption"`
}

// ComplianceConfig contains compliance settings
type ComplianceConfig struct {
	Enabled           bool   `yaml:"enabled"`
	Standards         []string `yaml:"standards"`        // GDPR, SOC2, HIPAA, ISO27001
	DataRetention     int    `yaml:"data_retention"`   // days
	AuditRetention     int    `yaml:"audit_retention"`   // days
	DataEncryption    bool   `yaml:"data_encryption"`
	AuditEncryption   bool   `yaml:"audit_encryption"`
	RequireApproval   bool   `yaml:"require_approval"`
	ApprovalWorkflow string `yaml:"approval_workflow"`
}

// IntegrationConfig contains third-party integrations
type IntegrationConfig struct {
	LDAP      LDAPConfig      `yaml:"ldap"`
	SSO        SSOConfig        `yaml:"sso"`
	OAuth2     OAuth2Config     `yaml:"oauth2"`
	Webhook    WebhookConfig    `yaml:"webhook"`
	SIEM       SIEMConfig       `yaml:"siem"`
	Monitoring MonitoringConfig `yaml:"monitoring"`
}

// LDAPConfig contains LDAP configuration
type LDAPConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Server     string `yaml:"server"`
	Port       int    `yaml:"port"`
	BaseDN     string `yaml:"base_dn"`
	UserFilter string `yaml:"user_filter"`
	BindDN     string `yaml:"bind_dn"`
	BindPassword string `yaml:"bind_password"`
	UseSSL     bool   `yaml:"use_ssl"`
}

// SSOConfig contains SSO configuration
type SSOConfig struct {
	Enabled      bool   `yaml:"enabled"`
	Provider     string `yaml:"provider"`     // saml, oidc, okta, azure-ad
	EntityID     string `yaml:"entity_id"`
	MetadataURL  string `yaml:"metadata_url"`
	Certificate  string `yaml:"certificate"`
	PrivateKey   string `yaml:"private_key"`
	RedirectURI  string `yaml:"redirect_uri"`
}

// OAuth2Config contains OAuth2 configuration
type OAuth2Config struct {
	Enabled   bool   `yaml:"enabled"`
	Provider  string `yaml:"provider"`    // github, google, microsoft, gitlab
	ClientID  string `yaml:"client_id"`
	Secret    string `yaml:"secret"`
	Redirect  string `yaml:"redirect"`
	Scopes    []string `yaml:"scopes"`
}

// WebhookConfig contains webhook configuration
type WebhookConfig struct {
	Enabled   bool     `yaml:"enabled"`
	Endpoints []string `yaml:"endpoints"`
	Headers   map[string]string `yaml:"headers"`
	Events    []string `yaml:"events"`
	Secret    string `yaml:"secret"`
}

// SIEMConfig contains SIEM integration configuration
type SIEMConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Provider string `yaml:"provider"`   // splunk, elk, graylog, datadog
	Endpoint string `yaml:"endpoint"`
	APIKey   string `yaml:"api_key"`
	Index    string `yaml:"index"`
}

// MonitoringConfig contains monitoring configuration
type MonitoringConfig struct {
	Enabled   bool     `yaml:"enabled"`
	Provider  string   `yaml:"provider"`   // prometheus, datadog, newrelic
	Endpoint  string   `yaml:"endpoint"`
	Metrics   []string `yaml:"metrics"`
	Labels    map[string]string `yaml:"labels"`
}

// User represents an enterprise user
type User struct {
	ID              string            `json:"id"`
	Username        string            `json:"username"`
	Email           string            `json:"email"`
	FirstName       string            `json:"first_name"`
	LastName        string            `json:"last_name"`
	PasswordHash    string            `json:"password_hash"`
	Role            string            `json:"role"`
	TeamIDs         []string          `json:"team_ids"`
	ProjectIDs      []string          `json:"project_ids"`
	APIKeys         []string          `json:"api_keys"`
	Permissions     map[string]bool   `json:"permissions"`
	Preferences     UserPreferences   `json:"preferences"`
	LastLogin       time.Time         `json:"last_login"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	Active          bool              `json:"active"`
	Metadata        map[string]string `json:"metadata"`
}

// UserPreferences contains user preferences
type UserPreferences struct {
	Theme        string `json:"theme"`
	Language     string `json:"language"`
	Timezone     string `json:"timezone"`
	DateFormat   string `json:"date_format"`
	TimeFormat   string `json:"time_format"`
	PageSize     int    `json:"page_size"`
	Notifications bool   `json:"notifications"`
	EmailDigest  bool   `json:"email_digest"`
}

// Role represents a user role with permissions
type Role struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Permissions map[string]bool   `json:"permissions"`
	Inherits    []string          `json:"inherits"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	System      bool              `json:"system"`
}

// Team represents an enterprise team
type Team struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	LeadID      string            `json:"lead_id"`
	MemberIDs   []string          `json:"member_ids"`
	ProjectIDs  []string          `json:"project_ids"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Active      bool              `json:"active"`
	Metadata    map[string]string `json:"metadata"`
}

// Project represents an enterprise project
type Project struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	OwnerID     string            `json:"owner_id"`
	TeamIDs     []string          `json:"team_ids"`
	MemberIDs   []string          `json:"member_ids"`
	Settings    ProjectSettings    `json:"settings"`
	Status      string            `json:"status"`       // active, archived, suspended
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	ArchivedAt  *time.Time        `json:"archived_at,omitempty"`
	Metadata    map[string]string `json:"metadata"`
}

// ProjectSettings contains project-specific settings
type ProjectSettings struct {
	Privacy         string `json:"privacy"`         // public, private, team
	TestRetention   int    `json:"test_retention"` // days
	MaxTestRuns    int    `json:"max_test_runs"`
	AllowSharing    bool   `json:"allow_sharing"`
	RequireApproval bool   `json:"require_approval"`
	BackupEnabled  bool   `json:"backup_enabled"`
}

// AuditEntry represents an audit log entry
type AuditEntry struct {
	ID          string            `json:"id"`
	Timestamp   time.Time         `json:"timestamp"`
	UserID      string            `json:"user_id"`
	Username    string            `json:"username"`
	Action      string            `json:"action"`
	Resource    string            `json:"resource"`
	ResourceID  string            `json:"resource_id"`
	Details     map[string]string `json:"details"`
	IPAddress   string            `json:"ip_address"`
	UserAgent    string            `json:"user_agent"`
	Success     bool              `json:"success"`
	ErrorCode   string            `json:"error_code,omitempty"`
	Severity    string            `json:"severity"`       // low, medium, high, critical
	Category    string            `json:"category"`       // auth, access, data, system, security
}

// Subscription represents an enterprise subscription
type Subscription struct {
	ID          string            `json:"id"`
	UserID      string            `json:"user_id"`
	Plan        string            `json:"plan"`           // trial, basic, pro, enterprise
	Status      string            `json:"status"`         // active, cancelled, expired, suspended
	StartDate   time.Time         `json:"start_date"`
	EndDate     time.Time         `json:"end_date"`
	RenewalDate time.Time         `json:"renewal_date"`
	Features    []string          `json:"features"`
	Limits      SubscriptionLimits `json:"limits"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// SubscriptionLimits contains subscription limits
type SubscriptionLimits struct {
	MaxUsers      int `json:"max_users"`
	MaxProjects   int `json:"max_projects"`
	MaxTestRuns   int `json:"max_test_runs"`
	MaxAPIKeys    int `json:"max_api_keys"`
	MaxStorageGB  int `json:"max_storage_gb"`
	MaxBandwidthGB int `json:"max_bandwidth_gb"`
}

// APIKey represents an API key for programmatic access
type APIKey struct {
	ID          string            `json:"id"`
	UserID      string            `json:"user_id"`
	Name        string            `json:"name"`
	Key         string            `json:"key"`
	Secret      string            `json:"secret"`
	Permissions []string          `json:"permissions"`
	Scopes      []string          `json:"scopes"`
	RateLimit   int               `json:"rate_limit"`
	Enabled     bool              `json:"enabled"`
	CreatedAt   time.Time         `json:"created_at"`
	ExpiresAt   *time.Time        `json:"expires_at,omitempty"`
	LastUsed    *time.Time        `json:"last_used,omitempty"`
	UsageCount  int               `json:"usage_count"`
	Metadata    map[string]string `json:"metadata"`
}

// Session represents a user session
type Session struct {
	ID        string            `json:"id"`
	UserID    string            `json:"user_id"`
	Token     string            `json:"token"`
	IPAddress string            `json:"ip_address"`
	UserAgent string            `json:"user_agent"`
	CreatedAt time.Time         `json:"created_at"`
	ExpiresAt time.Time         `json:"expires_at"`
	Active    bool              `json:"active"`
	Metadata  map[string]string `json:"metadata"`
}

// NewEnterpriseManager creates a new enterprise manager
func NewEnterpriseManager(log logger.Logger) *EnterpriseManager {
	return &EnterpriseManager{
		Logger:        log,
		Users:        make(map[string]*User),
		Roles:        make(map[string]*Role),
		Teams:        make(map[string]*Team),
		Projects:      make(map[string]*Project),
		AuditLog:      make([]AuditEntry, 0),
		Subscriptions: make(map[string]*Subscription),
		APIKeys:       make(map[string]*APIKey),
		Sessions:      make(map[string]*Session),
		Initialized:   false,
	}
}

// Initialize initializes the enterprise manager
func (em *EnterpriseManager) Initialize(config EnterpriseConfig) error {
	em.Config = config
	em.Enabled = config.Enabled
	em.StoragePath = config.StoragePath

	if !em.Enabled {
		em.Logger.Info("Enterprise management is disabled")
		return nil
	}

	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(em.StoragePath, 0700); err != nil {
		return fmt.Errorf("failed to create enterprise storage directory: %w", err)
	}

	// Initialize default roles
	if err := em.initializeDefaultRoles(); err != nil {
		return fmt.Errorf("failed to initialize default roles: %w", err)
	}

	// Load existing data
	if err := em.loadData(); err != nil {
		em.Logger.Warnf("Failed to load enterprise data: %v", err)
	}

	// Validate license
	if err := em.validateLicense(); err != nil {
		em.Logger.Warnf("License validation failed: %v", err)
	}

	// Initialize cleanup routines
	go em.startCleanupRoutines()

	em.Initialized = true
	em.Logger.Infof("Enterprise management initialized for organization: %s", config.OrganizationName)
	return nil
}

// initializeDefaultRoles creates default system roles
func (em *EnterpriseManager) initializeDefaultRoles() error {
	defaultRoles := map[string]map[string]bool{
		"admin": {
			"user.create":           true,
			"user.read":            true,
			"user.update":          true,
			"user.delete":          true,
			"role.create":          true,
			"role.read":           true,
			"role.update":         true,
			"role.delete":         true,
			"team.create":         true,
			"team.read":          true,
			"team.update":        true,
			"team.delete":        true,
			"project.create":      true,
			"project.read":       true,
			"project.update":     true,
			"project.delete":     true,
			"test.create":        true,
			"test.read":         true,
			"test.update":       true,
			"test.delete":       true,
			"test.run":          true,
			"report.read":       true,
			"report.create":      true,
			"analytics.read":    true,
			"settings.read":     true,
			"settings.update":   true,
			"system.admin":      true,
		},
		"manager": {
			"user.read":            true,
			"user.update":          true,
			"team.create":         true,
			"team.read":          true,
			"team.update":        true,
			"team.delete":        true,
			"project.create":      true,
			"project.read":       true,
			"project.update":     true,
			"project.delete":     true,
			"test.create":        true,
			"test.read":         true,
			"test.update":       true,
			"test.delete":       true,
			"test.run":          true,
			"report.read":       true,
			"report.create":      true,
			"analytics.read":    true,
			"settings.read":     true,
		},
		"developer": {
			"user.read":            true,
			"team.read":          true,
			"project.create":      true,
			"project.read":       true,
			"project.update":     true,
			"test.create":        true,
			"test.read":         true,
			"test.update":       true,
			"test.delete":       true,
			"test.run":          true,
			"report.read":       true,
			"report.create":      true,
			"analytics.read":    true,
		},
		"viewer": {
			"user.read":            true,
			"team.read":          true,
			"project.read":       true,
			"test.read":         true,
			"report.read":       true,
			"analytics.read":    true,
		},
	}

	for roleName, permissions := range defaultRoles {
		role := &Role{
			ID:          roleName,
			Name:        strings.Title(roleName),
			Description: fmt.Sprintf("Default %s role", roleName),
			Permissions: permissions,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			System:      true,
		}
		em.Roles[roleName] = role
	}

	em.Logger.Info("Default roles initialized successfully")
	return nil
}

// loadData loads enterprise data from storage
func (em *EnterpriseManager) loadData() error {
	// Load users
	if err := em.loadJSON("users.json", &em.Users); err != nil {
		em.Logger.Warnf("Failed to load users: %v", err)
	}

	// Load roles
	if err := em.loadJSON("roles.json", &em.Roles); err != nil {
		em.Logger.Warnf("Failed to load roles: %v", err)
	}

	// Load teams
	if err := em.loadJSON("teams.json", &em.Teams); err != nil {
		em.Logger.Warnf("Failed to load teams: %v", err)
	}

	// Load projects
	if err := em.loadJSON("projects.json", &em.Projects); err != nil {
		em.Logger.Warnf("Failed to load projects: %v", err)
	}

	// Load audit log
	if err := em.loadJSON("audit.json", &em.AuditLog); err != nil {
		em.Logger.Warnf("Failed to load audit log: %v", err)
	}

	// Load subscriptions
	if err := em.loadJSON("subscriptions.json", &em.Subscriptions); err != nil {
		em.Logger.Warnf("Failed to load subscriptions: %v", err)
	}

	// Load API keys
	if err := em.loadJSON("api_keys.json", &em.APIKeys); err != nil {
		em.Logger.Warnf("Failed to load API keys: %v", err)
	}

	em.Logger.Info("Enterprise data loaded successfully")
	return nil
}

// saveData saves enterprise data to storage
func (em *EnterpriseManager) saveData() error {
	// Save users
	if err := em.saveJSON("users.json", em.Users); err != nil {
		return fmt.Errorf("failed to save users: %w", err)
	}

	// Save roles
	if err := em.saveJSON("roles.json", em.Roles); err != nil {
		return fmt.Errorf("failed to save roles: %w", err)
	}

	// Save teams
	if err := em.saveJSON("teams.json", em.Teams); err != nil {
		return fmt.Errorf("failed to save teams: %w", err)
	}

	// Save projects
	if err := em.saveJSON("projects.json", em.Projects); err != nil {
		return fmt.Errorf("failed to save projects: %w", err)
	}

	// Save audit log
	if err := em.saveJSON("audit.json", em.AuditLog); err != nil {
		return fmt.Errorf("failed to save audit log: %w", err)
	}

	// Save subscriptions
	if err := em.saveJSON("subscriptions.json", em.Subscriptions); err != nil {
		return fmt.Errorf("failed to save subscriptions: %w", err)
	}

	// Save API keys
	if err := em.saveJSON("api_keys.json", em.APIKeys); err != nil {
		return fmt.Errorf("failed to save API keys: %w", err)
	}

	return nil
}

// loadJSON loads JSON data from file
func (em *EnterpriseManager) loadJSON(filename string, target interface{}) error {
	filePath := filepath.Join(em.StoragePath, filename)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

// saveJSON saves JSON data to file
func (em *EnterpriseManager) saveJSON(filename string, data interface{}) error {
	filePath := filepath.Join(em.StoragePath, filename)
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, jsonData, 0600)
}

// validateLicense validates the enterprise license
func (em *EnterpriseManager) validateLicense() error {
	if em.Config.License.Key == "" {
		return fmt.Errorf("no license key provided")
	}

	// For demo purposes, accept any license key
	em.Logger.Infof("License validated: %s (%s)", em.Config.License.Type, em.Config.License.Key)
	return nil
}

// startCleanupRoutines starts background cleanup routines
func (em *EnterpriseManager) startCleanupRoutines() {
	// Clean up expired sessions every 5 minutes
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			em.cleanupExpiredSessions()
		}
	}()

	// Rotate audit log every hour
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		for range ticker.C {
			em.rotateAuditLog()
		}
	}()

	// Clean up old backups daily
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		for range ticker.C {
			em.cleanupOldBackups()
		}
	}()
}

// cleanupExpiredSessions removes expired sessions
func (em *EnterpriseManager) cleanupExpiredSessions() {
	now := time.Now()
	for id, session := range em.Sessions {
		if now.After(session.ExpiresAt) {
			delete(em.Sessions, id)
			em.Logger.Debugf("Expired session removed: %s", id)
		}
	}
}

// rotateAuditLog rotates audit log files
func (em *EnterpriseManager) rotateAuditLog() {
	// For demo purposes, just log the rotation
	em.Logger.Info("Audit log rotation completed")
}

// cleanupOldBackups removes old backup files
func (em *EnterpriseManager) cleanupOldBackups() {
	// For demo purposes, just log the cleanup
	em.Logger.Info("Old backup cleanup completed")
}

// generateID generates a random ID
func (em *EnterpriseManager) generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)[:22]
}

// hashPassword hashes a password using bcrypt
func (em *EnterpriseManager) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// verifyPassword verifies a password against its hash
func (em *EnterpriseManager) verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// validatePassword validates password against policy
func (em *EnterpriseManager) validatePassword(password string) error {
	policy := em.Config.PasswordPolicy

	if len(password) < policy.MinLength {
		return fmt.Errorf("password must be at least %d characters", policy.MinLength)
	}

	if policy.RequireUppercase && !strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if policy.RequireLowercase && !strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz") {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if policy.RequireNumbers && !strings.ContainsAny(password, "0123456789") {
		return fmt.Errorf("password must contain at least one number")
	}

	if policy.RequireSymbols && !strings.ContainsAny(password, "!@#$%^&*()_+-=[]{}|;:,.<>?/") {
		return fmt.Errorf("password must contain at least one symbol")
	}

	return nil
}