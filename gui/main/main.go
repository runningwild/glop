package main

import(
  "fmt"
  "glop/gui"
  "glop/gos"
  "glop/gin"
  "glop/system"
  "runtime"
  "gl"
)

type BoxWidget struct {
  gui.EmbeddedWidget
  gui.Childless
  gui.Rectangle
  r,g,b,a float64
  on int
}
func (w *BoxWidget) Draw(region gui.Region) {
  w.Rectangle.Dims = region.Dims
  w.Rectangle.Point = region.Point
  if w.on > 0 {
    gl.Color3d(1, 0, 0)
  } else {
    gl.Color4d(w.r, w.g, w.b, w.a)
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
func MakeColorBoxWidget(r,g,b,a float64) *BoxWidget {
  var bw BoxWidget
  bw.EmbeddedWidget = &gui.BasicWidget{ CoreWidget : &bw }
  bw.r,bw.g,bw.b,bw.a = r,g,b,a
  return &bw
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
  gui.MustLoadFontAs("/Library/fonts/Tahoma.ttf", "standard")
  tw := gui.MakeTextEditLine("standard", "AAAVVV", 1, 1, 1, 1)
  fmt.Printf("tw: %v\n", tw)

  gl.MatrixMode(gl.PROJECTION)
  gl.Ortho(0, 800, 0, 600, 1, -1)
  gl.MatrixMode(gl.MODELVIEW)
  gl.LoadIdentity()
  ui := gui.Make(gin.In(), gui.Dims{ Dx : 800, Dy : 600})
  table := gui.MakeHorizontalTable(700)
  var vt *gui.VerticalTable
  vt = gui.MakeVerticalTable(500)
  vt.AddChild(MakeColorBoxWidget(1,0,0,1), true)
  vt.AddChild(MakeColorBoxWidget(0,1,0,1), false)
  vt.AddChild(MakeColorBoxWidget(0,0,1,1), true)
  table.AddChild(vt, false)
  vt = gui.MakeVerticalTable(500)
  vt.AddChild(MakeColorBoxWidget(1,0,0,1), false)
  vt.AddChild(MakeColorBoxWidget(0,1,0,1), true)
  vt.AddChild(MakeColorBoxWidget(0,0,1,1), true)
  table.AddChild(vt, false)
  vt = gui.MakeVerticalTable(500)
  vt.AddChild(MakeColorBoxWidget(1,0,0,1), true)
  vt.AddChild(MakeColorBoxWidget(0,1,0,1), true)
  vt.AddChild(MakeColorBoxWidget(0,0,1,1), true)
  table.AddChild(vt, false)
  ui.AddChild(table)
//  gin.In().RegisterEventListener(t)
  for gin.In().GetKey('q').FramePressCount() == 0 {
    sys.SwapBuffers()
    gl.ClearColor(0, 0, 0, 1)
    gl.Clear(gl.COLOR_BUFFER_BIT)
    sys.Think()
    ui.Draw()
  }
}








