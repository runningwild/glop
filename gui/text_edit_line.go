package gui

import (
  "glop/gin"
  "freetype-go.googlecode.com/hg/freetype"
  "gl"
)

type cursor struct {
  index  int     // Index before which the cursor is placed
  pos    float64 // position of the cursor in pixels from the left hand side
  moved  bool    // whether or not the cursor has been moved recently
  on     bool    // whether or not the curosr is showing
  period int64   // how fast the cursor should blink
  start  int64   // last time cursor.on was set to true
}
type TextEditLine struct {
  TextLine
  cursor cursor
}

func (w *TextEditLine) String() string {
  return "text edit line"
}

func MakeTextEditLine(font_name, text string, width int, r, g, b, a float64) *TextEditLine {
  var w TextEditLine
  w.TextLine = *MakeTextLine(font_name, text, width, r, g, b, a)
  w.EmbeddedWidget = &BasicWidget{CoreWidget: &w}

  w.scale = 1.0
  w.cursor.index = len(w.text)
  w.cursor.pos = w.findOffsetAtIndex(w.cursor.index)
  w.cursor.period = 500 // half a second
  return &w
}

func (w *TextEditLine) findIndexAtOffset(offset int) int {
  low := 0
  high := 1
  var low_off, high_off float64
  for high < len(w.text) && high_off < float64(offset) {
    low = high
    low_off = high_off
    high++
    high_off = w.findOffsetAtIndex(high)
  }
  if float64(offset)-low_off < high_off-float64(offset) {
    return low
  }
  return high
}

func (w *TextEditLine) findOffsetAtIndex(index int) float64 {
  pt := freetype.Pt(0, 0)
  if index > len(w.text) {
    index = len(w.text)
  }
  if index < 0 {
    index = 0
  }
  adv, _ := w.context.DrawString(w.text[:index], pt)
  return float64(adv.X>>8) * w.scale
}

func (w *TextEditLine) DoThink(t int64, focus bool) {
  changed := w.text != w.next_text
  w.TextLine.DoThink(t, false)
  if focus && w.cursor.start == 0 {
    w.cursor.start = t
    w.cursor.on = true
  }
  if !focus {
    w.cursor.start = 0
    w.cursor.on = false
  }
  if w.cursor.start > 0 {
    w.cursor.on = ((t-w.cursor.start)/w.cursor.period)%2 == 0
  }
  if w.cursor.moved || changed {
    w.cursor.pos = w.findOffsetAtIndex(w.cursor.index)
    w.cursor.moved = false
  }
}

func (w *TextEditLine) DoRespond(event_group EventGroup) (consume, change_focus bool) {
  event := event_group.Events[0]
  if event.Type != gin.Press {
    return
  }
  key_id := event.Key.Id()
  if event_group.Focus {
    if key_id == gin.Escape {
      change_focus = true
      return
    }
    if key_id == gin.Backspace {
      if len(w.text) > 0 && w.cursor.index > 0 {
        var pre, post string
        if w.cursor.index > 0 {
          pre = w.text[0 : w.cursor.index-1]
        }
        if w.cursor.index < len(w.text) {
          post = w.text[w.cursor.index:]
        }
        w.SetText(pre + post)
        w.cursor.index--
        w.cursor.moved = true
      }
    } else if key_id > 0 && key_id <= 127 && event.Type == gin.Press {
      w.SetText(w.text[0:w.cursor.index] + string([]byte{byte(key_id)}) + w.text[w.cursor.index:])
      w.cursor.index++
      w.cursor.moved = true
    } else if key_id == gin.MouseLButton {
      x, _ := event.Key.Cursor().Point()
      cx := w.TextLine.Render_region.X
      w.cursor.index = w.findIndexAtOffset(x - cx)
      w.cursor.moved = true
    }
    consume = true
  } else {
    change_focus = event.Key.Id() == gin.MouseLButton
  }
  return
}

func (w *TextEditLine) Draw(region Region) {
  region.PushClipPlanes()
  defer region.PopClipPlanes()
  gl.Disable(gl.TEXTURE_2D)
  gl.Color4d(0.3, 0.3, 0.3, 0.9)
  gl.Begin(gl.QUADS)
  gl.Vertex2i(region.X+1, region.Y+1)
  gl.Vertex2i(region.X+1, region.Y-1+region.Dy)
  gl.Vertex2i(region.X-1+region.Dx, region.Y-1+region.Dy)
  gl.Vertex2i(region.X-1+region.Dx, region.Y+1)
  gl.End()
  w.TextLine.preDraw(region)
  w.TextLine.coreDraw(region)
  gl.Disable(gl.TEXTURE_2D)
  if w.cursor.on {
    gl.Color3d(1, 0.3, 0)
  } else {
    gl.Color3d(0.5, 0.3, 0)
  }
  gl.Begin(gl.LINES)
  gl.Vertex2i(region.X+int(w.cursor.pos), region.Y)
  gl.Vertex2i(region.X+int(w.cursor.pos), region.Y+region.Dy)
  gl.End()
  w.TextLine.postDraw(region)
}
