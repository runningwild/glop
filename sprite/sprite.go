package sprite

import (
  "os"
  "fmt"
  "strconv"
  "glop/util/algorithm"
  "path/filepath"
  "image"
  _ "image/png"
  "gl"
  "gl/glu"
  "runtime"
  "rand"
)

// frameIndexes are used so that we can have maps that are keyed on the pair
// of facing and anim_id
type frameIndex struct {
  facing  uint8
  anim_id uint16
}
func (f frameIndex) Int() int {
  n := int(f.anim_id)
  n = n | (int(f.facing) << 16)
  return n
}
func makeFrameIndex(facing,anim_id int) frameIndex {
  return frameIndex{ facing:uint8(facing), anim_id:uint16(anim_id) }
}

// Texture coordinates of a frame in a sprite sheet
type spriteRect struct {
  x,y,x2,y2 float64
}
type spriteLevel struct {
  rects   map[int]spriteRect
  texture gl.Texture
}
func releaseSpriteLevel(s *spriteLevel) {
  s.texture.Delete()
}
func makeSpriteLevel() *spriteLevel {
  s := new(spriteLevel)
  s.texture = gl.GenTexture()
  runtime.SetFinalizer(s, releaseSpriteLevel)
  return s
}

// TODO: This was copied from the gui package, probably should just have some basic
// texture loading utils that do this common stuff
func nextPowerOf2(n uint32) uint32 {
  if n == 0 { return 1 }
  for i := uint(0); i < 32; i++ {
    p := uint32(1) << i
    if n <= p { return p }
  }
  return 0
}
func (s *spriteLevel) RenderToQuad(index frameIndex) {
  gl.MatrixMode(gl.MODELVIEW)
  gl.PushMatrix()
  defer gl.PopMatrix()

  gl.Enable(gl.TEXTURE_2D)
  gl.Enable(gl.BLEND)
  gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
  s.texture.Bind(gl.TEXTURE_2D)
  gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
  gl.Color4d(1.0, 1.0, 1.0, 1.0)
  rect := s.rects[index.Int()]
  gl.Begin(gl.QUADS)
    gl.TexCoord2d(rect.x, rect.y2)
    gl.Vertex2d(100, 100)
    gl.TexCoord2d(rect.x, rect.y)
    gl.Vertex2d(100, 400)
    gl.TexCoord2d(rect.x2, rect.y)
    gl.Vertex2d(400, 400)
    gl.TexCoord2d(rect.x2, rect.y2)
    gl.Vertex2d(400, 100)
  gl.End()
  gl.Disable(gl.BLEND)
  gl.Disable(gl.TEXTURE_2D)
}

func (s *spriteLevel) Loaded(index frameIndex) bool {
  _,ok := s.rects[index.Int()]
  return ok
}

// Clears out the frames that were loaded in this level and replaces them with new frames
func (s *spriteLevel) Reload(indexes []frameIndex, filenames []string) {
  s.rects = make(map[int]spriteRect)

  var images []image.Image
  for i := range indexes {
    filename := filenames[i]
    file,err := os.Open(filename)
    if err != nil {
      // TODO: LOG IT!
      return
    }
    defer file.Close()
    im,_,err := image.Decode(file)
    if err != nil {
      // TODO: LOG IT!
      return
    }
    images = append(images, im)
  }

  dx := 0
  dy := 0
  for _,im := range images {
    bounds := im.Bounds()
    if bounds.Dy() > dy {
      dy = bounds.Dy()
    }
    dx += bounds.Dx()
  }
  pdx := int(nextPowerOf2(uint32(dx)))
  pdy := int(nextPowerOf2(uint32(dy)))

  sheet := image.NewRGBA(pdx, pdy)
  cx := 0
  for i := range images {
    // blit the image onto the sheet
    bounds := images[i].Bounds()
    for y := 0; y < bounds.Dy(); y++ {
      for x := 0; x < bounds.Dx(); x++ {
        r,g,b,a := images[i].At(x,y).RGBA()
        base := 4*(x+cx) + sheet.Stride*y
        sheet.Pix[base] = uint8(r)
        sheet.Pix[base+1] = uint8(g)
        sheet.Pix[base+2] = uint8(b)
        sheet.Pix[base+3] = uint8(a)
      }
    }
    rect := spriteRect{
      x : float64(cx) / float64(pdx),
      y : 0,
      x2 : float64(cx + bounds.Dx()) / float64(pdx),
      y2 : float64(bounds.Dy()) / float64(pdy),
    }
    s.rects[indexes[i].Int()] = rect
    cx += bounds.Dx()
  }

  gl.Enable(gl.TEXTURE_2D)
  s.texture.Bind(gl.TEXTURE_2D)
  gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
  gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
  glu.Build2DMipmaps(gl.TEXTURE_2D, 4, pdx, pdy, gl.RGBA, sheet.Pix)
}

type Sprite struct {
  anim,state *Graph

  indexes   []frameIndex
  filenames []string

  // connection always stays loaded so that we can always transition between facings
  connection *spriteLevel

  // facing is reloaded whenever the sprite changes facing, it doesn't load anything that is
  // in connection, though
  facing *spriteLevel

  // the number of possible facings
  num_facings int

  // the facing number, always in the range [0, num_facings)
  cur_facing int

  // The current frame of animation that is displayed when Render is called
  cur_frame *Node

  // Ms remaining on this frame
  togo      int

  // If len(path) == 0 then this is the animation sequence that must be followed.  This is set
  // whenever a command is given to the sprite so that the command can be followed as quickly as
  // possible with no chance of doing any idle animations first
  path []int

  pending_cmds []string

  // maps a command to the node ids of any node that comes immediately after that command
  cmd_target map[string][]int
}

func (s *Sprite) RenderToQuad() {
  f := frameIndex{
    facing : uint8(s.cur_facing),
    anim_id : uint16(s.cur_frame.Id),
  }
  if s.connection.Loaded(f) {
    s.connection.RenderToQuad(f)
  } else {
    if s.facing.Loaded(f) {
      s.facing.RenderToQuad(f)
    } else {
      panic("Dammit")
    }
  }
}

type facingVisitor struct {
  base  string
  count map[int]bool
}
func (f *facingVisitor) VisitDir(path string, _ *os.FileInfo) bool {
  if path == f.base {
    return true
  }
  _,final := filepath.Split(path)
  num,err := strconv.Atoi(final)
  if err != nil {
    return false
  }
  if f.count == nil {
    f.count = make(map[int]bool)
  }
  f.count[num] = true
  return false
}
func (f *facingVisitor) VisitFile(_ string, _ *os.FileInfo) {}
func (f *facingVisitor) Valid() bool {
  for v := range f.count {
    if v < 0 { return false }
    if v >= len(f.count) { return false }
  }
  return true
}
func (f *facingVisitor) Count() int {
  return len(f.count)
}

func LoadSprite(path string) (*Sprite, os.Error) {
  anim_path := filepath.Join(path, "anim.xgml")
  state_path := filepath.Join(path, "state.xgml")
  anim_graph,err := LoadGraph(anim_path)
  if err != nil { return nil,err }
  state_graph,err := LoadGraph(state_path)
  if err != nil { return nil,err }
  ProcessAnimWithState(anim_graph, state_graph)
  image_names := make(map[string]bool)
  for _,n := range anim_graph.nodes {
    image_names[n.Name + ".png"] = true
  }

  sprite := Sprite{
    anim : anim_graph,
    state : state_graph,
  }

  f := facingVisitor {
    base : path,
  }
  filepath.Walk(path, &f, nil)
  if !f.Valid() || f.Count() == 0 {
    return nil, os.NewError("Sprite facing directories not set up properly")
  }
  sprite.num_facings = f.Count()

  for facing := 0; facing < sprite.num_facings; facing++ {
    for _,node := range anim_graph.nodes {
      fi := frameIndex{
        facing : uint8(facing),
        anim_id : uint16(node.Id),
      }
      sprite.indexes = append(sprite.indexes, fi)
      full_path := filepath.Join(path, fmt.Sprintf("%d", facing), node.Name + ".png")
      sprite.filenames = append(sprite.filenames, full_path)
    }
  }

  // TODO: Figure out how much should be loaded in connection
  // Right now we're just taking the frames on either end of a facing change and keeping those
  // permanently loaded.
  mids := make(map[int]bool)
  for i := range anim_graph.edges {
    if anim_graph.edges[i].Facing != 0 {
      mids[anim_graph.edges[i].Source] = true
      mids[anim_graph.edges[i].Target] = true
    }
  }
  var indexes []frameIndex
  var filenames []string
  for i := range sprite.indexes {
    if _,ok := mids[int(sprite.indexes[i].anim_id)]; ok {
      indexes = append(indexes, sprite.indexes[i])
      filenames = append(filenames, sprite.filenames[i])
    }
  }
  sprite.connection = makeSpriteLevel()
  sprite.connection.Reload(indexes, filenames)
  sprite.facing = makeSpriteLevel()
  sprite.loadFacing(0)
  sprite.cur_frame = sprite.anim.StartNode()
  sprite.togo = sprite.cur_frame.Time

  sprite.cmd_target = make(map[string][]int)
  for _,edge := range sprite.anim.edges {
    if edge.State == "" { continue }
    if _,ok := sprite.cmd_target[edge.State]; !ok {
      sprite.cmd_target[edge.State] = make([]int, 0, 1)
    }
    sprite.cmd_target[edge.State] = append(sprite.cmd_target[edge.State], edge.Target)
  }
  return &sprite,nil
}

func (s *Sprite) loadFacing(facing int) {
  var indexes []frameIndex
  var filenames []string
  for i := range s.indexes {
    if int(s.indexes[i].facing) != facing { continue }
    indexes = append(indexes, s.indexes[i])
    filenames = append(filenames, s.filenames[i])
  }
  s.facing.Reload(indexes, filenames)
}

func (s *Sprite) CurState() string {
  return fmt.Sprintf("%d: %s", s.cur_facing, s.cur_frame.Name)
}

func (s *Sprite) Command(cmd string) {
  if _,ok := s.cmd_target[cmd]; !ok {
    // TODO: Log a warning, can't call this cmd on this sprite
    return
  }
  s.pending_cmds = append(s.pending_cmds, cmd)
}

// Advances the TODO: WRITE COMMENTS!
func (s *Sprite) Think(dt int) {
  // We might call Think(0) within think just to rerun some logic, but dt < -1 is crazy
  if dt < 0 {
    return
    // TODO: Log a warning?
  }
  if s.togo > dt {
    s.togo -= dt
    return
  }
  dt -= s.togo

fmt.Printf("-------------------\n")
fmt.Printf("Cur: %v\n", s.cur_frame.Name)
fmt.Printf("Cmd: %v\n", s.pending_cmds)
fmt.Printf("Path: %v\n", s.path)
fmt.Printf("PathNames: ")
for _,p := range s.path {
  fmt.Printf("%s ", s.anim.nodes[p].Name)
}

  // If we don't have a path layed out but we do have pending commands we should used one
  // of those to get a new path
  if len(s.path) == 0 && len(s.pending_cmds) > 0 {
    cg := &commandGraph{
      Graph : s.anim,
      cmd : s.pending_cmds[0],
    }
    for len(s.pending_cmds) > 0 {
      var t float64
      t,s.path = algorithm.Dijkstra(cg, []int{s.cur_frame.Id}, s.cmd_target[s.pending_cmds[0]])
      if t < 0 {
        // TODO: Log a warning, got a command that we can't actually handle
        s.pending_cmds = s.pending_cmds[1:]
      } else {
        s.pending_cmds = s.pending_cmds[1:]
        break
      }
    }
    if len(s.path) > 0 {
      s.path = s.path[1:]
      if len(s.path) == 0 {
        s.path = nil
      } else {
        // TODO: If we queue a command just after we finished another instance of that same
        //       command we will *not* try to find a path because the shortest path is just
        //       to stay still.  Instead we will just wait until we've moved one frame, but
        //       this means that we will run dijkstra's multiple times for no reason, might
        //       be worth caching that or something.
        if len(s.pending_cmds) == 0 {
          s.pending_cmds = nil
        }
      }
    }
  }

  // TODO: Can't have self-edges in the animation graph, but *can* have them in the state graph
  prev := s.cur_frame
  if len(s.path) > 0 {
    s.cur_frame = s.anim.nodes[s.path[0]]
    s.path = s.path[1:]
    s.togo = s.cur_frame.Time
  } else {
    var edges []Edge
    weight := 0.0
    for _,edge := range s.cur_frame.Edges {
      if edge.State != "" { continue }
      edges = append(edges, edge)
      weight += edge.Weight
    }
    hit := rand.Float64() * weight
    sum := 0.0
    for _,edge := range edges {
      sum += edge.Weight
      if hit < sum {
        s.cur_frame = s.anim.nodes[edge.Target]
        s.togo = s.cur_frame.Time
        break
      }
    }
  }
  for _,edge := range prev.Edges {
    if edge.Source == prev.Id && edge.Target == s.cur_frame.Id {
      if edge.Facing != 0 {
        s.cur_facing = (s.cur_facing + edge.Facing + s.num_facings) % s.num_facings
        s.loadFacing(s.cur_facing)
        break
      }
    }
  }
  s.Think(dt)
}

func (s *Sprite) Stats() {
  for i := range s.state.nodes {
    source_state := s.state.nodes[i].Name
    for j := range s.state.nodes[i].Edges {
      target_state := s.state.nodes[s.state.nodes[i].Edges[j].Target].Name
      var source_anim,target_anim []int
      for i := range s.anim.nodes {
        if s.anim.nodes[i].State == source_state {
          source_anim = append(source_anim, i)
        }
        if s.anim.nodes[i].State == target_state && s.anim.nodes[i].IsCore() {
          target_anim = append(target_anim, i)
        }
      }
      var time float64
      var path []int
      for _,start := range source_anim {
        t,p := algorithm.Dijkstra(s.anim, []int{ start }, target_anim)
        if t > time {
          time = t
          path = p
        }
      }
      fmt.Printf("%s -> %s\n", source_state, target_state)
      fmt.Printf("time: %f\n", time)
      for i := range path {
        fmt.Printf("%d: %s\n", i, s.anim.nodes[path[i]].Name)
      }
      fmt.Printf("\n")
    }
  }
}
