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

func (dk *derivedKey) SetPressAmt(amt float64, ms int64, cause *Event) (event *Event) {
  can_press := false
  if cause != nil && cause.Type == Press {
    for _,binding := range dk.Bindings {
      if cause.Key.Id() == binding.PrimaryKey {
        can_press = true
      }
    }
  }
  can_release := dk.is_down
  if (dk.keyState.this.press_amt == 0) == (amt == 0) {
    event = nil
  } else {
    event = &Event {
      Key : &dk.keyState,
      Timestamp : ms,
    }
    if amt == 0 && can_release {
      event.Type = Release
      dk.keyState.this.release_count++
      dk.is_down = false
    } else if amt != 0 && can_press {
      event.Type = Press
      dk.keyState.this.press_count++
      dk.is_down = true
    } else {
      event = nil
    }
  }
  dk.keyState.this.press_sum += dk.keyState.this.press_amt * float64(ms - dk.keyState.last_press)
  dk.keyState.this.press_amt = amt
  dk.keyState.last_press = ms
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
/*
  fmt.Printf("Binding Primary: %v\n", b.PrimaryKey)
  pk := b.Input.GetKey(b.PrimaryKey)
  fmt.Printf("%v: %f\n", pk, pk.CurPressAmt())
  for i := range b.Modifiers {
    k := b.Input.GetKey(b.Modifiers[i])
    fmt.Printf("Modifier: %v, %f %t\n", k, k.CurPressAmt(), b.Down[i])
  }
*/
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

