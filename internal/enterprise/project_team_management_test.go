package enterprise

import (
	"context"
	"fmt"
	"testing"
	"time"

	"panoptic/internal/logger"

	"github.com/stretchr/testify/assert"
)

// ============================================================================
// PROJECT MANAGEMENT TESTS
// ============================================================================

// TestNewProjectManagement tests project management creation
func TestNewProjectManagement(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		Projects: make(map[string]*Project),
	}

	pm := NewProjectManagement(manager)

	assert.NotNil(t, pm, "ProjectManagement should not be nil")
	assert.Equal(t, manager, pm.Manager, "Manager should be set")
}

// TestCreateProject tests project creation
func TestCreateProject(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		Projects: make(map[string]*Project),
		Users:    make(map[string]*User),
		AuditLog: []AuditEntry{},
		Config: EnterpriseConfig{
			MaxProjects: 100,
		},
	}

	// Add test user
	manager.Users["testuser"] = &User{
		ID:         "user-123",
		Username:   "testuser",
		ProjectIDs: []string{},
	}

	pm := NewProjectManagement(manager)
	ctx := context.Background()

	req := CreateProjectRequest{
		Name:        "Test Project",
		Description: "A test project",
		OwnerID:     "user-123",
		MemberIDs:   []string{"user-123"},
		Settings:    ProjectSettings{Privacy: "private"},
		Metadata:    map[string]string{"env": "test"},
	}

	project, err := pm.CreateProject(ctx, req)

	assert.NoError(t, err, "Should create project without error")
	assert.NotNil(t, project, "Project should not be nil")
	assert.Equal(t, req.Name, project.Name)
	assert.Equal(t, req.OwnerID, project.OwnerID)
	assert.Equal(t, "active", project.Status)
	assert.Contains(t, project.MemberIDs, req.OwnerID, "Owner should be in members")
}

// TestCreateProject_InvalidRequest tests with invalid request
func TestCreateProject_InvalidRequest(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		Projects: make(map[string]*Project),
		Config: EnterpriseConfig{
			MaxProjects: 100,
		},
	}

	pm := NewProjectManagement(manager)
	ctx := context.Background()

	testCases := []struct {
		name string
		req  CreateProjectRequest
	}{
		{
			name: "empty name",
			req: CreateProjectRequest{
				OwnerID: "user-123",
			},
		},
		{
			name: "empty owner",
			req: CreateProjectRequest{
				Name: "Test",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			project, err := pm.CreateProject(ctx, tc.req)

			assert.Error(t, err, "Should error with invalid request")
			assert.Nil(t, project, "Project should be nil on error")
		})
	}
}

// TestGetProject tests project retrieval
func TestGetProject(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		Projects: make(map[string]*Project),
	}

	testProject := &Project{
		ID:      "project-123",
		Name:    "Test Project",
		OwnerID: "user-123",
		Status:  "active",
	}
	manager.Projects["project-123"] = testProject

	pm := NewProjectManagement(manager)
	ctx := context.Background()

	project, err := pm.GetProject(ctx, "project-123")

	assert.NoError(t, err, "Should get project without error")
	assert.NotNil(t, project, "Project should not be nil")
	assert.Equal(t, testProject.ID, project.ID)
}

// TestGetProject_NotFound tests getting nonexistent project
func TestGetProject_NotFound(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		Projects: make(map[string]*Project),
	}

	pm := NewProjectManagement(manager)
	ctx := context.Background()

	project, err := pm.GetProject(ctx, "nonexistent")

	assert.Error(t, err, "Should error when project not found")
	assert.Nil(t, project, "Project should be nil")
}

// TestUpdateProject tests project updates
func TestUpdateProject(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		Projects: make(map[string]*Project),
		Users:    make(map[string]*User),
		AuditLog: []AuditEntry{},
	}

	testProject := &Project{
		ID:          "project-123",
		Name:        "Old Name",
		Description: "Old Description",
		OwnerID:     "user-123",
		MemberIDs:   []string{"user-123"},
		Status:      "active",
	}
	manager.Projects["project-123"] = testProject
	manager.Users["testuser"] = &User{ID: "user-123", Username: "testuser"}

	pm := NewProjectManagement(manager)
	ctx := context.Background()

	newName := "New Name"
	req := UpdateProjectRequest{
		Name: &newName,
	}

	project, err := pm.UpdateProject(ctx, "project-123", req)

	assert.NoError(t, err, "Should update project without error")
	assert.NotNil(t, project, "Project should not be nil")
	assert.Equal(t, "New Name", project.Name)
}

// TestDeleteProject tests project deletion
func TestDeleteProject(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		Projects: make(map[string]*Project),
		Users:    make(map[string]*User),
		AuditLog: []AuditEntry{},
	}

	testProject := &Project{
		ID:        "project-123",
		Name:      "Test Project",
		OwnerID:   "user-123",
		MemberIDs: []string{"user-123"},
		Status:    "active",
	}
	manager.Projects["project-123"] = testProject
	manager.Users["testuser"] = &User{
		ID:         "user-123",
		Username:   "testuser",
		ProjectIDs: []string{"project-123"},
	}

	pm := NewProjectManagement(manager)
	ctx := context.Background()

	err := pm.DeleteProject(ctx, "project-123")

	assert.NoError(t, err, "Should delete project without error")
	project, exists := manager.Projects["project-123"]
	assert.True(t, exists, "Project should still exist (archived)")
	assert.Equal(t, "archived", project.Status, "Project should be archived")
}

// TestListProjects tests project listing
func TestListProjects(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		Projects: make(map[string]*Project),
	}

	// Add multiple projects
	for i := 1; i <= 5; i++ {
		project := &Project{
			ID:      generateTestID(i),
			Name:    generateTestName("Project", i),
			OwnerID: "user-123",
			Status:  "active",
		}
		manager.Projects[project.ID] = project
	}

	pm := NewProjectManagement(manager)
	ctx := context.Background()

	req := ListProjectsRequest{
		Page:     1,
		PageSize: 10,
	}

	response, err := pm.ListProjects(ctx, req)

	assert.NoError(t, err, "Should list projects without error")
	assert.NotNil(t, response, "Response should not be nil")
	assert.Equal(t, 5, response.Total, "Should have 5 total projects")
	assert.GreaterOrEqual(t, len(response.Projects), 5, "Should return projects")
}

// TestListProjects_WithFilter tests listing with filters
func TestListProjects_WithFilter(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		Projects: make(map[string]*Project),
	}

	// Add projects with different owners
	manager.Projects["p1"] = &Project{
		ID:      "p1",
		Name:    "Project 1",
		OwnerID: "user-1",
		Status:  "active",
	}
	manager.Projects["p2"] = &Project{
		ID:      "p2",
		Name:    "Project 2",
		OwnerID: "user-2",
		Status:  "active",
	}

	pm := NewProjectManagement(manager)
	ctx := context.Background()

	req := ListProjectsRequest{
		OwnerID:  "user-1",
		Page:     1,
		PageSize: 10,
	}

	response, err := pm.ListProjects(ctx, req)

	assert.NoError(t, err, "Should list projects without error")
	assert.NotNil(t, response, "Response should not be nil")
}

// ============================================================================
// TEAM MANAGEMENT TESTS
// ============================================================================

// TestNewTeamManagement tests team management creation
func TestNewTeamManagement(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger: *log,
		Teams:  make(map[string]*Team),
	}

	tm := NewTeamManagement(manager)

	assert.NotNil(t, tm, "TeamManagement should not be nil")
	assert.Equal(t, manager, tm.Manager, "Manager should be set")
}

// TestCreateTeam tests team creation
func TestCreateTeam(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		Teams:    make(map[string]*Team),
		Users:    make(map[string]*User),
		AuditLog: []AuditEntry{},
		Config: EnterpriseConfig{
			MaxProjects: 100, // No MaxTeams field exists
		},
	}

	// Add test user
	manager.Users["testuser"] = &User{
		ID:       "user-123",
		Username: "testuser",
		TeamIDs:  []string{},
	}

	tm := NewTeamManagement(manager)
	ctx := context.Background()

	req := CreateTeamRequest{
		Name:        "Test Team",
		Description: "A test team",
		LeadID:      "user-123",
		MemberIDs:   []string{"user-123"},
		Metadata:    map[string]string{"dept": "engineering"},
	}

	team, err := tm.CreateTeam(ctx, req)

	assert.NoError(t, err, "Should create team without error")
	assert.NotNil(t, team, "Team should not be nil")
	assert.Equal(t, req.Name, team.Name)
	assert.Equal(t, req.LeadID, team.LeadID)
	assert.Contains(t, team.MemberIDs, req.LeadID, "Lead should be in members")
}

// TestCreateTeam_InvalidRequest tests with invalid request
func TestCreateTeam_InvalidRequest(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger: *log,
		Teams:  make(map[string]*Team),
		Config: EnterpriseConfig{
			MaxProjects: 100,
		},
	}

	tm := NewTeamManagement(manager)
	ctx := context.Background()

	testCases := []struct {
		name string
		req  CreateTeamRequest
	}{
		{
			name: "empty name",
			req: CreateTeamRequest{
				LeadID: "user-123",
			},
		},
		{
			name: "empty lead",
			req: CreateTeamRequest{
				Name: "Test",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			team, err := tm.CreateTeam(ctx, tc.req)

			assert.Error(t, err, "Should error with invalid request")
			assert.Nil(t, team, "Team should be nil on error")
		})
	}
}

// TestGetTeam tests team retrieval
func TestGetTeam(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger: *log,
		Teams:  make(map[string]*Team),
	}

	testTeam := &Team{
		ID:     "team-123",
		Name:   "Test Team",
		LeadID: "user-123",
	}
	manager.Teams["team-123"] = testTeam

	tm := NewTeamManagement(manager)
	ctx := context.Background()

	team, err := tm.GetTeam(ctx, "team-123")

	assert.NoError(t, err, "Should get team without error")
	assert.NotNil(t, team, "Team should not be nil")
	assert.Equal(t, testTeam.ID, team.ID)
}

// TestGetTeam_NotFound tests getting nonexistent team
func TestGetTeam_NotFound(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger: *log,
		Teams:  make(map[string]*Team),
	}

	tm := NewTeamManagement(manager)
	ctx := context.Background()

	team, err := tm.GetTeam(ctx, "nonexistent")

	assert.Error(t, err, "Should error when team not found")
	assert.Nil(t, team, "Team should be nil")
}

// TestUpdateTeam tests team updates
func TestUpdateTeam(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		Teams:    make(map[string]*Team),
		Users:    make(map[string]*User),
		AuditLog: []AuditEntry{},
	}

	testTeam := &Team{
		ID:          "team-123",
		Name:        "Old Name",
		Description: "Old Description",
		LeadID:      "user-123",
		MemberIDs:   []string{"user-123"},
	}
	manager.Teams["team-123"] = testTeam
	manager.Users["testuser"] = &User{ID: "user-123", Username: "testuser"}

	tm := NewTeamManagement(manager)
	ctx := context.Background()

	newName := "New Name"
	req := UpdateTeamRequest{
		Name: &newName,
	}

	team, err := tm.UpdateTeam(ctx, "team-123", req)

	assert.NoError(t, err, "Should update team without error")
	assert.NotNil(t, team, "Team should not be nil")
	assert.Equal(t, "New Name", team.Name)
}

// TestDeleteTeam tests team deletion
func TestDeleteTeam(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		Teams:    make(map[string]*Team),
		Users:    make(map[string]*User),
		AuditLog: []AuditEntry{},
	}

	testTeam := &Team{
		ID:        "team-123",
		Name:      "Test Team",
		LeadID:    "user-123",
		MemberIDs: []string{"user-123"},
	}
	manager.Teams["team-123"] = testTeam
	manager.Users["testuser"] = &User{
		ID:       "user-123",
		Username: "testuser",
		TeamIDs:  []string{"team-123"},
	}

	tm := NewTeamManagement(manager)
	ctx := context.Background()

	err := tm.DeleteTeam(ctx, "team-123")

	assert.NoError(t, err, "Should delete team without error")
	_, exists := manager.Teams["team-123"]
	assert.False(t, exists, "Team should be deleted")
}

// TestAddTeamMember tests adding member to team
func TestAddTeamMember(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		Teams:    make(map[string]*Team),
		Users:    make(map[string]*User),
		AuditLog: []AuditEntry{},
	}

	testTeam := &Team{
		ID:        "team-123",
		Name:      "Test Team",
		LeadID:    "user-123",
		MemberIDs: []string{"user-123"},
	}
	manager.Teams["team-123"] = testTeam
	manager.Users["newuser"] = &User{
		ID:       "user-456",
		Username: "newuser",
		TeamIDs:  []string{},
	}

	tm := NewTeamManagement(manager)
	ctx := context.Background()

	err := tm.AddTeamMember(ctx, "team-123", "user-456")

	assert.NoError(t, err, "Should add team member without error")
	assert.Contains(t, testTeam.MemberIDs, "user-456", "New member should be in team")
}

// TestRemoveTeamMember tests removing member from team
func TestRemoveTeamMember(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		Teams:    make(map[string]*Team),
		Users:    make(map[string]*User),
		AuditLog: []AuditEntry{},
	}

	testTeam := &Team{
		ID:        "team-123",
		Name:      "Test Team",
		LeadID:    "user-123",
		MemberIDs: []string{"user-123", "user-456"},
	}
	manager.Teams["team-123"] = testTeam
	manager.Users["testuser"] = &User{
		ID:       "user-123",
		Username: "testuser",
		TeamIDs:  []string{"team-123"},
	}

	tm := NewTeamManagement(manager)
	ctx := context.Background()

	err := tm.RemoveTeamMember(ctx, "team-123", "user-456")

	assert.NoError(t, err, "Should remove team member without error")
	assert.NotContains(t, testTeam.MemberIDs, "user-456", "Member should be removed from team")
}

// TestListTeams tests team listing
func TestListTeams(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger: *log,
		Teams:  make(map[string]*Team),
	}

	// Add multiple teams
	for i := 1; i <= 5; i++ {
		team := &Team{
			ID:     generateTestID(i),
			Name:   generateTestName("Team", i),
			LeadID: "user-123",
		}
		manager.Teams[team.ID] = team
	}

	tm := NewTeamManagement(manager)
	ctx := context.Background()

	req := ListTeamsRequest{
		Page:     1,
		PageSize: 10,
	}

	response, err := tm.ListTeams(ctx, req)

	assert.NoError(t, err, "Should list teams without error")
	assert.NotNil(t, response, "Response should not be nil")
	assert.Equal(t, 5, response.Total, "Should have 5 total teams")
	assert.GreaterOrEqual(t, len(response.Teams), 5, "Should return teams")
}

// TestProject_Structure tests Project struct
func TestProject_Structure(t *testing.T) {
	project := &Project{
		ID:          "project-123",
		Name:        "Test Project",
		Description: "Test description",
		OwnerID:     "user-123",
		TeamIDs:     []string{"team-1"},
		MemberIDs:   []string{"user-123", "user-456"},
		Settings:    ProjectSettings{Privacy: "private"},
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    map[string]string{"key": "value"},
	}

	assert.NotNil(t, project, "Project should not be nil")
	assert.Equal(t, "project-123", project.ID)
	assert.Equal(t, 2, len(project.MemberIDs))
	assert.Equal(t, "active", project.Status)
}

// TestTeam_Structure tests Team struct
func TestTeam_Structure(t *testing.T) {
	team := &Team{
		ID:          "team-123",
		Name:        "Test Team",
		Description: "Test description",
		LeadID:      "user-123",
		MemberIDs:   []string{"user-123", "user-456"},
		ProjectIDs:  []string{"project-1"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    map[string]string{"key": "value"},
	}

	assert.NotNil(t, team, "Team should not be nil")
	assert.Equal(t, "team-123", team.ID)
	assert.Equal(t, 2, len(team.MemberIDs))
	assert.Equal(t, 1, len(team.ProjectIDs))
}

// ============================================================================
// HELPER FUNCTION TESTS
// ============================================================================

// TestCanAccessProject tests project access control
func TestCanAccessProject(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		Projects: make(map[string]*Project),
		Teams:    make(map[string]*Team),
		Users:    make(map[string]*User),
	}

	pm := NewProjectManagement(manager)

	// Create test project
	manager.Projects["project-1"] = &Project{
		ID:        "project-1",
		OwnerID:   "owner-1",
		MemberIDs: []string{"member-1", "member-2"},
		TeamIDs:   []string{"team-1"},
	}

	// Create test team
	manager.Teams["team-1"] = &Team{
		ID:        "team-1",
		MemberIDs: []string{"team-member-1"},
	}

	tests := []struct {
		name      string
		projectID string
		userID    string
		expected  bool
	}{
		{"Owner can access", "project-1", "owner-1", true},
		{"Member can access", "project-1", "member-1", true},
		{"Team member can access", "project-1", "team-member-1", true},
		{"Non-member cannot access", "project-1", "random-user", false},
		{"Non-existent project", "non-existent", "owner-1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pm.canAccessProject(context.Background(), tt.projectID, tt.userID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCanAccessTeam tests team access control
func TestCanAccessTeam(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger: *log,
		Teams:  make(map[string]*Team),
	}

	tm := NewTeamManagement(manager)

	// Create test team
	manager.Teams["team-1"] = &Team{
		ID:        "team-1",
		LeadID:    "lead-1",
		MemberIDs: []string{"member-1", "member-2"},
	}

	tests := []struct {
		name     string
		teamID   string
		userID   string
		expected bool
	}{
		{"Lead can access", "team-1", "lead-1", true},
		{"Member can access", "team-1", "member-1", true},
		{"Non-member cannot access", "team-1", "random-user", false},
		{"Non-existent team", "non-existent", "lead-1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tm.canAccessTeam(context.Background(), tt.teamID, tt.userID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetUsernameByID tests username retrieval by ID
func TestGetUsernameByID(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger: *log,
		Users:  make(map[string]*User),
	}

	tm := NewTeamManagement(manager)

	// Add test users
	manager.Users["user1"] = &User{
		ID:       "user-id-1",
		Username: "john_doe",
	}
	manager.Users["user2"] = &User{
		ID:       "user-id-2",
		Username: "jane_smith",
	}

	tests := []struct {
		name     string
		userID   string
		expected string
	}{
		{"Existing user 1", "user-id-1", "john_doe"},
		{"Existing user 2", "user-id-2", "jane_smith"},
		{"Non-existent user", "non-existent", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tm.getUsername(tt.userID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestAppendUnique tests unique append functionality
func TestAppendUnique(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected []string
	}{
		{"Append to empty slice", []string{}, "item1", []string{"item1"}},
		{"Append new item", []string{"item1"}, "item2", []string{"item1", "item2"}},
		{"Duplicate item", []string{"item1", "item2"}, "item1", []string{"item1", "item2"}},
		{"Multiple duplicates", []string{"a", "b", "c"}, "b", []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := appendUnique(tt.slice, tt.item)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestUpdateUserProjects tests user project list updates
func TestUpdateUserProjects(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger: *log,
		Users:  make(map[string]*User),
	}

	// Create test user
	manager.Users["user-1"] = &User{
		ID:         "user-1",
		Username:   "testuser",
		ProjectIDs: []string{},
		UpdatedAt:  time.Now().Add(-1 * time.Hour),
	}

	oldTime := manager.Users["user-1"].UpdatedAt

	// Update user projects
	manager.updateUserProjects("user-1", "project-1")

	assert.Equal(t, 1, len(manager.Users["user-1"].ProjectIDs))
	assert.Contains(t, manager.Users["user-1"].ProjectIDs, "project-1")
	assert.True(t, manager.Users["user-1"].UpdatedAt.After(oldTime))

	// Add duplicate project
	manager.updateUserProjects("user-1", "project-1")
	assert.Equal(t, 1, len(manager.Users["user-1"].ProjectIDs))

	// Update non-existent user
	manager.updateUserProjects("non-existent", "project-1")
	assert.Nil(t, manager.Users["non-existent"])
}

// TestUpdateUserTeams tests user team list updates
func TestUpdateUserTeams(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger: *log,
		Users:  make(map[string]*User),
	}

	// Create test user
	manager.Users["user-1"] = &User{
		ID:        "user-1",
		Username:  "testuser",
		TeamIDs:   []string{},
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}

	oldTime := manager.Users["user-1"].UpdatedAt

	// Update user teams
	manager.updateUserTeams("user-1", "team-1")

	assert.Equal(t, 1, len(manager.Users["user-1"].TeamIDs))
	assert.Contains(t, manager.Users["user-1"].TeamIDs, "team-1")
	assert.True(t, manager.Users["user-1"].UpdatedAt.After(oldTime))

	// Add duplicate team
	manager.updateUserTeams("user-1", "team-1")
	assert.Equal(t, 1, len(manager.Users["user-1"].TeamIDs))

	// Update non-existent user
	manager.updateUserTeams("non-existent", "team-1")
	assert.Nil(t, manager.Users["non-existent"])
}

// TestRemoveUserTeam tests removing team from user
func TestRemoveUserTeam(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger: *log,
		Users:  make(map[string]*User),
	}

	// Create test user with teams
	manager.Users["user-1"] = &User{
		ID:        "user-1",
		Username:  "testuser",
		TeamIDs:   []string{"team-1", "team-2", "team-3"},
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}

	oldTime := manager.Users["user-1"].UpdatedAt

	// Remove team
	manager.removeUserTeam("user-1", "team-2")

	assert.Equal(t, 2, len(manager.Users["user-1"].TeamIDs))
	assert.NotContains(t, manager.Users["user-1"].TeamIDs, "team-2")
	assert.Contains(t, manager.Users["user-1"].TeamIDs, "team-1")
	assert.Contains(t, manager.Users["user-1"].TeamIDs, "team-3")
	assert.True(t, manager.Users["user-1"].UpdatedAt.After(oldTime))

	// Remove non-existent team
	manager.removeUserTeam("user-1", "team-99")
	assert.Equal(t, 2, len(manager.Users["user-1"].TeamIDs))

	// Remove from non-existent user
	manager.removeUserTeam("non-existent", "team-1")
	assert.Nil(t, manager.Users["non-existent"])
}

// Helper functions
func generateTestName(prefix string, i int) string {
	return fmt.Sprintf("%s %d", prefix, i)
}
