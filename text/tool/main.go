package main

import (
	"code.google.com/p/freetype-go/freetype"
	"code.google.com/p/freetype-go/freetype/raster"
	"code.google.com/p/freetype-go/freetype/truetype"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"math"
	"os"
)

var path = flag.String("path", "", "Font file to use")
var runes = flag.String("runes", "abc", "Runes to process")

type boundingBoxer struct {
	boundsStarted bool

	// maxDist is the distance, above which, we don't bother to distinguish between distances.
	// this is so that we can spend more bits of precision right around an edge.
	maxDist float64

	bounds      image.Rectangle
	largeBounds image.Rectangle
	crossover   map[image.Point]bool
	inside      map[image.Point]bool

	distField []float64
	grayField []byte
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
	return color.Gray{bb.grayField[(x-bb.bounds.Min.X)+(y-bb.bounds.Min.Y)*bb.bounds.Dx()]}
}
func (bb *boundingBoxer) ColorModel() color.Model {
	return color.GrayModel
}
func (bb *boundingBoxer) Bounds() image.Rectangle {
	if bb.maxDist == 0 {
		return bb.largeBounds
	}
	return bb.bounds
}
func (bb *boundingBoxer) complete() {
	for insidePoint := range bb.inside {
		delete(bb.crossover, insidePoint)
	}

	bb.maxDist = math.Sqrt(float64(bb.bounds.Dx()*bb.bounds.Dy())) / 50
	fmt.Printf("max: %v\n", bb.maxDist)

	bb.bounds.Min.X -= int(bb.maxDist) + 1
	bb.bounds.Min.Y -= int(bb.maxDist) + 1
	bb.bounds.Max.X += int(bb.maxDist) + 1
	bb.bounds.Max.Y += int(bb.maxDist) + 1

	bb.distField = make([]float64, bb.bounds.Dx()*bb.bounds.Dy())
	for i := range bb.distField {
		bb.distField[i] = bb.maxDist
	}
	max := int(bb.maxDist + 1)
	for point := range bb.crossover {
		for dy := -max; dy <= max; dy++ {
			for dx := -max; dx <= max; dx++ {
				x := point.X + dx - bb.bounds.Min.X
				y := point.Y + dy - bb.bounds.Min.Y
				if x < 0 || x >= bb.bounds.Dx() || y < 0 || y >= bb.bounds.Dy() {
					continue
				}
				pos := x + y*bb.bounds.Dx()
				dist := math.Sqrt(float64(dx*dx + dy*dy))
				if dist < bb.distField[pos] {
					bb.distField[pos] = dist
				}
			}
		}
	}
	fmt.Printf("Bounds: %v\n", bb.bounds)
	bb.grayField = make([]byte, bb.bounds.Dx()*bb.bounds.Dy())
	for i := range bb.distField {
		x := (i % bb.bounds.Dx()) + bb.bounds.Min.X
		y := (i / bb.bounds.Dx()) + bb.bounds.Min.Y
		bb.grayField[i] = 127 - byte(127*(bb.distField[i]/bb.maxDist))
		if bb.inside[image.Point{x, y}] {
			bb.grayField[i] = 255 - bb.grayField[i]
		}
	}
}
func (bb *boundingBoxer) Set(x, y int, c color.Color) {
	bb.inside[image.Point{x, y}] = true
	bb.crossover[image.Point{x - 1, y}] = true
	bb.crossover[image.Point{x + 1, y}] = true
	bb.crossover[image.Point{x, y - 1}] = true
	bb.crossover[image.Point{x, y + 1}] = true
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

func main() {
	flag.Parse()
	data, err := ioutil.ReadFile(*path)
	if err != nil {
		fmt.Printf("Unable to read file: %v", err)
		os.Exit(1)
	}
	f, err := freetype.ParseFont(data)
	if err != nil {
		fmt.Printf("Unable to parse font file: %v", err)
		os.Exit(1)
	}
	glyph := truetype.NewGlyphBuf()
	fmt.Printf("Funits: %v\n", f.FUnitsPerEm())
	for _, r := range *runes {
		index := f.Index(r)
		glyph.Load(f, f.FUnitsPerEm(), index, truetype.FullHinting)
		fmt.Printf("Rune: %c\n", r)
		fmt.Printf("Advance: %v\n", glyph.AdvanceWidth)
		fmt.Printf("Bounds: %v\n", glyph.B)
		fmt.Printf("VMetric: %v\n", f.VMetric(f.FUnitsPerEm(), index))
		fmt.Printf("HMetric: %v\n", f.HMetric(f.FUnitsPerEm(), index))
		for _, p := range glyph.Point {
			if p.Flags&1 == 1 {
				fmt.Printf("%v %v\n", p.X, p.Y)
			}
		}
	}
	ctx := freetype.NewContext()
	dst := makeBoundingBoxer()
	ctx.SetSrc(image.NewUniform(color.White))
	ctx.SetDst(dst)
	ctx.SetClip(dst.largeBounds)
	ctx.SetFontSize(250)
	ctx.SetDPI(2000)
	ctx.SetFont(f)
	if err := glyph.Load(f, f.FUnitsPerEm(), f.Index('X'), truetype.FullHinting); err != nil {
		fmt.Printf("Unable to load glyph: %v\n", err)
		os.Exit(1)
	}
	var rp raster.Point
	rp.X = ctx.PointToFix32(0)
	rp.Y = ctx.PointToFix32(100)
	ctx.DrawString(*runes, rp)
	fmt.Printf("%v\n", len(dst.crossover))
	fmt.Printf("%v\n", dst.Bounds())
	dst.complete()
	fmt.Printf("%v\n", dst.Bounds())

	out, err := os.Create("output.png")
	if err != nil {
		fmt.Printf("Couldn't create output file: %v\n")
		os.Exit(1)
	}
	err = png.Encode(out, dst)
	if err != nil {
		fmt.Printf("Couldn't encode output image: %v\n")
		os.Exit(1)
	}
}
