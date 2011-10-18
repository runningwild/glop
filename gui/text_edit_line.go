package gui

import (
  "glop/gin"
  "freetype-go.googlecode.com/hg/freetype"
  "gl"
)

type TextEditLine struct {
  TextLine
  cursor_index int
  cursor_pos   float64
  cursor_moved bool
}

func MakeTextEditLine(font_name,text string, r,g,b,a float64) *TextEditLine {
  var w TextEditLine
  w.TextLine = *MakeTextLine(font_name, text, r,g,b,a)
  w.BasicWidget.CoreWidget = &w

  w.scale = 1.0
  w.cursor_index = len(w.text)
  w.cursor_pos = w.findOffsetAtIndex(w.cursor_index)
  return &w
}

func (w *TextEditLine) findIndexAtOffset(offset int) int {
  low := 0
  high := 1
  var low_off,high_off float64
  for high <= len(w.text) && high_off < float64(offset) {
    low = high
    low_off = high_off
    high++
    high_off = w.findOffsetAtIndex(high)
  }
  if float64(offset) - low_off < high_off - float64(offset) {
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
  adv,_ := w.context.DrawString(w.text[ : index], pt)
  return float64(adv.X >> 8) * w.scale
}

func (w *TextEditLine) DoThink(t int64) {
  changed := w.changed
  w.TextLine.DoThink(t)
  if w.cursor_moved || changed {
    w.cursor_pos = w.findOffsetAtIndex(w.cursor_index)
    w.cursor_moved = false
  }
}

func (w *TextEditLine) DoRespond(event_group EventGroup) (consume,take_focus bool) {
  event := event_group.Events[0]
  if event.Type != gin.Press { return }
  key_id := event.Key.Id()
  if event_group.Focus {
    if key_id == 8 {
      if len(w.text) > 0 && w.cursor_index > 0 {
        var pre,post string
        if w.cursor_index > 0 {
          pre = w.text[0 : w.cursor_index - 1]
        }
        if w.cursor_index < len(w.text) {
          post = w.text[w.cursor_index : ]
        }
        w.SetText(pre + post)
        w.changed = true
        w.cursor_index--
        w.cursor_moved = true
      }
    } else if key_id > 0 && key_id <= 127  && event.Type == gin.Press {
      w.SetText(w.text[0:w.cursor_index] + string([]byte{byte(key_id)}) + w.text[w.cursor_index:])
      w.changed = true
      w.cursor_index++
      w.cursor_moved = true
    } else if key_id == 304 {
      x,_ := event.Key.Cursor().Point()
      cx := w.TextLine.Render_region.X
      w.cursor_index = w.findIndexAtOffset(x - cx)
      w.cursor_moved = true
    }
    consume = true
  } else {
    take_focus = event.Key.Id() == 304
  }
  return
}

func (w *TextEditLine) Draw(region Region) {
  region.PushClipPlanes()
  defer region.PopClipPlanes()
  gl.Color4d(0.3, 0.3, 0.3, 0.9)
  gl.Begin(gl.QUADS)
    gl.Vertex2i(region.X+1, region.Y+1)
    gl.Vertex2i(region.X+1, region.Y-1 + region.Dy)
    gl.Vertex2i(region.X-1 + region.Dx, region.Y-1 + region.Dy)
    gl.Vertex2i(region.X-1 + region.Dx, region.Y+1)
  gl.End()
  w.TextLine.preDraw(region)
  w.TextLine.coreDraw(region)
  gl.Disable(gl.TEXTURE_2D)
  gl.Color3d(1, 0, 0)
  gl.Begin(gl.LINES)
    gl.Vertex2i(region.X + int(w.cursor_pos), region.Y)
    gl.Vertex2i(region.X + int(w.cursor_pos), region.Y + region.Dy)
  gl.End()
  w.TextLine.postDraw(region)
}