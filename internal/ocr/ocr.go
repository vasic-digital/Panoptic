// Package ocr provides image- and video-frame text extraction for Panoptic.
//
// It is the missing capability that turns Panoptic's existing heuristic
// element detector (internal/vision) and text-based error classifier
// (internal/ai) into a full "read the on-screen text" pipeline: given a
// PNG/JPEG frame OR an mp4/webm recording, it produces the actual prose the
// user would have seen on screen.
//
// Design constraints (CONST-051(B) decoupling / §11.4.74 reuse-don't-fake):
//   - This package is fully project-agnostic. It hardcodes NO consumer path,
//     hostname, app name, or prompt. Everything is passed in as parameters.
//   - It performs REAL OCR by shelling out to the `tesseract` binary and REAL
//     frame extraction via the `ffmpeg` binary — there is no simulated text,
//     no hardcoded "detected" string. Both tools are pure CLI invocations
//     (no cgo) so the package builds and tests everywhere.
//   - When a required external tool is absent, the package returns a typed
//     ErrToolAbsent rather than fabricating a PASS. Callers MUST surface this
//     as an honest SKIP-with-reason, never a fake success (§11.4 / §11.4.107).
package ocr

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// ErrToolAbsent is returned (wrapped) when a required external binary
// (ffmpeg or tesseract) is not present on the host PATH. Callers MUST treat
// this as a reason to SKIP-with-reason, NOT as a pass and NOT as a hard
// failure of the feature under test.
var ErrToolAbsent = errors.New("required external tool absent")

// Engine performs OCR and frame extraction via external CLI tools.
//
// FrameTool / OCRTool default to "ffmpeg" / "tesseract" but are overridable
// so a consumer can point at a vendored binary or a stub in tests.
type Engine struct {
	FrameTool string // binary used to extract frames (default "ffmpeg")
	OCRTool   string // binary used to OCR a frame (default "tesseract")
	Language  string // tesseract language, default "eng"
}

// NewEngine builds an Engine with default tool names.
func NewEngine() *Engine {
	return &Engine{FrameTool: "ffmpeg", OCRTool: "tesseract", Language: "eng"}
}

// frameToolName / ocrToolName apply defaults without mutating the receiver.
func (e *Engine) frameToolName() string {
	if e.FrameTool != "" {
		return e.FrameTool
	}
	return "ffmpeg"
}

func (e *Engine) ocrToolName() string {
	if e.OCRTool != "" {
		return e.OCRTool
	}
	return "tesseract"
}

func (e *Engine) lang() string {
	if e.Language != "" {
		return e.Language
	}
	return "eng"
}

// ToolsAvailable reports which external tools are present. A caller can use
// this up front to decide PASS/FAIL vs SKIP-with-reason before doing work.
func (e *Engine) ToolsAvailable() (ffmpeg bool, tesseract bool) {
	_, ffErr := exec.LookPath(e.frameToolName())
	_, tsErr := exec.LookPath(e.ocrToolName())
	return ffErr == nil, tsErr == nil
}

// FrameText pairs an extracted frame's file path with the text read from it.
type FrameText struct {
	FramePath string `json:"frame_path"`
	Index     int    `json:"index"`
	Text      string `json:"text"`
}

// OCRImage runs tesseract on a single image file and returns the recognised
// text. Returns a wrapped ErrToolAbsent if tesseract is not installed.
//
// Light-on-dark robustness (the dark-theme TUI fix): tesseract is trained for
// dark text on a light background. A light-on-dark frame (the common terminal
// / TUI theme) OCRs to little-or-nothing — the feature is plainly visible to a
// human yet "invisible" to the analyzer, a §11.4.107(10) analyzer-self-bluff.
// To close that gap GENERICALLY (no consumer specifics, helps any dark-theme
// UI per CONST-051(B)), OCRImage:
//   - always OCRs the frame as-is (raw), AND
//   - when the frame is predominantly DARK (detected via mean luma), ALSO OCRs
//     an upscaled + grayscale + negated variant and UNIONS the recognised text.
//
// The union never LOSES text a light frame already yields (raw is always run),
// and never FABRICATES text (a blank dark frame negates to a blank light frame
// → still empty), so it cannot introduce a false-positive. The negate pass is
// skipped entirely on light frames, so light-theme recordings pay no extra OCR.
func (e *Engine) OCRImage(ctx context.Context, imagePath string) (string, error) {
	tool := e.ocrToolName()
	if _, err := exec.LookPath(tool); err != nil {
		return "", fmt.Errorf("%w: %s", ErrToolAbsent, tool)
	}
	if _, err := os.Stat(imagePath); err != nil {
		return "", fmt.Errorf("image not found: %s: %w", imagePath, err)
	}

	rawText, rawErr := e.runTesseract(ctx, imagePath)
	if rawErr != nil {
		return "", rawErr
	}

	// Augment dark frames with a negated pass. If the frame-transform tool
	// (ffmpeg) is unavailable, or the frame is not dark, the raw text stands —
	// no fabrication, no hard failure (the transform is an enhancement, not a
	// gate).
	if e.frameIsDark(ctx, imagePath) {
		if negText, ok := e.ocrNegated(ctx, imagePath); ok {
			return unionText(rawText, negText), nil
		}
	}
	return rawText, nil
}

// runTesseract runs `tesseract <img> stdout -l <lang>` and returns its text.
func (e *Engine) runTesseract(ctx context.Context, imagePath string) (string, error) {
	cmd := exec.CommandContext(ctx, e.ocrToolName(), imagePath, "stdout", "-l", e.lang())
	out, err := cmd.Output()
	if err != nil {
		// tesseract exits non-zero on a genuinely unreadable/empty input on
		// some builds; surface stderr for diagnosis but do not fabricate text.
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return "", fmt.Errorf("tesseract failed on %s: %v: %s", imagePath, err, string(ee.Stderr))
		}
		return "", fmt.Errorf("tesseract failed on %s: %w", imagePath, err)
	}
	return string(out), nil
}

// darkLumaThreshold is the mean-luma (0..255) ceiling below which a frame is
// treated as "predominantly dark" and given the negate pass. 96 (~37%) cleanly
// separates the common dark TUI/terminal themes (typically <50) from light
// themes (typically >180) with a wide safety margin.
const darkLumaThreshold = 96.0

// frameIsDark reports whether the frame's mean luma is below darkLumaThreshold,
// using ffmpeg's signalstats (YAVG). When ffmpeg is unavailable or the probe
// cannot be parsed, it returns false (raw OCR stands — no enhancement, but no
// failure either).
func (e *Engine) frameIsDark(ctx context.Context, imagePath string) bool {
	tool := e.frameToolName()
	if _, err := exec.LookPath(tool); err != nil {
		return false
	}
	// signalstats prints "lavfi.signalstats.YAVG=<float>" to stderr via the
	// metadata filter; route it to a null muxer so no file is produced.
	cmd := exec.CommandContext(ctx, tool,
		"-hide_banner",
		"-i", imagePath,
		"-vf", "signalstats,metadata=print:key=lavfi.signalstats.YAVG",
		"-f", "null", "-",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	yavg, ok := parseYAVG(string(out))
	if !ok {
		return false
	}
	return yavg < darkLumaThreshold
}

// parseYAVG extracts the YAVG float from ffmpeg signalstats output.
func parseYAVG(s string) (float64, bool) {
	const key = "lavfi.signalstats.YAVG="
	idx := strings.LastIndex(s, key)
	if idx < 0 {
		return 0, false
	}
	rest := s[idx+len(key):]
	// Take the numeric run up to the first non-number character.
	end := 0
	for end < len(rest) {
		c := rest[end]
		if (c >= '0' && c <= '9') || c == '.' || c == '-' || c == '+' {
			end++
			continue
		}
		break
	}
	v, err := strconv.ParseFloat(strings.TrimSpace(rest[:end]), 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

// ocrNegated transcodes the frame through an upscale (3x) + grayscale + negate
// filter into a temp PNG and OCRs it. Returns the recognised text and true on
// success; ("", false) when the transform or OCR could not run (so the caller
// falls back to the raw text — never a hard failure, never fabrication).
func (e *Engine) ocrNegated(ctx context.Context, imagePath string) (string, bool) {
	tool := e.frameToolName()
	if _, err := exec.LookPath(tool); err != nil {
		return "", false
	}
	tmpDir, err := os.MkdirTemp("", "panoptic-ocr-neg-")
	if err != nil {
		return "", false
	}
	defer os.RemoveAll(tmpDir)
	out := filepath.Join(tmpDir, "negated.png")

	cmd := exec.CommandContext(ctx, tool,
		"-hide_banner", "-loglevel", "error",
		"-i", imagePath,
		"-vf", "scale=iw*3:ih*3:flags=lanczos,format=gray,negate",
		out, "-y",
	)
	if err := cmd.Run(); err != nil {
		return "", false
	}
	if fi, statErr := os.Stat(out); statErr != nil || fi.Size() == 0 {
		return "", false
	}
	text, ocrErr := e.runTesseract(ctx, out)
	if ocrErr != nil {
		return "", false
	}
	return text, true
}

// unionText concatenates two OCR passes' text so a marker readable in EITHER
// pass is present in the result. A trailing newline separates them so line-
// oriented consumers (AggregateText) see both passes' lines.
func unionText(a, b string) string {
	a = strings.TrimRight(a, "\n")
	b = strings.TrimRight(b, "\n")
	switch {
	case a == "":
		return b
	case b == "":
		return a
	default:
		return a + "\n" + b
	}
}

// ExtractFrames pulls representative frames from a video into outDir using
// ffmpeg. fps controls sampling density (frames per second of video). A
// value <= 0 defaults to 1 fps. Returns the sorted list of frame file paths.
//
// Returns a wrapped ErrToolAbsent if ffmpeg is not installed.
func (e *Engine) ExtractFrames(ctx context.Context, videoPath, outDir string, fps float64) ([]string, error) {
	tool := e.frameToolName()
	if _, err := exec.LookPath(tool); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrToolAbsent, tool)
	}
	if _, err := os.Stat(videoPath); err != nil {
		return nil, fmt.Errorf("video not found: %s: %w", videoPath, err)
	}
	if fps <= 0 {
		fps = 1.0
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return nil, fmt.Errorf("cannot create frame dir %s: %w", outDir, err)
	}
	pattern := filepath.Join(outDir, "frame_%05d.png")
	// -vf fps=N samples N frames per video-second. -y overwrites.
	cmd := exec.CommandContext(ctx, tool,
		"-y",
		"-i", videoPath,
		"-vf", "fps="+strconv.FormatFloat(fps, 'f', -1, 64),
		pattern,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("ffmpeg frame extraction failed for %s: %w: %s", videoPath, err, string(out))
	}
	frames, err := filepath.Glob(filepath.Join(outDir, "frame_*.png"))
	if err != nil {
		return nil, fmt.Errorf("cannot list extracted frames: %w", err)
	}
	if len(frames) == 0 {
		return nil, fmt.Errorf("ffmpeg produced no frames from %s (video may be empty or zero-length)", videoPath)
	}
	sort.Strings(frames)
	return frames, nil
}

// VideoToText extracts frames from a video and OCRs each one, returning a
// FrameText per frame. It is the high-level "read every screen of this
// recording" entry point. Both ffmpeg and tesseract must be present; a
// missing tool yields a wrapped ErrToolAbsent.
//
// keepFrames=false deletes the temporary frame directory before returning.
func (e *Engine) VideoToText(ctx context.Context, videoPath, frameDir string, fps float64, keepFrames bool) ([]FrameText, error) {
	frames, err := e.ExtractFrames(ctx, videoPath, frameDir, fps)
	if err != nil {
		return nil, err
	}
	if !keepFrames {
		defer os.RemoveAll(frameDir)
	}
	results := make([]FrameText, 0, len(frames))
	for i, f := range frames {
		text, ocrErr := e.OCRImage(ctx, f)
		if ocrErr != nil {
			// A tool-absent OCR error aborts the whole run honestly.
			if errors.Is(ocrErr, ErrToolAbsent) {
				return nil, ocrErr
			}
			// A per-frame OCR failure is recorded as empty text rather than
			// killing the run — a blank/garbled frame is itself a signal.
			text = ""
		}
		results = append(results, FrameText{FramePath: f, Index: i, Text: text})
	}
	return results, nil
}

// AggregateText joins the text of every frame, deduplicating consecutive
// identical lines so the result reads like the session transcript rather
// than N repeats of a static screen.
func AggregateText(frames []FrameText) string {
	var b strings.Builder
	var lastLine string
	for _, ft := range frames {
		for _, raw := range strings.Split(ft.Text, "\n") {
			line := strings.TrimRight(raw, " \t\r")
			if strings.TrimSpace(line) == "" {
				continue
			}
			if line == lastLine {
				continue
			}
			b.WriteString(line)
			b.WriteByte('\n')
			lastLine = line
		}
	}
	return b.String()
}
