package gin_test

import (
  . "gospec"
  "gospec"
  "glop/gin"
  "glop/system"
  "strings"
)

type mockSystem struct {
  next_events []system.KeyEvent
}
func (ms *mockSystem) Think() {
}
func (ms *mockSystem) Startup() {
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
func (ms *mockSystem) GetWindowPosition(window system.Window) (int,int) {
  return 0,0
}
func (ms *mockSystem) GetWindowSize(window system.Window) (int,int) {
  return 0,0
}
func (ms *mockSystem) injectEvent(index gin.KeyId, amt float64, timestamp int64) {
  ms.next_events = append(ms.next_events,
    system.KeyEvent{
      Index : int(index),
      Press_amt : amt,
      Timestamp : timestamp,
    },
  )
}

func NaturalKeySpec(c gospec.Context) {
  ms := &mockSystem{}
  input := gin.MakeInput(ms)
  c.Specify("Single key press or release per frame sets basic keyState values properly.", func() {
    keya := input.GetKey('a')
    keyb := input.GetKey('b')

    ms.injectEvent('a', 1, 5)
    input.Think(10, false)
    c.Expect(keya.FramePressCount(),   Equals, 1)
    c.Expect(keya.FrameReleaseCount(), Equals, 0)
    c.Expect(keyb.FramePressCount(),   Equals, 0)
    c.Expect(keyb.FrameReleaseCount(), Equals, 0)

    ms.injectEvent('b', 1, 15)
    input.Think(20, false)
    c.Expect(keya.FramePressCount(),   Equals, 0)
    c.Expect(keya.FrameReleaseCount(), Equals, 0)
    c.Expect(keyb.FramePressCount(),   Equals, 1)
    c.Expect(keyb.FrameReleaseCount(), Equals, 0)

    ms.injectEvent('a', 0, 25)
    input.Think(30, false)
    c.Expect(keya.FramePressCount(),   Equals, 0)
    c.Expect(keya.FrameReleaseCount(), Equals, 1)
    c.Expect(keyb.FramePressCount(),   Equals, 0)
    c.Expect(keyb.FrameReleaseCount(), Equals, 0)
  })

  c.Specify("Multiple key presses in a single frame work.", func() {
    keya := input.GetKey('a')
    keyb := input.GetKey('b')

    ms.injectEvent('a', 1, 4)
    ms.injectEvent('a', 0, 5)
    ms.injectEvent('a', 1, 6)
    ms.injectEvent('b', 1, 7)
    ms.injectEvent('a', 0, 8)
    ms.injectEvent('a', 1, 9)
    input.Think(10, false)
    c.Expect(keya.FramePressCount(),   Equals, 3)
    c.Expect(keya.FrameReleaseCount(), Equals, 2)
    c.Expect(keyb.FramePressCount(),   Equals, 1)
    c.Expect(keyb.FrameReleaseCount(), Equals, 0)
  })

  c.Specify("Key.FramePressSum() works.", func() {
    keya := input.GetKey('a')
    ms.injectEvent('a', 1, 3)
    input.Think(10, false)
    ms.injectEvent('a', 0, 14)
    ms.injectEvent('a', 1, 16)
    input.Think(20, false)
    c.Expect(keya.FramePressSum(),   Equals, 8.0)

    keyb := input.GetKey('b')
    ms.injectEvent('b', 1, 22)
    ms.injectEvent('b', 0, 24)
    input.Think(30, false)
    c.Expect(keyb.FramePressSum(),   Equals, 2.0)

    ms.injectEvent('b', 1, 35)
    input.Think(40, false)
    c.Expect(keyb.FramePressSum(),   Equals, 5.0)
  })
}

func DerivedKeySpec(c gospec.Context) {
  ms := &mockSystem{}
  input := gin.MakeInput(ms)
  ABc_binding := input.MakeBinding('a', []gin.KeyId{'b', 'c'}, []bool{true, false})
  Ef_binding := input.MakeBinding('e', []gin.KeyId{'f'}, []bool{false})
  ABc_Ef := input.BindDerivedKey("ABc_Ef", ABc_binding, Ef_binding)
  // ABc_Ef should be down if either ab and not c, or e and not f (or both)
  // That is to say the following (and no others) should all trigger it:
  // A B c e f
  // A B c e F
  // A B c E f
  // A B c E F
  // a b c E f
  // a b C E f
  // a B c E f
  // a B C E f
  // A b c E f
  // A b C E f
  // A B C E f

  c.Specify("Derived key presses happen only when a primary key is pressed after all modifiers are set.", func() {

    // Test that first binding can trigger a press
    ms.injectEvent('b', 1, 1)
    ms.injectEvent('a', 1, 1)
    input.Think(10, false)
    c.Expect(ABc_Ef.FramePressAmt(), Equals, 1.0)
    c.Expect(ABc_Ef.IsDown(), Equals, true)
    c.Expect(ABc_Ef.FramePressCount(), Equals, 1)

    c.Specify("Release happens once primary key is released", func() {
      ms.injectEvent('a', 0, 11)
      input.Think(20, false)
      c.Expect(ABc_Ef.IsDown(), Equals, false)
      c.Expect(ABc_Ef.FramePressCount(), Equals, 0)
      c.Expect(ABc_Ef.FrameReleaseCount(), Equals, 1)
    })

    c.Specify("Release happens when a modifier that should be pressed is released", func() {
      ms.injectEvent('b', 0, 11)
      input.Think(20, false)
      c.Expect(ABc_Ef.IsDown(), Equals, false)
      c.Expect(ABc_Ef.FramePressCount(), Equals, 0)
      c.Expect(ABc_Ef.FrameReleaseCount(), Equals, 1)
    })

    c.Specify("Release happens when a modifier that should be released is pressed", func() {
      ms.injectEvent('c', 1, 11)
      input.Think(20, false)
      c.Expect(ABc_Ef.IsDown(), Equals, false)
      c.Expect(ABc_Ef.FramePressCount(), Equals, 0)
      c.Expect(ABc_Ef.FrameReleaseCount(), Equals, 1)
    })

    c.Specify("Pressing a second binding should not generate another press on the derived key", func() {
      ms.injectEvent('e', 1, 11)
      input.Think(20, false)
      c.Expect(ABc_Ef.IsDown(), Equals, true)
      c.Expect(ABc_Ef.FramePressCount(), Equals, 0)
      c.Expect(ABc_Ef.FrameReleaseCount(), Equals, 0)
    })

    // Reset keys
    ms.injectEvent('a', 0, 21)
    ms.injectEvent('b', 0, 21)
    ms.injectEvent('c', 0, 21)
    ms.injectEvent('e', 0, 21)
    input.Think(30, false)

    // Test that second binding can trigger a press
    ms.injectEvent('e', 1, 31)
    input.Think(40, false)
    c.Expect(ABc_Ef.FramePressAmt(), Equals, 1.0)
    c.Expect(ABc_Ef.IsDown(), Equals, true)
    c.Expect(ABc_Ef.FramePressCount(), Equals, 1)

    // Reset keys
    ms.injectEvent('e', 0, 41)
    input.Think(50, false)

    // Test that first binding doesn't trigger a press if modifiers aren't set first
    ms.injectEvent('a', 1, 51)
    ms.injectEvent('b', 1, 51)
    input.Think(60, false)
    c.Expect(ABc_Ef.IsDown(), Equals, false)
    c.Expect(ABc_Ef.FramePressCount(), Equals, 0)
  })
}

func NestedDerivedKeySpec(c gospec.Context) {
  ms := &mockSystem{}
  input := gin.MakeInput(ms)
  AB_binding := input.MakeBinding('a', []gin.KeyId{'b'}, []bool{true})
  AB := input.BindDerivedKey("AB", AB_binding)
  AB_C_binding := input.MakeBinding('c', []gin.KeyId{AB.Id()}, []bool{true})
  AB_C := input.BindDerivedKey("AB_C", AB_C_binding)

  check := func(order string) {
    input.Think(10, false)
    if strings.Index(order, "b") < strings.Index(order, "a") {
      c.Expect(AB.IsDown(), Equals, true)
      c.Expect(AB.FramePressCount(), Equals, 1)
    } else {
      c.Expect(AB.IsDown(), Equals, false)
      c.Expect(AB.FramePressCount(), Equals, 0)
    }
    if order == "bac" {
      c.Expect(AB_C.IsDown(), Equals, true)
      c.Expect(AB_C.FramePressCount(), Equals, 1)
    } else {
      c.Expect(AB_C.IsDown(), Equals, false)
      c.Expect(AB_C.FramePressCount(), Equals, 0)
    }
  }

  c.Specify("Nested derived keys work like normal derived keys.", func() {
    c.Specify("Testing order 'bac'.", func() {
      ms.injectEvent('b', 1, 1)
      ms.injectEvent('a', 1, 1)
      ms.injectEvent('c', 1, 1)
      check("bac")
    })
    c.Specify("Testing order 'abc'.", func() {
      ms.injectEvent('a', 1, 1)
      ms.injectEvent('b', 1, 1)
      ms.injectEvent('c', 1, 1)
      check("abc")
    })
    c.Specify("Testing order 'acb'.", func() {
      ms.injectEvent('a', 1, 1)
      ms.injectEvent('c', 1, 1)
      ms.injectEvent('b', 1, 1)
      check("acb")
    })
    c.Specify("Testing order 'bca'.", func() {
      ms.injectEvent('b', 1, 1)
      ms.injectEvent('c', 1, 1)
      ms.injectEvent('a', 1, 1)
      check("bca")
    })
    c.Specify("Testing order 'cab'.", func() {
      ms.injectEvent('c', 1, 1)
      ms.injectEvent('a', 1, 1)
      ms.injectEvent('b', 1, 1)
      check("cab")
    })
    c.Specify("Testing order 'cba'.", func() {
      ms.injectEvent('c', 1, 1)
      ms.injectEvent('b', 1, 1)
      ms.injectEvent('a', 1, 1)
      check("cba")
    })
  })
}

