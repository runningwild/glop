package game

import (
  "glop/gin"
  "glop/gui"
  "os"
  "fmt"
  "path/filepath"
  "io/ioutil"
)

type Editor struct {
  level    *StaticLevelData
  selected map[*CellData]bool

  ui *gui.VerticalTable

  // All of these are for editing attributes of the terrain
  terrain_parent *gui.CollapseWrapper
  terrain_type   *gui.SelectTextBox
  current_type   string


  // Event processing stuff

  // When selecting tiles, if the user clicks on an already-selected tile it
  // will be deselected, as well as any other tiles the mouse is dragged over
  invert bool
}

func MakeEditor(level_data *StaticLevelData, dir,filename string) *Editor {
  var e Editor
  e.level = level_data
  e.selected = make(map[*CellData]bool, 50)
  e.ui = gui.MakeVerticalTable()
  e.ui.AddChild(gui.MakeTextLine("standard", "The Editor", 250, 1, 1, 1, 1))
  bg_name_widget := gui.MakeTextEditLine("standard", level_data.bg_path, 250, 1, 1, 1, 1)
  e.ui.AddChild(bg_name_widget)
  filename_widget := gui.MakeTextEditLine("standard", filename, 250, 1, 1, 1, 1)
  e.ui.AddChild(filename_widget)

  // Save everything to a whole new directory, including the background image
  e.ui.AddChild(gui.MakeButton("standard", "Save", 150, 1, 1, 0, 1, func(int64) {
    ldc := e.level.makeLevelDataContainer()
    bg_in_path := filepath.Join(dir, ldc.Level.Image)
    bg_in,err := os.Open(bg_in_path)
    if err != nil {
      fmt.Printf("Err: %s\n", err.String())
      return
    }
    defer bg_in.Close()

    bg_out_path := filepath.Join(dir, bg_name_widget.GetText())
    if bg_out_path != bg_in_path {
      image_data,err := ioutil.ReadAll(bg_in)
      if err != nil {
        fmt.Printf("Err: %s\n", err.String())
        return
      }
      err = ioutil.WriteFile(bg_out_path, image_data, 0664)
    }

    data_path := filepath.Join(dir, filename_widget.GetText())
    data_file,err := os.Create(data_path)
    if err != nil {
      fmt.Printf("Err: %s\n", err.String())
      return
    }
    err = ldc.Write(data_file)
    if err != nil {
      fmt.Printf("Err: %s\n", err.String())
    }
  }))
  terrain_data := gui.MakeVerticalTable()
  e.terrain_parent = gui.MakeCollapseWrapper(terrain_data)
  e.terrain_type = gui.MakeSelectTextBox(GetRegisteredTerrains(), 200)
  terrain_data.AddChild(e.terrain_type)
  e.terrain_parent.Collapsed = true

  e.ui.AddChild(e.terrain_parent)
  return &e
}

func (e *Editor) SelectCell(x,y int) {
  if e.invert {
    e.selected[&e.level.grid[x][y]] = false,false
  } else {
    e.selected[&e.level.grid[x][y]] = true
  }
  e.terrain_parent.Collapsed = len(e.selected) == 0
  if len(e.selected) > 0 {
    var name string
    for cell,_ := range e.selected {
      name = cell.Name()
      break
    }
    for cell,_ := range e.selected {
      if name != cell.Name() {
        name = ""
        break
      }
    }
    e.terrain_type.SetSelectedOption(name)
  }
}

func (e *Editor) GetGui() gui.Widget {
  return e.ui
}

func (e *Editor) Think() {
  if len(e.selected) > 0 {
    if e.current_type != e.terrain_type.GetSelectedOption() {
      for cell,_ := range e.selected {
        cell.staticCellData.Terrain = MakeTerrain(e.terrain_type.GetSelectedOption())
      }
    }
  }
  for i := range e.level.grid {
    for j := range e.level.grid[i] {
      if _,ok := e.selected[&e.level.grid[i][j]]; ok {
        e.level.grid[i][j].highlight |= Selected
      } else {
        e.level.grid[i][j].highlight &= ^Selected
      }
    }
  }
}

func (e *Editor) HandleEventGroup(event_group gin.EventGroup, x,y int) {
  if gin.In().GetKey(gin.MouseLButton).CurPressAmt() == 0 { return }
  if event_group.Events[0].Key.Id() == gin.MouseLButton && event_group.Events[0].Type == gin.Press {
    if gin.In().GetKey(gin.EitherShift).CurPressAmt() == 0 {
      e.selected = make(map[*CellData]bool)
    }
    _,ok := e.selected[&e.level.grid[x][y]]
    e.invert = ok
  }
  e.SelectCell(x, y)
}
