package gin

// TODO: Old comment, make sure it's valid
// A derived key generates events whenever any of its main keys generate events (in the same call
// to OnKeyEvents). It's press amount is the sum of the press amounts of its active main keys.

var (
  next_derived_key_id int
)

func init() {
  next_derived_key_id = 10000
}

func getDerivedKeyId() (id int) {
  id = next_derived_key_id
  next_derived_key_id++
  return
}

func BindDerivedKey(name string, bindings []binding) KeyId {
  dk := derivedKey {
    KeyId : KeyId{
      id : getDerivedKeyId(),
      name : name,
    },
    Bindings : bindings,
  }
  derived_keys = append(derived_keys, dk)
  all_keys = append(all_keys, dk)
  return dk.KeyId
}

// A derivedKey is down if any of its bindings is down
type derivedKey struct {
  KeyId
  KeyState
  Bindings []binding
}

// A Binding is considered down if PrimaryKey is down and all Modifiers' IsDown()s match the
// corresponding entry in Down
type binding struct {
  PrimaryKey KeyId
  Modifiers  []KeyId
  Down       []bool
}

func (b *binding) IsDown() bool {
  return true
}
