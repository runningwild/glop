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
func (r *rootWidget) Draw(dims Dims) {
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
  g.Root.handleEventGroup(event_group)
}
func (g *Gui) Think(ms int64) {
  g.Root.think(ms, g.focus)
  region := Region{
    Dims : g.Root.widget.(*rootWidget).Dims,
  }
  gl.MatrixMode(gl.MODELVIEW)
  gl.LoadIdentity();
  gl.MatrixMode(gl.PROJECTION)
  gl.LoadIdentity();

  region.setViewport()
  gl.ClearColor(0, 0, 0, 1)
  gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
  g.Root.layoutAndDraw(region)
  region.setViewport()
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
func (t *VerticalTable) Draw(dims Dims) {
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
func (b *BoxWidget) Draw(dims Dims) {
  gl.MatrixMode(gl.MODELVIEW)
  gl.PushMatrix()
  defer gl.PopMatrix()
  gl.LoadIdentity()
  gl.Ortho(0,1,0,1,-1,1)

  gl.Color4d(b.R, b.G, b.B, b.A)
  gl.Disable(gl.TEXTURE_2D)
  gl.Begin(gl.QUADS)
    gl.Vertex2d(0,0)
    gl.Vertex2d(1,0)
    gl.Vertex2d(1,1)
    gl.Vertex2d(0,1)
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
    xoff := int(anchor.Bx * float64(dims.Dx) - anchor.Wx * float64(requested[widget].Dx) + 0.5)
    yoff := int(anchor.By * float64(dims.Dy) - anchor.Wy * float64(requested[widget].Dy) + 0.5)
    dims := requested[widget]
    if xoff < 0 {
      dims.Dx += xoff
      xoff = 0
    }
    if yoff < 0 {
      dims.Dy += yoff
      yoff = 0
    }
    if xoff + dims.Dx > w.Dims.Dx {
      dims.Dx -= (xoff + dims.Dx) - w.Dims.Dx
    }
    if yoff + dims.Dy > w.Dims.Dy {
      dims.Dy -= (yoff + dims.Dy) - w.Dims.Dy
    }
    reg[widget] = Region{
      Point : Point{ X : xoff, Y : yoff },
      Dims : dims,
    }
  }
  return reg
}
func (w *AnchorBox) Draw(dims Dims) {
}
