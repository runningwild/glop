package algo_test

import (
  "gospec"
  "testing"
)


func TestAllSpecs(t *testing.T) {
  r := gospec.NewRunner()
  r.AddSpec(GraphSpec)
  gospec.MainGoTest(r, t)
}

