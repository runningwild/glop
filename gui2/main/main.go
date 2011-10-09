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
  gui.BasicWidget
  gui.StandardParent
  gui.Rectangle
}
func (w *VerticalTable) HandleEventGroup(gin_group gin.EventGroup) {
  w.Respond(gui.EventGroup{gin_group, false})
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
}
func (w *VerticalTable) DoRespond(event_group gui.EventGroup) bool {
  return false
}
func (w *VerticalTable) Draw(region gui.Region) {
  shrink := float64(region.Dy) / float64(w.Dims.Dy)
  for i := range w.Children {
    req := w.Children[i].Bounds()
    req.Dy = int(float64(req.Dy) * shrink)
    req.Point = region.Point
    w.Children[i].Draw(req)
    region.Y += req.Dy
  }
}
func MakeVerticalTable() *VerticalTable {
  var t VerticalTable
  t.BasicWidget.CoreWidget = &t
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
  gl.Begin(gl.QUADS)
    gl.Vertex2i(region.X, region.Y)
    gl.Vertex2i(region.X, region.Y+w.Rectangle.Dy)
    gl.Vertex2i(region.X+w.Rectangle.Dx, region.Y+w.Rectangle.Dy)
    gl.Vertex2i(region.X+w.Rectangle.Dx, region.Y)
  gl.End()
}
func (w *BoxWidget) DoThink(t int64) {
  w.Dims = gui.Dims{100, 100}
  fmt.Printf("Think\n")
  w.on = w.on >> 1
}
func (w *BoxWidget) DoRespond(event_group gui.EventGroup) bool {
  fmt.Printf("Event: %v\n", event_group.Events[0].Key)
  if event_group.Events[0].Key.Cursor() != nil {
    w.on = 3
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
  fmt.Printf("")
  t := MakeVerticalTable()
  N := 25
  for i := 0; i < N; i++ {
    t.AddWidget(MakeBoxWidget(float64(i) / float64(N)))
  }
  m := MakeBoxWidget(1)
  m.DoThink(1)
  runtime.LockOSThread()
  sys = system.Make(gos.GetSystemInterface())
  sys.Startup()
  sys.CreateWindow(10, 10, 800, 600)
  sys.EnableVSync(true)
  gl.Ortho(0, 800, 0, 600, 1, -1)
  gin.In().RegisterEventListener(t)
  for gin.In().GetKey('q').FramePressCount() == 0 {
    sys.SwapBuffers()
    gl.ClearColor(0, 0, 0, 1)
    gl.Clear(gl.COLOR_BUFFER_BIT)
    sys.Think()
    t.Draw(gui.Region{ gui.Point{0, 0}, gui.Dims{800, 600}})
  }
}








