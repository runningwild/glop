package game_test

import (
  "gospec"
  "testing"
)


func TestAllSpecs(t *testing.T) {
  r := gospec.NewRunner()
  r.AddSpec(WeaponSpecsSpec)
  gospec.MainGoTest(r, t)
}

