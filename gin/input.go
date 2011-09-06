package gin

import (
//  "glop/gos"
)

var (
  all_keys, down_keys, pressed_keys, derived_keys []Key
  key_map map[int]Key
)

func init() {
  all_keys = make([]Key, 16)[0:0]
  down_keys = make([]Key, 16)[0:0]
  pressed_keys = make([]Key, 16)[0:0]
  derived_keys = make([]Key, 16)[0:0]

  key_map = make(map[int]Key, 128)
}

const (
  // Key repeat-rate constants
  repeatDelay = 500  // Ms between key down event #1 and #2 while a key is down
  repeatRate  = 60   // Ms between later key down events while a key is down
)

/*

func updateDerivedKeys() {
  derived_key_states_.resize(GetNumDerivedKeys());
  for (int i = 0; i < GetNumDerivedKeys(); i++) {
    GlopKey key(i, kDeviceDerived);
    vector<GlopKey> released_keys;
    UpdateDerivedKeyState(key, &released_keys);
    if (released_keys.size() > 0)
      OnKeyEvent(KeyEvent(released_keys, KeyEvent::Release));
  }
}

// Figures out whether the OS should be displaying or hiding our cursor, and then informs it if
// it is doing the wrong thing.
func updateOsCursorVisibility() {
  bool is_in_focus, focus_changed;
  bool os_is_visible = true;
  int width, height;
  if (!is_cursor_visible_) {
    Os::GetWindowFocusState(window_->os_data_, &is_in_focus, &focus_changed);
    Os::GetWindowSize(window_->os_data_, &width, &height);
    os_is_visible = (!is_in_focus || mouse_x_ < 0 || mouse_y_ < 0 ||
                      mouse_x_ >= width || mouse_y_ >= height);
  }
  if (os_is_visible != os_is_cursor_visible_) {
    Os::ShowMouseCursor(os_is_visible);
    os_is_cursor_visible_ = os_is_visible;
  }
}

func clearAllKeyState() {
  // Set the current state
  is_num_lock_set_ = os_events[n - 1].is_num_lock_set;
  is_caps_lock_set_ = os_events[n - 1].is_caps_lock_set;
  mouse_x_ = os_events[n - 1].cursor_x - window_x_;
  mouse_y_ = os_events[n - 1].cursor_y - window_y_;

  // Release all keys. Releasing each non-derived key will also release the derived keys, but
  // we need to explicitly recalculate the press amount (to 0) for derived keys.
  for (int i = kMinDevice; i <= GetMaxDevice(); i++)
  for (int j = 0; j < GetNumKeys(i); j++)
  if (!GlopKey(j, i).IsDerivedKey()) {
    if (GetNonDerivedKeyTracker(GlopKey(j, i))->Clear() == KeyEvent::Release)
      UpdateDerivedKeyStatesAndProcessEvents(GlopKey(j, i), KeyEvent::Release);
  }
}

// Performs all per-frame logic for the input manager. lost_focus indicates whether the owning
// window either lost input focus or was recreated during the last frame. In either case, we need
// to forget all key down information because it may no longer be current.
func Think(lost_focus bool) {
  // Update all editable derived keys - this is useful in the case that the user has changed their
  // definitions since the last frame. We do it here rather than at key-binding time since the data
  // here is naturally tied to a single window, but key binding is naturally static across all
  // windows.
  updateDerivedKeys()

  // Update mouse and joystick status. If the number of joysticks changes, we do a full reset of
  // all input data.
  UpdateOsCursorVisibility();

  // TODO: If we add joystick support, do it here

  // What has happened since our last poll? Even if we have gone out of focus, we still promise to
  // call Os::GetInputEvents.
  os_events := gos.GetInputEvents()
  // TODO: Where should we get the widow pointer from?
  window_x, window_y = gos.WindowPosition(nil)
  Os::GetWindowPosition(window_->os_data_, &window_x_, &window_y_);
  if len(os_events) == 0 {
    panic("Expected at least one event from a call to gos.GetInputEvents()")
  }

  // If we have lost focus, clear all key state. Note that down_keys_frame_ is rebuilt every frame
  // regardless, so we do not need to worry about it here.
  if lost_focus {
    clearAllKeyState()
  }

  // Do all per-frame logic on keys.
  down_keys = down_keys[0:0]
  pressed_keys = pressed_keys[0:0]
  for _,key := range all_keys {
    key.Think()
  }
//  if (GlopKey(j, i).IsDerivedKey())
//    GetKeyState(GlopKey(j, i))->Think();
//  else
//    GetNonDerivedKeyTracker(GlopKey(j, i))->Think();

  // Now update key statuses for this frame. We only do this if !lost_focus. Otherwise, some down
  // key os_events might have been generated before losing focus, and the corresponding up key
  // event may never happen.
  if !lost_focus {
    
  }

  if (!lost_focus)
  for (int i = 0; i < n; i++) {
    // Calculate the time differential for this phase and add mouse motion os_events. Note that we
    // do not assume the times are strictly increasing. This is because the dummy input event is
    // likely generated asynchronously from the other os_events, and could thus cause an
    // inconsistent ordering.
    const int kTimeGranularity = 10;
    int new_t = os_events[i].timestamp;
    int old_t = (i == 0? last_poll_time_ : os_events[i-1].timestamp);
    
    // Handle the case where last_poll_time_ is not yet initialized.
    if (!last_poll_time_set_) {
      last_poll_time_ = old_t = new_t;
      last_poll_time_set_ = true;
    }
    int t_boundary = ((old_t + kTimeGranularity - 1) / kTimeGranularity) * kTimeGranularity;

    // Update last_poll_time_, accounting for both overflow and for the fact that new_t may be
    // less than last_poll_time_.
    if ( (new_t - last_poll_time_) >= 0)
      last_poll_time_ = new_t;
    for (int t = t_boundary; (t - new_t) < 0; t += kTimeGranularity) {
      // Send elapsed time messages
      for (int j = kMinDevice; j <= GetMaxDevice(); j++)
      for (int k = 0; k < GetNumKeys(j); k++)
      if (GlopKey(k, j).IsDerivedKey()) {
        GetKeyState(GlopKey(k, j))->OnDt(kTimeGranularity);
      } else {
        KeyTracker *info = GetNonDerivedKeyTracker(GlopKey(k, j));
        KeyEvent::Type type = info->OnDt(kTimeGranularity);
        if (type != KeyEvent::Nothing)
          UpdateDerivedKeyStatesAndProcessEvents(GlopKey(k, j), type);
      }
      float mouse_scale = mouse_sensitivity_ * kBaseMouseSensitivity / kTimeGranularity;
      SetNonDerivedKeyPressAmount(kMouseUp, -mouse_dy_ * mouse_scale);
      SetNonDerivedKeyPressAmount(kMouseRight, mouse_dx_ * mouse_scale);
      SetNonDerivedKeyPressAmount(kMouseDown, mouse_dy_ * mouse_scale);
      SetNonDerivedKeyPressAmount(kMouseLeft, -mouse_dx_ * mouse_scale);
      mouse_dx_ = mouse_dy_ = 0;
      OnKeyEvent(KeyEvent(kTimeGranularity));
    }

    // Update the new settings
    is_num_lock_set_ = os_events[i].is_num_lock_set;
    is_caps_lock_set_ = os_events[i].is_caps_lock_set;
    mouse_x_ = os_events[i].cursor_x - window_x_;
    mouse_y_ = os_events[i].cursor_y - window_y_;

    // Update the total mouse motion - note we do not actually send events until the time exceeds
    // kTimeGranularity, at which point we send them in the above loop.
    if (os_events[i].key == kNoKey) {
      mouse_dx_ += os_events[i].mouse_dx;
      mouse_dy_ += os_events[i].mouse_dy;
      continue;
    }

    // Process all key up/key down os_events from this phase
    ASSERT(!os_events[i].key.IsDerivedKey());
    SetNonDerivedKeyPressAmount(os_events[i].key, os_events[i].press_amount);
  }

  // Fill the down keys vector
  for (int i = kMinDevice; i <= GetMaxDevice(); i++)
  for (int j = 0; j < GetNumKeys(i); j++)
  if (GetKeyState(GlopKey(j, i))->IsDownFrame())
    down_keys_frame_.push_back(GlopKey(j, i));
}

*/
