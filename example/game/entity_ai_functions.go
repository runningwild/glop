package game

import (
  "fmt"
  "polish"
  "reflect"
)

func init() {
  fmt.Printf("")
}

func AddEntityContext(ent *Entity, context *polish.Context) {
  context.AddFunc("numVisibleEnemies", func() int { return ent.numVisibleEnemies() })
  context.AddFunc("nearestEnemy", func() *Entity { return ent.nearestEnemy() })
  context.AddFunc("distBetween", distBetween)
  context.AddFunc("attack", func(t *Entity) { ent.attack(t) })
  context.AddFunc("advanceTowards", func(t *Entity) { ent.advanceTowards(t) })
  context.AddFunc("done", func() { ent.done = true })
  context.SetValue("me", ent)
}

func (e *Entity) numVisibleEnemies() int {
  count := 0
  for v,_ := range e.visible {
    ent := e.level.GetCellAtVertex(v).ent
    if ent != nil && ent.side != e.side {
      count++
    }
  }
  return count
}

func (e *Entity) nearestEnemy() *Entity {
  var nearest *Entity
  for v,_ := range e.visible {
    ent := e.level.GetCellAtVertex(v).ent
    if ent == nil { continue }
    if ent.side == e.side { continue }
    if nearest == nil {
      nearest = ent
    } else if ent.pos.Dist(e.pos) < nearest.pos.Dist(e.pos) {
      nearest = ent
    }
  }
  return nearest
}

func distBetween(e1,e2 *Entity) int {
  return e1.pos.Dist(e2.pos)
}

func (e *Entity) attack(target *Entity) {
  panic("done")
}

func (e *Entity) getAction(typ reflect.Type) Action {
  for _,action := range e.actions {
    if reflect.TypeOf(action) == typ {
      return action
    }
  }
  return nil
}

func (e *Entity) doCmd(f func() bool) {
  e.cmds <- f
  if !(<-e.cont) {
    e.aig.Term() <- true    
  }
  fmt.Printf("Command complete\n")
}

func (e *Entity) advanceTowards(target *Entity) {
  if target == nil {
    panic("No target")
  }
  var move *ActionMove
  move = e.getAction(reflect.TypeOf(&ActionMove{})).(*ActionMove)
  if move == nil {
    panic("couldn't find the move action")
  }

  e.doCmd(func() bool {
    // TODO: This preamble should be in a level method
    if move.aiMoveToWithin(target.pos.Xi(), target.pos.Yi(), 1) == StandardAction {
      e.level.current_action = move
      e.level.mid_action = true
      return true
    }
    return false
  })
}
