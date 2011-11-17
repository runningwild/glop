package stats

import(
  "fmt"
  "json"
  "game/base"
)

type terrainVals map[base.Terrain]int

func (m *terrainVals) UnmarshalJSON(data []byte) error {
  standin := make(map[string]int)
  err := json.Unmarshal(data, &standin)
  if err != nil {
    return err
  }
  *m = make(map[base.Terrain]int)
  for k,v := range standin {
    (*m)[base.Terrain(k)] = v
  }
  return nil
}
type Attributes struct {
  // How much less the unit can see if it has to look through this terrain
  // 0 indicates that its vision is not affected by this terrain
  // any unspecified terrain blocks los
  LosMods terrainVals

  // Indicates how many addition AP (beyond the base of 1 per cell) must be
  // used to enter a cell with this terrain any unspecified terrain is
  // impassable
  MoveMods terrainVals

  // Modifiers to combat values - unspecified terrain does not modify comabt
  // values.  It is an error for any unit to have any terrain specified more
  // than once in either AttackMods or DefenseMods
  AttackMods, DefenseMods terrainVals
}

// for every key in b, if the key does not exist in a the key is added to a
// with the same corresponding value as in b.  If the key does exist in a then
// the values of both are passed to the function.  If the function returns
// true the value from b overwrites the value in a.
func smoosh(a,b terrainVals, f func(av,bv int) bool) {
  for key,bv := range b {
    av,ok := a[key]
    if !ok || f(bv,av) {
      a[key] = bv
    }
  }
}

// takes all attributes listed for a unit and combines them by taking the
// best parts of all attributes.
func processAttributes(attlist []string, attmap map[string]Attributes) Attributes {
  var atts Attributes
  atts.LosMods = make(map[base.Terrain]int)
  atts.MoveMods = make(map[base.Terrain]int)
  atts.AttackMods = make(map[base.Terrain]int)
  atts.DefenseMods = make(map[base.Terrain]int)

  for _,name := range attlist {
    terrainVals,ok := attmap[name]
    if !ok {
      panic(fmt.Sprintf("Attribute '%s' could not be found.", name))
    }

    smoosh(atts.LosMods, terrainVals.LosMods, func(a,b int) bool { return a < b })
    smoosh(atts.MoveMods, terrainVals.MoveMods, func(a,b int) bool { return a < b })
    smoosh(atts.AttackMods, terrainVals.AttackMods, func(a,b int) bool { return true })
    smoosh(atts.DefenseMods, terrainVals.DefenseMods, func(a,b int) bool { return true })
  }
  return atts
}

func LoadAttributes(path string) (map[string]Attributes, error) {
  var atts map[string]Attributes
  err := base.LoadJson(path, &atts)
  return atts, err
}
