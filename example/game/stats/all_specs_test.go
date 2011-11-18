package stats_test

import (
  "gospec"
  "testing"
)


func TestAllSpecs(t *testing.T) {
  r := gospec.NewRunner()
  r.AddSpec(StatsSpec)
  r.AddSpec(MakeEffectsSpec)
  r.AddSpec(EffectsSpec)
  r.AddSpec(DamageSpec)
  r.AddSpec(DupSpec)
  gospec.MainGoTest(r, t)
}

