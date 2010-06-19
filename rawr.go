package main

import "glop"
import "gl"

func main() {
  window := glop.CreateWindow()
  gl.ClearColor(1.0, 0.0, 1.0, 1.0)
  gl.Clear(0x00004000)
  glop.SwapBuffers(window)
//  glop.Foo(window)
  glop.Run()
}
