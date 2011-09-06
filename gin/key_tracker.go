package gin

// A helper class, built upon KeyState, that performs all logic for natural keys.
type keyTracker struct {
  KeyId
  KeyState

  requested_press_amt float64
  release_delay_left  int
  release_delay       int
  repeat_delay_left   int

  mouse_wheel_hack bool
}

func (kt *keyTracker) SetReleaseDelay(delay int, mouse_wheel_hack bool) {
  kt.release_delay = delay
  kt.mouse_wheel_hack = mouse_wheel_hack
}

func (kt *keyTracker) Clear() KeyEventType {
  kt.KeyState.SetPressAmount(0)
  return kt.KeyState.SetIsDown(false, true)
}

func (kt *keyTracker) SetPressAmount(amt float64) KeyEventType {
  kt.requested_press_amt = amt

  // Handle presses
  if amt > 0 {
    kt.KeyState.SetPressAmount(amt)
    kt.release_delay_left = kt.release_delay
    if !kt.IsDownNow() {
      kt.repeat_delay_left = repeatDelay
      return kt.KeyState.SetIsDown(true, true)
    }
    // For the mouse wheel, we hold the press amount down for a period of time to make it function
    // similar to other keys for smooth movement. However, tracking presses is done much better by
    // just sending events directly from the Os.
    if kt.mouse_wheel_hack {
      return RepeatPress
    }

  // Handle releases
  } else if kt.IsDownNow() {
    if kt.release_delay == 0 {
      kt.KeyState.SetPressAmount(0)
      return kt.KeyState.SetIsDown(false, true)
    } else if !kt.mouse_wheel_hack {
      kt.KeyState.SetPressAmount(0)
    }
  }
  return NoEvent
}

func (kt *keyTracker) SetIsDown(is_down bool) KeyEventType {
  amt := 0.0
  if is_down {
    amt = 1.0
  }
  return kt.SetPressAmount(amt)
}

func (kt *keyTracker) OnDt(dt int) KeyEventType {
  if dt < 0 {
    panic("Cannot have dt < 0")
  }
  if dt == 0 {
    return NoEvent
  }
  kt.KeyState.OnDt(dt)

  if kt.IsDownNow() {
    // Handle releases
    if kt.requested_press_amt == 0 {
      kt.release_delay_left -= dt
      if kt.release_delay_left <= 0 {
        kt.KeyState.SetPressAmount(0)
        return kt.KeyState.SetIsDown(false, true)
      }
    }

    // Handle repeat events
    if kt.IsDownNow() && !kt.mouse_wheel_hack {
      kt.repeat_delay_left -= dt
      if kt.repeat_delay_left <= 0 {
        kt.OnKeyEvent(RepeatPress)
        kt.repeat_delay_left += repeatRate
        return RepeatPress
      }
    }
  }
  return NoEvent
}
