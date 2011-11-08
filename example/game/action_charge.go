package game

import (
  "github.com/arbaal/mathgl"
  "glop/util/algorithm"
)

type chargeMode int
const(
  notCharging chargeMode = iota
  chargeMove
  chargeAttack
)

type ActionChargeAttack struct {
  basicIcon
  ent     *Entity
  mark    chargeTarget
  mode    chargeMode
  weapon  Weapon
  targets []chargeTarget
}

func makeChargeAttackAction(ent *Entity, weapon Weapon) Action {
  action := ActionChargeAttack{ ent : ent, weapon : weapon }
  action.path = weapon.Icon()
  return &action
}

func (a *ActionChargeAttack) Prep() bool {
  a.targets = a.getValidTargets()
  if len(a.targets) == 0 { return false }
  for _,target := range a.targets {
    a.ent.level.grid[int(target.ent.pos.X)][int(target.ent.pos.Y)].highlight |= Attackable
    for _,v := range target.path {
      x,y := a.ent.level.fromVertex(v)
      a.ent.level.grid[x][y].highlight |= Reachable
    }
  }
  return true
}

type chargeTarget struct {
  ent  *Entity
  path []int
}

func (a *ActionChargeAttack) getValidTargets() []chargeTarget {
  var valid []chargeTarget
  level := a.ent.level
  graph := &unitGraph{ a.ent.level, a.ent.Base.attributes.MoveMods }
  src := level.toVertex(int(a.ent.pos.X), int(a.ent.pos.Y))
  for _,ent := range level.Entities {
    if ent.side == a.ent.side { continue }
    var dst []int
    ex := int(ent.pos.X)
    ey := int(ent.pos.Y)
    for x := ex - 1; x <= ex + 1; x++ {
      for y := ey - 1; y <= ey + 1; y++ {
        dst = append(dst, level.toVertex(x, y))
      }
    }
    dist,path := algorithm.Dijkstra(graph, []int{ src }, dst)
    final_terrain := level.grid[int(ent.pos.X)][int(ent.pos.Y)].Terrain
    dist -= float64(a.ent.Base.attributes.MoveMods[final_terrain])
    if len(path) <= 2 { continue }
    if int(dist) > a.ent.AP { continue }
//    path = path[1 : ]
    valid = append(valid, chargeTarget{ ent, path })
  }
  return valid
}

func (a *ActionChargeAttack) Cancel() {
  a.targets = nil
  a.mode = notCharging
  a.ent.level.clearCache(Attackable | Reachable)
}

func (a *ActionChargeAttack) MouseOver(bx,by float64) {
}

func (a *ActionChargeAttack) MouseClick(bx,by float64) bool {
  x := int(bx)
  y := int(by)
  for _,target := range a.targets {
    if int(target.ent.pos.X) == x && int(target.ent.pos.Y) == y {
      a.mark = target
      a.mode = chargeMove
      return true
    }
  }
  return false
}

func (a *ActionChargeAttack) payForMove() bool {
  level := a.ent.level
  graph := unitGraph{level, a.ent.Base.attributes.MoveMods}
  src := level.toVertex(int(a.ent.pos.X), int(a.ent.pos.Y))
  cost := int(graph.costToMove(src, a.mark.path[0]))
  if cost > a.ent.AP {
    return false
  }
  a.ent.AP -= cost
  return true
}

func (a *ActionChargeAttack) doMove(dt int64) bool {
  x,y := a.ent.level.fromVertex(a.mark.path[0])
  tomove := a.ent.Move_speed * float32(dt)
  for tomove > 0 {
    moved,reached := a.ent.Advance(x, y, tomove)
    if moved == 0 && !reached { return false }
    tomove -= moved

    // Check to see if the entity has made it to a new cell
    if reached {
      a.ent.OnEntry()
      px,py := a.ent.level.fromVertex(a.mark.path[0])
      a.ent.level.grid[px][py].highlight &= ^Reachable
      a.mark.path = a.mark.path[1:]

      // If we have reached our destination *OR* if something has happened and
      // we no longer have the AP required to continue moving then this action
      // is complete - so we return true
      if len(a.mark.path) == 0 || !a.payForMove() {
        a.Cancel()
        a.ent.Advance(0, 0, 0)
        return true
      }
      x,y = a.ent.level.fromVertex(a.mark.path[len(a.mark.path) - 1])
    }
  }
  return false
}

func (a *ActionChargeAttack) doAttack() bool {
  target := Target{
    Type : EntityTarget,
    X : int(a.mark.ent.pos.X),
    Y : int(a.mark.ent.pos.Y),
  }
  ress := a.weapon.Damage(a.ent, target)

  a.ent.turnToFace(mathgl.Vec2{float32(a.mark.ent.pos.X), float32(a.mark.ent.pos.Y)})

  dist := maxNormi(target.X, target.Y, int(a.ent.pos.X), int(a.ent.pos.Y))

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
  return true
}

func (a *ActionChargeAttack) Maintain(dt int64) bool {
  switch a.mode {
    case chargeMove:
      if a.doMove(dt) {
        a.mode = chargeAttack
      }
      return false

    case chargeAttack:
      a.doAttack()
      a.mode = notCharging
      return true
  }
  panic("this should be unreachable")
}
