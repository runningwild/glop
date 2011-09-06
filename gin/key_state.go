package gin

import (
  "fmt"
)

const (
  doublePressThreshold = 200
)

// All keys, natural and derived, implement the Key interface
type Key interface {
  String() string
}

// All keys, natural and derived, contain a unique KeyId that can stringify itself into a
// human-recognizable name like "Left Shift", or "q"
type KeyId struct {
  id   int
  name string
}
func (k KeyId) String() string {
  return fmt.Sprintf("Key# %d", k)
}

type KeyEventType int
const (
  NoEvent KeyEventType = iota
  Release
  Press
  RepeatPress
  DoublePress
)

// Helper class that stores all information on a key's state and performs some basic logic that
// is common to all keys, both normal and derived. It is also provided publicly so that the logic
// can be duplicated elsewhere - e.g. GUI button tracking.
type KeyState struct {
  press_amt_now   float64  // Instantaneous press_amount for the key
  press_amt_frame float64  // Average press_amount for the key this frame

  is_down_now   bool  // Is the key down now - not quite the same as cur_press_amount>0
  is_down_frame bool  // Was the key down at any point this frame (See IsKeyDown)

  total_frame_time       int  // Total time spent during this frame
  double_press_time_left int  // Used for tracking double presses

  was_pressed            bool  // Was a key press event generated for this key this frame
  was_pressed_no_repeats bool  // Was a key press (not repeat) event generated this frame
  was_released           bool  // Was a key release event generated for this key this frame
}

// Mark this key as up or down. This is not necessarily the same thing as press_amount > 0.
// Some keys, such as mouse motion, marks the key artifically as "down" briefly after
// press_amount becomes 0. This makes things smoother. Release events are automatically
// generated. Press events can optionally be generated as well. For derived keys, however, we
// copy events from the keys they're derived from instead.
func (ks *KeyState) SetIsDown(is_down, generate_press_event bool) KeyEventType {
  if is_down == ks.is_down_now {
    return NoEvent
  }
  ks.is_down_now   = is_down
  ks.is_down_frame = ks.is_down_frame || is_down
  if !is_down {
    ks.was_released = true
    return Release
  }
  if generate_press_event {
    ks.was_pressed = true
    if ks.double_press_time_left > 0 {
      ks.double_press_time_left = 0
      return DoublePress
    }
    ks.double_press_time_left = doublePressThreshold
    return Press
  }
  return NoEvent
}

// Updates how far down the key is pressed. This is independent here of whether the key is
// down.
func (ks *KeyState) SetPressAmount(amt float64) {
  ks.press_amt_now = amt
}

// Registers that a press-type event was generated for this key. This is reflected only in
// the WasPressed function. Release events are generated automatically. The others are not
// necessarily (see SetIsDown).
func (ks *KeyState) OnKeyEvent(typ KeyEventType) {
  switch typ {
    case Press:
      fallthrough
    case DoublePress:
      ks.was_pressed_no_repeats = true
      fallthrough
    case RepeatPress:
      ks.was_pressed = true
    default:
      panic("Called KeyState.OnKeyEvent() with an invalid KeyEventType")
  }
}

// Registers that time has passed. Repeat events are not generated automatically, but this is
// used to track frame press amount.
func (ks *KeyState) OnDt(dt int) {
  ks.press_amt_frame =
      (ks.press_amt_frame * float64(ks.total_frame_time) + ks.press_amt_now *float64(dt)) /
      float64(ks.total_frame_time + dt)
  ks.total_frame_time += dt

  ks.double_press_time_left -= dt
  if ks.double_press_time_left < 0 {
    ks.double_press_time_left = 0
  }
}

// Registers that a new frame has begun
func (ks *KeyState) Think() {
  ks.press_amt_frame = ks.press_amt_now
  ks.total_frame_time = 0
  ks.is_down_frame = ks.is_down_now
  ks.was_pressed = false
  ks.was_pressed_no_repeats = false
  ks.was_released = false
}

// Instantaneous status - see top of the file and Input class below.
func (ks *KeyState) GetPressAmountNow() float64 {
  return ks.press_amt_now
}

func (ks *KeyState) IsDownNow() bool {
  return ks.is_down_now
}

// Frame status - see top of the file and Input class below.
func (ks *KeyState) GetPressAmountFrame() float64 {
  return ks.press_amt_frame
}

func (ks *KeyState) IsDownFrame() bool {
  return ks.is_down_frame
}

func (ks *KeyState) WasPressed() bool {
  return ks.was_pressed
}

func (ks *KeyState) WasPressedWithoutRepeats() bool {
  return ks.was_pressed_no_repeats
}

func (ks *KeyState) WasReleased() bool {
  return ks.was_released
}
