package gin_test

import (
	. "github.com/orfjackal/gospec/src/gospec"
	"github.com/orfjackal/gospec/src/gospec"
	"github.com/runningwild/glop/gin"
	"strings"
)

func injectEvent(events *[]gin.OsEvent, key_index gin.KeyIndex, device_index gin.DeviceIndex, device_type gin.DeviceType, amt float64, timestamp int64) {
	*events = append(*events,
		gin.OsEvent{
			KeyId: gin.KeyId{
				Index: key_index,
				Device: gin.DeviceId{
					Index: device_index,
					Type:  device_type,
				},
			},
			Press_amt: amt,
			Timestamp: timestamp,
		},
	)
}

func NaturalKeySpec(c gospec.Context) {
	input := gin.Make()
	keya := input.GetKeyFlat(gin.KeyA, gin.DeviceTypeKeyboard, 1)
	keyb := input.GetKeyFlat(gin.KeyB, gin.DeviceTypeKeyboard, 1)
	c.Specify("Single key press or release per frame sets basic keyState values properly.", func() {

		events := make([]gin.OsEvent, 0)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 5)
		input.Think(10, false, events)
		c.Expect(keya.FramePressCount(), Equals, 1)
		c.Expect(keya.FrameReleaseCount(), Equals, 0)
		c.Expect(keyb.FramePressCount(), Equals, 0)
		c.Expect(keyb.FrameReleaseCount(), Equals, 0)

		events = events[0:0]
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 15)
		input.Think(20, false, events)
		c.Expect(keya.FramePressCount(), Equals, 0)
		c.Expect(keya.FrameReleaseCount(), Equals, 0)
		c.Expect(keyb.FramePressCount(), Equals, 1)
		c.Expect(keyb.FrameReleaseCount(), Equals, 0)

		events = events[0:0]
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 0, 25)
		input.Think(30, false, events)
		c.Expect(keya.FramePressCount(), Equals, 0)
		c.Expect(keya.FrameReleaseCount(), Equals, 1)
		c.Expect(keyb.FramePressCount(), Equals, 0)
		c.Expect(keyb.FrameReleaseCount(), Equals, 0)
	})

	c.Specify("Multiple key presses in a single frame work.", func() {
		events := make([]gin.OsEvent, 0)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 4)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 0, 5)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 6)
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 7)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 0, 8)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 9)
		input.Think(10, false, events)
		c.Expect(keya.FramePressCount(), Equals, 3)
		c.Expect(keya.FrameReleaseCount(), Equals, 2)
		c.Expect(keyb.FramePressCount(), Equals, 1)
		c.Expect(keyb.FrameReleaseCount(), Equals, 0)
	})

	c.Specify("Redundant events don't generate redundant events.", func() {
		events := make([]gin.OsEvent, 0)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 4)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 5)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 6)
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 7)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 0, 8)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 0, 9)
		input.Think(10, false, events)
		c.Expect(keya.FramePressCount(), Equals, 1)
		c.Expect(keya.FrameReleaseCount(), Equals, 1)
	})

	c.Specify("Key.FramePressSum() works.", func() {
		events := make([]gin.OsEvent, 0)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 3)
		input.Think(10, false, events)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 0, 14)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 16)
		input.Think(20, false, events)
		c.Expect(keya.FramePressSum(), Equals, 8.0)

		events = events[0:0]
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 22)
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 0, 24)
		input.Think(30, false, events)
		c.Expect(keyb.FramePressSum(), Equals, 2.0)

		events = events[0:0]
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 35)
		input.Think(40, false, events)
		c.Expect(keyb.FramePressSum(), Equals, 5.0)
	})

	c.Specify("Key.FramePressAvg() works.", func() {
		events := make([]gin.OsEvent, 0)
		input.Think(10, false, events)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 10)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 0, 12)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 14)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 0, 16)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 18)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 0, 20)
		input.Think(20, false, events)
		c.Expect(keya.FramePressAvg(), Equals, 0.6)

		events = events[0:0]
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 25)
		input.Think(30, false, events)
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 0, 35)
		input.Think(40, false, events)
		c.Expect(keyb.FramePressAvg(), Equals, 0.5)
	})
}

func DerivedKeySpec(c gospec.Context) {
	input := gin.Make()
	keya := input.GetKeyFlat(gin.KeyA, gin.DeviceTypeKeyboard, 1)
	keyb := input.GetKeyFlat(gin.KeyB, gin.DeviceTypeKeyboard, 1)
	keyc := input.GetKeyFlat(gin.KeyC, gin.DeviceTypeKeyboard, 1)
	keye := input.GetKeyFlat(gin.KeyE, gin.DeviceTypeKeyboard, 1)
	keyf := input.GetKeyFlat(gin.KeyF, gin.DeviceTypeKeyboard, 1)
	ABc_binding := input.MakeBinding(keya.Id(), []gin.KeyId{keyb.Id(), keyc.Id()}, []bool{true, false})
	Ef_binding := input.MakeBinding(keye.Id(), []gin.KeyId{keyf.Id()}, []bool{false})
	ABc_Ef := input.BindDerivedKey("ABc_Ef", ABc_binding, Ef_binding)

	// ABc_Ef should be down if either ab and not c, or e and not f (or both)
	// That is to say the following (and no others) should all trigger it:
	// (Capital letter indicates the key is down, lowercase indicate it is not)
	// A B c e f
	// A B c e F
	// A B c E f
	// A B c E F
	// a b c E f
	// a b C E f
	// a B c E f
	// a B C E f
	// A b c E f
	// A b C E f
	// A B C E f

	c.Specify("Derived key presses happen only when a primary key is pressed after all modifiers are set.", func() {

		// Test that first binding can trigger a press
		events := make([]gin.OsEvent, 0)
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 1)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 1)
		input.Think(10, false, events)
		c.Expect(ABc_Ef.FramePressAmt(), Equals, 1.0)
		c.Expect(ABc_Ef.IsDown(), Equals, true)
		c.Expect(ABc_Ef.FramePressCount(), Equals, 1)
		events = events[0:0]

		c.Specify("Release happens once primary key is released", func() {
			injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 0, 11)
			input.Think(20, false, events)
			c.Expect(ABc_Ef.IsDown(), Equals, false)
			c.Expect(ABc_Ef.FramePressCount(), Equals, 0)
			c.Expect(ABc_Ef.FrameReleaseCount(), Equals, 1)
		})

		c.Specify("Key remains down when when a down modifier is released", func() {
			injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 0, 11)
			input.Think(20, false, events)
			c.Expect(ABc_Ef.IsDown(), Equals, true)
			c.Expect(ABc_Ef.FramePressCount(), Equals, 0)
			c.Expect(ABc_Ef.FrameReleaseCount(), Equals, 0)
		})

		c.Specify("Key remains down when an up modifier is pressed", func() {
			injectEvent(&events, 'c', 1, gin.DeviceTypeKeyboard, 1, 11)
			input.Think(20, false, events)
			c.Expect(ABc_Ef.IsDown(), Equals, true)
			c.Expect(ABc_Ef.FramePressCount(), Equals, 0)
			c.Expect(ABc_Ef.FrameReleaseCount(), Equals, 0)
		})
		c.Specify("Release isn't affect by bindings changing states first", func() {
			c.Specify("releasing b", func() {
				injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 0, 11)
				injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 0, 11)
				input.Think(20, false, events)
				c.Expect(ABc_Ef.IsDown(), Equals, false)
				c.Expect(ABc_Ef.FramePressCount(), Equals, 0)
				c.Expect(ABc_Ef.FrameReleaseCount(), Equals, 1)
			})
			c.Specify("pressing c", func() {
				injectEvent(&events, 'c', 1, gin.DeviceTypeKeyboard, 1, 11)
				injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 0, 11)
				input.Think(20, false, events)
				c.Expect(ABc_Ef.IsDown(), Equals, false)
				c.Expect(ABc_Ef.FramePressCount(), Equals, 0)
				c.Expect(ABc_Ef.FrameReleaseCount(), Equals, 1)
			})
		})

		c.Specify("Pressing a second binding should not generate another press on the derived key", func() {
			injectEvent(&events, 'e', 1, gin.DeviceTypeKeyboard, 1, 11)
			input.Think(20, false, events)
			c.Expect(ABc_Ef.IsDown(), Equals, true)
			c.Expect(ABc_Ef.FramePressCount(), Equals, 0)
			c.Expect(ABc_Ef.FrameReleaseCount(), Equals, 0)
		})

		// Reset keys
		events = events[0:0]
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 0, 21)
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 0, 21)
		injectEvent(&events, 'c', 1, gin.DeviceTypeKeyboard, 0, 21)
		injectEvent(&events, 'e', 1, gin.DeviceTypeKeyboard, 0, 21)
		input.Think(30, false, events)
		c.Expect(ABc_Ef.FramePressAmt(), Equals, 0.0)
		c.Expect(ABc_Ef.IsDown(), Equals, false)

		// Test that second binding can trigger a press
		events = events[0:0]
		injectEvent(&events, 'e', 1, gin.DeviceTypeKeyboard, 1, 31)
		input.Think(40, false, events)
		c.Expect(ABc_Ef.FramePressAmt(), Equals, 1.0)
		c.Expect(ABc_Ef.IsDown(), Equals, true)
		c.Expect(ABc_Ef.FramePressCount(), Equals, 1)

		// Reset keys
		events = events[0:0]
		injectEvent(&events, 'e', 1, gin.DeviceTypeKeyboard, 0, 41)
		input.Think(50, false, events)

		// Test that first binding doesn't trigger a press if modifiers aren't set first
		events = events[0:0]
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 51)
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 51)
		input.Think(60, false, events)
		c.Expect(ABc_Ef.IsDown(), Equals, false)
		c.Expect(ABc_Ef.FramePressCount(), Equals, 0)
	})
}

// Check that derived keys can properly differentiate between the same key
// pressed on different devices.
func DeviceSpec(c gospec.Context) {
	input := gin.Make()
	keya1 := input.GetKeyFlat(gin.KeyA, gin.DeviceTypeKeyboard, 1)
	keya2 := input.GetKeyFlat(gin.KeyA, gin.DeviceTypeKeyboard, 2)
	keya3 := input.GetKeyFlat(gin.KeyA, gin.DeviceTypeKeyboard, 3)
	A1_binding := input.MakeBinding(keya1.Id(), nil, nil)
	A2_binding := input.MakeBinding(keya2.Id(), nil, nil)
	A3_binding := input.MakeBinding(keya3.Id(), nil, nil)
	A1 := input.BindDerivedKey("A1", A1_binding)
	A2 := input.BindDerivedKey("A2", A2_binding)
	A3 := input.BindDerivedKey("A3", A3_binding)

	keya_any := input.GetKeyFlat(gin.KeyA, gin.DeviceTypeKeyboard, 0)
	AAny_binding := input.MakeBinding(keya_any.Id(), nil, nil)
	AAny := input.BindDerivedKey("AAny", AAny_binding)
	keyb_any := input.GetKeyFlat(gin.KeyB, gin.DeviceTypeKeyboard, 0)
	BAny_binding := input.MakeBinding(keyb_any.Id(), nil, nil)
	BAny := input.BindDerivedKey("BAny", BAny_binding)

	any_key_on_1 := input.GetKeyFlat(gin.AnyKey, gin.DeviceTypeKeyboard, 1)
	any_key_on_1_binding := input.MakeBinding(any_key_on_1.Id(), nil, nil)
	Any1 := input.BindDerivedKey("Any1", any_key_on_1_binding)
	any_key_on_2 := input.GetKeyFlat(gin.AnyKey, gin.DeviceTypeKeyboard, 2)
	any_key_on_2_binding := input.MakeBinding(any_key_on_2.Id(), nil, nil)
	Any2 := input.BindDerivedKey("Any2", any_key_on_2_binding)
	any_key_on_3 := input.GetKeyFlat(gin.AnyKey, gin.DeviceTypeKeyboard, 3)
	any_key_on_3_binding := input.MakeBinding(any_key_on_3.Id(), nil, nil)
	Any3 := input.BindDerivedKey("Any3", any_key_on_3_binding)

	the_any_key := input.GetKeyFlat(gin.AnyKey, gin.DeviceTypeAny, 0)
	the_any_key_binding := input.MakeBinding(the_any_key.Id(), nil, nil)
	Any_key := input.BindDerivedKey("Any Key", the_any_key_binding)

	c.Specify("Derived keys trigger from the specified devices and no others.", func() {
		// Test that first binding can trigger a press
		events := make([]gin.OsEvent, 0)
		injectEvent(&events, gin.KeyA, 1, gin.DeviceTypeKeyboard, 1, 1)
		input.Think(2, false, events)
		c.Expect(A1.IsDown(), Equals, true)
		c.Expect(A2.IsDown(), Equals, false)
		c.Expect(A3.IsDown(), Equals, false)

		injectEvent(&events, gin.KeyA, 1, gin.DeviceTypeKeyboard, 0, 3)
		injectEvent(&events, gin.KeyA, 2, gin.DeviceTypeKeyboard, 1, 4)
		input.Think(5, false, events)
		c.Expect(A1.IsDown(), Equals, false)
		c.Expect(A2.IsDown(), Equals, true)
		c.Expect(A3.IsDown(), Equals, false)

		injectEvent(&events, gin.KeyA, 2, gin.DeviceTypeKeyboard, 0, 6)
		injectEvent(&events, gin.KeyA, 3, gin.DeviceTypeKeyboard, 1, 7)
		input.Think(8, false, events)
		c.Expect(A1.IsDown(), Equals, false)
		c.Expect(A2.IsDown(), Equals, false)
		c.Expect(A3.IsDown(), Equals, true)
		events = events[0:0]
	})

	c.Specify("Derived key can specify a specific key on any device.", func() {
		// Test that first binding can trigger a press
		events := make([]gin.OsEvent, 0)
		injectEvent(&events, gin.KeyA, 1, gin.DeviceTypeKeyboard, 1, 1)
		input.Think(2, false, events)
		c.Expect(AAny.IsDown(), Equals, true)
		c.Expect(BAny.IsDown(), Equals, false)

		injectEvent(&events, gin.KeyA, 1, gin.DeviceTypeKeyboard, 0, 3)
		injectEvent(&events, gin.KeyA, 2, gin.DeviceTypeKeyboard, 1, 4)
		input.Think(5, false, events)
		c.Expect(AAny.IsDown(), Equals, true)
		c.Expect(BAny.IsDown(), Equals, false)

		injectEvent(&events, gin.KeyA, 2, gin.DeviceTypeKeyboard, 0, 6)
		injectEvent(&events, gin.KeyB, 1, gin.DeviceTypeKeyboard, 1, 7)
		input.Think(8, false, events)
		c.Expect(AAny.IsDown(), Equals, false)
		c.Expect(BAny.IsDown(), Equals, true)

		injectEvent(&events, gin.KeyB, 1, gin.DeviceTypeKeyboard, 0, 9)
		injectEvent(&events, gin.KeyB, 2, gin.DeviceTypeKeyboard, 1, 10)
		input.Think(11, false, events)
		c.Expect(AAny.IsDown(), Equals, false)
		c.Expect(BAny.IsDown(), Equals, true)
		events = events[0:0]
	})

	c.Specify("Derived key can specify any key on a specific device.", func() {
		// Test that first binding can trigger a press
		events := make([]gin.OsEvent, 0)
		injectEvent(&events, gin.KeyA, 1, gin.DeviceTypeKeyboard, 1, 1)
		input.Think(2, false, events)
		c.Expect(Any1.IsDown(), Equals, true)
		c.Expect(Any2.IsDown(), Equals, false)
		c.Expect(Any3.IsDown(), Equals, false)

		injectEvent(&events, gin.KeyA, 1, gin.DeviceTypeKeyboard, 0, 3)
		injectEvent(&events, gin.KeyA, 2, gin.DeviceTypeKeyboard, 1, 3)
		input.Think(4, false, events)
		c.Expect(Any1.IsDown(), Equals, false)
		c.Expect(Any2.IsDown(), Equals, true)
		c.Expect(Any3.IsDown(), Equals, false)

		injectEvent(&events, gin.KeyA, 2, gin.DeviceTypeKeyboard, 0, 5)
		injectEvent(&events, gin.KeyA, 3, gin.DeviceTypeKeyboard, 1, 5)
		input.Think(6, false, events)
		c.Expect(Any1.IsDown(), Equals, false)
		c.Expect(Any2.IsDown(), Equals, false)
		c.Expect(Any3.IsDown(), Equals, true)

		events = events[0:0]
	})

	c.Specify("Derived key can specify the any key.", func() {
		// Test that first binding can trigger a press
		events := make([]gin.OsEvent, 0)
		injectEvent(&events, gin.KeyA, 1, gin.DeviceTypeKeyboard, 1, 1)
		input.Think(2, false, events)
		c.Expect(Any_key.IsDown(), Equals, true)
		c.Expect(Any_key.FramePressCount(), Equals, 1)
		c.Expect(Any_key.FrameReleaseCount(), Equals, 0)

		injectEvent(&events, gin.KeyB, 2, gin.DeviceTypeKeyboard, 1, 3)
		input.Think(4, false, events)
		c.Expect(Any_key.IsDown(), Equals, true)
		c.Expect(Any_key.FramePressCount(), Equals, 0)
		c.Expect(Any_key.FrameReleaseCount(), Equals, 0)

		injectEvent(&events, gin.KeyA, 1, gin.DeviceTypeKeyboard, 0, 5)
		input.Think(6, false, events)
		c.Expect(Any_key.IsDown(), Equals, true)
		c.Expect(Any_key.FramePressCount(), Equals, 0)
		c.Expect(Any_key.FrameReleaseCount(), Equals, 0)

		injectEvent(&events, gin.KeyB, 2, gin.DeviceTypeKeyboard, 0, 7)
		input.Think(8, false, events)
		c.Expect(Any_key.IsDown(), Equals, false)
		c.Expect(Any_key.FramePressCount(), Equals, 0)
		c.Expect(Any_key.FrameReleaseCount(), Equals, 1)
		events = events[0:0]
	})
}

func NestedDerivedKeySpec(c gospec.Context) {
	input := gin.Make()
	keya := input.GetKeyFlat(gin.KeyA, gin.DeviceTypeKeyboard, 1)
	keyb := input.GetKeyFlat(gin.KeyB, gin.DeviceTypeKeyboard, 1)
	keyc := input.GetKeyFlat(gin.KeyC, gin.DeviceTypeKeyboard, 1)
	AB_binding := input.MakeBinding(keya.Id(), []gin.KeyId{keyb.Id()}, []bool{true})
	AB := input.BindDerivedKey("AB", AB_binding)
	AB_C_binding := input.MakeBinding(keyc.Id(), []gin.KeyId{AB.Id()}, []bool{true})
	AB_C := input.BindDerivedKey("AB_C", AB_C_binding)
	events := make([]gin.OsEvent, 0)

	check := func(order string) {
		input.Think(10, false, events)
		if strings.Index(order, "b") < strings.Index(order, "a") {
			c.Expect(AB.IsDown(), Equals, true)
			c.Expect(AB.FramePressCount(), Equals, 1)
		} else {
			c.Expect(AB.IsDown(), Equals, false)
			c.Expect(AB.FramePressCount(), Equals, 0)
		}
		if order == "bac" {
			c.Expect(AB_C.IsDown(), Equals, true)
			c.Expect(AB_C.FramePressCount(), Equals, 1)
		} else {
			c.Expect(AB_C.IsDown(), Equals, false)
			c.Expect(AB_C.FramePressCount(), Equals, 0)
		}
	}

	c.Specify("Nested derived keys work like normal derived keys.", func() {
		c.Specify("Testing order 'bac'.", func() {
			injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 1)
			injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 1)
			injectEvent(&events, 'c', 1, gin.DeviceTypeKeyboard, 1, 1)
			check("bac")
		})
		c.Specify("Testing order 'abc'.", func() {
			injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 1)
			injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 1)
			injectEvent(&events, 'c', 1, gin.DeviceTypeKeyboard, 1, 1)
			check("abc")
		})
		c.Specify("Testing order 'acb'.", func() {
			injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 1)
			injectEvent(&events, 'c', 1, gin.DeviceTypeKeyboard, 1, 1)
			injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 1)
			check("acb")
		})
		c.Specify("Testing order 'bca'.", func() {
			injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 1)
			injectEvent(&events, 'c', 1, gin.DeviceTypeKeyboard, 1, 1)
			injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 1)
			check("bca")
		})
		c.Specify("Testing order 'cab'.", func() {
			injectEvent(&events, 'c', 1, gin.DeviceTypeKeyboard, 1, 1)
			injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 1)
			injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 1)
			check("cab")
		})
		c.Specify("Testing order 'cba'.", func() {
			injectEvent(&events, 'c', 1, gin.DeviceTypeKeyboard, 1, 1)
			injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 1)
			injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 1)
			check("cba")
		})
	})
}

func EventSpec(c gospec.Context) {
	input := gin.Make()
	keya := input.GetKeyFlat(gin.KeyA, gin.DeviceTypeKeyboard, 1)
	keyb := input.GetKeyFlat(gin.KeyB, gin.DeviceTypeKeyboard, 1)
	keyc := input.GetKeyFlat(gin.KeyC, gin.DeviceTypeKeyboard, 1)
	keyd := input.GetKeyFlat(gin.KeyD, gin.DeviceTypeKeyboard, 1)
	AB_binding := input.MakeBinding(keya.Id(), []gin.KeyId{keyb.Id()}, []bool{true})
	AB := input.BindDerivedKey("AB", AB_binding)
	CD_binding := input.MakeBinding(keyc.Id(), []gin.KeyId{keyd.Id()}, []bool{true})
	CD := input.BindDerivedKey("CD", CD_binding)
	AB_CD_binding := input.MakeBinding(AB.Id(), []gin.KeyId{CD.Id()}, []bool{true})
	_ = input.BindDerivedKey("AB CD", AB_CD_binding)
	events := make([]gin.OsEvent, 0)

	check := func(lengths ...int) {
		groups := input.Think(10, false, events)
		c.Assume(len(groups), Equals, len(lengths))
		for i, length := range lengths {
			// To make it more clear what each test is doing we only check for the
			// number of natural key and derived key events generated on each
			// press/release, i.e. we don't count general key events.
			natural_events := 0
			for _, event := range groups[i].Events {
				if event.Key.Id().Index == gin.AnyKey ||
					event.Key.Id().Device.Type == gin.DeviceTypeAny ||
					event.Key.Id().Device.Index == 0 {
					continue
				}
				natural_events++
			}
			c.Assume(natural_events, Equals, length)
		}
	}

	c.Specify("Testing order 'abcd'.", func() {
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 1)
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 2)
		injectEvent(&events, 'c', 1, gin.DeviceTypeKeyboard, 1, 3)
		injectEvent(&events, 'd', 1, gin.DeviceTypeKeyboard, 1, 4)
		check(1, 1, 1, 1)
	})

	c.Specify("Testing order 'dbca'.", func() {
		injectEvent(&events, 'd', 1, gin.DeviceTypeKeyboard, 1, 1)
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 2)
		injectEvent(&events, 'c', 1, gin.DeviceTypeKeyboard, 1, 3)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 4)
		check(1, 1, 2, 3)
	})

	c.Specify("Testing order 'dcba'.", func() {
		injectEvent(&events, 'd', 1, gin.DeviceTypeKeyboard, 1, 1)
		injectEvent(&events, 'c', 1, gin.DeviceTypeKeyboard, 1, 2)
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 3)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 4)
		check(1, 2, 1, 3)
	})

	c.Specify("Testing order 'bcda'.", func() {
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 1)
		injectEvent(&events, 'c', 1, gin.DeviceTypeKeyboard, 1, 2)
		injectEvent(&events, 'd', 1, gin.DeviceTypeKeyboard, 1, 3)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 4)
		check(1, 1, 1, 2)
	})

	// This test also checks that a derived key stays down until the primary key is released
	// CD is used here after D is released to trigger AB_CD
	c.Specify("Testing order 'dcbDad'.", func() {
		injectEvent(&events, 'd', 1, gin.DeviceTypeKeyboard, 1, 1)
		injectEvent(&events, 'c', 1, gin.DeviceTypeKeyboard, 1, 2)
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 3)
		injectEvent(&events, 'd', 1, gin.DeviceTypeKeyboard, 0, 4)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 5)
		injectEvent(&events, 'd', 1, gin.DeviceTypeKeyboard, 1, 6)
		check(1, 2, 1, 1, 3, 1)
	})
}

func AxisSpec(c gospec.Context) {
	input := gin.Make()

	// TODO: This is the mouse x axis key, we need a constant for this or something
	x := input.GetKeyFlat(gin.MouseXAxis, gin.DeviceTypeMouse, 1)
	events := make([]gin.OsEvent, 0)

	c.Specify("Axes aggregate press amts and report IsDown() properly.", func() {
		injectEvent(&events, x.Id().Index, 1, gin.DeviceTypeMouse, 1, 5)
		injectEvent(&events, x.Id().Index, 1, gin.DeviceTypeMouse, 10, 6)
		injectEvent(&events, x.Id().Index, 1, gin.DeviceTypeMouse, -3, 7)
		input.Think(10, false, events)
		c.Expect(x.FramePressAmt(), Equals, -3.0)
		c.Expect(x.FramePressSum(), Equals, 8.0)
	})

	c.Specify("Axes can sum to zero and still be down.", func() {
		input.Think(0, false, events)
		events = events[0:0]
		c.Expect(x.FramePressSum(), Equals, 0.0)
		c.Expect(x.IsDown(), Equals, false)

		injectEvent(&events, x.Id().Index, 1, gin.DeviceTypeMouse, 5, 5)
		injectEvent(&events, x.Id().Index, 1, gin.DeviceTypeMouse, -5, 6)
		input.Think(10, false, events)
		events = events[0:0]
		c.Expect(x.FramePressSum(), Equals, 0.0)
		c.Expect(x.IsDown(), Equals, true)

		input.Think(20, false, events)
		c.Expect(x.FramePressSum(), Equals, 0.0)
		c.Expect(x.IsDown(), Equals, false)
	})
}

type listener struct {
	input   *gin.Input
	key_id  gin.KeyId
	context gospec.Context

	press_count   []int
	release_count []int
	press_amt     []float64
}

func (l *listener) ExpectPressCounts(v ...int) {
	l.press_count = v
}
func (l *listener) ExpectReleaseCounts(v ...int) {
	l.release_count = v
}
func (l *listener) ExpectPressAmts(v ...float64) {
	l.press_amt = v
}
func (l *listener) HandleEventGroup(eg gin.EventGroup) {
	k := l.input.GetKey(l.key_id)
	l.context.Expect(k.CurPressCount(), Equals, l.press_count[0])
	l.context.Expect(k.CurReleaseCount(), Equals, l.release_count[0])
	l.context.Expect(k.CurPressAmt(), Equals, l.press_amt[0])
	l.press_count = l.press_count[1:]
	l.release_count = l.release_count[1:]
	l.press_amt = l.press_amt[1:]
}
func (l *listener) Think(ms int64) {
	l.context.Expect(len(l.press_count), Equals, 0)
	l.context.Expect(len(l.release_count), Equals, 0)
	l.context.Expect(len(l.press_amt), Equals, 0)
}

func EventListenerSpec(c gospec.Context) {
	input := gin.Make()
	keya := input.GetKeyFlat(gin.KeyA, gin.DeviceTypeKeyboard, 1)
	keyb := input.GetKeyFlat(gin.KeyB, gin.DeviceTypeKeyboard, 1)
	AB_binding := input.MakeBinding(keya.Id(), []gin.KeyId{keyb.Id()}, []bool{true})
	AB := input.BindDerivedKey("AB", AB_binding)
	events := make([]gin.OsEvent, 0)

	c.Specify("Check keys report state properly while handling events", func() {
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 1)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 0, 2)
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 1, 3)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 1, 4)
		injectEvent(&events, 'b', 1, gin.DeviceTypeKeyboard, 0, 5)
		injectEvent(&events, 'a', 1, gin.DeviceTypeKeyboard, 0, 6)

		c.Specify("Test a", func() {
			la := &listener{
				input:   input,
				key_id:  keya.Id(),
				context: c,
			}
			input.RegisterEventListener(la)
			la.ExpectPressCounts(1, 1, 1, 2, 2, 2)
			la.ExpectReleaseCounts(0, 1, 1, 1, 1, 2)
			la.ExpectPressAmts(1, 0, 0, 1, 1, 0)
			input.Think(0, false, events)
		})
		c.Specify("Test b", func() {
			lb := &listener{
				input:   input,
				key_id:  keyb.Id(),
				context: c,
			}
			input.RegisterEventListener(lb)
			lb.ExpectPressCounts(0, 0, 1, 1, 1, 1)
			lb.ExpectReleaseCounts(0, 0, 0, 0, 1, 1)
			lb.ExpectPressAmts(0, 0, 1, 1, 0, 0)
			input.Think(0, false, events)
		})
		c.Specify("Test ab", func() {
			lab := &listener{
				input:   input,
				key_id:  AB.Id(),
				context: c,
			}
			input.RegisterEventListener(lab)
			lab.ExpectPressCounts(0, 0, 0, 1, 1, 1)
			lab.ExpectReleaseCounts(0, 0, 0, 0, 0, 1)
			lab.ExpectPressAmts(0, 0, 0, 1, 1, 0)
			input.Think(0, false, events)
		})
	})
}
