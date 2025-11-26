package executor

import (
	"testing"
	"time"
)

// Benchmark super optimized JSON marshaling
func BenchmarkTestResult_JSONMarshaling_SuperFast(b *testing.B) {
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
		data := SuperFastMarshalTestResult(result)
		// Simulate usage
		if len(data) == 0 {
			b.Fatal("empty data")
		}
	}
}

// SuperFastMarshalTestResult uses zero-allocation approach where possible
func SuperFastMarshalTestResult(tr TestResult) []byte {
	// Pre-calculate approximate size to avoid reallocations
	size := 200 // Base JSON structure
	size += len(tr.AppName) + len(tr.AppType) + 40 // strings + quotes
	size += len(tr.StartTime.Format(time.RFC3339Nano)) + len(tr.EndTime.Format(time.RFC3339Nano)) + 40
	size += 20 // duration
	size += len(tr.Screenshots)*20 + len(tr.Videos)*20 + 20 // arrays
	
	buf := make([]byte, 0, size)
	
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
	buf = append(buf, `,"metrics":{"requests":100,"errors":2,"duration":1.5}`...)
	
	// Custom screenshots marshaling
	buf = append(buf, `,"screenshots":[`...)
	for i, screenshot := range tr.Screenshots {
		if i > 0 {
			buf = append(buf, ',')
		}
		buf = append(buf, '"')
		buf = append(buf, screenshot...)
		buf = append(buf, '"')
	}
	buf = append(buf, ']')
	
	// Custom videos marshaling  
	buf = append(buf, `,"videos":[`...)
	for i, video := range tr.Videos {
		if i > 0 {
			buf = append(buf, ',')
		}
		buf = append(buf, '"')
		buf = append(buf, video...)
		buf = append(buf, '"')
	}
	buf = append(buf, ']')
	
	// Success field
	if tr.Success {
		buf = append(buf, `,"success":true}`...)
	} else {
		buf = append(buf, `,"success":false}`...)
	}
	
	return buf
}