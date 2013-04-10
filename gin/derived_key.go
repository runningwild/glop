package gin

import (
	"fmt"
)

var (
	next_derived_key_index KeyIndex
)

func init() {
	next_derived_key_index = 10000
}

func genDerivedKeyIndex() KeyIndex {
	next_derived_key_index++
	return next_derived_key_index
}

// TODO: Handle removal of dependencies
func (input *Input) registerDependence(derived Key, dep KeyId) {
	list, ok := input.id_to_deps[dep]
	if !ok {
		list = make([]Key, 0)
	}
	list = append(list, derived)
	input.id_to_deps[dep] = list
	if derived.Id() == dep {
		panic("Can't have a key depend on itself.")
	}
}

func (input *Input) BindDerivedKey(name string, bindings ...Binding) Key {
	return input.bindDerivedKeyWithIndex(
		name,
		genDerivedKeyIndex(),
		DeviceId{Index: 1, Type: DeviceTypeDerived},
		bindings...)
}

func (input *Input) bindDerivedKeyWithIndex(name string, index KeyIndex, device DeviceId, bindings ...Binding) Key {
	dk := &derivedKey{
		keyState: keyState{
			id: KeyId{
				Index:  index,
				Device: device,
			},
			name:       name,
			aggregator: &standardAggregator{},
		},
		Bindings:      bindings,
		bindings_down: make([]bool, len(bindings)),
	}

	// TODO: Decide whether or not this is true, might need to register them for
	// when the game loses focus.
	// I think it might not be necessary to register derived keys.
	// input.registerKeyIndex(dk.id.Index, &standardAggregator{}, name)

	for _, binding := range bindings {
		input.registerDependence(dk, binding.PrimaryKey)
		for _, modifier := range binding.Modifiers {
			input.registerDependence(dk, modifier)
		}
	}

	// TODO: Figure out a way to move this into Input.GetKey() or something.  It's
	// really dirty to have these maps/slices populated in multiple places.
	input.key_map[dk.id] = dk
	input.all_keys = append(input.all_keys, dk)
	return dk
}

// A derivedKey is down if any of its bindings are down
type derivedKey struct {
	keyState
	is_down bool

	Bindings []Binding

	// We store the down state of all of our bindings so that things 
	bindings_down []bool
}

func (dk *derivedKey) Think(ms int64) (bool, float64) {
	return dk.keyState.Think(ms)
}

func (dk *derivedKey) CurPressAmt() float64 {
	sum := 0.0
	for index := range dk.Bindings {
		if dk.bindings_down[index] {
			sum += dk.Bindings[index].primaryPressAmt()
		} else {
			sum += dk.Bindings[index].CurPressAmt()
		}
	}
	return sum
}

func (dk *derivedKey) IsDown() bool {
	return dk.numBindingsDown() > 0
}

func (dk *derivedKey) numBindingsDown() int {
	count := 0
	for _, down := range dk.bindings_down {
		if down {
			count++
		}
	}
	return count
}

func (dk *derivedKey) SetPressAmt(amt float64, ms int64, cause Event) (event Event) {
	fmt.Printf("dk.SetPressAmt()\n")
	index := -1
	for i, binding := range dk.Bindings {
		if cause.Key.Id() == binding.PrimaryKey {
			index = i
		}
	}
	event.Type = NoEvent
	event.Key = &dk.keyState
	if amt == 0 && index != -1 && dk.numBindingsDown() == 1 && dk.bindings_down[index] {
		event.Type = Release
	}
	if amt != 0 && index != -1 && dk.numBindingsDown() == 0 && !dk.bindings_down[index] {
		fmt.Printf("(%p) Generated press event for %v\n", dk, dk.Id())
		event.Type = Press
	}
	if index != -1 {
		dk.bindings_down[index] = dk.Bindings[index].CurPressAmt() != 0
	}
	dk.keyState.aggregator.SetPressAmt(amt, ms, event.Type)
	return
}

// A Binding is considered down if PrimaryKey is down and all Modifiers' IsDown()s match the
// corresponding entry in Down
type Binding struct {
	PrimaryKey KeyId
	Modifiers  []KeyId
	Down       []bool
	Input      *Input
}

func (input *Input) MakeBinding(primary KeyId, modifiers []KeyId, down []bool) Binding {
	if len(modifiers) != len(down) {
		panic("MakeBindings(primary, modifiers, down) - modifiers and down must have the same length.")
	}
	return Binding{
		PrimaryKey: primary,
		Modifiers:  modifiers,
		Down:       down,
		Input:      input,
	}
}

func (b *Binding) primaryPressAmt() float64 {
	return b.Input.key_map[b.PrimaryKey].CurPressAmt()
}

func (b *Binding) CurPressAmt() float64 {
	for i := range b.Modifiers {
		if b.Input.GetKey(b.Modifiers[i]).IsDown() != b.Down[i] {
			return 0
		}
	}
	if !b.Input.key_map[b.PrimaryKey].IsDown() {
		return 0
	}
	return b.Input.key_map[b.PrimaryKey].CurPressAmt()
}

type derivedKeyFamily struct {
	name             string
	index            KeyIndex
	binding_families []BindingFamily
	input            *Input
}

func (input *Input) BindDerivedKeyFamily(name string, bindings ...BindingFamily) KeyIndex {
	return input.bindDerivedKeyFamilyWithIndex(
		name,
		genDerivedKeyIndex(),
		bindings...)
}

func (input *Input) bindDerivedKeyFamilyWithIndex(name string, index KeyIndex, bindings ...BindingFamily) KeyIndex {
	dkf := derivedKeyFamily{
		name:             name,
		index:            index,
		binding_families: bindings,
		input:            input,
	}
	for _, binding := range bindings {
		input.index_to_family_deps[binding.PrimaryIndex] = append(input.index_to_family_deps[dkf.index], dkf)
		for _, mod := range binding.Modifiers {
			input.index_to_family_deps[mod] = append(input.index_to_family_deps[dkf.index], dkf)
		}
	}
	dkf.input.index_to_family[dkf.index] = dkf
	return dkf.index
}

func (dkf *derivedKeyFamily) GetKey(device DeviceId) Key {
	id := KeyId{Index: dkf.index, Device: device}
	if _, ok := dkf.input.key_map[id]; !ok {
		var bindings []Binding
		for _, binding_family := range dkf.binding_families {
			var modifiers []KeyId
			for _, index := range binding_family.Modifiers {
				modifiers = append(modifiers, KeyId{Index: index, Device: device})
			}
			bindings = append(bindings, Binding{
				PrimaryKey: KeyId{Index: binding_family.PrimaryIndex, Device: device},
				Modifiers:  modifiers,
				Down:       binding_family.Down,
				Input:      dkf.input,
			})
		}
		dkf.input.bindDerivedKeyWithIndex(dkf.name, dkf.index, device, bindings...)

		// On second thought i don't think i need any of this - it's done in the above
		// function call		
		// TODO: Maybe should set the key in the map here, it gets set elsewhere and
		// it might be nice to have that all in one place.
		// dkf.input.key_map[id] = key
		// dkf.input.all_keys = append(dkf.input.all_keys, key)
	}
	ret := dkf.input.GetKeyFlat(dkf.index, device.Type, device.Index)
	return ret
}

// A BindingFamily is like a binding, but it does not specify a device.  Instead
// a binding family can be down per device.  Example:
// bf := BindingFamily{
//   PrimaryKey: KeyA,
//   Modifiers: []KeyIndex{KeyB},
//   Down: []bool{false},
// }
type BindingFamily struct {
	PrimaryIndex KeyIndex
	Modifiers    []KeyIndex
	Down         []bool
}

func (input *Input) MakeBindingFamily(primary KeyIndex, modifiers []KeyIndex, down []bool) BindingFamily {
	if len(modifiers) != len(down) {
		panic("MakeBindingFamilys(primary, modifiers, down) - modifiers and down must have the same length.")
	}
	return BindingFamily{
		PrimaryIndex: primary,
		Modifiers:    modifiers,
		Down:         down,
	}
}
