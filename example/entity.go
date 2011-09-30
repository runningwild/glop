package main

import (
  "math"
  "fmt"
  "glop/sprite"
  "github.com/arbaal/mathgl"
)

type entity struct {
  // TODO: entity needs some sort of static data, like ap/round, etc...

  s *sprite.Sprite

  // in board coordinates per ms
  move_speed float32

  // Board coordinates of this entity's current position
  bx,by float32

  // number of ap remaining
  ap int

  // If the entity is currently moving then it will follow the vertices in path
  path [][2]int
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
  b = mathgl.Vec2{ e.bx, e.by }
  t = mathgl.Vec2{ float32(e.path[0][0]), float32(e.path[0][1]) }
  t.Subtract(&b)
  moved := t.Length()
  if moved <= 1e-5 {
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
  e.bx = b.X
  e.by = b.Y
  e.advance(dist - final_dist)
  facing := int(math.Atan2(float64(t.Y), float64(t.X)) / (2 * math.Pi) * 2 + 0.5)
  facing = (facing + 1) % 2
  fmt.Printf("Cur/Target: %d %d\n", e.s.StateFacing(), facing)
  if e.s.StateFacing() != facing {
    e.s.Command("stop")
    e.s.Command("turn_left")
    e.s.Command("move")
    fmt.Printf("post Cur/Target: %d %d\n", e.s.StateFacing(), facing)
  }
}

func (e *entity) Think(dt int64) {
  e.s.Think(dt)
  e.advance(e.move_speed * float32(dt))
}

