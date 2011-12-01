package game

func init() {
  registerActionType("spray attack", &ActionSpray{})
}
type ActionSpray struct {
  basicIcon
  nonInterrupt
  uninterruptable
  Ent    *Entity
  Cost   int
  Power  int
  Melee  int
  Length int
  Start  int
  End    int

  dir   BoardPos
  cells []BoardPos
}

func (a *ActionSpray) Prep() bool {
  a.cells = nil
  return a.Ent.CurAp() >= a.Cost
}

func (a *ActionSpray) Cancel() {
  a.Ent.level.clearCache(Targeted)
  a.cells = nil
  a.dir = MakeBoardPos(0, 0)
}

func (a *ActionSpray) getDir(bp BoardPos) BoardPos {
  diff := bp.Sub(a.Ent.Pos)
  dx := diff.Xi()
  if dx < 0 { dx = -dx }
  dy := diff.Yi()
  if dy < 0 { dy = -dy }
  if dx > dy {
    if diff.Xi() < 0 {
      return MakeBoardPos(-1, 0)
    }
    return MakeBoardPos(1, 0)
  }
  if diff.Yi() < 0 {
    return MakeBoardPos(0, -1)
  }
  return MakeBoardPos(0, 1)
}

func (a *ActionSpray) MouseOver(bx,by float64) {
  dir := a.getDir(MakeBoardPos(int(bx), int(by)))
  if dir.IntEquals(a.dir) {
    return
  }
  a.Ent.level.clearCache(Targeted)
  a.dir = dir
  a.cells = nil
  side := MakeBoardPos(a.dir.Yi(), a.dir.Xi())
  pos := a.Ent.Pos
  for i := 0; i < a.Length; i++ {
    pos = pos.Add(a.dir)
    var width int
    if a.Length <= 1 {
      width = a.Start
    } else {
      width = ((a.End - a.Start) * i) / (a.Length - 1)
    }
    row := pos.Add(side.Scale(-width))
    for j := 0; j < width*2 + 1; j++ {
      if row.Valid(a.Ent.level) {
        a.Ent.level.GetCellAtPos(row).highlight |= Targeted
      }
      a.cells = append(a.cells, row)
      row = row.Add(side)
    }
  }
}

func (a *ActionSpray) MouseClick(bx,by float64) ActionCommit {
  return StandardAction
}

func (a *ActionSpray) Maintain(dt int64) MaintenanceStatus {
  a.Ent.SpendAp(a.Cost)
  if a.Melee != 0 {
    a.Ent.s.Command("melee")
  } else {
    a.Ent.s.Command("ranged")
  }
  for _,cell := range a.cells {
    mark := a.Ent.level.GetCellAtPos(cell).ent
    if mark == nil { continue }
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
  }
  a.Cancel()
  return Complete
}
