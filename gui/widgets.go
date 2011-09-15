package gui

import (
  "glop/gin"
  "gl"
)

// This widget is actually superfluous, it's functionality can just be folded into Gui
type rootWidget struct {
  Stoic
  Unthinking
  StandardParent
  dx,dy    int
}
func (r *rootWidget) Draw(x,y,dx,dy int) {
  gl.LoadIdentity();
  gl.Ortho(0, float64(dx), 0, float64(dy), -1, 1)
  gl.MatrixMode(gl.MODELVIEW)
  gl.ClearColor(0, 0, 0, 1)
  gl.Clear(0x00004000)
  for _,child := range r.children {
    cdx,cdy := child.Size()
    child.Draw(x, y, cdx, cdy)  // Let them do whatever they want
  }
}
func (r *rootWidget) Size() (int,int) {
  return r.dx,r.dy
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
    widget : &rootWidget{dx : dx, dy : dy },
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
}

func (g *Gui) Draw() {
  dx,dy := g.Root.widget.Size()
  g.Root.widget.Draw(0, 0, dx, dy)
}

type Childless struct {}
func (c Childless)  AddChild(_ Widget, _ interface{}) {}
func (c Childless)  RemoveChild(_ Widget) {}

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

type Unthinking struct {}
func (u Unthinking) Think(_ int64, _ bool) bool {
  return false
}

type VerticalTable struct {
  Stoic
  StandardParent
  Unthinking
  dx,dy int
}
func (t *VerticalTable) Think(ms int64, has_focus bool) bool {
  t.dx = 0
  t.dy = 0
  for _,child := range t.children {
    dx,dy := child.Size()
    if dx > t.dx {
      t.dx = dx
    }
    t.dy += dy
  }
  return false
}
func (t *VerticalTable) Size() (int,int) {
  return t.dx, t.dy
}
func (t *VerticalTable) Draw(x,y,dx,dy int) {
  pos := y
  for _,child := range t.children {
    cdx,cdy := child.Size()
    if cdx > dx {
      cdx = dx
    }
    if cdy > dy {
      cdy = dy
    }
    child.Draw(x, pos, cdx, cdy)
    pos += cdy
  }
}

type BoxWidget struct {
  Childless
  Stoic
  Unthinking
  dx,dy int
  r,g,b,a float64
}
func MakeBoxWidget(dx,dy int, r,g,b,a float64) *BoxWidget {
  return &BoxWidget{ dx:dx, dy:dy, r:r, g:g, b:b, a:a }
}
func (b *BoxWidget) Size() (int,int) {
  return b.dx, b.dy
}
func (b *BoxWidget) Draw(x,y,dx,dy int) {
  gl.Color4d(b.r, b.g, b.b, b.a)
  fx := float64(x)
  fy := float64(y)
  fdx := float64(dx)
  fdy := float64(dy)
  gl.Begin(gl.QUADS)
    gl.Vertex2d(    fx,    fy)
    gl.Vertex2d(fdx+fx,    fy)
    gl.Vertex2d(fdx+fx,fdy+fy)
    gl.Vertex2d(    fx,fdy+fy)
  gl.End()
}

