package main

import "glop"
//import "gl"
import "runtime"

func main() {
  runtime.LockOSThread()
//  gl.Flush()
  window := glop.CreateWindow(10, 10, 500, 500)
//  gl.ClearColor(1.0, 0.0, 1.0, 1.0)
//  gl.Clear(0x00004000)
  glop.SwapBuffers(window)
//  glop.Foo(window)
  glop.Run()
}
