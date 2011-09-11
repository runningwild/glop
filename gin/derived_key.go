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
  }
  input.registerKey(dk, dk.id)

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
  Bindings []Binding
  is_down  bool
}

func (dk *derivedKey) CurPressAmt() float64 {
  sum := 0.0
  for _,binding := range dk.Bindings {
    sum += binding.CurPressAmt()
  }
  return sum
}

func (dk *derivedKey) IsDown() bool {
  return dk.is_down
}

func (dk *derivedKey) SetPressAmt(amt float64, ms int64, cause Event) (event Event) {
  is_primary := false
  if cause.Type != NoEvent {
    for _,binding := range dk.Bindings {
      if cause.Key.Id() == binding.PrimaryKey {
        is_primary = true
      }
    }
  }
  can_press := is_primary && cause.Type == Press
  can_release := is_primary && dk.is_down

  event.Type = NoEvent
  if (dk.keyState.CurPressAmt() == 0) != (amt == 0) {
    event.Key = &dk.keyState
    if amt == 0 && can_release {
      event.Type = Release
      dk.is_down = false
    } else if amt != 0 && can_press {
      event.Type = Press
      dk.is_down = true
    } else {
      event.Type = NoEvent
    }
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

