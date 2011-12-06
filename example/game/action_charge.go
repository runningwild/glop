package game

import (
  "game/base"
  "glop/util/algorithm"
  "fmt"
)

func init() {
  registerActionType("charge attack", &ActionCharge{})
}
type ActionCharge struct {
  basicAction
  basicIcon
  nonInterrupt

  // Cost to charge is the same as the cost to move that distance
  Power int

  valid_marks  []BoardPos
  pos_to_path  map[int][]BoardPos
  path         []BoardPos
  mark         *Entity
}

func (a *ActionCharge) Prep() bool {
  a.valid_marks = nil
  a.pos_to_path = make(map[int][]BoardPos)
  graph := &unitGraph{a.Level, a.Ent}
  vertex_to_boardpos := func(v interface{}) interface{} {
    return a.Level.MakeBoardPosFromVertex(v.(int))
  }
  for _,ent := range a.Level.Entities {
    if ent.Side == a.Ent.Side { continue }
    if base.MaxNormi(a.Ent.Pos.Xi(), a.Ent.Pos.Yi(), ent.Pos.Xi(), ent.Pos.Yi()) <= 2 {
      continue
    }
    var dsts []int
    cx := ent.Pos.Xi()
    cy := ent.Pos.Yi()
    for x := cx - 1; x <= cx + 1; x++ {
      for y := cy - 1; y <= cy + 1; y++ {
        if x == cx && y == cy { continue }
        dsts = append(dsts, ent.level.toVertex(x, y))
      }
    }
    ap, path := algorithm.Dijkstra(graph, []int{a.Ent.Pos.Vertex(a.Level)}, dsts)
    if int(ap) > a.Ent.CurAp() { continue }
    a.valid_marks = append(a.valid_marks, ent.Pos)
    a.pos_to_path[ent.Pos.Vertex(a.Level)] = algorithm.Map(path[1:], []BoardPos{}, vertex_to_boardpos).([]BoardPos)
  }
  if len(a.valid_marks) == 0 {
    return false
  }

  for _,mark := range a.valid_marks {
    fmt.Printf("Mark: %v\n", mark)
    a.Level.GetCellAtPos(mark).highlight |= Attackable
  }
  return true
}

func (a *ActionCharge) Cancel() {
  a.valid_marks = nil
  a.pos_to_path = nil
  a.Level.clearCache(Attackable | Reachable)
}

func (a *ActionCharge) MouseOver(bx,by float64) {
  // TODO: Might want to highlight the specific path that would be taken if
  // the user clicked here
  if len(a.valid_marks) == 0 { return }
  a.Level.clearCache(Reachable)
  for _,bp := range a.pos_to_path[a.Level.toVertex(int(bx), int(by))] {
    a.Level.GetCellAtPos(bp).highlight |= Reachable
  }
}

func (a *ActionCharge) MouseClick(bx,by float64) ActionCommit {
  path,ok := a.pos_to_path[a.Level.toVertex(int(bx), int(by))]
  if !ok { return NoAction }
  a.path = path

  var mark *Entity
  mark_cell := MakeBoardPos(int(bx), int(by))
  for _,mark = range a.Level.Entities {
    if mark.Pos.IntEquals(mark_cell) {
      a.mark = mark
      return StandardAction
    }
  }
  return NoAction
}

func (a *ActionCharge) Pause() bool {
  a.Ent.s.Command("stop")
  return true
}

func (a *ActionCharge) Maintain(dt int64) MaintenanceStatus {
  plen := len(a.path)
  if AdvanceEntity(a.Ent, &a.path, dt) {
    a.Ent.s.Command("melee")

    attack := a.Power + a.Ent.CurAttack() + ((Dice("5d5") - 2) / 3 - 4)
    defense := a.mark.CurDefense()

    a.mark.s.Command("defend")
    if attack <= defense {
      a.mark.s.Command("undamaged")
    } else {
      a.mark.DoDamage(attack - defense)
      if a.mark.CurHealth() <= 0 {
        a.mark.s.Command("killed")
      } else {
        a.mark.s.Command("damaged")
      }
    }

    a.Ent.turnToFace(a.mark.Pos)


    a.Cancel()
    return Complete
  }
  if len(a.path) < plen {
    return CheckForInterrupts
  }
  return InProgress
}
