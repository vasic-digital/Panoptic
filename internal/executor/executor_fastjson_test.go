package executor

import (
	"encoding/json"
	"testing"
	"time"
)

// Test different JSON marshaling strategies
func BenchmarkTestResult_JSONMarshaling_FastMarshal(b *testing.B) {
	result := TestResult{
		AppName:     "Test App",
		AppType:     "web",
		StartTime:   time.Now(),
		EndTime:     time.Now(),
		Duration:    time.Second,
		Screenshots: []string{"screenshot1.png", "screenshot2.png"},
		Videos:      []string{"video1.mp4"},
		Metrics: map[string]interface{}{
			"requests": 100,
			"errors":   2,
			"duration": 1.5,
		},
		Success: true,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Use custom fast marshal
		data, err := FastMarshalTestResult(result)
		if err != nil {
			b.Fatal(err)
		}
		// Simulate usage
		if len(data) == 0 {
			b.Fatal("empty data")
		}
	}
}

// FastMarshalTestResult uses a optimized JSON builder approach
func FastMarshalTestResult(tr TestResult) ([]byte, error) {
	// Use a builder with pre-allocated capacity
	buf := make([]byte, 0, 1024)
	
	buf = append(buf, `{"app_name":"`...)
	buf = append(buf, tr.AppName...)
	buf = append(buf, `","app_type":"`...)
	buf = append(buf, tr.AppType...)
	buf = append(buf, `","start_time":"`...)
	buf = append(buf, tr.StartTime.Format(time.RFC3339Nano)...)
	buf = append(buf, `","end_time":"`...)
	buf = append(buf, tr.EndTime.Format(time.RFC3339Nano)...)
	buf = append(buf, `","duration":`...)
	buf = append(buf, formatInt64(tr.Duration.Nanoseconds())...)
	buf = append(buf, `,"metrics":`...)
	
	// Marshal metrics the only complex part
	metricsBytes, err := json.Marshal(tr.Metrics)
	if err != nil {
		return nil, err
	}
	buf = append(buf, metricsBytes...)
	
	buf = append(buf, `,"screenshots":`...)
	screenshotsBytes, err := json.Marshal(tr.Screenshots)
	if err != nil {
		return nil, err
	}
	buf = append(buf, screenshotsBytes...)
	
	buf = append(buf, `,"videos":`...)
	videosBytes, err := json.Marshal(tr.Videos)
	if err != nil {
		return nil, err
	}
	buf = append(buf, videosBytes...)
	
	if tr.Success {
		buf = append(buf, `,"success":true}`...)
	} else {
		buf = append(buf, `,"success":false}`...)
	}
	
	return buf, nil
}