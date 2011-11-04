package game

import(
  "fmt"
  "path/filepath"
  "os"
  "io/ioutil"
  "json"
)

type mods map[Terrain]int
func (m *mods) UnmarshalJSON(data []byte) error {
  standin := make(map[string]int)
  err := json.Unmarshal(data, &standin)
  if err != nil {
    return err
  }
  *m = make(map[Terrain]int)
  for k,v := range standin {
    (*m)[Terrain(k)] = v
  }
  return nil
}
type Attributes struct {
  // How far the unit is able to see
  LosDistance int

  // How much less the unit can see if it has to look through this terrain
  // 0 indicates that its vision is not affected by this terrain
  // any unspecified terrain blocks los
  LosMods mods

  // Indicates how many addition AP (beyond the base of 1 per cell) must be
  // used to enter a cell with this terrain any unspecified terrain is
  // impassable
  MoveMods mods

  // Modifiers to combat values - unspecified terrain does not modify comabt
  // values.  It is an error for any unit to have any terrain specified more
  // than once in either AttackMods or DefenseMods
  AttackMods, DefenseMods mods
}

// for every key in b, if the key does not exist in a the key is added to a
// with the same corresponding value as in b.  If the key does exist in a then
// the values of both are passed to the function.  If the function returns
// true the value from b overwrites the value in a.
func smoosh(a,b mods, f func(av,bv int) bool) {
  for key,bv := range b {
    av,ok := a[key]
    if !ok || f(bv,av) {
      a[key] = bv
    }
  }
}

// takes all attributes listed for a unit and combines them by taking the
// best parts of all attributes.
func (unit *UnitType) processAttributes(attmap map[string]Attributes) {
  atts := &unit.attributes
  atts.LosDistance = 0
  atts.LosMods = make(map[Terrain]int)
  atts.MoveMods = make(map[Terrain]int)
  atts.AttackMods = make(map[Terrain]int)
  atts.DefenseMods = make(map[Terrain]int)

  for _,name := range unit.Attribute_names {
    mods,ok := attmap[name]
    if !ok {
      panic(fmt.Sprintf("Attribute '%s' specified by UnitType '%s' could not be found.", name, unit.Name))
    }

    if atts.LosDistance < mods.LosDistance {
      atts.LosDistance = mods.LosDistance
    }
    smoosh(atts.LosMods, mods.LosMods, func(a,b int) bool { return a < b })
    smoosh(atts.MoveMods, mods.MoveMods, func(a,b int) bool { return a < b })
    smoosh(atts.AttackMods, mods.AttackMods, func(a,b int) bool { return true })
    smoosh(atts.DefenseMods, mods.DefenseMods, func(a,b int) bool { return true })
  }
}

func loadJson(path string, target interface{}) error {
  f, err := os.Open(path)
  if err != nil {
    return err
  }
  data, err := ioutil.ReadAll(f)
  if err != nil {
    return err
  }
  err = json.Unmarshal(data, target)
  if err != nil {
    return err
  }
  return nil
}

func loadAttributes(dir string) (map[string]Attributes, error) {
  var atts map[string]Attributes
  err := loadJson(filepath.Join(dir, "attributes.json"), &atts)
  return atts, err
}
