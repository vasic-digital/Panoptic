package executor

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTestResult_MarshalJSON_Correctness(t *testing.T) {
	original := TestResult{
		AppName:     "Test App",
		AppType:     "web",
		StartTime:   time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
		EndTime:     time.Date(2023, 1, 1, 12, 1, 0, 0, time.UTC),
		Duration:    time.Minute,
		Screenshots: []string{"screenshot1.png", "screenshot2.png"},
		Videos:      []string{"video1.mp4"},
		Metrics: map[string]interface{}{
			"requests":        100,
			"errors":          2,
			"duration":        1.5,
			"click_actions":   []string{"button1", "link2"},
			"fill_actions":    []map[string]string{{"selector": "input1", "value": "test"}},
			"start_time":      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			"end_time":        time.Date(2023, 1, 1, 12, 1, 0, 0, time.UTC),
			"total_duration":  time.Minute,
			"custom_metric":   "custom value",
			"boolean_metric":  true,
		},
		Success: true,
		Error:   "",
	}

	// Marshal with our optimized implementation
	optimizedData, err := original.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Unmarshal both versions to verify they're equivalent
	var optimizedResult TestResult
	if err := json.Unmarshal(optimizedData, &optimizedResult); err != nil {
		t.Fatalf("Failed to unmarshal optimized JSON: %v", err)
	}

	// Test equivalence
	if optimizedResult.AppName != original.AppName {
		t.Errorf("AppName: got %q, want %q", optimizedResult.AppName, original.AppName)
	}
	if optimizedResult.AppType != original.AppType {
		t.Errorf("AppType: got %q, want %q", optimizedResult.AppType, original.AppType)
	}
	if optimizedResult.Success != original.Success {
		t.Errorf("Success: got %v, want %v", optimizedResult.Success, original.Success)
	}
	if len(optimizedResult.Screenshots) != len(original.Screenshots) {
		t.Errorf("Screenshots length: got %d, want %d", len(optimizedResult.Screenshots), len(original.Screenshots))
	}
	if len(optimizedResult.Videos) != len(original.Videos) {
		t.Errorf("Videos length: got %d, want %d", len(optimizedResult.Videos), len(original.Videos))
	}

	// Test specific metrics (note: JSON unmarshaling converts numbers to float64)
	if optimizedResult.Metrics["requests"].(float64) != float64(original.Metrics["requests"].(int)) {
		t.Errorf("Metrics requests: got %v (type %T), want %v", 
			optimizedResult.Metrics["requests"], optimizedResult.Metrics["requests"],
			float64(original.Metrics["requests"].(int)))
	}
	if optimizedResult.Metrics["duration"].(float64) != original.Metrics["duration"].(float64) {
		t.Errorf("Metrics duration: got %v, want %v", optimizedResult.Metrics["duration"], original.Metrics["duration"])
	}
	if optimizedResult.Metrics["custom_metric"].(string) != original.Metrics["custom_metric"].(string) {
		t.Errorf("Metrics custom_metric: got %v, want %v", optimizedResult.Metrics["custom_metric"], original.Metrics["custom_metric"])
	}
	if optimizedResult.Metrics["boolean_metric"].(bool) != original.Metrics["boolean_metric"].(bool) {
		t.Errorf("Metrics boolean_metric: got %v, want %v", optimizedResult.Metrics["boolean_metric"], original.Metrics["boolean_metric"])
	}

	// Test JSON string is valid
	var temp interface{}
	if err := json.Unmarshal(optimizedData, &temp); err != nil {
		t.Errorf("Generated JSON is invalid: %v\nJSON: %s", err, string(optimizedData))
	}

	t.Logf("Generated JSON length: %d bytes", len(optimizedData))
	t.Logf("JSON preview: %s", string(optimizedData[:min(200, len(optimizedData))]))
}

func TestTestResult_MarshalJSON_WithError(t *testing.T) {
	result := TestResult{
		AppName:    "Test App",
		AppType:    "web",
		StartTime:  time.Now(),
		EndTime:    time.Now().Add(time.Minute),
		Duration:   time.Minute,
		Metrics:    map[string]interface{}{},
		Success:    false,
		Error:      "Something went wrong",
	}

	data, err := result.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Verify error field is included
	if !contains(data, `"error":`) {
		t.Error("Error field should be included in JSON")
	}
	if !contains(data, "Something went wrong") {
		t.Error("Error message should be included in JSON")
	}

	// Verify it's valid JSON
	var temp interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		t.Errorf("Generated JSON is invalid: %v", err)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func contains(data []byte, substr string) bool {
	for i := 0; i <= len(data)-len(substr); i++ {
		if string(data[i:i+len(substr)]) == substr {
			return true
		}
	}
	return false
}