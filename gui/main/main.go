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
  gui.BasicZone
  r,g,b,a float64
  on int
}
func (w *BoxWidget) Draw(region gui.Region) {
  w.Render_region = region
  if w.on > 0 {
    gl.Color3d(1, 0, 0)
  } else {
    gl.Color4d(w.r, w.g, w.b, w.a)
  }
  gl.Disable(gl.TEXTURE_2D)
  gl.Begin(gl.QUADS)
    gl.Vertex2i(region.X, region.Y)
    gl.Vertex2i(region.X, region.Y+region.Dy)
    gl.Vertex2i(region.X+region.Dx, region.Y+region.Dy)
    gl.Vertex2i(region.X+region.Dx, region.Y)
  gl.End()
}
func (w *BoxWidget) DoThink(t int64) {
  w.on = w.on >> 1
}
func (w *BoxWidget) DoRespond(event_group gui.EventGroup) bool {
  if event_group.Events[0].Key.Cursor() != nil {
    w.on = 512-1
  }
  return false
}
func MakeColorBoxWidget(dx,dy int, r,g,b,a float64) *BoxWidget {
  var bw BoxWidget
  bw.EmbeddedWidget = &gui.BasicWidget{ CoreWidget : &bw }
  bw.Request_dims = gui.Dims{ dx, dy }
  bw.r,bw.g,bw.b,bw.a = r,g,b,a
  return &bw
}

type ExpandoBox struct {
  *BoxWidget
}
func MakeExpandoBox(dx,dy int, r,g,b,a float64) *ExpandoBox {
  var ex ExpandoBox
  ex.BoxWidget = MakeColorBoxWidget(dx, dy, r, g, b, a)
  return &ex
}
func (ex *ExpandoBox) Expandable() (bool,bool) {
  return true,true
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
  table := gui.MakeVerticalTable()
  s1 := gui.MakeHorizontalTable()
  s1.AddChild(MakeColorBoxWidget(121, 12, 1, 0, 0, 1))
  s1.AddChild(MakeExpandoBox(121, 12, 0, 0, 1, 1))
  table.AddChild(MakeColorBoxWidget(300, 100, 1,0,0,1))
  table.AddChild(MakeColorBoxWidget(100, 100, 0,1,0,1))
  table.AddChild(MakeColorBoxWidget(100, 100, 0,0,1,1))
  table.AddChild(MakeExpandoBox(10, 100, 1,1,1,1))
  table.AddChild(s1)
  table.AddChild(gui.MakeTextEditLine("standard", "Foo", 1, 1, 0, 1))
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








