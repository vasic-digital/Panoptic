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
func (e *Engine) OCRImage(ctx context.Context, imagePath string) (string, error) {
	tool := e.ocrToolName()
	if _, err := exec.LookPath(tool); err != nil {
		return "", fmt.Errorf("%w: %s", ErrToolAbsent, tool)
	}
	if _, err := os.Stat(imagePath); err != nil {
		return "", fmt.Errorf("image not found: %s: %w", imagePath, err)
	}
	// `tesseract <img> stdout -l <lang>` prints recognised text to stdout.
	cmd := exec.CommandContext(ctx, tool, imagePath, "stdout", "-l", e.lang())
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
