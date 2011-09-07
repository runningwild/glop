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


func (input *Input) BindDerivedKey(name string, bindings []binding) KeyId {
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
  return dk.id
}

// A derivedKey is down if any of its bindings are down
type derivedKey struct {
  keyState
  Bindings []binding
}

func (dk *derivedKey) CurPressAmt() float64 {
  sum := 0.0
  for _,binding := range dk.Bindings {
    sum += binding.CurPressAmt()
  }
  return sum
}

// A Binding is considered down if PrimaryKey is down and all Modifiers' IsDown()s match the
// corresponding entry in Down
type binding struct {
  PrimaryKey KeyId
  Modifiers  []KeyId
  Down       []bool
  input      *Input
}

func (b *binding) CurPressAmt() float64 {
  for i := range b.Modifiers {
    if (b.input.key_map[b.Modifiers[i]].CurPressAmt() != 0) != b.Down[i] {
      return 0
    }
  }
  return b.input.key_map[b.PrimaryKey].CurPressAmt()
}

