package ocr

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"
)

func TestToolsAvailable_ReportsHostState(t *testing.T) {
	e := NewEngine()
	ff, ts := e.ToolsAvailable()
	_, ffErr := exec.LookPath("ffmpeg")
	_, tsErr := exec.LookPath("tesseract")
	if ff != (ffErr == nil) {
		t.Errorf("ffmpeg availability mismatch: reported %v, LookPath err=%v", ff, ffErr)
	}
	if ts != (tsErr == nil) {
		t.Errorf("tesseract availability mismatch: reported %v, LookPath err=%v", ts, tsErr)
	}
}

func TestOCRImage_AbsentTool_ReturnsErrToolAbsent(t *testing.T) {
	e := &Engine{OCRTool: "tesseract_absent_xyz123", FrameTool: "ffmpeg"}
	_, err := e.OCRImage(context.Background(), "whatever.png")
	if !errors.Is(err, ErrToolAbsent) {
		t.Fatalf("expected ErrToolAbsent, got %v", err)
	}
}

func TestExtractFrames_AbsentTool_ReturnsErrToolAbsent(t *testing.T) {
	e := &Engine{FrameTool: "ffmpeg_absent_xyz123", OCRTool: "tesseract"}
	_, err := e.ExtractFrames(context.Background(), "whatever.mp4", t.TempDir(), 1)
	if !errors.Is(err, ErrToolAbsent) {
		t.Fatalf("expected ErrToolAbsent, got %v", err)
	}
}

func TestVideoToText_AbsentTool_ReturnsErrToolAbsent(t *testing.T) {
	e := &Engine{FrameTool: "ffmpeg_absent_xyz123", OCRTool: "tesseract_absent_xyz123"}
	_, err := e.VideoToText(context.Background(), "x.mp4", t.TempDir(), 1, false)
	if !errors.Is(err, ErrToolAbsent) {
		t.Fatalf("expected ErrToolAbsent, got %v", err)
	}
}

func TestAggregateText_DedupesAndDropsBlankLines(t *testing.T) {
	frames := []FrameText{
		{Index: 0, Text: "model llama3\n\n  \nready"},
		{Index: 1, Text: "ready\nThe answer is 4"},  // "ready" dedup'd
		{Index: 2, Text: "The answer is 4\nThe answer is 4"}, // dup dropped
	}
	got := AggregateText(frames)
	lines := strings.Split(strings.TrimSpace(got), "\n")
	want := []string{"model llama3", "ready", "The answer is 4"}
	if len(lines) != len(want) {
		t.Fatalf("expected %d lines, got %d: %q", len(want), len(lines), got)
	}
	for i := range want {
		if strings.TrimSpace(lines[i]) != want[i] {
			t.Errorf("line %d = %q, want %q", i, lines[i], want[i])
		}
	}
}
