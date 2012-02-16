package gui

import "github.com/runningwild/opengl/gl"

type TableParams struct {
  Spacing int
  Background struct {
    R,G,B,A float64
  }
  Border struct {
    R,G,B,A float64
  }
}

func defaultParams() TableParams {
  var p TableParams
  p.Background.A = 0.7
  p.Border.R = 1
  p.Border.G = 1
  p.Border.B = 1
  p.Border.A = 0.5
  return p
}

type Table interface {
  Widget
  GetChildren() []Widget
  AddChild(w Widget)
  RemoveChild(w Widget)
  RemoveAllChildren()
  Params() *TableParams
}

type VerticalTable struct {
  EmbeddedWidget
  NonResponder
  NonFocuser
  BasicZone
  StandardParent
  params TableParams
}

func MakeVerticalTable() *VerticalTable {
  var table VerticalTable
  table.EmbeddedWidget = &BasicWidget{CoreWidget: &table}
  table.params = defaultParams()
  return &table
}
func (w *VerticalTable) Params() *TableParams {
  return &w.params
}
func (w *VerticalTable) String() string {
  return "vertical table"
}
func (w *VerticalTable) DoThink(int64, bool) {
  w.Request_dims = Dims{}
  w.Ex = false
  w.Ey = false
  for _, child := range w.Children {
    ex, ey := child.Expandable()
    if ex {
      w.Ex = true
    }
    if ey {
      w.Ey = true
    }
    w.Request_dims.Dy += child.Requested().Dy
    if child.Requested().Dx > w.Request_dims.Dx {
      w.Request_dims.Dx = child.Requested().Dx
    }
  }
  w.Request_dims.Dy += w.params.Spacing * len(w.Children)
}
func (w *VerticalTable) Draw(region Region) {
  gl.Enable(gl.BLEND)
  gl.Disable(gl.TEXTURE_2D)
  dx := region.Dx
  if dx > w.Request_dims.Dx && !w.Ex {
    dx = w.Request_dims.Dx
  }
  dy := region.Dy
  if dy > w.Request_dims.Dy && !w.Ex {
    dy = w.Request_dims.Dy
  }
  gl.Color4d(
      w.params.Background.R,
      w.params.Background.G,
      w.params.Background.B,
      w.params.Background.A)
  gl.Begin(gl.QUADS)
  gl.Vertex2i(region.X, region.Y+region.Dy-dy)
  gl.Vertex2i(region.X, region.Y+region.Dy)
  gl.Vertex2i(region.X+dx, region.Y+region.Dy)
  gl.Vertex2i(region.X+dx, region.Y+region.Dy-dy)
  gl.End()
  gl.Color4d(
      w.params.Border.R,
      w.params.Border.G,
      w.params.Border.B,
      w.params.Border.A)
  gl.Begin(gl.LINES)
  gl.Vertex2i(region.X, region.Y+region.Dy-dy)
  gl.Vertex2i(region.X, region.Y+region.Dy)

  gl.Vertex2i(region.X, region.Y+region.Dy)
  gl.Vertex2i(region.X+dx, region.Y+region.Dy)

  gl.Vertex2i(region.X+dx, region.Y+region.Dy)
  gl.Vertex2i(region.X+dx, region.Y+region.Dy-dy)

  gl.Vertex2i(region.X+dx, region.Y+region.Dy-dy)
  gl.Vertex2i(region.X, region.Y+region.Dy-dy)
  gl.End()

  fill_available := region.Dy - w.Request_dims.Dy
  if fill_available < 0 {
    fill_available = 0
  }
  fill_request := 0
  for _, child := range w.Children {
    if _, ey := child.Expandable(); ey {
      fill_request += child.Requested().Dy
    }
  }
  var child_region Region
  child_region.Y = region.Y + region.Dy
  for _, child := range w.Children {
    child_region.Dims = child.Requested()
    if _, ey := child.Expandable(); ey && fill_request > 0 {
      child_region.Dy += (child_region.Dy * fill_available) / fill_request
    }
    if region.Dy < w.Request_dims.Dy {
      child_region.Dims.Dy *= region.Dy
      child_region.Dims.Dy /= w.Request_dims.Dy
    }
    if child_region.Dx > region.Dx {
      child_region.Dx = region.Dx
    }
    if ex, _ := child.Expandable(); child_region.Dx < region.Dx && ex {
      child_region.Dx = region.Dx
    }
    child_region.X = region.X
    child_region.Y -= child_region.Dy
    child_region.Y -= w.params.Spacing
    child.Draw(child_region)
  }
  w.Render_region = region
}

type HorizontalTable struct {
  EmbeddedWidget
  NonResponder
  NonFocuser
  BasicZone
  StandardParent
  params TableParams
}

func MakeHorizontalTable() *HorizontalTable {
  var table HorizontalTable
  table.EmbeddedWidget = &BasicWidget{CoreWidget: &table}
  table.params = defaultParams()
  return &table
}
func (w *HorizontalTable) Params() *TableParams {
  return &w.params
}
func (w *HorizontalTable) String() string {
  return "horizontal table"
}
func (w *HorizontalTable) DoThink(int64, bool) {
  w.Request_dims = Dims{}
  w.Ex = false
  w.Ey = false
  for _, child := range w.Children {
    ex, ey := child.Expandable()
    if ex {
      w.Ex = true
    }
    if ey {
      w.Ey = true
    }
    w.Request_dims.Dx += child.Requested().Dx
    if child.Requested().Dy > w.Request_dims.Dy {
      w.Request_dims.Dy = child.Requested().Dy
    }
  }
  w.Request_dims.Dx += w.params.Spacing * len(w.Children)
}
func (w *HorizontalTable) Draw(region Region) {
  gl.Enable(gl.BLEND)
  gl.Disable(gl.TEXTURE_2D)
  dx := region.Dx
  if dx > w.Request_dims.Dx && !w.Ex {
    dx = w.Request_dims.Dx
  }
  dy := region.Dy
  if dy > w.Request_dims.Dy && !w.Ex {
    dy = w.Request_dims.Dy
  }
  gl.Color4d(
      w.params.Background.R,
      w.params.Background.G,
      w.params.Background.B,
      w.params.Background.A)
  gl.Begin(gl.QUADS)
  gl.Vertex2i(region.X, region.Y)
  gl.Vertex2i(region.X, region.Y+region.Dy)
  gl.Vertex2i(region.X+dx, region.Y+region.Dy)
  gl.Vertex2i(region.X+dx, region.Y)
  gl.End()
  gl.Color4d(
      w.params.Border.R,
      w.params.Border.G,
      w.params.Border.B,
      w.params.Border.A)
  gl.Begin(gl.LINES)
  gl.Vertex2i(region.X, region.Y)
  gl.Vertex2i(region.X, region.Y+region.Dy)

  gl.Vertex2i(region.X, region.Y+region.Dy)
  gl.Vertex2i(region.X+dx, region.Y+region.Dy)

  gl.Vertex2i(region.X+dx, region.Y+region.Dy)
  gl.Vertex2i(region.X+dx, region.Y)

  gl.Vertex2i(region.X+dx, region.Y+region.Dy-dy)
  gl.Vertex2i(region.X, region.Y)
  gl.End()

  fill_available := region.Dx - w.Request_dims.Dx
  if fill_available < 0 {
    fill_available = 0
  }
  fill_request := 0
  for _, child := range w.Children {
    if ex, _ := child.Expandable(); ex {
      fill_request += child.Requested().Dx
    }
  }
  var child_region Region
  child_region.X = region.X
  for _, child := range w.Children {
    child_region.Dims = child.Requested()
    if ex, _ := child.Expandable(); ex && fill_request > 0 {
      child_region.Dx += (child_region.Dx * fill_available) / fill_request
    }
    if region.Dx < w.Request_dims.Dx {
      child_region.Dims.Dx *= region.Dx
      child_region.Dims.Dx /= w.Request_dims.Dx
    }
    if child_region.Dy > region.Dy {
      child_region.Dy = region.Dy
    }
    if _, ey := child.Expandable(); child_region.Dy < region.Dy && ey {
      child_region.Dy = region.Dy
    }
    child_region.Y = region.Y + region.Dy - child_region.Dy
    child.Draw(child_region)
    child_region.X += child_region.Dx
    child_region.X += w.params.Spacing
  }
  w.Render_region = region
}
