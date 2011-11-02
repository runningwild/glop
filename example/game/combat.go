package game

import(
  "rand"
  "regexp"
  "strings"
  "json"
  "fmt"
  "os"
  "io"
  "io/ioutil"
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

type TargetType int
const(
  NoTarget TargetType = 1 << iota
  EntityTarget
  CellTarget
)
type Target struct {
  Type TargetType
  X,Y int
}

type baseWeapon interface {
  // Cost in AP that the source Entity must spent to use this weapon.
  Cost(source *Entity) int

  // Returns a list of all valid targets.
  ValidTargets(source *Entity) []Target

  // Does any effects for when the player mouses over something while this
  // weapon is selected, such as highlighting the region that would be hit
  // with an AOE attack
  MouseOver(source *Entity, bx,by float64)

  // Returns a Resolution indicating everything that happened in once instance of an attack
  // with this weapon
  Damage(source,target *Entity) Resolution
}

type Weapon interface {
  // Using the weapon is an Action, so it must satisfy the Action interface
  Action

  baseWeapon
}


type WeaponInstance struct {
  Base string
  Name string
  Path string
}
func (wi WeaponInstance) Icon() string {
  return wi.Path
}

type WeaponSpec struct {
  WeaponInstance
  baseWeapon
}


// TODO: IT would be nice to change to using the reflect system, if possible, instead of
// requiring a WeaponMaker
type WeaponMaker func() baseWeapon

// map from WeaponInstance.Base to a WeaponMaker for the underlying type for that weapon
var weapon_registry map[string]WeaponMaker

// map from WeaponInstance.Name to a Weapon interface for that specific weapon
var weapon_specs_registry map[string]Weapon
func init() {
  weapon_registry = make(map[string]WeaponMaker)
  rstring := "\\s*{\\s*\"Instance\"\\s*:\\s*(\\{[^}]*\\})\\s*,\\s*\"Weapon\"\\s*:\\s*({(\\s|\\S)*})\\s*}\\s*$"
  rstring = strings.Replace(rstring, "\\s", "[ \t\r\n\f]", -1)
  rstring = strings.Replace(rstring, "\\S", "[^ \t\r\n\f]", -1)
  weapon_spec_regexp = regexp.MustCompile(rstring)

  weapon_specs_registry = make(map[string]Weapon)
}
func RegisterWeapon(base string, maker WeaponMaker) {
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
  json.Unmarshal(res[1], &w.WeaponInstance)
  maker,ok := weapon_registry[w.WeaponInstance.Base]
  if !ok {
    return os.NewError(fmt.Sprintf("Unable to make a weapon of type '%s'", string(res[1])))
  }
  w.baseWeapon = maker()
  json.Unmarshal(res[2], &w.baseWeapon)
  return nil
}


// Reads a file that contains an array of WeaponSpecs in JSON then adds them to the weapon
// spec registry
func LoadWeaponSpecs(spec io.Reader) os.Error {
  var specs []WeaponSpec
  data,err := ioutil.ReadAll(spec)
  if err != nil { return err }
  err = json.Unmarshal(data, &specs)
  if err != nil { return err }
  for i := range specs {
    name := specs[i].WeaponInstance.Name
    if _,ok := weapon_specs_registry[name]; ok {
      return os.NewError(fmt.Sprintf("Cannot register the weapon '%s' because a weapon has already been registered with that name.", name))
    }
    weapon_specs_registry[name] = specs[i]
  }
  return nil
}

func MakeWeapon(name string) Weapon {
  weapon,ok := weapon_specs_registry[name]
  if !ok {
    panic(fmt.Sprintf("Can't make the weapon '%s' because the spec wasn't loaded.", name))
  }
  if !ok { return nil }
  return weapon
}

type NoMouseOverEffect struct {}
func (w NoMouseOverEffect) MouseOver(*Entity, float64, float64) { }

type StaticCost int
func (w StaticCost) Cost(_ *Entity) int {
  return int(w)
}

type EntityRange int
func (r EntityRange) ValidTargets(source *Entity) []Target {
  return entitiesWithinRange(source, int(r))
}

type CellRange int
func (r CellRange) ValidTargets(source *Entity) []Target {
  return cellsWithinRange(source, int(r))
}

func entitiesWithinRange(source *Entity, rnge int) []Target {
  x := int(source.pos.X)
  y := int(source.pos.Y)
  var ret []Target
  for _,target := range source.level.Entities {
    x2 := int(target.pos.X)
    y2 := int(target.pos.Y)

    // Don't let units attack something they don't have LOS to
    if _,ok := source.visible[source.level.toVertex(x2, y2)]; !ok {
      continue
    }

    // Don't let units attack something on their own side
    if source.side == target.side {
      continue
    }

    // The obvious check - must not be further away than our weapon's range
    dist := maxNormi(x, y, x2, y2)
    if rnge < dist {
      continue
    }

    // If it passed all of the above tests then we can include it as a valid
    // target
    ret = append(ret, Target{ EntityTarget, x2, y2 })
  }
  return ret
}

func cellsWithinRange(source *Entity, rnge int) []Target {
  x := int(source.pos.X)
  y := int(source.pos.Y)
  minx := x - rnge
  if minx < 0 { minx = 0 }
  miny := y - rnge
  if miny < 0 { miny = 0 }
  maxx := x + rnge
  if maxx >= len(source.level.grid) { maxx = len(source.level.grid) }
  maxy := y + rnge
  if maxy >= len(source.level.grid[0]) { maxy = len(source.level.grid[0]) }

  var ret []Target
  for x2 := minx; x2 <= maxx; x2++ {
    for y2 := miny; y2 <= maxy; y2++ {
      // Don't let units attack something they don't have LOS to
      if _,ok := source.visible[source.level.toVertex(x2, y2)]; !ok {
        continue
      }

      // The obvious check - must not be further away than our weapon's range
      dist := maxNormi(x, y, x2, y2)
      if rnge < dist {
        continue
      }

      // If it passed all of the above tests then we can include it as a valid
      // target
      ret = append(ret, Target{ CellTarget, x2, y2 })
    }
  }
  return ret
}

func init() {
  RegisterWeapon("Club", func() baseWeapon { return &Club{} })
}
type Club struct {
  StaticCost
  EntityRange
  Factor int
  NoMouseOverEffect
}
func (c *Club) Damage(source,target *Entity) Resolution {
  mod := rand.Intn(10)
  if c.Factor * source.Base.Attack + mod > target.Base.Defense {
    amt := c.Factor * source.Base.Attack + mod - target.Base.Defense - 2
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

func init() {
  RegisterWeapon("Gun", func() baseWeapon { return &Gun{} })
}
type Gun struct {
  StaticCost
  EntityRange
  Power int
  NoMouseOverEffect
}
func (g *Gun) Damage(source,target *Entity) Resolution {
  dist := maxNormi(int(source.pos.X), int(source.pos.Y), int(target.pos.X), int(target.pos.Y))
  acc := 2 * int(g.EntityRange) - dist
  if rand.Intn(acc) < dist {
    return Resolution {
      Connect : Miss,
    }
  }

  if rand.Intn(target.Base.Defense) / 2 > source.Base.Attack {
    return Resolution {
      Connect : Dodge,
    }
  }

  return Resolution {
    Connect : Hit,
    Damage : Damage {
      Piercing : g.Power,
    },
  }
}

func init() {
  RegisterWeapon("StandardAOE", func() baseWeapon { return &StandardAOE{} })
}
type StandardAOE struct {
  StaticCost
  CellRange
  Power  int
  Size int
}
func (g *StandardAOE) affected(source *Entity, bx,by float64) [][2]int {
  var cx,cy int
  if g.Size % 2 == 0 {
    cx = int(bx - 0.5)
    cy = int(by - 0.5)
  } else {
    cx = int(bx)
    cy = int(by)
  }
  var minx,miny,maxx,maxy int
  minx = cx - (g.Size - 1) / 2
  if minx < 0 { minx = 0 }
  miny = cy - (g.Size - 1) / 2
  if miny < 0 { miny = 0 }
  maxx = cx + g.Size / 2
  if maxx >= len(source.level.grid) { maxx = len(source.level.grid) }
  maxy = cy + g.Size / 2
  if maxy >= len(source.level.grid[0]) { maxy = len(source.level.grid[0]) }

  var ret [][2]int
  for x := minx; x <= maxx; x++ {
    for y := miny; y <= maxy; y++ {
      // TODO: Remove things that can't be hit?  Maybe impassable terrain?
      ret = append(ret, [2]int{ x, y })
    }
  }
  return ret
}
func (g *StandardAOE) MouseOver(source *Entity, bx,by float64) {
  targets := g.ValidTargets(source)
  for _,target := range targets {
    if target.Type == CellTarget && target.X == int(bx) && target.Y == int(by) {
      affected := g.affected(source, bx, by)
      for _,pos := range affected {
        source.level.grid[pos[0]][pos[1]].highlight |= AttackMouseOver
      }
      return
    }
  }
}
func (g *StandardAOE) Damage(source,target *Entity) Resolution {
  dist := maxNormi(int(source.pos.X), int(source.pos.Y), int(target.pos.X), int(target.pos.Y))
  acc := 2 * int(g.CellRange) - dist
  if rand.Intn(acc) < dist {
    return Resolution {
      Connect : Miss,
    }
  }

  if rand.Intn(target.Base.Defense) / 2 > source.Base.Attack {
    return Resolution {
      Connect : Dodge,
    }
  }

  return Resolution {
    Connect : Hit,
    Damage : Damage {
      Piercing : g.Power,
    },
  }
}
