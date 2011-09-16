package main

import (
  "glop/gos"
  "glop/gui"
  "glop/system"
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

func gameLoop() {
  window := sys.CreateWindow(10, 10, 1200, 700)
  texture := gl.GenTexture()
  texture.Bind(gl.TEXTURE_2D)
  fontpath := os.Args[0] + "/../../fonts/luxisr.ttf"
  fontpath = path.Clean(fontpath)
  font,err := loadFont(fontpath)
  if err != nil {
    fmt.Printf("Failed to load font: %s\n", err.String())
    return
  }
  context,err := makeContext()
  if err != nil {
    fmt.Printf("Failed to make font context: %s\n", err.String())
    return
  }
  rgba := image.NewRGBA(1024, 1024)
  for i := 0; i < 00; i++ {
    drawText(font, context, rgba, texture, []string{"Rawr!!!"})
  }

  err = drawText(font, context, rgba, texture, []string{"Rawr!!! :-D"})
  if err != nil {
    fmt.Printf("Couldn't render texture: %s\n", err.String())
    return
  }

  ticker := time.Tick(1*16666667)
  text := make([]string, 100)[0:0]
  for {
    sys.SwapBuffers(window)
    <-ticker
    err := drawText(font, context, rgba, texture, text)
    if err != nil {
      fmt.Printf("Couldn't draw text: %s\n", err.String())
      return
    }
    gl.LoadIdentity();
    gl.Ortho(-2.4,2.4, -1.4,1.4, -1.1,1.1)
    gl.MatrixMode(gl.MODELVIEW)
    gl.ClearColor(0.2, 0.0, 1.0, 1.0)
    gl.Clear(0x00004000)
    gl.Color3d(0, 1, 0)
    gl.Color4d(1.0, 1.0, 1.0, 0.7)
    gl.Enable(gl.TEXTURE_2D)
    gl.Enable(gl.BLEND)
    texture.Bind(gl.TEXTURE_2D)
    gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
    gl.Begin(gl.QUADS)
      gl.TexCoord2d(0,1)
      gl.Vertex2d(-1.4,-1.4)
      gl.TexCoord2d(0,0)
      gl.Vertex2d(-1.4, 1.4)
      gl.TexCoord2d(1,0)
      gl.Vertex2d( 1.4, 1.4)
      gl.TexCoord2d(1,1)
      gl.Vertex2d( 1.4,-1.4)
    gl.End()
    gl.Disable(gl.TEXTURE_2D)
    sys.Think()
    groups := sys.GetInputEvents()
    text = make([]string, 0)
    text = append(text, fmt.Sprintf("%d Groups", len(groups)))
    for gi,group := range groups {
      if len(group.Events) > 0 && group.Events[0].Key.Id() == 'q' {
        return
      }
      for _,event := range group.Events {
        if event.Key.Id() == sys.Input().GetKey(300).Id() {
          text = append(text, fmt.Sprintf("%d: %v %v %f", gi, event.Type, event.Key, event.Key.FramePressSum()))
        }
      }
    }
//    text = make([]string, len(events))
 //   for i := range events {
//      text[i] = fmt.Sprintf("%v", events[i])
 //   }
  }
}

type Foo struct {
  *gui.BoxWidget
}
var window system.Window
func (f *Foo) Think(_ int64, has_focus bool, previous gui.Region, _ map[gui.Widget]gui.Dims) (bool, gui.Dims) {
  cx,cy := sys.GetCursorPos(window)
  cursor := gui.Point{ X : cx, Y : cy }
  if cursor.Inside(previous) {
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
  window = sys.CreateWindow(10, 10, 800, 600)
  ticker := time.Tick(5e7)
  ui := gui.Make(sys.Input(), 800, 600)
  table := ui.Root.InstallWidget(new(gui.VerticalTable), nil)
  table.InstallWidget(&Foo{gui.BoxWidget : gui.MakeBoxWidget(100, 100, 1, 1, 1, 1)}, nil)
  table.InstallWidget(&Foo{gui.BoxWidget : gui.MakeBoxWidget(100, 100, 1, 1, 1, 1)}, nil)
  table.InstallWidget(&Foo{gui.BoxWidget : gui.MakeBoxWidget(100, 100, 1, 1, 1, 1)}, nil)
  table.InstallWidget(&Foo{gui.BoxWidget : gui.MakeBoxWidget(100, 100, 1, 1, 1, 1)}, nil)
  table.InstallWidget(&Foo{gui.BoxWidget : gui.MakeBoxWidget(100, 100, 1, 1, 1, 1)}, nil)
  table.InstallWidget(&Foo{gui.BoxWidget : gui.MakeBoxWidget(100, 100, 1, 1, 1, 1)}, nil)
  table.InstallWidget(&Foo{gui.BoxWidget : gui.MakeBoxWidget(100, 100, 1, 1, 1, 1)}, nil)
  table.InstallWidget(&Foo{gui.BoxWidget : gui.MakeBoxWidget(100, 100, 1, 1, 1, 1)}, nil)
  for {
    sys.SwapBuffers(window)
    <-ticker
    sys.Think()
    ui.Draw()
    groups := sys.GetInputEvents()
    for _,group := range groups {
      if found,_ := group.FindEvent('e'); found {
        x,y := sys.GetCursorPos(window)
        fmt.Printf("MOUSE: %d %d\n", x, y)
      }
      if found,_ := group.FindEvent('q'); found {
        return
      }
    }
  }

//  gameLoop()
  fmt.Printf("")
}
