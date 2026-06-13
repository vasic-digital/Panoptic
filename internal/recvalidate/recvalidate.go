// Package recvalidate validates recorded TUI/chat session videos.
//
// It answers, for an mp4 (or webm) recording of a terminal chat where prompts
// are typed and an LLM replies, three questions with captured evidence:
//
//  1. Did the chat render a REAL model reply to each expected prompt
//     (actual prose text on screen, not a blank pane or a spinner)?
//  2. Did NO error/warning text appear on screen during the interaction
//     (e.g. "error", "failed", "unhealthy", "no provider", "0 models",
//     "Not selected", "panic")?
//  3. Was the intended model selected (its name visible on screen)?
//
// It is built ON TOP of Panoptic's existing capabilities (§11.4.74 reuse):
//   - internal/ocr     — REAL frame extraction (ffmpeg) + REAL OCR (tesseract).
//   - internal/ai      — ErrorDetector, the existing regex error classifier,
//     is reused verbatim to flag error-category tokens in the OCR'd text.
//
// Decoupling (CONST-051(B)): no consumer-specific path/app/prompt is hardcoded.
// The caller supplies the video path, the expected prompts, the intended
// model name, and (optionally) extra error tokens. The package is reusable by
// any project that records a chat-style TUI to video.
package recvalidate

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"panoptic/internal/ai"
	"panoptic/internal/logger"
	"panoptic/internal/ocr"
)

// chatErrorTokens are chat/TUI-specific failure phrases that Panoptic's
// generic ErrorDetector does not already cover. They are matched
// case-insensitively against the OCR'd screen text. This list is generic
// (provider/model/health wording common to LLM TUIs), not consumer-specific.
var chatErrorTokens = []string{
	"no provider",
	"no providers",
	"0 models",
	"no models",
	"not selected",
	"unhealthy",
	"provider error",
	"model error",
	"panic",
	"traceback",
	"stack trace",
	"connection refused",
	"rate limit",
	"quota exceeded",
}

// Frame extraction sampling default (frames per video-second).
const defaultFPS = 1.0

// Options configures a validation run.
type Options struct {
	// VideoPath is the recording to validate (mp4/webm/...).
	VideoPath string
	// ExpectedPrompts is the ordered list of prompts that were typed. The
	// validator asserts, for each, that a real reply followed it on screen.
	ExpectedPrompts []string
	// IntendedModel, if non-empty, MUST appear in the on-screen text.
	IntendedModel string
	// ExtraErrorTokens are additional case-insensitive failure phrases to
	// flag (lets a consumer add app-specific error wording without forking).
	ExtraErrorTokens []string
	// FPS controls frame sampling density (<=0 => 1 fps).
	FPS float64
	// FrameDir is where frames are written. If empty, a temp dir is used.
	FrameDir string
	// KeepFrames retains extracted frames as evidence (default false).
	KeepFrames bool
	// MinReplyChars is the minimum number of prose characters that must
	// appear AFTER a prompt on screen for the reply to count as "real".
	// Defaults to 12 if <=0.
	MinReplyChars int
}

// CheckResult is one assertion's verdict.
type CheckResult struct {
	Name     string `json:"name"`
	Pass     bool   `json:"pass"`
	Detail   string `json:"detail"`
	Evidence string `json:"evidence,omitempty"` // text/frame backing this verdict
}

// Report is the structured PASS/FAIL output of a validation run.
type Report struct {
	Pass            bool             `json:"pass"`
	Skipped         bool             `json:"skipped"`
	SkipReason      string           `json:"skip_reason,omitempty"`
	VideoPath       string           `json:"video_path"`
	FrameCount      int              `json:"frame_count"`
	FramesDir       string           `json:"frames_dir,omitempty"`
	AggregatedText  string           `json:"aggregated_text"`
	Checks          []CheckResult    `json:"checks"`
	DetectedErrors  []ai.DetectedError `json:"detected_errors,omitempty"`
}

// Validator drives a recorded-video validation using the OCR engine and the
// reused Panoptic ErrorDetector.
type Validator struct {
	engine   *ocr.Engine
	detector *ai.ErrorDetector
	log      logger.Logger
}

// NewValidator builds a Validator with the default OCR engine + a fresh
// Panoptic ErrorDetector (reusing the existing regex classifier).
func NewValidator(log logger.Logger) *Validator {
	return &Validator{
		engine:   ocr.NewEngine(),
		detector: ai.NewErrorDetector(log),
		log:      log,
	}
}

// NewValidatorWithEngine allows injecting a custom OCR engine (e.g. pointing
// at vendored binaries or a deterministic test stub).
func NewValidatorWithEngine(log logger.Logger, engine *ocr.Engine) *Validator {
	return &Validator{
		engine:   engine,
		detector: ai.NewErrorDetector(log),
		log:      log,
	}
}

// Validate runs the full pipeline against a video file and returns a Report.
// A missing ffmpeg/tesseract yields a Skipped report with a reason — never a
// fake PASS and never a hard error masquerading as a feature failure.
func (v *Validator) Validate(ctx context.Context, opts Options) (*Report, error) {
	fps := opts.FPS
	if fps <= 0 {
		fps = defaultFPS
	}
	frameDir := opts.FrameDir
	if frameDir == "" {
		frameDir = fmt.Sprintf("%s/panoptic_recvalidate_%d", tmpDir(), time.Now().UnixNano())
	}

	if ff, ts := v.engine.ToolsAvailable(); !ff || !ts {
		var missing []string
		if !ff {
			missing = append(missing, v.engine.FrameTool)
		}
		if !ts {
			missing = append(missing, v.engine.OCRTool)
		}
		return &Report{
			Pass:       false,
			Skipped:    true,
			SkipReason: "required OCR/frame tools absent: " + strings.Join(missing, ", "),
			VideoPath:  opts.VideoPath,
		}, nil
	}

	frames, err := v.engine.VideoToText(ctx, opts.VideoPath, frameDir, fps, opts.KeepFrames)
	if err != nil {
		return nil, fmt.Errorf("video OCR failed: %w", err)
	}
	aggregated := ocr.AggregateText(frames)

	rep := &Report{
		Pass:           true, // flips false on any failed check
		VideoPath:      opts.VideoPath,
		FrameCount:     len(frames),
		AggregatedText: aggregated,
	}
	if opts.KeepFrames {
		rep.FramesDir = frameDir
	}

	rep.runChecks(v, aggregated, opts)
	return rep, nil
}

// ValidateText runs the assertion layer against already-OCR'd text. It is the
// testable, tool-independent core (the self-validation golden fixtures feed
// text in directly, so the analyzer is provable without ffmpeg/tesseract).
func (v *Validator) ValidateText(aggregated string, opts Options) *Report {
	rep := &Report{
		Pass:           true,
		VideoPath:      opts.VideoPath,
		AggregatedText: aggregated,
	}
	rep.runChecks(v, aggregated, opts)
	return rep
}

// runChecks performs the three assertion families against the OCR'd text.
func (r *Report) runChecks(v *Validator, aggregated string, opts Options) {
	lower := strings.ToLower(aggregated)

	// CHECK 1: No error/warning tokens on screen.
	// 1a — reuse Panoptic's ErrorDetector (regex classifier) on the screen text.
	detected := v.detector.DetectErrors([]ai.ErrorMessage{{
		Message:   aggregated,
		Source:    "recorded-video-ocr",
		Timestamp: time.Now(),
		Level:     "screen",
	}})
	r.DetectedErrors = detected

	// 1b — chat/TUI-specific failure phrases.
	errTokens := append([]string{}, chatErrorTokens...)
	errTokens = append(errTokens, opts.ExtraErrorTokens...)
	var hitTokens []string
	for _, tok := range errTokens {
		if tok == "" {
			continue
		}
		if strings.Contains(lower, strings.ToLower(tok)) {
			hitTokens = append(hitTokens, tok)
		}
	}

	noErrPass := len(detected) == 0 && len(hitTokens) == 0
	detail := "no error/warning tokens detected on screen"
	evidence := ""
	if !noErrPass {
		var parts []string
		if len(detected) > 0 {
			names := make([]string, 0, len(detected))
			for _, d := range detected {
				names = append(names, d.Name+"("+d.Category+")")
			}
			parts = append(parts, "ErrorDetector="+strings.Join(names, ","))
		}
		if len(hitTokens) > 0 {
			parts = append(parts, "tokens="+strings.Join(hitTokens, ","))
		}
		detail = "error/warning text present on screen: " + strings.Join(parts, "; ")
		evidence = excerptAround(aggregated, hitTokens, detected)
	}
	r.add(CheckResult{Name: "no_error_tokens", Pass: noErrPass, Detail: detail, Evidence: evidence})

	// CHECK 2: Each expected prompt has a real reply (prose after the prompt).
	minReply := opts.MinReplyChars
	if minReply <= 0 {
		minReply = 12
	}
	for i, prompt := range opts.ExpectedPrompts {
		ok, ev := promptHasReply(aggregated, prompt, minReply, errTokens)
		name := fmt.Sprintf("prompt_%d_has_reply", i+1)
		d := "real prose reply present after prompt"
		if !ok {
			d = "no real reply found after prompt (blank/spinner-only)"
		}
		r.add(CheckResult{Name: name, Pass: ok, Detail: d, Evidence: ev})
	}

	// CHECK 3: Intended model is visible on screen (if requested).
	// OCR routinely mangles model strings ("llama3.2" -> "llamas,2"), so we
	// match the alphanumeric core fuzzily rather than by exact substring.
	if opts.IntendedModel != "" {
		ok, ev := fuzzyModelVisible(aggregated, opts.IntendedModel)
		d := "intended model name visible on screen: " + opts.IntendedModel
		if !ok {
			d = "intended model name NOT visible on screen: " + opts.IntendedModel
		}
		r.add(CheckResult{Name: "intended_model_selected", Pass: ok, Detail: d, Evidence: ev})
	}
	_ = lower
}

func (r *Report) add(c CheckResult) {
	r.Checks = append(r.Checks, c)
	if !c.Pass {
		r.Pass = false
	}
}

// promptHasReply reports whether real prose text follows the prompt on screen.
// It locates the prompt text in the aggregated transcript and checks that the
// content after it contains at least minChars of alphabetic prose that is not
// merely a spinner/ellipsis, the prompt echoed back, OR error chrome. Error
// chrome ("Error: no provider ...") must NOT count as a real model reply, so
// any error token present in the post-prompt region is excised before the
// prose count — otherwise a failure message would masquerade as an answer.
func promptHasReply(aggregated, prompt string, minChars int, errTokens []string) (bool, string) {
	lowerPrompt := strings.ToLower(strings.TrimSpace(prompt))
	if lowerPrompt == "" {
		return false, "empty prompt"
	}
	// Locate the prompt by OCR-tolerant token overlap rather than exact
	// substring — OCR substitutes characters ("What" -> "Mat"), so an exact
	// match would false-FAIL a perfectly good recording.
	endIdx := locatePromptEnd(aggregated, prompt)
	if endIdx < 0 {
		// Prompt itself never appeared on screen — cannot confirm a reply.
		return false, "prompt text not found on screen"
	}
	after := aggregated[endIdx:]
	// Bound the reply region to before the NEXT prompt marker so the answer to
	// a later prompt cannot satisfy an earlier one (best-effort: cut at "\n>").
	if cut := strings.Index(after, "\n>"); cut > 0 {
		after = after[:cut]
	}
	// Excise error-token content + generic error words so failure chrome can
	// never count as a real reply (anti-bluff: an error IS NOT an answer).
	cleaned := stripErrorChrome(after, errTokens)
	// Strip spinner/placeholder noise so a spinner can never count as a reply.
	cleaned = stripNoise(cleaned)
	prose := proseChars(cleaned)
	if prose >= minChars {
		ev := strings.TrimSpace(firstN(cleaned, 160))
		return true, ev
	}
	return false, fmt.Sprintf("only %d real-prose chars after prompt (need %d; error chrome excluded)", prose, minChars)
}

// stripErrorChrome removes any line containing an error token or a generic
// error indicator from the candidate reply region. A reply that is ENTIRELY
// error chrome therefore counts as zero prose.
func stripErrorChrome(s string, errTokens []string) string {
	tokens := append([]string{}, errTokens...)
	tokens = append(tokens, "error", "failed", "failure", "exception",
		"unhealthy", "cannot", "unable", "panic")
	var keep []string
	for _, line := range strings.Split(s, "\n") {
		low := strings.ToLower(line)
		drop := false
		for _, tok := range tokens {
			if tok != "" && strings.Contains(low, strings.ToLower(tok)) {
				drop = true
				break
			}
		}
		if !drop {
			keep = append(keep, line)
		}
	}
	return strings.Join(keep, "\n")
}

var spinnerRe = regexp.MustCompile(`[\.\|/\\\-_•▏▎▍▌▋▊▉█◐◓◑◒⣾⣽⣻⢿⡿⣟⣯⣷\s]`)

// stripNoise removes spinner glyphs / dots / whitespace runs.
func stripNoise(s string) string {
	// Replace spinner-class runs with single spaces, preserving real words.
	return spinnerRe.ReplaceAllString(s, " ")
}

// proseChars counts letters (Unicode) — a proxy for "real words on screen".
func proseChars(s string) int {
	n := 0
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			n++
		}
	}
	return n
}

func firstN(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// excerptAround returns a short excerpt of the text near the first error hit,
// for the evidence field.
func excerptAround(text string, tokens []string, detected []ai.DetectedError) string {
	lower := strings.ToLower(text)
	probe := ""
	if len(tokens) > 0 {
		probe = strings.ToLower(tokens[0])
	} else if len(detected) > 0 {
		probe = strings.ToLower(firstWord(detected[0].Message))
	}
	if probe == "" {
		return firstN(text, 160)
	}
	idx := strings.Index(lower, probe)
	if idx < 0 {
		return firstN(text, 160)
	}
	start := idx - 40
	if start < 0 {
		start = 0
	}
	end := idx + 80
	if end > len(text) {
		end = len(text)
	}
	return strings.TrimSpace(text[start:end])
}

func firstWord(s string) string {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

func tmpDir() string {
	return osTempDir()
}

// --- OCR-tolerant fuzzy matching helpers ---

var wordRe = regexp.MustCompile(`[A-Za-z0-9]+`)

// locatePromptEnd returns the byte offset in `text` just AFTER the line on
// which the prompt is most strongly matched (by word overlap), or -1 if the
// prompt's words do not sufficiently appear on any single line. Returning the
// end of the matched line lets the reply check inspect what follows.
func locatePromptEnd(text, prompt string) int {
	pWords := lowerWords(prompt)
	if len(pWords) == 0 {
		return -1
	}
	want := map[string]bool{}
	for _, w := range pWords {
		want[w] = true
	}
	// Require a majority of prompt words (>=60%, min 1) on one line. OCR drops
	// a few characters but rarely a majority of multi-word prompts.
	threshold := (len(want)*6 + 9) / 10 // ceil(0.6*n)
	if threshold < 1 {
		threshold = 1
	}

	offset := 0
	best := -1
	bestHits := 0
	for _, line := range strings.Split(text, "\n") {
		lineWords := wordRe.FindAllString(strings.ToLower(line), -1)
		hits := 0
		seen := map[string]bool{}
		for _, lw := range lineWords {
			if want[lw] && !seen[lw] {
				seen[lw] = true
				hits++
			}
		}
		if hits >= threshold && hits > bestHits {
			bestHits = hits
			best = offset + len(line) + 1 // +1 for the consumed newline
		}
		offset += len(line) + 1
	}
	return best
}

// fuzzyModelVisible reports whether the model name's alphanumeric core appears
// on screen, tolerating OCR substitutions. It strips non-alphanumerics from
// both the target and each on-screen token and accepts a token whose
// normalized edit distance to the target is small.
func fuzzyModelVisible(text, model string) (bool, string) {
	target := normAlnum(model)
	if target == "" {
		return false, ""
	}
	for _, tok := range wordRe.FindAllString(text, -1) {
		cand := normAlnum(tok)
		if cand == "" {
			continue
		}
		if cand == target {
			return true, tok
		}
		// Allow up to ceil(len/5) substitutions (OCR error budget).
		budget := (len(target) + 4) / 5
		if budget < 1 {
			budget = 1
		}
		if abs(len(cand)-len(target)) <= budget && levenshtein(cand, target) <= budget {
			return true, tok + " (~" + model + ")"
		}
	}
	return false, ""
}

func lowerWords(s string) []string {
	return wordRe.FindAllString(strings.ToLower(s), -1)
}

func normAlnum(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// levenshtein is a standard edit-distance for the small OCR error budget.
func levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	prev := make([]int, lb+1)
	cur := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		cur[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min3(cur[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[lb]
}

func min3(a, b, c int) int {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
}
