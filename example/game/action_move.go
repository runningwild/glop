package game

import "glop/util/algorithm"

func init() {
  registerActionType("move", &ActionMove{})
}
type ActionMove struct {
  basicIcon
  nonInterrupt
  Ent       *Entity

  reachable []BoardPos
  path      []BoardPos
}

func (a *ActionMove) Prep() bool {
  level := a.Ent.level
  bx := int(a.Ent.Pos.X)
  by := int(a.Ent.Pos.Y)
  graph := &unitGraph{level, a.Ent}
  reachable := algorithm.ReachableWithinLimit(graph, []int{level.toVertex(bx, by)}, float64(a.Ent.CurAp()))

  if len(reachable) == 0 {
    return false
  }

  vertex_to_boardpos := func(v interface{}) interface{} {
    return level.MakeBoardPosFromVertex(v.(int))
  }

  a.reachable = algorithm.Map(reachable, []BoardPos{}, vertex_to_boardpos).([]BoardPos)

  // Since this is a valid action we can go ahead and highlight all of the
  // tiles that the unit can move to
  for _,v := range a.reachable {
    level.GetCellAtPos(v).highlight |= Reachable
  }

  return true
}

func (a *ActionMove) Cancel() {
  a.reachable = nil
  a.path = nil
  a.Ent.level.clearCache(Reachable)
}

func (a *ActionMove) MouseOver(bx,by float64) {
  // TODO: Might want to highlight the specific path that would be taken if
  // the user clicked here
}

func (a *ActionMove) aiMoveToWithin(tx,ty,rnge int) bool {
  var dsts []int
  for x := tx - rnge; x <= tx + rnge; x++ {
    for y := ty - rnge; y <= ty + rnge; y++ {
      if x == tx && y == ty { continue }
      dsts = append(dsts, a.Ent.level.toVertex(x, y))
    }
  }
  graph := &unitGraph{a.Ent.level, a.Ent}
  _, path := algorithm.Dijkstra(graph, []int{a.Ent.Pos.Vertex(a.Ent.level)}, dsts)
  if path == nil {
    return false
  }
  if len(path) <= 1 || !canPayForMove(a.Ent, a.Ent.level.MakeBoardPosFromVertex(path[1])) {
    return false
  }
  vertex_to_boardpos := func(v interface{}) interface{} {
    return a.Ent.level.MakeBoardPosFromVertex(v.(int))
  }
  a.path = algorithm.Map(path[1:], []BoardPos{}, vertex_to_boardpos).([]BoardPos)
  return true
}

func (a *ActionMove) MouseClick(bx,by float64) ActionCommit {
  level := a.Ent.level
  dst := MakeBoardPos(int(bx), int(by))
  found := false
  for _,v := range a.reachable {
    if dst.IntEquals(v) {
      found = true
      break
    }
  }
  if !found { return NoAction }

  graph := &unitGraph{level, a.Ent}
  ap, path := algorithm.Dijkstra(graph, []int{a.Ent.Pos.Vertex(a.Ent.level)}, []int{dst.Vertex(a.Ent.level)})
  if len(path) <= 1 || int(ap) > a.Ent.CurAp() {
    return NoAction
  }

  vertex_to_boardpos := func(v interface{}) interface{} {
    return level.MakeBoardPosFromVertex(v.(int))
  }

  a.path = algorithm.Map(path[1:], []BoardPos{}, vertex_to_boardpos).([]BoardPos)
  a.reachable = nil

  level.clearCache(Reachable)
  for _,v := range a.path {
    level.GetCellAtPos(v).highlight |= Reachable
  }
  if !payForMove(a.Ent, a.path[0]) {
    a.path = nil
    return NoAction
  }
  return StandardAction
}

func (a *ActionMove) Pause() bool {
  a.Ent.s.Command("stop")
  a.Cancel()
  return false
}

// TODO: Need to make sure that we check for interrupts after every change in
// position - otherwise a slow frame (i.e. a large dt) can cause us to skip
// cells and potentially miss an interrupt that would fire from that cell.
func (a *ActionMove) Maintain(dt int64) MaintenanceStatus {
  plen := len(a.path)
  if AdvanceEntity(a.Ent, &a.path, dt) {
    a.Cancel()
    return Complete
  }
  if len(a.path) < plen {
    return CheckForInterrupts
  }
  return InProgress
}
