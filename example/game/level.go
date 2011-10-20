package game

import (
  "glop/gin"
  "glop/gui"
  "glop/util/algorithm"
  "glop/sprite"
  "gl"
  "math"
  "github.com/arbaal/mathgl"
  "json"
  "path/filepath"
  "io/ioutil"
  "os"
  "fmt"
)

type Terrain interface {
  Name() string
}

type Grass struct {}
func (t Grass) Name() string {
  return "grass"
}

type Dirt struct {}
func (t Dirt) Name() string {
  return "dirt"
}

type Water struct {}
func (t Water) Name() string {
  return "water"
}

type Brush struct {}
func (t Brush) Name() string {
  return "brush"
}

var terrain_registry map[string]Terrain
func init() {
  terrain_registry = make(map[string]Terrain)
  RegisterTerrain(Water{})
  RegisterTerrain(Dirt{})
  RegisterTerrain(Grass{})
  RegisterTerrain(Brush{})
}
func RegisterTerrain(t Terrain) {
  _,ok := terrain_registry[t.Name()]
  if ok {
    panic(fmt.Sprintf("Tried to register the terrain '%s' more than once.", t.Name()))
  }
  terrain_registry[t.Name()] = t
}
func MakeTerrain(name string) Terrain {
  t,ok := terrain_registry[name]
  if !ok {
    panic(fmt.Sprintf("Cannot load the unregistered terrain '%s'.", name))
  }
  return t
}


type staticCellData struct {
  Terrain
}
type cachedCellData struct {
  highlight Highlight
}

type CellData struct {
  staticCellData
  cachedCellData
}

type Highlight int
const (
  None Highlight = iota
  Reachable
  Attackable
  MouseOver
//  Impassable
//  OutOfRange
)

func (t *CellData) Clear() {
  t.cachedCellData = cachedCellData{}
}

func (t *CellData) Render(x,y,z,scale float32) {
  gl.Disable(gl.TEXTURE_2D)
  gl.Enable(gl.BLEND)
  gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
  gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
  var r,g,b,a float32
  a = 0.2
  switch t.Terrain.Name() {
    case "grass":
      r,g,b = 0.1, 0.9, 0.4
    case "brush":
      r,g,b = 0.0, 0.7, 0.2
    case "water":
      r,g,b = 0.0, 0.0, 1.0
    case "dirt":
      r,g,b = 0.6, 0.5, 0.3
    default:
      r,g,b = 1,0,0
  }
  x *= scale
  y *= scale
  gl.Color4f(r, g, b, a)
  gl.Begin(gl.QUADS)
    gl.Vertex3f(x, y, z)
    gl.Vertex3f(x, y+scale, z)
    gl.Vertex3f(x+scale, y+scale, z)
    gl.Vertex3f(x+scale, y, z)
  gl.End()

  if t.highlight != None {
    switch t.highlight {
      case Reachable:
        r,g,b,a = 0, 0.2, 0.9, 0.3
      case Attackable:
        r,g,b,a = 0.7, 0.2, 0.2, 0.9
      case MouseOver:
        r,g,b,a = 0.1, 0.9, 0.2, 0.4
      default:
        panic("Unknown highlight")
    }
    gl.Color4f(r, g, b, a)
    gl.Begin(gl.QUADS)
      gl.Vertex3f(x, y, z)
      gl.Vertex3f(x, y+scale, z)
      gl.Vertex3f(x+scale, y+scale, z)
      gl.Vertex3f(x+scale, y, z)
    gl.End()
  }
}

// Contains everything about a level that is stored on disk
type StaticLevelData struct {
  grid [][]CellData
}
type unitGraph struct {
  *Level
  move_cost map[string]int
}
func (s *StaticLevelData) NumVertex() int {
  return len(s.grid) * len(s.grid[0])
}
func (s *StaticLevelData) fromVertex(v int) (int,int) {
  return v % len(s.grid), v / len(s.grid)
}
func (s *StaticLevelData) toVertex(x,y int) int {
  return x + y * len(s.grid)
}

// Assumes that src and dst are adjacent
func (l unitGraph) costToMove(src,dst int) float64 {
  x,y := l.fromVertex(src)
  x2,y2 := l.fromVertex(dst)

  cost_c,ok := l.move_cost[l.grid[x2][y2].Terrain.Name()]
  if !ok { return -1 }
  if x == x2 || y == y2 {
    return float64(cost_c)
  }

  cost_a,ok := l.move_cost[l.grid[x][y2].Terrain.Name()]
  if !ok { return - 1 }
  cost_b,ok := l.move_cost[l.grid[x2][y].Terrain.Name()]
  if !ok { return - 1 }

  cost_ab := float64(cost_a + cost_b) / 2
  return math.Fmax(cost_ab, float64(cost_c))
}
func (l *unitGraph) Adjacent(v int) ([]int, []float64) {
  x,y := l.fromVertex(v)
  var adj []int
  var weight []float64

  // separate arrays for the adjacent diagonal cells, this way we make sure they are listed
  // at the end so that searches will prefer orthogonal adjacent cells
  var adj_diag []int
  var weight_diag []float64

  for dx := -1; dx <= 1; dx++ {
    if x + dx < 0 || x + dx >= len(l.grid) { continue }
    for dy := -1; dy <= 1; dy++ {
      if dx == 0 && dy == 0 { continue }
      if y + dy < 0 || y + dy >= len(l.grid[0]) { continue }

      // Don't want to be able to walk through other units
      occupied := false
      for i := range l.Entities {
        if int(l.Entities[i].pos.X) == x+dx && int(l.Entities[i].pos.Y) == y+dy {
          occupied = true
          break
        }
      }
      if occupied { continue }

      // Prevent moving along a diagonal if we couldn't get to that space normally via
      // either of the non-diagonal paths
      cost := l.costToMove(l.toVertex(x,y), l.toVertex(x+dx, y+dy))
      if cost < 0 { continue }
      if dx != 0 && dy != 0 {
        adj_diag = append(adj_diag, l.toVertex(x+dx, y+dy))
        weight_diag = append(weight_diag, cost)
      } else {
        adj = append(adj, l.toVertex(x+dx, y+dy))
        weight = append(weight, cost)
      }
    }
  }
  for i := range adj_diag {
    adj = append(adj, adj_diag[i])
    weight = append(weight, weight_diag[i])
  }
  return adj,weight
}

type Command int
const (
  NoCommand Command = iota
  Move
  Attack
)

// Contains everything for the playing of the game
type Level struct {
  StaticLevelData

  editor *Editor

  // unset when the cache is cleared, lets Think() know it has to refil the cache
  cached bool

  // The single gui element containing all other elements related to the
  // game
  game_gui *gui.HorizontalTable

  // The gui element rendering the terrain and all of the other drawables
  Terrain *gui.Terrain

  // The gui elements that show entity information
  selected_gui *EntityStatsWindow
  targeted_gui *EntityStatsWindow

  Entities []*Entity

  selected *Entity
  hovered  *Entity

  // window coords of the mouse
  winx,winy int


  // The most recently prepped valid command
  command Command


  // MOVE data

  // If a unit is selected this will hold the list of cells that are reachable
  // from that unit's position within its allotted AP
  reachable []int

  // ATTACK data
  in_range []int
}
func (l *Level) GetSelected() *Entity {
  return l.selected
}
func (l *Level) GetHovered() *Entity {
  return l.hovered
}
func (l *Level) Round() {
  for i := range l.Entities {
    l.Entities[i].OnRound()
  }
}

func (l *Level) Setup() {
  for i := range l.Entities {
    l.Entities[i].OnSetup()
  }
}

func (l *Level) clearCache() {
  if !l.cached { return }
  for i := range l.grid {
    for j := range l.grid[i] {
      l.grid[i][j].Clear()
    }
  }
  l.cached = false
}

func (l *Level) maintainCommand() {
  switch l.command {
    case Move:
    for _,v := range l.reachable {
      x,y := l.fromVertex(v)
      l.grid[x][y].highlight = Reachable
    }

    case Attack:
    for _,v := range l.in_range {
      x,y := l.fromVertex(v)
      l.grid[x][y].highlight = Attackable
    }

    case NoCommand:
    default:
      panic(fmt.Sprintf("Unknown command: %d", l.command))
  }
}

func (l *Level) PrepAttack() {
  if l.selected == nil { return }
  weapon := l.selected.Weapons[0]
  if weapon.Cost(l.selected) > l.selected.AP { return }

  l.in_range = nil
  for i := range l.Entities {
    if l.Entities[i] == l.selected { continue }
    if !weapon.InRange(l.selected, l.Entities[i]) { continue }
    l.in_range = append(l.in_range, l.toVertex(l.Entities[i].Coords()))
  }

  if len(l.in_range) == 0 { return }

  l.command = Attack
  l.clearCache()
}
func (l *Level) DoAttack(target *Entity) {
  if l.selected == nil { return }

  // First check range
  weapon := l.selected.Weapons[0]
  cost := weapon.Cost(l.selected)
  if cost > l.selected.AP { return }
  l.selected.AP -= cost

  // TODO: Should probably log a warning, this shouldn't have been able to happen
  if !weapon.InRange(l.selected, target) { return }

  res := weapon.Damage(l.selected, target)

  // Resolve the actual attack here
  l.selected.turnToFace(target.pos)

  x := int(target.pos.X)
  y := int(target.pos.Y)
  x2 := int(l.selected.pos.X)
  y2 := int(l.selected.pos.Y)
  dist := maxNormi(x, y, x2, y2)

  if dist > 2 {
    l.selected.s.Command("ranged")
  } else {
    l.selected.s.Command("melee")
  }
  target.s.Command("defend")

  if res.Connect == Hit {
    target.Health -= res.Damage.Piercing
    if target.Health <= 0 {
      target.s.Command("killed")
    } else {
      target.s.Command("damaged")
    }
  } else {
    target.s.Command("undamaged")
  }

  l.clearCache()
  l.command = NoCommand
}
func (l *Level) PrepMove() {
  if l.selected == nil { return }

  bx := int(l.selected.pos.X)
  by := int(l.selected.pos.Y)
  graph := &unitGraph{ l, l.selected.Base.Move_cost }
  l.reachable = algorithm.ReachableWithinLimit(graph, []int{ l.toVertex(bx, by) }, float64(l.selected.AP))

  if len(l.reachable) == 0 { return }

  l.command = Move
  l.clearCache()
}

func (l *Level) DoMove(click_x,click_y int) {
  if l.selected == nil { return }

  start := l.toVertex(int(l.selected.pos.X), int(l.selected.pos.Y))
  end := l.toVertex(click_x, click_y)
  graph := &unitGraph{ l, l.selected.Base.Move_cost }
  ap,path := algorithm.Dijkstra(graph, []int{ start }, []int{ end })
  fmt.Printf("Found a path: %f %v\n", ap, path)
  if len(path) == 0 || int(ap) > l.selected.AP { return }
  path = path[1:]
  l.selected.path = l.selected.path[0:0]
  l.reachable = nil
  for i := range path {
    x,y := l.fromVertex(path[i])
    l.selected.path = append(l.selected.path, [2]int{x,y})
  }

  l.clearCache()
  l.command = NoCommand
}

func (l *Level) Think(dt int64) {
  // Draw all sprites
  for i := range l.Entities {
    e := l.Entities[i]
    pbx := int(e.pos.X)
    pby := int(e.pos.Y)
    e.Think(dt)
    if pbx != int(e.pos.X) || pby != int(e.pos.Y) {
      l.clearCache()
    }
    l.Terrain.AddUprightDrawable(e.pos.X + 0.25, e.pos.Y + 0.25, e.s)
  }

  l.maintainCommand()

  // Draw tiles
  for i := range l.grid {
    for j := range l.grid[i] {
      l.Terrain.AddFlattenedDrawable(float32(i), float32(j), &l.grid[i][j])
    }
  }

  // Highlight the square under the cursor
  bx,by := l.Terrain.WindowToBoard(l.winx, l.winy)
  mx := int(bx)
  my := int(by)
  if mx >= 0 && my >= 0 && mx < len(l.grid) && my < len(l.grid[0]) {
    cell := l.grid[mx][my]
    cell.highlight = MouseOver
    l.Terrain.AddFlattenedDrawable(float32(mx), float32(my), &cell)
    l.hovered = nil
    for i := range l.Entities {
      x,y := l.Entities[i].Coords()
      if x == mx && y == my {
        l.hovered = l.Entities[i]
      }
    }
  }

  // Highlight selected entity
  if l.selected != nil {
    cell := l.grid[int(l.selected.pos.X)][int(l.selected.pos.Y)]
    cell.highlight = Reachable
    l.Terrain.AddFlattenedDrawable(l.selected.pos.X, l.selected.pos.Y, &cell)
  }
  l.selected_gui.SetEntity(l.selected)
  l.targeted_gui.SetEntity(l.hovered)

  l.cached = true
}

func (l *Level) HandleEventGroup(event_group gin.EventGroup) {
  x,y := gin.In().GetKey(gin.MouseLButton).Cursor().Point()
  l.winx = x
  l.winy = y
  bx,by := l.Terrain.WindowToBoard(x, y)

  if found,event := event_group.FindEvent(gin.MouseLButton); found && event.Type == gin.Press {
    click := mathgl.Vec2{ bx, by }

    var ent *Entity
    var dist float32 = float32(math.Inf(1))
    for i := range l.Entities {
      var cc mathgl.Vec2
      cc.Assign(&click)
      cc.Subtract(&mathgl.Vec2{ l.Entities[i].pos.X + 0.5, l.Entities[i].pos.Y + 0.5 })
      dx := cc.X
      if dx < 0 { dx = -dx }
      dy := cc.Y
      if dy < 0 { dy = -dy }
      d := float32(math.Fmax(float64(dx), float64(dy)))
      if d < dist {
        dist = d
        ent = l.Entities[i]
      }
    }

    if l.selected == nil && dist < 3 {
      l.selected = ent
      l.clearCache()
      l.command = NoCommand
      return
    }

    var target *Entity
    if l.selected != nil && dist < 0.5 {
      target = ent
    }

    switch l.command {
      case Move:
        if target == nil {
          l.DoMove(int(bx), int(by))
        }

      case Attack:
        if target != nil {
          l.DoAttack(target)
          target = nil
        }

      case NoCommand:
      default:
    }

    if target != nil {
      l.selected = target
    }
    l.clearCache()
    l.command = NoCommand
  }
}

type levelDataCell struct {
  Terrain string
}
type levelData struct {
  // TOOD: Need to track the file this came from so we can copy it when
  // we save
  Image string

  Cells [][]levelDataCell
}

type levelDataContainer struct {
  Level levelData
}

func (l *Level) SaveLevel(pathname string) os.Error {
  out,err := os.Create(pathname)
  if err != nil { return nil }
  defer out.Close()
  var ldc levelDataContainer
  ldc.Level.Image = "fudgecake.png"
  ldc.Level.Cells = make([][]levelDataCell, len(l.grid))
  for i := range ldc.Level.Cells {
    ldc.Level.Cells[i] = make([]levelDataCell, len(l.grid[0]))
  }
  for i := range ldc.Level.Cells {
    for j := range ldc.Level.Cells[i] {
      ldc.Level.Cells[i][j].Terrain = l.grid[i][j].Name()
    }
  }
  data,err := json.Marshal(&ldc)
  if err != nil { return err }
  _,err = out.Write(data)
  return err
}

func LoadLevel(pathname string) (*Level, os.Error) {
  datapath := filepath.Join(filepath.Clean(pathname), "data.json")
  datafile,err := os.Open(datapath)
  if err != nil {
    return nil, err
  }
  data,err := ioutil.ReadAll(datafile)
  if err != nil {
    return nil, err
  }
  var ldc levelDataContainer
  json.Unmarshal(data, &ldc)

  var level Level
  dx := len(ldc.Level.Cells)
  dy := len(ldc.Level.Cells[0])
  all_cells := make([]CellData, dx*dy)
  level.grid = make([][]CellData, dx)
  for i := range level.grid {
    level.grid[i] = all_cells[i*dy : (i+1)*dy]
  }
  for i := range level.grid {
    for j := range level.grid[i] {
      level.grid[i][j].Terrain = MakeTerrain(ldc.Level.Cells[i][j].Terrain)
    }
  }
  bg_path := filepath.Join(filepath.Clean(pathname), ldc.Level.Image)
  terrain,err := gui.MakeTerrain(bg_path, 100, dx, dy, 65)
  if err != nil {
    return nil, err
  }
  level.Terrain = terrain
  terrain.SetEventHandler(&level)

  level.editor = MakeEditor()
  level.game_gui = gui.MakeHorizontalTable()
  game_only_gui := gui.MakeVerticalTable()
  level.selected_gui = MakeStatsWindow()
  level.targeted_gui = MakeStatsWindow()
  entity_guis := gui.MakeHorizontalTable()
  entity_guis.AddChild(level.selected_gui)
  entity_guis.AddChild(level.targeted_gui)
  game_only_gui.AddChild(level.Terrain)
  game_only_gui.AddChild(entity_guis)
  level.game_gui.AddChild(game_only_gui)
  level.game_gui.AddChild(level.editor.GetGui())
  return &level, nil
}

func (l *Level) GetGui() gui.Widget {
  return l.game_gui
}

func (l *Level) ToggleEditor() {
  l.editor.ToggleGui()
}

func (l *Level) AddEntity(unit_type UnitType, x,y int, move_speed float32, sprite *sprite.Sprite) *Entity  {
  ent := &Entity{
    UnitStats : UnitStats {
      Base : &unit_type,
    },
    pos : mathgl.Vec2{ float32(x), float32(y) },
    s : sprite,
    level : l,
    CosmeticStats : CosmeticStats{
      Move_speed : move_speed,
    },
  }
  for _,name := range unit_type.Weapons {
    ent.Weapons = append(ent.Weapons, MakeWeapon(name))
  }
  l.Entities = append(l.Entities, ent)
  return ent
}