package game

import (
  "errors"
  "game/base"
  "game/stats"
  "glop/ai"
  "glop/gin"
  "glop/gui"
  "glop/util/algorithm"
  "encoding/gob"
  "gl"
  "math"
  "github.com/arbaal/mathgl"
  "encoding/json"
  "path/filepath"
  "io"
  "io/ioutil"
  "os"
  "fmt"
)

type UnitPlacement struct {
  // What side the unit initially in this cell belongs to.  0 Means that there
  // is no unit here (hence Name is irrelevant).
  Side int

  // If Side > 0 and Name == "" this cell is available for unit placement for
  // the specified side.  Otherwise Name indicates the name of the unit that
  // is initially placed in this cell at the beginning of the game.
  Name string
}

type Highlight uint32

type CellData struct {
  // Permanent data
  Terrain base.Terrain
  Unit    UnitPlacement

  // Transient data
  highlight Highlight
  ent       *Entity
}

const (
  None Highlight = 1 << iota

  Reachable
  // If the move action is selected this indicates cells that the unit can reach

  Attackable
  // If the attack action is selected this indicates cells that the unit can attack

  Targeted
  // For some actions that require selecting multiple targets this can serve to
  // indicate something that has been selected

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
      if t.highlight&Targeted != 0 {
        r, g, b, a = 0.7, 0.2, 0.2, 0.7
        draw_quad()
      } else if t.highlight&Attackable != 0 {
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
type staticLevelData struct {
  bg_path string
  grid    [][]CellData
}
func (s *staticLevelData) NumVertex() int {
  return len(s.grid) * len(s.grid[0])
}
func (s *staticLevelData) fromVertex(v int) (int, int) {
  return v % len(s.grid), v / len(s.grid)
}
func (s *staticLevelData) toVertex(x, y int) int {
  return x + y*len(s.grid)
}

type moveCoster interface {
  MoveCost(base.Terrain) int
}
type unitGraph struct {
  *Level
  moveCoster
}
// Assumes that src and dst are adjacent
func (l unitGraph) costToMove(src, dst int) float64 {
  x, y := l.fromVertex(src)
  x2, y2 := l.fromVertex(dst)

  cost_c := l.MoveCost(l.grid[x2][y2].Terrain)
  if cost_c < 0 {
    return -1
  }
  if x == x2 || y == y2 {
    return float64(cost_c + 1)
  }

  cost_a := l.MoveCost(l.grid[x][y2].Terrain)
  if cost_a < 0 {
    return -1
  }
  cost_b := l.MoveCost(l.grid[x2][y].Terrain)
  if cost_b < 0 {
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
        if int(l.Entities[i].Pos.X) == x+dx && int(l.Entities[i].Pos.Y) == y+dy {
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

type BoardPos struct {
  mathgl.Vec2
}

func MakeBoardPos(x,y int) BoardPos {
  return BoardPos{mathgl.Vec2{ float32(x), float32(y)} }
}

func (l *Level) MakeBoardPosFromVertex(v int) BoardPos {
  x,y := l.fromVertex(v)
  return MakeBoardPos(x, y)
}

func (bp BoardPos) Xi() int {
  return int(bp.X)
}

func (bp BoardPos) Yi() int {
  return int(bp.Y)
}

func (bp BoardPos) Sub(t BoardPos) BoardPos {
  return MakeBoardPos(bp.Xi() - t.Xi(), bp.Yi() - t.Yi())
}

func (bp BoardPos) Add(t BoardPos) BoardPos {
  return MakeBoardPos(bp.Xi() + t.Xi(), bp.Yi() + t.Yi())
}

func (bp BoardPos) Scale(scale int) BoardPos {
  return MakeBoardPos(bp.Xi() * scale, bp.Yi() * scale)
}

func (bp BoardPos) Vertex(level *Level) int {
  return bp.Xi() + bp.Yi()*len(level.grid)
}

func (bp BoardPos) IntEquals(t BoardPos) bool {
  if bp.Xi() != t.Xi() { return false }
  if bp.Yi() != t.Yi() { return false }
  return true
}

// Returns the maxnorm distance between two points
func (bp BoardPos) Dist(t BoardPos) int {
  return base.MaxNormi(bp.Xi(), bp.Yi(), t.Xi(), t.Yi())
}

func (bp BoardPos) Valid(level *Level) bool {
  if bp.Xi() < 0 { return false }
  if bp.Yi() < 0 { return false }
  if bp.Xi() >= len(level.grid) { return false }
  if bp.Yi() >= len(level.grid[0]) { return false }
  return true
}

// Contains everything for the playing of the game
type Level struct {
  staticLevelData

  // *** Begin exported fields
  // When saving the game the level is gobbed, so all exported fields are
  // saved, everything else is not.  All information must either be in these
  // fields, be derived from these fields, or be implicit in the fact that
  // saves are only allowed at certain times (between turns).

  // directory that everything is loaded from
  Directory string

  // Map that this level is using
  Mapname string
  
  // whose turn it is, side 1 goes first, then 2, then back to 1...
  Side     int

  Entities []*Entity
  // *** End exported fields

  editor     *Editor
  editor_gui *gui.CollapseWrapper

  side_gui *gui.TextLine

  // The single gui element containing all other elements related to the
  // game
  game_gui *gui.HorizontalTable

  // The gui element rendering the terrain and all of the other drawables
  terrain *gui.Terrain

  // The gui elements that show entity information
  selected_gui *EntityStatsWindow
  targeted_gui *EntityStatsWindow

  selected *Entity
  hovered  *Entity

  // window coords of the mouse
  winx, winy int

  // Indicates if an action is currently happening such that a new one cannot
  // start yet
  mid_action bool

  // The action that is currently executing
  current_action   Action

  // The next action to execute, should only be set if there isn't already an
  // action executing.  This action is stored here until all entities are in
  // the ready state so that animations don't get backed up.
  pending_action   Action

  // If an action has been paused while an interrupt happens then this is the
  // interrupt.  If the action is cancelled then the interrupt becomes the
  // current action.
  current_interupt Action

  // Contains all Actions that have been registered as interrupts.  Gets
  // partially cleared out every round.
  interrupts []Action

  //
  // Ai stuff
  //
  aig_errs      chan error
  ai_done       bool
  ai_evaluating bool
}

func (l *Level) Terrain() *gui.Terrain {
  return l.terrain
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

  // Don't let us change sides while the Ai is going
  if l.Side == 2 && !l.ai_done { return }


  l.ai_done = false

  l.selected = nil
  if l.current_action != nil {
    l.current_action.Cancel()
  }
  l.current_action = nil
  l.clearCache(all_highlights)
  l.Side = l.Side%2 + 1
  if l.Side == 1 {
    l.side_gui.SetText("It is The Jungle's turn to move")
  } else {
    l.side_gui.SetText("It is The Man's turn to move")
  }

  // Any unit whose turn it is should no longer have interrupts pending, so go
  // through and remove those entities' interrupts from their slices and the
  // Level's slice of interrupts
  interrupts := make(map[Action]bool)
  for _,ent := range l.Entities {
    if ent.Side == l.Side {
      for _,interrupt := range ent.interrupts {
        interrupts[interrupt] = true
      }
      ent.interrupts = nil

      // Also, if this is an Ai controlled entity, then it should be flagged
      // as not having finished its moves for this turn.
      ent.done = false
    }
  }
  l.interrupts = algorithm.Choose(l.interrupts, func (a interface{}) bool {
    _,ok := interrupts[a.(Action)]
    return !ok
  }).([]Action)

  // Filter out dead entities
  l.Entities = algorithm.Choose(l.Entities, func (a interface{}) bool { return a.(*Entity).CurHealth() > 0 }).([]*Entity)

  for i := range l.Entities {
    if l.Entities[i].Side != l.Side {
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
    if l.Entities[i].Side != l.Side {
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

func (l *Level) findInterrupt() Action {
  var interrupt Action
  for i := range l.interrupts {
    if l.interrupts[i].Interrupt() {
      interrupt = l.interrupts[i]
      break
    }
  }
  if interrupt == nil {
    return nil
  }
  l.interrupts = algorithm.Choose(l.interrupts, func (a interface{}) bool {
    return a.(Action) != interrupt
  }).([]Action)
  return interrupt
}

// Returns true if all living ents are in the ready state, meaning that if an
// animation is applied to them that they will be able to respond to it
// immediately.
func (l *Level) allEntsReady() bool {
  for _,ent := range l.Entities {
    if ent.s.NumPendingCommands() > 0 || (ent.s.CurAnim() != "killed" && ent.s.CurAnim() != "ready") {
      return false
    }
  }
  return true
}

func (l *Level) GetAnim() string {
  for _,ent := range l.Entities {
    if ent.Name == "Nubcake" {
      return ent.s.CurAnim()
    }
  }
  panic("wtf")
}

func (l *Level) ExpSave(filename string) error {
  f,err := os.Create(filename)
  if err != nil {
    return err
  }
  defer f.Close()
  enc := gob.NewEncoder(f)
  return enc.Encode(*l)
}

func (l *Level) ExpLoad(filename string) error {
  f,err := os.Open(filename)
  if err != nil {
    return err
  }
  defer f.Close()
  dec := gob.NewDecoder(f)

  // TODO: Do we need some sort of cleanup mechanism to take down old goroutines?
  *l = Level{}
  return dec.Decode(l)
}

func (l *Level) Think(dt int64) {
  if l.allEntsReady() {
    if l.current_action == nil && l.pending_action != nil {
      l.current_action = l.pending_action
      l.pending_action = nil
    }
  }
  // If there is an action happening right now then work on that action, unless
  // that action is being interrupted, in which case work on the interrupt
  // instead.
  if l.current_action != nil {
    if l.current_interupt != nil {
      if l.current_interupt.Maintain(dt) == Complete {
        l.current_interupt = l.findInterrupt()
      }
    }
    if l.mid_action && l.current_interupt == nil {
      cont := false
//      fmt.Printf("Checking current action\n")
      switch l.current_action.Maintain(dt) {
        case Complete:
//          fmt.Printf("complete\n")
        l.selected_gui.actions.SetSelectedIndex(-1)
        l.current_action = nil
        l.mid_action = false
        cont = true
        fallthrough

        case CheckForInterrupts:
        interrupt := l.findInterrupt()
        if interrupt != nil {
          if l.current_action != nil && l.current_action.Pause() {
            l.current_interupt = interrupt
          } else {
            l.current_action = interrupt
            l.mid_action = true
          }
          if l.ai_evaluating {
//           fmt.Printf("Stopping...\n")
            l.selected.cont <- aiEvalPause
            l.ai_evaluating = false
          }
        } else if cont {
          if l.ai_evaluating {
//           fmt.Printf("Continuing...\n")
            l.selected.cont <- aiEvalCont
//            l.ai_evaluating = false
          }
        }

        case InProgress:
//          fmt.Printf("in progress\n")
      }
    }
  }
  if l.Side == 2 {
    // Do Ai stuff here
    // Right now we just go through all entities and let the act in that order,
    // TODO: Need a higher-level ai graph that determines what order to
    // activate entities in
    if l.selected != nil {
      var err error
      select {
        case f := <-l.selected.cmds:
          if !f() {
//            fmt.Printf("Not f()\n")
            l.selected.cont <- aiEvalTerm
            l.ai_evaluating = false
          }
        case err = <-l.aig_errs:
//          fmt.Printf("aig done\n")
          if err == ai.TermError {
//            fmt.Printf("terminated\n")
            l.selected.done = true
            l.selected = nil
          }
          if err == ai.InterruptError {
            l.selected = nil
          }
//          fmt.Printf("Error evaluating Ai: %v\n", err)
        default:
      }
    }
    if l.selected == nil {
      for _,ent := range l.Entities {
        if ent.Side != 2 { continue }
        if !ent.done {
          l.selected = ent
          l.ai_evaluating = true
          go func(aig *ai.AiGraph) {
//            fmt.Printf("evaluating\n")
            l.aig_errs <- aig.Eval()
//            fmt.Printf("Completed evaluation\n")
          } (l.selected.aig)
          break
        }
      }
      if l.selected == nil {
        l.ai_done = true
        defer l.Round()
      }
    }
  }

  if l.selected != nil && l.Side == 1 {
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
    pbx := int(e.Pos.X)
    pby := int(e.Pos.Y)
    e.Think(dt)
    if pbx != int(e.Pos.X) || pby != int(e.Pos.Y) {
      l.clearCache(game_highlights)
    }
    if l.grid[int(e.Pos.X)][int(e.Pos.Y)].highlight&FogOfWar == 0 {
      l.terrain.AddUprightDrawable(e.Pos.X+0.25, e.Pos.Y+0.25, e.s)
    }
  }

  l.editor.Think()

  // Highlight selected entity
  if l.selected != nil {
    cell := &l.grid[int(l.selected.Pos.X)][int(l.selected.Pos.Y)]
    cell.highlight |= Selected
  }

  bx, by := l.terrain.WindowToBoard(l.winx, l.winy)
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

  for i := range l.grid {
    for j := range l.grid[i] {
      // Draw tiles
      l.terrain.AddFlattenedDrawable(float32(i), float32(j), &l.grid[i][j])

      // Erase entities in the grid so that we can replace them with their current
      // positions, since they may have changed
      l.grid[i][j].ent = nil
    }
  }

  // Update entity positions
  for _,ent := range l.Entities {
    l.GetCellAtPos(ent.Pos).ent = ent
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
    cc.Subtract(&mathgl.Vec2{l.Entities[i].Pos.X + 0.5, l.Entities[i].Pos.Y + 0.5})
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
    if ent != nil && ent.Side == l.Side {
      l.selected = ent
    }
  } else {
    if !l.mid_action {
      switch l.current_action.MouseClick(float64(click.X), float64(click.Y)) {
        case StandardAction:
        l.mid_action = true

        case StandardInterrupt:
        l.selected.interrupts = append(l.selected.interrupts, l.current_action)
        l.interrupts = append(l.interrupts, l.current_action)
        l.current_action = nil

        case NoAction:
      }
    }
  }
}

func (l *Level) HandleEventGroup(event_group gin.EventGroup) {
  if l.Side == 2 {
    // Side == 2 is the computer, we shouldn't be doing anything according to
    // the player's input until its the player's turn again.
    return
  }

  defer l.updateDependantGuis()

  cursor := event_group.Events[0].Key.Cursor()
  if cursor == nil {
    return
  }
  l.winx, l.winy = cursor.Point()
  bx, by := l.terrain.WindowToBoard(l.winx, l.winy)
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
  Terrain base.Terrain
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

func (sld *staticLevelData) makeLevelDataContainer() *levelDataContainer {
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
    l.pending_action = l.selected.actions[n]
    l.selected_gui.actions.SetSelectedIndex(n-1)
  }
}

/*
  Directory string
  
  // whose turn it is, side 1 goes first, then 2, then back to 1...
  Side     int

  Entities []*Entity
*/

func MakeLevel(directory,mapname string) *Level {
  var level Level
  level.Directory = directory
  level.Mapname = mapname

  return &level
}

func (l *Level) Fill() error {
  mappath := filepath.Join(l.Directory, "maps", l.Mapname)
  mapfile, err := os.Open(mappath)
  if err != nil {
    return err
  }
  mapdata, err := ioutil.ReadAll(mapfile)
  if err != nil {
    return err
  }
  var ldc levelDataContainer
  err = json.Unmarshal(mapdata, &ldc)
  if err != nil {
    return err
  }

  attmap, err := stats.LoadAttributes(filepath.Join(l.Directory, "attributes.json"))
  if err != nil {
    return err
  }
  stats.SetAttmap(attmap)

  all_units, err := LoadAllUnits(l.Directory)
  if err != nil {
    return err
  }
  unit_map := make(map[string]*UnitType)
  for _, unit := range all_units {
    unit_map[unit.Name] = unit
  }

  dx := len(ldc.Level.Cells)
  dy := len(ldc.Level.Cells[0])
  all_cells := make([]CellData, dx*dy)
  l.grid = make([][]CellData, dx)
  for i := range l.grid {
    l.grid[i] = all_cells[i*dy : (i+1)*dy]
  }
  for i := range l.grid {
    for j := range l.grid[i] {
      l.grid[i][j].Terrain = ldc.Level.Cells[i][j].Terrain
      l.grid[i][j].Unit = ldc.Level.Cells[i][j].Unit
      if l.grid[i][j].Unit.Name != "" {
        if unit, ok := unit_map[l.grid[i][j].Unit.Name]; !ok {
          return errors.New(fmt.Sprintf("Unable to find unit definition for '%s'.", l.grid[i][j].Unit))
        } else {
          side := l.grid[i][j].Unit.Side
          ent := l.addEntity(*unit, i, j, side, 0.0075, unit_map)
          if ent.Side == 2 {
            err = ent.MakeAi(filepath.Join(l.Directory, "ai", "basic.xgml"))
            if err != nil {
              return err
            }
          }
        }
      }
    }
  }
  l.bg_path = ldc.Level.Image
  bg_path := filepath.Join(filepath.Clean(l.Directory), "maps", l.bg_path)
  terrain, err := gui.MakeTerrain(bg_path, 50, dx, dy, 65)
  if err != nil {
    return err
  }
  l.terrain = terrain
  terrain.SetEventHandler(l)

  l.side_gui = gui.MakeTextLine("standard", "", 500, 1, 1, 1, 1)
  l.editor = MakeEditor(&l.staticLevelData, l.Directory, l.Mapname)
  l.game_gui = gui.MakeHorizontalTable()
  game_only_gui := gui.MakeVerticalTable()
  l.selected_gui = MakeStatsWindow(true)
  l.targeted_gui = MakeStatsWindow(false)
  entity_guis := gui.MakeHorizontalTable()
  entity_guis.AddChild(l.selected_gui)
  entity_guis.AddChild(l.targeted_gui)
  game_only_gui.AddChild(l.terrain)
  game_only_gui.AddChild(l.side_gui)
  game_only_gui.AddChild(entity_guis)
  l.game_gui.AddChild(game_only_gui)
  l.editor_gui = gui.MakeCollapseWrapper(l.editor.GetGui())
  l.game_gui.AddChild(l.editor_gui)

  // Ai stuff
  l.aig_errs = make(chan error)

  return nil
}

func (l *Level) GetGui() gui.Widget {
  return l.game_gui
}

func (l *Level) ToggleEditor() {
  l.editor_gui.Collapsed = !l.editor_gui.Collapsed
}

func (l *Level) addEntity(unit_type UnitType, x, y, side int, move_speed float32, unit_map map[string]*UnitType) *Entity {
  var ent Entity
  ent = Entity{
    Name : unit_type.Name,
    Stats: stats.MakeStats(unit_type.Health, unit_type.Ap, unit_type.Attack, unit_type.Defense, unit_type.LosDist, unit_type.Atts),
    Pos:   MakeBoardPos(x, y),
    Side:  side,
    CosmeticStats: CosmeticStats{
      Move_speed: move_speed,
    },
  }
  ent.fill(l, unit_map)

  l.Entities = append(l.Entities, &ent)
  l.grid[x][y].ent = &ent
  return &ent
}
