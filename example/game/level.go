package game

import (
  "errors"
  "glop/gin"
  "glop/gui"
  "glop/sprite"
  "gl"
  "math"
  "github.com/arbaal/mathgl"
  "json"
  "path/filepath"
  "io"
  "io/ioutil"
  "os"
  "fmt"
)

type Terrain string

type UnitPlacement struct {
  // What side the unit initially in this cell belongs to.  0 Means that there
  // is no unit here (hence Name is irrelevant).
  Side int

  // If Side > 0 and Name == "" this cell is available for unit placement for
  // the specified side.  Otherwise Name indicates the name of the unit that
  // is initially placed in this cell at the beginning of the game.
  Name string
}

type staticCellData struct {
}
type cachedCellData struct {
}

type Highlight uint32

type CellData struct {
  // Permanent data
  Terrain Terrain
  Unit    UnitPlacement

  // Transient data
  highlight Highlight
}

const (
  None Highlight = 1 << iota

  Reachable
  // If the move action is selected this indicates cells that the unit can reach

  Attackable
  // If the attack action is selected this indicates cells that the unit can attack

  AttackMouseOver
  // MouseOver effect when in attack mode - could be multiple cells for an AOE

  MouseOver
  // The cell that the mouse is currently position over (should only ever be one)

  Selected
  // The unit currently selected

  NoLOS
  // indicates that the selected unit does not have visibility to this tile

  FogOfWar
  // indicates that no unit on a team has visibility to this tile

  MaxHighlights
)
const game_highlights = MouseOver
const combat_highlights = Reachable | Attackable
const editor_highlights = Selected
const visibility_highlights = FogOfWar | NoLOS
const all_highlights = MaxHighlights - 1

func (t *CellData) Clear(mask Highlight) {
  t.highlight &= ^mask
}

func (t *CellData) Render(x, y, z, scale float32) {
  gl.Disable(gl.TEXTURE_2D)
  gl.Enable(gl.BLEND)
  gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
  gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
  var r, g, b, a float32
  a = 0.2
  switch t.Terrain {
  case "plains":
    r, g, b = 0.1, 0.7, 0.4
  case "brush":
    r, g, b = 0.2, 0.6, 0.0
  case "water":
    r, g, b = 0.0, 0.0, 1.0
  case "swamp":
    r, g, b = 0.6, 0.5, 0.3
  case "jungle":
    r, g, b = 0.0, 0.7, 0.0
  case "hills":
    r, g, b = 0.9, 0.1, 0.3
  default:
    r, g, b = 1, 0, 0
  }
  x *= scale
  y *= scale
  gl.Color4f(r, g, b, a)
  gl.Begin(gl.QUADS)
  gl.Vertex3f(x, y, z)
  gl.Vertex3f(x, y+scale, z)
  gl.Vertex3f(x+scale, y+scale, z)
  gl.Vertex3f(x+scale, y, z)
  gl.End()

  draw_quad := func() {
    gl.Color4f(r, g, b, a)
    gl.Begin(gl.QUADS)
    gl.Vertex3f(x, y, z)
    gl.Vertex3f(x, y+scale, z)
    gl.Vertex3f(x+scale, y+scale, z)
    gl.Vertex3f(x+scale, y, z)
    gl.End()
  }

  if t.highlight != None {
    if t.highlight&FogOfWar != 0 {
      r, g, b, a = 0, 0, 0, 0.8
      draw_quad()
    } else if t.highlight&NoLOS != 0 {
      r, g, b, a = 0, 0, 0, 0.6
      draw_quad()
    } else {
      if t.highlight&Reachable != 0 {
        r, g, b, a = 0, 0.2, 0.5, 0.2
        draw_quad()
      }
      if t.highlight&AttackMouseOver != 0 {
        r, g, b, a = 0.7, 0.2, 0.2, 0.5
        draw_quad()
      }
      if t.highlight&Attackable != 0 {
        r, g, b, a = 0.7, 0.2, 0.2, 0.2
        draw_quad()
      }
      if t.highlight&MouseOver != 0 {
        r, g, b, a = 0.1, 0.9, 0.2, 0.4
        draw_quad()
      }
      if t.highlight&Selected != 0 {
        r, g, b, a = 0.0, 0.7, 0.4, 0.5
        draw_quad()
      }
    }
  }
}

// Contains everything about a level that is stored on disk
type StaticLevelData struct {
  bg_path string
  grid    [][]CellData
}
func (s *StaticLevelData) NumVertex() int {
  return len(s.grid) * len(s.grid[0])
}
func (s *StaticLevelData) fromVertex(v int) (int, int) {
  return v % len(s.grid), v / len(s.grid)
}
func (s *StaticLevelData) toVertex(x, y int) int {
  return x + y*len(s.grid)
}

type unitGraph struct {
  *Level
  move_cost map[Terrain]int
}
// Assumes that src and dst are adjacent
func (l unitGraph) costToMove(src, dst int) float64 {
  x, y := l.fromVertex(src)
  x2, y2 := l.fromVertex(dst)

  cost_c, ok := l.move_cost[l.grid[x2][y2].Terrain]
  if !ok {
    return -1
  }
  if x == x2 || y == y2 {
    return float64(cost_c + 1)
  }

  cost_a, ok := l.move_cost[l.grid[x][y2].Terrain]
  if !ok {
    return -1
  }
  cost_b, ok := l.move_cost[l.grid[x2][y].Terrain]
  if !ok {
    return -1
  }

  cost_ab := float64(cost_a+cost_b+2) / 2
  return math.Max(cost_ab, float64(cost_c))
}
func (l *unitGraph) Adjacent(v int) ([]int, []float64) {
  x, y := l.fromVertex(v)
  var adj []int
  var weight []float64

  // separate arrays for the adjacent diagonal cells, this way we make sure they are listed
  // at the end so that searches will prefer orthogonal adjacent cells
  var adj_diag []int
  var weight_diag []float64

  for dx := -1; dx <= 1; dx++ {
    if x+dx < 0 || x+dx >= len(l.grid) {
      continue
    }
    for dy := -1; dy <= 1; dy++ {
      if dx == 0 && dy == 0 {
        continue
      }
      if y+dy < 0 || y+dy >= len(l.grid[0]) {
        continue
      }

      // Don't want to be able to walk through other units
      occupied := false
      for i := range l.Entities {
        if int(l.Entities[i].pos.X) == x+dx && int(l.Entities[i].pos.Y) == y+dy {
          occupied = true
          break
        }
      }
      if occupied {
        continue
      }

      // Prevent moving along a diagonal if we couldn't get to that space normally via
      // either of the non-diagonal paths
      cost := l.costToMove(l.toVertex(x, y), l.toVertex(x+dx, y+dy))
      if cost < 0 {
        continue
      }
      if dx != 0 && dy != 0 {
        adj_diag = append(adj_diag, l.toVertex(x+dx, y+dy))
        weight_diag = append(weight_diag, cost)
      } else {
        adj = append(adj, l.toVertex(x+dx, y+dy))
        weight = append(weight, cost)
      }
    }
  }
  for i := range adj_diag {
    adj = append(adj, adj_diag[i])
    weight = append(weight, weight_diag[i])
  }
  return adj, weight
}

type Command int

const (
  NoCommand Command = iota
  Move
  Attack
)

type BoardPos struct {
  mathgl.Vec2
  level *Level
}

func (l *Level) MakeBoardPos(x,y int) BoardPos {
  return BoardPos{mathgl.Vec2{ float32(x), float32(y)}, l}
}

func (l *Level) MakeBoardPosFromVertex(v int) BoardPos {
  x,y := l.fromVertex(v)
  return l.MakeBoardPos(x, y)
}

func (bp *BoardPos) Xi() int {
  return int(bp.X)
}

func (bp *BoardPos) Yi() int {
  return int(bp.Y)
}

func (bp *BoardPos) Vertex() int {
  return bp.Xi() + bp.Yi()*len(bp.level.grid)
}

func (bp *BoardPos) AreEqual(t *BoardPos) bool {
  if bp.Xi() != t.Xi() { return false }
  if bp.Yi() != t.Yi() { return false }
  return true
}

// Contains everything for the playing of the game
type Level struct {
  StaticLevelData
  directory string

  editor     *Editor
  editor_gui *gui.CollapseWrapper

  // whose turn it is, side 1 goes first, then 2, then back to 1...
  side     int
  side_gui *gui.TextLine

  // unset when the cache is cleared, lets Think() know it has to refil the cache
  cached bool

  // The single gui element containing all other elements related to the
  // game
  game_gui *gui.HorizontalTable

  // The gui element rendering the terrain and all of the other drawables
  Terrain *gui.Terrain

  // The gui elements that show entity information
  selected_gui *EntityStatsWindow
  targeted_gui *EntityStatsWindow

  Entities []*Entity

  selected *Entity
  hovered  *Entity

  // window coords of the mouse
  winx, winy int

  // The most recently prepped valid command
  command Command

  // Indicates if an action is currently happening such that a new one cannot
  // start yet
  mid_action bool

  current_action Action


  // MOVE data

  // If a unit is selected this will hold the list of cells that are reachable
  // from that unit's position within its allotted AP

  reachable []int
}

func (l *Level) GetCellAtVertex(v int) *CellData {
  x,y := l.fromVertex(v)
  return &l.grid[x][y]
}
func (l *Level) GetCellAtPos(bp BoardPos) *CellData {
  return &l.grid[bp.Xi()][bp.Yi()]
}
func (l *Level) GetSelected() *Entity {
  return l.selected
}
func (l *Level) GetHovered() *Entity {
  return l.hovered
}
func (l *Level) Round() {
  // Can't start the next round until the current action completes
  if l.mid_action { return }

  l.selected = nil
  if l.current_action != nil {
    l.current_action.Cancel()
  }
  l.current_action = nil
  l.clearCache(all_highlights)
  l.side = l.side%2 + 1
  if l.side == 1 {
    l.side_gui.SetText("It is The Jungle's turn to move")
  } else {
    l.side_gui.SetText("It is The Man's turn to move")
  }
  for i := range l.Entities {
    if l.Entities[i].side != l.side {
      continue
    }
    l.Entities[i].OnRound()
  }
  l.updateDependantGuis()
}

func (l *Level) Setup() {
  for i := range l.Entities {
    l.Entities[i].OnSetup()
  }
  l.Round()
}

func (l *Level) clearCache(mask Highlight) {
  for i := range l.grid {
    for j := range l.grid[i] {
      l.grid[i][j].Clear(mask)
    }
  }
  l.cached = false
}

func (l *Level) figureVisible() {
  if !l.editor_gui.Collapsed {
    return
  }

  // If the player hasn't selected a unit we will just show all tiles that are
  // visible to any unit on that side as being equally visible, but if the
  // player has selcted a unit we will put a very light fog over the cells
  // that unit doesn't have LOS to.

  l.clearCache(visibility_highlights)

  // First do the selected unit's visibility stuff
  if l.selected != nil {
    for x := range l.grid {
      for y := range l.grid[0] {
        if _, ok := l.selected.visible[l.toVertex(x, y)]; !ok {
          l.grid[x][y].highlight |= NoLOS
        }
      }
    }
  }

  // Now do the visibility for all units on that side, have to first find the
  // union of all visible vertices.
  visible := make(map[int]bool, 100)
  for i := range l.Entities {
    if l.Entities[i].side != l.side {
      continue
    }
    for v, _ := range l.Entities[i].visible {
      visible[v] = true
    }
  }
  for x := range l.grid {
    for y := range l.grid[0] {
      if _, ok := visible[l.toVertex(x, y)]; !ok {
        l.grid[x][y].highlight |= FogOfWar
      }
    }
  }
}

func (l *Level) Think(dt int64) {
  if l.current_action != nil {
    if l.mid_action && l.current_action.Maintain(dt) {
      l.selected_gui.actions.SetSelectedIndex(-1)
      l.current_action = nil
      l.mid_action = false
    }
  }

  if l.selected != nil {
    // If we are committed to an action let's make sure that the UI doesn't
    // let us choose other actions until it's done.
    if l.mid_action {
      for i := range l.selected.actions {
        if l.selected.actions[i] == l.current_action {
          l.selected_gui.actions.SetSelectedIndex(i - 1)
        }
      }
    } else {
      index := l.selected_gui.actions.GetSelectedIndex()
      if index >= 0 && l.current_action != l.selected.actions[index + 1] {
        l.SelectAction(index + 1)
      }
    }
  }

  l.clearCache(MouseOver | AttackMouseOver | Selected)
  l.figureVisible()

  // Draw all sprites
  for i := range l.Entities {
    e := l.Entities[i]
    pbx := int(e.pos.X)
    pby := int(e.pos.Y)
    e.Think(dt)
    if pbx != int(e.pos.X) || pby != int(e.pos.Y) {
      l.clearCache(game_highlights)
    }
    if l.grid[int(e.pos.X)][int(e.pos.Y)].highlight&FogOfWar == 0 {
      l.Terrain.AddUprightDrawable(e.pos.X+0.25, e.pos.Y+0.25, e.s)
    }
  }

  l.editor.Think()

  // Highlight selected entity
  if l.selected != nil {
    cell := &l.grid[int(l.selected.pos.X)][int(l.selected.pos.Y)]
    cell.highlight |= Selected
  }

  bx, by := l.Terrain.WindowToBoard(l.winx, l.winy)
  // mouseovers
  if l.current_action != nil {
    l.current_action.MouseOver(float64(bx), float64(by))
  }
  if l.selected == nil {
    // Highlight the square under the cursor
    mx := int(bx)
    my := int(by)
    if mx >= 0 && my >= 0 && mx < len(l.grid) && my < len(l.grid[0]) {
      cell := &l.grid[mx][my]
      cell.highlight |= MouseOver
    }
  }

  l.cached = true

  // Draw tiles
  for i := range l.grid {
    for j := range l.grid[i] {
      l.Terrain.AddFlattenedDrawable(float32(i), float32(j), &l.grid[i][j])
    }
  }
}

// Some gui elements are dependent on entities we have selected, etc...
// We need to be very careful when we modify these things, if we modified them
// in Think() then the gui would be modified between layout and render which
// would cause a single-frame gui glitch.  This is the only function that
// directly modifies the gui elements so that we can easily isolate these
// calls.  This function should only be called during event-handling or
// Round()
func (l *Level) updateDependantGuis() {
  l.selected_gui.SetEntity(l.selected)
  if l.hovered == l.selected {
    l.targeted_gui.SetEntity(nil)
  } else {
    l.targeted_gui.SetEntity(l.hovered)
  }
}

func (l *Level) handleClickInGameMode(click mathgl.Vec2) {
  var ent *Entity
  var dist float32 = float32(math.Inf(1))
  for i := range l.Entities {
    var cc mathgl.Vec2
    cc.Assign(&click)
    cc.Subtract(&mathgl.Vec2{l.Entities[i].pos.X + 0.5, l.Entities[i].pos.Y + 0.5})
    dx := cc.X
    if dx < 0 {
      dx = -dx
    }
    dy := cc.Y
    if dy < 0 {
      dy = -dy
    }
    d := float32(math.Max(float64(dx), float64(dy)))
    if d < dist {
      dist = d
      ent = l.Entities[i]
    }
  }
  // At this point ent is the entity closest to the point that the user clicked
  // and dist is the distance from the click to the center of the entity's cell

  if dist > 0.5 {
    ent = nil
  }
  if l.current_action == nil {
    if ent != nil && ent.side == l.side {
      l.selected = ent
    }
  } else {
    if l.current_action.MouseClick(float64(click.X), float64(click.Y)) {
      l.mid_action = true
    }
  }
}

func (l *Level) HandleEventGroup(event_group gin.EventGroup) {
  defer l.updateDependantGuis()

  cursor := event_group.Events[0].Key.Cursor()
  if cursor == nil {
    return
  }
  l.winx, l.winy = cursor.Point()
  bx, by := l.Terrain.WindowToBoard(l.winx, l.winy)
  if bx < 0 || by < 0 || int(bx) >= len(l.grid) || int(by) >= len(l.grid[0]) {
    return
  }

  l.hovered = nil
  for i := range l.Entities {
    x, y := l.Entities[i].Coords()
    if x == int(bx) && y == int(by) {
      l.hovered = l.Entities[i]
    }
  }

  if !l.editor_gui.Collapsed {
    l.editor.HandleEventGroup(event_group, int(bx), int(by))
    return
  }

  found, event := event_group.FindEvent(gin.MouseLButton)
  if !found || event.Type != gin.Press {
    return
  }
  click := mathgl.Vec2{bx, by}
  l.handleClickInGameMode(click)
}

type levelDataCell struct {
  Terrain Terrain
  Unit    UnitPlacement
}
type levelData struct {
  // TOOD: Need to track the file this came from so we can copy it when
  // we save
  Image string

  Cells [][]levelDataCell
}

type levelDataContainer struct {
  Level levelData
}

func (sld *StaticLevelData) makeLevelDataContainer() *levelDataContainer {
  var ldc levelDataContainer
  ldc.Level.Image = sld.bg_path
  ldc.Level.Cells = make([][]levelDataCell, len(sld.grid))
  for i := range ldc.Level.Cells {
    ldc.Level.Cells[i] = make([]levelDataCell, len(sld.grid[0]))
  }
  for i := range ldc.Level.Cells {
    for j := range ldc.Level.Cells[i] {
      ldc.Level.Cells[i][j].Terrain = sld.grid[i][j].Terrain
      ldc.Level.Cells[i][j].Unit = sld.grid[i][j].Unit
    }
  }
  return &ldc
}

func (ldc *levelDataContainer) Write(out io.Writer) error {
  data, err := json.Marshal(&ldc)
  if err != nil {
    return err
  }
  _, err = out.Write(data)
  return err
}

// TODO: Should save everything to a temporary directory and then copy
// from there to the appropriate place.  In the event that the save
// fails we don't want to be left with something inconsistent.
func (l *Level) SaveLevel(pathname string) error {
  out, err := os.Create(pathname)
  if err != nil {
    return nil
  }
  defer out.Close()
  var ldc levelDataContainer
  ldc.Level.Image = l.bg_path
  ldc.Level.Cells = make([][]levelDataCell, len(l.grid))
  for i := range ldc.Level.Cells {
    ldc.Level.Cells[i] = make([]levelDataCell, len(l.grid[0]))
  }
  for i := range ldc.Level.Cells {
    for j := range ldc.Level.Cells[i] {
      ldc.Level.Cells[i][j].Terrain = l.grid[i][j].Terrain
    }
  }
  data, err := json.Marshal(&ldc)
  if err != nil {
    return err
  }
  _, err = out.Write(data)
  return err
}

func (l *Level) SelectAction(n int) {
  if l.selected == nil { return }
  if l.mid_action { return }
  if n < 0 {
    if l.current_action != nil {
      l.current_action.Cancel()
      l.current_action = nil
      l.selected_gui.actions.SetSelectedIndex(-1)
    }
    return
  }
  if n >= len(l.selected.actions) { return }

  // If the current action is selected then we cancel it and reselect it.
  // This gives users a way to deselect anything they may have selected
  // with the current action without having to go through any other buttons
  // than the ones they were already using.

  if l.current_action != nil {
    l.current_action.Cancel()
    l.selected_gui.actions.SetSelectedIndex(-1)
  }

  if l.selected.actions[n].Prep() {
    l.current_action = l.selected.actions[n]
    l.selected_gui.actions.SetSelectedIndex(n-1)
  }
}

func LoadLevel(datadir, mapname string) (*Level, error) {
  datapath := filepath.Join(datadir, "maps", mapname)
  datafile, err := os.Open(datapath)
  if err != nil {
    return nil, err
  }
  data, err := ioutil.ReadAll(datafile)
  if err != nil {
    return nil, err
  }
  var ldc levelDataContainer
  err = json.Unmarshal(data, &ldc)
  if err != nil {
    panic(fmt.Sprintf("Error reading %s: %s\n", mapname, err.Error()))
  }

  var level Level
  level.directory = datadir

  dx := len(ldc.Level.Cells)
  dy := len(ldc.Level.Cells[0])
  all_cells := make([]CellData, dx*dy)
  level.grid = make([][]CellData, dx)
  for i := range level.grid {
    level.grid[i] = all_cells[i*dy : (i+1)*dy]
  }
  all_units, err := LoadAllUnits(datadir)
  if err != nil {
    panic(fmt.Sprintf("Error reading unit files: %s\n", err.Error()))
  }
  unit_map := make(map[string]*UnitType)
  for _, unit := range all_units {
    unit_map[unit.Name] = unit
  }
  for i := range level.grid {
    for j := range level.grid[i] {
      level.grid[i][j].Terrain = ldc.Level.Cells[i][j].Terrain
      level.grid[i][j].Unit = ldc.Level.Cells[i][j].Unit
      if level.grid[i][j].Unit.Name != "" {
        if unit, ok := unit_map[level.grid[i][j].Unit.Name]; !ok {
          return nil, errors.New(fmt.Sprintf("Unable to find unit definition for '%s'.", level.grid[i][j].Unit))
        } else {
          sprite, err := sprite.LoadSprite(filepath.Join(datadir, "sprites", unit.Sprite))
          if err != nil {
            return nil, err
          }
          side := level.grid[i][j].Unit.Side
          level.addEntity(*unit, i, j, side, 0.0075, sprite)
        }
      }
    }
  }
  level.bg_path = ldc.Level.Image
  bg_path := filepath.Join(filepath.Clean(level.directory), "maps", level.bg_path)
  terrain, err := gui.MakeTerrain(bg_path, 100, dx, dy, 65)
  if err != nil {
    fmt.Printf("Error making terrain: %s\n", err.Error())
  }
  level.Terrain = terrain
  terrain.SetEventHandler(&level)

  level.side_gui = gui.MakeTextLine("standard", "", 500, 1, 1, 1, 1)
  level.editor = MakeEditor(&level.StaticLevelData, datadir, mapname)
  level.game_gui = gui.MakeHorizontalTable()
  game_only_gui := gui.MakeVerticalTable()
  level.selected_gui = MakeStatsWindow(true)
  level.targeted_gui = MakeStatsWindow(false)
  entity_guis := gui.MakeHorizontalTable()
  entity_guis.AddChild(level.selected_gui)
  entity_guis.AddChild(level.targeted_gui)
  game_only_gui.AddChild(level.Terrain)
  game_only_gui.AddChild(level.side_gui)
  game_only_gui.AddChild(entity_guis)
  level.game_gui.AddChild(game_only_gui)
  level.editor_gui = gui.MakeCollapseWrapper(level.editor.GetGui())
  level.game_gui.AddChild(level.editor_gui)
  return &level, nil
}

func (l *Level) GetGui() gui.Widget {
  return l.game_gui
}

func (l *Level) ToggleEditor() {
  l.editor_gui.Collapsed = !l.editor_gui.Collapsed
}

func (l *Level) addEntity(unit_type UnitType, x, y, side int, move_speed float32, sprite *sprite.Sprite) *Entity {
  var ent Entity
  ent = Entity{
    UnitStats: UnitStats{
      Base: &unit_type,
    },
    pos:   l.MakeBoardPos(x, y),
    side:  side,
    s:     sprite,
    level: l,
    CosmeticStats: CosmeticStats{
      Move_speed: move_speed,
    },
  }
  ent.actions = append(ent.actions, MakeAction("move", &ent))
  for _, name := range unit_type.Weapons {
    ent.actions = append(ent.actions, MakeAction(name, &ent))
  }
  l.Entities = append(l.Entities, &ent)
  return &ent
}
