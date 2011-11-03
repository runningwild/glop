package main

import (
  "json"
  "glop/gos"
  "glop/gin"
  "glop/gui"
  "glop/system"
  "game"
  "runtime"
  "runtime/pprof"
  "fmt"
  "io/ioutil"
  "os"
  "flag"
  "time"
  "path/filepath"
  "strings"
)

var (
  sys       system.System
  font_path *string = flag.String("font", "fonts/skia.ttf", "relative path of a font")
  quit      chan bool
)

func LoadUnit(path string) (*game.UnitType, error) {
  f, err := os.Open(path)
  if err != nil {
    return nil, err
  }
  defer f.Close()
  data, err := ioutil.ReadAll(f)
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

func init() {
  sys = system.Make(gos.GetSystemInterface())
  quit = make(chan bool)
  go actualMain()
}

func actualMain() {
  runtime.LockOSThread()
  defer func() {
    quit <- true
  }()
  // Running a binary via osx's package mechanism will add a flag that begins with
  // '-psn' so we have to find it and pretend like we were expecting it so that go
  // doesn't asplode because of an unexpected flag.
  for _, arg := range os.Args {
    if len(arg) >= 4 && arg[0:4] == "-psn" {
      flag.Bool(arg[1:], false, "HERE JUST TO APPEASE GO'S FLAG PACKAGE")
    }
  }
  flag.Parse()
  sys.Startup()

  basedir := filepath.Join(os.Args[0], "..", "..")

  // TODO: Loading weapon specs should be done automatically - it just needs the datadir
  // Load weapon files
  dir, err := os.Open(filepath.Join(basedir, "weapons"))
  if err != nil {
    panic(err.Error())
  }
  names, err := dir.Readdir(0)
  if err != nil {
    panic(err.Error())
  }
  for _,name := range names {
    weapons, err := os.Open(filepath.Join(basedir, "weapons", name.Name))
    if err != nil {
      panic(err.Error())
    }
    err = game.LoadWeaponSpecs(weapons)
    if err != nil {
      panic(err.Error())
    }
  }

  gui.MustLoadFontAs(filepath.Join(basedir, *font_path), "standard")

  factor := 1.0
  wdx := int(factor * float64(1024))
  wdy := int(factor * float64(768))

  sys.CreateWindow(0, 0, wdx, wdy)
  _, _, wdx, wdy = sys.GetWindowDims()
  ui := gui.Make(gin.In(), gui.Dims{wdx, wdy})
  //  table := gui.MakeVerticalTable()
  //  ui.AddChild(table)

  level, err := game.LoadLevel(basedir, "bosworth.json")
  if err != nil {
    panic(err.Error())
  }
  ui.AddChild(level.GetGui())
  //  level.Terrain.Move(10,10)

  //  table := anch.InstallWidget(&gui.VerticalTable{}, gui.Anchor{0,0, 0,0})

  n := 0

  // TODO: Would be better to only be vsynced, but apparently it can turn itself off
  // when the window disappears, so we need a safety net to slow it down if necessary
  sys.EnableVSync(true)
  ticker := time.Tick(1e7)

  level.Setup()
  prev := time.Nanoseconds()

  profiling := false
  var load_widget gui.Widget
  for {
    n++
    next := time.Nanoseconds()
    dt := (next - prev) / 1000000
    prev = next

    sys.Think()
    level.Think(dt)
    ui.Draw()
    sys.SwapBuffers()
    <-ticker
    if gin.In().GetKey('q').FramePressCount() > 0 {
      return
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
      if gin.In().GetKey('e').FramePressCount()%2 == 1 {
        level.ToggleEditor()
      }
      if gin.In().GetKey('p').FramePressCount() > 0 {
        if !profiling {
          f, err := os.Create(filepath.Join(basedir, "profiles", "profile.prof"))
          if err != nil {
            fmt.Printf("Failed to write profile: %s\n", err.Error())
          }
          pprof.StartCPUProfile(f)
          profiling = true
        } else {
          pprof.StopCPUProfile()
          profiling = false
        }
      }
      if gin.In().GetKey('l').FramePressCount() > 0 {
        if load_widget != nil {
          ui.RemoveChild(load_widget)
          load_widget = nil
        } else {
          table := gui.MakeVerticalTable()
          dir, err := os.Open(filepath.Join(basedir, "maps"))
          if err != nil {
            panic(err.Error())
          }
          names, err := dir.Readdir(0)
          if err != nil {
            panic(err.Error())
          }
          for _, name := range names {
            if !strings.HasSuffix(name.Name, "json") {
              continue
            }
            var the_name = name.Name // closure madness
            table.AddChild(gui.MakeButton("standard", the_name, 300, 1, 1, 1, 1,
              func(int64) {
                nlevel, err := game.LoadLevel(basedir, the_name)
                if err != nil {
                  panic(err.Error())
                }
                ui.RemoveChild(level.GetGui())
                ui.AddChild(nlevel.GetGui())
                ui.RemoveChild(table)
                level = nlevel
                level.Setup()
                load_widget = nil
              }))
          }
          load_widget = table
          ui.AddChild(load_widget)
        }
      }
      level.Terrain.Zoom(zoom * 0.0025)
    }
  }

  fmt.Printf("")
}

func main() {
  <-quit
}
