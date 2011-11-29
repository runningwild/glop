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
  context.AddFunc("done", func() { ent.aig.Term() <- true })
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

  // This is so that we always iterate through the visible vertices in the same
  // order
  visible := make([]int, 0, len(e.visible))
  for v := range e.visible {
    visible = append(visible, v)
  }

  for _,v := range visible {
    ent := e.level.GetCellAtVertex(v).ent
    if ent == nil { continue }
    if ent.side == e.side { continue }
    if ent.CurHealth() <= 0 { continue }
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
  if target == nil {
    panic("No target")
  }
  var att *ActionBasicAttack
  att = e.getAction(reflect.TypeOf(&ActionBasicAttack{})).(*ActionBasicAttack)
  if att == nil {
    panic("couldn't find an attack action")
  }

  e.doCmd(func() bool {
    // TODO: This preamble should be in a level method
    if att.aiDoAttack(target) {
      e.level.pending_action = att
      e.level.mid_action = true
      return true
    }
    return false
  })
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
    if move.aiMoveToWithin(target.pos.Xi(), target.pos.Yi(), 1) {
      e.level.pending_action = move
      e.level.mid_action = true
      return true
    }
    return false
  })
}
