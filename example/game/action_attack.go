package game

type ActionBasicAttack struct {
  basicIcon
  ent     *Entity
  weapon  Weapon
  targets []Target
  mark    *Target
}

func makeBasicAttackAction(ent *Entity, weapon Weapon) Action {
  action := ActionBasicAttack{ ent : ent, weapon : weapon }
  action.path = weapon.Icon()
  return &action
}

func (a *ActionBasicAttack) Prep() bool {
  if a.ent.AP < a.weapon.Cost(a.ent) {
    return false
  }
  a.targets = a.weapon.ValidTargets(a.ent)
  for _,target := range a.targets {
    a.ent.level.grid[target.X][target.Y].highlight |= Attackable
  }
  return true
}

func (a *ActionBasicAttack) Cancel() {
  a.mark = nil
  a.ent.level.clearCache(Attackable)
}

func (a *ActionBasicAttack) MouseOver(bx,by float64) {
}

func (a *ActionBasicAttack) MouseClick(bx,by float64) bool {
  for i := range a.targets {
    if int(bx) == a.targets[i].X && int(by) == a.targets[i].Y {
      a.mark = &a.targets[i]
      return true
    }
  }
  return false
}

func (a *ActionBasicAttack) Maintain(dt int64) bool {
  if a.mark == nil { return false }
  cost := a.weapon.Cost(a.ent)
  if a.ent.AP < cost {
    a.Cancel()
    return true
  }
  a.ent.AP -= cost

  ress := a.weapon.Damage(a.ent, *a.mark)

  a.ent.turnToFace(a.ent.level.MakeBoardPos(a.mark.X, a.mark.Y))

  dist := maxNormi(a.mark.X, a.mark.Y, int(a.ent.pos.X), int(a.ent.pos.Y))

  // TODO: Melee/ranged should be determined by the weapon, not by the distance
  if dist >= 2 {
    a.ent.s.Command("ranged")
  } else {
    a.ent.s.Command("melee")
  }

  for _, res := range ress {
    res.Target.s.Command("defend")

    if res.Connect == Hit {
      res.Target.Health -= res.Damage.Piercing
      if res.Target.Health <= 0 {
        res.Target.s.Command("killed")
      } else {
        res.Target.s.Command("damaged")
      }
    } else {
      res.Target.s.Command("undamaged")
    }
  }
  a.Cancel()
  return true
}
