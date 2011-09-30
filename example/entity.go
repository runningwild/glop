package main

import (
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
      e.s.Command([]string{"stop"})
    }
    return
  }
  if e.s.CurState() != "walk" {
    e.s.Command([]string{"move"})
  }
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

  t.Normalize()
  t.Scale(dist)
  b.Add(&t)
  e.bx = b.X
  e.by = b.Y
  e.advance(dist - moved)
}

func (e *entity) Think(dt int64) {
  e.s.Think(dt)
  e.advance(e.move_speed * float32(dt))
}

