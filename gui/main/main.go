package main

import (
  "glop/gos"
  "glop/gin"
  "glop/system"
  "runtime"
)

var (
  sys system.System
  quit chan bool
)

func init() {
  runtime.LockOSThread()
  quit = make(chan bool)
}

func main() {
  go realmain()
  <-quit
}

func realmain() {
  sys = system.Make(gos.GetSystemInterface())
  sys.Startup()
  sys.CreateWindow(10, 10, 800, 600)
  sys.EnableVSync(true)

  for gin.In().GetKey('q').FramePressCount() == 0 {
    sys.SwapBuffers()
    sys.Think()
  }
  quit <- true
}
