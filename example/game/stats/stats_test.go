package stats_test

import (
  . "gospec"
  "gospec"
  "gob"
  "bytes"
  "game/stats"
  "game/base"
  "fmt"
)

func init() {
  stats.SetAttmap(map[string]stats.Attributes{
      "basic" : stats.Attributes{
        MoveMods : map[base.Terrain]int{
          "grass" : 0,
          "hills" : 2,
        },
      },
      "tag" : stats.Attributes{
      },
    })
}

func StatsSpec(c gospec.Context) {
  stat := stats.MakeStats(10, 10, 10, 10, 10, []string{ "basic", "tag" })

  c.Specify("Dynamic stats stuff", func() {
    c.Expect(stat.BaseAp(), Equals, 10)
    c.Expect(stat.BaseHealth(), Equals, 10)
    c.Expect(stat.CurAp(), Equals, 10)
    c.Expect(stat.CurHealth(), Equals, 10)

    // Make sure that the damage is reflected in the current health, not the
    // base health
    stat.DoDamage(3)
    c.Expect(stat.BaseHealth(), Equals, 10)
    c.Expect(stat.CurHealth(), Equals, 7)
    c.Expect(stat.BaseAp(), Equals, 10)
    c.Expect(stat.CurAp(), Equals, 10)

    // Make sure that the health is not regained at the end of a round
    stat.Round()
    c.Expect(stat.BaseHealth(), Equals, 10)
    c.Expect(stat.CurHealth(), Equals, 7)
    c.Expect(stat.BaseAp(), Equals, 10)
    c.Expect(stat.CurAp(), Equals, 10)

    // Make sure that Ap spending is reflected in the cur stat, not the base
    stat.SpendAp(5)
    c.Expect(stat.BaseHealth(), Equals, 10)
    c.Expect(stat.CurHealth(), Equals, 7)
    c.Expect(stat.BaseAp(), Equals, 10)
    c.Expect(stat.CurAp(), Equals, 5)

    // Make sure that Ap is replenished at the end of a round
    stat.Round()
    c.Expect(stat.BaseHealth(), Equals, 10)
    c.Expect(stat.CurHealth(), Equals, 7)
    c.Expect(stat.BaseAp(), Equals, 10)
    c.Expect(stat.CurAp(), Equals, 10)
  })
}

func init() {
  err := stats.RegisterEffectsFromJson([]byte(`
      [{
        "Type" : "static effect",
        "Name" : "Slow",
        "Int_params" : {
          "TimedEffect" : 2,
          "MoveCost" : 1
        }
      },
      {
        "Type" : "shield",
        "Name" : "Shield2",
        "Int_params" : {
          "Amount" : 2
        }
      },
      {
        "Type" : "shield",
        "Name" : "Shield4",
        "Int_params" : {
          "Amount" : 4
        }
      }]
    `))
  if err != nil {
    panic(fmt.Sprintf("Unable to register effect: %s\n", err.Error()))
  }
}

func MakeEffectsSpec(c gospec.Context) {
  c.Specify("Effects can be made with the MakeEffect function.", func() {
    stats.MakeEffect("Slow")
  })
}

func EffectsSpec(c gospec.Context) {
  stat := stats.MakeStats(10, 10, 10, 10, 10, []string{ "basic", "tag" })
  c.Specify("Movement cost is 1 + whatever is in the MoveMods", func() {
    c.Expect(stat.MoveCost("grass"), Equals, 1)
    c.Expect(stat.MoveCost("hills"), Equals, 3)
  })
  c.Specify("Movement can be affected by effects", func() {
    stat.AddEffect(stats.MakeEffect("Slow"), false)
    c.Expect(stat.MoveCost("grass"), Equals, 2)
    c.Expect(stat.MoveCost("hills"), Equals, 4)
  })
}

func DamageSpec(c gospec.Context) {
  stat := stats.MakeStats(10, 10, 10, 10, 10, []string{ "basic", "tag" })
  c.Specify("Damage and shield effects work and interact properly", func() {
    c.Expect(stat.BaseHealth(), Equals, 10)
    c.Expect(stat.CurHealth(), Equals, 10)
    stat.DoDamage(3)
    c.Expect(stat.CurHealth(), Equals, 7)

    shield := stats.MakeEffect("Shield2")
    stat.AddEffect(shield, false)
    c.Expect(shield.Active(), Equals, true)
    stat.DoDamage(5)
    c.Expect(stat.CurHealth(), Equals, 4)
    c.Expect(shield.Active(), Equals, false)

    shield = stats.MakeEffect("Shield2")
    stat.AddEffect(shield, false)
    c.Expect(shield.Active(), Equals, true)
    stat.DoDamage(1)
    c.Expect(stat.CurHealth(), Equals, 4)
    c.Expect(shield.Active(), Equals, true)
    stat.DoDamage(1)
    c.Expect(stat.CurHealth(), Equals, 4)
    c.Expect(shield.Active(), Equals, false)
    stat.DoDamage(1)
    c.Expect(stat.CurHealth(), Equals, 3)
    c.Expect(shield.Active(), Equals, false)
  })
}

func DupSpec(c gospec.Context) {
  stat := stats.MakeStats(10, 10, 10, 10, 10, []string{ "basic", "tag" })
  c.Specify("Only one effect of any name should be present at a time.", func() {
    c.Expect(stat.BaseHealth(), Equals, 10)
    c.Expect(stat.CurHealth(), Equals, 10)

    stat.AddEffect(stats.MakeEffect("Shield2"), false)
    stat.AddEffect(stats.MakeEffect("Shield2"), false)
    stat.AddEffect(stats.MakeEffect("Shield2"), false)
    stat.DoDamage(10)
    c.Expect(stat.CurHealth(), Equals, 2)
  })
  c.Specify("New effects overwrite old effects by the same name.", func() {
    c.Expect(stat.BaseHealth(), Equals, 10)
    c.Expect(stat.CurHealth(), Equals, 10)

    stat.AddEffect(stats.MakeEffect("Shield2"), false)
    stat.DoDamage(1)
    stat.AddEffect(stats.MakeEffect("Shield2"), false)
    stat.DoDamage(5)
    c.Expect(stat.CurHealth(), Equals, 7)
  })
  c.Specify("Different effects of the same time can coexist.", func() {
    c.Expect(stat.BaseHealth(), Equals, 10)
    c.Expect(stat.CurHealth(), Equals, 10)

    stat.AddEffect(stats.MakeEffect("Shield2"), false)
    stat.AddEffect(stats.MakeEffect("Shield4"), false)
    stat.DoDamage(10)
    c.Expect(stat.CurHealth(), Equals, 6)
  })
}

type container struct {
  S stats.Stats
}

func GobSpec(c gospec.Context) {
  var c1,c2 container
  c1.S = stats.MakeStats(10, 20, 30, 40, 50, []string{ "basic", "tag" })
  buffer := bytes.NewBuffer([]byte{})
  enc := gob.NewEncoder(buffer)
  dec := gob.NewDecoder(buffer)
  c.Specify("Stats can be gobbed and ungobbed without loss of data.", func() {
    err := enc.Encode(c1)
    if err != nil {
      panic(err.Error())
    }
    err = dec.Decode(&c2)
    if err != nil {
      panic(err.Error())
    }
    c.Expect(c1.S.BaseHealth(), Equals, c2.S.BaseHealth())
    c.Expect(c1.S.BaseAp(), Equals, c2.S.BaseAp())
    c.Expect(c1.S.BaseAttack(), Equals, c2.S.BaseAttack())
    c.Expect(c1.S.BaseDefense(), Equals, c2.S.BaseDefense())
    c.Expect(c1.S.BaseLosDist(), Equals, c2.S.BaseLosDist())
    c.Expect(c1.S.CurHealth(), Equals, c2.S.CurHealth())
    c.Expect(c1.S.CurAp(), Equals, c2.S.CurAp())
  })
}
