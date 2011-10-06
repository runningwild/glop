package game_test

import (
  . "gospec"
  "gospec"
  "game"
  "json"
  "bytes"
  "fmt"
)

const test_json string = `
    {
      "Instance" : {
        "Base" : "Gun",
        "Name" : "Ak-47"
      },
      "Weapon" : {
        "StaticCost": 7,
        "StaticRange" : 10,
        "Power" : 25
      }
    }
  `


func WeaponLoadingSpec(c gospec.Context) {
  ws := []game.WeaponSpec {
    game.WeaponSpec {
      Instance : game.WeaponInstance {
        Base : "Gun",
        Name : "Rifle",
      },
      Weapon : &game.Gun {
        StaticCost : 5,
        StaticRange : 10,
        Power : 25,
      },
    },
    game.WeaponSpec {
      Instance : game.WeaponInstance {
        Base : "Gun",
        Name : ".38 Special",
      },
      Weapon : &game.Gun {
        StaticCost : 2,
        StaticRange : 5,
        Power : 12,
      },
    },
  }
  data,err := json.Marshal(&ws)
  fmt.Printf("---\n%s\n---", data)
  c.Assume(err, Equals, nil)
  err = game.LoadWeaponSpecs(bytes.NewBuffer(data))
  c.Assume(err, Equals, nil)

  rifle_weapon := game.MakeWeapon("Rifle")
  rifle,ok := rifle_weapon.(*game.Gun)
  c.Assume(ok, Equals, true)
  c.Expect(rifle.StaticCost, Equals, game.StaticCost(5))
  c.Expect(rifle.StaticRange, Equals, game.StaticRange(10))
  c.Expect(rifle.Power, Equals, 25)

  special_weapon := game.MakeWeapon(".38 Special")
  special,ok := special_weapon.(*game.Gun)
  c.Assume(ok, Equals, true)
  c.Expect(special.StaticCost, Equals, game.StaticCost(2))
  c.Expect(special.StaticRange, Equals, game.StaticRange(5))
  c.Expect(special.Power, Equals, 12)

  whoops := game.MakeWeapon("Not a weapon")
  c.Expect(whoops, Equals, nil)
}

func WeaponSpecsSpec(c gospec.Context) {
  var w game.WeaponSpec
  err := json.Unmarshal([]byte(test_json), &w)
  c.Assume(err, Equals, nil)
  rifle,ok := w.Weapon.(*game.Gun)
  c.Assume(ok, Equals, true)
  c.Expect(rifle.StaticCost, Equals, game.StaticCost(7))
  c.Expect(rifle.StaticRange, Equals, game.StaticRange(10))
  c.Expect(rifle.Power, Equals, 25)
  data,err := json.Marshal(&w)
  c.Assume(err, Equals, nil)
  var w2 game.WeaponSpec
  err = json.Unmarshal(data, &w2)
  c.Assume(err, Equals, nil)
  rifle = w2.Weapon.(*game.Gun)
  c.Expect(rifle.StaticCost, Equals, game.StaticCost(7))
  c.Expect(rifle.StaticRange, Equals, game.StaticRange(10))
  c.Expect(rifle.Power, Equals, 25)
}
