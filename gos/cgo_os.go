package gos

// #include <glop.h>
import "C"

import (
  "unsafe"
)

type Window struct {
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

func CreateWindow(x,y,width,height int) *Window {
  var window Window
  w := (*unsafe.Pointer)(unsafe.Pointer(&window.window))
  c := (*unsafe.Pointer)(unsafe.Pointer(&window.context))
  C.CreateWindow(w, c, C.int(x), C.int(y), C.int(width), C.int(height))
  return &window
}

func SwapBuffers(window *Window) {
  C.SwapBuffers(unsafe.Pointer(window.context))
}

func Think() {
  C.Think()
}

type Mouse struct {
  X,Y,Dx,Dy int
}
type KeyEvent struct {
  Index     uint16
  Device    uint16
  Press_amt float64
  Mouse     Mouse
  Timestamp int
  Num_lock  int
  Caps_lock int
}

func GetInputEvents() []KeyEvent {
  var first_event *C.KeyEvent
  cp := (*unsafe.Pointer)(unsafe.Pointer(&first_event))
  var length C.int
  C.GetInputEvents(cp, &length)
  c_events := (*[1000]C.KeyEvent)(unsafe.Pointer(first_event))[:length]
  events := make([]KeyEvent, length)
  for i := range c_events {
    events[i] = KeyEvent{
      Index     : uint16(c_events[i].index),
      Device    : uint16(c_events[i].device),
      Press_amt : float64(c_events[i].press_amt),
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

func CursorPos(window *Window) (int,int) {
  var x,y int
  C.CurrentMousePos(unsafe.Pointer(window.window), unsafe.Pointer(&x), unsafe.Pointer(&y));
  return x,y
}
