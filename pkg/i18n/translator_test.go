package i18n

import (
	"context"
	"testing"
)

// TestNoopTranslator_T_ReturnsMessageIDVerbatim asserts that the
// default NoopTranslator returns the requested message ID unchanged,
// preserving standalone behaviour when no host bundle is wired.
func TestNoopTranslator_T_ReturnsMessageIDVerbatim(t *testing.T) {
	tr := NoopTranslator{}
	got := tr.T(context.Background(), "panoptic_cmd_errors_short")
	if got != "panoptic_cmd_errors_short" {
		t.Fatalf(
			"NoopTranslator.T returned %q, want %q",
			got, "panoptic_cmd_errors_short",
		)
	}
}

// TestNoopTranslator_T_EmptyIDFallback asserts the empty-ID sentinel
// — callers must never observe an empty user-facing string even on
// misuse, so the contract is enforced by the default implementation.
func TestNoopTranslator_T_EmptyIDFallback(t *testing.T) {
	tr := NoopTranslator{}
	got := tr.T(context.Background(), "")
	if got == "" {
		t.Fatalf(
			"NoopTranslator.T on empty messageID returned empty " +
				"string; contract says non-empty sentinel required",
		)
	}
	if got != "i18n.empty_message_id" {
		t.Fatalf(
			"NoopTranslator.T on empty messageID returned %q, " +
				"want %q", got, "i18n.empty_message_id",
		)
	}
}

// fakeTranslator returns a sentinel-prefixed string for every
// messageID so call-site tests can assert that resolution went
// through the registered Translator (not a static literal).
type fakeTranslator struct{}

func (fakeTranslator) T(_ context.Context, messageID string, _ ...any) string {
	return "<TRANSLATED:" + messageID + ">"
}

// TestSetTranslator_SwapAndReset exercises the package-global
// registry: install a fake, observe routing, reset with nil, observe
// fallback to NoopTranslator.
func TestSetTranslator_SwapAndReset(t *testing.T) {
	defer SetTranslator(nil) // ensure global state restored

	SetTranslator(fakeTranslator{})
	got := T("panoptic_cmd_vision_short")
	want := "<TRANSLATED:panoptic_cmd_vision_short>"
	if got != want {
		t.Fatalf(
			"after SetTranslator(fake), T returned %q, want %q",
			got, want,
		)
	}

	SetTranslator(nil)
	got = T("panoptic_cmd_vision_short")
	if got != "panoptic_cmd_vision_short" {
		t.Fatalf(
			"after SetTranslator(nil), T returned %q, want raw " +
				"message ID %q (NoopTranslator fallback)",
			got, "panoptic_cmd_vision_short",
		)
	}
}
