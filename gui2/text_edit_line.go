package gui

import (
  "glop/gin"
  "fmt"
  "freetype-go.googlecode.com/hg/freetype"
  "freetype-go.googlecode.com/hg/freetype/truetype"
  "gl"
  "strings"
)

type TextEditLine struct {
  TextLine
  cursor_pos int
}

func MakeTextEditLine(font_name,text string, r,g,b,a float64) *TextEditLine {
  var w TextEditLine
  w.BasicWidget.CoreWidget = &w
  font,ok := basic_fonts[font_name]
  if !ok {
    panic(fmt.Sprintf("Unable to find a font registered as '%s'.", font_name))
  }
  w.font = font
  w.glyph_buf = truetype.NewGlyphBuf()
  w.text = text
  w.psize = 72
  w.context = freetype.NewContext()
  w.context.SetDPI(132)
  w.context.SetFontSize(18)
  w.texture = gl.GenTexture()
  w.SetColor(r, g, b, a)
  w.figureDims()
  w.cursor_pos = len(w.text)
  return &w
}

//func (w *TextEditLine) DoThink(_ int64) {
//  if !w.changed { return }
//  w.changed = false
//  w.figureDims()
//}

func (w *TextEditLine) findIndexAtOffset(offset int) int {
  reader := strings.NewReader(w.text)
  n := 0
  cx := 0
  for cx < offset {
    rune,_,err := reader.ReadRune()
    if err != nil {
      w.cursor_pos = len(w.text)
      break
    }
    w.glyph_buf.Load(w.font, w.font.Index(rune))
    cx += int(w.glyph_buf.B.XMax - w.glyph_buf.B.XMin + 1)
    n++
  }
  return n
}

func (w *TextEditLine) DoRespond(event_group EventGroup) bool {
  event := event_group.Events[0]
  key_id := event.Key.Id()

    fmt.Printf("id: %d\n", key_id)
  if key_id > 0 && key_id <= 127  && event.Type == gin.Press {
    w.SetText(w.text[0:w.cursor_pos] + string([]byte{byte(key_id)}) + w.text[w.cursor_pos:])
    w.cursor_pos++
  } else if key_id == 304 {
    x,_ := event.Key.Cursor().Point()
    cx := w.Rectangle.X

    w.cursor_pos = w.findIndexAtOffset(x - cx)
    fmt.Printf("set pos: %d\n", w.cursor_pos)
  }
  return false
}

