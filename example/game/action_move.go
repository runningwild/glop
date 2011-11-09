package game

import "glop/util/algorithm"

type ActionMove struct {
  basicIcon
  ent       *Entity

  reachable []int
  path      []int
}

func makeMoveAction(ent *Entity) Action {
  return &ActionMove{ ent : ent }
}

func (a *ActionMove) Prep() bool {
  level := a.ent.level
  bx := int(a.ent.pos.X)
  by := int(a.ent.pos.Y)
  graph := &unitGraph{level, a.ent.Base.attributes.MoveMods}
  reachable := algorithm.ReachableWithinLimit(graph, []int{level.toVertex(bx, by)}, float64(a.ent.AP))

  if len(reachable) == 0 {
    return false
  }

  a.reachable = reachable

  // Since this is a valid action we can go ahead and highlight all of the
  // tiles that the unit can move to
  for _,v := range a.reachable {
    x,y := level.fromVertex(v)
    level.grid[x][y].highlight |= Reachable
  }

  return true
}

func (a *ActionMove) Cancel() {
  a.reachable = nil
  a.path = nil
  a.ent.level.clearCache(Reachable)
}

func (a *ActionMove) MouseOver(bx,by float64) {
  // TODO: Might want to highlight the specific path that would be taken if
  // the user clicked here
}

func (a *ActionMove) MouseClick(bx,by float64) bool {
  level := a.ent.level
  dst := level.toVertex(int(bx), int(by))
  found := false
  for _,v := range a.reachable {
    if dst == v {
      found = true
      break
    }
  }
  if !found { return false }

  src := level.toVertex(int(a.ent.pos.X), int(a.ent.pos.Y))
  graph := &unitGraph{level, a.ent.Base.attributes.MoveMods}
  ap, path := algorithm.Dijkstra(graph, []int{src}, []int{dst})
  if len(path) <= 1 || int(ap) > a.ent.AP {
    return false
  }
  a.path = path[1:]
  a.reachable = nil

  level.clearCache(Reachable)
  for _,v := range a.path {
    x,y := level.fromVertex(v)
    level.grid[x][y].highlight |= Reachable
  }
  if !a.payForMove() {
    a.path = nil
    return false
  }
  return true
}

// Subtracts the AP cost of moving into the next cell from the entity's 
// available AP.  Returns false if the entity didn't have enough AP.
func (a *ActionMove) payForMove() bool {
  level := a.ent.level
  graph := unitGraph{level, a.ent.Base.attributes.MoveMods}
  src := level.toVertex(int(a.ent.pos.X), int(a.ent.pos.Y))
  cost := int(graph.costToMove(src, a.path[0]))
  if cost > a.ent.AP {
    return false
  }
  a.ent.AP -= cost
  return true
}

func (a *ActionMove) Maintain(dt int64) bool {
  if len(a.path) == 0 { return false }
  pos := a.ent.level.MakeBoardPosFromVertex(a.path[0])
  tomove := a.ent.Move_speed * float32(dt)
  for tomove > 0 {
    moved,reached := a.ent.Advance(pos, tomove)
    if moved == 0 && !reached { return false }
    tomove -= moved

    // Check to see if the entity has made it to a new cell
    if reached {
      a.ent.OnEntry()
      dst := a.ent.level.MakeBoardPosFromVertex(a.path[0])
      a.ent.level.GetCellAtPos(dst).highlight &= ^Reachable
      a.path = a.path[1:]

      // If we have reached our destination *OR* if something has happened and
      // we no longer have the AP required to continue moving then this action
      // is complete - so we return true
      if len(a.path) == 0 || !a.payForMove() {
        a.Cancel()
        a.ent.Advance(BoardPos{}, 0)
        return true
      }
      pos = a.ent.level.MakeBoardPosFromVertex(a.path[len(a.path) - 1])
    }
  }
  return false
}

