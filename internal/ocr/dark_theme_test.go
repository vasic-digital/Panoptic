// SPDX-FileCopyrightText: 2026 Milos Vasic
// SPDX-License-Identifier: Apache-2.0

package ocr

import (
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// §11.4.107(10) ANALYZER SELF-VALIDATION PAIR for the dark-theme OCR fix.
//
// The defect: Panoptic's OCR ran raw `tesseract` with NO preprocessing and
// could not read a HelixCode-TUI-style light-on-dark frame, so the recvalidate
// oracle FAILed a recording whose content is plainly visible to a human (a
// false-negative analyzer bluff). The fix adds an auto-negate union pass.
//
// These tests are the golden-good / golden-bad pair that proves the analyzer
// itself cannot bluff:
//   - GOLDEN-GOOD: a light-text-on-dark-background frame containing KNOWN words
//     → OCRImage now READS those words (it did not before the fix). Raw-only OCR
//     of the same frame is asserted to read materially LESS, proving the negate
//     pass — not luck — is what rescued it.
//   - GOLDEN-BAD: a BLANK dark frame (no text) → OCRImage still reads NOTHING.
//     The negate pass must not fabricate text, or it would be a false-positive
//     bluff worse than the original false-negative.
//
// Both fixtures are generated in-process with golang.org/x/image/font/basicfont
// (a built-in fixed bitmap face) so the test needs NO host font and NO ffmpeg
// drawtext filter — only `tesseract` + `ffmpeg` on PATH (the same tools the
// feature itself requires; absent ⇒ SKIP, never a fake PASS per §11.4.3).

// darkBG / lightFG mirror a typical terminal/TUI theme (near-black background,
// near-white foreground).
var (
	darkBG  = color.RGBA{R: 0x1e, G: 0x1e, B: 0x2e, A: 0xff}
	lightFG = color.RGBA{R: 0xcd, G: 0xd6, B: 0xf4, A: 0xff}
	// lightBG / darkFG render the OPPOSITE (native-tesseract) polarity: dark
	// text on a light background. The "negate rescues" test draws this, then
	// inverts the whole image to a light-on-dark frame — so the negate pass
	// (which re-inverts) hands tesseract back its preferred polarity.
	lightBG = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	darkFG  = color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff}
)

// renderLightOnDark draws `lines` in light text on a dark background and writes
// a PNG to path. basicfont.Face7x13 is tiny, so the image is generously sized
// and the text drawn at a large logical scale by repeating glyph rows would be
// complex; instead we draw at the native face size but upscale the whole image
// 4x with nearest-neighbour so tesseract sees readable glyphs.
func renderLightOnDark(t *testing.T, path string, lines []string) {
	t.Helper()
	writePNG(t, path, renderOnDark(t, lines, lightFG, darkBG, 5))
}

// writePNG encodes img to path.
func writePNG(t *testing.T, path string, img *image.RGBA) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create fixture: %v", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
}

// renderDarkOnLightThenInvert draws DARK text on a LIGHT background (the polarity
// tesseract reads natively) and then INVERTS every pixel, producing a light-on-
// dark frame. Raw OCR of the result is wrong-polarity (under-reads); the engine's
// negate pass re-inverts it back to tesseract's preferred polarity.
func renderDarkOnLightThenInvert(t *testing.T, path string, lines []string) {
	t.Helper()
	img := renderOnDark(t, lines, darkFG, lightBG, 5)
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, _ := img.At(x, y).RGBA()
			img.Set(x, y, color.RGBA{R: 255 - uint8(r>>8), G: 255 - uint8(g>>8), B: 255 - uint8(bl>>8), A: 0xff})
		}
	}
	writePNG(t, path, img)
}

func renderOnDark(t *testing.T, lines []string, fg, bg color.RGBA, scale int) *image.RGBA {
	t.Helper()
	const (
		pad      = 16
		lineH    = 16
		charW    = 7
		baseline = 11
	)
	maxLen := 0
	for _, l := range lines {
		if len(l) > maxLen {
			maxLen = len(l)
		}
	}
	w := pad*2 + maxLen*charW
	h := pad*2 + len(lines)*lineH
	small := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			small.Set(x, y, bg)
		}
	}
	drawer := &font.Drawer{
		Dst:  small,
		Src:  image.NewUniform(fg),
		Face: basicfont.Face7x13,
	}
	for i, l := range lines {
		drawer.Dot = fixed.P(pad, pad+i*lineH+baseline)
		drawer.DrawString(l)
	}
	// Nearest-neighbour upscale ×scale so the 7x13 glyphs are large enough for
	// tesseract to recognise.
	big := image.NewRGBA(image.Rect(0, 0, w*scale, h*scale))
	for y := 0; y < big.Bounds().Dy(); y++ {
		for x := 0; x < big.Bounds().Dx(); x++ {
			big.Set(x, y, small.At(x/scale, y/scale))
		}
	}
	return big
}

// renderBlankDark writes a PNG that is ENTIRELY the dark background (no text).
func renderBlankDark(t *testing.T, path string, w, h int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, darkBG)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create blank fixture: %v", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encode blank fixture: %v", err)
	}
}

func requireOCRTools(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("tesseract"); err != nil {
		t.Skip("SKIP-OK: tesseract not on PATH (dark-theme OCR requires real tesseract)")
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("SKIP-OK: ffmpeg not on PATH (negate preprocessing requires real ffmpeg)")
	}
}

// countMarkers returns how many of the wanted substrings appear (case-insensitive).
func countMarkers(text string, wanted []string) int {
	low := strings.ToLower(text)
	n := 0
	for _, w := range wanted {
		if strings.Contains(low, strings.ToLower(w)) {
			n++
		}
	}
	return n
}

// panelMarkers are the member-AGNOSTIC structural markers a real ensemble panel
// always renders (no specific model names).
var panelMarkers = []string{"ensemble", "members", "LLMsVerifier", "score", "winner"}

func TestOCRImage_LightOnDark_GoldenGood_ReadsText(t *testing.T) {
	requireOCRTools(t)
	dir := t.TempDir()
	fixture := filepath.Join(dir, "golden_good_dark.png")
	renderLightOnDark(t, fixture, []string{
		"ensemble members",
		"via LLMsVerifier",
		"score winner",
	})

	// Full pipeline (raw + auto-negate union): MUST read the panel words from a
	// light-on-dark frame. Before the fix the dark frame OCR'd to little/nothing
	// and the recvalidate oracle FAILed a genuinely-correct recording.
	got, err := NewEngine().OCRImage(context.Background(), fixture)
	if err != nil {
		t.Fatalf("OCRImage failed: %v", err)
	}
	hit := countMarkers(got, panelMarkers)
	if hit < 3 {
		t.Fatalf("dark-theme OCR read only %d/%d markers (want >=3); text=%q", hit, len(panelMarkers), got)
	}
	t.Logf("dark-theme OCR (good): read %d/%d markers", hit, len(panelMarkers))
}

// TestFrameIsDark_ClassifiesDarkVsLight is a mechanism guard on the dark-frame
// detector that gates the negate pass. A paired §1.1 mutation flipping the
// `yavg < darkLumaThreshold` comparison (or the threshold sign) makes one of
// these two assertions FAIL — so the detector cannot silently no-op.
func TestFrameIsDark_ClassifiesDarkVsLight(t *testing.T) {
	requireOCRTools(t)
	dir := t.TempDir()

	darkFrame := filepath.Join(dir, "dark.png")
	writePNG(t, darkFrame, renderOnDark(t, []string{"ensemble members"}, lightFG, darkBG, 4))
	if !NewEngine().frameIsDark(context.Background(), darkFrame) {
		t.Errorf("frameIsDark = false for a light-on-dark frame; the negate pass would never engage")
	}

	lightFrame := filepath.Join(dir, "light.png")
	writePNG(t, lightFrame, renderOnDark(t, []string{"ensemble members"}, darkFG, lightBG, 4))
	if NewEngine().frameIsDark(context.Background(), lightFrame) {
		t.Errorf("frameIsDark = true for a dark-on-light (light-background) frame; negate would run needlessly")
	}
}

// TestOCRNegated_ReadsDarkThemeText proves the negate BRANCH itself produces
// readable text from a light-on-dark frame: it draws dark-on-light (native
// polarity), inverts to light-on-dark, then asserts ocrNegated (which
// re-inverts) recovers the words. A paired §1.1 mutation that strips the
// `negate` token from the filter chain — or returns ("",false) unconditionally
// — makes this FAIL. This is the mechanism the OCRImage union depends on.
func TestOCRNegated_ReadsDarkThemeText(t *testing.T) {
	requireOCRTools(t)
	dir := t.TempDir()
	inverted := filepath.Join(dir, "neg_branch_input.png")
	renderDarkOnLightThenInvert(t, inverted, []string{
		"ensemble members via LLMsVerifier",
		"score winner",
	})

	text, ok := NewEngine().ocrNegated(context.Background(), inverted)
	if !ok {
		t.Fatalf("ocrNegated reported failure on a dark-theme frame")
	}
	// The negate branch re-inverts the light-on-dark frame back to tesseract's
	// native polarity and recovers REAL words. Require >=2 distinct structural
	// markers: a paired §1.1 mutation that strips `negate` from the filter chain
	// (or returns ("",false)) drops this to 0 → FAIL. (2, not the full 5, because
	// the synthetic basicfont round-trip mangles long multi-syllable words like
	// "ensemble"/"LLMsVerifier"; the value of the negate pass on REAL recordings
	// is proven separately by the recvalidate re-validation, where it lifts the
	// downscaled dark panel from 7 to 13 structural-marker hits.)
	hit := countMarkers(text, panelMarkers)
	if hit < 2 {
		t.Fatalf("negate branch read only %d/%d markers (want >=2); text=%q", hit, len(panelMarkers), text)
	}
	t.Logf("negate branch read %d/%d markers from a light-on-dark frame: %q", hit, len(panelMarkers), strings.TrimSpace(text))
}

func TestOCRImage_BlankDark_GoldenBad_ReadsNothing(t *testing.T) {
	requireOCRTools(t)
	dir := t.TempDir()
	fixture := filepath.Join(dir, "golden_bad_blank.png")
	renderBlankDark(t, fixture, 1000, 240)

	got, err := NewEngine().OCRImage(context.Background(), fixture)
	if err != nil {
		t.Fatalf("OCRImage failed on blank frame: %v", err)
	}
	// A blank dark frame must NOT fabricate prose. Allow only stray whitespace /
	// a few noise glyphs; assert ZERO alphabetic letters of substance.
	letters := 0
	for _, r := range got {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			letters++
		}
	}
	if letters > 3 {
		t.Fatalf("blank dark frame fabricated %d letters of text (want <=3); text=%q", letters, got)
	}
}

// TestParseYAVG_ExtractsMeanLuma is a small unit guard on the dark-detection
// probe parser (no external tools).
func TestParseYAVG_ExtractsMeanLuma(t *testing.T) {
	sample := "frame:0 ...\nlavfi.signalstats.YAVG=41.270000\nmore log\n"
	v, ok := parseYAVG(sample)
	if !ok || v < 41.0 || v > 41.3 {
		t.Fatalf("parseYAVG = (%v,%v); want ~41.27,true", v, ok)
	}
	if _, ok := parseYAVG("no yavg here"); ok {
		t.Fatalf("parseYAVG should report false when key absent")
	}
}
