package executor

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GenerateComprehensiveReport creates a full HTML test report with results, screenshots, and video embeds.
func GenerateComprehensiveReport(outputPath string, results []TestResult) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create report directory: %w", err)
	}

	var b strings.Builder
	b.Grow(8192)

	// Count pass/fail
	passed, failed := 0, 0
	var totalDuration time.Duration
	for _, r := range results {
		if r.Success {
			passed++
		} else {
			failed++
		}
		totalDuration += r.Duration
	}

	// HTML header
	b.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Panoptic Test Report</title>
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:#1a1a2e;color:#e0e0e0;padding:20px}
.header{text-align:center;padding:30px 0;border-bottom:2px solid #16213e}
.header h1{font-size:2em;color:#e94560}
.header .subtitle{color:#888;margin-top:8px}
.summary{display:flex;justify-content:center;gap:30px;padding:20px 0;flex-wrap:wrap}
.stat{text-align:center;padding:15px 25px;background:#16213e;border-radius:8px;min-width:120px}
.stat .value{font-size:2em;font-weight:bold}
.stat .label{font-size:0.85em;color:#888;margin-top:4px}
.stat.pass .value{color:#4caf50}
.stat.fail .value{color:#f44336}
.stat.total .value{color:#2196f3}
.stat.time .value{color:#ff9800;font-size:1.4em}
.apps{padding:20px 0}
.app-card{background:#16213e;border-radius:8px;margin:15px 0;padding:20px;border-left:4px solid #4caf50}
.app-card.failed{border-left-color:#f44336}
.app-card .app-header{display:flex;justify-content:space-between;align-items:center;flex-wrap:wrap}
.app-card .app-name{font-size:1.3em;font-weight:bold}
.app-card .app-type{background:#0f3460;padding:3px 10px;border-radius:4px;font-size:0.8em}
.app-card .app-status{padding:4px 12px;border-radius:4px;font-weight:bold;font-size:0.9em}
.app-card .app-status.pass{background:#1b5e20;color:#a5d6a7}
.app-card .app-status.fail{background:#b71c1c;color:#ef9a9a}
.app-card .app-meta{margin-top:10px;font-size:0.9em;color:#888}
.app-card .app-error{margin-top:10px;padding:10px;background:#2a0a0a;border-radius:4px;color:#ef9a9a;font-family:monospace;font-size:0.85em;white-space:pre-wrap;word-break:break-all}
.screenshots{margin-top:15px}
.screenshots h3{font-size:1em;margin-bottom:8px;color:#aaa}
.screenshot-grid{display:flex;gap:10px;flex-wrap:wrap}
.screenshot-grid img{max-width:200px;max-height:150px;border-radius:4px;border:1px solid #333;cursor:pointer;transition:transform 0.2s}
.screenshot-grid img:hover{transform:scale(1.05)}
.screenshot-grid a{color:#64b5f6;font-size:0.85em;text-decoration:none}
.videos{margin-top:15px}
.videos h3{font-size:1em;margin-bottom:8px;color:#aaa}
.videos video{max-width:480px;border-radius:4px;border:1px solid #333}
.videos .video-link{color:#64b5f6;font-size:0.85em;text-decoration:none;display:block;margin-top:4px}
.footer{text-align:center;padding:30px 0;color:#555;font-size:0.85em;border-top:1px solid #16213e;margin-top:30px}
</style>
</head>
<body>
<div class="header">
<h1>Panoptic Test Report</h1>
<div class="subtitle">Generated: `)
	b.WriteString(html.EscapeString(time.Now().Format("2006-01-02 15:04:05 MST")))
	b.WriteString(`</div>
</div>

<div class="summary">
<div class="stat total"><div class="value">`)
	b.WriteString(fmt.Sprintf("%d", len(results)))
	b.WriteString(`</div><div class="label">Total Apps</div></div>
<div class="stat pass"><div class="value">`)
	b.WriteString(fmt.Sprintf("%d", passed))
	b.WriteString(`</div><div class="label">Passed</div></div>
<div class="stat fail"><div class="value">`)
	b.WriteString(fmt.Sprintf("%d", failed))
	b.WriteString(`</div><div class="label">Failed</div></div>
<div class="stat time"><div class="value">`)
	b.WriteString(formatDuration(totalDuration))
	b.WriteString(`</div><div class="label">Total Duration</div></div>
</div>

<div class="apps">
`)

	// Per-app cards
	for _, r := range results {
		statusClass := "pass"
		statusText := "PASSED"
		cardClass := ""
		if !r.Success {
			statusClass = "fail"
			statusText = "FAILED"
			cardClass = " failed"
		}

		b.WriteString(fmt.Sprintf(`<div class="app-card%s">
<div class="app-header">
<span class="app-name">%s</span>
<span class="app-type">%s</span>
<span class="app-status %s">%s</span>
</div>
<div class="app-meta">Duration: %s | Start: %s</div>
`,
			cardClass,
			html.EscapeString(r.AppName),
			html.EscapeString(r.AppType),
			statusClass, statusText,
			formatDuration(r.Duration),
			r.StartTime.Format("15:04:05"),
		))

		if r.Error != "" {
			b.WriteString(fmt.Sprintf(`<div class="app-error">%s</div>
`, html.EscapeString(r.Error)))
		}

		// Screenshots
		if len(r.Screenshots) > 0 {
			b.WriteString(`<div class="screenshots"><h3>Screenshots</h3><div class="screenshot-grid">
`)
			for _, s := range r.Screenshots {
				relPath := filepath.Base(s)
				// Check if file exists
				if _, err := os.Stat(s); err == nil {
					b.WriteString(fmt.Sprintf(`<a href="screenshots/%s" target="_blank"><img src="screenshots/%s" alt="%s" loading="lazy"></a>
`, html.EscapeString(relPath), html.EscapeString(relPath), html.EscapeString(relPath)))
				} else {
					b.WriteString(fmt.Sprintf(`<a href="#">%s (not found)</a>
`, html.EscapeString(relPath)))
				}
			}
			b.WriteString(`</div></div>
`)
		}

		// Videos
		if len(r.Videos) > 0 {
			b.WriteString(`<div class="videos"><h3>Videos</h3>
`)
			for _, v := range r.Videos {
				relPath := filepath.Base(v)
				if _, err := os.Stat(v); err == nil {
					b.WriteString(fmt.Sprintf(`<video controls preload="metadata"><source src="videos/%s" type="video/mp4">Your browser does not support video.</video>
<a class="video-link" href="videos/%s" download>Download: %s</a>
`, html.EscapeString(relPath), html.EscapeString(relPath), html.EscapeString(relPath)))
				} else {
					b.WriteString(fmt.Sprintf(`<a class="video-link" href="#">%s (not found)</a>
`, html.EscapeString(relPath)))
				}
			}
			b.WriteString(`</div>
`)
		}

		b.WriteString(`</div>
`)
	}

	// Footer
	b.WriteString(`</div>
<div class="footer">
Panoptic Automated Testing Framework &mdash; Report generated automatically
</div>
</body>
</html>
`)

	return os.WriteFile(outputPath, []byte(b.String()), 0644)
}

// formatDuration formats a time.Duration into a human-readable string.
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm %ds", m, s)
}
