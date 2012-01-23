package main

import (
  "glop/gos"
  "glop/gin"
  "glop/gui"
  "glop/system"
  "glop/render"
  "game"
  "game/stats"
  "runtime"
  "runtime/debug"
  "runtime/pprof"
  "fmt"
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

func init() {
  sys = system.Make(gos.GetSystemInterface())
  quit = make(chan bool)
}

func actualMain() {
  defer func() {
    quit <- true
  }()
  defer func() {
    if r := recover(); r != nil {
      data := debug.Stack()
      fmt.Printf("%s\n", string(data))
      out,err := os.Open("crash.txt")
      if err != nil {
        out.Write(data)
        out.Close()
      }
    }
  } ()
  sys.Startup()
  runtime.LockOSThread()
  for !sys.Think() {
    runtime.UnlockOSThread()
    runtime.Gosched()
    runtime.LockOSThread()
  }
  // Running a binary via osx's package mechanism will add a flag that begins with
  // '-psn' so we have to find it and pretend like we were expecting it so that go
  // doesn't asplode because of an unexpected flag.
  for _, arg := range os.Args {
    if len(arg) >= 4 && arg[0:4] == "-psn" {
      flag.Bool(arg[1:], false, "HERE JUST TO APPEASE GO'S FLAG PACKAGE")
    }
  }
  flag.Parse()

  basedir := filepath.Join(os.Args[0], "..", "..")

  // TODO: Loading weapon specs should be done automatically - it just needs the datadir

  factor := 1.0
  wdx := int(factor * float64(1024))
  wdy := int(factor * float64(768))

  render.Init()
  render.Queue(func() {
    sys.CreateWindow(0, 0, wdx, wdy)
    sys.EnableVSync(true)
    })
  render.Purge()
  _, _, wdx, wdy = sys.GetWindowDims()
  ui,err := gui.Make(gin.In(), gui.Dims{wdx, wdy}, filepath.Join(basedir, *font_path))
  if err != nil {
    panic(err.Error())
  }
  //  table := gui.MakeVerticalTable()
  //  ui.AddChild(table)

  effects_dir := filepath.Join(basedir, "effects")
  stats.RegisterAllEffectsInDir(effects_dir)

  actions_dir := filepath.Join(basedir, "actions")
  game.RegisterAllSpecsInDir(actions_dir)

  level := game.MakeLevel(basedir, "bosworth.json")
  render.Queue(func() {
    err = level.Fill()
    if err != nil {
      panic(err.Error())
    }
  })
  render.Purge()
  ui.AddChild(level.GetGui())
  //  level.Terrain.Move(10,10)

  //  table := anch.InstallWidget(&gui.VerticalTable{}, gui.Anchor{0,0, 0,0})

  n := 0

  // TODO: Would be better to only be vsynced, but apparently it can turn itself off
  // when the window disappears, so we need a safety net to slow it down if necessary
  ticker := time.Tick(1e7)

  level.Setup()
  prev := time.Now().UnixNano()

  profiling := false
  var load_widget gui.Widget
  for {
    n++
    next := time.Now().UnixNano()
    dt := (next - prev) / 1000000
    prev = next

    sys.Think()
    level.Think(dt)
    render.Queue(func() {
      ui.Draw()
    })
    sys.SwapBuffers()
    <-ticker
    if gin.In().GetKey('q').FramePressCount() > 0 {
      return
    }

    if ui.FocusWidget() == nil {
      cmd_keys := []gin.Key{
        gin.In().GetKey('`'),
        gin.In().GetKey('1'),
        gin.In().GetKey('2'),
        gin.In().GetKey('3'),
        gin.In().GetKey('4'),
        gin.In().GetKey('5'),
        gin.In().GetKey('6'),
        gin.In().GetKey('7'),
        gin.In().GetKey('8'),
        gin.In().GetKey('9'),
        gin.In().GetKey('0'),
      }
      for i := range cmd_keys {
        if cmd_keys[i].FramePressCount() > 0 {
          level.SelectAction(i)
        }
      }
      if gin.In().GetKey(gin.Escape).FramePressCount() > 0 {
        level.SelectAction(-1)
      }

      kw := gin.In().GetKey('w')
      ka := gin.In().GetKey('a')
      ks := gin.In().GetKey('s')
      kd := gin.In().GetKey('d')
      m_factor := 0.0075
      dx := m_factor * (kd.FramePressSum() - ka.FramePressSum())
      dy := m_factor * (kw.FramePressSum() - ks.FramePressSum())
      level.Terrain().Move(dx, dy)
      zoom := gin.In().GetKey('r').FramePressSum() - gin.In().GetKey('f').FramePressSum()
      if gin.In().GetKey('o').FramePressCount() > 0 {
        level.Round()
      }
      if gin.In().GetKey('e').FramePressCount()%2 == 1 {
        level.ToggleEditor()
      }
      if gin.In().GetKey('z').FramePressCount() > 0 {
        err := level.ExpLoad("/Users/runningwild/code/go-glop/example/game.dat")
        if err != nil {
          fmt.Printf("Error: %s\n", err.Error())
        }
        level.Fill()
      }
      if gin.In().GetKey('x').FramePressCount() > 0 {
        err := level.ExpSave("/Users/runningwild/code/go-glop/example/game.dat")
        if err != nil {
          fmt.Printf("Error: %s\n", err.Error())
        }
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
            if !strings.HasSuffix(name.Name(), "json") {
              continue
            }
            var the_name = name.Name() // closure madness
            table.AddChild(gui.MakeButton("standard", the_name, 300, 1, 1, 1, 1,
              func(int64) {
                nlevel := game.MakeLevel(basedir, the_name)
                err := level.Fill()
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
      level.Terrain().Zoom(zoom * 0.0025)
    }
  }

  fmt.Printf("")
}

func main() {
  go actualMain()
  <-quit
}
