package gin

// func (input *Input) registerDependence(derived Key, dep KeyId) {
// 	list, ok := input.id_to_deps[dep]
// 	if !ok {
// 		list = make([]Key, 0)
// 	}
// 	list = append(list, derived)
// 	input.id_to_deps[dep] = list
// }

// func (input *Input) BindDerivedKey(name string, bindings ...Binding) Key {
// 	return input.bindDerivedKeyWithIndex(name, genDerivedKeyIndex(), bindings...)
// }

// func (input *Input) bindDerivedKeyWithIndex(name string, index KeyIndex, bindings ...Binding) Key {
// 	dk := &derivedKey{
// 		keyState: keyState{
// 			id: KeyId{
// 				Index: index,
// 				Device: DeviceId{
// 					Index: 1,
// 					Type:  DeviceTypeDerived,
// 				},
// 			},
// 			name:       name,
// 			aggregator: &standardAggregator{},
// 		},
// 		Bindings:      bindings,
// 		bindings_down: make([]bool, len(bindings)),
// 	}

// 	// TODO: Decide whether or not this is true, might need to register them for
// 	// when the game loses focus.
// 	// I think it might not be necessary to register derived keys.
// 	// input.registerKeyIndex(dk.id.Index, &standardAggregator{}, name)

// 	for _, binding := range bindings {
// 		input.registerDependence(dk, binding.PrimaryKey)
// 		for _, modifier := range binding.Modifiers {
// 			input.registerDependence(dk, modifier)
// 		}
// 	}
// 	input.key_map[dk.id] = dk
// 	input.all_keys = append(input.all_keys, dk)
// 	return dk
// }

// A generalDerivedKey represents a group of natural keys.  A key is specified
// with (key index, device type, device index).  Given these there variables,
// the following are possible:
// (specific, specific, specific) - These are natural keys
// (specific, specific, general) - Specific key on any device of a specific type
// (specific, general, general) - Specific key on any device at all
// (general, specific, specific) - Any key on a specific device
// (general, specific, general) - Any key on any device of a specific type
// (general, general, general) - Any key on any device at all
// Note: It never makes sense to specify a device index without specifying the
// device type - doing so will cause glop to panic.
type generalDerivedKey struct {
	keyState
	press_amt float64

	// We need the input object itself so that we can get at all of the keys that
	// we depend on.
	input *Input
}

func (gdk *generalDerivedKey) CurPressAmt() float64 {
	sum := 0.0
	for _, key := range gdk.input.all_keys {
		if key.Id().Index == AnyKey ||
			key.Id().Device.Type == DeviceTypeAny ||
			key.Id().Device.Type == DeviceTypeDerived ||
			key.Id().Device.Index == DeviceIndexAny {
			continue
		}
		if gdk.Id().Index != AnyKey && key.Id().Index != gdk.Id().Index {
			// Not the appropriate key index
			continue
		}
		if gdk.Id().Device.Type != DeviceTypeAny &&
			key.Id().Device.Type != gdk.Id().Device.Type {
			// Not the appropriate device type
			continue
		}
		if gdk.Id().Device.Index != DeviceIndexAny &&
			key.Id().Device.Index != gdk.Id().Device.Index {
			// Not the appropriate device index
			continue
		}
		sum += key.CurPressAmt()
	}
	return sum
}

func (gdk *generalDerivedKey) IsDown() bool {
	return gdk.press_amt > 0
}

func (gdk *generalDerivedKey) SetPressAmt(amt float64, ms int64, cause Event) (event Event) {
	event.Type = NoEvent
	event.Key = &gdk.keyState
	old_press_amt := gdk.press_amt
	gdk.press_amt = gdk.CurPressAmt()
	if (old_press_amt == 0) == (gdk.press_amt == 0) {
		return
	}
	if gdk.press_amt > 0 {
		event.Type = Press
	} else {
		event.Type = Release
	}
	gdk.keyState.aggregator.SetPressAmt(gdk.press_amt, ms, event.Type)
	return
}
