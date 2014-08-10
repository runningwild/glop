package main

import (
	"code.google.com/p/freetype-go/freetype"
	"code.google.com/p/freetype-go/freetype/raster"
	"code.google.com/p/freetype-go/freetype/truetype"
	"fmt"
	"github.com/disintegration/gift"
	"image"
	"image/color"
	"math"
)

// boundingBoxer is a drawable image that reports a massive canvas.  Ideally this means that
// anything can be draw onto it.  After done drawing, call boundingBoxer.complete(), then it will
// report bounds that are just large enough to contain everything that was draw onto it.
type boundingBoxer struct {
	boundsStarted bool

	// maxDist is the distance, above which, we don't bother to distinguish between distances.
	// this is so that we can spend more bits of precision right around an edge.
	maxDist float64

	bounds      image.Rectangle
	largeBounds image.Rectangle
	crossover   map[image.Point]bool
	inside      map[image.Point]bool

	// The grayscale image data.
	grayField []uint16
}

func makeBoundingBoxer() *boundingBoxer {
	return &boundingBoxer{
		largeBounds: image.Rect(-10000, -10000, 10000, 10000),
		crossover:   make(map[image.Point]bool),
		inside:      make(map[image.Point]bool),
	}
}

func (bb *boundingBoxer) At(x, y int) color.Color {
	if bb.maxDist == 0 {
		return color.Gray{0}
	}
	if !(image.Point{x, y}).In(bb.bounds) {
		return color.Gray{0}
	}
	return color.Gray16{bb.grayField[(x-bb.bounds.Min.X)+(y-bb.bounds.Min.Y)*bb.bounds.Dx()]}
}
func (bb *boundingBoxer) ColorModel() color.Model {
	return color.Gray16Model
}
func (bb *boundingBoxer) Bounds() image.Rectangle {
	if bb.maxDist == 0 {
		return bb.largeBounds
	}
	return bb.bounds
}

type offsetAndDist struct {
	x, y int
	dist float64
}

func (bb *boundingBoxer) complete() {
	for insidePoint := range bb.inside {
		delete(bb.crossover, insidePoint)
	}

	// bb.maxDist = math.Sqrt(float64(bb.bounds.Dx()*bb.bounds.Dy())) / bb.something / 10
	bb.maxDist = 50

	bb.bounds.Min.X -= int(bb.maxDist) + 1
	bb.bounds.Min.Y -= int(bb.maxDist) + 1
	bb.bounds.Max.X += int(bb.maxDist) + 1
	bb.bounds.Max.Y += int(bb.maxDist) + 1

	distField := make([]float64, bb.bounds.Dx()*bb.bounds.Dy())
	for i := range distField {
		distField[i] = bb.maxDist
	}
	max := int(bb.maxDist)

	// offsests will contain the offsets from a crossover point that we should test for, and dists
	// will contain the distance of that offset.  This way we don't need to recalculate all of this
	// stuff for every crossover pixel we look at.
	var offsets []offsetAndDist
	for dy := -max; dy <= max; dy++ {
		for dx := -max; dx <= max; dx++ {
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist >= bb.maxDist {
				continue
			}
			offsets = append(offsets, offsetAndDist{dx, dy, dist})
		}
	}

	bbdx := bb.bounds.Dx()
	bbdy := bb.bounds.Dy()
	for point := range bb.crossover {
		for _, offset := range offsets {
			x := point.X + offset.x - bb.bounds.Min.X
			y := point.Y + offset.y - bb.bounds.Min.Y
			if x < 0 || x >= bbdx || y < 0 || y >= bbdy {
				continue
			}
			pos := x + y*bbdx
			if offset.dist < distField[pos] {
				distField[pos] = offset.dist
			}
		}
	}
	bb.grayField = make([]uint16, bb.bounds.Dx()*bb.bounds.Dy())
	for i := range distField {
		x := (i % bb.bounds.Dx()) + bb.bounds.Min.X
		y := (i / bb.bounds.Dx()) + bb.bounds.Min.Y
		bb.grayField[i] = 32767 - uint16(32767*(distField[i]/bb.maxDist))
		if bb.inside[image.Point{x, y}] {
			bb.grayField[i] = 65535 - bb.grayField[i]
		}
	}
}
func (bb *boundingBoxer) Set(x, y int, c color.Color) {
	bb.inside[image.Point{x, y}] = true
	bb.crossover[image.Point{x - 1, y}] = true
	bb.crossover[image.Point{x + 1, y}] = true
	bb.crossover[image.Point{x, y - 1}] = true
	bb.crossover[image.Point{x, y + 1}] = true

	// Update all bounds so that we can eventually report a bounding box for everything.
	if !bb.boundsStarted {
		bb.boundsStarted = true
		bb.bounds.Min = image.Point{x, y}
		bb.bounds.Max = image.Point{x, y}
	}
	if x < bb.bounds.Min.X {
		bb.bounds.Min.X = x
	}
	if y < bb.bounds.Min.Y {
		bb.bounds.Min.Y = y
	}
	if x > bb.bounds.Max.X {
		bb.bounds.Max.X = x
	}
	if y > bb.bounds.Max.Y {
		bb.bounds.Max.Y = y
	}
}

// Render draws rune r front the specified font at the specified dpi and scale.  It returns a
// grayscale image that is just large enough to contain the rune.
func Render(font *truetype.Font, r rune, dpi, scale float64) (*image.Gray, error) {
	glyph := truetype.NewGlyphBuf()
	index := font.Index(r)
	glyph.Load(font, font.FUnitsPerEm(), index, truetype.FullHinting)
	ctx := freetype.NewContext()
	boxer := makeBoundingBoxer()
	ctx.SetSrc(image.NewUniform(color.White))
	ctx.SetDst(boxer)
	ctx.SetClip(boxer.largeBounds)
	ctx.SetFontSize(250)
	ctx.SetDPI(dpi)
	ctx.SetFont(font)
	if err := glyph.Load(font, font.FUnitsPerEm(), font.Index(r), truetype.FullHinting); err != nil {
		return nil, fmt.Errorf("Unable to load glyph: %v\n", err)
	}
	var rp raster.Point
	rp.X = ctx.PointToFix32(0)
	rp.Y = ctx.PointToFix32(100)
	ctx.DrawString(string(r), rp)
	boxer.complete()

	g := gift.New(
		gift.Resize(int(float64(boxer.Bounds().Dx())*scale+0.5), int(float64(boxer.Bounds().Dy())*scale+0.5), gift.CubicResampling),
	)
	dst := image.NewGray(g.Bounds(boxer.Bounds()))
	g.Draw(dst, boxer)
	return dst, nil
}
