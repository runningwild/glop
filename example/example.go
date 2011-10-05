package main

import (
  "glop/gos"
  "glop/gin"
  "glop/gui"
  "glop/system"
  "glop/sprite"
  "game"
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
  font_path *string = flag.String("font", "fonts/skia.ttf", "relative path of a font")
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

  basedir := filepath.Join(os.Args[0], "..", "..")

  err := gui.LoadFont("standard", filepath.Join(basedir, *font_path))
  if err != nil {
    panic(err.String())
  }

  factor := 0.875
  wdx := int(factor * float64(1024))
  wdy := int(factor * float64(768))

  sys.CreateWindow(10, 10, wdx, wdy)
  ui := gui.Make(gin.In(), wdx, wdy)
  anch := ui.Root.InstallWidget(gui.MakeAnchorBox(gui.Dims{wdx, wdy}), nil)
  h1 := gui.MakeSingleLineText("standard", "", 1,0.9,0.9,1)
  h2 := gui.MakeSingleLineText("standard", "", 1,0.9,0.9,1)
  mappath := filepath.Join(os.Args[0], "..", "..", "maps", "bosworth")
  mappath = path.Clean(mappath)
  level,err := game.LoadLevel(mappath)
  if err != nil {
    panic(err.String())
  }
  anch.InstallWidget(level.Terrain, gui.Anchor{0,0,0,0})

  table := anch.InstallWidget(&gui.VerticalTable{}, gui.Anchor{0,0, 0,0})

  frame_count_widget := gui.MakeSingleLineText("standard", "Frame", 1,0,1,1)
  table.InstallWidget(frame_count_widget, nil)
  table.InstallWidget(h1, nil)
  table.InstallWidget(h2, nil)
  n := 0
  sys.EnableVSync(true)
//  ticker := time.Tick(3e7)

  bluepath := filepath.Join(basedir, "sprites", "blue")
  purplepath := filepath.Join(basedir, "sprites", "purple")
  guy,err := sprite.LoadSprite(bluepath)
  if err != nil {
    panic(err.String())
  }
  guy2,err := sprite.LoadSprite(purplepath)
  if err != nil {
    panic(err.String())
  }

  seal := game.UnitType {
    Name : "Navy Seal",
    Health : 175,
    Move_cost : map[game.Terrain]int{
      game.Grass : 2,
      game.Dirt  : 2,
      game.Water : 6,
      game.Brush : 4,
    },
    AP : 30,
    Attack  : 150,
    Defense : 140,
    game.Weapons : []game.Weapon {
      &game.Bayonet {},
    },
  }

  rifleman := game.UnitType {
    Name : "Rifleman",
    Health : 90,
    Move_cost : map[game.Terrain]int{
      game.Grass : 1,
      game.Dirt  : 1,
      game.Water : 15,
      game.Brush : 1,
    },
    AP : 30,
    Attack  : 100,
    Defense : 80,
    game.Weapons : []game.Weapon {
      &game.Rifle {
        Range : 35,
        Power : 55,
      },
    },
  }
  ent := level.AddEntity(seal, 1, 2, 0.0075, guy)
  ent2 := level.AddEntity(rifleman, 25, 20, 0.0075, guy2)
  level.Setup()
  prev := time.Nanoseconds()

  for {
    n++
    next := time.Nanoseconds()
    dt := (next - prev) / 1000000
    prev = next

    frame_count_widget.SetText(fmt.Sprintf("               %d", n/10))
    h1.SetText(fmt.Sprintf("%s: Health: %d, AP: %d", ent.Base.Name, ent.Health, ent.AP))
    h2.SetText(fmt.Sprintf("%s: Health: %d, AP: %d", ent2.Base.Name, ent2.Health, ent2.AP))
    sys.Think()
    sys.SwapBuffers()
    groups := sys.GetInputEvents()
    for _,group := range groups {
      if found,_ := group.FindEvent('q'); found {
        return
      }
    }
    
    kw := gin.In().GetKey('w')
    ka := gin.In().GetKey('a')
    ks := gin.In().GetKey('s')
    kd := gin.In().GetKey('d')
    m_factor := 0.0075
    dx := m_factor * (kd.FramePressSum() - ka.FramePressSum())
    dy := m_factor * (kw.FramePressSum() - ks.FramePressSum())
    level.Terrain.Move(dx, dy)
    zoom := gin.In().GetKey('r').FramePressSum() - gin.In().GetKey('f').FramePressSum()
    if gin.In().GetKey('m').FramePressCount() > 0 {
      level.PrepMove()
    }
    if gin.In().GetKey('k').FramePressCount() > 0 {
      level.PrepAttack()
    }
    if gin.In().GetKey('o').FramePressCount() > 0 {
      level.Round()
    }
    level.Terrain.Zoom(zoom * 0.0025)
    level.Think(dt)
  }

  fmt.Printf("")
}
