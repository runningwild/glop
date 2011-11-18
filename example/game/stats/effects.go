package stats

import (
  "game/base"
  "fmt"
  "os"
  "strings"
  "io/ioutil"
  "json"
  "reflect"
  "path/filepath"
)

type Effect interface {
  // Two effects with the same Name() are equivalent.  This is used to prevent
  // the same effect from being applied to one unit more than once at a time.
  Name() string

  // Can apply modifiers to BaseStats, but cannot actually change them, since we
  // need to keep the original data around
  ModifyStats(BaseStats) BaseStats

  // Can change DynamicStats however it wants, and may reference BaseStats to do
  // so.
  ModifyDynamicStats(*DynamicStats, BaseStats)

  // Given an amount of damage, returns a modified amount of damage that this
  // unit should take.
  ModifyIncomingDamage(int, BaseStats) int

  // Given an amount of damage, returns a modified amount of damage that this
  // unit should deal.
  ModifyOutgoingDamage(int, BaseStats) int

  // Modifies an Attribute based on the terrain and its current value
  ModifyMovement(base.Terrain, int) int
  ModifyLos(base.Terrain, int) int
  ModifyAttack(base.Terrain, int) int
  ModifyDefense(base.Terrain, int) int

  // Called at the end of every round.
  Round()

  // Should return false when this effect is done so that it can be removed.
  Active() bool
}

type NullEffect struct {
  Id string
}
func (n NullEffect) Name() string { return n.Id }
func (NullEffect) ModifyStats(b BaseStats) BaseStats { return b }
func (NullEffect) ModifyDynamicStats(*DynamicStats, BaseStats) { }
func (NullEffect) ModifyIncomingDamage(dmg int, b BaseStats) int { return dmg }
func (NullEffect) ModifyOutgoingDamage(dmg int, b BaseStats) int { return dmg }
func (NullEffect) ModifyMovement(t base.Terrain, n int) int { return n }
func (NullEffect) ModifyLos(t base.Terrain, n int) int { return n }
func (NullEffect) ModifyAttack(t base.Terrain, n int) int { return n }
func (NullEffect) ModifyDefense(t base.Terrain, n int) int { return n }

// TimedEffect provides an easy way to make an effect last for a specific amount
// of time.  For an effect to only last until the end of the turn set
// rounds = 1, for it to last until the end of the next turn set rounds = 2
type TimedEffect int
func (e *TimedEffect) Round() {
  (*e)--
}
func (e *TimedEffect) Active() bool {
  return (*e) > 0
}

func init() {
  registerEffectType("static effect", &StaticEffect{})
}
type StaticEffect struct {
  NullEffect
  TimedEffect
  Attack  int
  Defense int
  LosDist int
  Health  int
  Ap      int

  MoveCost int
  LosCost  int
}
func (e *StaticEffect) ModifyStats(b BaseStats) BaseStats {
  b.Attack += e.Attack
  b.Defense += e.Defense
  b.LosDist += e.LosDist
  return b
}
func (e *StaticEffect) ModifyDynamicStats(d *DynamicStats, b BaseStats) {
  d.Health += e.Health
  d.Ap += e.Ap
}
func (e *StaticEffect) ModifyMovement(_ base.Terrain, cost int) int {
  return cost + e.MoveCost
}
func (e *StaticEffect) ModifyLos(_ base.Terrain, cost int) int {
  return cost + e.LosCost
}

func init() {
  registerEffectType("shield", &ShieldEffect{})
}
type ShieldEffect struct {
  NullEffect
  Amount int
}
func (e *ShieldEffect) ModifyIncomingDamage(dmg int, b BaseStats) int {
  if e.Amount >= dmg {
    e.Amount -= dmg
    return 0
  }
  dmg -= e.Amount
  e.Amount = 0
  return dmg
}
func (e *ShieldEffect) Round() { }
func (e *ShieldEffect) Active() bool {
  return e.Amount > 0
}

type EffectSpec struct {
  Type       string
  Name       string
  Int_params map[string]int
}

var effect_type_registry map[string]reflect.Type

func registerEffectType(name string, effect Effect) {
  if effect_type_registry == nil {
    effect_type_registry = make(map[string]reflect.Type)
  }
  effect_type_registry[name] = reflect.TypeOf(effect).Elem()
}

func assignParams(effect_val reflect.Value, name string, int_params map[string]int) {
  reflect.Indirect(effect_val).FieldByName("Id").Set(reflect.ValueOf(name))
  for k,v := range int_params {
    field := reflect.Indirect(effect_val).FieldByName(k)
    if field.Kind() == reflect.Invalid {
      panic(fmt.Sprintf("int param '%s' specified, but corresponding field was not found.", k))
    }
    if field.Kind() != reflect.Int {
      panic(fmt.Sprintf("int param '%s' specified, but field is of type %s.", k, field.Kind()))
    }
    field.SetInt(int64(v))
  }
}

var effect_spec_registry map[string]EffectSpec

// TODO: This is basically so that we can test, or potentially reload specs on
// the fly, without having to deal with passing around a stats registry object.
// Singletons are a bad idea though, so maybe this should change.
func ClearSpecRegistry() {
    effect_spec_registry = nil
}
func registerEffectSpec(spec EffectSpec) {
  if effect_spec_registry == nil {
    effect_spec_registry = make(map[string]EffectSpec)
  }
  if _,ok := effect_spec_registry[spec.Name]; ok {
    panic(fmt.Sprintf("Tried to register the effect spec '%s' more than once.", spec.Name))
  }
  if _,ok := effect_type_registry[spec.Type]; !ok {
    panic(fmt.Sprintf("Tried to register an effect spec for the effect type '%s', which doesn't exist.", spec.Type))
  }
  effect_spec_registry[spec.Name] = spec

  // Test to make sure this thing can really make an effect without panicing,
  // this way we fail fast.
  MakeEffect(spec.Name)
}

func MakeEffect(spec_name string) Effect {
  spec,ok := effect_spec_registry[spec_name]
  if !ok {
    panic(fmt.Sprintf("Tried to load an unknown EffectSpec '%s'.", spec_name))
  }
  effect := reflect.New(effect_type_registry[spec.Type])
  assignParams(effect, spec_name, spec.Int_params)
  return effect.Interface().(Effect)
}

// Exposed mostly for testing
func RegisterEffectsFromJson(data []byte) error {
  var specs []EffectSpec
  err := json.Unmarshal(data, &specs)
  if err != nil { return err }
  for _,spec := range specs {
    registerEffectSpec(spec)
  }
  return nil
}

// Finds all *.json files in dir and registers all effects found in those files
func RegisterAllEffectsInDir(dir string) {
  err := filepath.Walk(dir, func(path string, info *os.FileInfo, err error) error {
    if info.IsDirectory() {
      return nil
    }
    if !strings.HasSuffix(path, ".json") {
      return nil
    }
    f,err := os.Open(path)
    if err != nil { return err }
    data,err := ioutil.ReadAll(f)
    f.Close()
    if err != nil { return err }
    err = RegisterEffectsFromJson(data)
    return err
  })
  if err != nil {
    panic(fmt.Sprintf("Unable to load all specs in directory '%s': %s\n", dir, err.Error()))
  }
}
