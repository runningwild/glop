package algorithm_test

import (
  "github.com/orfjackal/gospec/src/gospec"
  "testing"
)

func TestAllSpecs(t *testing.T) {
  r := gospec.NewRunner()
  r.AddSpec(DijkstraSpec)
  r.AddSpec(ReachableSpec)
  r.AddSpec(ChooserSpec)
  r.AddSpec(MapperSpec)
  r.AddSpec(TopoSpec)
  gospec.MainGoTest(r, t)
}
