package gin_test

import (
  "gospec"
  "testing"
)


func TestAllSpecs(t *testing.T) {
  r := gospec.NewRunner()
  r.AddSpec(NaturalKeySpec)
  r.AddSpec(DerivedKeySpec)
  r.AddSpec(NestedDerivedKeySpec)
  gospec.MainGoTest(r, t)
}

