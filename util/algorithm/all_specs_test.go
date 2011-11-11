package algo_test

import (
  "gospec"
  "testing"
)

func TestAllSpecs(t *testing.T) {
  r := gospec.NewRunner()
  r.AddSpec(DijkstraSpec)
  r.AddSpec(ReachableSpec)
  r.AddSpec(GenericSpec)
  gospec.MainGoTest(r, t)
}
