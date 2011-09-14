package gui

import (
  "glop/gin"
)

// The GUI is handled in four steps:
// 1. Handle Events
//   As event groups are received from gin they are passed, one by one, towards whatever widget
//   is in focus.  Each widget that these events are passed through may decide use the events,
//   for example, a table widget that receives an event saying that the tab key was pressed may
//   consume this event and change focus from one widget it contains to another.
//
// 2. Thinking
//   Widget.Think() is called for all widgets only after events are processed.  This gives
//   widgets a chance to take focus based on input other that event groups that are passed
//   around in step 1.  Care must be taken to ensure that widgets are not competing for focus.
//
// 3. Resize
//   Widgets are recursively queried about their size.  A widget should query all of its children
//   and then report back about their size.  All widgets should recalculate and cache their size
//   on a call to Resize(), and simply report this value on Size().
//
// 4. Draw
//   Widgets are recursively called to draw themselves.
//   TODO: Figure out how to set the scissor box for all widgets to enforce the size their parent
//         suggests for them

// All widgets must embed BaseWidget and so will automatically implement Focusable.  
type Focusable interface {
  // Informs the widget that it has been offered focus because of the EventGroup passed
  // to it.  This widget returns true iff it wants to take focus.  The default implementation
  // of this method always returns false.
  OfferedFocus(gin.EventGroup) bool

  // Returns true iff this widget has focus.
  HasFocus() bool

  // Returns true iff this widget or one of its children has focus
  ContainsFocus() bool

  // Returns the gui's Focus object.  This can be used to give or take focus.
  Focus() *Focus

  // Add or remove a widget as a child of this widget.  All user-defined widgets that can
  // contain other widgets need to call these functions on their child widgets, unless those
  // widgets do not ever want to receive focus.
  AddChild(*BaseWidget)
  RemoveChild(*BaseWidget)

  // These methods require that the widget has gained or lost focus and should update
  // itself and its parents accordingly.
  gainFocus()
  loseFocus()
  getParent() *BaseWidget
  getThis() *BaseWidget
}

// All user-defined widgets must embed BaseWidget.  This defines several methods that are used
// for tracking focus and passing events.
type BaseWidget struct {
  parent   *BaseWidget
  children []*BaseWidget

  focus *Focus

  has_focus       bool
  child_has_focus bool
}
func (fw *BaseWidget) Focus() *Focus {
  return fw.focus
}
func (fw *BaseWidget) AddChild(w *BaseWidget) {
  fw.children = append(fw.children, w)
}
func (fw *BaseWidget) RemoveChild(w *BaseWidget) {
  for i,c := range fw.children {
    if c == w {
      fw.children[i] = fw.children[len(fw.children)-1]
      fw.children = fw.children[0 : len(fw.children)-1]
      return
    }
  }
}
func (fw *BaseWidget) getParent() *BaseWidget {
  return fw.parent
}
func (fw *BaseWidget) getThis() *BaseWidget {
  return fw
}
func (fw *BaseWidget) OfferedFocus(gin.EventGroup) bool {
  return false
}
func (fw *BaseWidget) HasFocus() bool {
  return fw.has_focus
}
func (fw *BaseWidget) ContainsFocus() bool {
  return fw.child_has_focus || fw.has_focus
}
func (fw *BaseWidget) loseFocus() {
  fw.has_focus = false
  fw.child_has_focus = false
  if fw.parent == nil { return }
  fw.parent.loseFocus()
}
func (fw *BaseWidget) gainFocus() {
  fw.has_focus = true
  fw.child_has_focus = false
  p := fw.parent
  for p != nil {
    p.child_has_focus = true
    p = p.parent
  }
}

// A Focus object tracks what widget has focus.  The widget with focus is the one that events
// will be directed to.  Every incoming EventGroup will be sent first to the root widget, then
// it will pass it to a child widget and so on until it reaches the widget with focus.  There
// are cases when a widget will want to send events elsewhere, for example consider a table with
// two text boxes, A and B, A has focus, B does not.  If the user clicks on B the table widget
// will want to notify B that it should take focus, so it calls focus.Give(B).  This will result
// in B.TookFocus(event_group) being called, so it knows that it has focus and the event that
// made this happen.
type Focus struct {
  widgets []FocusWidget
}
type FocusWidget interface {
  Focusable
  Widget
}

// Whatever widget currently has focus loses it, and the widget passed to this function gains it.
func (f *Focus) Take(w FocusWidget) {
  if len(f.widgets) > 0 {
    f.widgets[len(f.widgets)-1].loseFocus()
  } else {
    f.widgets = append(f.widgets, nil)
  }
  w.gainFocus()
  f.widgets[len(f.widgets)-1] = w
}

// Gives a widget the opportunity to take focus.  If it does take focus the current widget with
// focus loses it.
func (f *Focus) Give(w FocusWidget, event_group gin.EventGroup) {
  if w.OfferedFocus(event_group) {
    if len(f.widgets) > 0 {
      f.widgets[len(f.widgets)-1].loseFocus()
    } else {
      f.widgets = append(f.widgets, nil)
    }
    f.widgets[len(f.widgets)-1] = w
  }
}

// Whatever widget has focus now loses it, but will regain it when Focus.Pop() is called
func (f *Focus) Push(w FocusWidget) {
  if len(f.widgets) > 0 {
    (f.widgets)[len(f.widgets)-1].loseFocus()
  }
  w.gainFocus()
  f.widgets = append(f.widgets, w)
}

func (f *Focus) Pop() {
  if len(f.widgets) == 0 {
    panic("Cannot pop an empty Focus stack")
  }
  f.widgets[len(f.widgets)-1].loseFocus()
  f.widgets = (f.widgets)[0 : len(f.widgets)-1]
  if len(f.widgets) > 0 {
    f.widgets[len(f.widgets)-1].gainFocus()
  }
}

type Widget interface {
  Draw()
  Think()
  HandleEventGroup(gin.EventGroup)
  Size() (int,int)
  Anchor(w Widget, srcx,srcy,dstx,dsty float64)
}
