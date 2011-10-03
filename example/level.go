package main

import (
  "glop/gin"
  "glop/gui"
  "glop/util/algorithm"
  "gl"
  "math"
  "github.com/arbaal/mathgl"
  "json"
  "path/filepath"
  "io/ioutil"
  "os"
  "fmt"
)

type staticCellData struct {
  move_cost int
}
type cachedCellData struct {
  highlight Highlight
}

type CellData struct {
  staticCellData
  cachedCellData
}

func maxNormi(x,y,x2,y2 int) int {
  dx := x2 - x
  if dx < 0 { dx = -dx }
  dy := y2 - y
  if dy < 0 { dy = -dy }
  if dx > dy { return dx }
  return dy
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
  switch t.move_cost {
    case 1:
      r,g,b = 0.1, 0.9, 0.4
    case 5:
      r,g,b = 0.0, 0.7, 0.2
    case 10:
      r,g,b = 0.0, 0.0, 1.0
    case 2:
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

func (s *StaticLevelData) NumVertex() int {
  return len(s.grid) * len(s.grid[0])
}
func (s *StaticLevelData) fromVertex(v int) (int,int) {
  return v % len(s.grid), v / len(s.grid)
}
func (s *StaticLevelData) toVertex(x,y int) int {
  return x + y * len(s.grid)
}
func (l *Level) Adjacent(v int) ([]int, []float64) {
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
      if l.grid[x+dx][y+dy].move_cost <= 0 { continue }

      // Don't want to be able to walk through other units
      occupied := false
      for i := range l.entities {
        if int(l.entities[i].bx) == x+dx && int(l.entities[i].by) == y+dy {
          occupied = true
          break
        }
      }
      if occupied { continue }

      // Prevent moving along a diagonal if we couldn't get to that space normally via
      // either of the non-diagonal paths
      if dx != 0 && dy != 0 {
        if l.grid[x+dx][y].move_cost > 0 && l.grid[x][y+dy].move_cost > 0 {
          cost_a := float64(l.grid[x+dx][y].move_cost + l.grid[x][y+dy].move_cost) / 2
          cost_b := float64(l.grid[x+dx][y+dy].move_cost)
          adj_diag = append(adj, l.toVertex(x+dx, y+dy))
          weight_diag = append(weight, math.Fmax(cost_a, cost_b))
        }
      } else {
        if l.grid[x+dx][y+dy].move_cost > 0 {
          adj = append(adj, l.toVertex(x+dx, y+dy))
          weight = append(weight, float64(l.grid[x+dx][y+dy].move_cost))
        }
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

  // unset when the cache is cleared, lets Think() know it has to refil the cache
  cached bool

  // The gui element rendering the terrain and all of the other drawables
  terrain *gui.Terrain

  entities []*entity

  selected *entity

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

func (l *Level) clearCache() {
  if !l.cached { return }
  print("Clearing\n")
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
  l.command = Attack
  l.clearCache()

  reach := 10
  l.in_range = nil
  for i := range l.entities {
    if l.entities[i] == l.selected { continue }
    x := int(l.entities[i].bx)
    y := int(l.entities[i].by)
    x2 := int(l.selected.bx)
    y2 := int(l.selected.by)
    dist := maxNormi(x, y, x2, y2)
    if dist > reach { continue }
    l.in_range = append(l.in_range, l.toVertex(x, y))
  }
}
func (l *Level) DoAttack(target *entity) {
  if l.selected == nil { return }

  // First check range
  reach := 10
  x := int(target.bx)
  y := int(target.by)
  x2 := int(l.selected.bx)
  y2 := int(l.selected.by)
  dist := maxNormi(x, y, x2, y2)
  if dist > reach { return }

  l.selected.s.Command("ranged")
  target.s.Command("defend")
  target.s.Command("undamaged")

  l.clearCache()
  l.command = NoCommand
}
func (l *Level) PrepMove() {
  if l.selected == nil { return }
  l.command = Move
  l.clearCache()

  bx := int(l.selected.bx)
  by := int(l.selected.by)
  l.reachable = algorithm.ReachableWithinLimit(l, []int{ l.toVertex(bx, by) }, float64(l.selected.ap))
}

func (l *Level) DoMove(click_x,click_y int) {
  if l.selected == nil { return }

  start := l.toVertex(int(l.selected.bx), int(l.selected.by))
  end := l.toVertex(click_x, click_y)
  ap,path := algorithm.Dijkstra(l, []int{ start }, []int{ end })
  if len(path) == 0 || int(ap) > l.selected.ap { return }
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
  for i := range l.entities {
    e := l.entities[i]
    pbx := int(e.bx)
    pby := int(e.by)
    e.Think(dt)
    if pbx != int(e.bx) || pby != int(e.by) {
      l.clearCache()
    }
    l.terrain.AddUprightDrawable(e.bx + 0.25, e.by + 0.25, e.s)
  }

  l.maintainCommand()

  // Draw tiles
  for i := range l.grid {
    for j := range l.grid[i] {
      l.terrain.AddFlattenedDrawable(float32(i), float32(j), &l.grid[i][j])
    }
  }

  // Highlight the square under the cursor
  bx,by := l.terrain.WindowToBoard(l.winx, l.winy)
  mx := int(bx)
  my := int(by)
  if mx >= 0 && my >= 0 && mx < len(l.grid) && my < len(l.grid[0]) {
    cell := l.grid[mx][my]
    cell.highlight = MouseOver
    l.terrain.AddFlattenedDrawable(float32(mx), float32(my), &cell)
  }

  // Highlight selected entity
  if l.selected != nil {
    cell := l.grid[int(l.selected.bx)][int(l.selected.by)]
    cell.highlight = Reachable
    l.terrain.AddFlattenedDrawable(l.selected.bx, l.selected.by, &cell)
  }

  l.cached = true
}

func (l *Level) HandleEventGroup(event_group gin.EventGroup) {
  x,y := gin.In().GetKey(304).Cursor().Point()
  l.winx = x
  l.winy = y
  bx,by := l.terrain.WindowToBoard(x, y)

  if found,event := event_group.FindEvent(304); found && event.Type == gin.Press {
    click := mathgl.Vec2{ bx, by }

    var ent *entity
    var dist float32 = float32(math.Inf(1))
    for i := range l.entities {
      var cc mathgl.Vec2
      cc.Assign(&click)
      cc.Subtract(&mathgl.Vec2{ l.entities[i].bx + 0.5, l.entities[i].by + 0.5 })
      dx := cc.X
      if dx < 0 { dx = -dx }
      dy := cc.Y
      if dy < 0 { dy = -dy }
      d := float32(math.Fmax(float64(dx), float64(dy)))
      if d < dist {
        dist = d
        ent = l.entities[i]
      }
    }

    if l.selected == nil && dist < 3 {
      l.selected = ent
      l.clearCache()
      l.command = NoCommand
      return
    }

    var target *entity
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
  Image string
  Cells [][]levelDataCell
}

type levelDataContainer struct {
  Level levelData
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
      switch ldc.Level.Cells[i][j].Terrain {
        case "grass":
          level.grid[i][j].move_cost = 1
        case "brush":
          level.grid[i][j].move_cost = 5
        case "water":
          level.grid[i][j].move_cost = 10
        case "dirt":
          level.grid[i][j].move_cost = 2
        default:
          panic("WTF")
      }
    }
  }
  bg_path := filepath.Join(filepath.Clean(pathname), ldc.Level.Image)
  terrain,err := gui.MakeTerrain(bg_path, 100, dx, dy, 65)
  if err != nil {
    return nil, err
  }
  level.terrain = terrain
  terrain.SetEventHandler(&level)
  return &level, nil
}

