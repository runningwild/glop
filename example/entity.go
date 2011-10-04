package main

import (
  "math"
  "fmt"
  "glop/sprite"
  "github.com/arbaal/mathgl"
  "rand"
)

type Terrain int
const(
  Grass Terrain = iota
  Dirt
  Water
  Brush
)

type Damage struct {
  Piercing int
  Smashing int
  Fire     int
}

type Connect int
const(
  Hit Connect = iota
  Miss
  Dodge
)
type Resolution struct {
  Connect Connect
  Damage  Damage
}

type Weapon interface {
  Reach() int
  Cost() int
  Damage(source,target *entity) Resolution
}

type Bayonet struct {}
func (b *Bayonet) Reach() int {
  return 2
}
func (b *Bayonet) Cost() int {
  return 5
}
func (b *Bayonet) Damage(source,target *entity) Resolution {
  mod := rand.Intn(10)
  if source.Base.Attack + mod > target.Base.Defense {
    amt := source.Base.Attack + mod - target.Base.Defense - 2
    if amt <= 0 {
      return Resolution {
        Connect : Dodge,
      }
    } else {
      return Resolution {
        Connect : Hit,
        Damage : Damage {
          Piercing : amt,
        },
      }
    }
  }
  return Resolution {
    Connect : Miss,
  }
}

type Rifle struct {
  Range int
  Power int
}
func (r *Rifle) Reach() int {
  return r.Range
}
func (r *Rifle) Cost() int {
  return 12
}
func (r *Rifle) Damage(source,target *entity) Resolution {
  dist := maxNormi(int(source.pos.X), int(source.pos.Y), int(target.pos.X), int(target.pos.Y))
  acc := r.Range - dist
  if rand.Intn(acc) == 0 {
    return Resolution {
      Connect : Miss,
    }
  }

  if rand.Intn(target.Base.Defense) > source.Base.Attack + r.Power {
    return Resolution {
      Connect : Dodge,
    }
  }

  return Resolution {
    Connect : Hit,
    Damage : Damage {
      Piercing : r.Power,
    },
  }
}


// contains the stats used to intialize a unit of this type
type UnitType struct {
  Name string

  Health int

  // map from Terrain to the AP required for this unit to move into that terrain
  // any Terrain not in this map is considered impassable by this unit
  Move_cost map[Terrain]int

  AP int

  // basic combat stats, will be replaced with more interesting things later
  Attack int
  Defense int

  Weapons []Weapon
}

type UnitStats struct {
  // Contains base stats before any modifier for this unit type
  Base   *UnitType
  Health int
  AP     int
}

type CosmeticStats struct {
  // in board coordinates per ms
  Move_speed float32
}

type entity struct {
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


func (e *entity) OnSetup() {
  e.Health = e.Base.Health
  e.AP = e.Base.AP
  e.prev_pos.Assign(&e.pos)
}
// On Turn is always called before OnRound
func (e *entity) OnTurn() {
}
func (e *entity) OnRound() {
  e.AP = e.Base.AP
}

func (e *entity) enterCell(x,y int) {
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

func (e *entity) advance(dist float32) {
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

func (e *entity) turnToFace(target mathgl.Vec2) {
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

func (e *entity) Think(dt int64) {
  e.s.Think(dt)
  e.advance(e.Move_speed * float32(dt))
}

