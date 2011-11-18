package stats_test

import (
  . "gospec"
  "gospec"
  "game/stats"
  "game/base"
  "fmt"
)

func StatsSpec(c gospec.Context) {
  b := stats.BaseStats{
    DynamicStats : stats.DynamicStats{
      Health : 10,
      Ap : 10,
    },
    Attack : 10,
    Defense : 10,
    LosDist : 10,
    Atts : []string{
      "basic",
      "tag",
    },
  }
  attmap := map[string]stats.Attributes{
    "basic" : stats.Attributes{
    },
    "tag" : stats.Attributes{
    },
  }
  c.Specify("Dynamic stats stuff", func() {
    stat := stats.MakeStats(b, attmap)
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
  b := stats.BaseStats{
    DynamicStats : stats.DynamicStats{
      Health : 10,
      Ap : 10,
    },
    Attack : 10,
    Defense : 10,
    LosDist : 10,
    Atts : []string{
      "basic",
      "tag",
    },
  }
  attmap := map[string]stats.Attributes{
    "basic" : stats.Attributes{
      MoveMods : map[base.Terrain]int{
        "grass" : 0,
        "hills" : 2,
      },
    },
    "tag" : stats.Attributes{
    },
  }
  c.Specify("Movement cost is 1 + whatever is in the MoveMods", func() {
    stat := stats.MakeStats(b, attmap)
    c.Expect(stat.MoveCost("grass"), Equals, 1)
    c.Expect(stat.MoveCost("hills"), Equals, 3)
  })
  c.Specify("Movement can be affected by effects", func() {
    stat := stats.MakeStats(b, attmap)
    stat.AddEffect(stats.MakeEffect("Slow"), false)
    c.Expect(stat.MoveCost("grass"), Equals, 2)
    c.Expect(stat.MoveCost("hills"), Equals, 4)
  })
}

func DamageSpec(c gospec.Context) {
  b := stats.BaseStats{
    DynamicStats : stats.DynamicStats{
      Health : 10,
      Ap : 10,
    },
  }
  c.Specify("Damage and shield effects work and interact properly", func() {
    stat := stats.MakeStats(b, nil)
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
  b := stats.BaseStats{
    DynamicStats : stats.DynamicStats{
      Health : 10,
      Ap : 10,
    },
  }
  c.Specify("Only one effect of any name should be present at a time.", func() {
    stat := stats.MakeStats(b, nil)
    c.Expect(stat.BaseHealth(), Equals, 10)
    c.Expect(stat.CurHealth(), Equals, 10)

    stat.AddEffect(stats.MakeEffect("Shield2"), false)
    stat.AddEffect(stats.MakeEffect("Shield2"), false)
    stat.AddEffect(stats.MakeEffect("Shield2"), false)
    stat.DoDamage(10)
    c.Expect(stat.CurHealth(), Equals, 2)
  })
  c.Specify("New effects overwrite old effects by the same name.", func() {
    stat := stats.MakeStats(b, nil)
    c.Expect(stat.BaseHealth(), Equals, 10)
    c.Expect(stat.CurHealth(), Equals, 10)

    stat.AddEffect(stats.MakeEffect("Shield2"), false)
    stat.DoDamage(1)
    stat.AddEffect(stats.MakeEffect("Shield2"), false)
    stat.DoDamage(5)
    c.Expect(stat.CurHealth(), Equals, 7)
  })
  c.Specify("Different effects of the same time can coexist.", func() {
    stat := stats.MakeStats(b, nil)
    c.Expect(stat.BaseHealth(), Equals, 10)
    c.Expect(stat.CurHealth(), Equals, 10)

    stat.AddEffect(stats.MakeEffect("Shield2"), false)
    stat.AddEffect(stats.MakeEffect("Shield4"), false)
    stat.DoDamage(10)
    c.Expect(stat.CurHealth(), Equals, 6)
  })
}
