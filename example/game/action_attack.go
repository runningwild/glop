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

  a.targets = getEntsWithinRange(a.Ent, a.Range, a.Ent.level)
  if len(a.targets) == 0 {
    return false
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
    a.Ent.s.Command("melee")
  } else {
    a.Ent.s.Command("ranged")
  }

  attack := a.Power + a.Ent.CurrentAttackMod() + ((Dice("5d5") - 2) / 3)
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
