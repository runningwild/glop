package main

import (
  "json"
  "glop/gos"
  "glop/gin"
  "glop/gui"
  "glop/system"
  "glop/sprite"
  "game"
  "runtime"
  "fmt"
  "io/ioutil"
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

func LoadUnit(path string) (*game.UnitType, os.Error) {
  f,err := os.Open(path)
  if err != nil {
    return nil, err
  }
  defer f.Close()
  data,err := ioutil.ReadAll(f)
  if err != nil {
    return nil, err
  }
  var unit game.UnitType
  err = json.Unmarshal(data, &unit)
  if err != nil {
    return nil, err
  }
  return &unit, nil
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

  gui.MustLoadFontAs(filepath.Join(basedir, *font_path), "standard")

  factor := 1.0
  wdx := int(factor * float64(1024))
  wdy := int(factor * float64(768))

  sys.CreateWindow(0, 0, wdx, wdy)
  _,_,wdx,wdy = sys.GetWindowDims()
  ui := gui.Make(gin.In(), gui.Dims{wdx, wdy})
//  table := gui.MakeVerticalTable()
//  ui.AddChild(table)

  mappath := filepath.Join(os.Args[0], "..", "..", "maps", "bosworth")
  mappath = path.Clean(mappath)
  level,err := game.LoadLevel(mappath)
  if err != nil {
    panic(err.String())
  }
  err = level.SaveLevel("/Users/runningwild/code/go-glop/example/fudgecake.json")
  if err != nil {
    panic(err.String())
  }
  ui.AddChild(level.GetGui())
//  level.Terrain.Move(10,10)

//  table := anch.InstallWidget(&gui.VerticalTable{}, gui.Anchor{0,0, 0,0})

  n := 0

  // TODO: Would be better to only be vsynced, but apparently it can turn itself off
  // when the window disappears, so we need a safety net to slow it down if necessary
  sys.EnableVSync(true)
  ticker := time.Tick(10e6)


  // Load weapon files
  weaponpath := filepath.Join(basedir, "weapons", "guns.json")
  weapons,err := os.Open(weaponpath)
  if err != nil {
    panic(err.String())
  }
  err = game.LoadWeaponSpecs(weapons)
  if err != nil {
    panic(err.String())
  }

  seal,err := LoadUnit(filepath.Join(basedir, "units", "seal.json"))
  if err != nil { panic(err.String()) }
  rifleman,err := LoadUnit(filepath.Join(basedir, "units", "rifleman.json"))
  if err != nil { panic(err.String()) }

  bluepath := filepath.Join(basedir, "sprites", "blue")
  purplepath := filepath.Join(basedir, "sprites", "purple")
  var ents []*game.Entity
  guy,_ := sprite.LoadSprite(bluepath)
  ents = append(ents, level.AddEntity(*seal, 1, 2, 0.0075, guy))
  guy,_ = sprite.LoadSprite(bluepath)
  ents = append(ents, level.AddEntity(*seal, 2, 4, 0.0075, guy))
  guy,_ = sprite.LoadSprite(bluepath)
  ents = append(ents, level.AddEntity(*seal, 5, 1, 0.0075, guy))
  guy,_ = sprite.LoadSprite(purplepath)
  ents = append(ents, level.AddEntity(*rifleman, 25, 20, 0.0075, guy))
  guy,_ = sprite.LoadSprite(purplepath)
  ents = append(ents, level.AddEntity(*rifleman, 25, 29, 0.0075, guy))
  guy,_ = sprite.LoadSprite(purplepath)
  ents = append(ents, level.AddEntity(*rifleman, 25, 25, 0.0075, guy))

//  var texts []*gui.SingleLineText
//  for i := range ents {
//    texts = append(texts, gui.MakeSingleLineText("standard", "", 1, 1, 1, 1))
//    table.InstallWidget(texts[i], nil)
//  }
//  table.InstallWidget(gui.MakeTextEntry("standard", "", 1,1,1,1), nil)
  level.Setup()
  prev := time.Nanoseconds()

  for {
    n++
    next := time.Nanoseconds()
    dt := (next - prev) / 1000000
    prev = next

//    for i := range ents {
//      texts[i].SetText(fmt.Sprintf("%s: Health: %d, AP: %d", ents[i].Base.Name, ents[i].Health, ents[i].AP))
//    }
    sys.Think()
    ui.Draw()
    sys.SwapBuffers()
    <-ticker
    groups := sys.GetInputEvents()
    for _,group := range groups {
      if found,_ := group.FindEvent('q'); found {
        return
      }
    }

    if ui.FocusWidget() == nil {
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
      if gin.In().GetKey('e').FramePressCount() % 2 == 1 {
        level.ToggleEditor()
      }
      level.Terrain.Zoom(zoom * 0.0025)
    }
    level.Think(dt)
  }

  fmt.Printf("")
}
