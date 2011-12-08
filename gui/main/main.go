package main

import (
  "fmt"
  "gl"
  "glop/gos"
  "glop/gin"
  "glop/gui"
  "glop/system"
  "runtime"
)

type BoxWidget struct {
  gui.CollapsableZone
  gui.EmbeddedWidget
  gui.Childless
  gui.NonFocuser
  r, g, b, a float64
  on         int
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
func (w *BoxWidget) DoRespond(event_group gui.EventGroup) (consume, take_focus bool) {
  if event_group.Events[0].Key.Cursor() != nil {
   // w.on = 512 - 1
    consume = true
  }
  return
}
func MakeColorBoxWidget(dx, dy int, r, g, b, a float64) *BoxWidget {
  var bw BoxWidget
  bw.EmbeddedWidget = &gui.BasicWidget{CoreWidget: &bw}
  bw.Request_dims = gui.Dims{dx, dy}
  bw.r, bw.g, bw.b, bw.a = r, g, b, a
  return &bw
}

type ExpandoBox struct {
  *BoxWidget
}

func (w *ExpandoBox) String() string {
  return "expando box"
}
func MakeExpandoBox(dx, dy int, r, g, b, a float64) *ExpandoBox {
  var ex ExpandoBox
  ex.BoxWidget = MakeColorBoxWidget(dx, dy, r, g, b, a)
  return &ex
}
func (ex *ExpandoBox) Expandable() (bool, bool) {
  return true, true
}

var (
  sys system.System
)

func init() {
  runtime.LockOSThread()
  gui.MustLoadFontAs("/Users/runningwild/code/go-glop/example/data/fonts/skia.ttf", "standard")
}

func main() {
  sys = system.Make(gos.GetSystemInterface())
  sys.Startup()
  wdx := 800
  wdy := 600
  sys.CreateWindow(10, 10, wdx, wdy)
  sys.EnableVSync(true)
  ui := gui.Make(gin.In(), gui.Dims{ wdx, wdy })

  vtable := gui.MakeVerticalTable()
  for i := 0; i < 10; i++ {
    vtable.AddChild(MakeColorBoxWidget(250, 250, 1, 1, 1, 1))
    vtable.AddChild(MakeColorBoxWidget(250, 250, 1, 0, 0, 1))
    vtable.AddChild(MakeColorBoxWidget(250, 250, 0, 1, 0, 1))
    vtable.AddChild(MakeColorBoxWidget(250, 250, 0, 0, 1, 1))
  }
  scroll := gui.MakeScrollFrame(vtable, 300, 100)
  v2 := gui.MakeVerticalTable()
  v2.AddChild(scroll)
  ui.AddChild(v2)

  v2.AddChild(gui.MakeFileChooser("/Users/runningwild/", func(w gui.Widget, n string, err error) {
    if err == nil {
      fmt.Printf("Got %s\n", n)
    } else {
      fmt.Printf("Err: %s\n", err.Error())
    }
    v2.RemoveChild(w)
  }))

  for gin.In().GetKey('q').FramePressCount() == 0 {
    ui.Draw()
    sys.SwapBuffers()
    sys.Think()
    gl.ClearColor(0, 0, 0, 1)
    gl.Clear(gl.COLOR_BUFFER_BIT)
  }
}
