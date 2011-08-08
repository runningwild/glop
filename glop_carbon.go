package glop

// #include <glop.h>
import "C"

import "unsafe"

type Window struct {
  window  uintptr  // windowData**
}


func init() {
  C.Init()
}

func CreateWindow(x,y,dx,dy int) *Window {
  var window Window
  w := (*unsafe.Pointer)(unsafe.Pointer(&window.window))
  C.CreateWindow(w, 0, C.int(x), C.int(y), C.int(dx), C.int(dy))
  return &window
}

func SwapBuffers(window *Window) {
  C.SwapBuffers(unsafe.Pointer(window))
}

func Think(window *Window) {
  C.Think(unsafe.Pointer(window))
}

