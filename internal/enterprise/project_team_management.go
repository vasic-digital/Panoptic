package enterprise

import (
	"context"
	"fmt"
	"time"

	"panoptic/internal/logger"
)

// ProjectManagement handles project operations
type ProjectManagement struct {
	Manager *EnterpriseManager
	Logger   logger.Logger
}

// NewProjectManagement creates new project management handler
func NewProjectManagement(manager *EnterpriseManager) *ProjectManagement {
	return &ProjectManagement{
		Manager: manager,
		Logger:  manager.Logger,
	}
}

// CreateProject creates a new project
func (pm *ProjectManagement) CreateProject(ctx context.Context, req CreateProjectRequest) (*Project, error) {
	// Validate request
	if err := pm.validateCreateProjectRequest(ctx, req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Check project limits
	if pm.Manager.ProjectsExceedLimit() {
		return nil, fmt.Errorf("maximum number of projects reached")
	}

	// Create project
	project := &Project{
		ID:          pm.Manager.generateID(),
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     req.OwnerID,
		TeamIDs:     req.TeamIDs,
		MemberIDs:   req.MemberIDs,
		Settings:    req.Settings,
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    req.Metadata,
	}

	// Ensure owner is in member list
	if !contains(project.MemberIDs, req.OwnerID) {
		project.MemberIDs = append(project.MemberIDs, req.OwnerID)
	}

	// Store project
	pm.Manager.Projects[project.ID] = project

	// Update user's projects
	pm.Manager.updateUserProjects(req.OwnerID, project.ID)

	// Log audit entry
	pm.Manager.logAuditEntry(AuditEntry{
		ID:         pm.Manager.generateID(),
		Timestamp:  time.Now(),
		UserID:     req.OwnerID,
		Username:   pm.getUsername(req.OwnerID),
		Action:     "project.create",
		Resource:   "project",
		ResourceID: project.ID,
		Details:    map[string]string{"name": req.Name, "owner": pm.getUsername(req.OwnerID)},
		Success:    true,
		Severity:   "medium",
		Category:   "data",
	})

	// Save data
	if err := pm.Manager.saveData(); err != nil {
		pm.Logger.Errorf("Failed to save project data: %v", err)
	}

	pm.Logger.Infof("Project created successfully: %s by %s", req.Name, pm.getUsername(req.OwnerID))
	return project, nil
}

// GetProject retrieves a project by ID
func (pm *ProjectManagement) GetProject(ctx context.Context, projectID string) (*Project, error) {
	project, exists := pm.Manager.Projects[projectID]
	if !exists {
		return nil, fmt.Errorf("project not found: %s", projectID)
	}

	return project, nil
}

// UpdateProject updates an existing project
func (pm *ProjectManagement) UpdateProject(ctx context.Context, projectID string, req UpdateProjectRequest) (*Project, error) {
	project, err := pm.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Check permissions
	if !pm.hasProjectPermission(ctx, project, "project.update") {
		return nil, fmt.Errorf("insufficient permissions to update project")
	}

	// Update fields
	if req.Name != nil && *req.Name != "" {
		project.Name = *req.Name
	}
	if req.Description != nil && *req.Description != "" {
		project.Description = *req.Description
	}
	if req.OwnerID != "" {
		project.OwnerID = req.OwnerID
		// Update member list
		if !contains(project.MemberIDs, req.OwnerID) {
			project.MemberIDs = append(project.MemberIDs, req.OwnerID)
		}
	}
	if req.TeamIDs != nil {
		project.TeamIDs = req.TeamIDs
	}
	if req.MemberIDs != nil {
		project.MemberIDs = req.MemberIDs
	}
	if req.Settings != nil {
		project.Settings = *req.Settings
	}
	if req.Metadata != nil {
		project.Metadata = *req.Metadata
	}

	project.UpdatedAt = time.Now()

	// Log audit entry
	pm.Manager.logAuditEntry(AuditEntry{
		ID:         pm.Manager.generateID(),
		Timestamp:  time.Now(),
		UserID:     project.OwnerID,
		Username:   pm.getUsername(project.OwnerID),
		Action:     "project.update",
		Resource:   "project",
		ResourceID: project.ID,
		Details:    map[string]string{"project_name": project.Name},
		Success:    true,
		Severity:   "medium",
		Category:   "data",
	})

	// Save data
	if err := pm.Manager.saveData(); err != nil {
		pm.Logger.Errorf("Failed to save project data: %v", err)
	}

	pm.Logger.Infof("Project updated successfully: %s", project.Name)
	return project, nil
}

// DeleteProject deletes a project
func (pm *ProjectManagement) DeleteProject(ctx context.Context, projectID string) error {
	project, err := pm.GetProject(ctx, projectID)
	if err != nil {
		return err
	}

	// Check permissions
	if !pm.hasProjectPermission(ctx, project, "project.delete") {
		return fmt.Errorf("insufficient permissions to delete project")
	}

	// Archive instead of delete
	now := time.Now()
	project.Status = "archived"
	project.ArchivedAt = &now
	project.UpdatedAt = now

	// Log audit entry
	pm.Manager.logAuditEntry(AuditEntry{
		ID:         pm.Manager.generateID(),
		Timestamp:  time.Now(),
		UserID:     project.OwnerID,
		Username:   pm.getUsername(project.OwnerID),
		Action:     "project.archive",
		Resource:   "project",
		ResourceID: project.ID,
		Details:    map[string]string{"project_name": project.Name},
		Success:    true,
		Severity:   "high",
		Category:   "data",
	})

	// Save data
	if err := pm.Manager.saveData(); err != nil {
		pm.Logger.Errorf("Failed to save project data: %v", err)
	}

	pm.Logger.Infof("Project archived successfully: %s", project.Name)
	return nil
}

// ListProjects lists all projects with filtering
func (pm *ProjectManagement) ListProjects(ctx context.Context, req ListProjectsRequest) (*ListProjectsResponse, error) {
	var projects []Project
	userID := req.UserID

	for _, project := range pm.Manager.Projects {
		// Filter by status
		if req.Status != "" && project.Status != req.Status {
			continue
		}

		// Filter by user access
		if userID != "" && !pm.canAccessProject(ctx, project.ID, userID) {
			continue
		}

		// Filter by owner
		if req.OwnerID != "" && project.OwnerID != req.OwnerID {
			continue
		}

		// Filter by team
		if req.TeamID != "" && !contains(project.TeamIDs, req.TeamID) {
			continue
		}

		projects = append(projects, *project)
	}

	// Apply pagination
	total := len(projects)
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize
	if end > total {
		end = total
	}
	if start >= total {
		return &ListProjectsResponse{
			Projects: []Project{},
			Total:    total,
			Page:     req.Page,
			PageSize: req.PageSize,
		}, nil
	}

	pagedProjects := projects[start:end]

	return &ListProjectsResponse{
		Projects: pagedProjects,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// TeamManagement handles team operations
type TeamManagement struct {
	Manager *EnterpriseManager
	Logger   logger.Logger
}

// NewTeamManagement creates new team management handler
func NewTeamManagement(manager *EnterpriseManager) *TeamManagement {
	return &TeamManagement{
		Manager: manager,
		Logger:  manager.Logger,
	}
}

// CreateTeam creates a new team
func (tm *TeamManagement) CreateTeam(ctx context.Context, req CreateTeamRequest) (*Team, error) {
	// Validate request
	if err := tm.validateCreateTeamRequest(ctx, req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Create team
	team := &Team{
		ID:          tm.Manager.generateID(),
		Name:        req.Name,
		Description: req.Description,
		LeadID:      req.LeadID,
		MemberIDs:   append([]string{req.LeadID}, req.MemberIDs...),
		ProjectIDs:  req.ProjectIDs,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Active:      true,
		Metadata:    req.Metadata,
	}

	// Store team
	tm.Manager.Teams[team.ID] = team

	// Update member teams
	for _, memberID := range team.MemberIDs {
		tm.Manager.updateUserTeams(memberID, team.ID)
	}

	// Log audit entry
	tm.Manager.logAuditEntry(AuditEntry{
		ID:         tm.Manager.generateID(),
		Timestamp:  time.Now(),
		UserID:     req.LeadID,
		Username:   tm.getUsername(req.LeadID),
		Action:     "team.create",
		Resource:   "team",
		ResourceID: team.ID,
		Details:    map[string]string{"name": req.Name, "lead": tm.getUsername(req.LeadID)},
		Success:    true,
		Severity:   "medium",
		Category:   "data",
	})

	// Save data
	if err := tm.Manager.saveData(); err != nil {
		tm.Logger.Errorf("Failed to save team data: %v", err)
	}

	tm.Logger.Infof("Team created successfully: %s", req.Name)
	return team, nil
}

// GetTeam retrieves a team by ID
func (tm *TeamManagement) GetTeam(ctx context.Context, teamID string) (*Team, error) {
	team, exists := tm.Manager.Teams[teamID]
	if !exists {
		return nil, fmt.Errorf("team not found: %s", teamID)
	}

	return team, nil
}

// UpdateTeam updates an existing team
func (tm *TeamManagement) UpdateTeam(ctx context.Context, teamID string, req UpdateTeamRequest) (*Team, error) {
	team, err := tm.GetTeam(ctx, teamID)
	if err != nil {
		return nil, err
	}

	// Check permissions
	if !tm.hasTeamPermission(ctx, team, "team.update") {
		return nil, fmt.Errorf("insufficient permissions to update team")
	}

	// Update fields
	if req.Name != nil && *req.Name != "" {
		team.Name = *req.Name
	}
	if req.Description != nil && *req.Description != "" {
		team.Description = *req.Description
	}
	if req.LeadID != "" {
		team.LeadID = req.LeadID
		if !contains(team.MemberIDs, req.LeadID) {
			team.MemberIDs = append(team.MemberIDs, req.LeadID)
		}
	}
	if req.MemberIDs != nil {
		team.MemberIDs = req.MemberIDs
	}
	if req.ProjectIDs != nil {
		team.ProjectIDs = req.ProjectIDs
	}
	if req.Metadata != nil {
		team.Metadata = *req.Metadata
	}

	team.UpdatedAt = time.Now()

	// Log audit entry
	tm.Manager.logAuditEntry(AuditEntry{
		ID:         tm.Manager.generateID(),
		Timestamp:  time.Now(),
		UserID:     team.LeadID,
		Username:   tm.getUsername(team.LeadID),
		Action:     "team.update",
		Resource:   "team",
		ResourceID: team.ID,
		Details:    map[string]string{"team_name": team.Name},
		Success:    true,
		Severity:   "medium",
		Category:   "data",
	})

	// Save data
	if err := tm.Manager.saveData(); err != nil {
		tm.Logger.Errorf("Failed to save team data: %v", err)
	}

	tm.Logger.Infof("Team updated successfully: %s", team.Name)
	return team, nil
}

// DeleteTeam deletes a team
func (tm *TeamManagement) DeleteTeam(ctx context.Context, teamID string) error {
	team, err := tm.GetTeam(ctx, teamID)
	if err != nil {
		return err
	}

	// Check permissions
	if !tm.hasTeamPermission(ctx, team, "team.delete") {
		return fmt.Errorf("insufficient permissions to delete team")
	}

	// Delete team
	delete(tm.Manager.Teams, team.ID)

	// Update member teams
	for _, memberID := range team.MemberIDs {
		tm.Manager.removeUserTeam(memberID, teamID)
	}

	// Log audit entry
	tm.Manager.logAuditEntry(AuditEntry{
		ID:         tm.Manager.generateID(),
		Timestamp:  time.Now(),
		UserID:     team.LeadID,
		Username:   tm.getUsername(team.LeadID),
		Action:     "team.delete",
		Resource:   "team",
		ResourceID: team.ID,
		Details:    map[string]string{"team_name": team.Name},
		Success:    true,
		Severity:   "high",
		Category:   "data",
	})

	// Save data
	if err := tm.Manager.saveData(); err != nil {
		tm.Logger.Errorf("Failed to save team data: %v", err)
	}

	tm.Logger.Infof("Team deleted successfully: %s", team.Name)
	return nil
}

// AddTeamMember adds a member to a team
func (tm *TeamManagement) AddTeamMember(ctx context.Context, teamID, userID string) error {
	team, err := tm.GetTeam(ctx, teamID)
	if err != nil {
		return err
	}

	// Check permissions
	if !tm.hasTeamPermission(ctx, team, "team.update") {
		return fmt.Errorf("insufficient permissions to update team")
	}

	// Check if user is already a member
	if contains(team.MemberIDs, userID) {
		return fmt.Errorf("user is already a team member")
	}

	// Add member
	team.MemberIDs = append(team.MemberIDs, userID)
	team.UpdatedAt = time.Now()

	// Update user teams
	tm.Manager.updateUserTeams(userID, teamID)

	// Log audit entry
	tm.Manager.logAuditEntry(AuditEntry{
		ID:         tm.Manager.generateID(),
		Timestamp:  time.Now(),
		UserID:     userID,
		Username:   tm.getUsername(userID),
		Action:     "team.member.add",
		Resource:   "team",
		ResourceID: team.ID,
		Details:    map[string]string{"team_name": team.Name, "user": tm.getUsername(userID)},
		Success:    true,
		Severity:   "low",
		Category:   "data",
	})

	// Save data
	if err := tm.Manager.saveData(); err != nil {
		tm.Logger.Errorf("Failed to save team data: %v", err)
	}

	tm.Logger.Infof("Member added to team successfully: %s -> %s", tm.getUsername(userID), team.Name)
	return nil
}

// RemoveTeamMember removes a member from a team
func (tm *TeamManagement) RemoveTeamMember(ctx context.Context, teamID, userID string) error {
	team, err := tm.GetTeam(ctx, teamID)
	if err != nil {
		return err
	}

	// Check permissions
	if !tm.hasTeamPermission(ctx, team, "team.update") {
		return fmt.Errorf("insufficient permissions to update team")
	}

	// Check if user is a member
	if !contains(team.MemberIDs, userID) {
		return fmt.Errorf("user is not a team member")
	}

	// Cannot remove team lead
	if team.LeadID == userID {
		return fmt.Errorf("cannot remove team lead")
	}

	// Remove member
	team.MemberIDs = remove(team.MemberIDs, userID)
	team.UpdatedAt = time.Now()

	// Update user teams
	tm.Manager.removeUserTeam(userID, teamID)

	// Log audit entry
	tm.Manager.logAuditEntry(AuditEntry{
		ID:         tm.Manager.generateID(),
		Timestamp:  time.Now(),
		UserID:     userID,
		Username:   tm.getUsername(userID),
		Action:     "team.member.remove",
		Resource:   "team",
		ResourceID: team.ID,
		Details:    map[string]string{"team_name": team.Name, "user": tm.getUsername(userID)},
		Success:    true,
		Severity:   "low",
		Category:   "data",
	})

	// Save data
	if err := tm.Manager.saveData(); err != nil {
		tm.Logger.Errorf("Failed to save team data: %v", err)
	}

	tm.Logger.Infof("Member removed from team successfully: %s -> %s", tm.getUsername(userID), team.Name)
	return nil
}

// ListTeams lists all teams with filtering
func (tm *TeamManagement) ListTeams(ctx context.Context, req ListTeamsRequest) (*ListTeamsResponse, error) {
	var teams []Team
	userID := req.UserID

	for _, team := range tm.Manager.Teams {
		// Filter by user access
		if userID != "" && !tm.canAccessTeam(ctx, team.ID, userID) {
			continue
		}

		// Filter by lead
		if req.LeadID != "" && team.LeadID != req.LeadID {
			continue
		}

		// Filter by project
		if req.ProjectID != "" && !contains(team.ProjectIDs, req.ProjectID) {
			continue
		}

		teams = append(teams, *team)
	}

	// Apply pagination
	total := len(teams)
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize
	if end > total {
		end = total
	}
	if start >= total {
		return &ListTeamsResponse{
			Teams: []Team{},
			Total: total,
			Page:  req.Page,
			PageSize: req.PageSize,
		}, nil
	}

	pagedTeams := teams[start:end]

	return &ListTeamsResponse{
		Teams: pagedTeams,
		Total: total,
		Page:  req.Page,
		PageSize: req.PageSize,
	}, nil
}

// Request types

type CreateProjectRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	OwnerID     string            `json:"owner_id"`
	TeamIDs     []string          `json:"team_ids"`
	MemberIDs   []string          `json:"member_ids"`
	Settings    ProjectSettings    `json:"settings"`
	Metadata    map[string]string `json:"metadata"`
}

type UpdateProjectRequest struct {
	Name        *string             `json:"name,omitempty"`
	Description *string             `json:"description,omitempty"`
	OwnerID     string              `json:"owner_id,omitempty"`
	TeamIDs     []string            `json:"team_ids,omitempty"`
	MemberIDs   []string            `json:"member_ids,omitempty"`
	Settings    *ProjectSettings    `json:"settings,omitempty"`
	Metadata    *map[string]string  `json:"metadata,omitempty"`
}

type ListProjectsRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Status   string `json:"status,omitempty"`
	OwnerID  string `json:"owner_id,omitempty"`
	TeamID   string `json:"team_id,omitempty"`
	UserID   string `json:"user_id,omitempty"`
}

type ListProjectsResponse struct {
	Projects []Project `json:"projects"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
}

type CreateTeamRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	LeadID      string            `json:"lead_id"`
	MemberIDs   []string          `json:"member_ids"`
	ProjectIDs  []string          `json:"project_ids"`
	Metadata    map[string]string `json:"metadata"`
}

type UpdateTeamRequest struct {
	Name        *string            `json:"name,omitempty"`
	Description *string            `json:"description,omitempty"`
	LeadID      string             `json:"lead_id,omitempty"`
	MemberIDs   []string           `json:"member_ids,omitempty"`
	ProjectIDs  []string           `json:"project_ids,omitempty"`
	Metadata    *map[string]string  `json:"metadata,omitempty"`
}

type ListTeamsRequest struct {
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
	LeadID    string `json:"lead_id,omitempty"`
	ProjectID string `json:"project_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
}

type ListTeamsResponse struct {
	Teams []Team `json:"teams"`
	Total int    `json:"total"`
	Page  int    `json:"page"`
	PageSize int    `json:"page_size"`
}

// Helper methods

func (pm *ProjectManagement) validateCreateProjectRequest(ctx context.Context, req CreateProjectRequest) error {
	if req.Name == "" {
		return fmt.Errorf("project name is required")
	}
	if req.OwnerID == "" {
		return fmt.Errorf("project owner is required")
	}
	return nil
}

func (pm *ProjectManagement) hasProjectPermission(ctx context.Context, project *Project, permission string) bool {
	// Check if user has permission (simplified for demo)
	return true
}

func (pm *ProjectManagement) getUsername(userID string) string {
	for _, user := range pm.Manager.Users {
		if user.ID == userID {
			return user.Username
		}
	}
	return "unknown"
}

func (tm *TeamManagement) getUsername(userID string) string {
	for _, user := range tm.Manager.Users {
		if user.ID == userID {
			return user.Username
		}
	}
	return "unknown"
}

func (pm *ProjectManagement) canAccessProject(ctx context.Context, projectID, userID string) bool {
	project, exists := pm.Manager.Projects[projectID]
	if !exists {
		return false
	}

	// Check if user is owner or member
	if project.OwnerID == userID || contains(project.MemberIDs, userID) {
		return true
	}

	// Check if user is in any team with access
	for _, teamID := range project.TeamIDs {
		team, teamExists := pm.Manager.Teams[teamID]
		if teamExists && contains(team.MemberIDs, userID) {
			return true
		}
	}

	return false
}

func (tm *TeamManagement) validateCreateTeamRequest(ctx context.Context, req CreateTeamRequest) error {
	if req.Name == "" {
		return fmt.Errorf("team name is required")
	}
	if req.LeadID == "" {
		return fmt.Errorf("team lead is required")
	}
	return nil
}

func (tm *TeamManagement) hasTeamPermission(ctx context.Context, team *Team, permission string) bool {
	// Check if user has permission (simplified for demo)
	return true
}

func (tm *TeamManagement) canAccessTeam(ctx context.Context, teamID, userID string) bool {
	team, exists := tm.Manager.Teams[teamID]
	if !exists {
		return false
	}

	// Check if user is lead or member
	return team.LeadID == userID || contains(team.MemberIDs, userID)
}

func (em *EnterpriseManager) updateUserProjects(userID, projectID string) {
	if user, exists := em.Users[userID]; exists {
		user.ProjectIDs = appendUnique(user.ProjectIDs, projectID)
		user.UpdatedAt = time.Now()
	}
}

func (em *EnterpriseManager) updateUserTeams(userID, teamID string) {
	if user, exists := em.Users[userID]; exists {
		user.TeamIDs = appendUnique(user.TeamIDs, teamID)
		user.UpdatedAt = time.Now()
	}
}

func (em *EnterpriseManager) removeUserTeam(userID, teamID string) {
	if user, exists := em.Users[userID]; exists {
		user.TeamIDs = remove(user.TeamIDs, teamID)
		user.UpdatedAt = time.Now()
	}
}

func (em *EnterpriseManager) getUsername(userID string) string {
	for _, user := range em.Users {
		if user.ID == userID {
			return user.Username
		}
	}
	return "unknown"
}

func (em *EnterpriseManager) ProjectsExceedLimit() bool {
	if em.Config.MaxProjects <= 0 {
		return false
	}

	count := 0
	for _, project := range em.Projects {
		if project.Status == "active" {
			count++
		}
	}

	return count >= em.Config.MaxProjects
}

func appendUnique(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}

func remove(slice []string, item string) []string {
	for i, s := range slice {
		if s == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}