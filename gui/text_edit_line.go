package gui

import (
  "glop/gin"
  "freetype-go.googlecode.com/hg/freetype"
  "gl"
  "strings"
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
  reader := strings.NewReader(w.text)
  n := -1
  cx := 0.0
  for cx <= float64(offset) {
    rune,_,err := reader.ReadRune()
    if err != nil {
      n = -1
      break
    }
    w.glyph_buf.Load(w.font, w.font.Index(rune))
    cx += float64(w.context.FUnitToPixelRU(int(w.glyph_buf.B.XMax - w.glyph_buf.B.XMin + 1))) * w.scale
    n++
  }
  return n
}

func (w *TextEditLine) findOffsetAtIndex(index int) float64 {
  pt := freetype.Pt(0, 0)
  adv,_ := w.context.DrawString(w.text[ : index], pt)
  return float64(adv.X >> 8) * w.scale
}

func (w *TextEditLine) findOffsetAtIndexOLD(index int) float64 {
  reader := strings.NewReader(w.text)
  n := 0
  cx := 0.0
  prune := 0
  var funit int16
  for n < index {
    rune,_,err := reader.ReadRune()
    if err != nil {
      cx = -1
      break
    }
    w.glyph_buf.Load(w.font, w.font.Index(rune))
    funit += w.glyph_buf.B.XMax - w.glyph_buf.B.XMin + 1
    if prune != 0 {
      funit += w.font.Kerning(w.font.Index(prune), w.font.Index(rune))
    }
    cx = float64(w.context.FUnitToPixelRU(int(funit))) * w.scale
    prune = rune
    n++
  }
  return cx
}

func (w *TextEditLine) DoThink(t int64) {
  changed := w.changed
  w.TextLine.DoThink(t)
  if w.cursor_moved || changed {
    w.cursor_pos = w.findOffsetAtIndex(w.cursor_index)
    w.cursor_moved = false
  }
}

func (w *TextEditLine) DoRespond(event_group EventGroup) bool {
  event := event_group.Events[0]
  key_id := event.Key.Id()

  if key_id > 0 && key_id <= 127  && event.Type == gin.Press {
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
  return false
}

func (w *TextEditLine) Draw(region Region) {
  w.TextLine.preDraw(region)
  w.TextLine.coreDraw(region)
  gl.Disable(gl.TEXTURE_2D)
  gl.Color3d(1, 0, 0)
  gl.Begin(gl.LINES)
    gl.Vertex2d(w.cursor_pos, 0)
    gl.Vertex2d(w.cursor_pos, float64(region.Dy))
  gl.End()
  w.TextLine.postDraw(region)
}