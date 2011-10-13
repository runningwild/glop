package game

import (
  "math"
  "fmt"
  "glop/gui"
  "glop/sprite"
  "github.com/arbaal/mathgl"
)

// contains the stats used to intialize a unit of this type
type UnitType struct {
  Name string

  Health int

  // map from Terrain name to the AP required for this unit to move into that terrain
  // any Terrain not named in this map is considered impassable by this unit
  Move_cost map[string]int

  AP int

  // basic combat stats, will be replaced with more interesting things later
  Attack int
  Defense int

  // List of the names of the weapons this unit comes with
  Weapons []string
}

type UnitStats struct {
  // Contains base stats before any modifier for this unit type
  Base    *UnitType
  Health  int
  AP      int
  Weapons []Weapon
}

type CosmeticStats struct {
  // in board coordinates per ms
  Move_speed float32
}

type EntityStatsWindow struct {
  gui.EmbeddedWidget
  gui.Rectangle
  gui.NonResponder
  gui.NonThinker

  table *gui.HorizontalTable
  ent   *Entity
}
func (e *Entity) MakeStatsWindow() *EntityStatsWindow {
  var esw EntityStatsWindow
  esw.EmbeddedWidget = &gui.BasicWidget{ CoreWidget : &esw }
  esw.ent = e
  esw.Dx = 200
  esw.Dy = 100

  esw.table = gui.MakeHorizontalTable()
  esw.table.AddChild(gui.MakeTextLine("standard", "Fixed", 1, 1, 1, 1), true)
  vert := gui.MakeVerticalTable()
  vert.AddChild(gui.MakeTextLine("standard", "F00", 1, 1, 1, 1), true)
  vert.AddChild(gui.MakeTextLine("standard", "B@r", 1, 1, 1, 1), true)
  vert.AddChild(gui.MakeTextLine("standard", "Wingerdinger", 1, 1, 1, 1), true)
  esw.table.AddChild(vert, false)

  return &esw
}
func (w *EntityStatsWindow) GetChildren() []gui.Widget {
  return []gui.Widget{ w.table }
}
func (w *EntityStatsWindow) Draw(region gui.Region) {
  w.Rectangle.Constrain(region)
  w.table.Draw(w.Bounds())
}

type Entity struct {
  UnitStats
  CosmeticStats

  s *sprite.Sprite

  level *Level

  // Board coordinates of this entity's current position
  pos mathgl.Vec2
  prev_pos mathgl.Vec2

  // If the entity is currently moving then it will follow the vertices in path
  path [][2]int
}

func (e *Entity) Coords() (x,y int) {
  return int(e.pos.X), int(e.pos.Y)
}

func (e *Entity) OnSetup() {
  e.Health = e.Base.Health
  e.AP = e.Base.AP
  e.prev_pos.Assign(&e.pos)
}
// On Turn is always called before OnRound
func (e *Entity) OnTurn() {
}
func (e *Entity) OnRound() {
  e.AP = e.Base.AP
}

func (e *Entity) enterCell(x,y int) {
  graph := unitGraph{ e.level, e.Base.Move_cost }
  src := e.level.toVertex(int(e.prev_pos.X), int(e.prev_pos.Y))
  dst := e.level.toVertex(x, y)
  e.AP -= int(graph.costToMove(src, dst))
  e.prev_pos.X = float32(x)
  e.prev_pos.Y = float32(y)
  if e.AP < 0 {
    // TODO: Log a warning
    e.AP = 0
  }
}

func (e *Entity) advance(dist float32) {
  if len(e.path) == 0 {
    if e.s.CurState() != "ready" {
      e.s.Command("stop")
    }
    return
  }
  if e.s.CurState() != "walk" {
    e.s.Command("move")
  }
  fmt.Printf("")
  if e.s.CurAnim() != "walk" { return }
  if dist <= 0 { return }
  var b,t mathgl.Vec2
  b = e.pos
  t = mathgl.Vec2{ float32(e.path[0][0]), float32(e.path[0][1]) }
  t.Subtract(&b)
  moved := t.Length()
  if moved <= 1e-5 {
    e.enterCell(e.path[0][0], e.path[0][1])
    e.path = e.path[1:]
    e.advance(dist - moved)
    return
  }
  final_dist := dist
  if final_dist > moved {
    final_dist = moved
  }
  t.Normalize()
  t.Scale(final_dist)
  b.Add(&t)
  e.pos.Assign(&b)

  if moved > dist {
    e.turnToFace(mathgl.Vec2{ float32(e.path[0][0]), float32(e.path[0][1]) })
  }

  e.advance(dist - final_dist)
}

func (e *Entity) turnToFace(target mathgl.Vec2) {
  target.Subtract(&e.pos)
  facing := math.Atan2(float64(target.Y), float64(target.X)) / (2 * math.Pi) * 360.0
  var face int
  if facing >= 22.5 || facing < -157.5 {
    face = 0
  } else {
    face = 1
  }
  if face != e.s.StateFacing() {
    e.s.Command("turn_left")
  }
}

func (e *Entity) Think(dt int64) {
  e.s.Think(dt)
  e.advance(e.Move_speed * float32(dt))
}

