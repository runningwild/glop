//target:glop/gos
package gos

// #include <glop.h>
import "C"

import "unsafe"

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

type KeyEvent struct {}
func GetInputEvents() []KeyEvent {
  var dummy_event C.KeyEvent
  cp := (*unsafe.Pointer)(unsafe.Pointer(&dummy_event))
  var length C.int
  C.GetInputEvents(cp, &length)
  return make([]KeyEvent, length)
}

func CursorPos(window *Window) (int,int) {
  var x,y int
  C.CurrentMousePos(unsafe.Pointer(window.window), unsafe.Pointer(&x), unsafe.Pointer(&y));
  return x,y
}
