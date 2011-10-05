package game

import(
  "rand"
  "exp/regexp"
  "json"
  "fmt"
  "os"
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
  // Cost in AP that the source Entity must spent to use this weapon.
  Cost(source *Entity) int

  // Returns true iff the source Entity can hit the target Entity with this weapon.
  InRange(source,target *Entity) bool

  // Returns a Resolution indicating everything that happened in once instance of an attack
  // with this weapon
  Damage(source,target *Entity) Resolution
}


type weaponInstance struct {
  Base string
  Name string
}
type WeaponSpec struct {
  Instance weaponInstance
  Weapon
}
type weaponMaker func() Weapon
var weapon_registry map[string]weaponMaker
func init() {
  weapon_registry = make(map[string]weaponMaker)
  weapon_spec_regexp = regexp.MustCompile("\\s*{\\s*\"Instance\"\\s*:\\s*(\\{[^}]*\\})\\s*,\\s*\"Weapon\"\\s*:\\s*({(\\s|\\S)*})\\s*}\\s*$")
}
func RegisterWeapon(base string, maker weaponMaker) {
  if _,ok := weapon_registry[base]; ok {
    panic(fmt.Sprintf("Tried to register the weapon '%s' twice.", base))
  }
  weapon_registry[base] = maker
}

var weapon_spec_regexp *regexp.Regexp
func init() {
}

func (w *WeaponSpec) UnmarshalJSON(data []byte) os.Error {
  res := weapon_spec_regexp.FindSubmatch(data)
  if res == nil {
    return os.NewError(fmt.Sprintf("Unable to unmarshal the JSON for the MarshalWeapon:\n%s\n", string(data)))
  }
  json.Unmarshal(res[1], &w.Instance)
  maker,ok := weapon_registry[w.Instance.Base]
  if !ok {
    return os.NewError(fmt.Sprintf("Unable to make a weapon of type '%s'", string(res[1])))
  }
  w.Weapon = maker()
  json.Unmarshal(res[2], &w.Weapon)
  return nil
}


type StaticCost int
func (w StaticCost) Cost(_ *Entity) int {
  return int(w)
}

type StaticRange int
func (w StaticRange) InRange(source,target *Entity) bool {
  x := int(source.pos.X)
  y := int(source.pos.Y)
  x2 := int(target.pos.X)
  y2 := int(target.pos.Y)
  dist := maxNormi(x, y, x2, y2)
  return int(w) >= dist
}

type Bayonet struct {
  StaticCost
  StaticRange
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
  StaticCost
  StaticRange
  Power int
}
func (r *Rifle) Damage(source,target *Entity) Resolution {
  dist := maxNormi(int(source.pos.X), int(source.pos.Y), int(target.pos.X), int(target.pos.Y))
  acc := int(r.StaticRange) - dist
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
