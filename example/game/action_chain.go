package game

func init() {
  registerActionType("chain attack", &ActionChainAttack{})
}
type ActionChainAttack struct {
  basicIcon
  nonInterrupt
  Ent   *Entity

  Power int
  Cost  int
  Range int
  Melee int
  Adds  int

  targets map[*Entity]bool
  marks   []*Entity
}

func (a *ActionChainAttack) Prep() bool {
  if a.Ent.CurAp() < a.Cost {
    return false
  }

  targets := getEntsWithinRange(a.Ent, a.Range, a.Ent.level)
  if len(targets) == 0 {
    return false
  }

  a.targets = make(map[*Entity]bool, len(a.targets))
  a.marks = nil
  for _,target := range targets {
    a.targets[target] = true
    a.Ent.level.GetCellAtPos(target.pos).highlight |= Attackable
  }
  return true
}

func (a *ActionChainAttack) Cancel() {
  a.marks = nil
  a.targets = nil
  a.Ent.level.clearCache(Attackable | Targeted)
}

func (a *ActionChainAttack) MouseOver(bx,by float64) {
}

func (a *ActionChainAttack) MouseClick(bx,by float64) ActionCommit {
  t := findTargetOnClick(bx, by, a.Ent.level, a.targets)
  if t == nil { return NoAction }
  a.Ent.level.GetCellAtPos(t.pos).highlight |= Targeted
  a.marks = append(a.marks, t)

  if len(a.marks) == a.Adds {
    a.Ent.SpendAp(a.Cost)
    return StandardAction
  }
  return NoAction
}

func (a *ActionChainAttack) Maintain(dt int64) bool {
  if len(a.marks) == 0 {
    a.Cancel()
    return true
  }

  mark := a.marks[0]
  for _,ent := range []*Entity{ a.Ent, mark } {
    if ent.s.NumPendingCommands() != 0 { return false }
    if ent.s.CurState() == "killed" {
      // The mark may have already died from a previous attack in this chain,
      // in that case we just skip this entity
      a.marks = a.marks[1 : ]
      return a.Maintain(dt)
    }
    if ent.s.CurAnim() != "ready" { return false }
  }

  a.marks = a.marks[1 : ]

  if a.Melee != 0 {
    a.Ent.s.Command("melee")
  } else {
    a.Ent.s.Command("ranged")
  }

  attack := a.Power + a.Ent.CurAttack() + ((Dice("5d5") - 2) / 3 - 4)
  defense := mark.CurDefense()

  mark.s.Command("defend")
  if attack <= defense {
    mark.s.Command("undamaged")

    // Chain attacks only continue after successful attacks
    a.Cancel()
    return true
  } else {
    mark.DoDamage(attack - defense)
    if mark.CurHealth() <= 0 {
      mark.s.Command("killed")
    } else {
      mark.s.Command("damaged")
    }
  }

  a.Ent.turnToFace(mark.pos)

  return false
}
