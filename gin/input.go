package gin

import (
  "fmt"
)

type Mouse struct {
  X,Y,Dx,Dy int
}

type OsEvent struct {
  // TODO: rename index to KeyId or something more appropriate
  KeyId     KeyId
  Press_amt float64
  Mouse     Mouse
  Timestamp int64
  Num_lock  int
  Caps_lock int
}

// Everything 'global' is put inside a struct so that tests can be run without stepping
// on each other
type Input struct {
  all_keys []Key
  key_map map[KeyId]Key

  dep_map map[KeyId][]Key
  // map from keyId to list of (derived) Keys that depend on it in some way
}

func Make() *Input {
  input := new(Input)
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
  NoEvent
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
}
func (e Event) String() string {
  if e.Key == nil || e.Type == NoEvent {
    return fmt.Sprintf("NoEvent")
  }
  return fmt.Sprintf("'%v %v'", e.Type, e.Key)
}

// An EventGroup is a series of events that were all created by a single OsEvent.
type EventGroup struct {
  Events    []Event
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

func (input *Input) pressKey(k Key, amt float64, t int64, cause Event, group *EventGroup) {
  event := k.SetPressAmt(amt, t, cause)
  deps := input.dep_map[k.Id()]

  for _,dep := range deps {
    input.pressKey(dep, dep.CurPressAmt(), t, event, group)
  }
  if event.Type != NoEvent {
    group.Events = append(group.Events, event)
  }
}

func (input *Input) Think(t int64, lost_focus bool, os_events []OsEvent) ([]EventGroup) {
  // If we have lost focus, clear all key state. Note that down_keys_frame_ is rebuilt every frame
  // regardless, so we do not need to worry about it here.
  if lost_focus {
//    clearAllKeyState()
  }

  // Generate all key events here.  Derived keys are handled through pressKey and all
  // events are aggregated into one array.  Events in this array will necessarily be in
  // sorted order.
  groups := make([]EventGroup, 10)[0:0]
  for _,os_event := range os_events {
    group := EventGroup{
      Events : make([]Event, 1)[0:0],
      Timestamp : os_event.Timestamp,
    }
    input.pressKey(
        input.key_map[os_event.KeyId],
        os_event.Press_amt,
        os_event.Timestamp,
        Event{},
        &group)
    groups = append(groups, group)
  }

  for _,key := range input.all_keys {
    key.Think(t)
  }

  return groups
}
