package gui

type tableBase struct {
  children []Widget
  fixed    []bool
}
func (w *tableBase) GetChildren() []Widget {
  return w.children
}
func (w *tableBase) AddChild(widget Widget, fixed bool) {
  w.children = append(w.children, widget)
  w.fixed = append(w.fixed, fixed)
}

type VerticalTable struct {
  EmbeddedWidget
  NonResponder
  Rectangle
  tableBase
}

func MakeVerticalTable() *VerticalTable {
  var table VerticalTable
  table.EmbeddedWidget = &BasicWidget{ CoreWidget : &table }
  return &table
}
func (w *VerticalTable) DoThink(t int64) {
  w.Dims = Dims{}
  for i := range w.children {
    if w.children[i].Bounds().Dx > w.Dims.Dx {
      w.Dims.Dx = w.children[i].Bounds().Dx
    }
    if w.fixed[i] {
      w.Dims.Dy += w.children[i].Bounds().Dy
    }
  }
}
func (w *VerticalTable) Draw(region Region) {
  fixed_available := region.Dy
  fixed_requested := w.Dy
  fill_available := fixed_available - fixed_requested
  if fixed_available > fixed_requested {
    fixed_available = fixed_requested
  }
  if fill_available < 0 {
    fill_available = 0
  }
  fill_requested := 0
  for i := range w.children {
    if !w.fixed[i] {
      fill_requested += w.children[i].Bounds().Dy
    }
  }
  y := region.Y
  for i := range w.children {
    req := w.children[i].Bounds()
    if w.fixed[i] {
      req.Dy = (req.Dy * fixed_available) / fixed_requested
    } else {
      req.Dy = (req.Dy * fill_available) / fill_requested
    }
    if req.Dx > region.Dx {
      req.Dx = region.Dx
    }
    req.X = region.X
    req.Y = y
    w.children[i].Draw(req)
    y += req.Dy
  }
  w.Rectangle.Dims = region.Dims
  w.Rectangle.Point = region.Point
}



type HorizontalTable struct {
  EmbeddedWidget
  NonResponder
  Rectangle
  tableBase
}

func MakeHorizontalTable() *HorizontalTable {
  var table HorizontalTable
  table.EmbeddedWidget = &BasicWidget{ CoreWidget : &table }
  return &table
}
func (w *HorizontalTable) DoThink(t int64) {
  w.Dims = Dims{}
  for i := range w.children {
    if w.children[i].Bounds().Dy > w.Dims.Dy {
      w.Dims.Dy = w.children[i].Bounds().Dy
    }
    if w.fixed[i] {
      w.Dims.Dx += w.children[i].Bounds().Dx
    }
  }
}
func (w *HorizontalTable) Draw(region Region) {
  fixed_available := region.Dx
  fixed_requested := w.Dx
  fill_available := fixed_available - fixed_requested
  if fixed_available > fixed_requested {
    fixed_available = fixed_requested
  }
  if fill_available < 0 {
    fill_available = 0
  }
  fill_requested := 0
  for i := range w.children {
    if !w.fixed[i] {
      fill_requested += w.children[i].Bounds().Dx
    }
  }
  x := region.X
  for i := range w.children {
    req := w.children[i].Bounds()
    if w.fixed[i] {
      req.Dx = (req.Dx * fixed_available) / fixed_requested
    } else {
      req.Dx = (req.Dx * fill_available) / fill_requested
    }
    if req.Dy > region.Dy {
      req.Dy = region.Dy
    }
    req.Y = region.Y
    req.X = x
    w.children[i].Draw(req)
    x += req.Dx
  }
  w.Rectangle.Dims = region.Dims
  w.Rectangle.Point = region.Point
}

