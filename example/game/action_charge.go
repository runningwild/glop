package game

import "glop/util/algorithm"
import "fmt"

func init() {
  registerActionType("charge attack", &ActionCharge{})
}
type ActionCharge struct {
  basicIcon
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
  graph := &unitGraph{a.Ent.level, a.Ent.Base.attributes.MoveMods}
  vertex_to_boardpos := func(v interface{}) interface{} {
    return a.Ent.level.MakeBoardPosFromVertex(v.(int))
  }
  for _,ent := range a.Ent.level.Entities {
    if ent.side == a.Ent.side { continue }
    if maxNormi(a.Ent.pos.Xi(), a.Ent.pos.Yi(), ent.pos.Xi(), ent.pos.Yi()) <= 2 {
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
    ap, path := algorithm.Dijkstra(graph, []int{a.Ent.pos.Vertex()}, dsts)
    if int(ap) > a.Ent.AP { continue }
    a.valid_marks = append(a.valid_marks, ent.pos)
    a.pos_to_path[ent.pos.Vertex()] = algorithm.Map(path[1:], []BoardPos{}, vertex_to_boardpos).([]BoardPos)
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

func (a *ActionCharge) MouseClick(bx,by float64) bool {
  path,ok := a.pos_to_path[a.Ent.level.toVertex(int(bx), int(by))]
  if !ok { return false }
  a.path = path

  var mark *Entity
  mark_cell := a.Ent.level.MakeBoardPos(int(bx), int(by))
  for _,mark = range a.Ent.level.Entities {
    if mark.pos.AreEqual(&mark_cell) {
      a.mark = mark
      return true
    }
  }
  return false
}

func (a *ActionCharge) Maintain(dt int64) bool {
  if AdvanceEntity(a.Ent, &a.path, dt) {
    a.Ent.s.Command("melee")

    mod_map := map[int]int{
       2 : -3,
       3 : -2,
       4 : -1,
       5 : -1,
       6 :  0,
       7 :  0,
       8 :  0,
       9 :  1,
      10 :  1,
      11 :  2,
      12 :  3,
    }
    attack := a.Power + a.Ent.CurrentAttackMod() + mod_map[Dice("2d6")]
    defense := a.mark.CurrentDefenseMod()


    a.mark.s.Command("defend")
    if attack <= defense {
      a.mark.s.Command("undamaged")
    } else {
      a.mark.Health -= attack - defense
      if a.mark.Health <= 0 {
        a.mark.s.Command("killed")
      } else {
        a.mark.s.Command("damaged")
      }
    }

    a.Ent.turnToFace(a.mark.pos)


    a.Cancel()
    return true
  }
  return false
}
