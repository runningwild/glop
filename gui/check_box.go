package gui

import (
  "glop/gin"
  "gl"
)

type checkBox struct {
  EmbeddedWidget
  Childless
  NonFocuser
  NonResponder
  NonThinker
  BasicZone
  selected bool
  disabled bool
}
func makeCheckBox() *checkBox {
  var cb checkBox
  cb.EmbeddedWidget = &BasicWidget{CoreWidget: &cb}
  cb.BasicZone.Request_dims.Dx = 30
  cb.BasicZone.Request_dims.Dy = 30
  return &cb
}
func (cb *checkBox) String() string {
  return "check box"
}
func (cb *checkBox) Click() {
  if cb.disabled { return }
  cb.selected = !cb.selected
}
func (cb *checkBox) Draw(region Region) {
  cb.Render_region = region
  if cb.disabled {
    gl.Color3d(0.6, 0.6, 0.6)
  } else {
    gl.Color3d(1, 1, 1)
  }
  gl.Begin(gl.QUADS)
    gl.Vertex2i(region.X, region.Y)
    gl.Vertex2i(region.X, region.Y + region.Dy)
    gl.Vertex2i(region.X + region.Dx, region.Y + region.Dy)
    gl.Vertex2i(region.X + region.Dx, region.Y)
    if !cb.selected && region.Dx >= 4 && region.Dy >= 4 {
      gl.Color3d(0, 0, 0)
      gl.Vertex2i(region.X + 2, region.Y + 2)
      gl.Vertex2i(region.X + 2, region.Y + region.Dy - 2)
      gl.Vertex2i(region.X + region.Dx - 2, region.Y + region.Dy - 2)
      gl.Vertex2i(region.X + region.Dx - 2, region.Y + 2)
    }
  gl.End()
}

type checkRow struct {
  EmbeddedWidget
  *HorizontalTable
  check_box *checkBox
}
func (cb *checkRow) String() string {
  return "check row"
}
func makeCheckRow(w Widget) *checkRow {
  var cr checkRow
  cr.EmbeddedWidget = &BasicWidget{CoreWidget: &cr}
  cr.HorizontalTable = MakeHorizontalTable()
  cr.check_box = makeCheckBox()
  cr.AddChild(cr.check_box)
  cr.AddChild(w)
  return &cr
}
func (cr *checkRow) DoRespond(group EventGroup) (consume, change_focus bool) {
    if found,event := group.FindEvent(gin.MouseLButton); found && event.Type == gin.Press {
    cr.check_box.Click()
    consume = true
    return
  }
  return
}

type CheckBoxes struct {
  *VerticalTable
}
func (cb *CheckBoxes) DoRespond(group EventGroup) (consume, change_focus bool) {
  return false, false
}
func MakeCheckBoxes(options []Widget, width int) *CheckBoxes {
  var cb CheckBoxes
  cb.VerticalTable = MakeVerticalTable()
  for _,w := range options {
    cb.VerticalTable.AddChild(makeCheckRow(w))
  }
  return &cb
}

func MakeCheckTextBox(text_options []string, width int) *CheckBoxes {
  options := make([]Widget, len(text_options))
  for i := range options {
    options[i] = MakeTextLine("standard", text_options[i], width, 1, 1, 1, 1)
  }
  return MakeCheckBoxes(options, width)
}

func (cb *CheckBoxes) String() string {
  return "check boxes"
}

func (cb *CheckBoxes) GetSelectedIndexes() []int {
  var indexes []int
  for i,w := range cb.VerticalTable.GetChildren() {
    if w.(*checkRow).check_box.selected {
      indexes = append(indexes, i)
    }
  }
  return indexes
}
