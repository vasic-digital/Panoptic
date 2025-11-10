package enterprise

import (
	"context"
	"fmt"
	"time"

	"panoptic/internal/logger"
)

// AuditManagement handles audit and compliance operations
type AuditManagement struct {
	Manager *EnterpriseManager
	Logger   logger.Logger
}

// NewAuditManagement creates new audit management handler
func NewAuditManagement(manager *EnterpriseManager) *AuditManagement {
	return &AuditManagement{
		Manager: manager,
		Logger:  manager.Logger,
	}
}

// GetAuditLog retrieves audit log entries with filtering
func (am *AuditManagement) GetAuditLog(ctx context.Context, req GetAuditLogRequest) (*GetAuditLogResponse, error) {
	var entries []AuditEntry

	for _, entry := range am.Manager.AuditLog {
		// Apply filters
		if req.UserID != "" && entry.UserID != req.UserID {
			continue
		}
		if req.Username != "" && entry.Username != req.Username {
			continue
		}
		if req.Action != "" && entry.Action != req.Action {
			continue
		}
		if req.Resource != "" && entry.Resource != req.Resource {
			continue
		}
		if req.Category != "" && entry.Category != req.Category {
			continue
		}
		if req.Severity != "" && entry.Severity != req.Severity {
			continue
		}
		if !req.StartTime.IsZero() && entry.Timestamp.Before(req.StartTime) {
			continue
		}
		if !req.EndTime.IsZero() && entry.Timestamp.After(req.EndTime) {
			continue
		}
		if req.Success != nil && *req.Success != entry.Success {
			continue
		}

		entries = append(entries, entry)
	}

	// Sort by timestamp (newest first)
	am.sortAuditEntries(entries, false)

	// Apply pagination
	total := len(entries)
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize
	if end > total {
		end = total
	}
	if start >= total {
		return &GetAuditLogResponse{
			Entries: []AuditEntry{},
			Total:   total,
			Page:    req.Page,
			PageSize: req.PageSize,
		}, nil
	}

	pagedEntries := entries[start:end]

	return &GetAuditLogResponse{
		Entries: pagedEntries,
		Total:   total,
		Page:    req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetAuditSummary retrieves audit log summary statistics
func (am *AuditManagement) GetAuditSummary(ctx context.Context, req GetAuditSummaryRequest) (*AuditSummaryResponse, error) {
	var filteredEntries []AuditEntry

	for _, entry := range am.Manager.AuditLog {
		// Apply date filter
		if !req.StartTime.IsZero() && entry.Timestamp.Before(req.StartTime) {
			continue
		}
		if !req.EndTime.IsZero() && entry.Timestamp.After(req.EndTime) {
			continue
		}

		filteredEntries = append(filteredEntries, entry)
	}

	summary := &AuditSummaryResponse{
		TotalEntries:   len(filteredEntries),
		SuccessCount:   0,
		FailureCount:   0,
		CategoryStats:  make(map[string]int),
		ActionStats:    make(map[string]int),
		SeverityStats:  make(map[string]int),
		UserStats:      make(map[string]int),
		HourlyStats:   make(map[string]int),
		DailyStats:     make(map[string]int),
	}

	for _, entry := range filteredEntries {
		// Count success/failure
		if entry.Success {
			summary.SuccessCount++
		} else {
			summary.FailureCount++
		}

		// Category statistics
		summary.CategoryStats[entry.Category]++

		// Action statistics
		summary.ActionStats[entry.Action]++

		// Severity statistics
		summary.SeverityStats[entry.Severity]++

		// User statistics
		summary.UserStats[entry.Username]++

		// Hourly statistics
		hour := entry.Timestamp.Format("2006-01-02T15")
		summary.HourlyStats[hour]++

		// Daily statistics
		day := entry.Timestamp.Format("2006-01-02")
		summary.DailyStats[day]++
	}

	summary.SuccessRate = float64(summary.SuccessCount) / float64(len(filteredEntries)) * 100
	summary.FailureRate = float64(summary.FailureCount) / float64(len(filteredEntries)) * 100

	return summary, nil
}

// GetComplianceStatus retrieves compliance status and reports
func (am *AuditManagement) GetComplianceStatus(ctx context.Context, req GetComplianceStatusRequest) (*ComplianceStatusResponse, error) {
	if !am.Manager.Config.Compliance.Enabled {
		return nil, fmt.Errorf("compliance features are not enabled")
	}

	response := &ComplianceStatusResponse{
		Standards:      am.Manager.Config.Compliance.Standards,
		DataRetention:   am.Manager.Config.Compliance.DataRetention,
		AuditRetention:  am.Manager.Config.Compliance.AuditRetention,
		DataEncryption: am.Manager.Config.Compliance.DataEncryption,
		AuditEncryption: am.Manager.Config.Compliance.AuditEncryption,
		RequireApproval: am.Manager.Config.Compliance.RequireApproval,
		ApprovalWorkflow: am.Manager.Config.Compliance.ApprovalWorkflow,
		Reports:       make(map[string]ComplianceReport),
		LastAssessment: time.Now().AddDate(0, 0, -30), // 30 days ago
		NextAssessment: time.Now().AddDate(0, 1, 0), // 1 month from now
		Status:        "compliant",
		Issues:        []ComplianceIssue{},
	}

	// Generate compliance reports for each standard
	for _, standard := range am.Manager.Config.Compliance.Standards {
		report := am.generateComplianceReport(standard)
		response.Reports[standard] = report

		if report.Status != "compliant" {
			response.Status = "non_compliant"
			response.Issues = append(response.Issues, report.Issues...)
		}
	}

	return response, nil
}

// ExportAuditLog exports audit log entries in various formats
func (am *AuditManagement) ExportAuditLog(ctx context.Context, req ExportAuditLogRequest) (*ExportAuditLogResponse, error) {
	// Get filtered entries
	getReq := GetAuditLogRequest{
		UserID:    req.UserID,
		Username:  req.Username,
		Action:    req.Action,
		Resource:  req.Resource,
		Category:  req.Category,
		Severity:  req.Severity,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Success:   req.Success,
		Page:      1,
		PageSize:  100000, // Get all entries
	}

	auditResponse, err := am.GetAuditLog(ctx, getReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit entries: %w", err)
	}

	// Export based on format
	var exportData string
	var contentType string
	var fileExtension string

	switch req.Format {
	case "json":
		exportData = am.exportToJSON(auditResponse.Entries)
		contentType = "application/json"
		fileExtension = "json"
	case "csv":
		exportData = am.exportToCSV(auditResponse.Entries)
		contentType = "text/csv"
		fileExtension = "csv"
	case "xml":
		exportData = am.exportToXML(auditResponse.Entries)
		contentType = "application/xml"
		fileExtension = "xml"
	default:
		return nil, fmt.Errorf("unsupported export format: %s", req.Format)
	}

	// Create export file
	filename := fmt.Sprintf("audit_log_%s.%s", time.Now().Format("20060102_150405"), fileExtension)
	backupFilePath := fmt.Sprintf("%s/%s", am.Manager.Config.StoragePath, filename)

	// In a real implementation, write to file and return download URL
	am.Logger.Infof("Audit log exported: %s (%d entries, %s format)", backupFilePath, len(auditResponse.Entries), req.Format)

	return &ExportAuditLogResponse{
		Filename:     filename,
		ContentType:  contentType,
		Size:         len(exportData),
		EntriesCount: len(auditResponse.Entries),
		ExportedAt:   time.Now(),
		ExpiresAt:    time.Now().Add(24 * time.Hour), // 24 hours
	}, nil
}

// CreateComplianceReport generates a compliance report
func (am *AuditManagement) CreateComplianceReport(ctx context.Context, req CreateComplianceReportRequest) (*ComplianceReport, error) {
	if !am.Manager.Config.Compliance.Enabled {
		return nil, fmt.Errorf("compliance features are not enabled")
	}

	report := am.generateComplianceReport(req.Standard)

	// Log audit entry
	am.Manager.logAuditEntry(AuditEntry{
		ID:        am.Manager.generateID(),
		Timestamp: time.Now(),
		Action:    "compliance.report.create",
		Resource:  "compliance_report",
		Details:   map[string]string{"standard": req.Standard, "status": report.Status},
		Success:   true,
		Severity:  "low",
		Category:  "system",
	})

	am.Logger.Infof("Compliance report generated: %s - %s", req.Standard, report.Status)
	return &report, nil
}

// GetRetentionStatus retrieves data retention status
func (am *AuditManagement) GetRetentionStatus(ctx context.Context, req GetRetentionStatusRequest) (*RetentionStatusResponse, error) {
	response := &RetentionStatusResponse{
		DataRetention:   am.Manager.Config.Compliance.DataRetention,
		AuditRetention:  am.Manager.Config.Compliance.AuditRetention,
		LastCleanup:    time.Now().AddDate(0, 0, -1), // 1 day ago
		NextCleanup:    time.Now().AddDate(0, 0, 1), // 1 day from now
		DataStats:      make(map[string]RetentionStats),
		AuditStats:     RetentionStats{
			TotalEntries:   len(am.Manager.AuditLog),
			EntriesToRetain: am.calculateEntriesToRetain(),
			EntriesToDelete: am.calculateEntriesToDelete(),
			OldestEntry:   am.getOldestAuditEntry(),
			NewestEntry:   am.getNewestAuditEntry(),
		},
	}

	// Calculate data retention stats for different data types
	dataTypes := []string{"users", "projects", "teams", "test_results", "reports"}
	for _, dataType := range dataTypes {
		stats := am.calculateDataRetentionStats(dataType)
		response.DataStats[dataType] = stats
	}

	return response, nil
}

// ExecuteCleanup executes data cleanup based on retention policies
func (am *AuditManagement) ExecuteCleanup(ctx context.Context, req ExecuteCleanupRequest) (*ExecuteCleanupResponse, error) {
	response := &ExecuteCleanupResponse{
		StartedAt: time.Now(),
		Results:   make(map[string]CleanupResult),
	}

	// Clean up audit log
	if req.IncludeAudit {
		result := am.cleanupAuditLog(req.DryRun)
		response.Results["audit"] = result
	}

	// Clean up data
	dataTypes := []string{"users", "projects", "teams", "test_results", "reports"}
	for _, dataType := range dataTypes {
		if req.IncludeData {
			result := am.cleanupData(dataType, req.DryRun)
			response.Results[dataType] = result
		}
	}

	response.CompletedAt = time.Now()
	response.Duration = response.CompletedAt.Sub(response.StartedAt)

	// Log audit entry
	am.Manager.logAuditEntry(AuditEntry{
		ID:        am.Manager.generateID(),
		Timestamp: response.CompletedAt,
		Action:    "cleanup.execute",
		Resource:  "retention_policy",
		Details: map[string]string{
			"dry_run": fmt.Sprintf("%t", req.DryRun),
			"duration": response.Duration.String(),
		},
		Success:  true,
		Severity: "medium",
		Category: "system",
	})

	am.Logger.Infof("Data cleanup executed: %s (dry_run: %t, duration: %s)", 
		response.Duration.String(), req.DryRun, req.DryRun)

	return response, nil
}

// Request types

type GetAuditLogRequest struct {
	Page      int        `json:"page"`
	PageSize  int        `json:"page_size"`
	UserID    string     `json:"user_id,omitempty"`
	Username  string     `json:"username,omitempty"`
	Action    string     `json:"action,omitempty"`
	Resource  string     `json:"resource,omitempty"`
	Category  string     `json:"category,omitempty"`
	Severity  string     `json:"severity,omitempty"`
	StartTime time.Time  `json:"start_time,omitempty"`
	EndTime   time.Time  `json:"end_time,omitempty"`
	Success   *bool      `json:"success,omitempty"`
}

type GetAuditLogResponse struct {
	Entries []AuditEntry `json:"entries"`
	Total   int          `json:"total"`
	Page    int          `json:"page"`
	PageSize int          `json:"page_size"`
}

type GetAuditSummaryRequest struct {
	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
}

type AuditSummaryResponse struct {
	TotalEntries    int               `json:"total_entries"`
	SuccessCount    int               `json:"success_count"`
	FailureCount    int               `json:"failure_count"`
	SuccessRate    float64           `json:"success_rate"`
	FailureRate    float64           `json:"failure_rate"`
	CategoryStats  map[string]int    `json:"category_stats"`
	ActionStats    map[string]int    `json:"action_stats"`
	SeverityStats  map[string]int    `json:"severity_stats"`
	UserStats      map[string]int    `json:"user_stats"`
	HourlyStats   map[string]int    `json:"hourly_stats"`
	DailyStats    map[string]int    `json:"daily_stats"`
}

type GetComplianceStatusRequest struct {
	Standards []string `json:"standards,omitempty"`
}

type ComplianceStatusResponse struct {
	Standards        []string                    `json:"standards"`
	DataRetention     int                         `json:"data_retention"`
	AuditRetention    int                         `json:"audit_retention"`
	DataEncryption    bool                        `json:"data_encryption"`
	AuditEncryption   bool                        `json:"audit_encryption"`
	RequireApproval  bool                        `json:"require_approval"`
	ApprovalWorkflow string                      `json:"approval_workflow"`
	Reports          map[string]ComplianceReport `json:"reports"`
	LastAssessment   time.Time                   `json:"last_assessment"`
	NextAssessment   time.Time                   `json:"next_assessment"`
	Status           string                      `json:"status"`        // compliant, non_compliant, in_progress
	Issues           []ComplianceIssue           `json:"issues"`
}

type ComplianceReport struct {
	Standard      string              `json:"standard"`
	Status        string              `json:"status"`
	LastAssessed  time.Time           `json:"last_assessed"`
	Score         int                 `json:"score"`
	MaxScore      int                 `json:"max_score"`
	Requirements  []ComplianceRequirement `json:"requirements"`
	Issues        []ComplianceIssue  `json:"issues"`
	Recommendations []string         `json:"recommendations"`
}

type ComplianceRequirement struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Mandatory   bool   `json:"mandatory"`
	Satisfied   bool   `json:"satisfied"`
	Score       int    `json:"score"`
	MaxScore    int    `json:"max_score"`
}

type ComplianceIssue struct {
	ID          string    `json:"id"`
	Requirement  string    `json:"requirement"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	Status      string    `json:"status"`        // open, in_progress, resolved
	CreatedAt   time.Time `json:"created_at"`
	ResolvedAt   *time.Time `json:"resolved_at,omitempty"`
}

type ExportAuditLogRequest struct {
	Format     string     `json:"format"`       // json, csv, xml
	UserID     string     `json:"user_id,omitempty"`
	Username   string     `json:"username,omitempty"`
	Action     string     `json:"action,omitempty"`
	Resource   string     `json:"resource,omitempty"`
	Category   string     `json:"category,omitempty"`
	Severity   string     `json:"severity,omitempty"`
	StartTime  time.Time  `json:"start_time,omitempty"`
	EndTime    time.Time  `json:"end_time,omitempty"`
	Success    *bool      `json:"success,omitempty"`
}

type ExportAuditLogResponse struct {
	Filename     string    `json:"filename"`
	ContentType  string    `json:"content_type"`
	Size         int       `json:"size"`
	EntriesCount int       `json:"entries_count"`
	ExportedAt   time.Time `json:"exported_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type CreateComplianceReportRequest struct {
	Standard string `json:"standard"`
}

type GetRetentionStatusRequest struct {
	DataTypes []string `json:"data_types,omitempty"`
}

type RetentionStatusResponse struct {
	DataRetention  int                         `json:"data_retention"`
	AuditRetention int                         `json:"audit_retention"`
	LastCleanup    time.Time                   `json:"last_cleanup"`
	NextCleanup    time.Time                   `json:"next_cleanup"`
	DataStats      map[string]RetentionStats  `json:"data_stats"`
	AuditStats     RetentionStats              `json:"audit_stats"`
}

type RetentionStats struct {
	TotalEntries    int       `json:"total_entries"`
	EntriesToRetain int       `json:"entries_to_retain"`
	EntriesToDelete int       `json:"entries_to_delete"`
	OldestEntry    time.Time `json:"oldest_entry"`
	NewestEntry    time.Time `json:"newest_entry"`
}

type ExecuteCleanupRequest struct {
	IncludeAudit bool `json:"include_audit"`
	IncludeData  bool `json:"include_data"`
	DryRun      bool `json:"dry_run"`
}

type ExecuteCleanupResponse struct {
	StartedAt   time.Time                  `json:"started_at"`
	CompletedAt time.Time                  `json:"completed_at"`
	Duration    time.Duration              `json:"duration"`
	Results     map[string]CleanupResult  `json:"results"`
}

type CleanupResult struct {
	ProcessedCount int    `json:"processed_count"`
	DeletedCount   int    `json:"deleted_count"`
	ErrorCount    int    `json:"error_count"`
	Errors        []string `json:"errors,omitempty"`
}

// Helper methods

func (am *AuditManagement) sortAuditEntries(entries []AuditEntry, ascending bool) {
	if ascending {
		for i := 0; i < len(entries)-1; i++ {
			for j := i + 1; j < len(entries); j++ {
				if entries[i].Timestamp.After(entries[j].Timestamp) {
					entries[i], entries[j] = entries[j], entries[i]
				}
			}
		}
	} else {
		for i := 0; i < len(entries)-1; i++ {
			for j := i + 1; j < len(entries); j++ {
				if entries[i].Timestamp.Before(entries[j].Timestamp) {
					entries[i], entries[j] = entries[j], entries[i]
				}
			}
		}
	}
}

func (am *AuditManagement) generateComplianceReport(standard string) ComplianceReport {
	// Generate mock compliance report for demo
	report := ComplianceReport{
		Standard:     standard,
		Status:       "compliant",
		LastAssessed: time.Now().AddDate(0, 0, -30),
		Score:        85,
		MaxScore:     100,
		Requirements: []ComplianceRequirement{
			{
				ID:          "REQ1",
				Name:        "Access Control",
				Description: "Proper access control mechanisms",
				Mandatory:   true,
				Satisfied:   true,
				Score:       25,
				MaxScore:    25,
			},
			{
				ID:          "REQ2",
				Name:        "Audit Logging",
				Description: "Comprehensive audit logging",
				Mandatory:   true,
				Satisfied:   true,
				Score:       30,
				MaxScore:    30,
			},
			{
				ID:          "REQ3",
				Name:        "Data Encryption",
				Description: "Data encryption at rest and in transit",
				Mandatory:   true,
				Satisfied:   true,
				Score:       30,
				MaxScore:    30,
			},
		},
		Issues: []ComplianceIssue{},
		Recommendations: []string{
			"Continue regular compliance assessments",
			"Monitor audit log retention policies",
		},
	}

	return report
}

func (am *AuditManagement) exportToJSON(entries []AuditEntry) string {
	// Simplified JSON export
	result := "[\n"
	for i, entry := range entries {
		result += fmt.Sprintf(`  {
    "id": "%s",
    "timestamp": "%s",
    "user_id": "%s",
    "username": "%s",
    "action": "%s",
    "resource": "%s",
    "success": %t,
    "severity": "%s"
  }`, entry.ID, entry.Timestamp.Format(time.RFC3339), entry.UserID, entry.Username, entry.Action, entry.Resource, entry.Success, entry.Severity)
		if i < len(entries)-1 {
			result += ","
		}
		result += "\n"
	}
	result += "\n]"
	return result
}

func (am *AuditManagement) exportToCSV(entries []AuditEntry) string {
	result := "timestamp,user_id,username,action,resource,success,severity\n"
	for _, entry := range entries {
		result += fmt.Sprintf("%s,%s,%s,%s,%s,%t,%s\n",
			entry.Timestamp.Format(time.RFC3339), entry.UserID, entry.Username, entry.Action, entry.Resource, entry.Success, entry.Severity)
	}
	return result
}

func (am *AuditManagement) exportToXML(entries []AuditEntry) string {
	// Simplified XML export
	result := `<?xml version="1.0" encoding="UTF-8"?>
<audit_entries>`
	for _, entry := range entries {
		result += fmt.Sprintf(`
  <audit_entry>
    <id>%s</id>
    <timestamp>%s</timestamp>
    <user_id>%s</user_id>
    <username>%s</username>
    <action>%s</action>
    <resource>%s</resource>
    <success>%t</success>
    <severity>%s</severity>
  </audit_entry>`, entry.ID, entry.Timestamp.Format(time.RFC3339), entry.UserID, entry.Username, entry.Action, entry.Resource, entry.Success, entry.Severity)
	}
	result += `
</audit_entries>`
	return result
}

func (am *AuditManagement) calculateEntriesToRetain() int {
	retentionDays := am.Manager.Config.Compliance.AuditRetention
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
	count := 0
	for _, entry := range am.Manager.AuditLog {
		if entry.Timestamp.After(cutoffTime) {
			count++
		}
	}
	return count
}

func (am *AuditManagement) calculateEntriesToDelete() int {
	total := len(am.Manager.AuditLog)
	return total - am.calculateEntriesToRetain()
}

func (am *AuditManagement) getOldestAuditEntry() time.Time {
	if len(am.Manager.AuditLog) == 0 {
		return time.Time{}
	}
	oldest := am.Manager.AuditLog[0].Timestamp
	for _, entry := range am.Manager.AuditLog {
		if entry.Timestamp.Before(oldest) {
			oldest = entry.Timestamp
		}
	}
	return oldest
}

func (am *AuditManagement) getNewestAuditEntry() time.Time {
	if len(am.Manager.AuditLog) == 0 {
		return time.Time{}
	}
	newest := am.Manager.AuditLog[0].Timestamp
	for _, entry := range am.Manager.AuditLog {
		if entry.Timestamp.After(newest) {
			newest = entry.Timestamp
		}
	}
	return newest
}

func (am *AuditManagement) calculateDataRetentionStats(dataType string) RetentionStats {
	// Mock calculation for demo
	return RetentionStats{
		TotalEntries:    1000,
		EntriesToRetain: 800,
		EntriesToDelete: 200,
		OldestEntry:    time.Now().AddDate(0, 0, -90),
		NewestEntry:    time.Now().AddDate(0, 0, -1),
	}
}

func (am *AuditManagement) cleanupAuditLog(dryRun bool) CleanupResult {
	retentionDays := am.Manager.Config.Compliance.AuditRetention
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
	
	var entriesToDelete []AuditEntry
	for _, entry := range am.Manager.AuditLog {
		if entry.Timestamp.Before(cutoffTime) {
			entriesToDelete = append(entriesToDelete, entry)
		}
	}

	result := CleanupResult{
		ProcessedCount: len(am.Manager.AuditLog),
		DeletedCount:   len(entriesToDelete),
		ErrorCount:     0,
	}

	if !dryRun {
		// In a real implementation, actually delete entries
		am.Logger.Infof("Would delete %d audit entries (dry_run: %t)", len(entriesToDelete), dryRun)
	}

	return result
}

func (am *AuditManagement) cleanupData(dataType string, dryRun bool) CleanupResult {
	// Mock cleanup for demo
	return CleanupResult{
		ProcessedCount: 1000,
		DeletedCount:   200,
		ErrorCount:     0,
	}
}