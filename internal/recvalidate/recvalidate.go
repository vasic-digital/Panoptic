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
	"provider error",
	"model error",
	"panic",
	"traceback",
	"stack trace",
	"connection refused",
	"rate limit",
	"quota exceeded",
}

// nonzeroUnhealthyRe matches a provider-health readout reporting a STRICTLY
// POSITIVE unhealthy count ("Unhealthy: 3"). The benign healthy state
// "Unhealthy: 0" MUST NOT trip the no-error check — a bare "unhealthy" substring
// match would false-flag a perfectly healthy TUI (the §11.4.107(10) analyzer
// self-bluff this guards against). OCR may render the colon as a stray glyph, so
// the separator is tolerant.
var nonzeroUnhealthyRe = regexp.MustCompile(`(?i)unhealthy\s*[:;.]?\s*[1-9]`)

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
	// ChromeLinePatterns are CONSUMER-SUPPLIED case-insensitive regex patterns
	// matching ambient UI chrome lines (sidebar labels, status panels, command
	// lists) that the full-frame OCR interleaves into the reply region and that
	// MUST NOT be miscounted as model-reply prose. Panoptic ships ZERO defaults
	// here — chrome is application-specific, so the consuming project passes its
	// own patterns (CONST-051(B) decoupling — no consumer UI strings live in this
	// reusable submodule).
	ChromeLinePatterns []string
	// ReplyMarkers are the assistant-turn prefixes the chat UI renders before a
	// model reply (e.g. "AI:"). Defaults to generic chat conventions when empty;
	// a consumer with a different convention overrides it. Matched
	// case-insensitively.
	ReplyMarkers []string
	// ErrorScopeReplies, when true, restricts the built-in error/warning scan
	// (the ErrorDetector classifier + chat error tokens) to the assistant-REPLY
	// regions only (the text AFTER each ReplyMarker), instead of the whole frame.
	//
	// WHY (CONST-051(B) consumer choice): a recording often includes incidental
	// terminal SCROLLBACK that pre-dates the session — startup warnings, a redis
	// connection log, a stray JSON error payload — none of which mean the feature
	// under test is broken. A bank that asserts STRUCTURAL on-screen presence
	// (e.g. "the ensemble members panel rendered") legitimately must NOT FAIL
	// because the operator's terminal printed a warning before the TUI launched.
	// Reply-scoping flags errors the MODEL actually emitted while still ignoring
	// ambient scrollback noise. DEFAULT is false (whole-frame scan, unchanged) —
	// the consumer opts in; this never weakens the default behaviour. Genuine
	// errors that appear inside an assistant reply still FAIL under either mode.
	ErrorScopeReplies bool
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
	Pass           bool               `json:"pass"`
	Skipped        bool               `json:"skipped"`
	SkipReason     string             `json:"skip_reason,omitempty"`
	VideoPath      string             `json:"video_path"`
	FrameCount     int                `json:"frame_count"`
	FramesDir      string             `json:"frames_dir,omitempty"`
	AggregatedText string             `json:"aggregated_text"`
	Checks         []CheckResult      `json:"checks"`
	DetectedErrors []ai.DetectedError `json:"detected_errors,omitempty"`
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

	replyMarkers := resolveReplyMarkers(opts.ReplyMarkers)

	// The error scan operates on the whole frame by default, or — when the
	// consumer opts into reply-scoping — only on the assistant-reply regions, so
	// incidental terminal SCROLLBACK (pre-session startup warnings, redis/log
	// noise) cannot false-FAIL a structural-presence bank (CONST-051(B) choice;
	// genuine in-reply errors still FAIL under either mode).
	errText := aggregated
	if opts.ErrorScopeReplies {
		errText = replyRegions(aggregated, replyMarkers)
	}
	errLower := strings.ToLower(errText)

	// CHECK 1: No error/warning tokens on screen.
	// 1a — reuse Panoptic's ErrorDetector (regex classifier) on the screen text.
	detected := v.detector.DetectErrors([]ai.ErrorMessage{{
		Message:   errText,
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
		if strings.Contains(errLower, strings.ToLower(tok)) {
			hitTokens = append(hitTokens, tok)
		}
	}
	// Provider health: flag ONLY a strictly-positive unhealthy count. The benign
	// "Unhealthy: 0" healthy state must never trip this (a bare substring match
	// would — the §11.4.107(10) analyzer self-bluff guarded here).
	if m := nonzeroUnhealthyRe.FindString(errText); m != "" {
		hitTokens = append(hitTokens, strings.TrimSpace(m))
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
	chromeRes := compileChromePatterns(opts.ChromeLinePatterns)
	for i, prompt := range opts.ExpectedPrompts {
		ok, ev := promptHasReply(aggregated, prompt, minReply, errTokens, replyMarkers, chromeRes)
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
func promptHasReply(aggregated, prompt string, minChars int, errTokens, replyMarkers []string, chromeRes []*regexp.Regexp) (bool, string) {
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
	// Anchor on the assistant turn marker (generic chat conventions / consumer
	// override) when one is present in the region: the real reply is the prose
	// AFTER that marker, not the ambient UI chrome (sidebar labels, status
	// panels, command lists) the full-frame OCR interleaves into the same
	// region. Combined with the consumer-supplied chrome patterns, this is what
	// stops an error-only turn from PASSing on surrounding static chrome (the
	// §11.4.107(10) self-bluff guarded by the golden-bad fixture). When no marker
	// is present the region is used as-is so marker-less UIs still validate.
	if region, ok := afterReplyMarker(after, replyMarkers); ok {
		after = region
	}
	// Excise error-token content + generic error words so failure chrome can
	// never count as a real reply (anti-bluff: an error IS NOT an answer).
	cleaned := stripErrorChrome(after, errTokens)
	// Drop ambient UI chrome lines (consumer-supplied patterns) so they can
	// never be miscounted as reply prose.
	cleaned = stripChromeLines(cleaned, chromeRes)
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
// genericReplyMarkers are generic chat-convention assistant-turn prefixes used
// when the consumer supplies none. They are conventions (not any consumer's UI
// strings), so they stay project-agnostic per CONST-051(B).
var genericReplyMarkers = []string{"ai:", "assistant:", "bot:", "response:", "model:"}

// resolveReplyMarkers lower-cases the consumer's markers, or returns the generic
// defaults when none are supplied.
func resolveReplyMarkers(markers []string) []string {
	if len(markers) == 0 {
		return genericReplyMarkers
	}
	out := make([]string, 0, len(markers))
	for _, m := range markers {
		if m = strings.ToLower(strings.TrimSpace(m)); m != "" {
			out = append(out, m)
		}
	}
	if len(out) == 0 {
		return genericReplyMarkers
	}
	return out
}

// compileChromePatterns compiles the CONSUMER-SUPPLIED chrome-line regexes
// (case-insensitive). Panoptic ships NONE of its own — chrome is application-
// specific (CONST-051(B)). Invalid patterns are skipped (the caller's mistake
// must not crash a validation run).
func compileChromePatterns(patterns []string) []*regexp.Regexp {
	var out []*regexp.Regexp
	for _, p := range patterns {
		if strings.TrimSpace(p) == "" {
			continue
		}
		if re, err := regexp.Compile("(?i)" + p); err == nil {
			out = append(out, re)
		}
	}
	return out
}

// replyRegions returns the concatenation of every assistant-reply region in the
// transcript — for each line that contains a reply marker, the text from the
// marker to the end of that line. When NO marker is present anywhere, it returns
// the whole text unchanged (so a marker-less recording still has its errors
// scanned rather than silently ignored — fail-closed, never fail-open).
//
// This is the error-scan analogue of afterReplyMarker: it isolates what the
// model actually emitted from ambient terminal scrollback so the reply-scoped
// error check (Options.ErrorScopeReplies) does not trip on pre-session noise.
func replyRegions(text string, markers []string) string {
	var b strings.Builder
	found := false
	for _, line := range strings.Split(text, "\n") {
		low := strings.ToLower(line)
		best := -1
		for _, m := range markers {
			if idx := strings.Index(low, m); idx >= 0 {
				end := idx + len(m)
				if best < 0 || end < best {
					best = end
				}
			}
		}
		if best >= 0 {
			found = true
			b.WriteString(line[best:])
			b.WriteByte('\n')
		}
	}
	if !found {
		// No reply marker anywhere — fail-closed: scan the whole transcript so a
		// marker-less UI's genuine errors are never silently skipped.
		return text
	}
	return b.String()
}

// afterReplyMarker returns the text following the FIRST assistant-turn marker in
// the region, and true when a marker was found. The reply is what the assistant
// said, not the chrome around it.
func afterReplyMarker(region string, markers []string) (string, bool) {
	low := strings.ToLower(region)
	best := -1
	for _, m := range markers {
		if idx := strings.Index(low, m); idx >= 0 {
			end := idx + len(m)
			if best < 0 || end < best {
				best = end
			}
		}
	}
	if best < 0 {
		return region, false
	}
	return region[best:], true
}

// stripChromeLines removes ambient UI chrome lines (matched by the consumer's
// patterns) so they cannot be miscounted as reply prose. With no patterns it is
// a no-op — Panoptic never assumes a consumer's UI layout.
func stripChromeLines(s string, chromeRes []*regexp.Regexp) string {
	if len(chromeRes) == 0 {
		return s
	}
	var keep []string
	for _, line := range strings.Split(s, "\n") {
		drop := false
		for _, re := range chromeRes {
			if re.MatchString(line) {
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
