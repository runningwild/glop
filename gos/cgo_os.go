package gos

// #include "include/glop.h"
import "C"

import (
  "glop/system"
  "unsafe"
)

type osxWindow struct {
  window  uintptr  // NSWindow*
  context uintptr  // NSOpenGLContext*
}


type osxSystemObject struct {}
var (
  osx_system_object osxSystemObject
)
func init() {
  C.Init()
}

func GetSystemObject() system.System {
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
func (osx *osxSystemObject) GetInputEvents() []system.KeyEvent {
  var first_event *C.KeyEvent
  cp := (*unsafe.Pointer)(unsafe.Pointer(&first_event))
  var length C.int
  C.GetInputEvents(cp, &length)
  c_events := (*[1000]C.KeyEvent)(unsafe.Pointer(first_event))[:length]
  events := make([]system.KeyEvent, length)
  for i := range c_events {
    events[i] = system.KeyEvent{
      Index     : int(c_events[i].index),
      Press_amt : float64(c_events[i].press_amt),
      Timestamp : int64(c_events[i].timestamp),
      Mouse : system.Mouse{
        Dx : int(c_events[i].mouse_dx),
        Dy : int(c_events[i].mouse_dy),
        X : int(c_events[i].cursor_x),
        Y : int(c_events[i].cursor_y),
      },
    }
  }
  return events
}

func (osx *osxSystemObject) CursorPos(window system.Window) (int,int) {
  osx_window := (*osxWindow)(unsafe.Pointer(window))
  var x,y int
  C.CurrentMousePos(unsafe.Pointer(osx_window.window), unsafe.Pointer(&x), unsafe.Pointer(&y));
  return x,y
}

// TODO: Duh
func (osx *osxSystemObject) WindowPos(window system.Window) (int,int) {
//  osx_window := (*osxWindow)(unsafe.Pointer(window))
  return 0,0
}
