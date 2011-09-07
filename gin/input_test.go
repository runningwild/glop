package gin_test

import (
  . "gospec"
  "gospec"
  "glop/gin"
  "glop/system"
)

type mockSystem struct {
  next_events []system.KeyEvent
}
func (ms *mockSystem) Think() {
}
func (ms *mockSystem) CreateWindow(x,y,width,height int) system.Window {
  return 0
}
func (ms *mockSystem) SwapBuffers(window system.Window) {
}
func (ms *mockSystem) GetInputEvents() []system.KeyEvent {
  ret := ms.next_events
  ms.next_events = []system.KeyEvent{}
  return ret
}
func (ms *mockSystem) CursorPos(window system.Window) (int,int) {
  return 0,0
}
func (ms *mockSystem) WindowPos(window system.Window) (int,int) {
  return 0,0
}

func BasicInputSpec(c gospec.Context) {
  ms := &mockSystem{}
  input := gin.MakeInput(ms)
  c.Specify("Single key press or release per frame sets basic keyState values properly.", func() {
    keya := input.GetKey('a')
    keyb := input.GetKey('b')

    ms.next_events = []system.KeyEvent{
      system.KeyEvent{
        Index     : 'a',
        Press_amt :  1,
        Timestamp :  5,
      },
    }
    input.Think(10, false)
    c.Expect(keya.FramePressCount(),   Equals, 1)
    c.Expect(keya.FrameReleaseCount(), Equals, 0)
    c.Expect(keyb.FramePressCount(),   Equals, 0)
    c.Expect(keyb.FrameReleaseCount(), Equals, 0)

    ms.next_events = []system.KeyEvent{
      system.KeyEvent{
        Index     : 'b',
        Press_amt :  1,
        Timestamp : 15,
      },
    }
    input.Think(20, false)
    c.Expect(keya.FramePressCount(),   Equals, 0)
    c.Expect(keya.FrameReleaseCount(), Equals, 0)
    c.Expect(keyb.FramePressCount(),   Equals, 1)
    c.Expect(keyb.FrameReleaseCount(), Equals, 0)

    ms.next_events = []system.KeyEvent{
      system.KeyEvent{
        Index     : 'a',
        Press_amt :  0,
        Timestamp : 25,
      },
    }
    input.Think(30, false)
    c.Expect(keya.FramePressCount(),   Equals, 0)
    c.Expect(keya.FrameReleaseCount(), Equals, 1)
    c.Expect(keyb.FramePressCount(),   Equals, 0)
    c.Expect(keyb.FrameReleaseCount(), Equals, 0)
  })
}

