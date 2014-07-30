package gui

// // An Anchor specifies where a widget should be positioned withing an AnchorBox
// // All values are between 0 and 1, inclusive.  wx,wy represent a point on the widget,
// // and bx,by represent a point on the AnchorBox.  During layout the widget will be positioned
// // such that these points line up.
// type Anchor struct {
// 	Wx, Wy, Bx, By float64
// }

// // An AnchorBox does layout according to Anchors.  An anchor must be specified when installing
// // a widget.
// type AnchorBox struct {
// 	EmbeddedWidget
// 	NonResponder
// 	NonThinker
// 	NonFocuser
// 	BasicZone
// 	children []Widget
// 	anchors  []Anchor
// }

// func MakeAnchorBox(dims Dims) *AnchorBox {
// 	var box AnchorBox
// 	box.EmbeddedWidget = &BasicWidget{CoreWidget: &box}
// 	box.Request_dims = dims
// 	return &box
// }
// func (w *AnchorBox) String() string {
// 	return "anchor box"
// }
// func (w *AnchorBox) AddChild(widget Widget, anchor Anchor) {
// 	w.children = append(w.children, widget)
// 	w.anchors = append(w.anchors, anchor)
// }
// func (w *AnchorBox) RemoveChild(widget Widget) {
// 	for i := range w.children {
// 		if w.children[i] == widget {
// 			w.children[i] = w.children[len(w.children)-1]
// 			w.children = w.children[0 : len(w.children)-1]
// 			w.anchors[i] = w.anchors[len(w.anchors)-1]
// 			w.anchors = w.anchors[0 : len(w.anchors)-1]
// 			return
// 		}
// 	}
// }
// func (w *AnchorBox) GetChildren() []Widget {
// 	return w.children
// }
// func (w *AnchorBox) Draw(region Region) {
// 	w.Render_region = region
// 	for i := range w.children {
// 		widget := w.children[i]
// 		anchor := w.anchors[i]
// 		var child_region Region
// 		child_region.Dims = widget.Requested()
// 		xoff := int(anchor.Bx*float64(region.Dx) - anchor.Wx*float64(child_region.Dx) + 0.5)
// 		yoff := int(anchor.By*float64(region.Dy) - anchor.Wy*float64(child_region.Dy) + 0.5)
// 		if xoff < 0 {
// 			child_region.Dx += xoff
// 			xoff = 0
// 		}
// 		if yoff < 0 {
// 			child_region.Dy += yoff
// 			yoff = 0
// 		}
// 		if xoff+child_region.Dx > w.Render_region.Dx {
// 			child_region.Dx -= (xoff + child_region.Dx) - w.Render_region.Dx
// 		}
// 		if yoff+child_region.Dy > w.Render_region.Dy {
// 			child_region.Dy -= (yoff + child_region.Dy) - w.Render_region.Dy
// 		}
// 		child_region.X = xoff
// 		child_region.Y = yoff
// 		ex, ey := widget.Expandable()
// 		if ex {
// 			child_region.X = region.X
// 			child_region.Dx = region.Dx
// 		}
// 		if ey {
// 			child_region.Y = region.Y
// 			child_region.Dy = region.Dy
// 		}
// 		widget.Draw(child_region)
// 	}
// }
