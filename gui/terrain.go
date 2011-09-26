package gui

import (
  "glop/gin"
  "glop/util/algorithm"
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
  "rand"
)

type staticCellData struct {
  // Number of AP required to move into this square, move_cost < 0 is impassable
  move_cost int
}
type dynamicCellData struct {
  highlight bool
}

type cell struct {
  staticCellData
  dynamicCellData
}

type Terrain struct {
  Childless

  // Length of the side of block in the source image.
  block_size int

  // Cursor position in window coordinates
  cx,cy int

  // Focus, in map coordinates
  fx,fy float64

  // The viewing angle, 0 means the map is viewed head-on, 90 means the map is viewed
  // on its edge (i.e. it would not be visible)
  angle float64

  // Zoom factor, 1.0 is standard
  zoom float64

  grid [][]cell

  // Region that the terrain was displayed in last frame
  prev Region

  bg      image.Image
  texture gl.Texture
}

func (t *Terrain) NumVertex() int {
  return len(t.grid) * len(t.grid[0])
}
func (t *Terrain) fromVertex(v int) (int,int) {
  return v % len(t.grid), v / len(t.grid)
}
func (t *Terrain) toVertex(x,y int) int {
  return x + y * len(t.grid)
}
func (t *Terrain) Adjacent(v int) ([]int, []float64) {
  x,y := t.fromVertex(v)
  var adj []int
  var weight []float64
  for dx := -1; dx <= 1; dx++ {
    if x + dx < 0 || x + dx >= len(t.grid) { continue }
    for dy := -1; dy <= 1; dy++ {
      if dx == 0 && dy == 0 { continue }
      if y + dy < 0 || y + dy >= len(t.grid[0]) { continue }
      if t.grid[x+dx][y+dy].move_cost < 0 { continue }
      // Prevent moving along a diagonal if we couldn't get to that space normally via
      // either of the non-diagonal paths
      if dx != 0 && dy != 0 {
        if t.grid[x+dx][y].move_cost >= 0 && t.grid[x][y+dy].move_cost >= 0 {
          cost_a := float64(t.grid[x+dx][y].move_cost + t.grid[x][y+dy].move_cost) / 2
          cost_b := float64(t.grid[x+dx][y+dy].move_cost)
          adj = append(adj, t.toVertex(x+dx, y+dy))
          weight = append(weight, math.Fmax(cost_a, cost_b))
        }
      } else {
        if t.grid[x+dx][y+dy].move_cost >= 0 {
          adj = append(adj, t.toVertex(x+dx, y+dy))
          weight = append(weight, float64(t.grid[x+dx][y+dy].move_cost))
        }
      }
    }
  }
  return adj,weight
}


func MakeTerrain(bg_path string, block_size int, angle float64) (*Terrain, os.Error) {
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
  t.angle = angle

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

  t.grid = make([][]cell, 32)
  for i := range t.grid {
    t.grid[i] = make([]cell, 32)
    for j := range t.grid[i] {
      switch rand.Int() % 3 {
        case 0:
          t.grid[i][j].move_cost = -1
        case 1:
          t.grid[i][j].move_cost = 1
        case 2:
          t.grid[i][j].move_cost = 5
      }
    }
  }

  t.zoom = 1.0

  return &t, nil
}

// The change in x and y screen coordinates to apply to point on the terrain the is in
// focus.  These coordinates will be scaled by the current zoom.
func (t *Terrain) Move(dx,dy float64) {
  dy /= math.Sin(t.angle * 3.1415926535 / 180)
  dx,dy = dy+dx, dy-dx
  t.fx += dx / t.zoom
  t.fy += dy / t.zoom
  t.fx = math.Fmax(t.fx, 0)
  t.fy = math.Fmax(t.fy, 0)
  t.fx = math.Fmin(t.fx, float64(len(t.grid)))
  t.fy = math.Fmin(t.fy, float64(len(t.grid[0])))
}

// Changes the current zoom from e^(zoom) to e^(zoom+dz)
func (t *Terrain) Zoom(dz float64) {
  exp := math.Log(t.zoom) + dz
  exp = math.Fmax(exp, -1.25)
  exp = math.Fmin(exp, 1.25)
  t.zoom = math.Exp(exp)
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
  // Set our viewing volume to be a box with width and height specified by dims, and
  // centered on the origin.  This way we can just draw our terrain with a particular
  // point at the origin and that point will wind up centered in the window
  gl.MatrixMode(gl.PROJECTION)
  gl.PushMatrix()
  defer gl.PopMatrix()
  defer gl.MatrixMode(gl.PROJECTION)
  half_dx_zoom := float64(dims.Dx) / (t.zoom * 2)
  half_dy_zoom := float64(dims.Dy) / (t.zoom * 2)
  gl.Ortho(-half_dx_zoom, half_dx_zoom, -half_dy_zoom, half_dy_zoom, -1000, 1000)

  gl.MatrixMode(gl.MODELVIEW)
  gl.PushMatrix()
  defer gl.PopMatrix()
  defer gl.MatrixMode(gl.MODELVIEW)
  gl.Rotated(45, 0,0,1)
  gl.Rotated(t.angle, 1,-1,0)

  // Move the terrain so that (t.fx,t.fy) is at the origin, and hence becomes centered
  // in the window
  xoff := (t.fx + 0.5) * float64(t.block_size)
  yoff := (t.fy + 0.5) * float64(t.block_size)
  gl.Translated(-xoff, -yoff, 0)

  // Compute the inverse of the XY component of the modelview matrix so we can figure
  // out what cell the mouse is hovering over.
  var mv_mat [16]float64
  gl.GetDoublev(gl.MODELVIEW_MATRIX, mv_mat[:])
  mx := mv_mat[0:2]
  my := mv_mat[4:6]
  det := mx[0]*my[1] - mx[1]*my[0]
  mxi := [2]float64{my[1] / det, -mx[1] / det}
  myi := [2]float64{-my[0] / det, mx[0] / det}
  cx := float64(t.cx - t.prev.X) / t.zoom - half_dx_zoom
  cy := float64(t.cy - t.prev.Y) / t.zoom - half_dy_zoom
  cx -= mv_mat[12]
  cy -= mv_mat[13]
  sx := float64(cx)*mxi[0] + cy*myi[0]
  sy := float64(cx)*mxi[1] + cy*myi[1]


  // Draw a simple border around the terrain
  gl.Disable(gl.TEXTURE_2D)
  gl.Color4d(.3,.3,.3,1)
  gl.Begin(gl.QUADS)
  fdx := float64(t.bg.Bounds().Dx())
  fdy := float64(t.bg.Bounds().Dy())
  fbs := float64(t.block_size)
  gl.Vertex2d(-fbs, -fbs)
  gl.Vertex2d(-fbs, fdy+fbs)
  gl.Vertex2d(fdx+fbs, fdy+fbs)
  gl.Vertex2d(fdx+fbs, -fbs)
  gl.End()


  gl.Enable(gl.TEXTURE_2D)
  gl.Enable(gl.BLEND)
  gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
  t.texture.Bind(gl.TEXTURE_2D)
  gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
  gl.Color4d(1.0, 1.0, 1.0, 1.0)
  gl.Begin(gl.QUADS)
    gl.TexCoord2d(0, 0)
    gl.Vertex2d(0, 0)

    gl.TexCoord2d(0, -1)
    gl.Vertex2d(0, fdy)

    gl.TexCoord2d(1, -1)
    gl.Vertex2d(fdx, fdy)

    gl.TexCoord2d(1, 0)
    gl.Vertex2d(fdx, 0)
  gl.End()
  gl.Disable(gl.TEXTURE_2D)



  for i := range t.grid {
    for j := range t.grid[0] {
      switch {
        case t.grid[i][j].move_cost < 0:
          gl.Color4d(0, 0, 0, 0.4)
        case t.grid[i][j].move_cost == 1:
          gl.Color4d(0.0, 0.7, 0.1, 0.4)
        case t.grid[i][j].move_cost == 5:
          gl.Color4d(0.8, 0.1, 0.1, 0.4)
      }
      bx := float64(t.block_size * i)
      bx2 := float64(t.block_size * (i + 1))
      by := float64(t.block_size * j)
      by2 := float64(t.block_size * (j + 1))
      gl.Begin(gl.QUADS)
        gl.Vertex2d(bx, by)
        gl.Vertex2d(bx, by2)
        gl.Vertex2d(bx2, by2)
        gl.Vertex2d(bx2, by)
      gl.End()
    }
  }

  hx := int(math.Floor(sx / float64(t.block_size)))
  hy := int(math.Floor(sy / float64(t.block_size)))
  in_bounds :=  hx >= 0 && hy >= 0 && hx < len(t.grid) && hy < len(t.grid[0])

  var reach []int
  if in_bounds {
    reach = algorithm.ReachableWithinLimit(t, []int{ t.toVertex(hx, hy) }, 10)
  }

  for _,r := range reach {
    gl.Color4d(0.0, 0.1, 1.0, 0.5)
    i,j := t.fromVertex(r)
    bx := float64(t.block_size * i)
    bx2 := float64(t.block_size * (i + 1))
    by := float64(t.block_size * j)
    by2 := float64(t.block_size * (j + 1))
    gl.Begin(gl.QUADS)
      gl.Vertex2d(bx, by)
      gl.Vertex2d(bx, by2)
      gl.Vertex2d(bx2, by2)
      gl.Vertex2d(bx2, by)
    gl.End()
  }

  if in_bounds {
    gl.Color4d(0.1, 0.2, 1, 0.4)
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
