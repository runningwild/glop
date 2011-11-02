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
  NonFocuser
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
  NonFocuser
  BasicZone
  Childless

  active  bool
  texture gl.Texture
  r,g,b,a float64
}
func MakeImageBox() *ImageBox {
  var ib ImageBox
  ib.EmbeddedWidget = &BasicWidget{ CoreWidget : &ib }
  runtime.SetFinalizer(&ib, freeTexture)
  ib.r, ib.g, ib.b, ib.a = 1, 1, 1, 1
  return &ib
}
func (w *ImageBox) String() string {
  return "image box"
}
func (w *ImageBox) SetShading(r,g,b,a float64) {
  w.r,w.g,w.b,w.a = r,g,b,a
}
func freeTexture(w *ImageBox) {
  if w.active {
    w.texture.Delete()
    w.active = false
  }
  w.texture = 0
}

// Does not take ownserhip of the texture, you must still free the texture
// when you are done with it.
func (w *ImageBox) SetImageByTexture(texture gl.Texture, dx,dy int) {
  w.UnsetImage()
  w.texture = texture
  w.Request_dims.Dx = dx
  w.Request_dims.Dy = dy
  w.active = false
}
func (w *ImageBox) UnsetImage() {
  freeTexture(w)
}
func (w *ImageBox) SetImage(path string) {
  w.UnsetImage()
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
  canvas := image.NewRGBA(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))
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

  // We check texture == 0 and not active because active only indicates if we
  // have a texture that we need to free later.  It's possible for us to have
  // a texture that someone else owns.
  if w.texture == 0 { return }

  gl.Enable(gl.TEXTURE_2D)
  w.texture.Bind(gl.TEXTURE_2D)
  gl.Enable(gl.BLEND)
  gl.Color4d(w.r, w.g, w.b, w.a)
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
  NonFocuser
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


type OptionContainer interface {
  SetSelectedOption(Widget)
}

type SelectableWidget interface {
  Widget

  // The selectable widget will call this function when clicked
  SetSelectFunc(func(int64))

  SetSelected(bool)
  GetData() interface{}
}

type selectableOption struct {
  Clickable
  data interface{}
  parent OptionContainer
}
func (so *selectableOption) GetData() interface{} {
  return so.data
}
func (so *selectableOption) SetSelectFunc(f func(int64)) {
  so.on_click = f
}

type textOption struct {
  TextLine
  selectableOption
}

func (w *textOption) DoRespond(event_group EventGroup) (consume,change_focus bool) {
  w.selectableOption.DoRespond(event_group)
  return
}

func (w *textOption) SetSelected(selected bool) {
  if selected {
    w.SetColor(0.9, 1, 0.9, 1)
  } else {
    w.SetColor(0.6, 0.4, 0.4, 1)
  }
}

func makeTextOption(text string, width int) SelectableWidget {
  var so textOption
  so.TextLine = *MakeTextLine("standard", text, width, 1, 1, 1, 1)
  so.data = text
  so.EmbeddedWidget = &BasicWidget{ CoreWidget : &so }
  return &so
}

type imageOption struct {
  ImageBox
  selectableOption
}

func (w *imageOption) DoRespond(event_group EventGroup) (consume,change_focus bool) {
  w.selectableOption.DoRespond(event_group)
  return
}

func (w *imageOption) SetSelected(selected bool) {
  if selected {
    w.SetShading(1, 1, 1, 1)
  } else {
    w.SetShading(0.5, 0.5, 0.5, 0.9)
  }
}

func makeImageOption(path string, data interface{}) SelectableWidget {
  var sio imageOption
  sio.ImageBox = *MakeImageBox()
  sio.ImageBox.SetImage(path)
  sio.data = data
  sio.EmbeddedWidget = &BasicWidget{ CoreWidget : &sio }
  return &sio
}

type SelectBox struct {
  Table
  selected int
}

func MakeSelectBox(options []SelectableWidget, vertical bool) *SelectBox {
  var sb SelectBox
  if vertical {
    sb.Table = MakeVerticalTable()
  } else {
    sb.Table = MakeHorizontalTable()
  }
  for i := range options {
    option := options[i]
    option.SetSelectFunc(func(int64) {
      sb.SetSelectedOption(option.GetData())
    })
    sb.AddChild(option)
    option.SetSelected(false)
  }
  return &sb
}

func MakeSelectTextBox(text_options []string, width int) *SelectBox {
  options := make([]SelectableWidget, len(text_options))
  for i := range options {
    options[i] = makeTextOption(text_options[i], width)
  }
  return MakeSelectBox(options, true)
}

func MakeSelectImageBox(paths []string, names []string) *SelectBox {
  options := make([]SelectableWidget, len(paths))
  for i := range options {
    options[i] = makeImageOption(paths[i], names[i])
  }
  return MakeSelectBox(options, false)
}

func (w *SelectBox) String() string {
  return "select box"
}

func (w *SelectBox) GetSelectedIndex() int {
  return w.selected
}

func (w *SelectBox) SetSelectedIndex(index int) {
  w.selectIndex(index)
}

func (w *SelectBox) GetSelectedOption() interface{} {
  if w.selected == -1 { return "" }
  return w.GetChildren()[w.selected].(SelectableWidget).GetData()
}

func (w *SelectBox) SetSelectedOption(option interface{}) {
  for i := range w.GetChildren() {
    if w.GetChildren()[i].(SelectableWidget).GetData() == option {
      w.selectIndex(i)
      return
    }
  }
  w.selectIndex(-1)
}

func (w *SelectBox) selectIndex(index int) {
  if w.selected >= 0 {
    w.GetChildren()[w.selected].(SelectableWidget).SetSelected(false)
  }
  if index < 0 || index >= len(w.GetChildren()) {
    index = -1
  } else {
    w.GetChildren()[index].(SelectableWidget).SetSelected(true)
  }
  w.selected = index
}

