package gos

// #include "darwin/include/glop.h"
import "C"

import (
  "glop/system"
  "glop/gin"
  "unsafe"
)

type osxSystemObject struct {
  window  uintptr // NSWindow*
  context uintptr // NSOpenGLContext*
  horizon int64
}

var (
  osx_system_object osxSystemObject
)

// Call after runtime.LockOSThread(), *NOT* in an init function
func (osx *osxSystemObject) Startup() {
  C.Init()
}

func GetSystemInterface() system.Os {
  return &osx_system_object
}

func (osx *osxSystemObject) Run() {
  C.Run()
}

func (osx *osxSystemObject) Quit() {
  C.Quit()
}

func (osx *osxSystemObject) CreateWindow(x, y, width, height int) {
  w := (*unsafe.Pointer)(unsafe.Pointer(&osx.window))
  c := (*unsafe.Pointer)(unsafe.Pointer(&osx.context))
  C.CreateWindow(w, c, C.int(x), C.int(y), C.int(width), C.int(height))
}

func (osx *osxSystemObject) SwapBuffers() {
  C.SwapBuffers(unsafe.Pointer(osx.context))
}

func (osx *osxSystemObject) Think() {
  C.Think()
}

// TODO: Make sure that events are given in sorted order (by timestamp)
// TODO: Adjust timestamp on events so that the oldest timestamp is newer than the
//       newest timestemp from the events from the previous call to GetInputEvents
//       Actually that should be in system
func (osx *osxSystemObject) GetInputEvents() ([]gin.OsEvent, int64) {
  var first_event *C.KeyEvent
  cp := (*unsafe.Pointer)(unsafe.Pointer(&first_event))
  var length C.int
  var horizon C.longlong
  C.GetInputEvents(cp, &length, &horizon)
  osx.horizon = int64(horizon)
  c_events := (*[1000]C.KeyEvent)(unsafe.Pointer(first_event))[:length]
  events := make([]gin.OsEvent, length)
  for i := range c_events {
    wx, wy := osx.rawCursorToWindowCoords(int(c_events[i].cursor_x), int(c_events[i].cursor_y))
    events[i] = gin.OsEvent{
      KeyId:     gin.KeyId(c_events[i].index),
      Press_amt: float64(c_events[i].press_amt),
      Timestamp: int64(c_events[i].timestamp),
      X:         wx,
      Y:         wy,
    }
  }
  return events, osx.horizon
}

func (osx *osxSystemObject) rawCursorToWindowCoords(x, y int) (int, int) {
  wx, wy, _, _ := osx.GetWindowDims()
  return x - wx, y - wy
}

func (osx *osxSystemObject) GetCursorPos() (int, int) {
  var x, y C.int
  C.GetMousePos(&x, &y)
  return osx.rawCursorToWindowCoords(int(x), int(y))
}

func (osx *osxSystemObject) GetWindowDims() (int, int, int, int) {
  var x, y, dx, dy C.int
  C.GetWindowDims(unsafe.Pointer(osx.window), &x, &y, &dx, &dy)
  return int(x), int(y), int(dx), int(dy)
}

func (osx *osxSystemObject) EnableVSync(enable bool) {
  var _enable C.int
  if enable {
    _enable = 1
  }
  C.EnableVSync(unsafe.Pointer(osx.context), _enable)
}
