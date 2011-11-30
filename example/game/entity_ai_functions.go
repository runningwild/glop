package game

import (
  "glop/ai"
  "fmt"
  "polish"
  "reflect"
)

func init() {
  fmt.Printf("")
}

func AddEntityContext(ent *Entity, context *polish.Context) {
  // These functions are self-explanitory, they are all relative to the
  // current entity
  context.AddFunc("numVisibleEnemies",
      func() int {
        return ent.numVisibleEntities(false)
      })
  context.AddFunc("numVisibleAllies",
      func() int {
        return ent.numVisibleEntities(true)
      })
  context.AddFunc("nearestEnemy",
      func() *Entity {
        return ent.nearestEntity(false)
      })
  context.AddFunc("nearestAlly",
      func() *Entity {
        return ent.nearestEntity(true)
      })

  // Returns the distance between two entities, does not take into account
  // movement penalties.
  context.AddFunc("distBetween", distBetween)

  // Attacks a target entity
  context.AddFunc("attack",
      func(t *Entity) {
        ent.attack(t)
      })

  // Advances the entity as far as possible towards another target entity
  context.AddFunc("advanceTowards",
      func(t *Entity) {
        ent.advanceTowards(t)
      })

  // Query an indidual entity's stats.  Depending on the game it may or may not
  // be legal to query certain stats on an enemy.
  context.AddFunc("curHealth",
      func(t *Entity) int {
        return t.CurHealth()
      })
  context.AddFunc("maxHealth",
      func(t *Entity) int {
        return t.BaseHealth()
      })
  context.AddFunc("curAttack",
      func(t *Entity) int {
        return t.CurAttack()
      })
  context.AddFunc("maxAttack",
      func(t *Entity) int {
        return t.BaseAttack()
      })
  context.AddFunc("curDefense",
      func(t *Entity) int {
        return t.CurDefense()
      })
  context.AddFunc("maxDefense",
      func(t *Entity) int {
        return t.BaseDefense()
      })

  // Ends an entity's turn
  context.AddFunc("done",
      func() {
        ent.aig.Term() <- ai.TermError
      })

  // This entity, the one currently taking its turn
  context.SetValue("me", ent)
}

func (e *Entity) numVisibleEntities(ally bool) int {
  count := 0
  for v,_ := range e.visible {
    ent := e.level.GetCellAtVertex(v).ent
    if ent != nil && (ent.side == e.side) == ally {
      count++
    }
  }
  return count
}

func (e *Entity) nearestEntity(ally bool) *Entity {
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
    if (ent.side == e.side) != ally { continue }
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

type aiEvalSignal int
const (
  aiEvalCont aiEvalSignal = iota
  aiEvalTerm
  aiEvalPause
)

func (e *Entity) doCmd(f func() bool) {
  e.cmds <- f
  cont := <-e.cont
  fmt.Printf("cont val: %d\n", cont)
  switch cont {
    case aiEvalCont:
    fmt.Printf("aiEvalCont\n")

    case aiEvalTerm:
    fmt.Printf("ai.Term...\n")
    e.aig.Term() <- ai.TermError
    fmt.Printf("ai.Term sent\n")

    case aiEvalPause:
    fmt.Printf("ai.Interrupt...\n")
    e.aig.Term() <- ai.InterruptError
    fmt.Printf("ai.Interrupt sent\n")
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
