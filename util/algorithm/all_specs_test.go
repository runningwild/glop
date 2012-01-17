package algo_test

import (
  "gospec"
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
