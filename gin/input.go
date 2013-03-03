package gin

import (
	"fmt"
)

const (
	AnyKey               KeyIndex = 0
	Space                         = 32
	Backspace                     = 8
	Tab                           = 9
	Return                        = 13
	Escape                        = 27
	KeyA                          = 97
	KeyB                          = 98
	KeyC                          = 99
	KeyD                          = 100
	KeyE                          = 101
	KeyF                          = 102
	KeyG                          = 103
	KeyH                          = 104
	KeyI                          = 105
	KeyJ                          = 106
	KeyK                          = 107
	KeyL                          = 108
	KeyM                          = 109
	KeyN                          = 110
	KeyO                          = 111
	KeyP                          = 112
	KeyQ                          = 113
	KeyR                          = 114
	KeyS                          = 115
	KeyT                          = 116
	KeyU                          = 117
	KeyV                          = 118
	KeyW                          = 119
	KeyX                          = 120
	KeyY                          = 121
	KeyZ                          = 122
	F1                            = 129
	F2                            = 130
	F3                            = 131
	F4                            = 132
	F5                            = 133
	F6                            = 134
	F7                            = 135
	F8                            = 136
	F9                            = 137
	F10                           = 138
	F11                           = 139
	F12                           = 140
	CapsLock                      = 150
	NumLock                       = 151
	ScrollLock                    = 152
	PrintScreen                   = 153
	Pause                         = 154
	LeftShift                     = 155
	RightShift                    = 156
	LeftControl                   = 157
	RightControl                  = 158
	LeftAlt                       = 159
	RightAlt                      = 160
	LeftGui                       = 161
	RightGui                      = 162
	Right                         = 166
	Left                          = 167
	Up                            = 168
	Down                          = 169
	KeyPadDivide                  = 170
	KeyPadMultiply                = 171
	KeyPadSubtract                = 172
	KeyPadAdd                     = 173
	KeyPadEnter                   = 174
	KeyPadDecimal                 = 175
	KeyPadEquals                  = 176
	KeyPad0                       = 177
	KeyPad1                       = 178
	KeyPad2                       = 179
	KeyPad3                       = 180
	KeyPad4                       = 181
	KeyPad5                       = 182
	KeyPad6                       = 183
	KeyPad7                       = 184
	KeyPad8                       = 185
	KeyPad9                       = 186
	KeyDelete                     = 190
	KeyHome                       = 191
	KeyInsert                     = 192
	KeyEnd                        = 193
	KeyPageUp                     = 194
	KeyPageDown                   = 195
	MouseXAxis                    = 300
	MouseYAxis                    = 301
	MouseWheelVertical            = 302
	MouseWheelHorizontal          = 303
	MouseLButton                  = 304
	MouseRButton                  = 305
	MouseMButton                  = 306

	// standard derived keys start here
	EitherShift = 1000 + iota
	EitherControl
	EitherAlt
	EitherGui
	ShiftTab
	DeleteOrBackspace
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
	id_to_deps map[KeyId][]Key

	// map from KeyIndex to an aggregator of the appropriate type for that index,
	// this allows us to construct Keys for devices as the events happen, rather
	// than needing to know what the devices are during initialization.
	index_to_agg_type map[KeyIndex]aggregatorType

	// map from KeyIndex to a human-readable name for that key
	index_to_name map[KeyIndex]string

	// The listeners will receive all events immediately after those events have been used to
	// update all key states.  The order in which listeners are notified of a particular event
	// group can change from group to group.
	listeners []Listener
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
	input.id_to_deps = make(map[KeyId][]Key, 16)
	input.index_to_agg_type = make(map[KeyIndex]aggregatorType)
	input.index_to_name = make(map[KeyIndex]string)

	input.registerKeyIndex(AnyKey, aggregatorTypeStandard, "AnyKey")
	for c := 'a'; c <= 'z'; c++ {
		name := fmt.Sprintf("Key %c", c+'A'-'a')
		input.registerKeyIndex(KeyIndex(c), aggregatorTypeStandard, name)
	}
	for _, c := range "0123456789`[]\\-=;',./" {
		name := fmt.Sprintf("Key %c", c)
		input.registerKeyIndex(KeyIndex(c), aggregatorTypeStandard, name)
	}
	input.registerKeyIndex(Space, aggregatorTypeStandard, "Space")
	input.registerKeyIndex(Backspace, aggregatorTypeStandard, "Backspace")
	input.registerKeyIndex(Tab, aggregatorTypeStandard, "Tab")
	input.registerKeyIndex(Return, aggregatorTypeStandard, "Return")
	input.registerKeyIndex(Escape, aggregatorTypeStandard, "Escape")
	input.registerKeyIndex(F1, aggregatorTypeStandard, "F1")
	input.registerKeyIndex(F2, aggregatorTypeStandard, "F2")
	input.registerKeyIndex(F3, aggregatorTypeStandard, "F3")
	input.registerKeyIndex(F4, aggregatorTypeStandard, "F4")
	input.registerKeyIndex(F5, aggregatorTypeStandard, "F5")
	input.registerKeyIndex(F6, aggregatorTypeStandard, "F6")
	input.registerKeyIndex(F7, aggregatorTypeStandard, "F7")
	input.registerKeyIndex(F8, aggregatorTypeStandard, "F8")
	input.registerKeyIndex(F9, aggregatorTypeStandard, "F9")
	input.registerKeyIndex(F10, aggregatorTypeStandard, "F10")
	input.registerKeyIndex(F11, aggregatorTypeStandard, "F11")
	input.registerKeyIndex(F12, aggregatorTypeStandard, "F12")
	input.registerKeyIndex(CapsLock, aggregatorTypeStandard, "CapsLock")
	input.registerKeyIndex(NumLock, aggregatorTypeStandard, "NumLock")
	input.registerKeyIndex(ScrollLock, aggregatorTypeStandard, "ScrollLock")
	input.registerKeyIndex(PrintScreen, aggregatorTypeStandard, "PrintScreen")
	input.registerKeyIndex(Pause, aggregatorTypeStandard, "Pause")
	input.registerKeyIndex(LeftShift, aggregatorTypeStandard, "LeftShift")
	input.registerKeyIndex(RightShift, aggregatorTypeStandard, "RightShift")
	input.registerKeyIndex(LeftControl, aggregatorTypeStandard, "LeftControl")
	input.registerKeyIndex(RightControl, aggregatorTypeStandard, "RightControl")
	input.registerKeyIndex(LeftAlt, aggregatorTypeStandard, "LeftAlt")
	input.registerKeyIndex(RightAlt, aggregatorTypeStandard, "RightAlt")
	input.registerKeyIndex(LeftGui, aggregatorTypeStandard, "LeftGui")
	input.registerKeyIndex(RightGui, aggregatorTypeStandard, "RightGui")
	input.registerKeyIndex(Right, aggregatorTypeStandard, "Right")
	input.registerKeyIndex(Left, aggregatorTypeStandard, "Left")
	input.registerKeyIndex(Up, aggregatorTypeStandard, "Up")
	input.registerKeyIndex(Down, aggregatorTypeStandard, "Down")
	input.registerKeyIndex(KeyPadDivide, aggregatorTypeStandard, "KeyPadDivide")
	input.registerKeyIndex(KeyPadMultiply, aggregatorTypeStandard, "KeyPadMultiply")
	input.registerKeyIndex(KeyPadSubtract, aggregatorTypeStandard, "KeyPadSubtract")
	input.registerKeyIndex(KeyPadAdd, aggregatorTypeStandard, "KeyPadAdd")
	input.registerKeyIndex(KeyPadEnter, aggregatorTypeStandard, "KeyPadEnter")
	input.registerKeyIndex(KeyPadDecimal, aggregatorTypeStandard, "KeyPadDecimal")
	input.registerKeyIndex(KeyPadEquals, aggregatorTypeStandard, "KeyPadEquals")
	input.registerKeyIndex(KeyPad0, aggregatorTypeStandard, "KeyPad0")
	input.registerKeyIndex(KeyPad1, aggregatorTypeStandard, "KeyPad1")
	input.registerKeyIndex(KeyPad2, aggregatorTypeStandard, "KeyPad2")
	input.registerKeyIndex(KeyPad3, aggregatorTypeStandard, "KeyPad3")
	input.registerKeyIndex(KeyPad4, aggregatorTypeStandard, "KeyPad4")
	input.registerKeyIndex(KeyPad5, aggregatorTypeStandard, "KeyPad5")
	input.registerKeyIndex(KeyPad6, aggregatorTypeStandard, "KeyPad6")
	input.registerKeyIndex(KeyPad7, aggregatorTypeStandard, "KeyPad7")
	input.registerKeyIndex(KeyPad8, aggregatorTypeStandard, "KeyPad8")
	input.registerKeyIndex(KeyPad9, aggregatorTypeStandard, "KeyPad9")
	input.registerKeyIndex(KeyDelete, aggregatorTypeStandard, "KeyDelete")
	input.registerKeyIndex(KeyHome, aggregatorTypeStandard, "KeyHome")
	input.registerKeyIndex(KeyInsert, aggregatorTypeStandard, "KeyInsert")
	input.registerKeyIndex(KeyEnd, aggregatorTypeStandard, "KeyEnd")
	input.registerKeyIndex(KeyPageUp, aggregatorTypeStandard, "KeyPageUp")
	input.registerKeyIndex(KeyPageDown, aggregatorTypeStandard, "KeyPageDown")

	input.registerKeyIndex(MouseXAxis, aggregatorTypeAxis, "X Axis")
	input.registerKeyIndex(MouseYAxis, aggregatorTypeAxis, "Y Axis")
	input.registerKeyIndex(MouseWheelVertical, aggregatorTypeWheel, "MouseWheel")
	input.registerKeyIndex(MouseLButton, aggregatorTypeStandard, "MouseLButton")
	input.registerKeyIndex(MouseRButton, aggregatorTypeStandard, "MouseRButton")
	input.registerKeyIndex(MouseMButton, aggregatorTypeStandard, "MouseMButton")

	// input.bindDerivedKeyWithId("Shift", EitherShift, input.MakeBinding(LeftShift, nil, nil), input.MakeBinding(RightShift, nil, nil))
	// input.bindDerivedKeyWithId("Control", EitherControl, input.MakeBinding(LeftControl, nil, nil), input.MakeBinding(RightControl, nil, nil))
	// input.bindDerivedKeyWithId("Alt", EitherAlt, input.MakeBinding(LeftAlt, nil, nil), input.MakeBinding(RightAlt, nil, nil))
	// input.bindDerivedKeyWithId("Gui", EitherGui, input.MakeBinding(LeftGui, nil, nil), input.MakeBinding(RightGui, nil, nil))
	// input.bindDerivedKeyWithId("ShiftTab", ShiftTab, input.MakeBinding(Tab, []KeyId{EitherShift}, []bool{true}))
	// input.bindDerivedKeyWithId("DeleteOrBackspace", DeleteOrBackspace, input.MakeBinding(KeyDelete, nil, nil), input.MakeBinding(Backspace, nil, nil))
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

func (input *Input) registerKeyIndex(index KeyIndex, agg_type aggregatorType, name string) {
	if index < 0 {
		panic(fmt.Sprintf("Cannot register a key with index %d, indexes must be greater than 0.", index))
	}
	if prev, ok := input.index_to_agg_type[index]; ok {
		panic(fmt.Sprintf("Cannot register key index %d, it has already been registered with the name %s and aggregator %v.", index, prev, agg_type))
	}
	input.index_to_agg_type[index] = agg_type
	input.index_to_name[index] = name
}

func (input *Input) GetKeyFlat(key_index KeyIndex, device_type DeviceType, device_index DeviceIndex) Key {
	return input.GetKey(KeyId{
		Index: key_index,
		Device: DeviceId{
			Index: device_index,
			Type:  device_type,
		},
	})
}

func (input *Input) GetKey(id KeyId) Key {
	if id.Device.Type >= DeviceTypeMax || id.Device.Type < 0 {
		panic(fmt.Sprintf("Specied invalid DeviceType, %d.", id.Device))
	}
	key, ok := input.key_map[id]
	if !ok {
		// If we're looking for a general key we know how to create those
		if id.Index == AnyKey || id.Device.Type == DeviceTypeAny || id.Device.Index == 0 {
			if id.Device.Type == DeviceTypeAny && id.Device.Index != 0 {
				panic("Cannot specify a Device Index but not a Device Type.")
			}
			input.key_map[id] = &generalDerivedKey{
				keyState: keyState{
					id:         id,
					name:       "Name me?",
					aggregator: &standardAggregator{},
				},
				input: input,
			}
		} else {
			// Check if the index is valid, if it is then we can just create a new key
			// the appropriate device.
			agg_type, ok := input.index_to_agg_type[id.Index]
			if !ok {
				panic(fmt.Sprintf("No key registered with id == %v.", id))
			}
			var agg aggregator
			switch agg_type {
			case aggregatorTypeStandard:
				agg = &standardAggregator{}
			case aggregatorTypeAxis:
				agg = &axisAggregator{}
			case aggregatorTypeWheel:
				agg = &wheelAggregator{}
			default:
				panic(fmt.Sprintf("Unknown aggregator type specified: %T.", agg_type))
			}
			// TODO: duplicate the aggregator, probably needs to use reflection
			input.key_map[id] = &keyState{
				id:         id,
				name:       input.index_to_name[id.Index],
				aggregator: agg,
			}
		}
		key = input.key_map[id]
		input.all_keys = append(input.all_keys, key)
	}
	return key
}
func (input *Input) GetKeyByName(name string) Key {
	for _, key := range input.key_map {
		if key.Name() == name {
			return key
		}
	}
	return nil
}

func (input *Input) informDeps(event Event, group *EventGroup) {
	id := event.Key.Id()
	any_device := id.Device
	any_device.Index = 0
	deps := input.id_to_deps[id]
	for _, dep := range input.id_to_deps[KeyId{Index: id.Index, Device: any_device}] {
		deps = append(deps, dep)
	}
	for _, dep := range deps {
		input.pressKey(dep, dep.CurPressAmt(), event, group)
	}
	if event.Type != NoEvent {
		group.Events = append(group.Events, event)
	}
}

var foo int

func (input *Input) pressKey(k Key, amt float64, cause Event, group *EventGroup) {
	event := k.SetPressAmt(amt, group.Timestamp, cause)
	input.informDeps(event, group)
	if k.Id().Index != AnyKey && k.Id().Device.Type != DeviceTypeAny && k.Id().Device.Type != DeviceTypeDerived && k.Id().Device.Index != 0 {
		general_keys := []Key{
			input.GetKeyFlat(AnyKey, k.Id().Device.Type, k.Id().Device.Index),
			input.GetKeyFlat(AnyKey, k.Id().Device.Type, 0),
			input.GetKeyFlat(AnyKey, DeviceTypeAny, 0),
			input.GetKeyFlat(k.Id().Index, k.Id().Device.Type, 0),
			input.GetKeyFlat(k.Id().Index, DeviceTypeAny, 0),
		}
		for _, general_key := range general_keys {
			input.pressKey(general_key, amt, cause, group)
		}
	}
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
		// if cursor := input.cursor_keys[os_event.KeyId]; cursor != nil {
		// 	cursor.X = os_event.X
		// 	cursor.Y = os_event.Y
		// }

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
		gen, amt := key.Think(t)
		if !gen {
			continue
		}
		group := EventGroup{Timestamp: t}
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
