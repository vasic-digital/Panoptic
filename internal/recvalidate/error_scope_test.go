// SPDX-FileCopyrightText: 2026 Milos Vasic
// SPDX-License-Identifier: Apache-2.0

package recvalidate

import (
	"strings"
	"testing"

	"panoptic/internal/logger"
)

// Tests for Options.ErrorScopeReplies — the reply-scoped error scan that lets a
// STRUCTURAL-presence bank ignore incidental terminal scrollback (pre-session
// startup warnings, a redis connection log) while still FAILing on genuine
// errors the model emitted inside a reply.
//
// FORENSIC ANCHOR (FACT, 2026-06-14): the HelixCode ensemble-members recording
// rendered the members panel correctly, but the terminal scrollback above it
// contained pre-session `redis: ... connection pool: failed` lines and a stray
// `"message": "Invalid input"` JSON payload. The whole-frame error scan FAILed
// the genuinely-correct recording (a §11.4.107(10) analyzer false-negative).

func newV(t *testing.T) *Validator {
	t.Helper()
	return NewValidator(*logger.NewLogger(false))
}

// scrollbackThenPanel mirrors the real recording: incidental error scrollback
// FIRST, then the assistant-reply panel with a clean structural marker.
const scrollbackThenPanel = "" +
	"redis: 2026/06/14 18:28:05 pool.go:617: redis: connection pool: failed: no such host\n" +
	"session persistence disabled.\n" +
	"\"message\": \"Invalid input: bad field\"\n" +
	"AI: ensemble: 2/14 members (strategy: confidence_weighted)\n" +
	"  [winner] Groq  score=0.92  -> llama-3.1-8b-instant (via LLMsVerifier)\n"

func TestErrorScopeReplies_IgnoresIncidentalScrollback(t *testing.T) {
	v := newV(t)
	opts := Options{
		ExpectedPrompts:   []string{"ensemble"},
		MinReplyChars:     8,
		ReplyMarkers:      []string{"AI:"},
		ErrorScopeReplies: true,
	}
	rep := v.ValidateText(scrollbackThenPanel, opts)
	noErr := findCheck(t, rep, "no_error_tokens")
	if !noErr.Pass {
		t.Fatalf("reply-scoped scan should IGNORE incidental scrollback errors, but FAILed: %s | %s",
			noErr.Detail, noErr.Evidence)
	}
}

func TestErrorScopeReplies_DefaultOff_FlagsScrollback(t *testing.T) {
	v := newV(t)
	// SAME text, default (whole-frame) scan: the incidental scrollback errors
	// MUST still be flagged — the scoping is strictly opt-in, never weakens the
	// default behaviour.
	opts := Options{
		ExpectedPrompts: []string{"ensemble"},
		MinReplyChars:   8,
		ReplyMarkers:    []string{"AI:"},
		// ErrorScopeReplies omitted => false
	}
	rep := v.ValidateText(scrollbackThenPanel, opts)
	noErr := findCheck(t, rep, "no_error_tokens")
	if noErr.Pass {
		t.Fatalf("default whole-frame scan must still FLAG the scrollback redis/Invalid-input errors, but PASSed")
	}
}

// inReplyError places a genuine error INSIDE the assistant reply — it MUST FAIL
// even under reply-scoping (the scope ignores scrollback, never real reply errors).
const inReplyError = "" +
	"redis: connection pool: failed: no such host\n" + // incidental scrollback (would be ignored)
	"AI: provider error: no provider available for this request\n"

func TestErrorScopeReplies_StillFailsOnInReplyError(t *testing.T) {
	v := newV(t)
	opts := Options{
		ExpectedPrompts:   []string{"anything"},
		MinReplyChars:     1,
		ReplyMarkers:      []string{"AI:"},
		ErrorScopeReplies: true,
	}
	rep := v.ValidateText(inReplyError, opts)
	noErr := findCheck(t, rep, "no_error_tokens")
	if noErr.Pass {
		t.Fatalf("an error INSIDE the assistant reply must FAIL even under reply-scoping, but PASSed")
	}
	if !strings.Contains(strings.ToLower(noErr.Detail), "no provider") {
		t.Logf("note: detail=%q (expected to cite the in-reply provider error)", noErr.Detail)
	}
}

// TestReplyRegions_FailsClosedWithoutMarker proves the fail-closed rule: a
// transcript with NO reply marker is scanned WHOLE (errors not silently
// skipped), so reply-scoping can never become a fail-open bluff.
func TestReplyRegions_FailsClosedWithoutMarker(t *testing.T) {
	noMarker := "redis: connection pool: failed: no such host\nsome other line\n"
	got := replyRegions(noMarker, []string{"ai:"})
	if got != noMarker {
		t.Fatalf("replyRegions must return the WHOLE text when no marker present (fail-closed); got %q", got)
	}
}

func findCheck(t *testing.T, rep *Report, name string) CheckResult {
	t.Helper()
	for _, c := range rep.Checks {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("check %q not found in report", name)
	return CheckResult{}
}
