// Package-level translator registry for panoptic. Cobra command
// metadata (Short:, Long:) is materialised at package-init time when
// there is no context available, so call sites resolve through this
// package-global accessor. Host programs swap the active Translator
// via SetTranslator(...) BEFORE invoking cobra.Command.Execute(). The
// default is NoopTranslator{} so the submodule remains standalone.
//
// Concurrency: SetTranslator is intended for one-shot wiring at
// program start. A sync.RWMutex guards swaps to keep T() safe for
// any later runtime change without imposing init-order constraints
// on consumers.
package i18n

import (
	"context"
	"sync"
)

var (
	mu     sync.RWMutex
	active Translator = NoopTranslator{}
)

// SetTranslator installs t as the active Translator. Passing nil
// resets the registry to NoopTranslator so callers always observe a
// usable translator.
func SetTranslator(t Translator) {
	mu.Lock()
	defer mu.Unlock()
	if t == nil {
		active = NoopTranslator{}
		return
	}
	active = t
}

// Translator returns the currently registered Translator. Safe for
// concurrent use; result captured by value (interface), so a later
// SetTranslator call does not mutate the returned reference.
func ActiveTranslator() Translator {
	mu.RLock()
	defer mu.RUnlock()
	return active
}

// T is a convenience helper resolving messageID through the active
// Translator with a background context. Cobra metadata sites use this
// because they have no request-scoped context.
func T(messageID string, args ...any) string {
	return ActiveTranslator().T(context.Background(), messageID, args...)
}
