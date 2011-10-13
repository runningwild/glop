package gui


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
  EmbeddedWidget
  NonResponder
  NonThinker
  Rectangle
  children []Widget
  anchors  []Anchor
}
func MakeAnchorBox(dims Dims) *AnchorBox {
  var box AnchorBox
  box.EmbeddedWidget = &BasicWidget{ CoreWidget : &box }
  box.Dims = dims
  return &box
}
func (w *AnchorBox) AddChild(widget Widget, anchor Anchor) {
  w.children = append(w.children, widget)
  w.anchors = append(w.anchors, anchor)
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
func (w *AnchorBox) GetChildren() []Widget {
  return w.children
}
func (w *AnchorBox) Draw(region Region) {
  for i := range w.children {
    widget := w.children[i]
    anchor := w.anchors[i]
    child_dims := widget.Bounds()
    xoff := int(anchor.Bx * float64(region.Dx) - anchor.Wx * float64(child_dims.Dx) + 0.5)
    yoff := int(anchor.By * float64(region.Dy) - anchor.Wy * float64(child_dims.Dy) + 0.5)
    if xoff < 0 {
      child_dims.Dx += xoff
      xoff = 0
    }
    if yoff < 0 {
      child_dims.Dy += yoff
      yoff = 0
    }
    if xoff + child_dims.Dx > w.Dims.Dx {
      child_dims.Dx -= (xoff + child_dims.Dx) - w.Dims.Dx
    }
    if yoff + child_dims.Dy > w.Dims.Dy {
      child_dims.Dy -= (yoff + child_dims.Dy) - w.Dims.Dy
    }
    child_dims.X = xoff
    child_dims.Y = yoff
    widget.Draw(child_dims)
  }
}
