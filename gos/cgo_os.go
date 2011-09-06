package gos

// #include "include/glop.h"
import "C"

import (
  "unsafe"
)

type osxWindow struct {
  window  uintptr  // NSWindow*
  context uintptr  // NSOpenGLContext*
}


func init() {
  C.Init()
}

func Run() {
  C.Run()
}

func Quit() {
  C.Quit()
}

func CreateWindow(x,y,width,height int) Window {
  var window osxWindow
  w := (*unsafe.Pointer)(unsafe.Pointer(&window.window))
  c := (*unsafe.Pointer)(unsafe.Pointer(&window.context))
  C.CreateWindow(w, c, C.int(x), C.int(y), C.int(width), C.int(height))
  return Window(unsafe.Pointer(&window))
}

func SwapBuffers(window Window) {
  osx_window := (*osxWindow)(unsafe.Pointer(window))
  C.SwapBuffers(unsafe.Pointer(osx_window.context))
}

func Think() {
  C.Think()
}

// TODO: Make sure that events are given in sorted order (by timestamp)
// TODO: Adjust timestamp on events so that the oldest timestamp is newer than the
//       newest timestemp from the events from the previous call to GetInputEvents
//       Actually that should be in system
func GetInputEvents() []KeyEvent {
  var first_event *C.KeyEvent
  cp := (*unsafe.Pointer)(unsafe.Pointer(&first_event))
  var length C.int
  C.GetInputEvents(cp, &length)
  c_events := (*[1000]C.KeyEvent)(unsafe.Pointer(first_event))[:length]
  events := make([]KeyEvent, length)
  for i := range c_events {
    events[i] = KeyEvent{
      Index     : int(c_events[i].index),
      Press_amt : float64(c_events[i].press_amt),
      Timestamp : int64(c_events[i].timestamp),
      Mouse : Mouse{
        Dx : int(c_events[i].mouse_dx),
        Dy : int(c_events[i].mouse_dy),
        X : int(c_events[i].cursor_x),
        Y : int(c_events[i].cursor_y),
      },
    }
  }
  return events
}

func CursorPos(window Window) (int,int) {
  osx_window := (*osxWindow)(unsafe.Pointer(window))
  var x,y int
  C.CurrentMousePos(unsafe.Pointer(osx_window.window), unsafe.Pointer(&x), unsafe.Pointer(&y));
  return x,y
}

// TODO: Duh
func WindowPos(window Window) (int,int) {
//  osx_window := (*osxWindow)(unsafe.Pointer(window))
  return 0,0
}