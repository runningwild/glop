package gin

import (
  "fmt"
  "glop/gos"
)

var (
  all_keys []Key
  key_map map[KeyId]Key

  dep_map map[KeyId][]Key
  // map from keyId to list of (derived) Keys that depend on it in some way
)

type EventType int
const (
  Press   EventType = iota
  Release
)

// TODO: Consider making a Timestamp type (int64)
type Event struct {
  Key       Key
  Type      EventType
  Timestamp int64
}

func init() {
  all_keys = make([]Key, 16)[0:0]
  key_map = make(map[KeyId]Key, 128)
  dep_map = make(map[KeyId][]Key, 16)
}

func registerKey(key Key, id KeyId) {
  if prev,ok := key_map[id]; ok {
    panic(fmt.Sprintf("Cannot register key '%v' with id %d, '%v' is already registered with that id.", key, id, prev))
  }
  key_map[id] = key
  all_keys = append(all_keys, key)
}

const (
  // Key repeat-rate constants
  repeatDelay = 500  // Ms between key down event #1 and #2 while a key is down
  repeatRate  = 60   // Ms between later key down events while a key is down
)


func pressKey(k Key, amt float64, t int64, events []*Event) {
  event := k.SetPressAmt(amt, t)
  deps,ok := dep_map[k.Id()]
  if !ok {
    if event != nil {
      events = append(events, event)
    }
  }
  length := len(events)
  for _,dep := range deps {
    pressKey(dep, dep.CurPressAmt(), t, events)
  }
  if len(events) == length {
    events = append(events, event)
  }
}

func Think(t int64, lost_focus bool) {
  os_events := gos.GetInputEvents()
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
    pressKey(key_map[KeyId(os_event.Index)], os_event.Press_amt, os_event.Timestamp, events)
  }

  for _,key := range all_keys {
    key.Think(t)
  }
}
