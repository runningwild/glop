package game

func init() {
  registerActionType("basic attack", &ActionBasicAttack{})
}
type ActionBasicAttack struct {
  basicIcon
  Ent     *Entity
  Power   int
  Cost    int
  Range   int
  Melee   int

  targets []*Entity
  mark    *Entity
}

func (a *ActionBasicAttack) Prep() bool {
  if a.Ent.AP < a.Cost {
    return false
  }

  a.targets = nil
  for _,ent := range a.Ent.level.Entities {
    if ent.side == a.Ent.side { continue }
    dist := maxNormi(a.Ent.pos.Xi(), a.Ent.pos.Yi(), ent.pos.Xi(), ent.pos.Yi())
    if _,ok := a.Ent.visible[ent.pos.Vertex()]; !ok { continue }
    if dist > a.Range { continue }
    a.targets = append(a.targets, ent)
  }

  for _,target := range a.targets {
    a.Ent.level.GetCellAtPos(target.pos).highlight |= Attackable
  }
  return true
}

func (a *ActionBasicAttack) Cancel() {
  a.mark = nil
  a.targets = nil
  a.Ent.level.clearCache(Attackable)
}

func (a *ActionBasicAttack) MouseOver(bx,by float64) {
}

func (a *ActionBasicAttack) MouseClick(bx,by float64) bool {
  for i := range a.targets {
    if int(bx) == a.targets[i].pos.Xi() && int(by) == a.targets[i].pos.Yi() {
      a.mark = a.targets[i]
      return true
    }
  }
  return false
}

func (a *ActionBasicAttack) Maintain(dt int64) bool {
  if a.mark == nil { return false }
  if a.Ent.AP < a.Cost {
    a.Cancel()
    return true
  }
  a.Ent.AP -= a.Cost

  if a.Melee != 0 {
    a.Ent.s.Command("ranged")
  } else {
    a.Ent.s.Command("melee")
  }

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
