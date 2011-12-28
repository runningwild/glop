package system

import (
  "glop/gin"
)

type System interface {
  // Call after runtime.LockOSThread(), *NOT* in an init function
  Startup()

  // Call System.Think() every frame
  Think() bool

  CreateWindow(x, y, width, height int)
  // TODO: implement this:
  // DestroyWindow(Window)

  // Gets the cursor position in window coordinates with the cursor at the bottom left
  // corner of the window
  GetCursorPos() (x, y int)

  // Hides/Unhides the cursor.  A hidden cursor is invisible and its position is
  // locked.  It should still generate mouse move events.
  HideCursor(bool)

  GetWindowDims() (x, y, dx, dy int)

  SwapBuffers()
  GetInputEvents() []gin.EventGroup

  EnableVSync(bool)

  // These probably shouldn't be here, probably always want to do the Think() approach
  //  Run()
  //  Quit()
}

// This is the interface implemented by any operating system that supports
// glop.  The glop/gos package for that OS should export a function called
// GetSystemInterface() which takes no parameters and returns an object that
// implements the system.Os interface.
type Os interface {
  // This is properly called after runtime.LockOSThread(), not in an init function
  Startup()

  // Think() is called on a regular basis and always from main thread.
  Think() bool

  // Create a window with the appropriate dimensions and bind an OpenGl contxt to it.
  // Currently glop only supports a single window, but this function could be called
  // more than once since a window could be destroyed so it can be recreated at different
  // dimensions or in full sreen mode.
  CreateWindow(x, y, width, height int)

  // TODO: implement this:
  // DestroyWindow(Window)

  // Gets the cursor position in window coordinates with the cursor at the bottom left
  // corner of the window
  GetCursorPos() (x, y int)

  // Hides/Unhides the cursor.  A hidden cursor is invisible and its position is
  // locked.  It should still generate mouse move events.
  HideCursor(bool)

  GetWindowDims() (x, y, dx, dy int)

  // Swap the OpenGl buffers on this window
  SwapBuffers()

  // Returns all of the events in the order that they happened since the last call to
  // this function.  The events do not have to be in order according to KeyEvent.Timestamp,
  // but they will be sorted according to this value.  The timestamp returned is the event
  // horizon, no future events will have a timestamp less than or equal to it.
  GetInputEvents() ([]gin.OsEvent, int64)

  EnableVSync(bool)

  // These probably shouldn't be here, probably always want to do the Think() approach
  //  Run()
  //  Quit()
}

type sysObj struct {
  os       Os
  events   []gin.EventGroup
  start_ms int64
}

func Make(os Os) System {
  return &sysObj{
    os: os,
  }
}
func (sys *sysObj) Startup() {
  sys.os.Startup()
  _, sys.start_ms = sys.os.GetInputEvents()
}
func (sys *sysObj) Think() bool {
  if !sys.os.Think() {
    return false
  }
  events, horizon := sys.os.GetInputEvents()
  for i := range events {
    events[i].Timestamp -= sys.start_ms
  }
  sys.events = gin.In().Think(horizon-sys.start_ms, false, events)
  return true
}
func (sys *sysObj) CreateWindow(x, y, width, height int) {
  sys.os.CreateWindow(x, y, width, height)
}
func (sys *sysObj) GetCursorPos() (int, int) {
  return sys.os.GetCursorPos()
}
func (sys *sysObj) HideCursor(hide bool) {
  sys.os.HideCursor(hide)
}
func (sys *sysObj) GetWindowDims() (int, int, int, int) {
  return sys.os.GetWindowDims()
}
func (sys *sysObj) SwapBuffers() {
  sys.os.SwapBuffers()
}
func (sys *sysObj) GetInputEvents() []gin.EventGroup {
  return sys.events
}
func (sys *sysObj) EnableVSync(enable bool) {
  sys.os.EnableVSync(enable)
}
