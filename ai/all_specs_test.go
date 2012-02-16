package ai_test

import (
  "github.com/orfjackal/gospec/src/gospec"
  "testing"
)


func TestAllSpecs(t *testing.T) {
  r := gospec.NewRunner()
  r.AddSpec(AiSpec)
  r.AddSpec(TermSpec)
  gospec.MainGoTest(r, t)
}

