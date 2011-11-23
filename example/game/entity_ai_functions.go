package game

import (
  "polish"
)

// All functions available in AiGraphs are here, even if they are one-liners,
// just for the sake of completeness.  The functions are added to the
// polish.Context exactly as they are defined here.  Methods defined on *Entity
// here are added to the polish.Context with the calling *Entity defined as the
// *Entity that the Ai is operating on, so it does not need to be passed.

func AddEntityContext(ent *Entity, context *polish.Context) {
  context.AddFunc("numVisibleEnemies", func() int { return ent.numVisibleEnemies() })
  context.AddFunc("nearestEnemy", func() *Entity { return ent.nearestEnemy() })
  context.AddFunc("distBetween", distBetween)
  context.AddFunc("attack", func(t *Entity) { ent.attack(t) })
  context.AddFunc("advanceTowards", func(t *Entity) { ent.advanceTowards(t) })
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
  
}

func (e *Entity) advanceTowards(target *Entity) {
  
}
