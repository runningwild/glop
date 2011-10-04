package game

import(
  "rand"
)

var terrains map[string]Terrain
var weapons map[string]Weapon

type Terrain int
const(
  Grass Terrain = iota
  Dirt
  Water
  Brush
)

type Damage struct {
  Piercing int
  Smashing int
  Fire     int
}

type Connect int
const(
  Hit Connect = iota
  Miss
  Dodge
)
type Resolution struct {
  Connect Connect
  Damage  Damage
}

type Weapon interface {
  Reach() int
  Cost() int
  Damage(source,target *Entity) Resolution
}

type Bayonet struct {}
func (b *Bayonet) Reach() int {
  return 2
}
func (b *Bayonet) Cost() int {
  return 5
}
func (b *Bayonet) Damage(source,target *Entity) Resolution {
  mod := rand.Intn(10)
  if source.Base.Attack + mod > target.Base.Defense {
    amt := source.Base.Attack + mod - target.Base.Defense - 2
    if amt <= 0 {
      return Resolution {
        Connect : Dodge,
      }
    } else {
      return Resolution {
        Connect : Hit,
        Damage : Damage {
          Piercing : amt,
        },
      }
    }
  }
  return Resolution {
    Connect : Miss,
  }
}

type Rifle struct {
  Range int
  Power int
}
func (r *Rifle) Reach() int {
  return r.Range
}
func (r *Rifle) Cost() int {
  return 12
}
func (r *Rifle) Damage(source,target *Entity) Resolution {
  dist := maxNormi(int(source.pos.X), int(source.pos.Y), int(target.pos.X), int(target.pos.Y))
  acc := r.Range - dist
  if rand.Intn(acc) == 0 {
    return Resolution {
      Connect : Miss,
    }
  }

  if rand.Intn(target.Base.Defense) > source.Base.Attack + r.Power {
    return Resolution {
      Connect : Dodge,
    }
  }

  return Resolution {
    Connect : Hit,
    Damage : Damage {
      Piercing : r.Power,
    },
  }
}
