package gui

import(
  "glop/gin"
  "gl"
)

type Point struct {
  X,Y int
}
func (p Point) Add(q Point) Point {
  return Point{
    X : p.X + q.X,
    Y : p.Y + q.Y,
  }
}
func (p Point) Inside(r Region) bool {
  if p.X < r.X { return false }
  if p.Y < r.Y { return false }
  if p.X > r.X + r.Dx { return false }
  if p.Y > r.Y + r.Dy { return false }
  return true
}
type Dims struct {
  Dx,Dy int
}
type Region struct {
  Point
  Dims
}
func (r Region) Add(p Point) Region {
  return Region{
    r.Point.Add(p),
    r.Dims,
  }
}

// Need a global stack of regions because opengl only handles pushing/popping
// the state of the enable bits for each clip plane, not the planes themselves
var clippers []Region
func (r Region) setClipPlanes() {
  var eqs [][4]float64
  eqs = append(eqs, [4]float64{ 1, 0, 0, -float64(r.X)})
  eqs = append(eqs, [4]float64{-1, 0, 0, float64(r.X + r.Dx)})
  eqs = append(eqs, [4]float64{ 0, 1, 0, -float64(r.Y)})
  eqs = append(eqs, [4]float64{ 0,-1, 0, float64(r.Y + r.Dy)})
  gl.ClipPlane(gl.CLIP_PLANE0, &eqs[0][0])
  gl.ClipPlane(gl.CLIP_PLANE1, &eqs[1][0])
  gl.ClipPlane(gl.CLIP_PLANE2, &eqs[2][0])
  gl.ClipPlane(gl.CLIP_PLANE3, &eqs[3][0])
}
func (r Region) PushClipPlanes() {
  if len(clippers) == 0 {
    gl.Enable(gl.CLIP_PLANE0)
    gl.Enable(gl.CLIP_PLANE1)
    gl.Enable(gl.CLIP_PLANE2)
    gl.Enable(gl.CLIP_PLANE3)
  }
  r.setClipPlanes()
  clippers = append(clippers, r)
}
func (r Region) PopClipPlanes() {
  clippers = clippers[0 : len(clippers) - 1]
  if len(clippers) == 0 {
    gl.Disable(gl.CLIP_PLANE0)
    gl.Disable(gl.CLIP_PLANE1)
    gl.Disable(gl.CLIP_PLANE2)
    gl.Disable(gl.CLIP_PLANE3)
  } else {
    clippers[len(clippers) - 1].setClipPlanes()
  }
}


//func (r Region) setViewport() {
//  gl.Viewport(r.Point.X, r.Point.Y, r.Dims.Dx, r.Dims.Dy)
//}

type Zone interface {
  // Returns the dimensions that this Widget would like available to
  // render itself.  A Widget should only update the value it returns from
  // this method when its Think() method is called.
  Requested() Dims

  // Returns ex,ey, where ex and ey indicate whether this Widget is
  // capable of expanding along the X and Y axes, respectively.
  Expandable() (bool,bool)

  // Returns the region that this Widget used to render itself the last
  // time it was rendered.  Should be completely contained within the
  // region that was passed to it on its last call to Render.
  Rendered() Region
}

type EventGroup struct {
  gin.EventGroup
  Focus bool
}

type Widget interface {
  Zone
  Think(int64)

  // Returns true if this widget or any of its children consumed the
  // event group
  Respond(*Gui,EventGroup) bool

  Draw(Region)
}
type CoreWidget interface {
  DoThink(int64)

  // If take_focus is true, then the EventGroup will be consumed,
  // regardless of the value of consume
  DoRespond(EventGroup) (consume,take_focus bool)
  Zone

  Draw(Region)
  GetChildren() []Widget
}
type EmbeddedWidget interface {
  Think(int64)
  Respond(*Gui, EventGroup) (consume bool)
}
type BasicWidget struct {
  CoreWidget
}
func (w *BasicWidget) Think(t int64) {
  kids := w.GetChildren()
  for i := range kids {
    kids[i].Think(t)
  }
  w.DoThink(t)
}
func (w *BasicWidget) Respond(gui *Gui, event_group EventGroup) bool {
  cursor := event_group.Events[0].Key.Cursor()
  if cursor != nil {
    var p Point
    p.X, p.Y = cursor.Point()
    if !p.Inside(w.Rendered()) {
      return false
    }
  }
  consume,take_focus := w.DoRespond(event_group)
  if take_focus {
    gui.TakeFocus(w)
  }
  if take_focus || consume { return true }
  kids := w.GetChildren()
  for i := range kids {
    if kids[i].Respond(gui, event_group) { return true }
  }
  return false
}

type BasicZone struct {
  Request_dims  Dims
  Render_region Region
  Ex,Ey         bool
}

func (bz *BasicZone) Requested() Dims {
  return bz.Request_dims
}
func (bz *BasicZone) Rendered() Region {
  return bz.Render_region
}
func (bz *BasicZone) Expandable() (bool,bool) {
  return bz.Ex, bz.Ey
}

type NonThinker struct {}
func (n NonThinker) DoThink(int64) {}

type NonResponder struct {}
func (n NonResponder) DoRespond(EventGroup) (bool,bool) {
  return false,false
}

type Childless struct {}
func (c Childless) GetChildren() []Widget { return nil }

type StandardParent struct {
  Children []Widget
}
func (s *StandardParent) GetChildren() []Widget {
  return s.Children
}
func (s *StandardParent) AddChild(w Widget) {
  s.Children = append(s.Children, w)
}
func (s *StandardParent) RemoveChild(w Widget) {
  for i := range s.Children {
    if s.Children[i] == w {
      s.Children[i] = s.Children[len(s.Children)-1]
      s.Children = s.Children[0 : len(s.Children)-1]
      return
    }
  }
}


type rootWidget struct {
  EmbeddedWidget
  StandardParent
  BasicZone
  NonResponder
  NonThinker
}

func (r *rootWidget) Draw(region Region) {
  r.Render_region = region
  for i := range r.Children {
    r.Children[i].Draw(region)
  }
}

type Gui struct {
  root  rootWidget

  // Stack of widgets that have focus
  focus []Widget
}

func Make(dispatcher gin.EventDispatcher, dims Dims) *Gui {
  var g Gui
  g.root.EmbeddedWidget = &BasicWidget{ CoreWidget : &g.root }
  g.root.Request_dims = dims
  g.root.Render_region.Dims = dims
  dispatcher.RegisterEventListener(&g)
  return &g
}

func (g *Gui) Draw() {
  gl.MatrixMode(gl.PROJECTION)
  gl.LoadIdentity();
  region := g.root.Render_region
  gl.Ortho(float64(region.X), float64(region.X + region.Dx), float64(region.Y), float64(region.Y + region.Dy), 1000, -1000)
  gl.ClearColor(0, 0, 0, 1)
  gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
  gl.MatrixMode(gl.MODELVIEW)
  gl.LoadIdentity();
  g.root.Draw(region)
}

// TODO: Shouldn't be exposing this
func (g *Gui) Think(t int64) {
  g.root.Think(t)
}

// TODO: Shouldn't be exposing this
func (g *Gui) HandleEventGroup(gin_group gin.EventGroup) {
  event_group := EventGroup{gin_group, false}
  if len(g.focus) > 0 {
    event_group.Focus = true
    consume := g.focus[len(g.focus)-1].Respond(g, event_group)
    if consume { return }
    event_group.Focus = false
  }
  g.root.Respond(g, event_group)
}

func (g *Gui) AddChild(w Widget) {
  g.root.AddChild(w)
}

func (g *Gui) RemoveChild(w Widget) {
  g.root.RemoveChild(w)
}

func (g *Gui) TakeFocus(w Widget) {
  if len(g.focus) == 0 {
    g.focus = append(g.focus, nil)
  }
  g.focus[len(g.focus)-1] = w
}
