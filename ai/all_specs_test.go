package ai_test

import (
  "gospec"
  "testing"
)


func TestAllSpecs(t *testing.T) {
  r := gospec.NewRunner()
  r.AddSpec(AiSpec)
  r.AddSpec(TermSpec)
  gospec.MainGoTest(r, t)
}

