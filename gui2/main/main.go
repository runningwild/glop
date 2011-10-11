package main

import(
  "fmt"
  "glop/gui2"
  "glop/gos"
  "glop/gin"
  "glop/system"
  "runtime"
  "gl"
)

type VerticalTable struct {
  gui.EmbeddedWidget
  gui.StandardParent
  gui.Rectangle
}
func (w *VerticalTable) DoThink(t int64) {
  w.Dims = gui.Dims{}
  w.Dims.Dx = 400
  for i := range w.Children {
    if w.Children[i].Bounds().Dx > w.Dims.Dx {
      w.Dims.Dx = w.Children[i].Bounds().Dx
    }
    w.Dims.Dy += w.Children[i].Bounds().Dy
  }
  if w.Dims.Dx > 700 { w.Dims.Dx = 700 }
}
func (w *VerticalTable) DoRespond(event_group gui.EventGroup) bool {
  return false
}
func (w *VerticalTable) Draw(region gui.Region) {
  shrink := 1.0
  if region.Dy < w.Dims.Dy {
    shrink = float64(region.Dy) / float64(w.Dims.Dy)
  }
  for i := range w.Children {
    req := w.Children[i].Bounds()
    req.Dy = int(float64(req.Dy) * shrink)
    if req.Dx > w.Dims.Dx {
      req.Dx = w.Dims.Dx
    }
    req.Point = region.Point
    w.Children[i].Draw(req)
    region.Y += req.Dy
  }
}
func MakeVerticalTable() *VerticalTable {
  var t VerticalTable
  t.EmbeddedWidget = &gui.BasicWidget{ CoreWidget : &t }
  return &t
}

type BoxWidget struct {
  gui.EmbeddedWidget
  gui.Childless
  gui.Rectangle
  shade float64
  on int
}
func (w *BoxWidget) Draw(region gui.Region) {
  w.Rectangle.Constrain(region)
  if w.on > 0 {
    gl.Color3d(1, 0, 0)
  } else {
    gl.Color3d(w.shade, w.shade, w.shade)
  }
  gl.Disable(gl.TEXTURE_2D)
  gl.Begin(gl.QUADS)
    gl.Vertex2i(region.X, region.Y)
    gl.Vertex2i(region.X, region.Y+w.Rectangle.Dy)
    gl.Vertex2i(region.X+w.Rectangle.Dx, region.Y+w.Rectangle.Dy)
    gl.Vertex2i(region.X+w.Rectangle.Dx, region.Y)
  gl.End()
}
func (w *BoxWidget) DoThink(t int64) {
  w.Dims = gui.Dims{100, 100}
  w.on = w.on >> 1
}
func (w *BoxWidget) DoRespond(event_group gui.EventGroup) bool {
  if event_group.Events[0].Key.Cursor() != nil {
    w.on = 512-1
  }
  return false
}
func MakeBoxWidget(shade float64) *BoxWidget {
  var b BoxWidget
  b.EmbeddedWidget = &gui.BasicWidget{ CoreWidget : &b }
  b.shade = shade
  return &b
}

var (
  sys system.System
)

func main() {
  runtime.LockOSThread()
  sys = system.Make(gos.GetSystemInterface())
  sys.Startup()
  sys.CreateWindow(10, 10, 800, 600)
  sys.EnableVSync(true)

  fmt.Printf("")
  t := MakeVerticalTable()

  gui.MustLoadFontAs("/Library/fonts/Tahoma.ttf", "standard")
  tw := gui.MakeTextEditLine("standard", "AAAVVV", 1, 1, 1, 1)
  fmt.Printf("tw: %v\n", tw)


  N := 10
  for i := 0; i < N; i++ {
    t.AddChild(MakeBoxWidget(float64(i) / float64(N)))
    if i == N/2 {
      t.AddChild(tw)
    }
  }
  m := MakeBoxWidget(1)
  m.DoThink(1)
  gl.MatrixMode(gl.PROJECTION)
  gl.Ortho(0, 800, 0, 600, 1, -1)
  gl.MatrixMode(gl.MODELVIEW)
  gl.LoadIdentity()
  ui := gui.MakeGui(gin.In(), gui.Dims{ Dx : 800, Dy : 600})
  ui.AddChild(t)
//  gin.In().RegisterEventListener(t)
  for gin.In().GetKey('q').FramePressCount() == 0 {
    sys.SwapBuffers()
    gl.ClearColor(0, 0, 0, 1)
    gl.Clear(gl.COLOR_BUFFER_BIT)
    sys.Think()
    ui.Draw()
  }
}








