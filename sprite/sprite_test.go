package sprite_test

import (
  "fmt"
  "github.com/runningwild/glop/sprite"
  . "github.com/orfjackal/gospec/src/gospec"
  "github.com/orfjackal/gospec/src/gospec"
)

func LoadSpriteSpec(c gospec.Context) {
  c.Specify("Sample sprite loads correctly", func() {
    s,err := sprite.LoadSprite("test_sprite")
    c.Expect(err, Equals, nil)
    for i := 0; i < 2000; i++ {
      s.Think(50)
      fmt.Printf("%s\n", s.Anim())
    }
    fmt.Printf("Commanding\n")
    s.Command("defend")
    s.Command("undamaged")
    s.Command("defend")
    s.Command("undamaged")
    for i := 0; i < 3000; i++ {
      s.Think(50)
    }
    s.Command("turn_right")
    s.Command("turn_right")
    s.Command("turn_right")
    s.Command("turn_right")
    s.Command("turn_right")
    s.Command("turn_right")
    s.Command("turn_left")
    s.Command("turn_left")
    s.Command("turn_right")
    s.Command("turn_right")
    s.Command("turn_right")
    s.Command("turn_left")
    s.Command("turn_left")
    s.Think(5000)
    c.Expect(s.Facing(), Equals, 1)
    for i := 0; i < 3000; i++ {
      s.Think(50)
    }
  })
}

func CommandNSpec(c gospec.Context) {
  c.Specify("Sample sprite loads correctly", func() {
    s,err := sprite.LoadSprite("test_sprite")
    c.Expect(err, Equals, nil)
    for i := 0; i < 2000; i++ {
      s.Think(50)
      fmt.Printf("%s\n", s.Anim())
    }
    s.CommandN([]string{
      "turn_right",
      "turn_right",
      "turn_right",
      "turn_right",
      "turn_right",
      "turn_right",
      "turn_left",
      "turn_left",
      "turn_right",
      "turn_right",
      "turn_right",
      "turn_left",
      "turn_left"})
    s.Think(5000)
    c.Expect(s.Facing(), Equals, 1)
    for i := 0; i < 3000; i++ {
      s.Think(50)
    }
  })
}
