package gui

import(
  "glop/gin"
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
func (s *StandardParent) AddWidget(w Widget) {
  s.Children = append(s.Children, w)
}
func (s *StandardParent) RemoveWidget(w Widget) {
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
