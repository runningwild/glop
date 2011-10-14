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

type ImageBox struct {
  EmbeddedWidget
  NonResponder
  NonThinker
  Rectangle
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
  w.Dims.Dx = dx
  w.Dims.Dy = dy
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

  w.Dx = img.Bounds().Dx()
  w.Dy = img.Bounds().Dy()
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
