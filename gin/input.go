package gin

import (
  "fmt"
  "glop/system"
)

// Everything 'global' is put inside a struct so that tests can be run without stepping
// on each other
type Input struct {
  sys system.System
  // Used for handling all os-level stuff, can be replaced with a mock System object for testing

  all_keys []Key
  key_map map[KeyId]Key

  dep_map map[KeyId][]Key
  // map from keyId to list of (derived) Keys that depend on it in some way
}

func MakeInput(sys system.System) *Input {
  input := new(Input)
  input.sys = sys
  input.all_keys = make([]Key, 16)[0:0]
  input.key_map = make(map[KeyId]Key, 128)
  input.dep_map = make(map[KeyId][]Key, 16)

  for c := 'a'; c <= 'z'; c++ {
    input.registerNaturalKey(KeyId(c), fmt.Sprintf("%c", c))
  }
  return input
}

type EventType int
const (
  Press   EventType = iota
  Release
)
func (event EventType) String() string {
  switch event {
    case Press:
      return "press"
    case Release:
      return "release"
  }
  panic(fmt.Sprintf("%d is not a valid EventType", event))
  return ""
}
// TODO: Consider making a Timestamp type (int64)
type Event struct {
  Key       Key
  Type      EventType
  Timestamp int64
}

func init() {
}

func (input *Input) registerKey(key Key, id KeyId) {
  if prev,ok := input.key_map[id]; ok {
    panic(fmt.Sprintf("Cannot register key '%v' with id %d, '%v' is already registered with that id.", key, id, prev))
  }
  input.key_map[id] = key
  input.all_keys = append(input.all_keys, key)
}

func (input *Input) registerNaturalKey(id KeyId, name string) {
  input.registerKey(&keyState{id : id, name : name}, id)
}

func (input *Input) GetKey(id KeyId) Key {
  key,ok := input.key_map[id]
  if !ok {
    return nil
  }
  return key
}

func (input *Input) pressKey(k Key, amt float64, t int64, events []*Event, cause *Event) {
  event := k.SetPressAmt(amt, t, cause)
  deps,ok := input.dep_map[k.Id()]
  if !ok {
    if event != nil {
      events = append(events, event)
    }
  }

  length := len(events)
  for _,dep := range deps {
    input.pressKey(dep, dep.CurPressAmt(), t, events, event)
  }
  if len(events) == length {
    events = append(events, event)
  }
}

func (input *Input) Think(t int64, lost_focus bool) {
  os_events := input.sys.GetInputEvents()
  if len(os_events) == 0 {
    panic("Expected at least one event from a call to gos.GetInputEvents()")
  }

  // If we have lost focus, clear all key state. Note that down_keys_frame_ is rebuilt every frame
  // regardless, so we do not need to worry about it here.
  if lost_focus {
//    clearAllKeyState()
  }

  // Generate all key events here.  Derived keys are handled through pressKey and all
  // events are aggregated into one array.  Events in this array will necessarily be in
  // sorted order.
  events := make([]*Event, 10)[0:0]
  for _,os_event := range os_events {
    input.pressKey(
        input.key_map[KeyId(os_event.Index)],
        os_event.Press_amt,
        os_event.Timestamp,
        events,
        nil)
  }

  for _,key := range input.all_keys {
    key.Think(t)
  }
}
