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
  basicIcon
  nonInterrupt
  Ent       *Entity
  Power     int

  valid_marks  []BoardPos
  pos_to_path  map[int][]BoardPos
  path         []BoardPos
  mark         *Entity
}

func (a *ActionCharge) Prep() bool {
  a.valid_marks = nil
  a.pos_to_path = make(map[int][]BoardPos)
  graph := &unitGraph{a.Ent.level, a.Ent}
  vertex_to_boardpos := func(v interface{}) interface{} {
    return a.Ent.level.MakeBoardPosFromVertex(v.(int))
  }
  for _,ent := range a.Ent.level.Entities {
    if ent.Side == a.Ent.Side { continue }
    if base.MaxNormi(a.Ent.pos.Xi(), a.Ent.pos.Yi(), ent.pos.Xi(), ent.pos.Yi()) <= 2 {
      continue
    }
    var dsts []int
    cx := ent.pos.Xi()
    cy := ent.pos.Yi()
    for x := cx - 1; x <= cx + 1; x++ {
      for y := cy - 1; y <= cy + 1; y++ {
        if x == cx && y == cy { continue }
        dsts = append(dsts, ent.level.toVertex(x, y))
      }
    }
    ap, path := algorithm.Dijkstra(graph, []int{a.Ent.pos.Vertex(a.Ent.level)}, dsts)
    if int(ap) > a.Ent.CurAp() { continue }
    a.valid_marks = append(a.valid_marks, ent.pos)
    a.pos_to_path[ent.pos.Vertex(a.Ent.level)] = algorithm.Map(path[1:], []BoardPos{}, vertex_to_boardpos).([]BoardPos)
  }
  if len(a.valid_marks) == 0 {
    return false
  }

  for _,mark := range a.valid_marks {
    fmt.Printf("Mark: %v\n", mark)
    a.Ent.level.GetCellAtPos(mark).highlight |= Attackable
  }
  return true
}

func (a *ActionCharge) Cancel() {
  a.valid_marks = nil
  a.pos_to_path = nil
  a.Ent.level.clearCache(Attackable | Reachable)
}

func (a *ActionCharge) MouseOver(bx,by float64) {
  // TODO: Might want to highlight the specific path that would be taken if
  // the user clicked here
  if len(a.valid_marks) == 0 { return }
  a.Ent.level.clearCache(Reachable)
  for _,bp := range a.pos_to_path[a.Ent.level.toVertex(int(bx), int(by))] {
    a.Ent.level.GetCellAtPos(bp).highlight |= Reachable
  }
}

func (a *ActionCharge) MouseClick(bx,by float64) ActionCommit {
  path,ok := a.pos_to_path[a.Ent.level.toVertex(int(bx), int(by))]
  if !ok { return NoAction }
  a.path = path

  var mark *Entity
  mark_cell := MakeBoardPos(int(bx), int(by))
  for _,mark = range a.Ent.level.Entities {
    if mark.pos.IntEquals(mark_cell) {
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

    a.Ent.turnToFace(a.mark.pos)


    a.Cancel()
    return Complete
  }
  if len(a.path) < plen {
    return CheckForInterrupts
  }
  return InProgress
}
