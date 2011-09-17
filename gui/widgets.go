package gui

import (
  "glop/gin"
  "gl"
)

// What widgets do we need
// SingleLineText
// ParagraphText
// Tables
// RenderFrame
// Basic form items:
//   Text Entry
//   Radio Buttons
//   Check Box
//   COMBO BOX!?

// This widget is actually superfluous, it's functionality can just be folded into Gui
type rootWidget struct {
  Stoic
  Unthinking
  StandardParent
}
func (r *rootWidget) Draw(region Region) {
  gl.LoadIdentity();
  gl.Ortho(0, float64(r.Dims.Dx), 0, float64(r.Dims.Dy), -1, 1)
  gl.MatrixMode(gl.MODELVIEW)
  gl.ClearColor(0, 0, 0, 1)
  gl.Clear(0x00004000)
}
func (r *rootWidget) Layout(dims Dims, req map[Widget]Dims) map[Widget]Region {
  m := make(map[Widget]Region)
  for _,child := range r.children {
    m[child] = Region{ Dims : req[child] }
  }
  return m
}

type Gui struct {
  Root  *Node
  focus *Focus
}
// Creates the gui object and specifies the size of the screen that it will render to.  The gui
// naturally renders to the rectangular region from (0,0) to (dx,dy)
func Make(input *gin.Input, dx,dy int) *Gui {
  g := new(Gui)
  g.focus = new(Focus)
  input.RegisterEventListener(g)
  g.Root = &Node{
    widget : &rootWidget{ Unthinking : Unthinking{ Dims : Dims{ Dx : dx, Dy : dy } } },
    parent : nil,
    children : nil,
  }
  return g
}

// This method shouldn't be exported, perhaps we can make it a method on a private
// member variable
func (g *Gui) HandleEventGroup(event_group gin.EventGroup) {
  if g.focus.top() == nil {
    return
  }
  path := g.focus.top().pathFromRoot()
  for _,p := range path {
    consume,give,target := p.widget.HandleEventGroup(event_group)
    if give {
      if target == nil {
        g.focus.Pop()
      } else {
        g.focus.Take(target)
      }
      return
    }
    if consume {
      return
    }
  }
}
func (g *Gui) Think(ms int64) {
  g.Root.think(ms, g.focus)
  region := Region{
    Dims : g.Root.widget.(*rootWidget).Dims,
  }
  g.Root.layoutAndDraw(region)
}

type Childless struct {}
func (c Childless) AddChild(_ Widget, _ interface{}) {}
func (c Childless) RemoveChild(_ Widget) {}
func (c Childless) Layout(_ Dims, _ map[Widget]Dims) map[Widget]Region { return nil }

type StandardParent struct {
  children []Widget
}
func (sp *StandardParent) AddChild(widget Widget, _ interface{}) {
  sp.children = append(sp.children, widget)
}
func (sp *StandardParent) RemoveChild(widget Widget) {
  for i := range sp.children {
    if sp.children[i] == widget {
      sp.children[i] = sp.children[len(sp.children)-1]
      sp.children = sp.children[0 : len(sp.children)-1]
      return
    }
  }
}


type Stoic struct {}
func (s Stoic) HandleEventGroup(_ gin.EventGroup) (bool, bool, *Node) {
  return false, false, nil
}

type Unthinking struct {
  Dims Dims
}
func (u Unthinking) Think(_ int64, _ bool, _ Region, _ map[Widget]Dims) (bool,Dims) {
  return false, u.Dims
}

type VerticalTable struct {
  Stoic
  StandardParent
}
func (t *VerticalTable) Think(_ int64, _ bool, _ Region, child_dims map[Widget]Dims) (bool, Dims) {
  var dims Dims
  for _,child_dim := range child_dims {
    if child_dim.Dx > dims.Dx {
      dims.Dx = child_dim.Dx
    }
    dims.Dy += child_dim.Dy
  }
  return false, dims
}
func (t *VerticalTable) Layout(dims Dims, requested map[Widget]Dims) map[Widget]Region {
  reg := make(map[Widget]Region)
  var cur Point
  for _,widget := range t.children {
    reg[widget] = Region{
      Point : cur,
      Dims : requested[widget],
    }
    cur.Y += requested[widget].Dy
  }
  return reg
}
func (t *VerticalTable) Draw(region Region) {
}

type BoxWidget struct {
  Childless
  Stoic
  Unthinking
  R,G,B,A float64
}
func MakeBoxWidget(dx,dy int, r,g,b,a float64) *BoxWidget {
  return &BoxWidget{ Unthinking : Unthinking{ Dims : Dims{ Dx : dx, Dy : dy }}, R:r, G:g, B:b, A:a }
}
func (b *BoxWidget) Draw(region Region) {
  gl.Color4d(b.R, b.G, b.B, b.A)
  fx := float64(region.X)
  fy := float64(region.Y)
  fdx := float64(region.Dx)
  fdy := float64(region.Dy)
  gl.Begin(gl.QUADS)
    gl.Vertex2d(    fx,    fy)
    gl.Vertex2d(fdx+fx,    fy)
    gl.Vertex2d(fdx+fx,fdy+fy)
    gl.Vertex2d(    fx,fdy+fy)
  gl.End()
}

// An Anchor specifies where a widget should be positioned withing an AnchorBox
// All values are between 0 and 1, inclusive.  wx,wy represent a point on the widget,
// and bx,by represent a point on the AnchorBox.  During layout the widget will be positioned
// such that these points line up.
type Anchor struct {
  Wx,Wy,Bx,By float64
}

// An AnchorBox does layout according to Anchors.  An anchor must be specified when installing
// a widget.
type AnchorBox struct {
  Stoic
  Unthinking
  children []Widget
  anchors  []Anchor
}
func MakeAnchorBox(dims Dims) *AnchorBox {
  w := AnchorBox{
    Unthinking : Unthinking{ Dims : dims },
  }
  return &w
}
func (w *AnchorBox) AddChild(widget Widget, _anchor interface{}) {
  w.children = append(w.children, widget)
  w.anchors = append(w.anchors, _anchor.(Anchor))
}
func (w *AnchorBox) RemoveChild(widget Widget) {
  for i := range w.children {
    if w.children[i] == widget {
      w.children[i] = w.children[len(w.children)-1]
      w.children = w.children[0 : len(w.children)-1]
      w.anchors[i] = w.anchors[len(w.anchors)-1]
      w.anchors = w.anchors[0 : len(w.anchors)-1]
      return
    }
  }
}
func (w *AnchorBox) Layout(dims Dims, requested map[Widget]Dims) map[Widget]Region {
  reg := make(map[Widget]Region)
  for i := range w.children {
    widget := w.children[i]
    anchor := w.anchors[i]
    xoff := anchor.Bx * float64(dims.Dx) - anchor.Wx * float64(requested[widget].Dx)
    yoff := anchor.By * float64(dims.Dy) - anchor.Wy * float64(requested[widget].Dy)
    reg[widget] = Region{
      Point : Point{ X : int(xoff), Y : int(yoff) },
      Dims : requested[widget],
    }
  }
  return reg
}
func (w *AnchorBox) Draw(region Region) {
}