package ai_test

import (
  . "gospec"
  "gospec"
  "yed"
  "polish"
  "ai"
)

func AiSpec(c gospec.Context) {
  c.Specify("Load a simple .xgml file.", func() {
    g, err := yed.ParseFromFile("state.xgml")
    c.Assume(err, Equals, nil)
    aig := ai.AiGraph{
      Graph: &g.Graph,
      Context: polish.MakeContext(),
    }
    polish.AddIntMathContext(aig.Context)

    dist := 0
    dist_func := func() int {
      return dist
    }

    var nearest int = 7
    nearest_func := func() int {
      return nearest
    }

    attacks := 0
    attack_func := func() int {
      attacks++
      return 0
    }

    aig.Context.AddFunc("dist", dist_func)
    aig.Context.AddFunc("nearest", nearest_func)
    aig.Context.AddFunc("move", func() int { nearest--; return 0 })
    aig.Context.AddFunc("wait", func() int { return 0 })
    aig.Context.AddFunc("attack", attack_func)
    aig.Eval()

    c.Expect(attacks, Equals, 0)
    c.Expect(nearest, Equals, 4)
  })
}