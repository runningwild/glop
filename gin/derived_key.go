package gin

var (
  next_derived_key_id KeyId
)
func init() {
  next_derived_key_id = KeyId(10000)
}

func genDerivedKeyId() (id KeyId) {
  id = next_derived_key_id
  next_derived_key_id++
  return
}

// TODO: Handle removal of dependencies
func (input *Input) registerDependence(derived Key, dep KeyId) {
  list,ok := input.dep_map[dep]
  if !ok {
    list = make([]Key, 0)
  }
  list = append(list, derived)
  input.dep_map[dep] = list
}


func (input *Input) BindDerivedKey(name string, bindings ...Binding) Key {
  dk := &derivedKey {
    keyState : keyState {
      id : genDerivedKeyId(),
      name : name,
      aggregator : &standardAggregator{},
    },
    Bindings : bindings,
    bindings_down : make([]bool, len(bindings)),
  }

  // Currently we don't have a way to register derived keys with cursor
  // association, but if one of the bindings includes a key with such an
  // association any event handler will be able to get at this data.
  input.registerKey(dk, dk.id, nil)

  for _,binding := range bindings {
    input.registerDependence(dk, binding.PrimaryKey)
    for _,modifier := range binding.Modifiers {
      input.registerDependence(dk, modifier)
    }
  }
  return dk
}

func (input *Input) MakeBinding(primary KeyId, modifiers []KeyId, down []bool) Binding {
  return Binding{
    PrimaryKey : primary,
    Modifiers  : modifiers,
    Down       : down,
    Input      : input,
  }
}

// A derivedKey is down if any of its bindings are down
type derivedKey struct {
  keyState
  is_down  bool

  Bindings      []Binding
  bindings_down []bool
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
  for _,down := range dk.bindings_down {
    if down {
      count++
    }
  }
  return count
}

func (dk *derivedKey) SetPressAmt(amt float64, ms int64, cause Event) (event Event) {
  index := -1
  if cause.Type != NoEvent {
    for i,binding := range dk.Bindings {
      if cause.Key.Id() == binding.PrimaryKey {
        index = i
      }
    }
  }
  event.Type = NoEvent
  event.Key = &dk.keyState
  if amt == 0 && index != -1 && dk.numBindingsDown() == 1 && dk.bindings_down[index] {
    event.Type = Release
  }
  if amt != 0 && index != -1 && dk.numBindingsDown() == 0 && !dk.bindings_down[index] {
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

func (b *Binding) primaryPressAmt() float64 {
  return b.Input.key_map[b.PrimaryKey].CurPressAmt()
}

func (b *Binding) CurPressAmt() float64 {
  for i := range b.Modifiers {
    if b.Input.key_map[b.Modifiers[i]].IsDown() != b.Down[i] {
      return 0
    }
  }
  if !b.Input.key_map[b.PrimaryKey].IsDown() {
    return 0
  }
  return b.Input.key_map[b.PrimaryKey].CurPressAmt()
}

