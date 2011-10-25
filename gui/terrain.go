package gui

import (
  "fmt"
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

func init() {
  fmt.Printf("")
}

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
  EmbeddedWidget
  BasicZone
  NonThinker
  NonFocuser


  // Length of the side of block in the source image.
  block_size int

  // All events received by the terrain are passed to the handler
  handler gin.EventHandler

  // Focus, in map coordinates
  fx,fy float32

  // The viewing angle, 0 means the map is viewed head-on, 90 means the map is viewed
  // on its edge (i.e. it would not be visible)
  angle float32

  // Zoom factor, 1.0 is standard
  zoom float32

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

  // Don't need to keep the image around once it's loaded into texture memory,
  // only need to keep around the dimensions
  bg_dims Dims
  texture gl.Texture
}

func (t *Terrain) String() string {
  return "terrain"
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

func MakeTerrain(bg_path string, block_size,dx,dy int, angle float32) (*Terrain, os.Error) {
  var t Terrain
  t.EmbeddedWidget = &BasicWidget{ CoreWidget : &t }

  f,err := os.Open(bg_path)
  if err != nil {
    return nil, err
  }
  defer f.Close()
  bg,_,err := image.Decode(f)
  if err != nil {
    return nil, err
  }

  t.block_size = block_size
  t.angle = angle

  t.bg_dims.Dx = bg.Bounds().Dx()
  t.bg_dims.Dy = bg.Bounds().Dy()
  rgba := image.NewRGBA(t.bg_dims.Dx, t.bg_dims.Dy)
  draw.Draw(rgba, bg.Bounds(), bg, image.Point{0,0}, draw.Over)

  gl.Enable(gl.TEXTURE_2D)
  t.texture = gl.GenTexture()
  t.texture.Bind(gl.TEXTURE_2D)
  gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
  glu.Build2DMipmaps(gl.TEXTURE_2D, 4, t.bg_dims.Dx, t.bg_dims.Dy, gl.RGBA, rgba.Pix)

  if err != nil {
    return nil,err
  }
  t.zoom = 1.0

  t.makeMat()
  t.Request_dims.Dx = 100
  t.Request_dims.Dy = 100
  t.Ex = true
  t.Ey = true
  return &t, nil
}

func (t *Terrain) makeMat() {
  var m mathgl.Mat4
  t.mat.Translation(float32(t.Render_region.Dx/2 + t.Render_region.X), float32(t.Render_region.Dy/2 + t.Render_region.Y), 0)
  m.RotationZ(45 * math.Pi / 180)
  t.mat.Multiply(&m)
  m.RotationAxisAngle(mathgl.Vec3{ X : -1, Y : 1}, -float32(t.angle) * math.Pi / 180)
  t.mat.Multiply(&m)

  s := float32(t.zoom)
  m.Scaling(s, s, s)
  t.mat.Multiply(&m)

  // Move the terrain so that (t.fx,t.fy) is at the origin, and hence becomes centered
  // in the window
  xoff := (t.fx + 0.5) * float32(t.block_size)
  yoff := (t.fy + 0.5) * float32(t.block_size)
  m.Translation(-xoff, -yoff, 0)
  t.mat.Multiply(&m)

  t.imat.Assign(&t.mat)
  t.imat.Inverse()
}

// Transforms a cursor position in window coordinates to board coordinates.  Does not check
// to make sure that the values returned represent a valid position on the board.
func (t *Terrain) WindowToBoard(wx,wy int) (float32, float32) {
  mx := float32(wx)
  my := float32(wy)
  return t.modelviewToBoard(mx, my)
}

func (t *Terrain) modelviewToBoard(mx,my float32) (float32,float32) {
  mz := (my - float32(t.Render_region.Y + t.Render_region.Dy/2)) * float32(math.Tan(float64(t.angle * math.Pi / 180)))
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

func clamp(f,min,max float32) float32 {
  if f < min { return min }
  if f > max { return max }
  return f
}

// The change in x and y screen coordinates to apply to point on the terrain the is in
// focus.  These coordinates will be scaled by the current zoom.
func (t *Terrain) Move(dx,dy float64) {
  if dx == 0 && dy == 0 { return }
  dy /= math.Sin(float64(t.angle) * math.Pi / 180)
  dx,dy = dy+dx, dy-dx
  t.fx += float32(dx) / t.zoom
  t.fy += float32(dy) / t.zoom
  t.fx = clamp(t.fx, 0, float32(t.bg_dims.Dx / t.block_size))
  t.fy = clamp(t.fy, 0, float32(t.bg_dims.Dy / t.block_size))
  t.makeMat()
}

// Changes the current zoom from e^(zoom) to e^(zoom+dz)
func (t *Terrain) Zoom(dz float64) {
  if dz == 0 { return }
  exp := math.Log(float64(t.zoom)) + dz
  exp = float64(clamp(float32(exp), -1.25, 1.25))
  t.zoom = float32(math.Exp(exp))
  t.makeMat()
}

func (t *Terrain) Draw(region Region) {
  region.PushClipPlanes()
  defer region.PopClipPlanes()
  if t.Render_region.X != region.X || t.Render_region.Y != region.Y || t.Render_region.Dx != region.Dx || t.Render_region.Dy != region.Dy {
    t.Render_region = region
    t.makeMat()
  }
  gl.MatrixMode(gl.MODELVIEW)
  gl.PushMatrix()
  gl.LoadIdentity()
  gl.MultMatrixf(&t.mat[0])
  defer gl.PopMatrix()

  gl.Disable(gl.DEPTH_TEST)
  gl.Disable(gl.TEXTURE_2D)
  gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
  gl.Color3d(1, 0, 0)
  gl.Enable(gl.BLEND)
  gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
  fdx := float32(t.bg_dims.Dx)
  fdy := float32(t.bg_dims.Dy)

  // Draw a simple border around the terrain
  gl.Color4d(1,.3,.3,1)
  gl.Begin(gl.QUADS)
    fbs := float32(t.block_size)
    gl.Vertex2f(   -fbs,    -fbs)
    gl.Vertex2f(   -fbs, fdy+fbs)
    gl.Vertex2f(fdx+fbs, fdy+fbs)
    gl.Vertex2f(fdx+fbs,    -fbs)
  gl.End()

  gl.Enable(gl.TEXTURE_2D)
  t.texture.Bind(gl.TEXTURE_2D)
  gl.Color4d(1.0, 1.0, 1.0, 1.0)
  gl.Begin(gl.QUADS)
    gl.TexCoord2f(0, 0)
    gl.Vertex2f(0, 0)
    gl.TexCoord2f(0, -1)
    gl.Vertex2f(0, fdy)
    gl.TexCoord2f(1, -1)
    gl.Vertex2f(fdx, fdy)
    gl.TexCoord2f(1, 0)
    gl.Vertex2f(fdx, 0)
  gl.End()

  gl.Disable(gl.TEXTURE_2D)
  gl.Color4f(0,0,0, 0.5)
  gl.Begin(gl.LINES)
  for i := float32(0); i < float32(t.bg_dims.Dx); i += float32(t.block_size) {
    gl.Vertex2f(i, 0)
    gl.Vertex2f(i, float32(t.bg_dims.Dy))
  }
  for j := float32(0); j < float32(t.bg_dims.Dy); j += float32(t.block_size) {
    gl.Vertex2f(0, j)
    gl.Vertex2f(float32(t.bg_dims.Dx), j)
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

func (t *Terrain) DoRespond(event_group EventGroup) (bool,bool) {
  if t.handler != nil {
    t.handler.HandleEventGroup(event_group.EventGroup)
  }
  return false,false
}

