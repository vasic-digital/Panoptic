package executor

import (
	"context"
	"testing"
	"time"

	"panoptic/internal/config"
	"panoptic/internal/logger"
)

// TestAnalyzeActions tests action grouping for parallelization
func TestAnalyzeActions(t *testing.T) {
	tests := []struct {
		name     string
		actions  []config.Action
		expected int // number of groups expected
	}{
		{
			name:     "empty actions",
			actions:  []config.Action{},
			expected: 0,
		},
		{
			name: "all parallelizable",
			actions: []config.Action{
				{Type: "click", Name: "click1"},
				{Type: "click", Name: "click2"},
				{Type: "fill", Name: "fill1"},
			},
			expected: 1,
		},
		{
			name: "mixed sequential and parallel",
			actions: []config.Action{
				{Type: "navigate", Name: "nav1"},
				{Type: "click", Name: "click1"},
				{Type: "fill", Name: "fill1"},
				{Type: "submit", Name: "submit1"},
				{Type: "click", Name: "click2"},
				{Type: "fill", Name: "fill2"},
			},
			expected: 4, // navigate || [click,fill] || submit || [click,fill]
		},
		{
			name: "all sequential",
			actions: []config.Action{
				{Type: "navigate", Name: "nav1"},
				{Type: "submit", Name: "submit1"},
				{Type: "navigate", Name: "nav2"},
			},
			expected: 3,
		},
	}
	
	log := logger.NewLogger(false)
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &Executor{
				config: &config.Config{
					Actions: tt.actions,
				},
				logger: log,
			}
			
			groups := executor.analyzeActions()
			if len(groups) != tt.expected {
				t.Errorf("Expected %d groups, got %d", tt.expected, len(groups))
			}
			
			// Verify group properties
			for i, group := range groups {
				if group.Parallelizable && len(group.Actions) <= 1 {
					t.Errorf("Group %d marked as parallelizable but has %d actions", i, len(group.Actions))
				}
			}
		})
	}
}

// TestExecuteActionGroup tests parallel and sequential action execution
func TestExecuteActionGroup(t *testing.T) {
	log := logger.NewLogger(false)
	tempDir := t.TempDir()
	
	cfg := &config.Config{
		Name:     "Test",
		Apps:     []config.AppConfig{},
		Actions:  []config.Action{},
		Settings: config.Settings{},
	}
	
	executor := NewExecutor(cfg, tempDir, log)
	
	// Mock platform for testing
	mockPlatform := &MockPlatform{
		metrics: make(map[string]interface{}),
	}
	
	app := config.AppConfig{
		Name: "Test App",
		Type: "web",
	}
	
	result := &TestResult{
		AppName:  app.Name,
		AppType:  app.Type,
		StartTime: time.Now(),
		Metrics:  make(map[string]interface{}),
	}
	
	recordingFile := ""
	ctx := context.Background()
	
	// Test sequential group
	t.Run("Sequential Group", func(t *testing.T) {
		sequentialGroup := ActionGroup{
			Actions: []config.Action{
				{Type: "navigate", Name: "nav1", Value: "https://example.com"},
				{Type: "submit", Name: "submit1"},
			},
			Parallelizable: false,
		}
		
		err := executor.executeActionGroup(ctx, mockPlatform, app, sequentialGroup, result, &recordingFile)
		if err != nil {
			t.Errorf("Sequential group execution failed: %v", err)
		}
		
		// Verify actions were executed
		if len(mockPlatform.executedActions) != 2 {
			t.Errorf("Expected 2 actions executed, got %d", len(mockPlatform.executedActions))
		}
	})
	
	// Test parallel group
	t.Run("Parallel Group", func(t *testing.T) {
		mockPlatform.executedActions = []config.Action{} // Reset
		
		parallelGroup := ActionGroup{
			Actions: []config.Action{
				{Type: "click", Name: "click1", Selector: "#btn1"},
				{Type: "fill", Name: "fill1", Selector: "#input1", Value: "test"},
				{Type: "click", Name: "click2", Selector: "#btn2"},
			},
			Parallelizable: true,
		}
		
		err := executor.executeActionGroup(ctx, mockPlatform, app, parallelGroup, result, &recordingFile)
		if err != nil {
			t.Errorf("Parallel group execution failed: %v", err)
		}
		
		// Verify actions were executed
		if len(mockPlatform.executedActions) != 3 {
			t.Errorf("Expected 3 actions executed, got %d", len(mockPlatform.executedActions))
		}
	})
}

// MockPlatform is a simple mock for testing
type MockPlatform struct {
	executedActions []config.Action
	metrics        map[string]interface{}
}

func (m *MockPlatform) Initialize(app config.AppConfig) error {
	m.metrics["initialized"] = true
	return nil
}

func (m *MockPlatform) Navigate(url string) error {
	m.executedActions = append(m.executedActions, config.Action{Type: "navigate", Value: url})
	return nil
}

func (m *MockPlatform) Click(selector string) error {
	m.executedActions = append(m.executedActions, config.Action{Type: "click", Selector: selector})
	return nil
}

func (m *MockPlatform) Fill(selector, value string) error {
	m.executedActions = append(m.executedActions, config.Action{Type: "fill", Selector: selector, Value: value})
	return nil
}

func (m *MockPlatform) Submit(selector string) error {
	m.executedActions = append(m.executedActions, config.Action{Type: "submit", Selector: selector})
	return nil
}

func (m *MockPlatform) Wait(duration int) error {
	m.executedActions = append(m.executedActions, config.Action{Type: "wait", WaitTime: duration})
	return nil
}

func (m *MockPlatform) Screenshot(filename string) error {
	m.executedActions = append(m.executedActions, config.Action{Type: "screenshot", Name: filename})
	return nil
}

func (m *MockPlatform) StartRecording(filename string) error {
	m.executedActions = append(m.executedActions, config.Action{Type: "record", Name: filename})
	return nil
}

func (m *MockPlatform) StopRecording() error {
	return nil
}

func (m *MockPlatform) GetMetrics() map[string]interface{} {
	return m.metrics
}

func (m *MockPlatform) Close() error {
	m.metrics["closed"] = true
	return nil
}

// BenchmarkSequentialVsParallel compares sequential vs parallel execution
func BenchmarkSequentialVsParallel(b *testing.B) {
	log := logger.NewLogger(false)
	tempDir := b.TempDir()
	
	// Create many parallelizable actions
	actions := make([]config.Action, 20)
	for i := 0; i < 20; i++ {
		actions[i] = config.Action{
			Type:     "click",
			Name:     "click",
			Selector: "#btn",
		}
	}
	
	cfg := &config.Config{
		Name:     "Benchmark",
		Actions:  actions,
		Settings: config.Settings{},
	}
	
	app := config.AppConfig{
		Name: "Benchmark App",
		Type: "web",
	}
	
	ctx := context.Background()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		executor := NewExecutor(cfg, tempDir, log)
		mockPlatform := &MockPlatform{
			metrics: make(map[string]interface{}),
		}
		
		groups := executor.analyzeActions()
		for _, group := range groups {
			result := &TestResult{}
			recordingFile := ""
			executor.executeActionGroup(ctx, mockPlatform, app, group, result, &recordingFile)
		}
	}
}