package game

func init() {
  registerActionType("basic attack", &ActionBasicAttack{})
}
type ActionBasicAttack struct {
  basicIcon
  nonInterrupt
  uninterruptable
  Ent     *Entity
  Power   int
  Cost    int
  Range   int
  Melee   int

  targets []*Entity
  mark    *Entity
}

func (a *ActionBasicAttack) Prep() bool {
  if a.Ent.CurAp() < a.Cost {
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

func (a *ActionBasicAttack) MouseClick(bx,by float64) ActionCommit {
  for i := range a.targets {
    if int(bx) == a.targets[i].pos.Xi() && int(by) == a.targets[i].pos.Yi() {
      a.mark = a.targets[i]
      return StandardAction
    }
  }
  return NoAction
}

func (a *ActionBasicAttack) Maintain(dt int64) MaintenanceStatus {
  if a.mark == nil || a.Ent.CurAp() < a.Cost {
    a.Cancel()
    return Complete
  }
  a.Ent.SpendAp(a.Cost)

  if a.Melee != 0 {
    a.Ent.s.Command("melee")
  } else {
    a.Ent.s.Command("ranged")
  }


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
