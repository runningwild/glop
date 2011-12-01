package game

import "game/base"

// Subtracts the AP cost of moving into the next cell from the Entity's 
// available AP.  Returns false if the Entity didn't have enough AP.
func payForMove(ent *Entity, dst BoardPos) bool {
  level := ent.level
  graph := unitGraph{level, ent}
  cost := int(graph.costToMove(ent.pos.Vertex(ent.level), dst.Vertex(ent.level)))
  if cost > ent.CurAp() {
    return false
  }
  ent.SpendAp(cost)
  return true
}

func canPayForMove(ent *Entity, dst BoardPos) bool {
  level := ent.level
  graph := unitGraph{level, ent}
  cost := int(graph.costToMove(ent.pos.Vertex(ent.level), dst.Vertex(ent.level)))
  return cost <= ent.CurAp()
}

func AdvanceEntity(ent *Entity, path *[]BoardPos, dt int64) bool {
  if len(*path) == 0 { return true }
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

// Returns all Entitys that are not on the same side as the src Entity, are
// within rng of the src Entity, and are visible to the src Entity.
func getEntsWithinRange(src *Entity, rng int, level *Level) []*Entity {
  var targets []*Entity
  for _,ent := range level.Entities {
    if ent.Side == src.Side { continue }
    dist := base.MaxNormi(src.pos.Xi(), src.pos.Yi(), ent.pos.Xi(), ent.pos.Yi())
    if _,ok := src.visible[ent.pos.Vertex(ent.level)]; !ok { continue }
    if dist > rng { continue }
    targets = append(targets, ent)
  }
  return targets
}

// If there is an entity at bx,by and that entity is in targets, then returns
// that entity, otherwise returns nil.
func findTargetOnClick(bx,by float64, level *Level, targets map[*Entity]bool) *Entity {
  bp := MakeBoardPos(int(bx), int(by))
  t := level.GetCellAtPos(bp).ent
  if _,ok := targets[t]; ok {
    return t
  }
  return nil
}

func findMultipleUniqueTargets(bx,by float64, level *Level, targets,marks *map[*Entity]bool, count int) bool {
  t := findTargetOnClick(bx, by, level, *targets)
  if _,ok := (*marks)[t]; ok {
    return true
  }
  if t != nil {
    (*marks)[t] = true
    level.GetCellAtPos(MakeBoardPos(int(bx),int(by))).highlight |= Targeted
  }
  return len(*marks) == count
}