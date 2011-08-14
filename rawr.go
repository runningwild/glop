package main

import "glop/gos"
import "mingle"
import "runtime"
import "time"
import "fmt"

func main() {
  fmt.Printf("")
  runtime.LockOSThread()
  window := gos.CreateWindow(10, 10, 500, 500)
  gl.Flush()
  r := 0.0
  for {
    gl.ClearColor((gl.Clampf)(r), 0.0, 1.0, 1.0)
    gl.Clear(0x00004000)
    gos.SwapBuffers(window)
    gos.Think()
    v := gos.GetInputEvents()
    fmt.Printf("Length: %d\n", len(v))
    r += 0.0101
    time.Sleep(1000*1000*10)
  }
}
