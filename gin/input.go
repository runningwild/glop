package gin

import (
  "fmt"
)

type Mouse struct {
  X,Y float64
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

  // map from keyId to list of (derived) Keys that depend on it in some way
  dep_map map[KeyId][]Key

  // The listeners will receive all events immediately after those events have been used to
  // update all key states.  The order in which listeners are notified of a particular event
  // group can change from group to group.
  listeners []Listener
}

func Make() *Input {
  input := new(Input)
  input.all_keys = make([]Key, 0, 16)
  input.key_map = make(map[KeyId]Key, 128)
  input.dep_map = make(map[KeyId][]Key, 16)

  for c := 'a'; c <= 'z'; c++ {
    input.registerNaturalKey(KeyId(c), fmt.Sprintf("%c", c))
  }
  for c := '0'; c <= '9'; c++ {
    input.registerNaturalKey(KeyId(c), fmt.Sprintf("%c", c))
  }
  ascii_keys := "`[]\\-=;',./"
  for _,key := range ascii_keys {
    input.registerNaturalKey(KeyId(key), fmt.Sprintf("%c", key))
  }
  input.registerNaturalKey(32, "Space")
  input.registerNaturalKey(8, "Backspace")
  input.registerNaturalKey(9, "Tab")
  input.registerNaturalKey(13, "Return")
  input.registerNaturalKey(27, "Escape")
  input.registerNaturalKey(129, "F1")
  input.registerNaturalKey(130, "F2")
  input.registerNaturalKey(131, "F3")
  input.registerNaturalKey(132, "F4")
  input.registerNaturalKey(133, "F5")
  input.registerNaturalKey(134, "F6")
  input.registerNaturalKey(135, "F7")
  input.registerNaturalKey(136, "F8")
  input.registerNaturalKey(137, "F9")
  input.registerNaturalKey(138, "F10")
  input.registerNaturalKey(139, "F11")
  input.registerNaturalKey(140, "F12")
  input.registerNaturalKey(150, "CapsLock")
  input.registerNaturalKey(151, "NumLock")
  input.registerNaturalKey(152, "ScrollLock")
  input.registerNaturalKey(153, "PrintScreen")
  input.registerNaturalKey(154, "Pause")
  input.registerNaturalKey(155, "LeftShift")
  input.registerNaturalKey(156, "RightShift")
  input.registerNaturalKey(157, "LeftControl")
  input.registerNaturalKey(158, "RightControl")
  input.registerNaturalKey(159, "LeftAlt")
  input.registerNaturalKey(160, "RightAlt")
  input.registerNaturalKey(161, "LeftGui")
  input.registerNaturalKey(162, "RightGui")
  input.registerNaturalKey(166, "Right")
  input.registerNaturalKey(167, "Left")
  input.registerNaturalKey(168, "Up")
  input.registerNaturalKey(169, "Down")
  input.registerNaturalKey(170, "KeyPadDivide")
  input.registerNaturalKey(171, "KeyPadMultiply")
  input.registerNaturalKey(172, "KeyPadSubtract")
  input.registerNaturalKey(173, "KeyPadAdd")
  input.registerNaturalKey(174, "KeyPadEnter")
  input.registerNaturalKey(175, "KeyPadDecimal")
  input.registerNaturalKey(176, "KeyPadEquals")
  input.registerNaturalKey(177, "KeyPad0")
  input.registerNaturalKey(178, "KeyPad1")
  input.registerNaturalKey(179, "KeyPad2")
  input.registerNaturalKey(180, "KeyPad3")
  input.registerNaturalKey(181, "KeyPad4")
  input.registerNaturalKey(182, "KeyPad5")
  input.registerNaturalKey(183, "KeyPad6")
  input.registerNaturalKey(184, "KeyPad7")
  input.registerNaturalKey(185, "KeyPad8")
  input.registerNaturalKey(186, "KeyPad9")
  input.registerNaturalKey(190, "KeyDelete")
  input.registerNaturalKey(191, "KeyHome")
  input.registerNaturalKey(192, "KeyInsert")
  input.registerNaturalKey(193, "KeyEnd")
  input.registerNaturalKey(194, "KeyPageUp")
  input.registerNaturalKey(195, "KeyPageDown")
  input.registerAxisKey(300, "MouseXAxis")
  input.registerAxisKey(301, "MouseYAxis")
  input.registerNaturalKey(302, "MouseWheelUp")
  input.registerNaturalKey(303, "MouseWheelDown")
  input.registerNaturalKey(304, "MouseLButton")
  input.registerNaturalKey(305, "MouseRButton")
  input.registerNaturalKey(306, "MouseMButton")

  return input
}

type EventType int
const (
  Press   EventType = iota
  Release
  Adjust   // The key was and is down, but the value of it has changed
  NoEvent
)
func (event EventType) String() string {
  switch event {
    case Press:
      return "press"
    case Release:
      return "release"
    case NoEvent:
      return "noevent"
    case Adjust:
      return "adjust"
  }
  panic(fmt.Sprintf("%d is not a valid EventType", event))
  return ""
}
// TODO: Consider making a Timestamp type (int64)
type Event struct {
  Key   Key
  Type  EventType
  Mouse Mouse
}
func (e Event) String() string {
  if e.Key == nil || e.Type == NoEvent {
    return fmt.Sprintf("NoEvent")
  }
  if e.Key == nil {
    return fmt.Sprintf("'%v %v'", e.Type, nil)
  }
  return fmt.Sprintf("'%v %v'", e.Type, e.Key)
}

// An EventGroup is a series of events that were all created by a single OsEvent.
type EventGroup struct {
  Events    []Event
  Timestamp int64
}
// Returns a bool indicating whether an event corresponding to the given KeyId is present
// in the EventGroup, and if so the Event returned is a copy of that event.
func (eg *EventGroup) FindEvent(id KeyId) (bool, Event) {
  for i := range eg.Events {
    if eg.Events[i].Key.Id() == id {
      return true, eg.Events[i]
    }
  }
  return false, Event{}
}

func (input *Input) registerKey(key Key, id KeyId) {
  if id <= 0 {
    panic(fmt.Sprintf("Cannot register a key with id %d, ids must be greater than 0.", id))
  }
  if prev,ok := input.key_map[id]; ok {
    panic(fmt.Sprintf("Cannot register key '%v' with id %d, '%v' is already registered with that id.", key, id, prev))
  }
  input.key_map[id] = key
  input.all_keys = append(input.all_keys, key)
}

func (input *Input) registerNaturalKey(id KeyId, name string) {
  input.registerKey(&keyState{id : id, name : name, aggregator : &standardAggregator{}}, id)
}

func (input *Input) registerAxisKey(id KeyId, name string) {
  input.registerKey(&keyState{id : id, name : name, aggregator : &axisAggregator{}}, id)
}

func (input *Input) GetKey(id KeyId) Key {
  key,ok := input.key_map[id]
  if !ok {
    panic(fmt.Sprintf("No key registered with id == %d.", id))
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

// The Input object can have a single Listener registered with it.  This object will receive
// event groups as they are processed.  During HandleEventGroup a listener can query keys as to
// their current state (i.e. with Cur*() methods) and these will accurately report their state
// given that the current event group has happened and no future events have happened yet.
// Frame*() methods on keys will report state from last frame.
// Listener.Think() will be called after all the events for a frame have been processed.
type Listener interface {
  HandleEventGroup(EventGroup)
  Think(int64)
}

func (input *Input) RegisterEventListener(listener Listener) {
  input.listeners = append(input.listeners, listener)
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
  var groups []EventGroup
  for _,os_event := range os_events {
    group := EventGroup{
      Timestamp : os_event.Timestamp,
    }
    input.pressKey(
        input.GetKey(os_event.KeyId),
        os_event.Press_amt,
        os_event.Timestamp,
        Event{},
        &group)

    // Include the position of the cursor on all events.  This is mostly for the sake
    // of mouse click events but it could be useful otherwise and it makes all Events
    // the same in terms of structure.
    for i := range group.Events {
      group.Events[i].Mouse = os_event.Mouse
    }

    if len(group.Events) > 0 {
      groups = append(groups, group)
    }
    for _,listener := range input.listeners {
      listener.HandleEventGroup(group)
    }
  }

  for _,key := range input.all_keys {
    key.Think(t)
  }

  for _,listener := range input.listeners {
    listener.Think(t)
  }
  return groups
}
