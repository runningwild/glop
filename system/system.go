package system

type Window uintptr

type System interface {
  // Call after runtime.LockOSThread(), *NOT* in an init function
  Startup()

  // Call System.Think() every frame
  Think()

  CreateWindow(x,y,width,height int) Window
  // TODO: implement this:
  // DestroyWindow(Window)

  // Self-explanitory getters
  GetWindowPosition(window Window) (int,int)
  GetWindowSize(window Window) (int,int)

  SwapBuffers(window Window)
  GetInputEvents() []KeyEvent

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
  Think()

  // Create a window with the appropriate dimensions and bind an OpenGl contxt to it.
  // Currently glop only supports a single window, but this function could be called
  // more than once since a window could be destroyed so it can be recreated at different
  // dimensions or in full sreen mode.
  CreateWindow(x,y,width,height int) Window

  // TODO: implement this:
  // DestroyWindow(Window)

  // Self-explanitory getters
  GetWindowPosition(window Window) (int,int)
  GetWindowSize(window Window) (int,int)

  // Swap the OpenGl buffers on this window
  SwapBuffers(window Window)

  // Returns all of the events in the order that they happened since the last call to
  // this function.  The events do not have to be in order according to KeyEvent.Timestamp,
  // but they will be sorted according to this value.  There must always be a terminal
  // event that has a Timestamp equal to the largest value such that no future calls
  // to this function will yield an event with a Timestamp less than or equal to it.
  // TODO: Mention that KeyEvent.Index on the terminal event should probably be Dummy or something
  GetInputEvents() []KeyEvent

  // These probably shouldn't be here, probably always want to do the Think() approach
//  Run()
//  Quit()
}

type Mouse struct {
  X,Y,Dx,Dy int
}
type KeyEvent struct {
  // TODO: rename index to KeyId or something more appropriate
  Index     int
  Press_amt float64
  Mouse     Mouse
  Timestamp int64
  Num_lock  int
  Caps_lock int
}

