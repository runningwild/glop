package game_test

import (
  . "gospec"
  "gospec"
  "game"
  "json"
  "bytes"
  "fmt"

  "os"
  "path/filepath"
)

const test_json string = `
    {
      "WeaponInstance" : {
        "Base" : "Gun",
        "Name" : "Desert Eagle",
        "Path" : "weapons/eagle.png"
      },
      "BaseWeapon" : {
        "Power" : 50,
        "StaticCost" : 6,
        "EntityRange" : 5
      }
    }
  `

func FooSpec(c gospec.Context) {
  basedir := "/Users/runningwild/code/go-glop/example/data"
  dir, err := os.Open(filepath.Join(basedir, "weapons"))
  if err != nil {
    panic(err.Error())
  }
  names, err := dir.Readdir(0)
  if err != nil {
    panic(err.Error())
  }
  for _,name := range names {
    weapons, err := os.Open(filepath.Join(basedir, "weapons", name.Name))
    if err != nil {
      panic(err.Error())
    }
    err = game.LoadWeaponSpecs(weapons)
    if err != nil {
      panic(err.Error())
    }
  }
}

func WeaponLoadingSpec(c gospec.Context) {
  ws := []game.WeaponSpec{
    game.WeaponSpec{
      WeaponInstance: game.WeaponInstance{
        Base: "Gun",
        Name: "Rifle",
        Path: "weapons/rifle.png",
      },
      BaseWeapon: &game.Gun{
        StaticCost:  5,
        EntityRange: 10,
        Power:       25,
      },
    },
    game.WeaponSpec{
      WeaponInstance: game.WeaponInstance{
        Base: "Gun",
        Name: ".38 Special",
        Path: "weapons/special.png",
      },
      BaseWeapon: &game.Gun{
        StaticCost:  2,
        EntityRange: 5,
        Power:       12,
      },
    },
  }
  data, err := json.Marshal(&ws)
  fmt.Printf("---\n%s\n---", data)
  c.Assume(err, Equals, nil)
  err = game.LoadWeaponSpecs(bytes.NewBuffer(data))
  c.Assume(err, Equals, nil)

  rifle_weapon := game.MakeWeapon("Rifle")
  rifle_spec, ok := rifle_weapon.(game.WeaponSpec)
  c.Assume(ok, Equals, true)
  rifle, ok := rifle_spec.BaseWeapon.(*game.Gun)
  c.Assume(ok, Equals, true)
  c.Expect(rifle.StaticCost, Equals, game.StaticCost(5))
  c.Expect(rifle.EntityRange, Equals, game.EntityRange(10))
  c.Expect(rifle.Power, Equals, 25)

  special_weapon := game.MakeWeapon(".38 Special")
  special_spec, ok := special_weapon.(game.WeaponSpec)
  c.Assume(ok, Equals, true)
  special, ok := special_spec.BaseWeapon.(*game.Gun)
  c.Assume(ok, Equals, true)
  c.Expect(special.StaticCost, Equals, game.StaticCost(2))
  c.Expect(special.EntityRange, Equals, game.EntityRange(5))
  c.Expect(special.Power, Equals, 12)

  //  whoops := game.MakeWeapon("Not a weapon")
  //  c.Expect(whoops, Equals, nil)
}

func WeaponSpecsSpec(c gospec.Context) {
  var w game.WeaponSpec
  err := json.Unmarshal([]byte(test_json), &w)
  c.Assume(err, Equals, nil)
  inst := w.WeaponInstance
  rifle, ok := w.BaseWeapon.(*game.Gun)
  c.Assume(ok, Equals, true)
  c.Expect(inst.Name, Equals, "Desert Eagle")
  c.Expect(inst.Path, Equals, "weapons/eagle.png")
  c.Expect(rifle.StaticCost, Equals, game.StaticCost(6))
  c.Expect(rifle.EntityRange, Equals, game.EntityRange(5))
  c.Expect(rifle.Power, Equals, 50)
  data, err := json.Marshal(&w)
  c.Assume(err, Equals, nil)
  var w2 game.WeaponSpec
  err = json.Unmarshal(data, &w2)
  c.Assume(err, Equals, nil)
  inst = w2.WeaponInstance
  rifle = w2.BaseWeapon.(*game.Gun)
  c.Expect(inst.Name, Equals, "Desert Eagle")
  c.Expect(inst.Path, Equals, "weapons/eagle.png")
  c.Expect(rifle.StaticCost, Equals, game.StaticCost(6))
  c.Expect(rifle.EntityRange, Equals, game.EntityRange(5))
  c.Expect(rifle.Power, Equals, 50)
}
