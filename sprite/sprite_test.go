package sprite_test

import (
  "github.com/runningwild/glop/sprite"
  . "github.com/orfjackal/gospec/src/gospec"
  "github.com/orfjackal/gospec/src/gospec"
)

func LoadSpriteSpec(c gospec.Context) {
  c.Specify("Sample sprite loads correctly", func() {
    s, err := sprite.LoadSprite("test_sprite")
    c.Expect(err, Equals, nil)
    for i := 0; i < 2000; i++ {
      s.Think(50)
    }
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
    // s.Think(5000)
    for i := 0; i < 300; i++ {
      s.Think(50)
    }
    c.Expect(s.Facing(), Equals, 1)
  })
}

func CommandNSpec(c gospec.Context) {
  c.Specify("Sample sprite loads correctly", func() {
    s, err := sprite.LoadSprite("test_sprite")
    c.Expect(err, Equals, nil)
    for i := 0; i < 2000; i++ {
      s.Think(50)
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

func SyncSpec(c gospec.Context) {
  c.Specify("Sample sprite loads correctly", func() {
    s1, err := sprite.LoadSprite("test_sprite")
    c.Expect(err, Equals, nil)
    s2, err := sprite.LoadSprite("test_sprite")
    c.Expect(err, Equals, nil)
    sprite.CommandSync([]*sprite.Sprite{s1, s2}, [][]string{[]string{"melee"}, []string{"defend", "damaged"}}, "hit")
    hit := false
    for i := 0; i < 20; i++ {
      s1.Think(50)
      s2.Think(50)
      if s1.Anim() == "melee_01" && s2.Anim() == "damaged_01" {
        hit = true
      }
    }
    c.Expect(hit, Equals, true)
  })
}
