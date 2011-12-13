package gui

import (
  "glop/gin"
  "math"
)

type ScrollFrame struct {
  EmbeddedWidget
  NonFocuser
  NonThinker
  BasicZone
  StandardParent
  amt float64
  max float64
}

func MakeScrollFrame(w Widget, dx,dy int) *ScrollFrame {
  var frame ScrollFrame
  frame.Children = append(frame.Children, w)
  frame.EmbeddedWidget = &BasicWidget{CoreWidget: &frame}
  frame.Request_dims.Dx = dx
  frame.Request_dims.Dy = dy
  frame.amt = math.Inf(1)
  return &frame
}
func (w *ScrollFrame) String() string {
  return "scroll frame"
}
func (w *ScrollFrame) DoRespond(group EventGroup) (consume, take_focus bool) {
  if found, event := group.FindEvent(gin.MouseWheelVertical); found {
    inc := 2 * event.Key.FramePressAmt()
    w.amt += inc
    consume = true
  } else {
    return
  }
  if w.amt < 0 {
    w.amt = 0
    consume = false
  }
  if w.amt > w.max {
    w.amt = w.max
    consume = false
  }
  return
}
func (w *ScrollFrame) Draw(region Region) {
  if region.Dx > w.Request_dims.Dx {
    region.Dx = w.Request_dims.Dx
  }
  if region.Dy > w.Request_dims.Dy {
    region.Dy = w.Request_dims.Dy
  }
  region.PushClipPlanes()
  defer region.PopClipPlanes()
  if region.Dy >= w.Children[0].Requested().Dy {
    w.Children[0].Draw(region)
    w.Render_region = w.Children[0].Rendered()
    return
  }
  w.Render_region = region
  w.max = float64(w.Children[0].Requested().Dy - region.Dy)
  if w.amt > w.max {
    w.amt = w.max
  }
  w.Children[0].Draw(Region{ region.Point, w.Children[0].Requested() }.Add(Point{0, -int(w.amt) }))
}
