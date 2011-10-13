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
  Bounds() Region
  Contains(Point) bool
}

type EventGroup struct {
  gin.EventGroup
  Focus bool
}

type Widget interface {
  Zone
  Think(int64)
  Respond(EventGroup)
  Draw(Region)
}
type CoreWidget interface {
  DoThink(int64)
  DoRespond(EventGroup) bool
  Zone

  Draw(Region)
  GetChildren() []Widget
}
type EmbeddedWidget interface {
  Think(int64)
  Respond(EventGroup)
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
func (w *BasicWidget) Respond(event_group EventGroup) {
  cursor := event_group.Events[0].Key.Cursor()
  if cursor != nil {
    var p Point
    p.X, p.Y = cursor.Point()
    if !w.Contains(p) {
      return
    }
  }
  if w.DoRespond(event_group) { return }
  kids := w.GetChildren()
  for i := range kids {
    kids[i].Respond(event_group)
  }
}

type Rectangle Region
func (r *Rectangle) Bounds() Region {
  return Region(*r)
}
func (r *Rectangle) Contains(p Point) bool {
  return p.Inside(Region(*r))
}
func (r *Rectangle) Constrain(r2 Region) {
  if r2.Dx < r.Dx {
    r.Dx = r2.Dx
  }
  if r2.Dy < r.Dy {
    r.Dy = r2.Dy
  }
  r.X = r2.X
  r.Y = r2.Y
}

type NonThinker struct {}
func (n NonThinker) DoThink(int64) {}

type NonResponder struct {}
func (n NonResponder) DoRespond(EventGroup) bool {
  return false
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
  Rectangle
  NonResponder
  NonThinker
}

func (r *rootWidget) Draw(region Region) {
  for i := range r.Children {
    r.Children[i].Draw(region)
  }
}

type Gui struct {
  root rootWidget
}

func Make(dispatcher gin.EventDispatcher, dims Dims) *Gui {
  var g Gui
  g.root.EmbeddedWidget = &BasicWidget{ CoreWidget : &g.root }
  g.root.Rectangle = Rectangle{ Dims : dims }
  dispatcher.RegisterEventListener(&g)
  return &g
}

func (g *Gui) Draw() {
  gl.MatrixMode(gl.PROJECTION)
  gl.LoadIdentity();
  gl.Ortho(float64(g.root.X), float64(g.root.X + g.root.Dx), float64(g.root.Y), float64(g.root.Y + g.root.Dy), 1000, -1000)
  gl.ClearColor(0, 0, 0, 1)
  gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
  gl.MatrixMode(gl.MODELVIEW)
  gl.LoadIdentity();
  g.root.Draw(g.root.Bounds())
}

// TODO: Shouldn't be exposing this
func (g *Gui) Think(t int64) {
  g.root.Think(t)
}

// TODO: Shouldn't be exposing this
func (g *Gui) HandleEventGroup(gin_group gin.EventGroup) {
  g.root.Respond(EventGroup{gin_group, false})
}

func (g *Gui) AddChild(w Widget) {
  g.root.AddChild(w)
}

func (g *Gui) RemoveChild(w Widget) {
  g.root.RemoveChild(w)
}


