package gui

import (
  "github.com/runningwild/glop/gin"
)

type ComboBox struct {
  BasicZone
  StandardParent
  table    Table
  scroll   *ScrollFrame
  opened_region Region
  selected int
  open     bool
  clicked  bool
}

func (cb *ComboBox) Think(gui *Gui, t int64) {
  if cb.clicked {
    cb.clicked = false
    cb.open = false
    cb.opened_region = Region{}
  }
  cb.scroll.Think(gui, t)
  if f,ok := gui.FocusWidget().(*ComboBox); ok && f == cb {
    if !cb.open {
      gui.DropFocus()
    }
  }
  if cb.selected >= 0 && cb.selected < len(cb.table.GetChildren()) {
    cb.Request_dims = cb.table.GetChildren()[cb.selected].Requested()
  } else {
    cb.Request_dims = cb.table.GetChildren()[0].Requested()
  }
}

func (cb *ComboBox) Draw(region Region) {
  if cb.open { return }
  child := cb.table.GetChildren()[cb.selected]
  child.Draw(region)
  cb.Render_region = child.Rendered()
}
func (cb *ComboBox) DrawFocused(region Region) {
  if !cb.open { return }
  if cb.opened_region.Size() == 0 {
    cb.opened_region = cb.table.GetChildren()[cb.selected].Rendered()
  }
  r := cb.opened_region
  r.Y -= (cb.table.Requested().Dy - cb.opened_region.Dy) / 2
  r.Dy = cb.table.Requested().Dy
  cb.scroll.Draw(r.Fit(region))
  cb.Render_region = cb.scroll.Rendered()
}
func (cb *ComboBox) Respond(gui *Gui, group EventGroup) bool {
  if cb.open {
    if found,event := group.FindEvent(gin.Escape); found && event.Type == gin.Press {
      cb.clicked = true
      return true
    }
  }
  cursor := group.Events[0].Key.Cursor()
  if cursor != nil {
    var p Point
    p.X, p.Y = cursor.Point()
    if !p.Inside(cb.Rendered()) {
      return false
    }
  }
  if cb.clicked { return false }
  if group.Focus {
    cb.scroll.Respond(gui, group)
  }
  if !cb.open {
    if found,event := group.FindEvent(gin.MouseLButton); found && event.Type == gin.Press {
      gui.TakeFocus(cb)
      cb.open = true
    }
  }
  return cursor != nil
}

func MakeComboBox(options []SelectableWidget, width int) *ComboBox {
  var cb ComboBox
  cb.table = MakeVerticalTable()
  for i,option := range options {
    cb.table.AddChild(option)
    opt := i
    option.SetSelectFunc(func(int64) {
      cb.selected = opt
      cb.clicked = true
    })
  }
  cb.scroll = MakeScrollFrame(cb.table, width, 300)
  cb.AddChild(cb.scroll)
  return &cb
}

func MakeComboTextBox(text_options []string, width int) *ComboBox {
  options := make([]SelectableWidget, len(text_options))
  for i := range options {
    options[i] = makeTextOption(text_options[i], width)
  }
  return MakeComboBox(options, width)
}

func (w *ComboBox) String() string {
  return "combo box"
}

func (w *ComboBox) GetComboedIndex() int {
  return w.selected
}

func (w *ComboBox) SetSelectedIndex(index int) {
  w.selectIndex(index)
}

func (w *ComboBox) GetComboedOption() interface{} {
  if w.selected == -1 {
    return ""
  }
  return w.table.GetChildren()[w.selected].(SelectableWidget).GetData()
}

func (w *ComboBox) SetSelectedOption(option interface{}) {
  for i := range w.GetChildren() {
    if w.GetChildren()[i].(SelectableWidget).GetData() == option {
      w.selectIndex(i)
      return
    }
  }
  w.selectIndex(-1)
}

func (w *ComboBox) selectIndex(index int) {
  w.selected = index
}
