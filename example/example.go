package main

import (
  "glop/gos"
  "glop/gui"
  "glop/gin"
  "glop/system"
  "glop/sprite"
  "runtime"
  "time"
  "fmt"
  "io/ioutil"
  "image"
  "image/draw"
  "freetype-go.googlecode.com/hg/freetype"
  "freetype-go.googlecode.com/hg/freetype/truetype"
  "gl"
  "gl/glu"
  "os"
  "path"
)

var (
  sys system.System
)

func init() {
  sys = system.Make(gos.GetSystemInterface())
}

func loadFont(filename string) (*truetype.Font, os.Error) {
  data,err := ioutil.ReadFile(filename)
  if err != nil {
    return nil, err
  }
  font,err := freetype.ParseFont(data)
  if err != nil {
    return nil, err
  }
  return font,nil
}

func makeContext() (*freetype.Context, os.Error) {
  c := freetype.NewContext()
  c.SetDPI(200)
  c.SetFontSize(15)
  return c, nil
}

func drawText(font *truetype.Font, c *freetype.Context, rgba *image.RGBA, texture gl.Texture, text []string) os.Error {
  fg, bg := image.Black, image.White
  draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)
  c.SetFont(font)
  c.SetDst(rgba)
  c.SetSrc(fg)
  c.SetClip(rgba.Bounds())
  pt := freetype.Pt(10, 10+c.FUnitToPixelRU(font.UnitsPerEm()))
  for _, s := range text {
    _, err := c.DrawString(s, pt)
    if err != nil {
      return err
    }
    pt.Y += c.PointToFix32(15 * 1.5)
  }
  gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
  glu.Build2DMipmaps(gl.TEXTURE_2D, 4, 1024, 1024, gl.RGBA, rgba.Pix)
  return nil
}

type Foo struct {
  *gui.BoxWidget
}
var window system.Window
func (f *Foo) Think(_ int64, has_focus bool, previous gui.Region, _ map[gui.Widget]gui.Dims) (bool, gui.Dims) {
  cx,cy := sys.GetCursorPos(window)
  cursor := gui.Point{ X : cx, Y : cy }
  if cursor.Inside(previous) {
    f.Dims.Dx += 5
    f.G = 0
    f.B = 0
  } else {
    f.G = 1
    f.B = 1
  }
  return f.BoxWidget.Think(0, false, previous, nil)
}

func main() {
  runtime.LockOSThread()
  sys.Startup()

  fontpath := os.Args[0] + "/../../fonts/skia.ttf"
  fontpath = path.Clean(fontpath)
  font,err := loadFont(fontpath)
  if err != nil {
    panic(err.String())
  }

  window = sys.CreateWindow(10, 10, 768, 576)
  ticker := time.Tick(5e7)
  ui := gui.Make(sys.Input(), 768, 576)
  anch := ui.Root.InstallWidget(gui.MakeAnchorBox(gui.Dims{768, 576}), nil)
  anch.InstallWidget(gui.MakeBoxWidget(250,250,0,0,1,1), gui.Anchor{0.5, 0.5, 0.5, 0.5})
  anch.InstallWidget(
      &Foo{gui.BoxWidget : gui.MakeBoxWidget(100, 100, 1, 1, 1, 1)},
      gui.Anchor{ Bx:0.7, By:1, Wx:0.5, Wy:1})
  anch.InstallWidget(
      &Foo{gui.BoxWidget : gui.MakeBoxWidget(100, 100, 1, 1, 1, 1)},
      gui.Anchor{ Bx:0, By:0.5, Wx:0, Wy:0.5})
  anch.InstallWidget(
      &Foo{gui.BoxWidget : gui.MakeBoxWidget(100, 100, 1, 1, 1, 1)},
      gui.Anchor{ Bx:0.2, By:0.2, Wx:1, Wy:1})
  text_widget := gui.MakeSingleLineText(font, "Funk Monkey 7$")
  anch.InstallWidget(gui.MakeBoxWidget(450,50,0,1,0,1), gui.Anchor{0,1, 0,1})
  anch.InstallWidget(text_widget, gui.Anchor{0,1,0,1})
  frame_count_widget := gui.MakeSingleLineText(font, "Frame")
  anch.InstallWidget(gui.MakeBoxWidget(250,50,0,1,0,1), gui.Anchor{1,1, 1,1})
  anch.InstallWidget(frame_count_widget, gui.Anchor{1,1,1,1})

  n := 0
  spritepath := os.Args[0] + "/../../sprites/test_sprite"
  spritepath = path.Clean(spritepath)
  s,err := sprite.LoadSprite(spritepath)
  if err != nil {
    panic(err.String())
  }
  s.Stats()

  key_bindings := make(map[byte]string)
  key_bindings['a'] = "defend"
  key_bindings['s'] = "undamaged"
  key_bindings['d'] = "damaged"
  key_bindings['f'] = "killed"
  key_bindings['z'] = "ranged"
  key_bindings['x'] = "melee"
  key_bindings['c'] = "move"
  key_bindings['v'] = "stop"
  key_bindings['o'] = "turn_left"
  key_bindings['p'] = "turn_right"

  for {
    n++
    frame_count_widget.SetText(fmt.Sprintf("%d", n))
    s.RenderToQuad()
    s.Think(50)
    sys.SwapBuffers(window)
    <-ticker
    text_widget.SetText(s.CurState())
    sys.Think()
    groups := sys.GetInputEvents()

    for _,group := range groups {
      for key,cmd := range key_bindings {
        if found,event := group.FindEvent(gin.KeyId(key)); found && event.Type == gin.Press {
          s.Command(cmd)
        }
      }
      if found,_ := group.FindEvent('q'); found {
        return
      }
    }
  }

  fmt.Printf("")
}
