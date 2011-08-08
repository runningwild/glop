package main

import(
  "glop"
  "mingle"
  "runtime"
  "time"
)

func main() {
  runtime.LockOSThread()
  window := glop.CreateWindow(10, 10, 500, 500)
  gl.Flush()
  r := 0.0
  for {
    gl.ClearColor((gl.Clampf)(r), 0.0, 1.0, 1.0)
    gl.Clear(0x00004000)
    glop.SwapBuffers(window)
    println(r)
    glop.Think(window)
    r += 0.0101
    time.Sleep(1000*1000*100)
  }
}
