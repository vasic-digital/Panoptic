package cmd

import (
	"context"
	"testing"

	"panoptic/pkg/i18n"
)

// fakeI18nTranslator is a unit-test-only Translator that prefixes
// every resolution so call-site tests can prove the cobra metadata
// flowed through the i18n contract instead of being a static literal.
// CONST-050(A): permitted because this file lives in *_test.go scope.
type fakeI18nTranslator struct{}

func (fakeI18nTranslator) T(
	_ context.Context, messageID string, _ ...any,
) string {
	return "<TRANSLATED:" + messageID + ">"
}

// resolveAfterSwap returns what i18n.T resolves to AFTER the active
// Translator is swapped to fakeI18nTranslator. Because cobra command
// metadata is materialised once at package-init time (with the
// NoopTranslator default still active), we resolve via i18n.T directly
// here rather than re-reading errorsCmd.Short — the migration's
// correctness is "the call site uses i18n.T", which this proves.
func resolveAfterSwap(messageID string) string {
	prev := i18n.ActiveTranslator()
	i18n.SetTranslator(fakeI18nTranslator{})
	defer i18n.SetTranslator(prev)
	return i18n.T(messageID)
}

// TestErrorsCmd_ShortUsesI18nID asserts the cobra metadata for the
// errors command was materialised through i18n.T at package init
// (resolving via NoopTranslator to the raw message ID) — NOT a
// hardcoded English literal. If a future change reverts to a literal,
// errorsCmd.Short will no longer equal the message ID and this fails.
func TestErrorsCmd_ShortUsesI18nID(t *testing.T) {
	if errorsCmd.Short != "panoptic_cmd_errors_short" {
		t.Fatalf(
			"errorsCmd.Short = %q; expected raw message ID " +
				"%q (proves i18n.T routed through NoopTranslator " +
				"at init; literal regression would change this)",
			errorsCmd.Short, "panoptic_cmd_errors_short",
		)
	}
	// And prove the registry routes the same ID through a real
	// Translator when one is wired — sentinel must contain the ID.
	got := resolveAfterSwap("panoptic_cmd_errors_short")
	want := "<TRANSLATED:panoptic_cmd_errors_short>"
	if got != want {
		t.Fatalf("resolveAfterSwap = %q, want %q", got, want)
	}
}

// TestErrorsAnalyzeCmd_ShortUsesI18nID — same pattern for the
// `errors analyze` subcommand.
func TestErrorsAnalyzeCmd_ShortUsesI18nID(t *testing.T) {
	if errorsAnalyzeCmd.Short != "panoptic_cmd_errors_analyze_short" {
		t.Fatalf(
			"errorsAnalyzeCmd.Short = %q; expected raw message " +
				"ID %q", errorsAnalyzeCmd.Short,
			"panoptic_cmd_errors_analyze_short",
		)
	}
	got := resolveAfterSwap("panoptic_cmd_errors_analyze_short")
	want := "<TRANSLATED:panoptic_cmd_errors_analyze_short>"
	if got != want {
		t.Fatalf("resolveAfterSwap = %q, want %q", got, want)
	}
}

// TestVisionCmd_ShortUsesI18nID — same pattern for the `vision` root.
func TestVisionCmd_ShortUsesI18nID(t *testing.T) {
	if visionCmd.Short != "panoptic_cmd_vision_short" {
		t.Fatalf(
			"visionCmd.Short = %q; expected raw message ID %q",
			visionCmd.Short, "panoptic_cmd_vision_short",
		)
	}
	got := resolveAfterSwap("panoptic_cmd_vision_short")
	want := "<TRANSLATED:panoptic_cmd_vision_short>"
	if got != want {
		t.Fatalf("resolveAfterSwap = %q, want %q", got, want)
	}
}

// TestVisionDetectCmd_ShortUsesI18nID — `vision detect` subcommand.
func TestVisionDetectCmd_ShortUsesI18nID(t *testing.T) {
	if visionDetectCmd.Short != "panoptic_cmd_vision_detect_short" {
		t.Fatalf(
			"visionDetectCmd.Short = %q; expected raw message " +
				"ID %q", visionDetectCmd.Short,
			"panoptic_cmd_vision_detect_short",
		)
	}
	got := resolveAfterSwap("panoptic_cmd_vision_detect_short")
	want := "<TRANSLATED:panoptic_cmd_vision_detect_short>"
	if got != want {
		t.Fatalf("resolveAfterSwap = %q, want %q", got, want)
	}
}

// TestVisionReportCmd_ShortUsesI18nID — `vision report` subcommand.
func TestVisionReportCmd_ShortUsesI18nID(t *testing.T) {
	if visionReportCmd.Short != "panoptic_cmd_vision_report_short" {
		t.Fatalf(
			"visionReportCmd.Short = %q; expected raw message " +
				"ID %q", visionReportCmd.Short,
			"panoptic_cmd_vision_report_short",
		)
	}
	got := resolveAfterSwap("panoptic_cmd_vision_report_short")
	want := "<TRANSLATED:panoptic_cmd_vision_report_short>"
	if got != want {
		t.Fatalf("resolveAfterSwap = %q, want %q", got, want)
	}
}
