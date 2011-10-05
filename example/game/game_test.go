package game_test

import (
  . "gospec"
  "gospec"
  "game"
  "json"
  "fmt"
)

func WeaponSpecsSpec(c gospec.Context) {
  test_json := []byte(`
    {
      "Instance" : {
        "Base" : "Rifle",
        "Name" : "Ak-47"
      },
      "Weapon" : {
        "StaticCost": 7,
        "StaticRange" : 10,
        "Power" : 25
      }
    }
  `)

  game.RegisterWeapon("Rifle", func() game.Weapon { return new(game.Rifle) })
  game.RegisterWeapon("Bayonet", func() game.Weapon { return new(game.Bayonet) })
  var w game.WeaponSpec
  err := json.Unmarshal(test_json, &w)
  c.Assume(err, Equals, nil)
  rifle,ok := w.Weapon.(*game.Rifle)
  c.Assume(ok, Equals, true)
  c.Expect(rifle.StaticCost, Equals, game.StaticCost(7))
  c.Expect(rifle.StaticRange, Equals, game.StaticRange(10))
  c.Expect(rifle.Power, Equals, 25)
  data,err := json.Marshal(&w)
  c.Assume(err, Equals, nil)
  var w2 game.WeaponSpec
  err = json.Unmarshal(data, &w2)
  c.Assume(err, Equals, nil)
  rifle = w2.Weapon.(*game.Rifle)
  c.Expect(rifle.StaticCost, Equals, game.StaticCost(7))
  c.Expect(rifle.StaticRange, Equals, game.StaticRange(10))
  c.Expect(rifle.Power, Equals, 25)
}
