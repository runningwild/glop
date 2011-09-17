package gui

import (
  "image"
  "image/draw"
  "freetype-go.googlecode.com/hg/freetype"
  "freetype-go.googlecode.com/hg/freetype/truetype"
  "gl"
  "gl/glu"
  "io/ioutil"
  "os"
  "path"
)

func mustLoadFont(filename string) *truetype.Font {
  data,err := ioutil.ReadFile(filename)
  if err != nil {
    panic(err.String())
  }
  font,err := freetype.ParseFont(data)
  if err != nil {
    panic(err.String())
  }
  return font
}

func drawText(font *truetype.Font, c *freetype.Context, rgba *image.RGBA, text string) (int,int) {
  fg, bg := image.Black, image.White
  draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)
  c.SetFont(font)
  c.SetDst(rgba)
  c.SetSrc(fg)
  c.SetClip(rgba.Bounds())
  pt := freetype.Pt(0, int(float64(c.FUnitToPixelRU(font.UnitsPerEm())) * 0.85) )
  adv,_ := c.DrawString(text, pt)
  pt.X += adv.X
  py := int(float64(pt.Y >> 8) / 0.85 + 0.01)
  return int(pt.X >> 8), py
}

var basic_font *truetype.Font

func init() {
  fontpath := os.Args[0] + "/../../fonts/skia.ttf"
  fontpath = path.Clean(fontpath)
  basic_font = mustLoadFont(fontpath)
}

type SingleLineText struct {
  Childless
  Stoic
  text      string
  changed   bool
  dims      Dims
  rdims     Dims
  psize     int
  font      *truetype.Font
  context   *freetype.Context
  glyph_buf *truetype.GlyphBuf
  texture   gl.Texture
  rgba      *image.RGBA
}

func nextPowerOf2(n uint32) uint32 {
  if n == 0 { return 1 }
  for i := uint(0); i < 32; i++ {
    p := uint32(1) << i
    if n <= p { return p }
  }
  return 0
}

func (t *SingleLineText) figureDims() {
  t.dims.Dx, t.dims.Dy = drawText(t.font, t.context, image.NewRGBA(1, 1), t.text)
  t.rdims = Dims{
    Dx : int(nextPowerOf2(uint32(t.dims.Dx))),
    Dy : int(nextPowerOf2(uint32(t.dims.Dy))),
  }  
  t.rgba = image.NewRGBA(t.rdims.Dx, t.rdims.Dy)
  drawText(t.font, t.context, t.rgba, t.text)


  gl.Enable(gl.TEXTURE_2D)
  t.texture.Bind(gl.TEXTURE_2D)
  gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
  glu.Build2DMipmaps(gl.TEXTURE_2D, 4, t.rdims.Dx, t.rdims.Dy, gl.RGBA, t.rgba.Pix)
}

func MakeSingleLineText(font *truetype.Font, str string) *SingleLineText {
  var t SingleLineText
  t.glyph_buf = truetype.NewGlyphBuf()
  t.text = str
  t.font = font
  t.psize = 72
  t.context = freetype.NewContext()
  t.context.SetDPI(113)
  t.context.SetFontSize(18)
  t.texture = gl.GenTexture()
  t.figureDims()
  return &t
}

func (t *SingleLineText) SetText(str string) {
  if t.text != str {
    t.text = str
    t.changed = true
  }
}

func (t *SingleLineText) Think(_ int64, _ bool, _ Region, _ map[Widget]Dims) (bool,Dims) {
  if !t.changed {
    return false, t.dims
  }
  t.changed = false
  t.figureDims()
  return false, t.dims
}

func (t *SingleLineText) Draw(region Region) {
  gl.Enable(gl.TEXTURE_2D)
  gl.Enable(gl.BLEND)
  t.texture.Bind(gl.TEXTURE_2D)
  gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
  fx := float64(region.X)
  fy := float64(region.Y)
  fdx := float64(t.dims.Dx)
  fdy := float64(t.dims.Dy)
  tx := float64(t.dims.Dx)/float64(t.rdims.Dx)
  ty := float64(t.dims.Dy)/float64(t.rdims.Dy)
  gl.Color4d(1.0, 1.0, 1.0, 0.7)
  gl.Begin(gl.QUADS)
    gl.TexCoord2d(0,ty)
    gl.Vertex2d(fx, fy)
    gl.TexCoord2d(0,0)
    gl.Vertex2d(fx, fy+fdy)
    gl.TexCoord2d(tx,0)
    gl.Vertex2d(fx+fdx, fy+fdy)
    gl.TexCoord2d(tx,ty)
    gl.Vertex2d(fx+fdx, fy)
  gl.End()
  gl.Disable(gl.TEXTURE_2D)
}
