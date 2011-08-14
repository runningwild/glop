package main

import (
  "glop/gos"
  "runtime"
  "io/ioutil"
  "image"
  "image/draw"
  "freetype-go.googlecode.com/hg/freetype"
  "freetype-go.googlecode.com/hg/freetype/truetype"
  "gl"
  "gl/glu"
  "os"
  "time"
)

func loadFont(filename string) (*truetype.Font, os.Error) {
  data,err := ioutil.ReadFile(filename)
  if err != nil {
    return nil, err
  }
  font,err := freetype.ParseFont(data)
  if err != nil {
    return nil, err
  }
  return font,nil
}

func makeContext() (*freetype.Context, os.Error) {
  c := freetype.NewContext()
  c.SetDPI(200)
  c.SetFontSize(15)
  return c, nil
}

func drawText(font *truetype.Font, c *freetype.Context, rgba *image.RGBA, texture gl.Texture, text []string) os.Error {
  fg, bg := image.Black, image.White
  draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)
  c.SetFont(font)
  c.SetDst(rgba)
  c.SetSrc(fg)
  c.SetClip(rgba.Bounds())
  pt := freetype.Pt(10, 10+c.FUnitToPixelRU(font.UnitsPerEm()))
  for _, s := range text {
    _, err := c.DrawString(s, pt)
    if err != nil {
      return err
    }
    pt.Y += c.PointToFix32(15 * 1.5)
  }
  gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
  glu.Build2DMipmaps(gl.TEXTURE_2D, 4, 512, 512, gl.RGBA, rgba.Pix)
  return nil
}

func main() {
  runtime.LockOSThread()
  window := gos.CreateWindow(10, 10, 500, 500)
  for {
    time.Sleep(4*16666667)
    //<-ticker



    gos.SwapBuffers(window)
    gos.Think()
    gos.GetInputEvents()
  }
}
