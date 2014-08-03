package kd_test

import (
	"github.com/orfjackal/gospec/src/gospec"
	"testing"
)

// List of all specs here
func TestAllSpecs(t *testing.T) {
	r := gospec.NewRunner()
	r.AddSpec(PartitionSpec)
	r.AddSpec(TreeSpec)
	gospec.MainGoTest(r, t)
}
