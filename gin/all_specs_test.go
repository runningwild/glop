package gin_test

import (
  "gospec"
  "testing"
)


func TestAllSpecs(t *testing.T) {
  r := gospec.NewRunner()
  r.AddSpec(BasicInputSpec)
  gospec.MainGoTest(r, t)
}

