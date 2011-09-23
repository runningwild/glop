package main

import (
  "glop/gos"
  "glop/gui"
  "glop/system"
  "runtime"
  "time"
  "fmt"
  "os"
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
  anch := ui.Root.InstallWidget(gui.MakeAnchorBox(gui.Dims{700, 500}), nil)
  manch := anch.InstallWidget(gui.MakeAnchorBox(gui.Dims{400,300}), gui.Anchor{1,1,1,1})
  text_widget := gui.MakeSingleLineText("standard", "Funk Monkey 7$", 1,0.9,0.9,1)

  terrain,err := gui.MakeTerrain("../../maps/chess.png", int(1000.0/8))
  if err != nil {
    panic(err.String())
  } else {
    manch.InstallWidget(terrain, gui.Anchor{0,0,0,0})
  }

  table := anch.InstallWidget(&gui.VerticalTable{}, gui.Anchor{0,0, 0,0})

  table.InstallWidget(text_widget, gui.Anchor{0,1,0,1})
  frame_count_widget := gui.MakeSingleLineText("standard", "Frame", 0,0,1,1)
  table.InstallWidget(frame_count_widget, gui.Anchor{1,1,1,1})
  n := 0
  for {
    n++
    terrain.HighlightBlockAtCursor(sys.GetCursorPos())
    frame_count_widget.SetText(fmt.Sprintf("%d", n))
    sys.SwapBuffers()
    <-ticker
    sys.Think()
    groups := sys.GetInputEvents()
    fmt.Printf("Num groups: %d\n", len(groups))
    for _,group := range groups {
      if found,_ := group.FindEvent('q'); found {
        return
      }
    }
  }

  fmt.Printf("")
}
