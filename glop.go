package glop

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

// TODO: Can we just ignore this?
//func ShutDown() {
//  C.ShutDown()
//}

func CreateWindow() *Window {
  var window Window
  x := (*unsafe.Pointer)(unsafe.Pointer(&window.window))
  y := (*unsafe.Pointer)(unsafe.Pointer(&window.context))
  C.CreateWindow(x, y)
  return &window
}

func SwapBuffers(window *Window) {
  C.SwapBuffers(unsafe.Pointer(window.context))
}

func Think() {
  C.Think()
}

func CursorPos(window *Window) (int,int) {
  var x,y int
  C.CurrentMousePos(unsafe.Pointer(window.window), unsafe.Pointer(&x), unsafe.Pointer(&y));
  return x,y
}
