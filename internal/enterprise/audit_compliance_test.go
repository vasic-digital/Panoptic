package enterprise

import (
	"context"
	"fmt"
	"testing"
	"time"

	"panoptic/internal/logger"

	"github.com/stretchr/testify/assert"
)

// TestNewAuditManagement tests audit management creation
func TestNewAuditManagement(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		AuditLog: []AuditEntry{},
	}

	am := NewAuditManagement(manager)

	assert.NotNil(t, am, "AuditManagement should not be nil")
	assert.Equal(t, manager, am.Manager, "Manager should be set")
}

// TestGetAuditLog tests audit log retrieval
func TestGetAuditLog(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		AuditLog: []AuditEntry{},
	}

	// Add multiple audit entries
	for i := 1; i <= 10; i++ {
		entry := AuditEntry{
			ID:         generateTestID(i),
			Timestamp:  time.Now().Add(time.Duration(-i) * time.Hour),
			UserID:     "user-123",
			Username:   "testuser",
			Action:     "test.action",
			Resource:   "test_resource",
			ResourceID: generateTestID(i),
			Category:   "test",
			Severity:   "medium",
			Success:    true,
			Details:    map[string]string{"test": "data"},
		}
		manager.AuditLog = append(manager.AuditLog, entry)
	}

	am := NewAuditManagement(manager)
	ctx := context.Background()

	req := GetAuditLogRequest{
		Page:     1,
		PageSize: 5,
	}

	response, err := am.GetAuditLog(ctx, req)

	assert.NoError(t, err, "Should get audit log without error")
	assert.NotNil(t, response, "Response should not be nil")
	assert.Equal(t, 10, response.Total, "Should have 10 total entries")
	assert.Equal(t, 5, len(response.Entries), "Should return 5 entries")
}

// TestGetAuditLog_WithFilters tests audit log with filters
func TestGetAuditLog_WithFilters(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		AuditLog: []AuditEntry{},
	}

	// Add entries with different users
	manager.AuditLog = append(manager.AuditLog, AuditEntry{
		ID:        "entry-1",
		Timestamp: time.Now(),
		UserID:    "user-1",
		Username:  "user1",
		Action:    "login",
		Category:  "auth",
		Severity:  "low",
		Success:   true,
	})
	manager.AuditLog = append(manager.AuditLog, AuditEntry{
		ID:        "entry-2",
		Timestamp: time.Now(),
		UserID:    "user-2",
		Username:  "user2",
		Action:    "delete",
		Category:  "data",
		Severity:  "high",
		Success:   false,
	})

	am := NewAuditManagement(manager)
	ctx := context.Background()

	// Filter by user
	req := GetAuditLogRequest{
		UserID:   "user-1",
		Page:     1,
		PageSize: 10,
	}

	response, err := am.GetAuditLog(ctx, req)

	assert.NoError(t, err, "Should get audit log without error")
	assert.NotNil(t, response, "Response should not be nil")
	assert.Equal(t, 1, response.Total, "Should have 1 entry for user-1")
}

// TestGetAuditLog_DateFilter tests date filtering
func TestGetAuditLog_DateFilter(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		AuditLog: []AuditEntry{},
	}

	now := time.Now()
	manager.AuditLog = append(manager.AuditLog, AuditEntry{
		ID:        "entry-1",
		Timestamp: now.Add(-2 * time.Hour),
		UserID:    "user-1",
		Action:    "test1",
		Success:   true,
	})
	manager.AuditLog = append(manager.AuditLog, AuditEntry{
		ID:        "entry-2",
		Timestamp: now,
		UserID:    "user-1",
		Action:    "test2",
		Success:   true,
	})

	am := NewAuditManagement(manager)
	ctx := context.Background()

	req := GetAuditLogRequest{
		StartTime: now.Add(-1 * time.Hour),
		Page:      1,
		PageSize:  10,
	}

	response, err := am.GetAuditLog(ctx, req)

	assert.NoError(t, err, "Should get audit log without error")
	assert.NotNil(t, response, "Response should not be nil")
	assert.Equal(t, 1, response.Total, "Should have 1 entry after start time")
}

// TestGetAuditSummary tests audit summary generation
func TestGetAuditSummary(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		AuditLog: []AuditEntry{},
	}

	now := time.Now()
	// Add successful entry
	manager.AuditLog = append(manager.AuditLog, AuditEntry{
		ID:        "entry-1",
		Timestamp: now,
		UserID:    "user-1",
		Username:  "user1",
		Action:    "login",
		Category:  "auth",
		Severity:  "low",
		Success:   true,
	})
	// Add failed entry
	manager.AuditLog = append(manager.AuditLog, AuditEntry{
		ID:        "entry-2",
		Timestamp: now,
		UserID:    "user-1",
		Username:  "user1",
		Action:    "delete",
		Category:  "data",
		Severity:  "high",
		Success:   false,
	})

	am := NewAuditManagement(manager)
	ctx := context.Background()

	req := GetAuditSummaryRequest{}

	summary, err := am.GetAuditSummary(ctx, req)

	assert.NoError(t, err, "Should get audit summary without error")
	assert.NotNil(t, summary, "Summary should not be nil")
	assert.Equal(t, 2, summary.TotalEntries, "Should have 2 total entries")
	assert.Equal(t, 1, summary.SuccessCount, "Should have 1 success")
	assert.Equal(t, 1, summary.FailureCount, "Should have 1 failure")
	assert.Equal(t, float64(50), summary.SuccessRate, "Success rate should be 50%")
	assert.NotNil(t, summary.CategoryStats, "Category stats should be populated")
	assert.NotNil(t, summary.ActionStats, "Action stats should be populated")
}

// TestGetComplianceStatus tests compliance status retrieval
func TestGetComplianceStatus(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		AuditLog: []AuditEntry{},
		Config: EnterpriseConfig{
			Compliance: ComplianceConfig{
				Enabled:        true,
				Standards:      []string{"SOC2", "GDPR"},
				DataRetention:  365,
				AuditRetention: 730,
				DataEncryption: true,
				AuditEncryption: true,
			},
		},
	}

	am := NewAuditManagement(manager)
	ctx := context.Background()

	req := GetComplianceStatusRequest{}

	response, err := am.GetComplianceStatus(ctx, req)

	assert.NoError(t, err, "Should get compliance status without error")
	assert.NotNil(t, response, "Response should not be nil")
	assert.Equal(t, []string{"SOC2", "GDPR"}, response.Standards)
	assert.Equal(t, 365, response.DataRetention)
	assert.Equal(t, 730, response.AuditRetention)
	assert.True(t, response.DataEncryption)
	assert.NotNil(t, response.Reports, "Reports should be populated")
}

// TestGetComplianceStatus_Disabled tests when compliance is disabled
func TestGetComplianceStatus_Disabled(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		AuditLog: []AuditEntry{},
		Config: EnterpriseConfig{
			Compliance: ComplianceConfig{
				Enabled: false,
			},
		},
	}

	am := NewAuditManagement(manager)
	ctx := context.Background()

	req := GetComplianceStatusRequest{}

	response, err := am.GetComplianceStatus(ctx, req)

	assert.Error(t, err, "Should error when compliance disabled")
	assert.Nil(t, response, "Response should be nil")
	assert.Contains(t, err.Error(), "not enabled")
}

// TestExportAuditLog tests audit log export
func TestExportAuditLog(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		AuditLog: []AuditEntry{},
		Config: EnterpriseConfig{
			StoragePath: "/tmp",
		},
	}

	// Add test entries
	manager.AuditLog = append(manager.AuditLog, AuditEntry{
		ID:        "entry-1",
		Timestamp: time.Now(),
		UserID:    "user-1",
		Username:  "testuser",
		Action:    "test.action",
		Success:   true,
	})

	am := NewAuditManagement(manager)
	ctx := context.Background()

	testCases := []struct {
		name   string
		format string
	}{
		{"JSON format", "json"},
		{"CSV format", "csv"},
		{"XML format", "xml"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := ExportAuditLogRequest{
				Format: tc.format,
			}

			response, err := am.ExportAuditLog(ctx, req)

			assert.NoError(t, err, "Should export audit log without error")
			assert.NotNil(t, response, "Response should not be nil")
			assert.NotEmpty(t, response.Filename, "Filename should be set")
			assert.Greater(t, response.Size, 0, "Size should be greater than 0")
			assert.Equal(t, 1, response.EntriesCount, "Should export 1 entry")
		})
	}
}

// TestExportAuditLog_InvalidFormat tests export with invalid format
func TestExportAuditLog_InvalidFormat(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		AuditLog: []AuditEntry{},
		Config: EnterpriseConfig{
			StoragePath: "/tmp",
		},
	}

	am := NewAuditManagement(manager)
	ctx := context.Background()

	req := ExportAuditLogRequest{
		Format: "invalid",
	}

	response, err := am.ExportAuditLog(ctx, req)

	assert.Error(t, err, "Should error with invalid format")
	assert.Nil(t, response, "Response should be nil")
	assert.Contains(t, err.Error(), "unsupported")
}

// TestCreateComplianceReport tests compliance report generation
func TestCreateComplianceReport(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		AuditLog: []AuditEntry{},
		Config: EnterpriseConfig{
			Compliance: ComplianceConfig{
				Enabled:   true,
				Standards: []string{"SOC2"},
			},
		},
	}

	am := NewAuditManagement(manager)
	ctx := context.Background()

	req := CreateComplianceReportRequest{
		Standard: "SOC2",
	}

	report, err := am.CreateComplianceReport(ctx, req)

	assert.NoError(t, err, "Should create compliance report without error")
	assert.NotNil(t, report, "Report should not be nil")
	assert.Equal(t, "SOC2", report.Standard)
	assert.NotEmpty(t, report.Status, "Status should be set")
}

// TestGetRetentionStatus tests retention status retrieval
func TestGetRetentionStatus(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		AuditLog: []AuditEntry{},
		Config: EnterpriseConfig{
			Compliance: ComplianceConfig{
				DataRetention:  365,
				AuditRetention: 730,
			},
		},
	}

	// Add some audit entries
	for i := 0; i < 5; i++ {
		manager.AuditLog = append(manager.AuditLog, AuditEntry{
			ID:        generateTestID(i),
			Timestamp: time.Now().Add(time.Duration(-i*24) * time.Hour),
			Success:   true,
		})
	}

	am := NewAuditManagement(manager)
	ctx := context.Background()

	req := GetRetentionStatusRequest{}

	response, err := am.GetRetentionStatus(ctx, req)

	assert.NoError(t, err, "Should get retention status without error")
	assert.NotNil(t, response, "Response should not be nil")
	assert.Equal(t, 365, response.DataRetention)
	assert.Equal(t, 730, response.AuditRetention)
	assert.NotNil(t, response.AuditStats, "Audit stats should be populated")
	assert.Equal(t, 5, response.AuditStats.TotalEntries)
	assert.NotNil(t, response.DataStats, "Data stats should be populated")
}

// TestExecuteCleanup tests data cleanup execution
func TestExecuteCleanup(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		AuditLog: []AuditEntry{},
		Config: EnterpriseConfig{
			Compliance: ComplianceConfig{
				AuditRetention: 730,
			},
		},
	}

	// Add some old audit entries
	for i := 0; i < 10; i++ {
		manager.AuditLog = append(manager.AuditLog, AuditEntry{
			ID:        generateTestID(i),
			Timestamp: time.Now().Add(time.Duration(-i*365) * 24 * time.Hour),
			Success:   true,
		})
	}

	am := NewAuditManagement(manager)
	ctx := context.Background()

	req := ExecuteCleanupRequest{
		DryRun:       true,
		IncludeAudit: true,
		IncludeData:  true,
	}

	response, err := am.ExecuteCleanup(ctx, req)

	assert.NoError(t, err, "Should execute cleanup without error")
	assert.NotNil(t, response, "Response should not be nil")
	assert.NotNil(t, response.Results, "Results should be populated")
	assert.Greater(t, response.Duration.Nanoseconds(), int64(0), "Duration should be positive")
}

// TestExecuteCleanup_DryRun tests cleanup in dry run mode
func TestExecuteCleanup_DryRun(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		AuditLog: []AuditEntry{},
		Config: EnterpriseConfig{
			Compliance: ComplianceConfig{
				AuditRetention: 730,
			},
		},
	}

	originalCount := 5
	for i := 0; i < originalCount; i++ {
		manager.AuditLog = append(manager.AuditLog, AuditEntry{
			ID:        generateTestID(i),
			Timestamp: time.Now().Add(time.Duration(-i*365) * 24 * time.Hour),
			Success:   true,
		})
	}

	am := NewAuditManagement(manager)
	ctx := context.Background()

	req := ExecuteCleanupRequest{
		DryRun:       true,
		IncludeAudit: true,
	}

	response, err := am.ExecuteCleanup(ctx, req)

	assert.NoError(t, err, "Should execute dry run without error")
	assert.NotNil(t, response, "Response should not be nil")
	// In dry run mode, old entries should not be deleted (but cleanup operation logs an audit entry)
	assert.GreaterOrEqual(t, len(manager.AuditLog), originalCount, "Original entries should still exist")
}

// TestSortAuditEntries tests audit entry sorting
func TestSortAuditEntries(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		AuditLog: []AuditEntry{},
	}

	am := NewAuditManagement(manager)

	now := time.Now()
	entries := []AuditEntry{
		{ID: "3", Timestamp: now.Add(-3 * time.Hour)},
		{ID: "1", Timestamp: now.Add(-1 * time.Hour)},
		{ID: "2", Timestamp: now.Add(-2 * time.Hour)},
	}

	// Sort ascending
	am.sortAuditEntries(entries, true)
	assert.Equal(t, "3", entries[0].ID, "First entry should be oldest")
	assert.Equal(t, "1", entries[2].ID, "Last entry should be newest")

	// Sort descending
	am.sortAuditEntries(entries, false)
	assert.Equal(t, "1", entries[0].ID, "First entry should be newest")
	assert.Equal(t, "3", entries[2].ID, "Last entry should be oldest")
}

// TestGenerateComplianceReport tests compliance report generation
func TestGenerateComplianceReport(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		AuditLog: []AuditEntry{},
		Config: EnterpriseConfig{
			Compliance: ComplianceConfig{
				Enabled: true,
			},
		},
	}

	am := NewAuditManagement(manager)

	standards := []string{"SOC2", "GDPR", "HIPAA", "PCI-DSS"}
	for _, standard := range standards {
		report := am.generateComplianceReport(standard)

		assert.Equal(t, standard, report.Standard)
		assert.NotEmpty(t, report.Status, "Status should be set")
		assert.Greater(t, report.MaxScore, 0, "Max score should be positive")
		assert.NotNil(t, report.Requirements, "Requirements should be populated")
	}
}

// TestExportFormats tests different export formats
func TestExportFormats(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		AuditLog: []AuditEntry{},
	}

	am := NewAuditManagement(manager)

	entries := []AuditEntry{
		{
			ID:        "entry-1",
			Timestamp: time.Now(),
			UserID:    "user-1",
			Username:  "testuser",
			Action:    "test.action",
			Resource:  "test",
			Success:   true,
			Severity:  "low",
			Category:  "test",
		},
	}

	// Test JSON export
	jsonData := am.exportToJSON(entries)
	assert.NotEmpty(t, jsonData, "JSON export should not be empty")
	assert.Contains(t, jsonData, "entry-1", "JSON should contain entry ID")

	// Test CSV export
	csvData := am.exportToCSV(entries)
	assert.NotEmpty(t, csvData, "CSV export should not be empty")
	assert.Contains(t, csvData, "timestamp", "CSV should contain timestamp header")

	// Test XML export
	xmlData := am.exportToXML(entries)
	assert.NotEmpty(t, xmlData, "XML export should not be empty")
	assert.Contains(t, xmlData, "<audit_entries>", "XML should contain root element")
}

// TestCalculateRetentionStats tests retention calculations
func TestCalculateRetentionStats(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		AuditLog: []AuditEntry{},
		Config: EnterpriseConfig{
			Compliance: ComplianceConfig{
				AuditRetention: 365,
			},
		},
	}

	// Add entries of different ages
	for i := 0; i < 10; i++ {
		manager.AuditLog = append(manager.AuditLog, AuditEntry{
			ID:        generateTestID(i),
			Timestamp: time.Now().Add(time.Duration(-i*100) * 24 * time.Hour),
			Success:   true,
		})
	}

	am := NewAuditManagement(manager)

	toRetain := am.calculateEntriesToRetain()
	toDelete := am.calculateEntriesToDelete()

	assert.GreaterOrEqual(t, toRetain, 0, "Entries to retain should be non-negative")
	assert.GreaterOrEqual(t, toDelete, 0, "Entries to delete should be non-negative")
	assert.Equal(t, len(manager.AuditLog), toRetain+toDelete, "Sum should equal total entries")
}

// TestGetOldestNewestAuditEntry tests timestamp retrieval
func TestGetOldestNewestAuditEntry(t *testing.T) {
	log := logger.NewLogger(false)
	manager := &EnterpriseManager{
		Logger:   *log,
		AuditLog: []AuditEntry{},
	}

	now := time.Now()
	oldest := now.Add(-10 * 24 * time.Hour)
	newest := now

	manager.AuditLog = append(manager.AuditLog, AuditEntry{
		ID:        "old",
		Timestamp: oldest,
	})
	manager.AuditLog = append(manager.AuditLog, AuditEntry{
		ID:        "new",
		Timestamp: newest,
	})
	manager.AuditLog = append(manager.AuditLog, AuditEntry{
		ID:        "middle",
		Timestamp: now.Add(-5 * 24 * time.Hour),
	})

	am := NewAuditManagement(manager)

	oldestTime := am.getOldestAuditEntry()
	newestTime := am.getNewestAuditEntry()

	assert.True(t, oldestTime.Equal(oldest), "Should return oldest timestamp")
	assert.True(t, newestTime.Equal(newest), "Should return newest timestamp")
}

// TestAuditEntry_Structure tests AuditEntry struct
func TestAuditEntry_Structure(t *testing.T) {
	entry := AuditEntry{
		ID:         "entry-123",
		Timestamp:  time.Now(),
		UserID:     "user-123",
		Username:   "testuser",
		Action:     "test.action",
		Resource:   "test_resource",
		ResourceID: "resource-123",
		Details:    map[string]string{"key": "value"},
		Success:    true,
		Severity:   "medium",
		Category:   "test",
		IPAddress:  "127.0.0.1",
		UserAgent:  "test-agent",
	}

	assert.NotNil(t, entry, "Entry should not be nil")
	assert.Equal(t, "entry-123", entry.ID)
	assert.Equal(t, "testuser", entry.Username)
	assert.True(t, entry.Success)
	assert.Equal(t, 1, len(entry.Details))
}

// TestComplianceReport_Structure tests ComplianceReport struct
func TestComplianceReport_Structure(t *testing.T) {
	report := ComplianceReport{
		Standard:     "SOC2",
		Status:       "compliant",
		LastAssessed: time.Now(),
		Score:        95,
		MaxScore:     100,
		Requirements: []ComplianceRequirement{
			{
				ID:          "req-1",
				Name:        "Test Requirement",
				Description: "Test description",
				Mandatory:   true,
				Satisfied:   true,
				Score:       10,
				MaxScore:    10,
			},
		},
		Issues:          []ComplianceIssue{},
		Recommendations: []string{"Keep up the good work"},
	}

	assert.NotNil(t, report, "Report should not be nil")
	assert.Equal(t, "SOC2", report.Standard)
	assert.Equal(t, 95, report.Score)
	assert.Equal(t, 1, len(report.Requirements))
}

// Helper function to generate test IDs
func generateTestID(i int) string {
	return fmt.Sprintf("test-id-%d", i)
}
