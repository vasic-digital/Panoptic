package enterprise

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"panoptic/internal/logger"
)

// UserManagement handles user operations
type UserManagement struct {
	Manager *EnterpriseManager
	Logger   logger.Logger
}

// NewUserManagement creates new user management handler
func NewUserManagement(manager *EnterpriseManager) *UserManagement {
	return &UserManagement{
		Manager: manager,
		Logger:  manager.Logger,
	}
}

// CreateUser creates a new user
func (um *UserManagement) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
	// Validate request
	if err := um.validateCreateUserRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Check if user already exists
	if _, exists := um.Manager.Users[req.Username]; exists {
		return nil, fmt.Errorf("user with username '%s' already exists", req.Username)
	}

	if _, exists := um.Manager.Users[req.Email]; exists {
		return nil, fmt.Errorf("user with email '%s' already exists", req.Email)
	}

	// Check user limits
	if um.Manager.UsersExceedLimit() {
		return nil, fmt.Errorf("maximum number of users reached")
	}

	// Hash password
	hashedPassword, err := um.Manager.hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &User{
		ID:           um.Manager.generateID(),
		Username:     req.Username,
		Email:        req.Email,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		PasswordHash: hashedPassword,
		Role:         req.Role,
		TeamIDs:      req.TeamIDs,
		ProjectIDs:   req.ProjectIDs,
		Permissions:  um.getRolePermissions(req.Role),
		Preferences:  UserPreferences{
			Theme:        "light",
			Language:     "en",
			Timezone:     "UTC",
			DateFormat:   "2006-01-02",
			TimeFormat:   "15:04:05",
			PageSize:     25,
			Notifications: true,
			EmailDigest:  true,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Active:    true,
		Metadata:  req.Metadata,
	}

	// Store user
	um.Manager.Users[user.Username] = user

	// Create default subscription if needed
	if um.Manager.Config.DefaultRole == "" || req.Role == "" {
		subscription := &Subscription{
			ID:         um.Manager.generateID(),
			UserID:     user.ID,
			Plan:       "trial",
			Status:     "active",
			StartDate:  time.Now(),
			EndDate:    time.Now().AddDate(0, 0, 30),
			RenewalDate: time.Now().AddDate(0, 1, 0),
			Features:   []string{"basic_testing", "limited_api"},
			Limits: SubscriptionLimits{
				MaxUsers:      1,
				MaxProjects:   1,
				MaxTestRuns:   100,
				MaxAPIKeys:    1,
				MaxStorageGB:  1,
				MaxBandwidthGB: 10,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		um.Manager.Subscriptions[subscription.ID] = subscription
	}

	// Log audit entry
	um.Manager.logAuditEntry(AuditEntry{
		ID:         um.Manager.generateID(),
		Timestamp:  time.Now(),
		UserID:     user.ID,
		Username:   req.Username,
		Action:     "user.create",
		Resource:   "user",
		ResourceID: user.ID,
		Details:    map[string]string{"username": req.Username, "role": req.Role},
		Success:    true,
		Severity:   "medium",
		Category:   "auth",
	})

	// Save data
	if err := um.Manager.saveData(); err != nil {
		um.Logger.Errorf("Failed to save user data: %v", err)
	}

	um.Logger.Infof("User created successfully: %s (%s)", req.Username, req.Email)
	return user, nil
}

// GetUser retrieves a user by ID or username
func (um *UserManagement) GetUser(ctx context.Context, identifier string) (*User, error) {
	user, exists := um.Manager.Users[identifier]
	if !exists {
		// Try by ID
		for _, u := range um.Manager.Users {
			if u.ID == identifier {
				user = u
				exists = true
				break
			}
		}
	}

	if !exists {
		return nil, fmt.Errorf("user not found: %s", identifier)
	}

	return user, nil
}

// UpdateUser updates an existing user
func (um *UserManagement) UpdateUser(ctx context.Context, identifier string, req UpdateUserRequest) (*User, error) {
	user, err := um.GetUser(ctx, identifier)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.FirstName != nil && *req.FirstName != "" {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil && *req.LastName != "" {
		user.LastName = *req.LastName
	}
	if req.Email != nil && *req.Email != "" {
		user.Email = *req.Email
	}
	if req.Role != nil && *req.Role != "" {
		user.Role = *req.Role
		user.Permissions = um.getRolePermissions(*req.Role)
	}
	if req.TeamIDs != nil {
		user.TeamIDs = req.TeamIDs
	}
	if req.ProjectIDs != nil {
		user.ProjectIDs = req.ProjectIDs
	}
	if req.Metadata != nil {
		user.Metadata = *req.Metadata
	}

	user.UpdatedAt = time.Now()

	// Log audit entry
	um.Manager.logAuditEntry(AuditEntry{
		ID:         um.Manager.generateID(),
		Timestamp:  time.Now(),
		UserID:     user.ID,
		Username:   user.Username,
		Action:     "user.update",
		Resource:   "user",
		ResourceID: user.ID,
		Details:    map[string]string{"updated_fields": "user_info"},
		Success:    true,
		Severity:   "medium",
		Category:   "auth",
	})

	// Save data
	if err := um.Manager.saveData(); err != nil {
		um.Logger.Errorf("Failed to save user data: %v", err)
	}

	um.Logger.Infof("User updated successfully: %s", user.Username)
	return user, nil
}

// DeleteUser deletes a user
func (um *UserManagement) DeleteUser(ctx context.Context, identifier string) error {
	user, err := um.GetUser(ctx, identifier)
	if err != nil {
		return err
	}

	// Check if user can be deleted (not last admin)
	if user.Role == "admin" && um.getAdminCount() <= 1 {
		return fmt.Errorf("cannot delete last admin user")
	}

	// Delete user
	delete(um.Manager.Users, user.Username)

	// Clean up related data
	um.cleanupUserData(user.ID)

	// Log audit entry
	um.Manager.logAuditEntry(AuditEntry{
		ID:         um.Manager.generateID(),
		Timestamp:  time.Now(),
		UserID:     user.ID,
		Username:   user.Username,
		Action:     "user.delete",
		Resource:   "user",
		ResourceID: user.ID,
		Details:    map[string]string{"username": user.Username},
		Success:    true,
		Severity:   "high",
		Category:   "auth",
	})

	// Save data
	if err := um.Manager.saveData(); err != nil {
		um.Logger.Errorf("Failed to save user data: %v", err)
	}

	um.Logger.Infof("User deleted successfully: %s", user.Username)
	return nil
}

// AuthenticateUser authenticates a user
func (um *UserManagement) AuthenticateUser(ctx context.Context, username, password string) (*Session, error) {
	user, err := um.GetUser(ctx, username)
	if err != nil {
		// Log failed authentication
		um.Manager.logAuditEntry(AuditEntry{
			ID:         um.Manager.generateID(),
			Timestamp:  time.Now(),
			Username:   username,
			Action:     "auth.login",
			Resource:   "user",
			Details:    map[string]string{"reason": "user_not_found"},
			Success:    false,
			Severity:   "medium",
			Category:   "auth",
		})
		return nil, fmt.Errorf("invalid credentials")
	}

	if !user.Active {
		um.Manager.logAuditEntry(AuditEntry{
			ID:         um.Manager.generateID(),
			Timestamp:  time.Now(),
			UserID:     user.ID,
			Username:   username,
			Action:     "auth.login",
			Resource:   "user",
			Details:    map[string]string{"reason": "user_inactive"},
			Success:    false,
			Severity:   "medium",
			Category:   "auth",
		})
		return nil, fmt.Errorf("account is inactive")
	}

	if !um.Manager.verifyPassword(password, user.PasswordHash) {
		um.Manager.logAuditEntry(AuditEntry{
			ID:         um.Manager.generateID(),
			Timestamp:  time.Now(),
			UserID:     user.ID,
			Username:   username,
			Action:     "auth.login",
			Resource:   "user",
			Details:    map[string]string{"reason": "invalid_password"},
			Success:    false,
			Severity:   "medium",
			Category:   "auth",
		})
		return nil, fmt.Errorf("invalid credentials")
	}

	// Create session
	session := &Session{
		ID:        um.Manager.generateID(),
		UserID:    user.ID,
		Token:     um.generateSessionToken(),
		IPAddress: um.getClientIP(ctx),
		UserAgent: um.getUserAgent(ctx),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(time.Duration(um.Manager.Config.SessionTimeout) * time.Minute),
		Active:    true,
	}

	// Store session
	um.Manager.Sessions[session.ID] = session

	// Update user last login
	user.LastLogin = time.Now()
	user.UpdatedAt = time.Now()

	// Log successful authentication
	um.Manager.logAuditEntry(AuditEntry{
		ID:         um.Manager.generateID(),
		Timestamp:  time.Now(),
		UserID:     user.ID,
		Username:   username,
		Action:     "auth.login",
		Resource:   "user",
		ResourceID: user.ID,
		Details:    map[string]string{"session_id": session.ID},
		Success:    true,
		Severity:   "low",
		Category:   "auth",
	})

	// Save data
	if err := um.Manager.saveData(); err != nil {
		um.Logger.Errorf("Failed to save user data: %v", err)
	}

	um.Logger.Infof("User authenticated successfully: %s", username)
	return session, nil
}

// LogoutUser logs out a user
func (um *UserManagement) LogoutUser(ctx context.Context, sessionID string) error {
	session, exists := um.Manager.Sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	// Delete session
	delete(um.Manager.Sessions, sessionID)
	session.Active = false

	// Log audit entry
	user, _ := um.GetUser(ctx, session.UserID)
	um.Manager.logAuditEntry(AuditEntry{
		ID:         um.Manager.generateID(),
		Timestamp:  time.Now(),
		UserID:     session.UserID,
		Username:   um.getUsername(session.UserID),
		Action:     "auth.logout",
		Resource:   "session",
		ResourceID: sessionID,
		Success:    true,
		Severity:   "low",
		Category:   "auth",
	})

	um.Logger.Infof("User logged out successfully: %s", user.Username)
	return nil
}

// ValidateSession validates a session
func (um *UserManagement) ValidateSession(ctx context.Context, sessionID string) (*Session, error) {
	session, exists := um.Manager.Sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	if !session.Active {
		return nil, fmt.Errorf("session is inactive")
	}

	if time.Now().After(session.ExpiresAt) {
		delete(um.Manager.Sessions, sessionID)
		return nil, fmt.Errorf("session expired")
	}

	return session, nil
}

// ListUsers lists all users with pagination and filtering
func (um *UserManagement) ListUsers(ctx context.Context, req ListUsersRequest) (*ListUsersResponse, error) {
	var users []User
	for _, user := range um.Manager.Users {
		// Apply filters
		if req.Role != "" && user.Role != req.Role {
			continue
		}
		if req.Active != nil && *req.Active != user.Active {
			continue
		}
		if req.TeamID != "" && !contains(user.TeamIDs, req.TeamID) {
			continue
		}
		if req.ProjectID != "" && !contains(user.ProjectIDs, req.ProjectID) {
			continue
		}

		users = append(users, *user)
	}

	// Apply pagination
	total := len(users)
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize
	if end > total {
		end = total
	}
	if start >= total {
		return &ListUsersResponse{
			Users: []User{},
			Total: total,
			Page:  req.Page,
			PageSize: req.PageSize,
		}, nil
	}

	pagedUsers := users[start:end]

	return &ListUsersResponse{
		Users: pagedUsers,
		Total: total,
		Page:  req.Page,
		PageSize: req.PageSize,
	}, nil
}

// Request types

type CreateUserRequest struct {
	Username  string            `json:"username"`
	Email     string            `json:"email"`
	FirstName string            `json:"first_name"`
	LastName  string            `json:"last_name"`
	Password  string            `json:"password"`
	Role      string            `json:"role"`
	TeamIDs   []string          `json:"team_ids"`
	ProjectIDs []string         `json:"project_ids"`
	Metadata  map[string]string `json:"metadata"`
}

type UpdateUserRequest struct {
	FirstName  *string           `json:"first_name,omitempty"`
	LastName   *string           `json:"last_name,omitempty"`
	Email      *string           `json:"email,omitempty"`
	Role       *string           `json:"role,omitempty"`
	TeamIDs    []string          `json:"team_ids,omitempty"`
	ProjectIDs []string          `json:"project_ids,omitempty"`
	Preferences *UserPreferences `json:"preferences,omitempty"`
	Metadata   *map[string]string `json:"metadata,omitempty"`
}

type ListUsersRequest struct {
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
	Role      string `json:"role,omitempty"`
	Active    *bool  `json:"active,omitempty"`
	TeamID    string `json:"team_id,omitempty"`
	ProjectID string `json:"project_id,omitempty"`
}

type ListUsersResponse struct {
	Users    []User `json:"users"`
	Total    int    `json:"total"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

// Helper methods

func (um *UserManagement) validateCreateUserRequest(req CreateUserRequest) error {
	if req.Username == "" {
		return fmt.Errorf("username is required")
	}
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.FirstName == "" {
		return fmt.Errorf("first name is required")
	}
	if req.LastName == "" {
		return fmt.Errorf("last name is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	if req.Role == "" {
		req.Role = um.Manager.Config.DefaultRole
	}

	return um.Manager.validatePassword(req.Password)
}

func (um *UserManagement) getRolePermissions(roleName string) map[string]bool {
	role, exists := um.Manager.Roles[roleName]
	if !exists {
		return map[string]bool{}
	}
	return role.Permissions
}

func (um *UserManagement) getAdminCount() int {
	count := 0
	for _, user := range um.Manager.Users {
		if user.Role == "admin" && user.Active {
			count++
		}
	}
	return count
}

func (um *UserManagement) cleanupUserData(userID string) {
	// Delete user's API keys
	for keyID, apiKey := range um.Manager.APIKeys {
		if apiKey.UserID == userID {
			delete(um.Manager.APIKeys, keyID)
		}
	}

	// Delete user's sessions
	for sessionID, session := range um.Manager.Sessions {
		if session.UserID == userID {
			delete(um.Manager.Sessions, sessionID)
		}
	}

	// Delete user's subscriptions
	for subID, subscription := range um.Manager.Subscriptions {
		if subscription.UserID == userID {
			delete(um.Manager.Subscriptions, subID)
		}
	}
}

func (um *UserManagement) generateSessionToken() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

func (um *UserManagement) getClientIP(ctx context.Context) string {
	// In a real implementation, extract from request context
	return "127.0.0.1"
}

func (um *UserManagement) getUserAgent(ctx context.Context) string {
	// In a real implementation, extract from request context
	return "panoptic-client"
}

func (um *UserManagement) getUsername(userID string) string {
	for _, user := range um.Manager.Users {
		if user.ID == userID {
			return user.Username
		}
	}
	return "unknown"
}

func (em *EnterpriseManager) logAuditEntry(entry AuditEntry) {
	entry.ID = em.generateID()
	em.AuditLog = append(em.AuditLog, entry)

	// Trim audit log if too large
	if len(em.AuditLog) > 10000 {
		em.AuditLog = em.AuditLog[1000:] // Keep last 9000 entries
	}

	// Log to SIEM if configured
	if em.Config.Integration.SIEM.Enabled {
		em.sendToSIEM(entry)
	}
}

func (em *EnterpriseManager) sendToSIEM(entry AuditEntry) {
	// In a real implementation, send to SIEM system
	em.Logger.Debugf("Audit entry sent to SIEM: %s - %s", entry.Action, entry.Resource)
}

func (em *EnterpriseManager) UsersExceedLimit() bool {
	if em.Config.MaxUsers <= 0 {
		return false
	}

	count := 0
	for _, user := range em.Users {
		if user.Active {
			count++
		}
	}

	return count >= em.Config.MaxUsers
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}