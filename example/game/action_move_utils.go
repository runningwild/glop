package game

// Subtracts the AP cost of moving into the next cell from the Entity's 
// available AP.  Returns false if the Entity didn't have enough AP.
func payForMove(ent *Entity, dst BoardPos) bool {
  level := ent.level
  graph := unitGraph{level, ent.Base.attributes.MoveMods}
  cost := int(graph.costToMove(ent.pos.Vertex(), dst.Vertex()))
  if cost > ent.AP {
    return false
  }
  ent.AP -= cost
  return true
}

func AdvanceEntity(ent *Entity, path *[]BoardPos, dt int64) bool {
  if len(*path) == 0 { return false }
  dst := (*path)[0]
  tomove := ent.Move_speed * float32(dt)
  for tomove > 0 {
    moved,reached := ent.Advance(dst, tomove)
    if moved == 0 && !reached { return false }
    tomove -= moved

    // Check to see if the Entity has made it to a new cell
    if reached {
      ent.OnEntry()
      ent.level.GetCellAtPos(dst).highlight &= ^Reachable
      *path = (*path)[1:]

      // If we have reached our destination *OR* if something has happened and
      // we no longer have the AP required to continue moving then this action
      // is complete - so we return true
      if len(*path) == 0 || !payForMove(ent, (*path)[0]) {
        ent.Advance(BoardPos{}, 0)
        return true
      }
      dst = (*path)[0]
    }
  }
  return false
}
