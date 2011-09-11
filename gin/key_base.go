package gin

import (
  "fmt"
)


type Key interface {
  String() string
  // Human readable name

  Id() KeyId
  // Unique Id

  SetPressAmt(amt float64, ms int64, cause Event) Event
  // Sets the instantaneous press amount for this key at a specific time and returns the
  // event generated, if any

  CurPressAmt() float64

  IsDown() bool
  // Not necessarily the same as CurPressAmt() > 0, derived keys can have a press amount
  // without being 'pressed'

  FramePressCount() int
  FrameReleaseCount() int
  FramePressAmt() float64
  FramePressSum() float64
  Think(ms int64)
}

// Simple struct that aggregates presses and press_amts during a frame so they can be viewed
// between Think()s
type keyStats struct {
  press_count   int
  release_count int
  press_amt     float64
  press_sum     float64
}

type KeyId int

// natural keys and derived keys all embed a keyState
type keyState struct {
  id   KeyId   // Unique id among all keys ever
  name string  // Human readable name for the key, 'Right Shift', 'q', 'Space Bar', etc...

  this    keyStats  // keyStats for this frame
  prev    keyStats  // keyStats for the previous frame.  Won't change between Think()s

  last_think int64  // time that the last call to Think() happened
  last_press int64  // time that the lass call to SetPressAmt() or Think() happened
}

func (ks *keyState) String() string {
  return fmt.Sprintf("%d: %s", ks.id, ks.name)
}

func (ks *keyState) Id() KeyId {
  return ks.id
}


// Tells this key that how much it was pressed at a particular time.  Times must be
// monotonically increasing.
// If this press was caused by another event (as is the case with derived keys), then
// cause is the event that made this happen.
func (ks *keyState) SetPressAmt(amt float64, ms int64, cause Event) (event Event) {
  event.Type = NoEvent
  if (ks.this.press_amt == 0) != (amt == 0) {
    event.Key = ks
    if amt == 0 {
      event.Type = Release
      ks.this.release_count++
    } else {
      event.Type = Press
      ks.this.press_count++
    }
  }
  ks.this.press_sum += ks.this.press_amt * float64(ms - ks.last_press)
  ks.this.press_amt = amt
  ks.last_press = ms
  return
}

func (ks *keyState) CurPressAmt() float64 {
  return ks.this.press_amt
}
func (ks *keyState) IsDown() bool {
  return ks.this.press_amt != 0
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
func (ks *keyState) FramePressSum() float64 {
  return ks.prev.press_sum
}
func (ks *keyState) Think(t int64) {
  ks.this.press_sum += ks.this.press_amt * float64(t - ks.last_press)
  ks.prev = ks.this
  ks.this = keyStats{
    press_amt : ks.prev.press_amt,
  }
  ks.last_think = t
  ks.last_press = t
}

