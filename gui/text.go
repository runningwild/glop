package gui

import (
  "io"
  "encoding/gob"
  "image"
  "image/color"
  "image/draw"
  "runtime"
  "sort"
  "unsafe"
  "code.google.com/p/freetype-go/freetype"
  "code.google.com/p/freetype-go/freetype/raster"
  "code.google.com/p/freetype-go/freetype/truetype"
  "github.com/runningwild/glop/render"
  // "github.com/runningwild/opengl/gl"
  gl "github.com/chsc/gogl/gl21"
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
  Pos      image.Rectangle
  Bounds   image.Rectangle
  Advance  float64
}
type dictData struct {
  // The Pix data from the original image.Rgba
  Pix  []byte

  // Dx and Dy of the original image.Rgba
  Dx, Dy int

  // Map from rune to that rune's runeInfo.
  Info  map[rune]runeInfo

  // runeInfo for all r < 256 will be stored here as well as in info so we can
  // avoid map lookups if possible.
  Ascii_info []runeInfo

  // At what vertical value is the line on which text is logically rendered.
  // This is determined by the positioning of the '.' rune.
  Baseline int

  Miny,Maxy int
}
type Dictionary struct {
  data dictData

  texture uint32

  strs map[string]strBuffer

  dlists map[string]uint32
}
type strBuffer struct {
  vbuffer uint32
  vs []dictVert

  ibuffer uint32
  is []uint16
}
type dictVert struct {
  x,y float32
  u,v float32
}


func (d *Dictionary) getInfo(r rune) runeInfo {
  var info runeInfo
  if r >= 0 && r < 256 {
    info = d.data.Ascii_info[r]
  } else {
    info,_ = d.data.Info[r]
  }
  return info
}

// Figures out how wide a string will be if rendered at its natural size.
func (d *Dictionary) figureWidth(s string) float64 {
  w := 0.0
  for _,r := range s {
    w += d.getInfo(r).Advance
  }
  return w
}

type Justification int
const (
  Center Justification = iota
  Left
  Right
)

func (d *Dictionary) MaxHeight() float64 {
  return float64(d.data.Maxy - d.data.Miny)
}

func (d *Dictionary) RenderString(s string, x, y, z, height float64, just Justification) {
  strbuf, ok := d.strs[s]
  if !ok {
    defer d.RenderString(s, x, y, z, height, just)
  }
  size := unsafe.Sizeof(dictVert{})
  scale := height / float64(d.data.Maxy - d.data.Miny)
  width := float32(d.figureWidth(s) * scale)
  x_pos := float32(x)
  switch just {
    case Center:
      x_pos -= width / 2
    case Right:
      x_pos -= width
  }
  if ok {
    gl.PushMatrix()
    defer gl.PopMatrix()
    gl.Translated(float64(x_pos), y, z)
    gl.Scaled(scale, scale, 1)

    gl.PushAttrib(gl.COLOR_BUFFER_BIT)
    defer gl.PopAttrib()
    gl.Enable(gl.BLEND)
    gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

    gl.Enable(gl.TEXTURE_2D)
    gl.BindTexture(gl.TEXTURE_2D, d.texture)

    gl.BindBuffer(gl.ARRAY_BUFFER, strbuf.vbuffer)

    gl.EnableClientState(gl.VERTEX_ARRAY)
    gl.VertexPointer(2, gl.FLOAT, gl.Sizei(size), nil)

    gl.EnableClientState(gl.TEXTURE_COORD_ARRAY)
    gl.TexCoordPointer(2, gl.FLOAT, gl.Sizei(size), gl.Pointer(unsafe.Offsetof(strbuf.vs[0].u)))

    gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, strbuf.ibuffer)
    gl.DrawElements(gl.TRIANGLES, gl.Sizei(len(strbuf.is)), gl.UNSIGNED_SHORT, nil)

    gl.DisableClientState(gl.VERTEX_ARRAY)
    gl.DisableClientState(gl.TEXTURE_COORD_ARRAY)
    return
  }
  x_pos = 0
  for _,r := range s {
    info := d.getInfo(r)
    xleft := x_pos + float32(info.Bounds.Min.X)
    xright := x_pos + float32(info.Bounds.Max.X)
    ytop := float32(d.data.Maxy - info.Bounds.Min.Y)
    ybot := float32(height) + float32(d.data.Miny - info.Bounds.Max.Y)
    start := uint16(len(strbuf.vs))
    strbuf.is = append(strbuf.is, start+0)
    strbuf.is = append(strbuf.is, start+1)
    strbuf.is = append(strbuf.is, start+2)
    strbuf.is = append(strbuf.is, start+0)
    strbuf.is = append(strbuf.is, start+2)
    strbuf.is = append(strbuf.is, start+3)
    strbuf.vs = append(strbuf.vs, dictVert{
      x: xleft,
      y: ybot,
      u: float32(info.Pos.Min.X) / float32(d.data.Dx),
      v: float32(info.Pos.Max.Y) / float32(d.data.Dy),
    })
    strbuf.vs = append(strbuf.vs, dictVert{
      x: xleft,
      y: ytop,
      u: float32(info.Pos.Min.X) / float32(d.data.Dx),
      v: float32(info.Pos.Min.Y) / float32(d.data.Dy),
    })
    strbuf.vs = append(strbuf.vs, dictVert{
      x: xright,
      y: ytop,
      u: float32(info.Pos.Max.X) / float32(d.data.Dx),
      v: float32(info.Pos.Min.Y) / float32(d.data.Dy),
    })
    strbuf.vs = append(strbuf.vs, dictVert{
      x: xright,
      y: ybot,
      u: float32(info.Pos.Max.X) / float32(d.data.Dx),
      v: float32(info.Pos.Max.Y) / float32(d.data.Dy),
    })
    x_pos += float32(info.Advance)
  }
  gl.GenBuffers(1, &strbuf.vbuffer)
  gl.BindBuffer(gl.ARRAY_BUFFER, strbuf.vbuffer)
  gl.BufferData(gl.ARRAY_BUFFER, gl.Sizeiptr(int(size)*len(strbuf.vs)), gl.Pointer(&strbuf.vs[0].x), gl.STATIC_DRAW)

  gl.GenBuffers(1, &strbuf.ibuffer)
  gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, strbuf.ibuffer)
  gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, gl.Sizeiptr(int(unsafe.Sizeof(strbuf.is[0]))*len(strbuf.is)), gl.Pointer(&strbuf.is[0]), gl.STATIC_DRAW)
  d.strs[s] = strbuf

}

// Renders the string on the quad spanning the specified coordinates.  The
// text will be rendering the current color.
func (d *Dictionary) RenderStringOld(s string, x, y, z, height float64, just Justification) {
  list, ok := d.dlists[s]
  if !ok {
    defer d.RenderStringOld(s, x, y, z, height, just)
  }

  scale := height / float64(d.data.Maxy - d.data.Miny)
  width := d.figureWidth(s) * scale
  x_pos := x
  switch just {
    case Center:
      x_pos -= width / 2
    case Right:
      x_pos -= width
  }

  gl.PushMatrix()
  defer gl.PopMatrix()
  gl.Translated(x_pos, y, z)
  gl.Scaled(scale, scale, 1)

  gl.PushAttrib(gl.COLOR_BUFFER_BIT)
  defer gl.PopAttrib()
  gl.Enable(gl.BLEND)
  gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

  gl.Enable(gl.TEXTURE_2D)
  gl.BindTexture(gl.TEXTURE_2D, d.texture)

  if ok {
    gl.CallList(list)
    return
  }

  list = gl.GenLists(1)
  d.dlists[s] = list
  gl.NewList(list, gl.COMPILE)

  x_pos = 0

  gl.Begin(gl.QUADS)
  for _,r := range s {
    info := d.getInfo(r)
    xleft := x_pos + float64(info.Bounds.Min.X)
    xright := x_pos + float64(info.Bounds.Max.X)
    ytop := float64(d.data.Maxy - info.Bounds.Min.Y)
    ybot := height + float64(d.data.Miny - info.Bounds.Max.Y)

    gl.TexCoord2d(float64(info.Pos.Min.X) / float64(d.data.Dx), float64(info.Pos.Max.Y) / float64(d.data.Dy))
    gl.Vertex2d(xleft, ybot)

    gl.TexCoord2d(float64(info.Pos.Min.X) / float64(d.data.Dx), float64(info.Pos.Min.Y) / float64(d.data.Dy))
    gl.Vertex2d(xleft, ytop)

    gl.TexCoord2d(float64(info.Pos.Max.X) / float64(d.data.Dx), float64(info.Pos.Min.Y) / float64(d.data.Dy))
    gl.Vertex2d(xright, ytop)

    gl.TexCoord2d(float64(info.Pos.Max.X) / float64(d.data.Dx), float64(info.Pos.Max.Y) / float64(d.data.Dy))
    gl.Vertex2d(xright, ybot)

    x_pos += info.Advance
  }
  gl.End()
  gl.EndList()
}

func (d *Dictionary) Store(w io.Writer) error {
  return gob.NewEncoder(w).Encode(d.data)
}

func LoadDictionary(r io.Reader) (*Dictionary, error) {
  var d Dictionary
  err := gob.NewDecoder(r).Decode(&d.data)
  if err != nil {
    return nil, err
  }
  d.setupGlStuff()
  return &d, nil
}

func MakeDictionary(font *truetype.Font, size int) *Dictionary {
  alphabet := " abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()[]{};:'\",.<>/?\\|`~-_=+"
  context := freetype.NewContext()
  context.SetFont(font)
  width := 300
  height := 300
  context.SetSrc(image.White)
  dpi := 150
  context.SetFontSize(float64(size))
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
    rune_info[r] = runeInfo{ Bounds: sub.bounds, Advance: adv_x }
  }
  packed := packImages(letters)

  for _,r := range alphabet {
    ri := rune_info[r]
    ri.Pos = packed.GetRect(rune_mapping[r])
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
  draw.Draw(pim, pim.Bounds(), packed, image.Point{}, draw.Src)
  var dict Dictionary
  dict.data.Pix = pim.Pix
  dict.data.Dx = pim.Bounds().Dx()
  dict.data.Dy = pim.Bounds().Dy()
  dict.data.Info = rune_info

  dict.data.Ascii_info = make([]runeInfo, 256)
  for r := rune(0); r < 256; r++ {
    if info,ok := dict.data.Info[r]; ok {
      dict.data.Ascii_info[r] = info
    }
  }
  dict.data.Baseline = dict.data.Info['.'].Bounds.Min.Y

  dict.data.Miny = int(1e9)
  dict.data.Maxy = int(-1e9)
  for _,info := range dict.data.Info {
    if info.Bounds.Min.Y < dict.data.Miny {
      dict.data.Miny = info.Bounds.Min.Y
    }
    if info.Bounds.Max.Y > dict.data.Maxy {
      dict.data.Maxy = info.Bounds.Max.Y
    }
  }

  dict.setupGlStuff()

  return &dict
}

// Sets up anything that wouldn't have been loaded from disk, including
// all opengl data, and sets up finalizers for that data.
func (d *Dictionary) setupGlStuff() {
  d.dlists = make(map[string]uint32)
  d.strs = make(map[string]strBuffer)
  // TODO: This finalizer is untested
  runtime.SetFinalizer(d, func(d *Dictionary) {
    render.Queue(func() {
      for _,v := range d.dlists {
        gl.DeleteLists(v, 1)
      }
    })
  })

  render.Queue(func() {
    gl.Enable(gl.TEXTURE_2D)
    gl.GenTextures(1, &d.texture)
    gl.BindTexture(gl.TEXTURE_2D, d.texture)
    gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
    gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
    gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
    gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
    gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
    glu.Build2DMipmaps(gl.TEXTURE_2D, 4, d.data.Dx, d.data.Dy, gl.RGBA, d.data.Pix)
    gl.Disable(gl.TEXTURE_2D)
  })
}
