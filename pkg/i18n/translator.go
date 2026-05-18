// Package i18n provides a project-decoupled translator interface for
// panoptic. Per CONST-046 (no hardcoded user-facing content) and
// CONST-051(B) (submodule decoupling), this package MUST NOT import
// from any consuming project's namespace (e.g. helix_code/...). It
// supplies a tiny contract — Translator — that callers depend on, plus
// a NoopTranslator default implementation that preserves the message
// ID verbatim so panoptic remains usable as a standalone module without
// requiring any host-side wiring.
//
// Reference: round-90 design doc
// docs/superpowers/specs/2026-05-19-const046-i18n-architecture-design.md
// Phase 3 round 99a.
package i18n

import "context"

// Translator is the contract panoptic call sites depend on for any
// user-facing text. Implementations are supplied by the host program
// (typically wired via a higher-level adapter that bridges to a real
// i18n bundle store). The contract is intentionally minimal so the
// submodule stays project-not-aware (CONST-051(B)).
type Translator interface {
	// T resolves messageID against the active locale carried in ctx
	// (or the implementation's default), substituting args by name.
	// Implementations MUST return a non-empty string even on lookup
	// miss — callers depend on T() never returning the empty string.
	T(ctx context.Context, messageID string, args ...any) string
}

// NoopTranslator is the safe default Translator. It returns the
// messageID verbatim, ignoring args. This guarantees panoptic remains
// usable when no host translator is wired (e.g. CLI invocation, unit
// tests, standalone embedding). The returned string is the message ID
// itself, which downstream rendering can recognise as un-translated
// and surface to operators for missing-bundle diagnosis.
type NoopTranslator struct{}

// T implements Translator. Returns messageID unchanged when non-empty;
// returns a constant sentinel for the empty-ID misuse case so callers
// never observe an empty user-facing string.
func (NoopTranslator) T(_ context.Context, messageID string, _ ...any) string {
	if messageID == "" {
		return "i18n.empty_message_id"
	}
	return messageID
}
