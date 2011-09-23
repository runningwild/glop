package gui

import (
  "glop/gin"
  "os"
  "path"
  "path/filepath"
  "image"
  "image/draw"
  _ "image/png"
  _ "image/jpeg"
  "gl"
  "gl/glu"
  "math"
)

type Terrain struct {
  Childless

  // Default zoom is 1.0 and implies a 1:1 ratio of pixels on source image to
  // pixels on the screen
  zoom    float64

  // Length of the side of block in the source image.
  block_size int

  // Cursor position in window coordinates
  cx,cy int

  // Region that the terrain was displayed in last frame
  prev Region

  bg      image.Image
  texture gl.Texture
}

func MakeTerrain(bg_path string, block_size int) (*Terrain, os.Error) {
  var t Terrain

  bg_path = filepath.Join(os.Args[0], bg_path)
  bg_path = path.Clean(bg_path)
  f,err := os.Open(bg_path)
  if err != nil {
    return nil, err
  }
  defer f.Close()
  t.bg,_,err = image.Decode(f)
  if err != nil {
    return nil, err
  }

  t.block_size = block_size

  rgba := image.NewRGBA(t.bg.Bounds().Dx(), t.bg.Bounds().Dy())
  draw.Draw(rgba, t.bg.Bounds(), t.bg, image.Point{0,0}, draw.Over)

  gl.Enable(gl.TEXTURE_2D)
  t.texture = gl.GenTexture()
  t.texture.Bind(gl.TEXTURE_2D)
  gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
  glu.Build2DMipmaps(gl.TEXTURE_2D, 4, t.bg.Bounds().Dx(), t.bg.Bounds().Dy(), gl.RGBA, rgba.Pix)

  return &t, nil
}

func (t *Terrain) HighlightBlockAtCursor(x,y int) {
  t.cx = x
  t.cy = y
}
func (t *Terrain) Think(_ int64, _ bool, prev Region, _ map[Widget]Dims) (bool,Dims) {
  t.prev = prev
  return false,Dims{t.bg.Bounds().Dx(), t.bg.Bounds().Dy()}
}
func mulMat(v [4]float64, mat [16]float64) [4]float64 {
  var ret [4]float64
  for i := 0; i <= 3; i++ {
    for j := 0; j <= 3; j++ {
      ret[i] += v[j]*mat[j + 4*i]
    }
  }
  return ret
}
func (t *Terrain) Draw(dims Dims) {
  gl.MatrixMode(gl.MODELVIEW)
  gl.PushMatrix()
  defer gl.PopMatrix()
//  gl.Translated(float64(region.X), float64(region.Y), 0)



  gl.Rotated(45, 0,0,1)
  gl.Rotated(65, 1,-1,0)

  var mv_mat [16]float64
  gl.GetDoublev(gl.MODELVIEW_MATRIX, mv_mat[:])

  mx := mv_mat[0:2]
  my := mv_mat[4:6]
  det := mx[0]*my[1] - mx[1]*my[0]
  mxi := [2]float64{my[1] / det, -mx[1] / det}
  myi := [2]float64{-my[0] / det, mx[0] / det}
  cx := t.cx - t.prev.X
  cy := t.cy - t.prev.Y
  sx := float64(cx)*mxi[0] + float64(cy)*myi[0]
  sy := float64(cx)*mxi[1] + float64(cy)*myi[1]

  gl.Enable(gl.TEXTURE_2D)
  gl.Enable(gl.BLEND)
  gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
  t.texture.Bind(gl.TEXTURE_2D)
  gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
  tx := 0.0
  ty := 0.0
  tx2 := float64(t.prev.Dx) / float64(t.bg.Bounds().Dx())
  ty2 := float64(t.prev.Dy) / float64(t.bg.Bounds().Dy())
  tx2 = 1
  ty2 = 1
  gl.Color4d(1.0, 1.0, 1.0, 1.0)
  gl.Begin(gl.QUADS)
    gl.TexCoord2d(tx, -ty)
    gl.Vertex2d(0, 0)

    gl.TexCoord2d(tx, -ty2)
    gl.Vertex2d(0, float64(t.bg.Bounds().Dy()))

    gl.TexCoord2d(tx2, -ty2)
    gl.Vertex2d(float64(t.bg.Bounds().Dx()), float64(t.bg.Bounds().Dy()))

    gl.TexCoord2d(tx2, -ty)
    gl.Vertex2d(float64(t.bg.Bounds().Dx()), 0)
  gl.End()
  gl.Disable(gl.TEXTURE_2D)

  hx := int(math.Floor(sx / float64(t.block_size)))
  hy := int(math.Floor(sy / float64(t.block_size)))
  if hx >= 0 {
    gl.Color4d(1,0,0,0.5)
    bx := float64(t.block_size * hx)
    bx2 := float64(t.block_size * (hx + 1))
    by := float64(t.block_size * hy)
    by2 := float64(t.block_size * (hy + 1))
    gl.Begin(gl.QUADS)
      gl.Vertex2d(bx, by)
      gl.Vertex2d(bx, by2)
      gl.Vertex2d(bx2, by2)
      gl.Vertex2d(bx2, by)
    gl.End()
  }
}

func (t *Terrain) HandleEventGroup(event_group gin.EventGroup) (bool, bool, *Node) {
  return false, false, nil
}
