package gos

type Window uintptr

type System interface {
  // Call System.Think() every frame
  Think()

  CreateWindow(x,y,width,height int) Window
  SwapBuffers(window Window)
  GetInputEvents() []KeyEvent
  CursorPos(window Window) (int,int)
  WindowPos(window Window) (int,int)

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

