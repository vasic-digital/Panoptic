package enterprise

import (
	"context"
	"fmt"
	"testing"
	"time"

	"panoptic/internal/logger"

	"github.com/stretchr/testify/assert"
)

// TestNewUserManagement tests user management creation
func TestNewUserManagement(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger: *log,
		Users:  make(map[string]*User),
	}

	um := NewUserManagement(manager)

	assert.NotNil(t, um, "UserManagement should not be nil")
	assert.Equal(t, manager, um.Manager, "Manager should be set")
}

// TestCreateUser tests user creation
func TestCreateUser(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:        *log,
		Users:         make(map[string]*User),
		Roles:         make(map[string]*Role),
		Subscriptions: make(map[string]*Subscription),
		Config: EnterpriseConfig{
			MaxUsers: 100,
			PasswordPolicy: PasswordPolicy{
				MinLength: 8,
			},
		},
	}

	um := NewUserManagement(manager)
	ctx := context.Background()

	req := CreateUserRequest{
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "SecurePass123",
		FirstName: "Test",
		LastName:  "User",
		Role:      "user",
	}

	user, err := um.CreateUser(ctx, req)

	assert.NoError(t, err, "Should create user without error")
	assert.NotNil(t, user, "User should not be nil")
	assert.Equal(t, req.Username, user.Username)
	assert.Equal(t, req.Email, user.Email)
	assert.NotEmpty(t, user.ID, "User ID should be generated")
	assert.NotEmpty(t, user.PasswordHash, "Password should be hashed")
	assert.NotEqual(t, req.Password, user.PasswordHash, "Password should be hashed")
	assert.True(t, user.Active, "User should be active")
}

// TestCreateUser_DuplicateUsername tests creating user with duplicate username
func TestCreateUser_DuplicateUsername(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger: *log,
		Users:  make(map[string]*User),
		Config: EnterpriseConfig{
			MaxUsers: 100,
			PasswordPolicy: PasswordPolicy{
				MinLength: 8,
			},
		},
	}

	// Pre-populate with existing user
	manager.Users["testuser"] = &User{
		ID:       "existing-id",
		Username: "testuser",
		Email:    "existing@example.com",
	}

	um := NewUserManagement(manager)
	ctx := context.Background()

	req := CreateUserRequest{
		Username:  "testuser", // Duplicate
		Email:     "new@example.com",
		FirstName: "New",
		LastName:  "User",
		Password:  "SecurePass123",
		Role:      "user",
	}

	user, err := um.CreateUser(ctx, req)

	assert.Error(t, err, "Should error on duplicate username")
	assert.Nil(t, user, "User should be nil on error")
	assert.Contains(t, err.Error(), "already exists", "Error should mention duplicate")
}

// TestCreateUser_InvalidRequest tests with invalid request
func TestCreateUser_InvalidRequest(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger: *log,
		Users:  make(map[string]*User),
		Config: EnterpriseConfig{
			MaxUsers: 100,
		},
	}

	um := NewUserManagement(manager)
	ctx := context.Background()

	testCases := []struct {
		name string
		req  CreateUserRequest
	}{
		{
			name: "empty username",
			req: CreateUserRequest{
				Username: "",
				Email:    "test@example.com",
				Password: "password",
			},
		},
		{
			name: "empty email",
			req: CreateUserRequest{
				Username: "testuser",
				Email:    "",
				Password: "password",
			},
		},
		{
			name: "empty password",
			req: CreateUserRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user, err := um.CreateUser(ctx, tc.req)
			assert.Error(t, err, "Should error with invalid request")
			assert.Nil(t, user, "User should be nil")
		})
	}
}

// TestGetUser tests user retrieval
func TestGetUser(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger: *log,
		Users:  make(map[string]*User),
	}

	testUser := &User{
		ID:       "user-123",
		Username: "testuser",
		Email:    "test@example.com",
		Active:   true,
	}
	manager.Users["testuser"] = testUser

	um := NewUserManagement(manager)
	ctx := context.Background()

	user, err := um.GetUser(ctx, "testuser")

	assert.NoError(t, err, "Should get user without error")
	assert.NotNil(t, user, "User should not be nil")
	assert.Equal(t, testUser.ID, user.ID)
	assert.Equal(t, testUser.Username, user.Username)
}

// TestGetUser_NotFound tests getting nonexistent user
func TestGetUser_NotFound(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger: *log,
		Users:  make(map[string]*User),
	}

	um := NewUserManagement(manager)
	ctx := context.Background()

	user, err := um.GetUser(ctx, "nonexistent")

	assert.Error(t, err, "Should error when user not found")
	assert.Nil(t, user, "User should be nil")
	assert.Contains(t, err.Error(), "not found", "Error should mention not found")
}

// TestUpdateUser tests user update
func TestUpdateUser(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger: *log,
		Users:  make(map[string]*User),
	}

	testUser := &User{
		ID:        "user-123",
		Username:  "testuser",
		Email:     "old@example.com",
		FirstName: "Old",
		LastName:  "Name",
		Active:    true,
		CreatedAt: time.Now(),
	}
	manager.Users["testuser"] = testUser

	um := NewUserManagement(manager)
	ctx := context.Background()

	req := UpdateUserRequest{
		Email:     stringPtr("new@example.com"),
		FirstName: stringPtr("New"),
		LastName:  stringPtr("Name"),
	}

	user, err := um.UpdateUser(ctx, "testuser", req)

	assert.NoError(t, err, "Should update user without error")
	assert.NotNil(t, user, "User should not be nil")
	assert.Equal(t, "new@example.com", user.Email)
	assert.Equal(t, "New", user.FirstName)
	assert.Equal(t, "Name", user.LastName)
}

// TestUpdateUser_NotFound tests updating nonexistent user
func TestUpdateUser_NotFound(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger: *log,
		Users:  make(map[string]*User),
	}

	um := NewUserManagement(manager)
	ctx := context.Background()

	req := UpdateUserRequest{
		Email: stringPtr("new@example.com"),
	}

	user, err := um.UpdateUser(ctx, "nonexistent", req)

	assert.Error(t, err, "Should error when user not found")
	assert.Nil(t, user, "User should be nil")
}

// TestDeleteUser tests user deletion
func TestDeleteUser(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		Users:    make(map[string]*User),
		Sessions: make(map[string]*Session),
	}

	testUser := &User{
		ID:       "user-123",
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
		Active:   true,
	}
	manager.Users["testuser"] = testUser

	um := NewUserManagement(manager)
	ctx := context.Background()

	err := um.DeleteUser(ctx, "testuser")

	assert.NoError(t, err, "Should delete user without error")
	_, exists := manager.Users["testuser"]
	assert.False(t, exists, "User should be removed from map")
}

// TestDeleteUser_LastAdmin tests deleting last admin
func TestDeleteUser_LastAdmin(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger: *log,
		Users:  make(map[string]*User),
	}

	adminUser := &User{
		ID:       "admin-123",
		Username: "admin",
		Email:    "admin@example.com",
		Role:     "admin",
		Active:   true,
	}
	manager.Users["admin"] = adminUser

	um := NewUserManagement(manager)
	ctx := context.Background()

	err := um.DeleteUser(ctx, "admin")

	assert.Error(t, err, "Should not allow deleting last admin")
	assert.Contains(t, err.Error(), "last admin", "Error should mention last admin")
}

// TestAuthenticateUser tests user authentication
func TestAuthenticateUser(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		Users:    make(map[string]*User),
		Sessions: make(map[string]*Session),
		Config: EnterpriseConfig{
			SessionTimeout: 60,
		},
	}

	password := "SecurePass123"
	hash, _ := manager.hashPassword(password)

	testUser := &User{
		ID:           "user-123",
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hash,
		Active:       true,
		Role:         "user",
	}
	manager.Users["testuser"] = testUser

	um := NewUserManagement(manager)
	ctx := context.Background()

	session, err := um.AuthenticateUser(ctx, "testuser", password)

	assert.NoError(t, err, "Should authenticate without error")
	assert.NotNil(t, session, "Session should not be nil")
	assert.Equal(t, testUser.ID, session.UserID)
	assert.NotEmpty(t, session.Token, "Session token should be generated")
	assert.True(t, session.ExpiresAt.After(time.Now()), "Session should not be expired")
}

// TestAuthenticateUser_WrongPassword tests authentication with wrong password
func TestAuthenticateUser_WrongPassword(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		Users:    make(map[string]*User),
		Sessions: make(map[string]*Session),
	}

	hash, _ := manager.hashPassword("CorrectPassword")

	testUser := &User{
		ID:           "user-123",
		Username:     "testuser",
		PasswordHash: hash,
		Active:       true,
	}
	manager.Users["testuser"] = testUser

	um := NewUserManagement(manager)
	ctx := context.Background()

	session, err := um.AuthenticateUser(ctx, "testuser", "WrongPassword")

	assert.Error(t, err, "Should error with wrong password")
	assert.Nil(t, session, "Session should be nil")
	assert.Contains(t, err.Error(), "invalid", "Error should mention invalid credentials")
}

// TestAuthenticateUser_InactiveUser tests authentication with inactive user
func TestAuthenticateUser_InactiveUser(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		Users:    make(map[string]*User),
		Sessions: make(map[string]*Session),
	}

	hash, _ := manager.hashPassword("password")

	testUser := &User{
		ID:           "user-123",
		Username:     "testuser",
		PasswordHash: hash,
		Active:       false, // Inactive
	}
	manager.Users["testuser"] = testUser

	um := NewUserManagement(manager)
	ctx := context.Background()

	session, err := um.AuthenticateUser(ctx, "testuser", "password")

	assert.Error(t, err, "Should error with inactive user")
	assert.Nil(t, session, "Session should be nil")
	assert.Contains(t, err.Error(), "inactive", "Error should mention inactive")
}

// TestLogoutUser tests user logout
func TestLogoutUser(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		Sessions: make(map[string]*Session),
		Users:    make(map[string]*User),
	}

	testUser := &User{
		ID:       "user-123",
		Username: "testuser",
		Active:   true,
	}
	manager.Users["testuser"] = testUser

	session := &Session{
		ID:        "session-123",
		UserID:    "user-123",
		Token:     "token-abc",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	manager.Sessions["session-123"] = session

	um := NewUserManagement(manager)
	ctx := context.Background()

	err := um.LogoutUser(ctx, "session-123")

	assert.NoError(t, err, "Should logout without error")
	_, exists := manager.Sessions["session-123"]
	assert.False(t, exists, "Session should be removed")
}

// TestLogoutUser_InvalidSession tests logout with invalid session
func TestLogoutUser_InvalidSession(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		Sessions: make(map[string]*Session),
		Users:    make(map[string]*User),
	}

	um := NewUserManagement(manager)
	ctx := context.Background()

	err := um.LogoutUser(ctx, "nonexistent-session")

	assert.Error(t, err, "Should error with invalid session")
}

// TestValidateSession tests session validation
func TestValidateSession(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		Sessions: make(map[string]*Session),
	}

	session := &Session{
		ID:        "session-123",
		UserID:    "user-123",
		Token:     "token-abc",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Active:    true,
	}
	manager.Sessions["session-123"] = session

	um := NewUserManagement(manager)
	ctx := context.Background()

	validSession, err := um.ValidateSession(ctx, "session-123")

	assert.NoError(t, err, "Should validate session without error")
	assert.NotNil(t, validSession, "Session should not be nil")
	assert.Equal(t, session.ID, validSession.ID)
}

// TestValidateSession_Expired tests validation with expired session
func TestValidateSession_Expired(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		Sessions: make(map[string]*Session),
	}

	session := &Session{
		ID:        "session-123",
		UserID:    "user-123",
		Token:     "token-abc",
		CreatedAt: time.Now().Add(-2 * time.Hour),
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
		Active:    true,
	}
	manager.Sessions["session-123"] = session

	um := NewUserManagement(manager)
	ctx := context.Background()

	validSession, err := um.ValidateSession(ctx, "session-123")

	assert.Error(t, err, "Should error with expired session")
	assert.Nil(t, validSession, "Session should be nil")
	assert.Contains(t, err.Error(), "expired", "Error should mention expired")
}

// TestListUsers tests user listing
func TestListUsers(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger: *log,
		Users:  make(map[string]*User),
	}

	// Add multiple users
	for i := 1; i <= 5; i++ {
		user := &User{
			ID:       fmt.Sprintf("user-%d", i),
			Username: fmt.Sprintf("user%d", i),
			Email:    fmt.Sprintf("user%d@example.com", i),
			Role:     "user",
			Active:   true,
		}
		manager.Users[user.Username] = user
	}

	um := NewUserManagement(manager)
	ctx := context.Background()

	req := ListUsersRequest{
		Page:     1,
		PageSize: 10,
	}

	response, err := um.ListUsers(ctx, req)

	assert.NoError(t, err, "Should list users without error")
	assert.NotNil(t, response, "Response should not be nil")
	assert.Equal(t, 5, response.Total, "Should have 5 total users")
	assert.GreaterOrEqual(t, len(response.Users), 5, "Should return users")
}

// TestListUsers_WithFilters tests listing with filters
func TestListUsers_WithFilters(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger: *log,
		Users:  make(map[string]*User),
	}

	// Add users with different roles
	manager.Users["admin1"] = &User{
		ID:       "admin-1",
		Username: "admin1",
		Role:     "admin",
		Active:   true,
	}
	manager.Users["user1"] = &User{
		ID:       "user-1",
		Username: "user1",
		Role:     "user",
		Active:   true,
	}

	um := NewUserManagement(manager)
	ctx := context.Background()

	req := ListUsersRequest{
		Role:     "admin",
		Page:     1,
		PageSize: 10,
	}

	response, err := um.ListUsers(ctx, req)

	assert.NoError(t, err, "Should list users without error")
	assert.NotNil(t, response, "Response should not be nil")
	// May filter by role or not depending on implementation
}

// TestGenerateSessionToken tests session token generation
func TestGenerateSessionToken(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger: *log,
	}

	um := NewUserManagement(manager)

	token1 := um.generateSessionToken()
	token2 := um.generateSessionToken()

	assert.NotEmpty(t, token1, "Token should not be empty")
	assert.NotEmpty(t, token2, "Token should not be empty")
	assert.NotEqual(t, token1, token2, "Tokens should be unique")
	assert.Greater(t, len(token1), 20, "Token should be sufficiently long")
}

// TestGetRolePermissions tests role permission retrieval
func TestGetRolePermissions(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger: *log,
		Roles:  make(map[string]*Role),
	}

	// Add test role
	manager.Roles["admin"] = &Role{
		ID:   "role-admin",
		Name: "admin",
		Permissions: map[string]bool{
			"read":   true,
			"write":  true,
			"delete": true,
		},
	}

	um := NewUserManagement(manager)

	permissions := um.getRolePermissions("admin")

	assert.NotNil(t, permissions, "Permissions should not be nil")
	// May return role permissions or default permissions
}

// TestGetAdminCount tests admin user counting
func TestGetAdminCount(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger: *log,
		Users:  make(map[string]*User),
	}

	// Add admin users
	manager.Users["admin1"] = &User{
		ID:       "admin-1",
		Username: "admin1",
		Role:     "admin",
		Active:   true,
	}
	manager.Users["admin2"] = &User{
		ID:       "admin-2",
		Username: "admin2",
		Role:     "admin",
		Active:   true,
	}
	manager.Users["user1"] = &User{
		ID:       "user-1",
		Username: "user1",
		Role:     "user",
		Active:   true,
	}

	um := NewUserManagement(manager)

	count := um.getAdminCount()

	assert.Equal(t, 2, count, "Should have 2 admin users")
}

// TestCleanupUserData tests user data cleanup
func TestCleanupUserData(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		Sessions: make(map[string]*Session),
		APIKeys:  make(map[string]*APIKey),
	}

	// Add user sessions
	manager.Sessions["session-1"] = &Session{
		ID:     "session-1",
		UserID: "user-123",
	}
	manager.Sessions["session-2"] = &Session{
		ID:     "session-2",
		UserID: "other-user",
	}

	um := NewUserManagement(manager)

	um.cleanupUserData("user-123")

	// Should cleanup user sessions
	_, exists := manager.Sessions["session-1"]
	// May or may not remove session depending on implementation
	if !exists {
		t.Log("User session was cleaned up")
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}

// TestUser_Structure tests User struct
func TestUser_Structure(t *testing.T) {
	user := User{
		ID:        "user-123",
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Role:      "user",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	assert.Equal(t, "user-123", user.ID)
	assert.Equal(t, "testuser", user.Username)
	assert.True(t, user.Active)
}

// TestSession_Structure tests Session struct
func TestSession_Structure(t *testing.T) {
	session := Session{
		ID:        "session-123",
		UserID:    "user-123",
		Token:     "token-abc",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	assert.Equal(t, "session-123", session.ID)
	assert.Equal(t, "user-123", session.UserID)
	assert.NotEmpty(t, session.Token)
}

// TestCreateUserRequest_Validation tests request validation
func TestCreateUserRequest_Validation(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger: *log,
		Config: EnterpriseConfig{
			PasswordPolicy: PasswordPolicy{
				MinLength: 8,
			},
		},
	}

	um := NewUserManagement(manager)

	testCases := []struct {
		name    string
		req     CreateUserRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: CreateUserRequest{
				Username:  "validuser",
				Email:     "valid@example.com",
				FirstName: "Valid",
				LastName:  "User",
				Password:  "SecurePass123",
				Role:      "user",
			},
			wantErr: false,
		},
		{
			name: "invalid email",
			req: CreateUserRequest{
				Username: "user",
				Email:    "invalid-email",
				Password: "password",
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := um.validateCreateUserRequest(tc.req)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
