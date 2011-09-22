package main

import (
  "glop/gos"
  "glop/gui"
  "glop/gin"
  "glop/system"
  "glop/sprite"
  "runtime"
  "time"
  "fmt"
  "os"
  "path"
  "path/filepath"
  "flag"
)

var (
  sys system.System
  font_path *string = flag.String("font", "../../fonts/skia.ttf", "relative path of a font")
  sprite_path *string=flag.String("sprite", "../../sprites/test_sprite", "relative path of sprite")
)

func init() {
  sys = system.Make(gos.GetSystemInterface())
}


type Foo struct {
  *gui.BoxWidget
}
func (f *Foo) Think(_ int64, has_focus bool, previous gui.Region, _ map[gui.Widget]gui.Dims) (bool, gui.Dims) {
  cx,cy := sys.GetCursorPos()
  cursor := gui.Point{ X : cx, Y : cy }
  if cursor.Inside(previous) {
    f.Dims.Dx += 5
    f.G = 0
    f.B = 0
  } else {
    f.G = 1
    f.B = 1
  }
  return f.BoxWidget.Think(0, false, previous, nil)
}

func main() {
  runtime.LockOSThread()

  // Running a binary via osx's package mechanism will add a flag that begins with
  // '-psn' so we have to find it and pretend like we were expecting it so that go
  // doesn't asplode because of an unexpected flag.
  for _,arg := range os.Args {
    if len(arg) >= 4 && arg[0:4] == "-psn" {
      flag.Bool(arg[1:], false, "HERE JUST TO APPEASE GO'S FLAG PACKAGE")
    }
  }
  flag.Parse()
  sys.Startup()

  err := gui.LoadFont("standard", *font_path)
  if err != nil {
    panic(err.String())
  }

  sys.CreateWindow(10, 10, 768, 576)
  ticker := time.Tick(5e7)
  ui := gui.Make(sys.Input(), 768, 576)
  anch := ui.Root.InstallWidget(gui.MakeAnchorBox(gui.Dims{768, 576}), nil)
  anch.InstallWidget(gui.MakeBoxWidget(250,250,0,0,1,1), gui.Anchor{0.5, 0.5, 0.5, 0.5})
  anch.InstallWidget(
      &Foo{gui.BoxWidget : gui.MakeBoxWidget(100, 100, 1, 1, 1, 1)},
      gui.Anchor{ Bx:0.7, By:1, Wx:0.5, Wy:1})
  anch.InstallWidget(
      &Foo{gui.BoxWidget : gui.MakeBoxWidget(100, 100, 1, 1, 1, 1)},
      gui.Anchor{ Bx:0, By:0.5, Wx:0, Wy:0.5})
  anch.InstallWidget(
      &Foo{gui.BoxWidget : gui.MakeBoxWidget(100, 100, 1, 1, 1, 1)},
      gui.Anchor{ Bx:0.2, By:0.2, Wx:1, Wy:1})
  text_widget := gui.MakeSingleLineText("standard", "Funk Monkey 7$")
  anch.InstallWidget(gui.MakeBoxWidget(450,50,0,1,0,1), gui.Anchor{0,1, 0,1})
  anch.InstallWidget(text_widget, gui.Anchor{0,1,0,1})
  frame_count_widget := gui.MakeSingleLineText("standard", "Frame")
  anch.InstallWidget(gui.MakeBoxWidget(250,50,0,1,0,1), gui.Anchor{1,1, 1,1})
  anch.InstallWidget(frame_count_widget, gui.Anchor{1,1,1,1})

  n := 0
  spritepath := filepath.Join(os.Args[0], *sprite_path)
  spritepath = path.Clean(spritepath)
  s,err := sprite.LoadSprite(spritepath)
  if err != nil {
    panic(err.String())
  }
  s.Stats()

  key_bindings := make(map[byte]string)
  key_bindings['a'] = "defend"
  key_bindings['s'] = "undamaged"
  key_bindings['d'] = "damaged"
  key_bindings['f'] = "killed"
  key_bindings['z'] = "ranged"
  key_bindings['x'] = "melee"
  key_bindings['c'] = "move"
  key_bindings['v'] = "stop"
  key_bindings['o'] = "turn_left"
  key_bindings['p'] = "turn_right"

  for {
    n++
    frame_count_widget.SetText(fmt.Sprintf("%d", n))
    s.RenderToQuad()
    s.Think(50)
    sys.SwapBuffers()
    <-ticker
    text_widget.SetText(s.CurState())
    sys.Think()
    groups := sys.GetInputEvents()
    fmt.Printf("Num groups: %d\n", len(groups))
    for _,group := range groups {
      for key,cmd := range key_bindings {
        if found,event := group.FindEvent(gin.KeyId(key)); found && event.Type == gin.Press {
          s.Command(cmd)
        }
      }
      if found,_ := group.FindEvent('q'); found {
        return
      }
    }
  }

  fmt.Printf("")
}
