package game

import (
  "game/base"
  "glop/gin"
  "glop/gui"
  "os"
  "fmt"
  "path/filepath"
  "io/ioutil"
  "sort"
)

type Editor struct {
  level    *StaticLevelData
  selected map[*CellData]bool

  ui *gui.VerticalTable

  cell_parent *gui.CollapseWrapper

  // attributes of the terrain
  terrain_type *gui.SelectBox

  // units
  starting_unit *gui.SelectBox

  // side
  starting_side *gui.SelectBox

  // Event processing stuff

  // When selecting tiles, if the user clicks on an already-selected tile it
  // will be deselected, as well as any other tiles the mouse is dragged over
  invert bool
}

func MakeEditor(level_data *StaticLevelData, dir, filename string) *Editor {
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
    bg_in_path := filepath.Join(dir, "maps", ldc.Level.Image)
    bg_in, err := os.Open(bg_in_path)
    if err != nil {
      fmt.Printf("Err: %s\n", err.Error())
      return
    }
    defer bg_in.Close()

    bg_out_path := filepath.Join(dir, "maps", bg_name_widget.GetText())
    if bg_out_path != bg_in_path {
      image_data, err := ioutil.ReadAll(bg_in)
      if err != nil {
        fmt.Printf("Err: %s\n", err.Error())
        return
      }
      err = ioutil.WriteFile(bg_out_path, image_data, 0664)
    }

    data_path := filepath.Join(dir, "maps", filename_widget.GetText())
    data_file, err := os.Create(data_path)
    if err != nil {
      fmt.Printf("Err: %s\n", err.Error())
      return
    }
    err = ldc.Write(data_file)
    if err != nil {
      fmt.Printf("Err: %s\n", err.Error())
    }
  }))

  attributes := gui.MakeVerticalTable()
  e.cell_parent = gui.MakeCollapseWrapper(attributes)
  e.cell_parent.Collapsed = true
  e.ui.AddChild(e.cell_parent)

  var terrain_names []string
  err := base.LoadJson(filepath.Join(dir, "terrains.json"), &terrain_names)
  if err != nil {
    fmt.Printf("err: %s\n", err.Error())
  }
  e.terrain_type = gui.MakeSelectTextBox(terrain_names, 200)
  attributes.AddChild(e.terrain_type)

  units, _ := LoadAllUnits(dir)
  unit_names := make([]string, len(units))
  for i, unit := range units {
    unit_names[i] = unit.Name
  }
  unit_names = append(unit_names, "")
  sort.Strings(unit_names)
  e.starting_unit = gui.MakeSelectTextBox(unit_names, 200)
  attributes.AddChild(e.starting_unit)

  e.starting_side = gui.MakeSelectTextBox([]string{"None", "The Jungle", "The Man"}, 200)
  attributes.AddChild(e.starting_side)

  return &e
}

func (e *Editor) SelectCell(x, y int) {
  if e.invert {
    delete(e.selected, &e.level.grid[x][y])
  } else {
    e.selected[&e.level.grid[x][y]] = true
  }

  e.cell_parent.Collapsed = len(e.selected) == 0

  if len(e.selected) > 0 {
    var terrain base.Terrain
    var unit string
    var side int
    for cell, _ := range e.selected {
      terrain = cell.Terrain
      unit = cell.Unit.Name
      side = cell.Unit.Side
      break
    }
    for cell, _ := range e.selected {
      if terrain != cell.Terrain {
        terrain = ""
      }
      if unit != cell.Unit.Name {
        unit = ""
      }
      if side != cell.Unit.Side {
        side = 0
      }
    }
    e.terrain_type.SetSelectedOption(string(terrain))
    if unit == "" {
      e.starting_unit.SetSelectedIndex(-1)
    } else {
      e.starting_unit.SetSelectedOption(unit)
    }
    e.starting_side.SetSelectedIndex(side)
  }
}

func (e *Editor) GetGui() gui.Widget {
  return e.ui
}

// TODO: Right now if you select two squares, one with a unit and one without, the unit will be erased because the gui will be set to not having a unit and both cells will be set to match the gui.  Instead we need to make select boxes either report that they were clicked, or we need to manually track the value in it.
func (e *Editor) Think() {
  for cell, _ := range e.selected {
    if e.terrain_type.GetSelectedIndex() != -1 {
      cell.Terrain = base.Terrain(e.terrain_type.GetSelectedOption().(string))
    }
    if e.starting_unit.GetSelectedIndex() != -1 {
      cell.Unit.Name = e.starting_unit.GetSelectedOption().(string)
    }
    cell.Unit.Side = e.starting_side.GetSelectedIndex()
  }
  for i := range e.level.grid {
    for j := range e.level.grid[i] {
      if _, ok := e.selected[&e.level.grid[i][j]]; ok {
        e.level.grid[i][j].highlight |= Selected
      } else {
        e.level.grid[i][j].highlight &= ^Selected
      }
    }
  }
}

func (e *Editor) HandleEventGroup(event_group gin.EventGroup, x, y int) {
  if gin.In().GetKey(gin.MouseLButton).CurPressAmt() == 0 {
    return
  }
  if event_group.Events[0].Key.Id() == gin.MouseLButton && event_group.Events[0].Type == gin.Press {
    if gin.In().GetKey(gin.EitherShift).CurPressAmt() == 0 {
      e.selected = make(map[*CellData]bool)
    }
    _, ok := e.selected[&e.level.grid[x][y]]
    e.invert = ok
  }
  e.SelectCell(x, y)
}
