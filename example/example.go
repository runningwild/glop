package main

import (
  "glop/gos"
  "glop/gin"
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
  "github.com/arbaal/mathgl"
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
  anch := ui.Root.InstallWidget(gui.MakeAnchorBox(gui.Dims{wdx - 50, wdy - 50}), nil)
  manch := anch.InstallWidget(gui.MakeAnchorBox(gui.Dims{wdx - 150, wdy - 150}), gui.Anchor{1,1,1,1})
  h1 := gui.MakeSingleLineText("standard", "", 1,0.9,0.9,1)
  h2 := gui.MakeSingleLineText("standard", "", 1,0.9,0.9,1)
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

  seal := UnitType {
    Name : "Navy Seal",
    Health : 175,
    Move_cost : map[Terrain]int{
      Grass : 2,
      Dirt  : 2,
      Water : 6,
      Brush : 4,
    },
    AP : 30,
    Attack  : 150,
    Defense : 140,
    Weapons : []Weapon {
      &Bayonet {},
    },
  }

  rifleman := UnitType {
    Name : "Rifleman",
    Health : 90,
    Move_cost : map[Terrain]int{
      Grass : 1,
      Dirt  : 1,
      Water : 15,
      Brush : 1,
    },
    AP : 30,
    Attack  : 100,
    Defense : 80,
    Weapons : []Weapon {
      &Rifle {
        Range : 35,
        Power : 55,
      },
    },
  }

  ent := &entity{
    UnitStats : UnitStats {
      Base : &seal,
    },
    pos : mathgl.Vec2{ 1, 2 },
    s : guy,
    level : level,
    CosmeticStats : CosmeticStats{
      Move_speed : 0.0075,
    },
  }
  level.entities = append(level.entities, ent)
  ent2 := &entity{
    UnitStats : UnitStats {
      Base : &rifleman,
    },
    pos : mathgl.Vec2{ 25, 20 },
    s : guy2,
    level : level,
    CosmeticStats : CosmeticStats{
      Move_speed : 0.0075,
    },
  }
  level.entities = append(level.entities, ent2)
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
    level.terrain.Move(dx, dy)
    zoom := gin.In().GetKey('r').FramePressSum() - gin.In().GetKey('f').FramePressSum()
    for i := range level.entities {
      level.entities[i].AP += 100 * gin.In().GetKey('p').FramePressCount()
    }
    if gin.In().GetKey('m').FramePressCount() > 0 {
      level.PrepMove()
    }
    if gin.In().GetKey('k').FramePressCount() > 0 {
      level.PrepAttack()
    }
    if gin.In().GetKey('o').FramePressCount() > 0 {
      level.Round()
    }
    level.terrain.Zoom(zoom * 0.0025)
    level.Think(dt)
  }

  fmt.Printf("")
}
