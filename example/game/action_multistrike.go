package game

func init() {
  registerActionType("multistrike", &ActionMultiStrike{})
}
type ActionMultiStrike struct {
  basicIcon
  Ent     *Entity
  Power   int
  Cost    int
  Range   int
  Melee   int
  Count   int

  targets map[*Entity]bool
  marks   map[*Entity]bool
}

func (a *ActionMultiStrike) Prep() bool {
  if a.Ent.AP < a.Cost {
    return false
  }

  targets := getEntsWithinRange(a.Ent, a.Range, a.Ent.level)
  if len(targets) == 0 {
    return false
  }

  a.targets = make(map[*Entity]bool, len(a.targets))
  a.marks = make(map[*Entity]bool, a.Count)
  for _,target := range targets {
    a.targets[target] = true
    a.Ent.level.GetCellAtPos(target.pos).highlight |= Attackable
  }
  return true
}

func (a *ActionMultiStrike) Cancel() {
  a.marks = nil
  a.targets = nil
  a.Ent.level.clearCache(Attackable | Targeted)
}

func (a *ActionMultiStrike) MouseOver(bx,by float64) {
}

func (a *ActionMultiStrike) MouseClick(bx,by float64) bool {
  bp := MakeBoardPos(int(bx), int(by))
  t := a.Ent.level.GetCellAtPos(bp).ent
  if t == nil { return false }
  if _,ok := a.targets[t]; !ok { return false }
  if _,ok := a.marks[t]; ok {
    return true
  }
  a.marks[t] = true
  a.Ent.level.GetCellAtPos(bp).highlight |= Targeted
  return len(a.marks) == a.Count
}

func (a *ActionMultiStrike) Maintain(dt int64) bool {
  if a.marks == nil || a.Ent.AP < a.Cost {
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

  for mark,_ := range a.marks {
    defense := mark.CurrentDefenseMod()

    mark.s.Command("defend")
    if attack <= defense {
      mark.s.Command("undamaged")
    } else {
      mark.Health -= attack - defense
      if mark.Health <= 0 {
        mark.s.Command("killed")
      } else {
        mark.s.Command("damaged")
      }
    }

    // TODO: This is kinda dumb, we just change facing a bunch and stay facing
    // at the last target (which is random).  Might want to do something like
    // face the average of all of the targets
    a.Ent.turnToFace(mark.pos)
  }

  a.Cancel()
  return true
}
