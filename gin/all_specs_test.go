package gin_test

import (
	"github.com/orfjackal/gospec/src/gospec"
	"testing"
)

func TestAllSpecs(t *testing.T) {
	r := gospec.NewRunner()
	r.AddSpec(NaturalKeySpec)
	r.AddSpec(DerivedKeyBugSpec)
	r.AddSpec(DerivedKeySpec)
	r.AddSpec(DeviceSpec)
	r.AddSpec(DeviceFamilySpec)
	r.AddSpec(NestedDerivedKeySpec)
	r.AddSpec(EventSpec)
	r.AddSpec(GeneralSpec)
	r.AddSpec(AxisSpec)
	r.AddSpec(EventListenerSpec)
	r.AddSpec(FocusSpec)
	gospec.MainGoTest(r, t)
}
