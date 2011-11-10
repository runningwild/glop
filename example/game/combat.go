package game

import (
  "errors"
  "regexp"
  "strings"
  "json"
  "fmt"
  "io"
  "io/ioutil"
)

type Damage struct {
  Piercing int
  Smashing int
  Fire     int
}

type Connect int

const (
  Hit Connect = iota
  Miss
  Dodge
)

type Resolution struct {
  Connect Connect
  Damage  Damage
  Target  *Entity
}

type TargetType int

const (
  NoTarget TargetType = 1 << iota
  EntityTarget
  CellTarget
)

// NEXT: Need to make Target objects from MouseOver so we can pass them arond to ensure that MouseOver and Damage are both hitting the exact same thing.
type Target struct {
  Type TargetType
  X, Y int
}

type AttackActionMaker func(*Entity,Weapon) Action

type BaseWeapon interface {
  // Cost in AP that the source Entity must spent to use this weapon.
  Cost(source *Entity) int

  // Returns a maker for the appropriate action for this Weapon, such as
  // makeBasicAttackAction.
  ActionMaker() AttackActionMaker

  // Returns a list of all valid targets.
  ValidTargets(source *Entity) []Target

  // Does any effects for when the player mouses over something while this
  // weapon is selected, such as highlighting the region that would be hit
  // with an AOE attack
  // Returns the Target object that should be used to target whatever is
  // at the specified board coordinates
  MouseOver(source *Entity, bx, by float64) Target

  // Returns a Resolution indicating everything that happened in once instance of an attack
  // with this weapon
  Damage(source *Entity, target Target) []Resolution
}

type Weapon interface {
  // Using the weapon is an Action, so it must satisfy the Action interface
  Icon() string

  // Returns an Action object that can be used in the Action system
  GetAction(source *Entity) Action

  BaseWeapon
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
  BaseWeapon
}

func (wi WeaponSpec) GetAction(source *Entity) Action {
  return wi.ActionMaker()(source, wi)
}

// TODO: IT would be nice to change to using the reflect system, if possible, instead of
// requiring a WeaponMaker
type WeaponMaker func() BaseWeapon

// map from WeaponInstance.Base to a WeaponMaker for the underlying type for that weapon
var weapon_registry map[string]WeaponMaker

// map from WeaponInstance.Name to a Weapon interface for that specific weapon
var weapon_specs_registry map[string]Weapon

func init() {
  weapon_registry = make(map[string]WeaponMaker)
  rstring := "\\s*{\\s*\"WeaponInstance\"\\s*:\\s*(\\{[^}]*\\})\\s*,\\s*\"BaseWeapon\"\\s*:\\s*({(\\s|\\S)*})\\s*}\\s*$"
  rstring = strings.Replace(rstring, "\\s", "[ \t\r\n\f]", -1)
  rstring = strings.Replace(rstring, "\\S", "[^ \t\r\n\f]", -1)
  weapon_spec_regexp = regexp.MustCompile(rstring)

  weapon_specs_registry = make(map[string]Weapon)
}
func RegisterWeapon(base string, maker WeaponMaker) {
  if _, ok := weapon_registry[base]; ok {
    panic(fmt.Sprintf("Tried to register the weapon '%s' twice.", base))
  }
  weapon_registry[base] = maker
}

var weapon_spec_regexp *regexp.Regexp

func init() {
}

func (w *WeaponSpec) UnmarshalJSON(data []byte) error {
  res := weapon_spec_regexp.FindSubmatch(data)
  if res == nil {
    return errors.New(fmt.Sprintf("Unable to unmarshal the JSON for the MarshalWeapon:\n%s\n", string(data)))
  }
  json.Unmarshal(res[1], &w.WeaponInstance)
  maker, ok := weapon_registry[w.WeaponInstance.Base]
  if !ok {
    return errors.New(fmt.Sprintf("Unable to make a weapon of type '%s'", string(res[1])))
  }
  w.BaseWeapon = maker()
  json.Unmarshal(res[2], &w.BaseWeapon)
  return nil
}

// Reads a file that contains an array of WeaponSpecs in JSON then adds them to the weapon
// spec registry
func LoadWeaponSpecs(spec io.Reader) error {
  var specs []WeaponSpec
  data, err := ioutil.ReadAll(spec)
  if err != nil {
    return err
  }
  err = json.Unmarshal(data, &specs)
  if err != nil {
    fmt.Printf("Got the error: %s when reading:\n%s\n", err.Error(), string(data))
    return err
  }
  for i := range specs {
    name := specs[i].WeaponInstance.Name
    if _, ok := weapon_specs_registry[name]; ok {
      return errors.New(fmt.Sprintf("Cannot register the weapon '%s' because a weapon has already been registered with that name.", name))
    }
    weapon_specs_registry[name] = specs[i]
  }
  return nil
}

func MakeWeapon(name string) Weapon {
  weapon, ok := weapon_specs_registry[name]
  if !ok {
    panic(fmt.Sprintf("Can't make the weapon '%s' because the spec wasn't loaded.", name))
  }
  if !ok {
    return nil
  }
  return weapon
}

type SingleTargetMouseOver struct{}

func (w SingleTargetMouseOver) MouseOver(source *Entity, x, y float64) Target {
  for _,ent := range source.level.Entities {
    if int(ent.pos.X) == int(x) && int(ent.pos.Y) == int(y) {
      return Target{
        Type : EntityTarget,
        X : int(x),
        Y : int(y),
      }
    }
  }
  return Target{
    Type : NoTarget,
  }
}

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
  for _, target := range source.level.Entities {
    x2 := int(target.pos.X)
    y2 := int(target.pos.Y)

    // Don't let units attack something they don't have LOS to
    if _, ok := source.visible[source.level.toVertex(x2, y2)]; !ok {
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
    ret = append(ret, Target{EntityTarget, x2, y2})
  }
  return ret
}

func cellsWithinRange(source *Entity, rnge int) []Target {
  x := int(source.pos.X)
  y := int(source.pos.Y)
  minx := x - rnge
  if minx < 0 {
    minx = 0
  }
  miny := y - rnge
  if miny < 0 {
    miny = 0
  }
  maxx := x + rnge
  if maxx >= len(source.level.grid) {
    maxx = len(source.level.grid) - 1
  }
  maxy := y + rnge
  if maxy >= len(source.level.grid[0]) {
    maxy = len(source.level.grid[0]) - 1
  }

  var ret []Target
  for x2 := minx; x2 <= maxx; x2++ {
    for y2 := miny; y2 <= maxy; y2++ {
      // Don't let units attack something they don't have LOS to
      if _, ok := source.visible[source.level.toVertex(x2, y2)]; !ok {
        continue
      }

      // The obvious check - must not be further away than our weapon's range
      dist := maxNormi(x, y, x2, y2)
      if rnge < dist {
        continue
      }

      // If it passed all of the above tests then we can include it as a valid
      // target
      ret = append(ret, Target{CellTarget, x2, y2})
    }
  }
  return ret
}

