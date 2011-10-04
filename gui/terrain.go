package gui

import (
  "glop/gin"
  "glop/sprite"
  "os"
  "image"
  "image/draw"
  _ "image/png"
  _ "image/jpeg"
  "gl"
  "gl/glu"
  "math"
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

  // All events received by the terrain are passed to the handler
  handler gin.EventHandler

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

  // All drawables that will be drawn parallel to the window
  upright_drawables []sprite.ZDrawable
  upright_positions []mathgl.Vec3

  // All drawables that will be drawn on the surface of the terrain
  flattened_drawables []sprite.ZDrawable
  flattened_positions []mathgl.Vec3

  // Region that the terrain was displayed in last frame
  prev Region

  // Don't need to keep the image around once it's loaded into texture memory,
  // only need to keep around the dimensions
  bg      image.Image
  texture gl.Texture
}

func (t *Terrain) AddUprightDrawable(x,y float32, zd sprite.ZDrawable) {
  t.upright_drawables = append(t.upright_drawables, zd)
  t.upright_positions = append(t.upright_positions, mathgl.Vec3{ x, y, 0 })
}

// x,y: board coordinates that the drawable should be drawn at.
// zd: drawable that will be rendered after the terrain has been rendered, it will be rendered
//     with the same modelview matrix as the rest of the terrain
func (t *Terrain) AddFlattenedDrawable(x,y float32, zd sprite.ZDrawable) {
  t.flattened_drawables = append(t.flattened_drawables, zd)
  t.flattened_positions = append(t.flattened_positions, mathgl.Vec3{ x, y, 0 })
}

func MakeTerrain(bg_path string, block_size,dx,dy int, angle float64) (*Terrain, os.Error) {
  var t Terrain

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

  rgba := image.NewRGBA(t.bg.Bounds())
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

// Transforms a cursor position in window coordinates to board coordinates.  Does not check
// to make sure that the values returned represent a valid position on the board.
func (t *Terrain) WindowToBoard(wx,wy int) (float32, float32) {
  mx := float32(wx - t.prev.X) - float32(t.prev.Dims.Dx) / 2
  my := float32(wy - t.prev.Y) - float32(t.prev.Dims.Dy) / 2
  return t.modelviewToBoard(mx, my)
}

func (t *Terrain) modelviewToBoard(mx,my float32) (float32,float32) {
  mz := my * float32(math.Tan(t.angle * 3.1415926535 / 180))
  v := mathgl.Vec4{ X : mx, Y : my, Z : mz, W : 1 }
  v.Transform(&t.imat)
  return v.X / float32(t.block_size), v.Y / float32(t.block_size)
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
  dy /= math.Sin(t.angle * 3.1415926535 / 180)
  dx,dy = dy+dx, dy-dx
  t.fx += dx / t.zoom
  t.fy += dy / t.zoom
  t.fx = math.Fmax(t.fx, 0)
  t.fy = math.Fmax(t.fy, 0)
  t.fx = math.Fmin(t.fx, float64(t.bg.Bounds().Dx()) / float64(t.block_size))
  t.fy = math.Fmin(t.fy, float64(t.bg.Bounds().Dy()) / float64(t.block_size))
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
  gl.Color4d(0,0,0, 0.5)
  gl.Begin(gl.LINES)
  for i := 0.0; i < float64(t.bg.Bounds().Dx()); i += float64(t.block_size) {
    gl.Vertex2d(i, 0)
    gl.Vertex2d(i, float64(t.bg.Bounds().Dy()))
  }
  for j := 0.0; j < float64(t.bg.Bounds().Dy()); j += float64(t.block_size) {
    gl.Vertex2d(0, j)
    gl.Vertex2d(float64(t.bg.Bounds().Dx()), j)
  }
  gl.End()

  for i := range t.flattened_positions {
    v := t.flattened_positions[i]
    t.flattened_drawables[i].Render(v.X, v.Y, 0, float32(t.block_size))
  }
  t.flattened_positions = t.flattened_positions[0:0]
  t.flattened_drawables = t.flattened_drawables[0:0]

  for i := range t.upright_positions {
    vx,vy,vz := t.boardToModelview(t.upright_positions[i].X, t.upright_positions[i].Y)
    t.upright_positions[i] = mathgl.Vec3{ vx, vy, vz }
  }
  sprite.ZSort(t.upright_positions, t.upright_drawables)
  gl.Disable(gl.TEXTURE_2D)
  gl.PushMatrix()
  gl.LoadIdentity()
  for i := range t.upright_positions {
    v := t.upright_positions[i]
    t.upright_drawables[i].Render(v.X, v.Y, v.Z, float32(t.zoom))
  }
  t.upright_positions = t.upright_positions[0:0]
  t.upright_drawables = t.upright_drawables[0:0]
  gl.PopMatrix()
}

func (t *Terrain) SetEventHandler(handler gin.EventHandler) {
  t.handler = handler
}

func (t *Terrain) HandleEventGroup(event_group gin.EventGroup) (bool, bool, *Node) {
  if t.handler != nil {
    t.handler.HandleEventGroup(event_group)
  }
  return false, false, nil
}

