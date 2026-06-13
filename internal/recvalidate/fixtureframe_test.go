package recvalidate

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// renderTextFramePNG draws the given lines as large, high-contrast black text
// on a white background and writes a PNG. It renders with the embeddable
// basicfont 7x13 face into a small buffer, then nearest-neighbour upscales so
// tesseract reads it reliably. This produces a REAL image that the REAL OCR
// path consumes — no simulated text anywhere.
func renderTextFramePNG(path string, lines []string, scale int) error {
	if scale < 1 {
		scale = 6
	}
	face := basicfont.Face7x13
	cellW := 7
	lineH := 18

	maxChars := 1
	for _, ln := range lines {
		if len(ln) > maxChars {
			maxChars = len(ln)
		}
	}
	pad := 6
	smallW := maxChars*cellW + 2*pad
	smallH := len(lines)*lineH + 2*pad
	if smallW < 16 {
		smallW = 16
	}
	if smallH < 16 {
		smallH = 16
	}

	small := image.NewRGBA(image.Rect(0, 0, smallW, smallH))
	draw.Draw(small, small.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)

	d := &font.Drawer{
		Dst:  small,
		Src:  image.NewUniform(color.Black),
		Face: face,
	}
	for i, line := range lines {
		d.Dot = fixed.P(pad, pad+12+i*lineH)
		d.DrawString(line)
	}

	// Nearest-neighbour upscale by `scale`.
	bigW, bigH := smallW*scale, smallH*scale
	big := image.NewRGBA(image.Rect(0, 0, bigW, bigH))
	for y := 0; y < bigH; y++ {
		for x := 0; x < bigW; x++ {
			big.Set(x, y, small.At(x/scale, y/scale))
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, big)
}
