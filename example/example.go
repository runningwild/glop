package main

import (
  "glop/gos"
  "glop/gui"
  "glop/system"
  "glop/sprite"
  "runtime"
  "fmt"
  "os"
  "flag"
  "time"
  "path"
  "path/filepath"
)

var (
  sys system.System
  font_path *string = flag.String("font", "../../fonts/skia.ttf", "relative path of a font")
  sprite_path *string=flag.String("sprite", "../../sprites/test_sprite", "relative path of sprite")
)

func init() {
  sys = system.Make(gos.GetSystemInterface())
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

  factor := 0.875
  wdx := int(factor * float64(1024))
  wdy := int(factor * float64(768))

  sys.CreateWindow(10, 10, wdx, wdy)
  ui := gui.Make(sys.Input(), wdx, wdy)
  anch := ui.Root.InstallWidget(gui.MakeAnchorBox(gui.Dims{wdx - 50, wdy - 50}), nil)
  manch := anch.InstallWidget(gui.MakeAnchorBox(gui.Dims{wdx - 150, wdy - 150}), gui.Anchor{1,1,1,1})
  text_widget := gui.MakeSingleLineText("standard", "Funk Monkey 7$", 1,0.9,0.9,1)
  mappath := filepath.Join(os.Args[0], "..", "..", "maps", "bosworth")
  mappath = path.Clean(mappath)
  level,err := LoadLevel(mappath)
  if err != nil {
    panic(err.String())
  }
  manch.InstallWidget(level.terrain, gui.Anchor{0,0,0,0})

  table := anch.InstallWidget(&gui.VerticalTable{}, gui.Anchor{0,0, 0,0})

  frame_count_widget := gui.MakeSingleLineText("standard", "Frame", 1,0,1,1)
  table.InstallWidget(frame_count_widget, nil)
  table.InstallWidget(text_widget, nil)
  n := 0
  sys.EnableVSync(true)
//  ticker := time.Tick(3e7)

  spritepath := filepath.Join(os.Args[0], "..", "..", "sprites", "test_sprite")
  spritepath = path.Clean(spritepath)
  guy,err := sprite.LoadSprite(spritepath)
  if err != nil {
    panic(err.String())
  }
  guy2,err := sprite.LoadSprite(spritepath)
  if err != nil {
    panic(err.String())
  }
  ent := &entity{
    bx : 1,
    by : 2,
    move_speed : 0.01,
    s : guy,
  }
  ent2 := &entity{
    bx : 3,
    by : 5,
    move_speed : 0.01,
    s : guy2,
  }
  level.entities = append(level.entities, ent)
  level.entities = append(level.entities, ent2)
  prev := time.Nanoseconds()

  for {
    n++
    next := time.Nanoseconds()
    dt := (next - prev) / 1000000
    prev = next

    frame_count_widget.SetText(fmt.Sprintf("               %d", n/10))
    sys.Think()
    sys.SwapBuffers()
    groups := sys.GetInputEvents()
    for _,group := range groups {
      if found,_ := group.FindEvent('q'); found {
        return
      }
    }
    level.Think(dt)
    
    kw := sys.Input().GetKey('w')
    ka := sys.Input().GetKey('a')
    ks := sys.Input().GetKey('s')
    kd := sys.Input().GetKey('d')
    m_factor := 0.0075
    dx := m_factor * (kd.FramePressSum() - ka.FramePressSum())
    dy := m_factor * (kw.FramePressSum() - ks.FramePressSum())
    level.terrain.Move(dx, dy)
    zoom := sys.Input().GetKey('r').FramePressSum() - sys.Input().GetKey('f').FramePressSum()
    for i := range level.entities {
      level.entities[i].ap += sys.Input().GetKey('p').FramePressCount()
    }
    level.terrain.Zoom(zoom * 0.0025)
  }

  fmt.Printf("")
}
