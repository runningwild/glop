package gui

import (
  "fmt"
  "image"
  "image/draw"
  "image/color"
  "code.google.com/p/freetype-go/freetype"
  "code.google.com/p/freetype-go/freetype/truetype"
  "github.com/runningwild/opengl/gl"
  "github.com/runningwild/opengl/glu"
  "io/ioutil"
)

type guiError struct {
  ErrorString string
}
func (g *guiError) Error() string {
  return g.ErrorString
}

func MustLoadFontAs(path, name string) {
  if _, ok := basic_fonts[name]; ok {
    panic(fmt.Sprintf("Cannot load two fonts with the same name: '%s'.", name))
  }
  data, err := ioutil.ReadFile(path)
  if err != nil {
    panic(err.Error())
  }
  font, err := freetype.ParseFont(data)
  if err != nil {
    panic(err.Error())
  }
  basic_fonts[name] = font
}

func LoadFontAs(path, name string) error {
  if _, ok := basic_fonts[name]; ok {
    return &guiError{ fmt.Sprintf("Cannot load two fonts with the same name: '%s'.", name) }
  }
  data, err := ioutil.ReadFile(path)
  if err != nil {
    return err
  }
  font, err := freetype.ParseFont(data)
  if err != nil {
    return err
  }
  basic_fonts[name] = font
  return nil
}

func GetDict(name string) *Dictionary {
  d,ok := basic_dicts[name]
  if ok {
    return d
  }
  basic_dicts[name] = MakeDictionary(basic_fonts[name])
  return basic_dicts[name]
}

func drawText(font *truetype.Font, c *freetype.Context, color color.Color, rgba *image.RGBA, text string) (int, int) {
  fg := image.NewUniform(color)
  bg := image.Transparent
  draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)
  c.SetFont(font)
  c.SetDst(rgba)
  c.SetSrc(fg)
  c.SetClip(rgba.Bounds())
  // height is the fraction of the font that is above the line, 1.0 would mean
  // that the font never falls below the line
  // TODO: wtf - this is all wrong!
  // fix fonts - we can't change the font size easily
  height := 1.3
  pt := freetype.Pt(0, int(float64(c.FUnitToPixelRU(font.UnitsPerEm()))*height))
  adv, _ := c.DrawString(text, pt)
  pt.X += adv.X
  py := int(float64(pt.Y>>8)/height + 0.01)
  return int(pt.X >> 8), py
}

var basic_fonts map[string]*truetype.Font
var basic_dicts map[string]*Dictionary

func init() {
  basic_fonts = make(map[string]*truetype.Font)
  basic_dicts = make(map[string]*Dictionary)
}

type TextLine struct {
  EmbeddedWidget
  Childless
  NonResponder
  NonFocuser
  BasicZone
  text      string
  next_text string
  initted   bool
  rdims     Dims
  font      *truetype.Font
  context   *freetype.Context
  glyph_buf *truetype.GlyphBuf
  texture   gl.Texture
  rgba      *image.RGBA
  color     color.Color
  scale     float64
}

func (w *TextLine) String() string {
  return "text line"
}

func nextPowerOf2(n uint32) uint32 {
  if n == 0 {
    return 1
  }
  for i := uint(0); i < 32; i++ {
    p := uint32(1) << i
    if n <= p {
      return p
    }
  }
  return 0
}

func (w *TextLine) figureDims() {
  // Always draw the text as white on a transparent background so that we can change
  // the color easily through opengl
  w.rdims.Dx, w.rdims.Dy = drawText(w.font, w.context, color.RGBA{255, 255, 255, 255}, image.NewRGBA(image.Rect(0, 0, 1, 1)), w.text)
  texture_dims := Dims{
    Dx: int(nextPowerOf2(uint32(w.rdims.Dx))),
    Dy: int(nextPowerOf2(uint32(w.rdims.Dy))),
  }
  w.rgba = image.NewRGBA(image.Rect(0, 0, texture_dims.Dx, texture_dims.Dy))
  drawText(w.font, w.context, color.RGBA{255, 255, 255, 255}, w.rgba, w.text)

  gl.Enable(gl.TEXTURE_2D)
  w.texture.Bind(gl.TEXTURE_2D)
  gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
  glu.Build2DMipmaps(gl.TEXTURE_2D, 4, w.rgba.Bounds().Dx(), w.rgba.Bounds().Dy(), gl.RGBA, w.rgba.Pix)
  gl.Disable(gl.TEXTURE_2D)
}

type Button struct {
  *TextLine
  Clickable
}

func MakeButton(font_name, text string, width int, r, g, b, a float64, f func(int64)) *Button {
  var btn Button
  btn.TextLine = MakeTextLine(font_name, text, width, r, g, b, a)
  btn.TextLine.EmbeddedWidget = &BasicWidget{CoreWidget: &btn}
  btn.on_click = f
  return &btn
}

func MakeTextLine(font_name, text string, width int, r, g, b, a float64) *TextLine {
  var w TextLine
  w.EmbeddedWidget = &BasicWidget{CoreWidget: &w}
  font, ok := basic_fonts[font_name]
  if !ok {
    panic(fmt.Sprintf("Unable to find a font registered as '%s'.", font_name))
  }
  w.font = font
  w.glyph_buf = truetype.NewGlyphBuf()
  w.next_text = text
  w.context = freetype.NewContext()
  w.context.SetDPI(132)
  w.context.SetFontSize(12)
  w.SetColor(r, g, b, a)
  w.Request_dims = Dims{width, 35}
  return &w
}

func (w *TextLine) SetColor(r, g, b, a float64) {
  w.color = color.RGBA{
    R: uint8(255 * r),
    G: uint8(255 * g),
    B: uint8(255 * b),
    A: uint8(255 * a),
  }
}

func (w *TextLine) GetText() string {
  return w.next_text
}

func (w *TextLine) SetText(str string) {
  if w.text != str {
    w.next_text = str
  }
}

func (w *TextLine) DoThink(int64, bool) {
}

func (w *TextLine) preDraw(region Region) {
  if !w.initted {
    w.initted = true
    w.texture = gl.GenTexture()
    w.text = w.next_text
    w.figureDims()
  }
  if w.text != w.next_text {
    w.text = w.next_text
    w.figureDims()
  }

  gl.PushMatrix()

  gl.Color3d(0, 0, 0)
  gl.Begin(gl.QUADS)
    gl.Vertex2i(region.X, region.Y)
    gl.Vertex2i(region.X, region.Y + region.Dy)
    gl.Vertex2i(region.X + region.Dx, region.Y + region.Dy)
    gl.Vertex2i(region.X + region.Dx, region.Y)
  gl.End()

  gl.PushAttrib(gl.TEXTURE_BIT)
  gl.Enable(gl.TEXTURE_2D)
  w.texture.Bind(gl.TEXTURE_2D)

  gl.PushAttrib(gl.COLOR_BUFFER_BIT)
  gl.Enable(gl.BLEND)
  gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
}

func (w *TextLine) postDraw(region Region) {
  gl.PopAttrib()
  gl.PopAttrib()
  gl.PopMatrix()
}

func (w *TextLine) Draw(region Region) {
  region.PushClipPlanes()
  defer region.PopClipPlanes()
  w.preDraw(region)
  w.coreDraw(region)
  w.postDraw(region)
}

func (w *TextLine) coreDraw(region Region) {
  if region.Size() == 0 { return }
  gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
  gl.Color4d(1.0, 1.0, 1.0, 1.0)
  req := w.Request_dims
  if req.Dx > region.Dx {
    req.Dx = region.Dx
  }
  if req.Dy > region.Dy {
    req.Dy = region.Dy
  }
  if req.Dx*region.Dy < req.Dy*region.Dx {
    req.Dy = (region.Dy * req.Dx) / region.Dx
  } else {
    req.Dx = (region.Dx * req.Dy) / region.Dy
  }
  w.Render_region.Dims = req
  w.Render_region.Point = region.Point
  tx := float64(w.rdims.Dx) / float64(w.rgba.Bounds().Dx())
  ty := float64(w.rdims.Dy) / float64(w.rgba.Bounds().Dy())
  //  w.scale = float64(w.Render_region.Dx) / float64(w.rdims.Dx)
  {
    r, g, b, a := w.color.RGBA()
    gl.Color4d(float64(r)/65535, float64(g)/65535, float64(b)/65535, float64(a)/65535)
  }
  gl.Begin(gl.QUADS)
  gl.TexCoord2d(0, 0)
  gl.Vertex2i(region.X, region.Y)
  gl.TexCoord2d(0, -ty)
  gl.Vertex2i(region.X, region.Y+w.rdims.Dy)
  gl.TexCoord2d(tx, -ty)
  gl.Vertex2i(region.X+w.rdims.Dx, region.Y+w.rdims.Dy)
  gl.TexCoord2d(tx, 0)
  gl.Vertex2i(region.X+w.rdims.Dx, region.Y)
  gl.End()
}
