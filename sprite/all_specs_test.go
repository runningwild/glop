package sprite_test

import (
  "github.com/orfjackal/gospec/src/gospec"
  "testing"
)

func TestAllSpecs(t *testing.T) {
  r := gospec.NewRunner()
  r.AddSpec(LoadSpriteSpec)
  gospec.MainGoTest(r, t)
}
