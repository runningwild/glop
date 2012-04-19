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
  r.AddSpec(Chooser2Spec)
  r.AddSpec(MapperSpec)
  r.AddSpec(Mapper2Spec)
  r.AddSpec(TopoSpec)
  gospec.MainGoTest(r, t)
}
