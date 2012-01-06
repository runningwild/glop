package main

import (
  "fmt"
  "gl"
  "glop/gos"
  "glop/gin"
  "glop/gui"
  "glop/system"
  "glop/sprite"
  "time"
  "runtime"
  "glop/render"
  "path/filepath"
  "os"
  "sort"
)

type loadResult struct {
  anim *sprite.Sprite
  err error
}

var (
  sys system.System
  datadir string
  key_map KeyMap
  action_map KeyMap
  loaded chan loadResult
)

func init() {
  runtime.LockOSThread()
  datadir = filepath.Join(os.Args[0], "..", "..")
  var key_binds KeyBinds
  LoadJson(filepath.Join(datadir, "bindings.json"), &key_binds)
  key_map = key_binds.MakeKeyMap()
  key_binds = nil
  LoadJson(filepath.Join(datadir, "actions.json"), &key_binds)
  action_map = key_binds.MakeKeyMap()
  loaded = make(chan loadResult)
}

func GetStoreVal(key string) string {
  var store map[string]string
  LoadJson(filepath.Join(datadir, "store"), &store)
  if store == nil {
    store = make(map[string]string)
  }
  val := store[key]
  return val
}

func SetStoreVal(key,val string) {
  var store map[string]string
  path := filepath.Join(datadir, "store")
  LoadJson(path, &store)
  if store == nil {
    store = make(map[string]string)
  }
  store[key] = val
  SaveJson(path, store)
}

type spriteBox struct {
  gui.EmbeddedWidget
  gui.BasicZone
  gui.NonThinker
  gui.NonResponder
  gui.NonFocuser
  gui.Childless
  s *sprite.Sprite
}
func makeSpriteBox(s *sprite.Sprite) *spriteBox {
  var sb spriteBox
  sb.EmbeddedWidget = &gui.BasicWidget{ CoreWidget: &sb }
  sb.Request_dims = gui.Dims{ 300, 300 }
  return &sb
}
func (sb *spriteBox) String() string {
  return "sprite box"
}
func (sb *spriteBox) Draw(region gui.Region) {
  if sb.s != nil {
    gl.Color4f(0.2, 0.1, 0.0, 1)
    gl.Begin(gl.QUADS)
    gl.Vertex2i(region.X, region.Y)
    gl.Vertex2i(region.X, region.Y + region.Dy)
    gl.Vertex2i(region.X + region.Dx, region.Y + region.Dy)
    gl.Vertex2i(region.X + region.Dx, region.Y)
    gl.End()
    sb.s.Render(float32(region.X + region.Dx / 2), float32(region.Y), 0, 1)
  }
}


func main() {
  sys = system.Make(gos.GetSystemInterface())
  sys.Startup()
  wdx := 700
  wdy := 500
  render.Init()
  var ui *gui.Gui
  render.Queue(func() {
    sys.CreateWindow(50, 50, wdx, wdy)
    sys.EnableVSync(true)
    ui,_ = gui.Make(gin.In(), gui.Dims{ wdx, wdy }, filepath.Join(datadir, "fonts", "skia.ttf"))
  })
  render.Purge()

  anchor := gui.MakeAnchorBox(gui.Dims{ wdx, wdy })
  ui.AddChild(anchor)

  error_msg := gui.MakeTextLine("standard", "", wdx, 1, 0.5, 0.5, 1)
  anchor.AddChild(error_msg, gui.Anchor{ 0,0, 0,0.2})

  actions_list := gui.MakeVerticalTable()
  keyname_list := gui.MakeVerticalTable()
  both_lists := gui.MakeHorizontalTable()
  both_lists.AddChild(actions_list)
  both_lists.AddChild(keyname_list)
  anchor.AddChild(both_lists, gui.Anchor{ 1,0.5, 1,0.5 })
  var actions []string
  for action := range action_map {
    actions = append(actions, action)
  }
  sort.Strings(actions)
  for _,action := range actions {
    actions_list.AddChild(gui.MakeTextLine("standard", action, 150, 1, 1, 1, 1))
    keyname_list.AddChild(gui.MakeTextLine("standard", action_map[action].Name(), 50, 1, 1, 1, 1))
  }

  current_anim := gui.MakeTextLine("standard", "", 300, 1, 1, 1, 1)
  current_state := gui.MakeTextLine("standard", "", 300, 1, 1, 1, 1)
  frame_data := gui.MakeVerticalTable()
  frame_data.AddChild(current_anim)
  frame_data.AddChild(current_state)
  anchor.AddChild(frame_data, gui.Anchor{ 0,1, 0,1 })

  speed := 100
  speed_text := gui.MakeTextLine("standard", "Speed: 100%", 150, 1, 1, 1, 1)
  anchor.AddChild(speed_text, gui.Anchor{ 0,0, 0,0 })

  sprite_box := makeSpriteBox(nil)
  anchor.AddChild(sprite_box, gui.Anchor{ 0.5,0.5, 0.25, 0.5 })

  var chooser gui.Widget
  curdir := GetStoreVal("curdir")
  if curdir == "" {
    curdir = datadir
  } else {
    go func() {
      anim, err := sprite.LoadSprite(curdir)
      loaded <- loadResult{ anim, err } 
    } ()
  }
  then := time.Now()
  for key_map["quit"].FramePressCount() == 0 {
    render.Purge()
    sys.SwapBuffers()
    sys.Think()
    now := time.Now()
    dt := (now.Nanosecond() - then.Nanosecond()) / 1000000
    then = now
    select {
    case load := <-loaded:
      if load.err != nil {
        error_msg.SetText(load.err.Error())
        current_anim.SetText("")
      } else {
        sprite_box.s = load.anim
        error_msg.SetText("")
      }
    default:
    }
    if sprite_box.s != nil {
      sprite_box.s.Think(int64(float64(dt) * float64(speed) / 100))
      current_anim.SetText(sprite_box.s.CurAnim())
      current_state.SetText(sprite_box.s.CurState())
    }
    render.Queue(func() {
      gl.ClearColor(0, 0, 0, 1)
      gl.Clear(gl.COLOR_BUFFER_BIT)
      ui.Draw()
    })

    if sprite_box.s != nil {
      if action_map["reset"].FramePressCount() > 0 {
        go func() {
          anim, err := sprite.LoadSprite(curdir)
          loaded <- loadResult{ anim, err } 
        } ()
      } else {
        for k,v := range action_map {
          for i := 0; i < v.FramePressCount(); i++ {
            sprite_box.s.Command(k)
          }
        }
      }
    }

    if key_map["load"].FramePressCount() > 0 && chooser == nil {
      anch := gui.MakeAnchorBox(gui.Dims{ wdx, wdy })
      file_chooser := gui.MakeFileChooser(curdir,
        func(path string, err error) {
          if err == nil && len(path) > 0 {
            curdir,_ = filepath.Split(path)
            go func() {
              anim, err := sprite.LoadSprite(curdir)
              if err == nil {
                SetStoreVal("curdir", curdir)
              }
              loaded <- loadResult{ anim, err } 
            } ()
          }
          ui.RemoveChild(chooser)
          chooser = nil
        },
        func(path string, is_dir bool) bool {
          return true
        })
      anch.AddChild(file_chooser, gui.Anchor{ 0.5, 0.5, 0.5, 0.5 })
      chooser = anch
      ui.AddChild(chooser)
    }

    delta := key_map["speed up"].FramePressAmt() - key_map["slow down"].FramePressAmt()
    if delta != 0 {
      speed += int(delta)
      if speed < 1 {
        speed = 1
      }
      if speed > 100 {
        speed = 100
      }
      speed_text.SetText(fmt.Sprintf("Speed: %d%%", speed))
    }
  }
}
