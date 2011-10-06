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
  "sort"
  "github.com/arbaal/mathgl"
)

type zArray struct {
  vs        []mathgl.Vec3
  drawables []ZDrawable
}
func (za *zArray) Len() int {
  return len(za.vs)
}
func (za *zArray) Swap(i,j int) {
  za.vs[i],za.vs[j] = za.vs[j],za.vs[i]
  za.drawables[i],za.drawables[j] = za.drawables[j],za.drawables[i]
}
func (za *zArray) Less(i,j int) bool {
  return za.vs[i].Z > za.vs[j].Z
}

// Convenience function that sorts the elements in drawables and vs by decreasing
// order of the Z component of the vectors in vs
func ZSort(vs []mathgl.Vec3, drawables []ZDrawable) {
  sort.Sort(&zArray{vs,drawables})
}

// A ZDrawable is anything that can draw itself on an XY plane at a particular
// value for Z.
type ZDrawable interface {
  // Renders the drawable on the XY plane specified by z.  The values x and y
  // indicate an anchor point the the drawable can render itself relative to.
  Render(x,y,z,scale float32)
}

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

// Texture coordinates of a frame in a sprite sheet, as well as the anchor point
// for that frame.
type spriteRect struct {
  x,y,x2,y2     float32
  anch_x,anch_y float32

  // width and height of the rect in pixels
  dx,dy float32
}
type renderParams struct {
  index frameIndex
  x,y,z,scale float32
}
func (rp renderParams) params() (index frameIndex, x,y,z,scale float32) {
  return rp.index, rp.x, rp.y, rp.z, rp.scale
}
type spriteLevel struct {
  // indexes[i] is the frame of animation the corresponds to the image from filenames[i]
  indexes   []frameIndex
  filenames []string

  // when the spriteLevel is loaded a sprite sheet is generated using all of the images
  // listed in filenames.  rects is a map from the frameIndexes in indexes to the region
  // in the sprite sheet corresponding to that frameIndex.
  rects   map[int]spriteRect

  // Texture that holds the sprite sheet when this spriteLevel is loaded
  texture gl.Texture

  // number of Load() calls - number of Unload() calls.  When this reaches
  // zero it will free its data
  count   int

  // channels used to manage the loading/unloading/rendering in a safe way
  load_count     chan int
  render         chan renderParams

  // Loaded when load() is called, but not actually bound to a texture until the
  // first time it is rendered.
  sheet *image.RGBA
}

// TODO: runtime.SetFinalizer()!!
func makeSpriteLevel(indexes []frameIndex, filenames []string) *spriteLevel {
  var sl spriteLevel
  sl.indexes = make([]frameIndex, len(indexes))
  copy(sl.indexes, indexes)
  sl.filenames = make([]string, len(filenames))
  copy(sl.filenames, filenames)
  sl.load_count = make(chan int)
  sl.render = make(chan renderParams)
  go sl.routine()
  return &sl
}

// all sprite level commands go through here so that we can call a load early and check
// back later for the results in a nice thread-safe fashion
func (sl *spriteLevel) routine() {
  for {
    select {
      case count := <-sl.load_count:
        prev := sl.count
        sl.count += count
        if (prev == 0) != (sl.count == 0) {
          if sl.count == 0 {
            sl.unload()
          } else {
            sl.load()
          }
        }
        if sl.count < 0 {
          panic("Cannot unload a sprite level more times than it is loaded.")
        }

      case data := <-sl.render:
        sl.renderToQuad(data.params())
        sl.render <- renderParams{}
    }
  }
}

func (sl *spriteLevel) Load() {
  sl.load_count <- 1
}

func (sl *spriteLevel) Unload() {
  sl.load_count <- -1
}

// TODO: Might want to have the load part happen in a separate go-routine so we don't block
// here if we're loading a lot of textures.  In that case we should have a default sprite or
// something that displays if a spriteLevel isn't available yet.
func (sl *spriteLevel) load() {
  sl.rects = make(map[int]spriteRect)

  var images []image.Image
  for i := range sl.indexes {
    filename := sl.filenames[i]
    file,err := os.Open(filename)
    if err != nil {
      panic(fmt.Sprintf("Unable to load texture '%s': %s", filename, err.String()))
      return
    }
    im,_,err := image.Decode(file)
    file.Close()
    if err != nil {
      panic(fmt.Sprintf("Unable to decode texture '%s': %s", filename, err.String()))
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

  sheet := image.NewRGBA(image.Rect(0, 0, pdx, pdy))
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
      x : float32(cx) / float32(pdx),
      y : 0,
      x2 : float32(cx + bounds.Dx()) / float32(pdx),
      y2 : float32(bounds.Dy()) / float32(pdy),
      anch_x : 0.5,
      anch_y : 0.0,
      dx : float32(bounds.Dx()),
      dy : float32(bounds.Dy()),
    }
    sl.rects[sl.indexes[i].Int()] = rect
    cx += bounds.Dx()
  }

  // This assignment will be visible in future calls to renderToQuad because
  // spriteLevel.routine() explicitly synchronizes these calls.
  sl.sheet = sheet
}
func (sl *spriteLevel) unload() {
  sl.rects = nil
  sl.texture.Delete()
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
func (s *spriteLevel) RenderToQuad(index frameIndex, x,y,z,scale float32) {
  s.render <- renderParams{ index, x, y, z, scale }
  <-s.render
}
func (s *spriteLevel) renderToQuad(index frameIndex, x,y,z,scale float32) {
  if s.sheet != nil {
    gl.Enable(gl.TEXTURE_2D)
    s.texture = gl.GenTexture()
    s.texture.Bind(gl.TEXTURE_2D)
    gl.TexEnvf(gl.TEXTURE_ENV, gl.TEXTURE_ENV_MODE, gl.MODULATE)
    gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
    gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
    gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
    gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
    pdx := int(nextPowerOf2(uint32(s.sheet.Bounds().Dx())))
    pdy := int(nextPowerOf2(uint32(s.sheet.Bounds().Dy())))
    glu.Build2DMipmaps(gl.TEXTURE_2D, 4, pdx, pdy, gl.RGBA, s.sheet.Pix)
    s.sheet = nil
  }
  if s.texture == 0 { return }
  gl.Enable(gl.TEXTURE_2D)
  gl.Enable(gl.BLEND)
  gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
  s.texture.Bind(gl.TEXTURE_2D)
  gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
  gl.Color4d(1.0, 1.0, 1.0, 1.0)
  rect := s.rects[index.Int()]
  x1 := x - scale * rect.anch_x * rect.dx
  x2 := x + scale * (1 - rect.anch_x) * rect.dx
  y1 := y - scale * rect.anch_y * rect.dy
  y2 := y + scale * (1 - rect.anch_y) * rect.dy
  gl.Begin(gl.QUADS)
    gl.TexCoord2f(rect.x, rect.y2)
    gl.Vertex3f(x1, y1, z)
    gl.TexCoord2f(rect.x, rect.y)
    gl.Vertex3f(x1, y2, z)
    gl.TexCoord2f(rect.x2, rect.y)
    gl.Vertex3f(x2, y2, z)
    gl.TexCoord2f(rect.x2, rect.y2)
    gl.Vertex3f(x2, y1, z)
  gl.End()
//  gl.Disable(gl.BLEND)
//  gl.Disable(gl.TEXTURE_2D)
}

// Data that can be shared between two different instance of the same Sprite
// sharedSprite is *NOT* thread-safe.
type sharedSprite struct {
  anim,state *Graph

  indexes   []frameIndex
  filenames []string

  // connection always stays loaded so that we can always transition between facings
  connection *spriteLevel

  // facing is reloaded whenever the sprite changes facing, it doesn't load anything that is
  // in connection, though
  facing *spriteLevel

  // map from facing number to spriteLevel for that facing.  These spriteLevels do not contain
  // any of the frames that are loaded into connections
  facings []*spriteLevel

  // the number of possible facings
  num_facings int

  // maps a command to the node ids of any node that comes immediately after that command
  cmd_target map[string][]int
}

type Sprite struct {
  *sharedSprite

  // the facing number for the animation, always in the range [0, num_facings)
  anim_facing int

  // the facing number for the state, always in the range [0, num_facings]
  state_facing int

  // The current frame of animation that is displayed when Render is called
  cur_frame *Node

  // The current state based on all commands that have been received so far.  The
  // current animation frame may lag behind this.
  cur_state *Node

  // Ms remaining on this frame
  togo      int64

  // If len(path) == 0 then this is the animation sequence that must be followed.  This is set
  // whenever a command is given to the sprite so that the command can be followed as quickly as
  // possible with no chance of doing any idle animations first
  path []int

  pending_cmds []string
}

func (s *Sprite) Render(x,y,z,scale float32) {
  f := frameIndex{
    facing : uint8(s.anim_facing),
    anim_id : uint16(s.cur_frame.Id),
  }
  if _,ok := s.connection.rects[f.Int()]; ok {
    s.connection.renderToQuad(f, x, y, z, scale)
  } else {
    s.facings[s.anim_facing].renderToQuad(f, x, y, z, scale)
  }
}

type SpriteManager struct {
  loaded_sprites map[string]*sharedSprite
}
var spriteManager *SpriteManager
func init() {
  spriteManager = MakeSpriteManager()
}

func MakeSpriteManager() *SpriteManager {
  sm := new(SpriteManager)
  sm.loaded_sprites = make(map[string]*sharedSprite)
  return sm
}

func LoadSprite(path string) (*Sprite, os.Error) {
  return spriteManager.LoadSprite(path)
}

func (sm *SpriteManager) LoadSprite(path string) (*Sprite, os.Error) {
  if _,ok := sm.loaded_sprites[path]; !ok {
    ss := new(sharedSprite)
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

    ss.anim = anim_graph
    ss.state = state_graph

    face_count := make(map[int]bool)
    walker := func(the_path string, info *os.FileInfo, err os.Error) os.Error {
      if the_path == path { return nil }
      if info.IsRegular() { return nil }
      _,final := filepath.Split(the_path)
      num,err := strconv.Atoi(final)
      if err != nil {
        return err
      }
      face_count[num] = true
      return filepath.SkipDir
    }
    filepath.Walk(path, walker)
    valid := true
    if len(face_count) == 0 {
      valid = false
    }
    for f := range face_count {
      if f < 0 || f >= len(face_count) {
        valid = false
      }
    }
    if !valid {
      return nil, os.NewError("Sprite facing directories not set up properly")
    }
    ss.num_facings = len(face_count)

    for facing := 0; facing < ss.num_facings; facing++ {
      for _,node := range anim_graph.nodes {
        fi := frameIndex{
          facing : uint8(facing),
          anim_id : uint16(node.Id),
        }
        ss.indexes = append(ss.indexes, fi)
        full_path := filepath.Join(path, fmt.Sprintf("%d", facing), node.Name + ".png")
        ss.filenames = append(ss.filenames, full_path)
      }
    }

    // TODO: Figure out how much should be loaded in connection
    // Right now we're just taking the frames on either end of a facing change and keeping those
    // permanently loaded.
    mids := make(map[int]bool)
    for _,node := range anim_graph.nodes {
      for _,edge := range node.Edges {
        if edge.Facing == 0 { continue }
        mids[edge.Source] = true
        mids[edge.Target] = true
      }
    }
    var indexes []frameIndex
    var filenames []string
    for i := range ss.indexes {
      if _,ok := mids[int(ss.indexes[i].anim_id)]; ok {
        indexes = append(indexes, ss.indexes[i])
        filenames = append(filenames, ss.filenames[i])
      }
    }

    ss.connection = makeSpriteLevel(indexes, filenames)
    ss.connection.Load()

    ss.facings = make([]*spriteLevel, ss.num_facings)
    for i := range ss.facings {
      var indexes []frameIndex
      var filenames []string
      for j := range ss.indexes {
        if int(ss.indexes[j].facing) != i { continue }
        if _,ok := ss.connection.rects[ss.indexes[j].Int()]; ok { continue }
        indexes = append(indexes, ss.indexes[j])
        filenames = append(filenames, ss.filenames[j])
      }
      ss.facings[i] = makeSpriteLevel(indexes, filenames)
    }

    ss.cmd_target = make(map[string][]int)
    for _,node := range ss.anim.nodes {
      for _,edge := range node.Edges {
        if edge.State == "" { continue }
        if _,ok := ss.cmd_target[edge.State]; !ok {
          ss.cmd_target[edge.State] = make([]int, 0, 1)
        }
        ss.cmd_target[edge.State] = append(ss.cmd_target[edge.State], edge.Target)
      }
    }
    sm.loaded_sprites[path] = ss
  }

  var sprite Sprite
  sprite.sharedSprite = sm.loaded_sprites[path]

  sprite.facings[0].Load()
  sprite.cur_frame = sprite.anim.StartNode()
  sprite.cur_state = sprite.state.StartNode()
  sprite.togo = sprite.cur_frame.Time
  return &sprite, nil
}

func (s *Sprite) CurAnim() string {
  return s.cur_frame.State
  return fmt.Sprintf("%d: %s -> %v", s.anim_facing, s.cur_frame.Name, s.pending_cmds)
}
func (s *Sprite) CurState() string {
  return s.cur_state.Name
}

func (s *Sprite) Command(cmd string) {
  edge := s.cur_state.FindEdge(cmd)
  if edge == nil { return }
  s.cur_state = s.state.nodes[edge.Target]
  if edge.Facing != 0 {
    s.state_facing = (s.state_facing + edge.Facing + s.num_facings) % s.num_facings
    s.facings[s.state_facing].Load()
  }
  if cmd != "" {
    s.pending_cmds = append(s.pending_cmds, cmd)
    s.Command("")
  }
}

func (s *Sprite) StateFacing() int {
  return s.state_facing
}

func (s *Sprite) AnimFacing() int {
  return s.anim_facing
}

// TODO: WRITE COMMENTS!
func (s *Sprite) Think(dt int64) {
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
    edge := s.cur_frame.FindEdge("")
    if edge != nil {
      s.cur_frame = s.anim.nodes[edge.Target]
      s.togo = s.cur_frame.Time
    }
  }
  for _,edge := range prev.Edges {
    if edge.Source == prev.Id && edge.Target == s.cur_frame.Id {
      if edge.Facing != 0 {
        // We are currently using the connections spriteSheet, so it's ok to unload the facings
        // spriteSheet
//        s.facings[s.anim_facing].Unload()
        s.anim_facing = (s.anim_facing + edge.Facing + s.num_facings) % s.num_facings
//        s.facings[s.anim_facing].Load()
        break
      }
    }
  }
  s.Think(dt)
}

func (s *sharedSprite) Stats() {
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
