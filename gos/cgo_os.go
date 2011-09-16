package gos

// #include "include/glop.h"
import "C"

import (
  "glop/system"
  "glop/gin"
  "unsafe"
)

type osxWindow struct {
  window  uintptr  // NSWindow*
  context uintptr  // NSOpenGLContext*
}


type osxSystemObject struct {
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

func (osx *osxSystemObject) CreateWindow(x,y,width,height int) system.Window {
  var window osxWindow
  w := (*unsafe.Pointer)(unsafe.Pointer(&window.window))
  c := (*unsafe.Pointer)(unsafe.Pointer(&window.context))
  C.CreateWindow(w, c, C.int(x), C.int(y), C.int(width), C.int(height))
  return system.Window(unsafe.Pointer(&window))
}

func (osx *osxSystemObject) SwapBuffers(window system.Window) {
  osx_window := (*osxWindow)(unsafe.Pointer(window))
  C.SwapBuffers(unsafe.Pointer(osx_window.context))
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
    events[i] = gin.OsEvent{
      KeyId     : gin.KeyId(c_events[i].index),
      Press_amt : float64(c_events[i].press_amt),
      Timestamp : int64(c_events[i].timestamp),
      Mouse : gin.Mouse{
        X : float64(c_events[i].cursor_x),
        Y : float64(c_events[i].cursor_y),
      },
    }
  }
  return events, osx.horizon
}

func (osx *osxSystemObject) GetCursorPos() (x,y int) {
  _x := unsafe.Pointer(&x)
  _y := unsafe.Pointer(&y)
  C.GetMousePos(_x, _y)
  return
}

func (osx *osxSystemObject) GetWindowDims(window system.Window) (int,int,int,int) {
  var x,y,dx,dy int
  osx_window := (*osxWindow)(unsafe.Pointer(window))
  _x := unsafe.Pointer(&x)
  _y := unsafe.Pointer(&y)
  _dx := unsafe.Pointer(&dx)
  _dy := unsafe.Pointer(&dy)
  C.GetWindowDims(unsafe.Pointer(osx_window.window), _x, _y, _dx, _dy)
  return x, y, dx, dy
}
