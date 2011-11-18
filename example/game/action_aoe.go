package game

import "game/stats"

func init() {
  registerActionType("aoe", &ActionAoe{})
}
type ActionAoe struct {
  basicIcon
  nonInterrupt
  Ent     *Entity
  Cost    int
  Range   int
  Size    int
  Allies  int
  Enemies int
  Effects []string
}

func (a *ActionAoe) Prep() bool {
  if a.Ent.CurAp() < a.Cost {
    return false
  }

  if a.Range == 0 {
    for dx := -a.Size; dx <= a.Size; dx++ {
      for dy := -a.Size; dy <= a.Size; dy++ {
        t := a.Ent.pos.Add(MakeBoardPos(dx, dy))
        if t.Valid(a.Ent.level) {
          a.Ent.level.GetCellAtPos(t).highlight |= Attackable
        }
      }
    }
  } else {
    
  }
  return true
}

func (a *ActionAoe) Cancel() {
  a.Ent.level.clearCache(Attackable)
}

func (a *ActionAoe) MouseOver(bx,by float64) {
}

func (a *ActionAoe) MouseClick(bx,by float64) ActionCommit {
  return StandardAction
}

func (a *ActionAoe) Maintain(dt int64) bool {
  a.Ent.SpendAp(a.Cost)
  if a.Range == 0 {
    for dx := -a.Size; dx <= a.Size; dx++ {
      for dy := -a.Size; dy <= a.Size; dy++ {
        t := a.Ent.pos.Add(MakeBoardPos(dx, dy))
        if !t.Valid(a.Ent.level) { continue }
        ent := a.Ent.level.GetCellAtPos(t).ent
        if ent == nil { continue }
        ent.s.Command("defend")
        ent.s.Command("undamaged")
        for _,effect := range a.Effects {
          ent.AddEffect(stats.MakeEffect(effect))
        }
      }
    }
  } else {
    
  }
  a.Cancel()
  return true
}
