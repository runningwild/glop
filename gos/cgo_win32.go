package gos

// #include "include/glop.h"
import "C"

import (
  "fmt"
  "glop/system"
  "glop/gin"
  "unsafe"
)

type win32SystemObject struct {
  horizon int64
  window  uintptr
}
var (
  win32_system_object win32SystemObject
)

// Call after runtime.LockOSThread(), *NOT* in an init function
func (win32 *win32SystemObject) Startup() {
  C.GlopInit()
}

func GetSystemInterface() system.Os {
  return &win32_system_object
}

func (win32 *win32SystemObject) Run() {
//  C.Run()
}

func (win32 *win32SystemObject) Quit() {
//  C.Quit()
}

func (win32 *win32SystemObject) CreateWindow(x,y,width,height int) {
  title := []byte{'a','b','c',0}
  win32.window = uintptr(unsafe.Pointer(C.GlopCreateWindow(
    unsafe.Pointer(&title[0]),
    C.int(x), C.int(y), C.int(width), C.int(height), 0, 0, 0)))
}

func (win32 *win32SystemObject) SwapBuffers() {
  C.GlopSwapBuffers(unsafe.Pointer(win32.window))
}

func (win32 *win32SystemObject) Think() {
  C.GlopThink()
}

// TODO: Make sure that events are given in sorted order (by timestamp)
// TODO: Adjust timestamp on events so that the oldest timestamp is newer than the
//       newest timestemp from the events from the previous call to GetInputEvents
//       Actually that should be in system
func (win32 *win32SystemObject) GetInputEvents() ([]gin.OsEvent, int64) {
  var first_event *C.GlopKeyEvent
  cp := (*unsafe.Pointer)(unsafe.Pointer(&first_event))
  var length C.int
  var horizon C.longlong
  fmt.Printf("Here: %v\n", win32.window)
  C.GlopGetInputEvents(unsafe.Pointer(win32.window), cp, unsafe.Pointer(&length), unsafe.Pointer(&horizon))
  fmt.Printf("Events: %v\n", length)
  fmt.Printf("Horizon: %v\n", horizon)
  win32.horizon = int64(horizon)
  print("here\n")
  c_events := (*[1000]C.GlopKeyEvent)(unsafe.Pointer(first_event))[:length]
  print("here\n")
  events := make([]gin.OsEvent, length)
  print("here\n")
  for i := range c_events {
    events[i] = gin.OsEvent{
      KeyId     : gin.KeyId(c_events[i].index),
      Press_amt : float64(c_events[i].press_amt),
      Timestamp : int64(c_events[i].timestamp),
      Mouse : gin.Mouse{
        X : float64(c_events[i].cursor_x),
        Y : float64(c_events[i].cursor_y),
      },
    }
  }
  print("here\n")
  return events, win32.horizon
}

func (win32 *win32SystemObject) GetCursorPos() (x,y int) {
//  _x := unsafe.Pointer(&x)
//  _y := unsafe.Pointer(&y)
//  C.GetMousePos(_x, _y)
  return
}

func (win32 *win32SystemObject) GetWindowDims() (int,int,int,int) {
/*
  var x,y,dx,dy int
  osx_window := (*osxWindow)(unsafe.Pointer(window))
  _x := unsafe.Pointer(&x)
  _y := unsafe.Pointer(&y)
  _dx := unsafe.Pointer(&dx)
  _dy := unsafe.Pointer(&dy)
  C.GetWindowDims(unsafe.Pointer(osx_window.window), _x, _y, _dx, _dy)
*/
//  return x, y, dx, dy
  return 0,0,0,0
}
