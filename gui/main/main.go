package main

import (
  "gl"
  "glop/gos"
  "glop/gin"
  "glop/gui"
  "glop/system"
  "runtime"
  "glop/render"
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
}

func main() {
  sys = system.Make(gos.GetSystemInterface())
  sys.Startup()
  wdx := 800
  wdy := 600
  render.Init()
  var ui *gui.Gui
  render.Queue(func() {
    sys.CreateWindow(10, 10, wdx, wdy)
    sys.EnableVSync(true)
    ui,_ = gui.Make(gin.In(), gui.Dims{ wdx, wdy }, "/Users/runningwild/code/go-glop/example/data/fonts/skia.ttf")
  })
  render.Purge()

  var ws []gui.Widget
  options := []string{ "foo", "ding", "bar" }
  target := make(map[string]bool)
  target["foo"] = true
  cb :=  gui.MakeCheckTextBox(options, 300, target)

  tabs := gui.MakeTabFrame(ws)

  ui.AddChild(cb)
//  ui.AddChild(v2)
  count := 0
  for gin.In().GetKey('q').FramePressCount() == 0 {
    count++
    if count % 300 == 0 {
      delete(target, "foo")
    }
    if count % 300 == 150 {
      target["bar"] = false
    }
    render.Queue(func() {
      gl.ClearColor(0, 0, 0, 1)
      gl.Clear(gl.COLOR_BUFFER_BIT)
      ui.Draw()
    })
    render.Purge()
    sys.SwapBuffers()
    sys.Think()
    if gin.In().GetKey('1').FramePressCount() > 0 {
      tabs.SelectTab(0)
    }
    if gin.In().GetKey('2').FramePressCount() > 0 {
      tabs.SelectTab(1)
    }
    if gin.In().GetKey('3').FramePressCount() > 0 {
      tabs.SelectTab(2)
    }
    if gin.In().GetKey('4').FramePressCount() > 0 {
      tabs.SelectTab(3)
    }
  }
}
