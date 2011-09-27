package gui

import (
  "glop/gin"
  "glop/util/algorithm"
  "glop/sprite"
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
  "github.com/arbaal/mathgl"
)

type staticCellData struct {
  // Number of AP required to move into this square, move_cost < 0 is impassable
  move_cost int
}
type dynamicCellData struct {
  s *sprite.Sprite
}

type cell struct {
  staticCellData
  dynamicCellData
}

// TODO: change to float32 everywhere since that's what mathgl uses
type Terrain struct {
  Childless

  // Length of the side of block in the source image.
  block_size int

  // Most recently clicked position on the board
  hx,hy int

  // Focus, in map coordinates
  fx,fy float64

  // The viewing angle, 0 means the map is viewed head-on, 90 means the map is viewed
  // on its edge (i.e. it would not be visible)
  angle float64

  // Zoom factor, 1.0 is standard
  zoom float64

  // The modelview matrix that is sent to opengl.  Updated any time focus, zoom, or viewing
  // angle changes
  mat mathgl.Mat4

  // Inverse of mat
  imat mathgl.Mat4

  grid [][]cell
  positions []mathgl.Vec3
  drawables []sprite.ZDrawable

  // Region that the terrain was displayed in last frame
  prev Region

  bg      image.Image
  texture gl.Texture
}

func (t *Terrain) AddZDrawable(x,y float32, zd sprite.ZDrawable) {
  vx,vy,vz := t.boardToModelview(x, y)
  t.positions = append(t.positions, mathgl.Vec3{ vx, vy, vz })
  t.drawables = append(t.drawables, zd)
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

func MakeTerrain(bg_path string, block_size,dx,dy int, angle float64) (*Terrain, os.Error) {
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

  t.grid = make([][]cell, dx)
  for i := range t.grid {
    t.grid[i] = make([]cell, dy)
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
  if err != nil {
    return nil,err
  }
  t.zoom = 1.0
  t.makeMat()

  return &t, nil
}

func (t *Terrain) makeMat() {
  var m mathgl.Mat4
  t.mat.RotationZ(45 * 3.1415926535 / 180)
  m.RotationAxisAngle(mathgl.Vec3{ X : -1, Y : 1}, -float32(t.angle) * 3.1415926535 / 180)
  t.mat.Multiply(&m)

  s := float32(t.zoom)
  m.Scaling(s, s, s)
  t.mat.Multiply(&m)

  // Move the terrain so that (t.fx,t.fy) is at the origin, and hence becomes centered
  // in the window
  xoff := (t.fx + 0.5) * float64(t.block_size)
  yoff := (t.fy + 0.5) * float64(t.block_size)
  m.Translation(-float32(xoff), -float32(yoff), 0)
  t.mat.Multiply(&m)

  t.imat.Assign(&t.mat)
  t.imat.Inverse()
}

func (t *Terrain) modelviewToBoard(mx,my float32) (int,int) {
  mz := my * float32(math.Tan(t.angle * 3.1415926535 / 180))
  v := mathgl.Vec4{ X : mx, Y : my, Z : mz, W : 1 }
  v.Transform(&t.imat)
  return int(v.X / float32(t.block_size)), int(v.Y / float32(t.block_size))
}

func (t *Terrain) boardToModelview(mx,my float32) (x,y,z float32) {
  v := mathgl.Vec4{ X : mx * float32(t.block_size), Y : my * float32(t.block_size), W : 1 }
  v.Transform(&t.mat)
  x,y,z = v.X, v.Y, v.Z
  return
}

// The change in x and y screen coordinates to apply to point on the terrain the is in
// focus.  These coordinates will be scaled by the current zoom.
func (t *Terrain) Move(dx,dy float64) {
  if dx == 0 && dy == 0 { return }
  dy /= math.Cos(t.angle * 3.1415926535 / 180)
  dx,dy = dy+dx, dy-dx
  t.fx += dx / t.zoom
  t.fy += dy / t.zoom
  t.fx = math.Fmax(t.fx, 0)
  t.fy = math.Fmax(t.fy, 0)
  t.fx = math.Fmin(t.fx, float64(len(t.grid)))
  t.fy = math.Fmin(t.fy, float64(len(t.grid[0])))
  t.makeMat()
}

// Changes the current zoom from e^(zoom) to e^(zoom+dz)
func (t *Terrain) Zoom(dz float64) {
  if dz == 0 { return }
  exp := math.Log(t.zoom) + dz
  exp = math.Fmax(exp, -1.25)
  exp = math.Fmin(exp, 1.25)
  t.zoom = math.Exp(exp)
  t.makeMat()
}

func (t *Terrain) Think(_ int64, _ bool, prev Region, _ map[Widget]Dims) (bool,Dims) {
  t.prev = prev
  return false,Dims{t.bg.Bounds().Dx(), t.bg.Bounds().Dy()}
}

func (t *Terrain) Draw(dims Dims) {
  // Set our viewing volume to be a box with width and height specified by dims, and
  // centered on the origin.  This way we can just draw our terrain with a particular
  // point at the origin and that point will wind up centered in the window
  gl.MatrixMode(gl.PROJECTION)
  gl.PushMatrix()
  half_dx_zoom := float64(dims.Dx) / 2
  half_dy_zoom := float64(dims.Dy) / 2
  gl.Ortho(-half_dx_zoom, half_dx_zoom, -half_dy_zoom, half_dy_zoom, 1000, -1000)
  defer gl.PopMatrix()
  defer gl.MatrixMode(gl.PROJECTION)

  gl.MatrixMode(gl.MODELVIEW)
  gl.PushMatrix()
  gl.MultMatrixf(&t.mat[0])
  defer gl.PopMatrix()
  defer gl.MatrixMode(gl.MODELVIEW)

  gl.Disable(gl.DEPTH_TEST)
  gl.Enable(gl.BLEND)
  gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
  fdx := float64(t.bg.Bounds().Dx())
  fdy := float64(t.bg.Bounds().Dy())

  // Draw a simple border around the terrain
  gl.Disable(gl.TEXTURE_2D)
  gl.Color4d(1,.3,.3,1)
  gl.Begin(gl.QUADS)
    fbs := float64(t.block_size)
    gl.Vertex3d(   -fbs,    -fbs, 1)
    gl.Vertex3d(   -fbs, fdy+fbs, 1)
    gl.Vertex3d(fdx+fbs, fdy+fbs, 1)
    gl.Vertex3d(fdx+fbs,    -fbs, 1)
  gl.End()

  gl.Enable(gl.TEXTURE_2D)
  t.texture.Bind(gl.TEXTURE_2D)
  gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
  gl.Color4d(1.0, 1.0, 1.0, 1.0)
  gl.Begin(gl.QUADS)
    gl.TexCoord2d(0, 0)
    gl.Vertex3d(0, 0, 0)
    gl.TexCoord2d(0, -1)
    gl.Vertex3d(0, fdy, 0)
    gl.TexCoord2d(1, -1)
    gl.Vertex3d(fdx, fdy, 0)
    gl.TexCoord2d(1, 0)
    gl.Vertex3d(fdx, 0, 0)
  gl.End()
  gl.Disable(gl.TEXTURE_2D)

  sprite.ZSort(t.positions, t.drawables)

  gl.PushMatrix()
  gl.LoadIdentity()
  for i := range t.positions {
    v := t.positions[i]
    t.drawables[i].Render(v.X, v.Y, v.Z, float32(t.zoom))
  }
  t.positions = t.positions[0:0]
  t.drawables = t.drawables[0:0]
  gl.PopMatrix()

  return

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

  in_bounds :=  t.hx >= 0 && t.hy >= 0 && t.hx < len(t.grid) && t.hy < len(t.grid[0])
  var reach []int
  if in_bounds {
    reach = algorithm.ReachableWithinLimit(t, []int{ t.toVertex(t.hx, t.hy) }, 10)
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
    bx := float64(t.block_size * t.hx)
    bx2 := float64(t.block_size * (t.hx + 1))
    by := float64(t.block_size * t.hy)
    by2 := float64(t.block_size * (t.hy + 1))
    gl.Begin(gl.QUADS)
      gl.Vertex2d(bx, by)
      gl.Vertex2d(bx, by2)
      gl.Vertex2d(bx2, by2)
      gl.Vertex2d(bx2, by)
    gl.End()
  }
}

func (t *Terrain) HandleEventGroup(event_group gin.EventGroup) (bool, bool, *Node) {
  if found,event := event_group.FindEvent(304); found && event.Type == gin.Press {
    x, y := event.Key.Cursor().Point()
    mx := float32(x - t.prev.X) - float32(t.prev.Dims.Dx) / 2
    my := float32(y - t.prev.Y) - float32(t.prev.Dims.Dy) / 2
    t.hx, t.hy = t.modelviewToBoard(mx, my)
  }
  return false, false, nil
}
