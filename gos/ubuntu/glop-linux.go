package glop

// #include <glop.h>
import "C"

import "unsafe"

type Window struct {
  window  uintptr  // OsWindowData*
}


func init() {
  C.Init()
}

func Run() {
  C.Run()
}

// TODO: Can we just ignore this?
//func ShutDown() {
//  C.ShutDown()
//}

func CreateWindow(x,y,width,height int) *Window {
  var window Window
  w := (*unsafe.Pointer)(unsafe.Pointer(&window.window))
  C.CreateWindow(w, C.int(x), C.int(y), C.int(width), C.int(height))
  return &window
}

func SwapBuffers(window *Window) {
  C.SwapBuffers(unsafe.Pointer(window))
}

func Think() {
  C.Think()
}

func CursorPos(window *Window) (int,int) {
  var x,y int
  C.CurrentMousePos(unsafe.Pointer(window.window), unsafe.Pointer(&x), unsafe.Pointer(&y));
  return x,y
}

