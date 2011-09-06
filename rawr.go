package main

import (
  "glop/gos"
  "runtime"
  "time"
  "fmt"
  "io/ioutil"
  "image"
  "image/draw"
  "freetype-go.googlecode.com/hg/freetype"
  "freetype-go.googlecode.com/hg/freetype/truetype"
  "gl"
  "gl/glu"
  "os"
  "path"
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

func gameLoop() {
  defer gos.Quit()
  window := gos.CreateWindow(10, 10, 700, 700)
  texture := gl.GenTexture()
  texture.Bind(gl.TEXTURE_2D)
  fontpath := os.Args[0] + "/../../fonts/traveling_typewriter.ttf"
  fontpath = path.Clean(fontpath)
  font,err := loadFont(fontpath)
  if err != nil {
    fmt.Printf("Failed to load font: %s\n", err.String())
    return
  }
  context,err := makeContext()
  if err != nil {
    fmt.Printf("Failed to make font context: %s\n", err.String())
    return
  }
  rgba := image.NewRGBA(512, 512)
  for i := 0; i < 00; i++ {
    drawText(font, context, rgba, texture, []string{"Rawr!!!"})
  }

  err = drawText(font, context, rgba, texture, []string{"Rawr!!! :-D"})
  if err != nil {
    fmt.Printf("Couldn't render texture: %s\n", err.String())
    return
  }

  ticker := time.Tick(4*16666667)
  text := make([]string, 100)[0:0]
  for {
    gos.SwapBuffers(window)
    <-ticker
    err := drawText(font, context, rgba, texture, text)
    if err != nil {
      fmt.Printf("Couldn't draw text: %s\n", err.String())
      return
    }
    gl.LoadIdentity();
    gl.Ortho(-1.1,1.1, -1.1,1.1, -1.1,1.1)
    gl.MatrixMode(gl.MODELVIEW)
    gl.ClearColor(0.2, 0.0, 1.0, 1.0)
    gl.Clear(0x00004000)
    gl.Color3d(0, 1, 0)
    gl.Color4d(1.0, 1.0, 1.0, 0.7)
    gl.Enable(gl.TEXTURE_2D)
    gl.Enable(gl.BLEND)
    texture.Bind(gl.TEXTURE_2D)
    gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
    gl.Begin(gl.QUADS)
      gl.TexCoord2d(0,1)
      gl.Vertex2d(-1,-1)
      gl.TexCoord2d(0,0)
      gl.Vertex2d(-1, 1)
      gl.TexCoord2d(1,0)
      gl.Vertex2d( 1, 1)
      gl.TexCoord2d(1,1)
      gl.Vertex2d( 1,-1)
    gl.End()
    gl.Disable(gl.TEXTURE_2D)
    gos.Think()
    events := gos.GetInputEvents()
    for _,event := range events {
      if event.Index == 113 {
        return
      }
    }
    text = make([]string, len(events))
    for i := range events {
      text[i] = fmt.Sprintf("%v", events[i])
    }
  }
}

func main() {
  runtime.LockOSThread()
  gameLoop()
  return
  gos.Run()
  return
//return
  fmt.Printf("")
  //texture_size := 512
  //factor := 5.0
  //gl.Flush()
  //gl.Viewport(0, 0, 500, 500)

  //for {
    //time.Sleep(1)
    ////<-ticker
    //err = drawText(font, context, rgba, texture, text)
    //if err != nil {
      //fmt.Printf("Couldn't draw text: %s\n", err.String())
      //return
    //}
    //gl.MatrixMode(gl.PROJECTION)
    //gl.LoadIdentity();
    //gl.Ortho(-1,1, -1,1, -1,1)
    //gl.MatrixMode(gl.MODELVIEW)
////    gl.Translated(r/100,0,0)
    ////gl.Ortho(0, float64(texture_size)/factor, float64(texture_size)/factor, 0, 0, 1)
    //gl.ClearColor((gl.GLclampf)(r), 0.0, 1.0, 1.0)
    //gl.Clear(0x00004000)
    //gl.Color3d(0, 1, 0)
    //gl.Color4d(1.0, 1.0, 1.0, 0.7)
    //gl.Enable(gl.TEXTURE_2D)
    //gl.Enable(gl.BLEND)
    //texture.Bind(gl.TEXTURE_2D)
    //gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
    //gl.Begin(gl.QUADS)
      //gl.TexCoord2d(0,1)
      //gl.Vertex2d(0,0)
      //gl.TexCoord2d(0,0)
      //gl.Vertex2d(0,1)
      //gl.TexCoord2d(1,0)
      //gl.Vertex2d(1,1)
      //gl.TexCoord2d(1,1)
      //gl.Vertex2d(1,0)
    //gl.End()
    //gl.Disable(gl.TEXTURE_2D)
//


    //gos.SwapBuffers(window)
    //gos.Think()
    //v := gos.GetInputEvents()
    //text = make([]string, len(v))
    //for i := range v {
      //text[i] = fmt.Sprintf("(%d %d)  d(%d %d)", v[i].Mouse.Dx, v[i].Mouse.Dy, v[i].Mouse.X, v[i].Mouse.Y)
      //if v[i].Index == 113 {
        //return
      //}
    //}
    //r += 0.0101
  //}
}
