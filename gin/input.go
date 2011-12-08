package gin

import (
  "fmt"
)

const (
  Space          = 32
  Backspace      = 8
  Tab            = 9
  Return         = 13
  Escape         = 27
  F1             = 129
  F2             = 130
  F3             = 131
  F4             = 132
  F5             = 133
  F6             = 134
  F7             = 135
  F8             = 136
  F9             = 137
  F10            = 138
  F11            = 139
  F12            = 140
  CapsLock       = 150
  NumLock        = 151
  ScrollLock     = 152
  PrintScreen    = 153
  Pause          = 154
  LeftShift      = 155
  RightShift     = 156
  LeftControl    = 157
  RightControl   = 158
  LeftAlt        = 159
  RightAlt       = 160
  LeftGui        = 161
  RightGui       = 162
  Right          = 166
  Left           = 167
  Up             = 168
  Down           = 169
  KeyPadDivide   = 170
  KeyPadMultiply = 171
  KeyPadSubtract = 172
  KeyPadAdd      = 173
  KeyPadEnter    = 174
  KeyPadDecimal  = 175
  KeyPadEquals   = 176
  KeyPad0        = 177
  KeyPad1        = 178
  KeyPad2        = 179
  KeyPad3        = 180
  KeyPad4        = 181
  KeyPad5        = 182
  KeyPad6        = 183
  KeyPad7        = 184
  KeyPad8        = 185
  KeyPad9        = 186
  KeyDelete      = 190
  KeyHome        = 191
  KeyInsert      = 192
  KeyEnd         = 193
  KeyPageUp      = 194
  KeyPageDown    = 195
  MouseXAxis     = 300
  MouseYAxis     = 301
  MouseWheelVertical   = 302
  MouseWheelHorizontal = 303
  MouseLButton   = 304
  MouseRButton   = 305
  MouseMButton   = 306

  // standard derived keys start here
  EitherShift = 1000 + iota
)

type Cursor interface {
  Name() string
  Point() (int, int)
}

type cursor struct {
  name string

  // Window coordinates of the cursor with the origin set as the lower-left
  // corner of the window.
  X, Y int
}

func (c *cursor) Name() string {
  return c.name
}
func (c *cursor) Point() (int, int) {
  return c.X, c.Y
}

type OsEvent struct {
  // TODO: rename index to KeyId or something more appropriate
  KeyId     KeyId
  Press_amt float64

  // For all cursor-based events this is the location of the cursor in window
  // coordinates when the event happened.  For mouse motion and mouse clicks this
  // is the location of the mouse.  For non-cursor-based events these values are
  // meaningless
  X, Y int

  Timestamp int64
  Num_lock  int
  Caps_lock int
}

// Everything 'global' is put inside a struct so that tests can be run without stepping
// on each other
type Input struct {
  all_keys []Key
  key_map  map[KeyId]Key

  // map from keyId to list of (derived) Keys that depend on it in some way
  dep_map map[KeyId][]Key

  // The listeners will receive all events immediately after those events have been used to
  // update all key states.  The order in which listeners are notified of a particular event
  // group can change from group to group.
  listeners []Listener

  // NOTE: Currently the only cursor supported is the mouse
  // Map from KeyId to the cursor associated with that key.  All KeyIds should be registered
  // in this map and will map to nil if they are not cursor keys.
  cursor_keys map[KeyId]*cursor

  cursors map[string]*cursor
}

// The standard input object
var input_obj *Input

func init() {
  input_obj = Make()
}

// TODO: You fucked up, the name of this function should be Input, and it should
//       return an interfact or something that is not called Input
func In() *Input {
  return input_obj
}

// Creates a new input object, mostly for testing.  Most users will just query
// gin.Input, which is created during initialization
func Make() *Input {
  input := new(Input)
  input.all_keys = make([]Key, 0, 512)
  input.key_map = make(map[KeyId]Key, 512)
  input.dep_map = make(map[KeyId][]Key, 16)
  input.cursor_keys = make(map[KeyId]*cursor, 512)
  input.cursors = make(map[string]*cursor, 2)

  for c := 'a'; c <= 'z'; c++ {
    input.registerNaturalKey(KeyId(c), fmt.Sprintf("%c", c))
  }
  for c := '0'; c <= '9'; c++ {
    input.registerNaturalKey(KeyId(c), fmt.Sprintf("%c", c))
  }
  ascii_keys := "`[]\\-=;',./"
  for _, key := range ascii_keys {
    input.registerNaturalKey(KeyId(key), fmt.Sprintf("%c", key))
  }
  input.registerNaturalKey(Space, "Space")
  input.registerNaturalKey(Backspace, "Backspace")
  input.registerNaturalKey(Tab, "Tab")
  input.registerNaturalKey(Return, "Return")
  input.registerNaturalKey(Escape, "Escape")
  input.registerNaturalKey(F1, "F1")
  input.registerNaturalKey(F2, "F2")
  input.registerNaturalKey(F3, "F3")
  input.registerNaturalKey(F4, "F4")
  input.registerNaturalKey(F5, "F5")
  input.registerNaturalKey(F6, "F6")
  input.registerNaturalKey(F7, "F7")
  input.registerNaturalKey(F8, "F8")
  input.registerNaturalKey(F9, "F9")
  input.registerNaturalKey(F10, "F10")
  input.registerNaturalKey(F11, "F11")
  input.registerNaturalKey(F12, "F12")
  input.registerNaturalKey(CapsLock, "CapsLock")
  input.registerNaturalKey(NumLock, "NumLock")
  input.registerNaturalKey(ScrollLock, "ScrollLock")
  input.registerNaturalKey(PrintScreen, "PrintScreen")
  input.registerNaturalKey(Pause, "Pause")
  input.registerNaturalKey(LeftShift, "LeftShift")
  input.registerNaturalKey(RightShift, "RightShift")
  input.registerNaturalKey(LeftControl, "LeftControl")
  input.registerNaturalKey(RightControl, "RightControl")
  input.registerNaturalKey(LeftAlt, "LeftAlt")
  input.registerNaturalKey(RightAlt, "RightAlt")
  input.registerNaturalKey(LeftGui, "LeftGui")
  input.registerNaturalKey(RightGui, "RightGui")
  input.registerNaturalKey(Right, "Right")
  input.registerNaturalKey(Left, "Left")
  input.registerNaturalKey(Up, "Up")
  input.registerNaturalKey(Down, "Down")
  input.registerNaturalKey(KeyPadDivide, "KeyPadDivide")
  input.registerNaturalKey(KeyPadMultiply, "KeyPadMultiply")
  input.registerNaturalKey(KeyPadSubtract, "KeyPadSubtract")
  input.registerNaturalKey(KeyPadAdd, "KeyPadAdd")
  input.registerNaturalKey(KeyPadEnter, "KeyPadEnter")
  input.registerNaturalKey(KeyPadDecimal, "KeyPadDecimal")
  input.registerNaturalKey(KeyPadEquals, "KeyPadEquals")
  input.registerNaturalKey(KeyPad0, "KeyPad0")
  input.registerNaturalKey(KeyPad1, "KeyPad1")
  input.registerNaturalKey(KeyPad2, "KeyPad2")
  input.registerNaturalKey(KeyPad3, "KeyPad3")
  input.registerNaturalKey(KeyPad4, "KeyPad4")
  input.registerNaturalKey(KeyPad5, "KeyPad5")
  input.registerNaturalKey(KeyPad6, "KeyPad6")
  input.registerNaturalKey(KeyPad7, "KeyPad7")
  input.registerNaturalKey(KeyPad8, "KeyPad8")
  input.registerNaturalKey(KeyPad9, "KeyPad9")
  input.registerNaturalKey(KeyDelete, "KeyDelete")
  input.registerNaturalKey(KeyHome, "KeyHome")
  input.registerNaturalKey(KeyInsert, "KeyInsert")
  input.registerNaturalKey(KeyEnd, "KeyEnd")
  input.registerNaturalKey(KeyPageUp, "KeyPageUp")
  input.registerNaturalKey(KeyPageDown, "KeyPageDown")

  input.registerCursor("Mouse")
  input.registerCursorAxisKey(MouseXAxis, "MouseXAxis", "Mouse")
  input.registerCursorAxisKey(MouseYAxis, "MouseYAxis", "Mouse")
  input.registerCursorWheelKey(MouseWheelVertical, "MouseWheelVertical", "Mouse")
  input.registerCursorKey(MouseLButton, "MouseLButton", "Mouse")
  input.registerCursorKey(MouseRButton, "MouseRButton", "Mouse")
  input.registerCursorKey(MouseMButton, "MouseMButton", "Mouse")

  input.bindDerivedKeyWithId("EitherShift", EitherShift, input.MakeBinding(LeftShift, nil, nil), input.MakeBinding(RightShift, nil, nil))
  return input
}

type EventType int

const (
  NoEvent EventType = iota
  Press
  Release
  Adjust // The key was and is down, but the value of it has changed
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
  Key  Key
  Type EventType
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

func (input *Input) registerCursor(name string) {
  input.cursors[name] = &cursor{name: name}
}
func (input *Input) registerKey(key Key, id KeyId, cursor_name string) {
  if id <= 0 {
    panic(fmt.Sprintf("Cannot register a key with id %d, ids must be greater than 0.", id))
  }
  if prev, ok := input.key_map[id]; ok {
    panic(fmt.Sprintf("Cannot register key '%v' with id %d, '%v' is already registered with that id.", key, id, prev))
  }
  input.key_map[id] = key
  if cursor_name != "" {
    input.cursor_keys[id] = input.cursors[cursor_name]
  } else {
    input.cursor_keys[id] = nil
  }
  input.all_keys = append(input.all_keys, key)
}

func (input *Input) registerNaturalKey(id KeyId, name string) {
  input.registerKey(&keyState{id: id, name: name, aggregator: &standardAggregator{}}, id, "")
}

func (input *Input) registerCursorKey(id KeyId, name string, cursor_name string) {
  input.registerKey(&keyState{id: id, name: name, aggregator: &standardAggregator{}, cursor: input.cursors[cursor_name]}, id, cursor_name)
}

func (input *Input) registerAxisKey(id KeyId, name string) {
  input.registerKey(&keyState{id: id, name: name, aggregator: &axisAggregator{}}, id, "")
}

func (input *Input) registerCursorAxisKey(id KeyId, name string, cursor_name string) {
  input.registerKey(&keyState{id: id, name: name, aggregator: &axisAggregator{}, cursor: input.cursors[cursor_name]}, id, cursor_name)
}

func (input *Input) registerCursorWheelKey(id KeyId, name string, cursor_name string) {
  input.registerKey(&keyState{id: id, name: name, aggregator: &wheelAggregator{}, cursor: input.cursors[cursor_name]}, id, cursor_name)
}

func (input *Input) GetCursor(name string) Cursor {
  cursor, ok := input.cursors[name]
  if !ok {
    panic(fmt.Sprintf("No cursor registered with name == '%s'.", name))
    return nil
  }
  return cursor
}
func (input *Input) GetKey(id KeyId) Key {
  key, ok := input.key_map[id]
  if !ok {
    panic(fmt.Sprintf("No key registered with id == %d.", id))
    return nil
  }
  return key
}

func (input *Input) informDeps(event Event, group *EventGroup) {
  deps := input.dep_map[event.Key.Id()]

  for _, dep := range deps {
    input.pressKey(dep, dep.CurPressAmt(), event, group)
  }
  if event.Type != NoEvent {
    group.Events = append(group.Events, event)
  }
}

func (input *Input) pressKey(k Key, amt float64, cause Event, group *EventGroup) {
  event := k.SetPressAmt(amt, group.Timestamp, cause)
  input.informDeps(event, group)
}

// The Input object can have a single Listener registered with it.  This object will receive
// event groups as they are processed.  During HandleEventGroup a listener can query keys as to
// their current state (i.e. with Cur*() methods) and these will accurately report their state
// given that the current event group has happened and no future events have happened yet.
// Frame*() methods on keys will report state from last frame.
// Listener.Think() will be called after all the events for a frame have been processed.
type EventHandler interface {
  HandleEventGroup(EventGroup)
}
type Listener interface {
  EventHandler
  Think(int64)
}
type EventDispatcher interface {
  RegisterEventListener(Listener)
}

func (input *Input) RegisterEventListener(listener Listener) {
  input.listeners = append(input.listeners, listener)
}

func (input *Input) Think(t int64, lost_focus bool, os_events []OsEvent) []EventGroup {
  // If we have lost focus, clear all key state. Note that down_keys_frame_ is rebuilt every frame
  // regardless, so we do not need to worry about it here.
  if lost_focus {
    //    clearAllKeyState()
  }
  // Generate all key events here.  Derived keys are handled through pressKey and all
  // events are aggregated into one array.  Events in this array will necessarily be in
  // sorted order.
  var groups []EventGroup
  for _, os_event := range os_events {
    group := EventGroup{
      Timestamp: os_event.Timestamp,
    }
    input.pressKey(
      input.GetKey(os_event.KeyId),
      os_event.Press_amt,
      Event{},
      &group)

    // Sets the cursor position if this is a cursor based event.
    // TODO: Currently only the mouse is supported as a cursor, but if we want to support
    //       joysticks as cursor_keys, since they don't naturally have a position associated
    //       with them, we will need to somehow associate cursor_keys with axes and treat the
    //       mouse and joysticks separately.
    if cursor := input.cursor_keys[os_event.KeyId]; cursor != nil {
      cursor.X = os_event.X
      cursor.Y = os_event.Y
    }

    //    for i := range group.Events {
    //      group.Events[i].Mouse = os_event.Mouse
    //    }

    if len(group.Events) > 0 {
      groups = append(groups, group)
      for _, listener := range input.listeners {
        listener.HandleEventGroup(group)
      }
    }
  }

  for _, key := range input.all_keys {
    gen,amt := key.Think(t)
    if !gen { continue }
    group := EventGroup{ Timestamp: t }
    input.pressKey(key, amt, Event{}, &group)
    if len(group.Events) > 0 {
      groups = append(groups, group)
      for _, listener := range input.listeners {
        listener.HandleEventGroup(group)
      }
    }
  }

  for _, listener := range input.listeners {
    listener.Think(t)
  }
  return groups
}
