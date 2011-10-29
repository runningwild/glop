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
  gui.CollapsableZone
  gui.EmbeddedWidget
  gui.Childless
  gui.NonFocuser
  r,g,b,a float64
  on int
  vbo gl.Buffer
}
func (w *BoxWidget) String() string {
  return "box widget"
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
func (w *BoxWidget) DoThink(t int64, _ bool) {
  w.on = w.on >> 1
}
func (w *BoxWidget) DoRespond(event_group gui.EventGroup) (consume,take_focus bool) {
  if event_group.Events[0].Key.Cursor() != nil {
    w.on = 512-1
    consume = true
  }
  return
}
func MakeColorBoxWidget(dx,dy int, r,g,b,a float64) *BoxWidget {
  var bw BoxWidget
  bw.EmbeddedWidget = &gui.BasicWidget{ CoreWidget : &bw }
  bw.Request_dims = gui.Dims{ dx, dy }
  bw.r,bw.g,bw.b,bw.a = r,g,b,a
  bw.vbo = gl.GenBuffer()
  bw.vbo.Bind(gl.ARRAY_BUFFER)
  return &bw
}

type ExpandoBox struct {
  *BoxWidget
}
func (w *ExpandoBox) String() string {
  return "expando box"
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

func init() {
  runtime.LockOSThread()
}

func main() {
  sys = system.Make(gos.GetSystemInterface())
  sys.Startup()
  sys.CreateWindow(10, 10, 800, 600)
  sys.EnableVSync(true)

  fmt.Printf("")
  gui.MustLoadFontAs("/Library/fonts/Tahoma.ttf", "standard")

  gl.MatrixMode(gl.PROJECTION)
  gl.Ortho(0, 800, 0, 600, 1, -1)
  gl.MatrixMode(gl.MODELVIEW)
  gl.LoadIdentity()
  ui := gui.Make(gin.In(), gui.Dims{ Dx : 800, Dy : 600})
  table := gui.MakeVerticalTable()
  box := MakeColorBoxWidget(100, 100, 0, 1, 1, 1)
  table.AddChild(box)
  ui.AddChild(table)
//  gin.In().RegisterEventListener(t)
  for gin.In().GetKey('q').FramePressCount() == 0 {
    sys.SwapBuffers()
    if gin.In().GetKey('w').FramePressCount() > 0 {
      box.Collapsed = true
    }
    if gin.In().GetKey('e').FramePressCount() > 0 {
      box.Collapsed = false
    }
    gl.ClearColor(0, 0, 0, 1)
    gl.Clear(gl.COLOR_BUFFER_BIT)
    sys.Think()
    ui.Draw()
  }
}


