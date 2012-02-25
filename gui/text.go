package gui

import (
  "image"
  "image/color"
  "image/draw"
  "image/png"
  "os"
  "sort"
  "code.google.com/p/freetype-go/freetype"
  "code.google.com/p/freetype-go/freetype/raster"
  "code.google.com/p/freetype-go/freetype/truetype"
  "github.com/runningwild/glop/render"
  "github.com/runningwild/opengl/gl"
  "github.com/runningwild/opengl/glu"
)

type subImage struct {
  im     image.Image
  bounds image.Rectangle
}
type transparent struct {}
func (t transparent) RGBA() (r, g, b, a uint32) {
  return 0, 0, 0, 0
}
func (si *subImage) ColorModel() color.Model {
  return si.im.ColorModel()
}
func (si *subImage) Bounds() image.Rectangle {
  return si.bounds
}
func (si *subImage) At(x, y int) color.Color {
  b := si.bounds
  if (image.Point{x, y}).In(b) {
    return si.im.At(x, y)
  }
  return transparent{}
}

// Returns a sub-image of the input image.  The bounding rectangle is the
// smallest possible rectangle that includes all pixels that have alpha > 0,
// with one pixel of border on all sides.
func minimalSubImage(src image.Image) *subImage {
  bounds := src.Bounds()
  var new_bounds image.Rectangle
  new_bounds.Max = bounds.Min
  new_bounds.Min = bounds.Max
  for x := bounds.Min.X; x <= bounds.Max.X; x++ {
    for y := bounds.Min.Y; y <= bounds.Max.Y; y++ {
      c := src.At(x, y)
      _,_,_,a := c.RGBA()
      if a > 0 {
        if x < new_bounds.Min.X {
          new_bounds.Min.X = x
        }
        if y < new_bounds.Min.Y {
          new_bounds.Min.Y = y
        }
        if x > new_bounds.Max.X {
          new_bounds.Max.X = x
        }
        if y > new_bounds.Max.Y {
          new_bounds.Max.Y = y
        }
      }
    }
  }

  // // We want one row/col of boundary between characters so that we don't get
  // // annoying artifacts
  new_bounds.Min.X--
  new_bounds.Min.Y--
  new_bounds.Max.X++
  new_bounds.Max.Y++

  if new_bounds.Min.X > new_bounds.Max.X || new_bounds.Min.Y > new_bounds.Max.Y {
    new_bounds = image.Rect(0, 0, 0, 0)
  }

  return &subImage{ src, new_bounds }
}

// This stupid thing is just so that our idiot-packedImage can answer queries
// faster.  If we're going to query every pixel then it makes sense to check
// the largest rectangles first, since they will be the correct response more
// often than the smaller rectangles.
type packedImageSortByArea struct {
  *packedImage
}
func (p *packedImageSortByArea) Len() int {
  return len(p.ims)
}
func (p *packedImageSortByArea) Less(i, j int) bool {
  ai := p.ims[i].Bounds().Dx() * p.ims[i].Bounds().Dy()
  aj := p.ims[j].Bounds().Dx() * p.ims[j].Bounds().Dy()
  return ai > aj
}
func (p *packedImageSortByArea) Swap(i, j int) {
  p.ims[i], p.ims[j] = p.ims[j], p.ims[i]
  p.off[i], p.off[j] = p.off[j], p.off[i]
}

type packedImage struct {
  ims []image.Image
  off []image.Point
  bounds image.Rectangle
}
func (p *packedImage) Len() int {
  return len(p.ims)
}
func (p *packedImage) Less(i, j int) bool {
  return p.ims[i].Bounds().Dy() < p.ims[j].Bounds().Dy()
}
func (p *packedImage) Swap(i, j int) {
  p.ims[i], p.ims[j] = p.ims[j], p.ims[i]
  p.off[i], p.off[j] = p.off[j], p.off[i]
}
func (p *packedImage) GetRect(im image.Image) image.Rectangle {
  for i := range p.ims {
    if im == p.ims[i] {
      return p.ims[i].Bounds().Add(p.off[i])
    }
  }
  return image.Rectangle{}
}
func (p *packedImage) ColorModel() color.Model {
  return p.ims[0].ColorModel()
}
func (p *packedImage) Bounds() image.Rectangle {
  return p.bounds
}
func (p *packedImage) At(x, y int) color.Color {
  point := image.Point{x, y}
  for i := range p.ims {
    if point.In(p.ims[i].Bounds().Add(p.off[i])) {
      return p.ims[i].At(x - p.off[i].X, y - p.off[i].Y)
    }
  }
  return transparent{}
}

func packImages(ims []image.Image) *packedImage {
  var p packedImage
  if len(ims) == 0 {
    panic("Cannot pack zero images")
  }
  p.ims = ims
  p.off = make([]image.Point, len(p.ims))
  sort.Sort(&p)

  run := 0
  height := 0
  max_width := 512
  max_height := 0
  for i := 1; i < len(p.off); i++ {
    run += p.ims[i - 1].Bounds().Dx()
    if run + p.ims[i].Bounds().Dx() > max_width {
      run = 0
      height += max_height
      max_height = 0
    }
    if p.ims[i].Bounds().Dy() > max_height {
      max_height = p.ims[i].Bounds().Dy()
    }
    p.off[i].X = run
    p.off[i].Y = height
  }
  for i := range p.ims {
    p.off[i] = p.off[i].Sub(p.ims[i].Bounds().Min)
  }

  // Done packing - now figure out the resulting bounds
  p.bounds.Min.X = 1e9  // if we exceed this something else will break first
  p.bounds.Min.Y = 1e9
  p.bounds.Max.X = -1e9
  p.bounds.Max.Y = -1e9
  for i := range p.ims {
    b := p.ims[i].Bounds()
    min := b.Add(p.off[i]).Min
    max := b.Add(p.off[i]).Max
    if min.X < p.bounds.Min.X {
      p.bounds.Min.X = min.X
    }
    if min.Y < p.bounds.Min.Y {
      p.bounds.Min.Y = min.Y
    }
    if max.X > p.bounds.Max.X {
      p.bounds.Max.X = max.X
    }
    if max.Y > p.bounds.Max.Y {
      p.bounds.Max.Y = max.Y
    }
  }

  sort.Sort(&packedImageSortByArea{&p})

  return &p
}

type runeInfo struct {
  pos      image.Rectangle
  bounds   image.Rectangle
  advance  float64
}
type Dictionary struct {
  Rgba  *image.RGBA
  info  map[rune]runeInfo

  // At what vertical value is the line on which text is logically rendered.
  // This is determined by the positioning of the '.' rune.
  baseline int

  miny,maxy int

  texture gl.Texture
}

// Renders the string on the quad spanning the specified coordinates.  The
// text will be rendering the current color.
func (d *Dictionary) RenderString(s string, x, y, x2, y2, z float64) {
  width := 0.0
  for _,r := range s {
    info,ok := d.info[r]
    if !ok {
      continue
    }
    width += info.advance
  }
  scale := (y2 - y) / float64(d.maxy - d.miny)

  gl.PushAttrib(gl.COLOR_BUFFER_BIT)
  defer gl.PopAttrib()
  gl.Enable(gl.BLEND)  
  gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

  gl.Enable(gl.TEXTURE_2D)
  d.texture.Bind(gl.TEXTURE_2D)
  gl.Begin(gl.QUADS)
  x_pos := x
  for _,r := range s {
    info,ok := d.info[r]
    if !ok {
      continue
    }
    xleft := x_pos + float64(info.bounds.Min.X) * scale
    xright := x_pos + float64(info.bounds.Max.X) * scale
    ytop := y + scale * float64(d.maxy - info.bounds.Min.Y)
    ybot := y2 + scale * float64(d.miny - info.bounds.Max.Y)

    gl.TexCoord2d(float64(info.pos.Min.X) / float64(d.Rgba.Bounds().Dx()), float64(info.pos.Max.Y) / float64(d.Rgba.Bounds().Dy()))
    gl.Vertex2d(xleft, ybot)

    gl.TexCoord2d(float64(info.pos.Min.X) / float64(d.Rgba.Bounds().Dx()), float64(info.pos.Min.Y) / float64(d.Rgba.Bounds().Dy()))
    gl.Vertex2d(xleft, ytop)

    gl.TexCoord2d(float64(info.pos.Max.X) / float64(d.Rgba.Bounds().Dx()), float64(info.pos.Min.Y) / float64(d.Rgba.Bounds().Dy()))
    gl.Vertex2d(xright, ytop)

    gl.TexCoord2d(float64(info.pos.Max.X) / float64(d.Rgba.Bounds().Dx()), float64(info.pos.Max.Y) / float64(d.Rgba.Bounds().Dy()))
    gl.Vertex2d(xright, ybot)

    x_pos += info.advance * scale
  }
  gl.End()
}

func MakeDictionary(font *truetype.Font) *Dictionary {
  alphabet := " abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()[]{};:'\",.<>/?\\|`~-_=+"
  context := freetype.NewContext()
  context.SetFont(font)
  width := 300
  height := 300
  context.SetSrc(image.White)
  point_size := 25.0
  dpi := 150
  context.SetFontSize(point_size)
  context.SetDPI(dpi)
  var letters []image.Image
  rune_mapping := make(map[rune]image.Image)
  rune_info := make(map[rune]runeInfo)
  for _,r := range alphabet {
    canvas := image.NewRGBA(image.Rect(-width/2, -height/2, width/2, height/2))
    context.SetDst(canvas)
    context.SetClip(canvas.Bounds())

    advance,_ := context.DrawString(string([]rune{r}), raster.Point{})
    sub := minimalSubImage(canvas)
    letters = append(letters, sub)
    rune_mapping[r] = sub
    adv_x := float64(advance.X) / 256.0
    rune_info[r] = runeInfo{ bounds: sub.bounds, advance: adv_x }
  }
  packed := packImages(letters)

  for _,r := range alphabet {
    ri := rune_info[r]
    ri.pos = packed.GetRect(rune_mapping[r])
    rune_info[r] = ri
  }

  dx := 1
  for dx < packed.Bounds().Dx() {
    dx = dx << 1
  }
  dy := 1
  for dy < packed.Bounds().Dy() {
    dy = dy << 1
  }

  pim := image.NewRGBA(image.Rect(0, 0, dx, dy))
  draw.Draw(pim, pim.Bounds(), packed, image.Point{}, draw.Over)
  var dict Dictionary
  dict.Rgba = pim
  dict.info = rune_info
  dict.baseline = dict.info['.'].bounds.Min.Y

  dict.miny = int(1e9)
  dict.maxy = int(-1e9)
  for _,info := range dict.info {
    if info.bounds.Min.Y < dict.miny {
      dict.miny = info.bounds.Min.Y
    }
    if info.bounds.Max.Y > dict.maxy {
      dict.maxy = info.bounds.Max.Y
    }
  }

  render.Queue(func() {
    gl.Enable(gl.TEXTURE_2D)
    dict.texture = gl.GenTexture()
    dict.texture.Bind(gl.TEXTURE_2D)
    gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
    gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
    gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
    gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
    gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
    glu.Build2DMipmaps(gl.TEXTURE_2D, 4, dict.Rgba.Bounds().Dx(), dict.Rgba.Bounds().Dy(), gl.RGBA, dict.Rgba.Pix)
    gl.Disable(gl.TEXTURE_2D)
  })

  f,err := os.Create("/Users/runningwild/code/src/github.com/runningwild/tester/dict.png")
  if err != nil {
    panic(err)
  }
  defer f.Close()
  err = png.Encode(f, dict.Rgba)
  if err != nil {
    panic(err)
  }

  return &dict
}

