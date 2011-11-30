package game

func init() {
  registerActionType("counter attack", &ActionCounterAttack{})
}
type ActionCounterAttack struct {
  basicIcon
  uninterruptable
  Ent     *Entity
  Power   int
  Cost    int
  Range   int
  Melee   int

  mark *Entity
}

func (a *ActionCounterAttack) Prep() bool {
  if a.Ent.CurAp() < a.Cost {
    return false
  }
  return true
}

func (a *ActionCounterAttack) Cancel() {
}

func (a *ActionCounterAttack) MouseOver(bx,by float64) {
}

func (a *ActionCounterAttack) MouseClick(bx,by float64) ActionCommit {
  a.Ent.SpendAp(a.Cost)
  return StandardInterrupt
}

func (a *ActionCounterAttack) Interrupt() bool {
  print("Checking counter attack\n")
  for dx := -a.Range; dx <= a.Range; dx++ {
    for dy := -a.Range; dy <= a.Range; dy++ {
      t := a.Ent.pos.Add(MakeBoardPos(dx, dy))
      if t.Valid(a.Ent.level) {
        mark := a.Ent.level.GetCellAtPos(t).ent
        if mark != nil && mark.side != a.Ent.side {
          a.mark = mark
          return true
        }
      }
    }
  }
  return false
}

func (a *ActionCounterAttack) Maintain(dt int64) MaintenanceStatus {
  if a.mark == nil {
    a.Cancel()
    print("Cancelig\n")
    return Complete
  }
    print("Do it!\n")

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
