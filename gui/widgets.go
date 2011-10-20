package gui

import (
  "image"
  _ "image/png"
  _ "image/jpeg"
  "os"
  "gl"
  "gl/glu"
  "runtime"
)

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
  BasicZone
  children []Widget
  anchors  []Anchor
}
func MakeAnchorBox(dims Dims) *AnchorBox {
  var box AnchorBox
  box.EmbeddedWidget = &BasicWidget{ CoreWidget : &box }
  box.Request_dims = dims
  return &box
}
func (w *AnchorBox) String() string {
  return "anchor box"
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
  w.Render_region = region
  for i := range w.children {
    widget := w.children[i]
    anchor := w.anchors[i]
    var child_region Region
    child_region.Dims = widget.Requested()
    xoff := int(anchor.Bx * float64(region.Dx) - anchor.Wx * float64(child_region.Dx) + 0.5)
    yoff := int(anchor.By * float64(region.Dy) - anchor.Wy * float64(child_region.Dy) + 0.5)
    if xoff < 0 {
      child_region.Dx += xoff
      xoff = 0
    }
    if yoff < 0 {
      child_region.Dy += yoff
      yoff = 0
    }
    if xoff + child_region.Dx > w.Render_region.Dx {
      child_region.Dx -= (xoff + child_region.Dx) - w.Render_region.Dx
    }
    if yoff + child_region.Dy > w.Render_region.Dy {
      child_region.Dy -= (yoff + child_region.Dy) - w.Render_region.Dy
    }
    child_region.X = xoff
    child_region.Y = yoff
    widget.Draw(child_region)
  }
}

type ImageBox struct {
  EmbeddedWidget
  NonResponder
  NonThinker
  BasicZone
  Childless

  active  bool
  texture gl.Texture
}
func MakeImageBox() *ImageBox {
  var ib ImageBox
  ib.EmbeddedWidget = &BasicWidget{ CoreWidget : &ib }
  runtime.SetFinalizer(&ib, freeTexture)
  return &ib
}
func (w *ImageBox) String() string {
  return "image box"
}
func freeTexture(w *ImageBox) {
  if w.active {
    w.texture.Delete()
  }
}
func (w *ImageBox) UnsetImage() {
  w.active = false
}
func (w *ImageBox) SetImageByTexture(texture gl.Texture, dx,dy int) {
  w.texture = texture
  w.Request_dims.Dx = dx
  w.Request_dims.Dy = dy
  w.active = true
}
func (w *ImageBox) SetImage(path string) {
  data,err := os.Open(path)
  if err != nil {
    // TODO: Log error
    return
  }

  var img image.Image
  img,_,err = image.Decode(data)
  if err != nil {
    // TODO: Log error
    return
  }

  w.Request_dims.Dx = img.Bounds().Dx()
  w.Request_dims.Dy = img.Bounds().Dy()
  canvas := image.NewRGBA(img.Bounds().Dx(), img.Bounds().Dy())
  for y := 0; y < canvas.Bounds().Dy(); y++ {
    for x := 0; x < canvas.Bounds().Dx(); x++ {
      r,g,b,a := img.At(x,y).RGBA()
      base := 4*x + canvas.Stride*y
      canvas.Pix[base]   = uint8(r)
      canvas.Pix[base+1] = uint8(g)
      canvas.Pix[base+2] = uint8(b)
      canvas.Pix[base+3] = uint8(a)
    }
  }

  w.texture = gl.GenTexture()
  gl.Enable(gl.TEXTURE_2D)
  w.texture.Bind(gl.TEXTURE_2D)
  gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
  glu.Build2DMipmaps(gl.TEXTURE_2D, 4, img.Bounds().Dx(), img.Bounds().Dy(), gl.RGBA, canvas.Pix)

  w.active = true
}
func (w *ImageBox) Draw(region Region) {
  w.Render_region = region
  if !w.active { return }
  gl.Enable(gl.TEXTURE_2D)
  w.texture.Bind(gl.TEXTURE_2D)
  gl.Color4d(1.0, 1.0, 1.0, 1.0)
  gl.Begin(gl.QUADS)
    gl.TexCoord2f(0, 0)
    gl.Vertex2i(region.X, region.Y)
    gl.TexCoord2f(0, -1)
    gl.Vertex2i(region.X, region.Y + region.Dy)
    gl.TexCoord2f(1, -1)
    gl.Vertex2i(region.X + region.Dx, region.Y + region.Dy)
    gl.TexCoord2f(1, 0)
    gl.Vertex2i(region.X + region.Dx, region.Y)
  gl.End()
  gl.Disable(gl.TEXTURE_2D)
}

type CollapseWrapper struct {
  EmbeddedWidget
  Wrapper
  CollapsableZone
  NonResponder
}

func MakeCollapseWrapper(w Widget) *CollapseWrapper {
  var cw CollapseWrapper
  cw.EmbeddedWidget = &BasicWidget{ CoreWidget : &cw }
  cw.Child = w
  return &cw
}

func (w *CollapseWrapper) String() string {
  return "collapse wrapper"
}


func (w *CollapseWrapper) DoThink(int64, bool) {
  w.Request_dims = w.Child.Requested()
  w.Render_region = w.Child.Rendered()
}

func (w *CollapseWrapper) Draw(region Region) {
  if w.Collapsed {
    w.Render_region = Region{}
    return
  }
  w.Child.Draw(region)
  w.Render_region = region
}

type SelectTextBox struct {
  EmbeddedWidget
  Wrapper
  BasicZone
  NonThinker
  NonResponder
  options  []string
  selected int
}

func MakeSelectTextBox(options []string, width int) *SelectTextBox {
  var stb SelectTextBox
  stb.EmbeddedWidget = &BasicWidget{ CoreWidget : &stb }
  v := MakeVerticalTable()
  for i := range options {
    v.AddChild(MakeTextLine("standard", options[i], width, 1, 1, 1, 1))
  }
  stb.Child = v
  return &stb
}

func (w *SelectTextBox) String() string {
  return "select text box"
}

func (w *SelectTextBox) DoThink(int64, bool) {
  w.Request_dims = w.Child.Requested()
  w.Render_region = w.Child.Rendered()
}


