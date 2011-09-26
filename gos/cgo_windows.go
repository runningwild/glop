package gos

// #include "windows/include/glop.h"
import "C"

import (
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
  title := []byte("Mob Rules")
  title = append(title, 0)
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
  C.GlopGetInputEvents(unsafe.Pointer(win32.window), cp, unsafe.Pointer(&length), unsafe.Pointer(&horizon))
  win32.horizon = int64(horizon)
  c_events := (*[1000]C.GlopKeyEvent)(unsafe.Pointer(first_event))[:length]
  events := make([]gin.OsEvent, length)
  for i := range c_events {
    wx,wy := osx.rawCursorToWindowCoords(int(c_events[i].cursor_x), int(c_events[i].cursor_y))
    events[i] = gin.OsEvent{
      KeyId     : gin.KeyId(c_events[i].index),
      Press_amt : float64(c_events[i].press_amt),
      Timestamp : int64(c_events[i].timestamp),
      X : wx,
      Y : wy,
    }
  }
  return events, win32.horizon
}

func (win32 *win32SystemObject) rawCursorToWindowCoords(x,y int) (int,int) {
  wx,wy,_,wdy := win32.GetWindowDims()
  return x - wx, wy + wdy - y
}

func (win32 *win32SystemObject) GetCursorPos() (int,int) {
  var x,y C.int
  C.GlopGetMousePosition(&x, &y)
  return win32.rawCursorToWindowCoords(int(x), int(y))
}

func (win32 *win32SystemObject) GetWindowDims() (int,int,int,int) {
  var x,y,dx,dy C.int
  C.GlopGetWindowDims(unsafe.Pointer(win32.window), &x, &y, &dx, &dy)
  return int(x), int(y), int(dx), int(dy)
}
