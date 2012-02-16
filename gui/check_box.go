package gui

import (
  "github.com/runningwild/glop/gin"
  "github.com/runningwild/opengl/gl"
  "reflect"
)

type checkBoxSelection int
const (
  checkBoxSelected checkBoxSelection = iota
  checkBoxUnselected
  checkBoxUnknown
)

type checkBox struct {
  EmbeddedWidget
  Childless
  NonFocuser
  NonResponder
  NonThinker
  BasicZone
  selected checkBoxSelection
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
  switch cb.selected {
    case checkBoxSelected:
      cb.selected = checkBoxUnselected
    case checkBoxUnselected:
      cb.selected = checkBoxSelected
    case checkBoxUnknown:
      cb.selected = checkBoxSelected
  }
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
    if cb.selected == checkBoxUnknown || cb.selected == checkBoxUnselected {
      if cb.selected == checkBoxUnknown {
        gl.Color3d(0.4, 0.4, 0.4)
      } else {
        gl.Color3d(0, 0, 0)
      }
      if region.Dx >= 4 && region.Dy >= 4 {
        gl.Vertex2i(region.X + 2, region.Y + 2)
        gl.Vertex2i(region.X + 2, region.Y + region.Dy - 2)
        gl.Vertex2i(region.X + region.Dx - 2, region.Y + region.Dy - 2)
        gl.Vertex2i(region.X + region.Dx - 2, region.Y + 2)
      }
    }
  gl.End()
}

type checkRow struct {
  EmbeddedWidget
  *HorizontalTable
  check_box    *checkBox
  target,index reflect.Value
}
func (cb *checkRow) String() string {
  return "check row"
}
func makeCheckRow(w Widget, target,index reflect.Value) *checkRow {
  var cr checkRow
  cr.EmbeddedWidget = &BasicWidget{CoreWidget: &cr}
  cr.HorizontalTable = MakeHorizontalTable()
  cr.check_box = makeCheckBox()
  cr.target = target
  cr.index = index
  cr.AddChild(cr.check_box)
  cr.AddChild(w)
  return &cr
}
func (cr *checkRow) DoRespond(group EventGroup) (consume, change_focus bool) {
  if found,event := group.FindEvent(gin.MouseLButton); found && event.Type == gin.Press {
    cr.check_box.Click()
    var selected reflect.Value
    if cr.check_box.selected == checkBoxSelected {
      selected = reflect.ValueOf(true)
    } else {
      selected = reflect.ValueOf(false)
    }
    cr.target.SetMapIndex(cr.index, selected)
    consume = true
    return
  }
  return
}
func (cr *checkRow) DoThink(t int64, focus bool) {
  val := cr.target.MapIndex(cr.index)
  if val.IsValid() {
    if val.Bool() {
      cr.check_box.selected = checkBoxSelected
    } else {
      cr.check_box.selected = checkBoxUnselected
    }
  } else {
    cr.check_box.selected = checkBoxUnknown
  }
  cr.HorizontalTable.DoThink(t, focus)
}

type CheckBoxes struct {
  *VerticalTable
  target reflect.Value
}
func (cb *CheckBoxes) DoRespond(group EventGroup) (consume, change_focus bool) {
  return false, false
}

// target = reflect.ValueOf(&map[<option_type>]bool)
func MakeCheckBoxes(options []Widget, indexes []reflect.Value, width int, target reflect.Value) *CheckBoxes {
  var cb CheckBoxes
  cb.VerticalTable = MakeVerticalTable()
  cb.target = target
  for i := range options {
    cb.VerticalTable.AddChild(makeCheckRow(options[i], target, indexes[i]))
  }
  return &cb
}

func MakeCheckTextBox(text_options []string, width int, target map[string]bool) *CheckBoxes {
  options := make([]Widget, len(text_options))
  indexes := make([]reflect.Value, len(text_options))
  for i := range options {
    options[i] = MakeTextLine("standard", text_options[i], width, 1, 1, 1, 1)
    indexes[i] = reflect.ValueOf(text_options[i])
  }
  return MakeCheckBoxes(options, indexes, width, reflect.ValueOf(target))
}

func (cb *CheckBoxes) String() string {
  return "check boxes"
}

func (cb *CheckBoxes) GetSelectedIndexes() []int {
  var indexes []int
  for i,w := range cb.VerticalTable.GetChildren() {
    if w.(*checkRow).check_box.selected == checkBoxSelected {
      indexes = append(indexes, i)
    }
  }
  return indexes
}
