package gin

import (
  "fmt"
)


type Key interface {
  String() string
  // Human readable name

  Id() KeyId
  // Unique Id

  SetPressAmt(amt float64, ms int64) *Event
  // Sets the instantaneous press amount for this key at a specific time and returns the
  // event generated, if any

  CurPressAmt() float64
  
  FramePressCount() int
  FrameReleaseCount() int
  FramePressAmt() float64
  FramePressAvg() float64
  Think(ms int64)
}

// Simple struct that aggregates presses and press_amts during a frame so they can be viewed
// between Think()s
type keyStats struct {
  press_count   int
  release_count int
  press_amt     float64
  press_avg     float64
}

type KeyId int

// natural keys and derived keys all embed a keyState
type keyState struct {
  id   KeyId   // Unique id among all keys ever
  name string  // Human readable name for the key, 'Right Shift', 'q', 'Space Bar', etc...

  this keyStats  // keyStats for this frame
  prev keyStats  // keyStats for the previous frame.  Won't change between Think()s

  last_think int64  // time that the last call to Think() happened
  last_press int64  // time that the lass call to SetPressAmt() or Think() happened
}

func (ks *keyState) String() string {
  return fmt.Sprintf("%d: %s", ks.id, ks.name)
}

func (ks *keyState) Id() KeyId {
  return ks.id
}

func (ks *keyState) SetPressAmt(amt float64, ms int64) (event *Event) {
  if (ks.this.press_amt == 0) == (amt == 0) {
    event = nil
  } else {
    event = &Event {
      Key : ks,
      Timestamp : ms,
    }
    if amt == 0 {
      event.Type = Release
      ks.this.release_count++
    } else {
      event.Type = Press
      ks.this.press_count++
    }
  }
  ks.this.press_avg += amt * float64(ms - ks.last_press)
  ks.this.press_amt = amt
  ks.last_press = ms
  return
}

func (ks *keyState) CurPressAmt() float64 {
  return ks.prev.press_amt
}
func (ks *keyState) FramePressCount() int {
  return ks.prev.press_count
}
func (ks *keyState) FrameReleaseCount() int {
  return ks.prev.release_count
}
func (ks *keyState) FramePressAmt() float64 {
  return ks.prev.press_amt
}
func (ks *keyState) FramePressAvg() float64 {
  return ks.prev.press_avg
}
func (ks *keyState) Think(t int64) {
  ks.this.press_avg += ks.this.press_amt * float64(t - ks.last_press)
  ks.this.press_avg /= float64(t - ks.last_think)
  ks.prev = ks.this
  ks.this = keyStats{}
  ks.last_think = t
  ks.last_press = t
}

