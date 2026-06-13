package recvalidate

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"panoptic/internal/logger"
	"panoptic/internal/ocr"
)

// ---------------------------------------------------------------------------
// Self-validation of the analyzer (§11.4.107(10), MANDATORY).
//
// The analyzer MUST pass a golden-GOOD transcript (normal model reply, no
// errors) and FAIL a golden-BAD transcript (error text such as "no provider"
// / "0 models"). A validator that passes its golden-bad fixture is itself a
// bluff. The paired-mutation test below proves the error check is the reason
// the golden-bad fails — strip the check and golden-bad starts passing.
// ---------------------------------------------------------------------------

// goldenGood is what an OCR of a healthy chat session would read: prompt,
// then a real prose model reply, intended model named, NO error tokens.
const goldenGood = `HelixCode TUI  —  model: llama3.2  (healthy)
> What is 2 + 2?
The answer is 4. Two plus two equals four, a basic arithmetic fact.
> Name a primary color.
Red is a primary color, along with blue and yellow.
`

// goldenBad is what an OCR of a broken session would read: provider/model
// failure chrome, no real reply rendered.
const goldenBad = `HelixCode TUI  —  model: Not selected
> What is 2 + 2?
Error: no provider configured. 0 models available. Provider unhealthy.
`

func goodOpts() Options {
	return Options{
		VideoPath:       "golden_good.mp4",
		ExpectedPrompts: []string{"What is 2 + 2?", "Name a primary color."},
		IntendedModel:   "llama3.2",
		MinReplyChars:   12,
	}
}

func badOpts() Options {
	return Options{
		VideoPath:       "golden_bad.mp4",
		ExpectedPrompts: []string{"What is 2 + 2?"},
		IntendedModel:   "llama3.2",
		MinReplyChars:   12,
	}
}

func TestAnalyzer_GoldenGood_Passes(t *testing.T) {
	v := NewValidator(*logger.NewLogger(false))
	rep := v.ValidateText(goldenGood, goodOpts())
	if !rep.Pass {
		t.Fatalf("golden-GOOD must PASS but failed; checks=%+v", rep.Checks)
	}
	for _, c := range rep.Checks {
		if !c.Pass {
			t.Errorf("golden-GOOD: check %q unexpectedly failed: %s", c.Name, c.Detail)
		}
	}
}

func TestAnalyzer_GoldenBad_Fails(t *testing.T) {
	v := NewValidator(*logger.NewLogger(false))
	rep := v.ValidateText(goldenBad, badOpts())
	if rep.Pass {
		t.Fatalf("golden-BAD must FAIL but PASSED — analyzer is a bluff; checks=%+v", rep.Checks)
	}
	// It must specifically fail the error-token check AND the model check AND
	// the reply check — not pass-by-accident on one.
	got := map[string]bool{}
	for _, c := range rep.Checks {
		got[c.Name] = c.Pass
	}
	if got["no_error_tokens"] {
		t.Errorf("golden-BAD: no_error_tokens check should FAIL (no provider / 0 models present)")
	}
	if got["intended_model_selected"] {
		t.Errorf("golden-BAD: model check should FAIL (model is 'Not selected')")
	}
	if got["prompt_1_has_reply"] {
		t.Errorf("golden-BAD: reply check should FAIL (only error chrome, no real reply)")
	}
}

// TestPairedMutation_ErrorCheckIsLoadBearing is the §1.1 paired mutation.
// We simulate stripping the error-detection assertion and assert that doing
// so would let the golden-BAD error-token check pass — proving the real
// check is the thing catching the bluff. (We do not mutate source here; we
// assert the discriminating property directly: the golden-BAD text DOES
// contain the error tokens the check looks for, and the golden-GOOD does not.)
func TestPairedMutation_ErrorCheckIsLoadBearing(t *testing.T) {
	v := NewValidator(*logger.NewLogger(false))

	badRep := v.ValidateText(goldenBad, badOpts())
	goodRep := v.ValidateText(goldenGood, goodOpts())

	badErr := checkByName(badRep, "no_error_tokens")
	goodErr := checkByName(goodRep, "no_error_tokens")
	if badErr == nil || goodErr == nil {
		t.Fatal("no_error_tokens check missing from a report")
	}
	if badErr.Pass {
		t.Fatal("mutation guard: golden-BAD error check must be FAIL while live")
	}
	if !goodErr.Pass {
		t.Fatal("mutation guard: golden-GOOD error check must be PASS while live")
	}
	// The mutation (removing the check) collapses this distinction: both would
	// report Pass=true. We prove the distinction exists ⇒ the check is real.
}

func TestReplyCheck_SpinnerIsNotARealReply(t *testing.T) {
	// A spinner / ellipsis after the prompt must NOT count as a reply.
	spinnerOnly := "> What is 2 + 2?\n... | / - \\ ... ⣾⣽⣻\n"
	v := NewValidator(*logger.NewLogger(false))
	rep := v.ValidateText(spinnerOnly, Options{
		ExpectedPrompts: []string{"What is 2 + 2?"},
		MinReplyChars:   12,
	})
	if checkByName(rep, "prompt_1_has_reply").Pass {
		t.Fatal("a spinner-only screen must NOT pass as a real reply (anti-bluff)")
	}
}

func checkByName(r *Report, name string) *CheckResult {
	for i := range r.Checks {
		if r.Checks[i].Name == name {
			return &r.Checks[i]
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// End-to-end OCR path (integration). Skipped honestly when ffmpeg/tesseract
// are absent — never a fake PASS (§11.4 / ErrToolAbsent contract).
// ---------------------------------------------------------------------------

func toolsPresent() bool {
	_, ffErr := exec.LookPath("ffmpeg")
	_, tsErr := exec.LookPath("tesseract")
	return ffErr == nil && tsErr == nil
}

func TestEngine_VideoToText_RealOCR(t *testing.T) {
	if !toolsPresent() {
		t.Skip("SKIP-OK: ffmpeg/tesseract absent — honest SKIP-with-reason, not a fake pass")
	}
	dir := t.TempDir()

	// Build a REAL frame PNG with high-contrast text (rendered in pure Go),
	// assemble it into a real mp4 with ffmpeg, then OCR it back.
	goodPNG := filepath.Join(dir, "good_frame.png")
	if err := renderTextFramePNG(goodPNG, []string{"The answer is 4", "Two plus two equals four"}, 6); err != nil {
		t.Fatalf("cannot synth good frame: %v", err)
	}
	mp4 := filepath.Join(dir, "session.mp4")
	if err := pngToMP4(goodPNG, mp4); err != nil {
		t.Fatalf("cannot build mp4: %v", err)
	}

	eng := ocr.NewEngine()
	frames, err := eng.VideoToText(context.Background(), mp4, filepath.Join(dir, "frames"), 1, false)
	if err != nil {
		t.Fatalf("VideoToText failed: %v", err)
	}
	if len(frames) == 0 {
		t.Fatal("no frames OCR'd from real mp4")
	}
	agg := strings.ToLower(ocr.AggregateText(frames))
	// tesseract should read at least the distinctive words back.
	if !strings.Contains(agg, "answer") && !strings.Contains(agg, "four") && !strings.Contains(agg, "two") {
		t.Fatalf("OCR did not read expected text back; got: %q", agg)
	}
}

func TestValidate_EndToEnd_GoldenGoodVideo(t *testing.T) {
	if !toolsPresent() {
		t.Skip("SKIP-OK: ffmpeg/tesseract absent — honest SKIP-with-reason, not a fake pass")
	}
	dir := t.TempDir()
	png := filepath.Join(dir, "frame.png")
	// One frame carrying prompt + real reply + model name.
	if err := renderTextFramePNG(png, []string{
		"model llama3",
		"What is 2 plus 2",
		"The answer is four equals four",
	}, 6); err != nil {
		t.Fatalf("synth frame: %v", err)
	}
	mp4 := filepath.Join(dir, "good.mp4")
	if err := pngToMP4(png, mp4); err != nil {
		t.Fatalf("mp4: %v", err)
	}
	v := NewValidator(*logger.NewLogger(false))
	rep, err := v.Validate(context.Background(), Options{
		VideoPath:       mp4,
		ExpectedPrompts: []string{"What is 2 plus 2"},
		IntendedModel:   "llama3",
		FPS:             1,
		MinReplyChars:   8,
	})
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}
	if rep.Skipped {
		t.Fatalf("unexpected skip: %s", rep.SkipReason)
	}
	if rep.FrameCount == 0 {
		t.Fatal("no frames processed")
	}
	// The error check must pass (no error chrome in this frame).
	if !checkByName(rep, "no_error_tokens").Pass {
		t.Errorf("good video flagged errors it should not have: %s", checkByName(rep, "no_error_tokens").Evidence)
	}
}

func TestValidate_ToolAbsent_IsHonestSkip(t *testing.T) {
	// Force absent tools by pointing the engine at non-existent binaries.
	v := NewValidatorWithEngine(*logger.NewLogger(false), &ocr.Engine{
		FrameTool: "ffmpeg_definitely_not_installed_xyz",
		OCRTool:   "tesseract_definitely_not_installed_xyz",
	})
	rep, err := v.Validate(context.Background(), Options{
		VideoPath:       "irrelevant.mp4",
		ExpectedPrompts: []string{"hi"},
	})
	if err != nil {
		t.Fatalf("tool-absent must be a clean Skip report, got error: %v", err)
	}
	if !rep.Skipped {
		t.Fatal("tool-absent must produce Skipped=true (honest SKIP, never fake PASS)")
	}
	if rep.Pass {
		t.Fatal("tool-absent must NOT report Pass=true")
	}
}

// --- helpers: assemble a real mp4 from a real PNG using ffmpeg ---

func pngToMP4(png, mp4 string) error {
	args := []string{"-y", "-loop", "1", "-i", png, "-t", "1", "-r", "1",
		"-pix_fmt", "yuv420p", mp4}
	out, err := exec.Command("ffmpeg", args...).CombinedOutput()
	if err != nil {
		return wrap(err, string(out))
	}
	return nil
}

func wrap(err error, ctx string) error {
	return &wrappedErr{err: err, ctx: ctx}
}

type wrappedErr struct {
	err error
	ctx string
}

func (w *wrappedErr) Error() string { return w.err.Error() + ": " + w.ctx }

// TestNoError_HealthyZeroUnhealthy_Passes is the §11.4.107(10) self-validation
// for the analyzer false-positives found while validating the HelixCode TUI
// videos: the benign "Unhealthy: 0" health readout (= ZERO unhealthy = healthy)
// and the "Not selected" pre-selection label MUST NOT trip the no-error check.
func TestNoError_HealthyZeroUnhealthy_Passes(t *testing.T) {
	healthy := "> /model\nProvider Status\nHealthy: 5\nUnhealthy: 0\nCurrent Model\nNot selected\n" +
		"> Hello there\nAI: Hi! How can I help you today, friend?\n"
	v := NewValidator(*logger.NewLogger(false))
	rep := v.ValidateText(healthy, Options{ExpectedPrompts: []string{"Hello there"}, MinReplyChars: 12})
	c := checkByName(rep, "no_error_tokens")
	if c == nil || !c.Pass {
		t.Fatalf("healthy 'Unhealthy: 0' + 'Not selected' must NOT flag an error; detail=%v", c)
	}
}

// TestNoError_NonzeroUnhealthy_Fails is the discriminating partner: a STRICTLY
// POSITIVE unhealthy count IS a real error and MUST trip the check. Strip the
// nonzeroUnhealthyRe guard and this FAILs (paired §1.1).
func TestNoError_NonzeroUnhealthy_Fails(t *testing.T) {
	unhealthy := "Provider Status\nHealthy: 2\nUnhealthy: 3\n> Hi\nAI: hello back to you my friend\n"
	v := NewValidator(*logger.NewLogger(false))
	rep := v.ValidateText(unhealthy, Options{ExpectedPrompts: []string{"Hi"}, MinReplyChars: 12})
	c := checkByName(rep, "no_error_tokens")
	if c == nil || c.Pass {
		t.Fatalf("'Unhealthy: 3' (nonzero) MUST flag an error; detail=%v", c)
	}
}

// TestReplyCheck_ErrorReplyWithChrome_Fails reproduces the exact V2 ensemble
// bluff: an error-only assistant turn ("AI: [Error: ... all members failed]")
// surrounded by static sidebar chrome MUST NOT count as a real reply. Before
// the afterReplyMarker + stripChromeLines hardening, the ambient chrome prose
// made this PASS — the analyzer self-bluff this guards (§11.4.107(10)).
func TestReplyCheck_ErrorReplyWithChrome_Fails(t *testing.T) {
	errReply := "> Do you see my codebase?\n" +
		"AI: [Error: helix-agent ensemble: all 4 member(s) failed (first error: groq request failed)]\n" +
		"Provider Status\nHealthy: 5\nUnhealthy: 0\nChat Statistics\nMessages: 1\n" +
		"/help - Show help\n/clear - Clear chat\nWelcome to HelixCode AI Chat\n"
	v := NewValidator(*logger.NewLogger(false))
	rep := v.ValidateText(errReply, Options{
		ExpectedPrompts:    []string{"Do you see my codebase?"},
		MinReplyChars:      12,
		ChromeLinePatterns: testChromePatterns(),
	})
	c := checkByName(rep, "prompt_1_has_reply")
	if c == nil || c.Pass {
		t.Fatalf("an error-only reply + chrome MUST NOT pass as a real reply; detail=%v", c)
	}
}

// testChromePatterns are the consumer-supplied chrome patterns a chat-TUI caller
// passes (the patterns live HERE / at the call site, NOT in Panoptic source —
// CONST-051(B): no consumer UI strings in the reusable submodule).
func testChromePatterns() []string {
	return []string{
		`^\s*/[a-z]+`,
		`provider status|chat statistics|current model|available models`,
		`healthy:|unhealthy:|total:|tokens used|messages:|llm info|select model`,
		`press enter|welcome to|start a conversation|use the buttons|status:`,
		`quick actions|recent activity|workers:|tasks:|system:`,
	}
}

// TestReplyCheck_RealReplyAfterMarker_Passes is the golden-good partner: a real
// prose answer after the AI: marker (with the same surrounding chrome) MUST
// pass — proving the hardening rejects errors without rejecting real replies.
func TestReplyCheck_RealReplyAfterMarker_Passes(t *testing.T) {
	realReply := "> What is 2 plus 2?\n" +
		"AI: 2 plus 2 equals 4. That is a basic arithmetic sum.\n" +
		"Provider Status\nHealthy: 5\nUnhealthy: 0\nChat Statistics\nMessages: 1\n/help - Show help\n"
	v := NewValidator(*logger.NewLogger(false))
	rep := v.ValidateText(realReply, Options{ExpectedPrompts: []string{"What is 2 plus 2?"}, MinReplyChars: 12})
	c := checkByName(rep, "prompt_1_has_reply")
	if c == nil || !c.Pass {
		t.Fatalf("a real prose reply after AI: MUST pass; detail=%v", c)
	}
}
