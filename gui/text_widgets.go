package gui

import (
  "glop/gin"
  "fmt"
  "image"
  "image/draw"
  "freetype-go.googlecode.com/hg/freetype"
  "freetype-go.googlecode.com/hg/freetype/truetype"
  "gl"
  "gl/glu"
  "io/ioutil"
  "os"
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

func drawText(font *truetype.Font, c *freetype.Context, color image.Color, rgba *image.RGBA, text string) (int,int) {
  fg := image.NewColorImage(color)
  bg := image.Transparent
  draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)
  c.SetFont(font)
  c.SetDst(rgba)
  c.SetSrc(fg)
  c.SetClip(rgba.Bounds())
  // height is the fraction of the font that is above the line, 1.0 would mean
  // that the font never falls below the line
  height := 0.75
  pt := freetype.Pt(0, int(float64(c.FUnitToPixelRU(font.UnitsPerEm())) * height) )
  adv,_ := c.DrawString(text, pt)
  pt.X += adv.X
  py := int(float64(pt.Y >> 8) / height + 0.01)
  return int(pt.X >> 8), py
}

var basic_fonts map[string]*truetype.Font

func init() {
  basic_fonts = make(map[string]*truetype.Font)
}

func LoadFont(name,path string) os.Error {
  if _,ok := basic_fonts[name]; ok {
    return os.NewError(fmt.Sprintf("Cannot load two fonts with the same name: '%s'", name))
  }
  basic_fonts[name] = mustLoadFont(path)
  return nil
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
  color     image.Color
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
  t.dims.Dx, t.dims.Dy = drawText(t.font, t.context, t.color, image.NewRGBA(1, 1), t.text)
  t.rdims = Dims{
    Dx : int(nextPowerOf2(uint32(t.dims.Dx))),
    Dy : int(nextPowerOf2(uint32(t.dims.Dy))),
  }  
  t.rgba = image.NewRGBA(t.rdims.Dx, t.rdims.Dy)
  drawText(t.font, t.context, t.color, t.rgba, t.text)


  gl.Enable(gl.TEXTURE_2D)
  t.texture.Bind(gl.TEXTURE_2D)
  gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
  glu.Build2DMipmaps(gl.TEXTURE_2D, 4, t.rdims.Dx, t.rdims.Dy, gl.RGBA, t.rgba.Pix)
}

func MakeSingleLineText(font_name,text string, r,g,b,a float64) *SingleLineText {
  var t SingleLineText
  font,ok := basic_fonts[font_name]
  if !ok {
    return nil
  }
  t.font = font
  t.glyph_buf = truetype.NewGlyphBuf()
  t.text = text
  t.psize = 72
  t.context = freetype.NewContext()
  t.context.SetDPI(132)
  t.context.SetFontSize(18)
  t.texture = gl.GenTexture()
  t.SetColor(r, g, b, a)
  t.figureDims()
  return &t
}

func (t *SingleLineText) SetColor(r,g,b,a float64) {
  t.color = image.RGBAColor{
    R : uint8(255 * r),
    G : uint8(255 * g),
    B : uint8(255 * b),
    A : uint8(255 * a),
  }
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

func (t *SingleLineText) Draw(dims Dims) {
  gl.MatrixMode(gl.PROJECTION)
  gl.PushMatrix()
  defer gl.PopMatrix()
  defer gl.MatrixMode(gl.PROJECTION)
  gl.LoadIdentity()

  gl.Ortho(0, float64(dims.Dx), 0, float64(dims.Dy), -1, 1)
  gl.Enable(gl.TEXTURE_2D)
  gl.Enable(gl.BLEND)
  gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
  t.texture.Bind(gl.TEXTURE_2D)
  gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
  fdx := float64(t.rdims.Dx)
  fdy := float64(t.rdims.Dy)
  tx := float64(t.dims.Dx)/float64(t.rdims.Dx)
  ty := float64(t.dims.Dy)/float64(t.rdims.Dy)
  tx = 1
  ty = 1
  gl.Color4d(1.0, 1.0, 1.0, 1.0)
  gl.Begin(gl.QUADS)
    gl.TexCoord2d(0,0)
    gl.Vertex2d(0, 0)
    gl.TexCoord2d(0,-ty)
    gl.Vertex2d(0, fdy)
    gl.TexCoord2d(tx,-ty)
    gl.Vertex2d(fdx, fdy)
    gl.TexCoord2d(tx,0)
    gl.Vertex2d(fdx, 0)
  gl.End()
  gl.Disable(gl.TEXTURE_2D)
}

func MakeTextEntry(font_name,text string, r,g,b,a float64, dx,dy int) *TextEntry {
  ret := MakeSingleLineText(font_name, text, r,g,b,a)
  text := &TextEntry{ *ret, 0 }
  text.mdims.Dx = dx
  text.mdims.Dx = dy
}
type TextEntry struct {
  SingleLineText
  cursor_pos int
  mdims Dims
}

func (t *SingleLineText) Think(_ int64, _ bool, _ Region, _ map[Widget]Dims) (bool,Dims) {

}

func (t *TextEntry) HandleEventGroup(group gin.EventGroup) (consume bool, give bool, target *Node) {
  if group.Events[0].Type != gin.Press { return }
  key := group.Events[0].Key
  fmt.Printf("Jey id: %d\n", key.Id())
  if key.Id() >= 'a' && key.Id() <= 'z' {
    consume = true
    t.SetText(t.text + string([]byte{byte(key.Id())}))
  }
  return
}


