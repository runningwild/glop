package gin_test

import (
  . "gospec"
  "gospec"
  "glop/gin"
  "glop/system"
)

type mockSystem struct {}
func (ms *mockSystem) Think() {
}
func (ms *mockSystem) CreateWindow(x,y,width,height int) system.Window {
  return 0
}
func (ms *mockSystem) SwapBuffers(window system.Window) {
}
func (ms *mockSystem) GetInputEvents() []system.KeyEvent {
  return []system.KeyEvent{}
}
func (ms *mockSystem) CursorPos(window system.Window) (int,int) {
  return 0,0
}
func (ms *mockSystem) WindowPos(window system.Window) (int,int) {
  return 0,0
}

func BasicInputSpec(c gospec.Context) {
  ms := &mockSystem{}
  c.Specify("Empty spec, just making sure things are connected.", func() {
    gin.SetSystemObject(ms)
    c.Expect(1, Equals, 1)
//    VecExpect(c, p.Seg(3)[1], Equals, s3[1])
  })
}

