package game

func init() {
  registerActionType("multistrike", &ActionMultiStrike{})
}
type ActionMultiStrike struct {
  basicAction
  basicIcon
  nonInterrupt
  uninterruptable

  Cost    int
  Power   int
  Range   int
  Melee   int
  Count   int  // Number of possible targets

  targets map[*Entity]bool
  marks   map[*Entity]bool
}

func (a *ActionMultiStrike) Prep() bool {
  if a.Ent.CurAp() < a.Cost {
    return false
  }

  targets := getEntsWithinRange(a.Ent, a.Range, a.Level)
  if len(targets) == 0 {
    return false
  }

  a.targets = make(map[*Entity]bool, len(a.targets))
  a.marks = make(map[*Entity]bool, a.Count)
  for _,target := range targets {
    a.targets[target] = true
    a.Level.GetCellAtPos(target.Pos).highlight |= Attackable
  }
  return true
}

func (a *ActionMultiStrike) Cancel() {
  a.marks = nil
  a.targets = nil
  a.Level.clearCache(Attackable | Targeted)
}

func (a *ActionMultiStrike) MouseOver(bx,by float64) {
}

func (a *ActionMultiStrike) MouseClick(bx,by float64) ActionCommit {
  if findMultipleUniqueTargets(bx, by, a.Level, &a.targets, &a.marks, a.Count) {
    return StandardAction
  }
  return NoAction
}

func (a *ActionMultiStrike) Maintain(dt int64) MaintenanceStatus {
  if a.marks == nil || a.Ent.CurAp() < a.Cost {
    a.Cancel()
    return Complete
  }
  a.Ent.SpendAp(a.Cost)

  if a.Melee != 0 {
    a.Ent.s.Command("melee")
  } else {
    a.Ent.s.Command("ranged")
  }


  for mark,_ := range a.marks {
    attack := a.Power + a.Ent.CurAttack() + ((Dice("5d5") - 2) / 3 - 4)
    defense := mark.CurDefense()

    mark.s.Command("defend")
    if attack <= defense {
      mark.s.Command("undamaged")
    } else {
      mark.DoDamage(attack - defense)
      if mark.CurHealth() <= 0 {
        mark.s.Command("killed")
      } else {
        mark.s.Command("damaged")
      }
    }

    // TODO: This is kinda dumb, we just change facing a bunch and stay facing
    // at the last target (which is random).  Might want to do something like
    // face the average of all of the targets
    a.Ent.turnToFace(mark.Pos)
  }

  a.Cancel()
  return Complete
}
