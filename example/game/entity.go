package game

import (
  "glop/ai"
  "yed"
  "polish"
  "game/stats"
  "math"
  "fmt"
  "glop/gui"
  "glop/sprite"
  "encoding/json"
  "io/ioutil"
  "os"
  "path/filepath"
  "strings"
)

// contains the stats used to intialize a unit of this type
type UnitType struct {
  Name string

  // All of the stats that are needed to create a stats.Stats interface
  Health  int
  Ap      int
  Attack  int
  Defense int
  LosDist int
  Atts    []string

  // Name of the sprite that should be used to represent this unit
  Sprite string

  // List of the names of the weapons this unit comes with
  Weapons []string
}

type CosmeticStats struct {
  // in board coordinates per ms
  Move_speed float32
}

type EntityStatsWindow struct {
  gui.EmbeddedWidget
  gui.BasicZone
  gui.NonResponder
  gui.NonFocuser

  ent     *Entity
  table   *gui.VerticalTable
  image   *gui.ImageBox
  name    *gui.TextLine
  health  *gui.TextLine
  ap      *gui.TextLine
  actions *gui.SelectBox

  // If this is false then events on this window will be immediately rejected
  // This is so we can have multiple windows, but only one can be used to
  // affect anything game related - so you can mouse-over units that aren't
  // under your control and see their stats, but not modify them, since they
  // aren't yours
  clickable bool
}

func MakeStatsWindow(clickable bool) *EntityStatsWindow {
  var esw EntityStatsWindow
  esw.EmbeddedWidget = &gui.BasicWidget{CoreWidget: &esw}
  esw.Request_dims.Dx = 350
  esw.Request_dims.Dy = 175
  esw.clickable = clickable

  top := gui.MakeHorizontalTable()

  esw.image = gui.MakeImageBox()
  top.AddChild(esw.image)

  esw.name = gui.MakeTextLine("standard", "", 275, 1, 1, 1, 1)
  esw.health = gui.MakeTextLine("standard", "", 275, 1, 1, 1, 1)
  esw.ap = gui.MakeTextLine("standard", "", 275, 1, 1, 1, 1)
  vert := gui.MakeVerticalTable()
  vert.AddChild(esw.name)
  vert.AddChild(esw.health)
  vert.AddChild(esw.ap)
  top.AddChild(vert)

  esw.table = gui.MakeVerticalTable()
  esw.table.AddChild(top)
  esw.actions = gui.MakeSelectImageBox([]string{}, []string{})
  esw.table.AddChild(esw.actions)

  return &esw
}

// Short-circuits the typical event-handling - if this window wasn't set to
// clickable then nothing will be able to get to it.
func (w *EntityStatsWindow) Respond(g *gui.Gui, e gui.EventGroup) bool {
  if w.clickable {
    return w.table.Respond(g, e)
  }
  return false
}
func (w *EntityStatsWindow) String() string {
  return "entity stats window"
}
func (w *EntityStatsWindow) update() {
  if w.ent == nil {
    return
  }
  w.health.SetText(fmt.Sprintf("Health: %d/%d", w.ent.CurHealth(), w.ent.BaseHealth()))
  w.ap.SetText(fmt.Sprintf("Ap: %d/%d", w.ent.CurAp(), w.ent.BaseAp()))
}
func (w *EntityStatsWindow) DoThink(int64, bool) {
  if w.ent == nil {
    return
  }
  w.update()
}
func (w *EntityStatsWindow) GetEntity() *Entity {
  return w.ent
}
func (w *EntityStatsWindow) SetEntity(e *Entity) {
  if e == w.ent {
    return
  }
  w.ent = e

  w.health.SetText("")
  w.ap.SetText("")
  w.name.SetText("")
  w.image.UnsetImage()
  w.table.RemoveChild(w.actions)
  if e != nil {
    thumb := e.s.Thumbnail()
    w.image.SetImageByTexture(thumb.Texture(), thumb.Dx(), thumb.Dy())
    w.name.SetText(e.Name)
    var paths, names []string
    for i := 1; i < len(e.actions); i++ {
      paths = append(paths, filepath.Join(e.level.Directory, "icons", e.actions[i].IconPath()))
      names = append(names, e.actions[i].IconPath())
    }
    w.actions = gui.MakeSelectImageBox(paths, names)
    w.table.AddChild(w.actions)
    w.actions.SetSelectedIndex(-1)
    w.update()
  }
}
func (w *EntityStatsWindow) GetChildren() []gui.Widget {
  return []gui.Widget{w.table}
}
func (w *EntityStatsWindow) Draw(region gui.Region) {
  w.Render_region = region
  w.table.Draw(region)
}

type Entity struct {
  // *** Begin exported fields
  // When saving the game all entities are gobbed, so all exported fields are
  // saved, everything else is not.  All information must either be in these
  // fields, be derived from these fields, or be implicit in the fact that
  // saves are only allowed at certain times (not while an action is executing).
  Name string
  stats.Stats
  CosmeticStats

  // 0 indicates that the unit is unaffiliated
  Side int

  // Board coordinates of this entity's current position
  Pos BoardPos
  // *** End exported fields


  s *sprite.Sprite

  level *Level

  // Board coordinates of this entity's previous position, needed to update the
  // level grid as this entity moves around the board.
  prev_pos BoardPos

  // set of vertices that this unit can see from its current location
  visible map[int]bool

  actions []Action

  // If the Entity has used an interrupt action it will be put into this slice.
  interrupts []Action

  //
  // AI stuff
  //
  aig  *ai.AiGraph

  // Whether this entity is done with its actions for this round
  done bool

  // If the entity needs to execute an action it sends a closure along this
  // channel that starts up everything it needs.
  cmds chan func() bool

  // For every closure sent along cmds, a response will be sent back along this
  // channel.  If true the ai can keep working, if false it should terminate.
  cont chan aiEvalSignal
}

type gameError struct {
  ErrorString string
}
func (g *gameError) Error() string {
  return g.ErrorString
}

// fill() should be called immediately after an Entity is ungobbed.  This will
// fill out any fields that are not stored by gob but can be determined from
// the values stored by gob (such as actions and visible).
func (e *Entity) fill(level *Level, units map[string]*UnitType) error {
  e.prev_pos = MakeBoardPos(e.Pos.Xi(), e.Pos.Yi())
  unit, ok := units[e.Name]
  if !ok {
    return &gameError{ fmt.Sprintf("Unknown unit type '%s'", e.Name) }
  }
  var err error
  e.s, err = sprite.LoadSprite(filepath.Join(level.Directory, "sprites", unit.Sprite))
  if err != nil {
    return err
  }
  e.level = level
  e.cmds = make(chan func() bool)
  e.cont = make(chan aiEvalSignal)

  e.actions = append(e.actions, MakeAction("move", e))
  for _, name := range unit.Weapons {
    e.actions = append(e.actions, MakeAction(name, e))
  }
  return nil
}

func (e *Entity) MakeAi(filename string) error {
  doc,err := yed.ParseFromFile(filename)
  if err != nil {
    return err
  }
  e.aig = ai.NewGraph()
  e.aig.Graph = &doc.Graph
  e.aig.Context = polish.MakeContext()
  polish.AddIntMathContext(e.aig.Context)
  AddEntityContext(e, e.aig.Context)
  return nil
}

func bresenham(x, y, x2, y2 int) [][2]int {
  dx := x2 - x
  if dx < 0 {
    dx = -dx
  }
  dy := y2 - y
  if dy < 0 {
    dy = -dy
  }

  var ret [][2]int
  steep := dy > dx
  if steep {
    x, y = y, x
    x2, y2 = y2, x2
    dx, dy = dy, dx
    ret = make([][2]int, dy)[0:0]
  } else {
    ret = make([][2]int, dx)[0:0]
  }

  err := dx >> 1
  cy := y

  xstep := 1
  if x2 < x {
    xstep = -1
  }
  ystep := 1
  if y2 < y {
    ystep = -1
  }
  for cx := x; cx != x2; cx += xstep {
    if !steep {
      ret = append(ret, [2]int{cx, cy})
    } else {
      ret = append(ret, [2]int{cy, cx})
    }
    err -= dy
    if err < 0 {
      cy += ystep
      err += dx
    }
  }
  if !steep {
    ret = append(ret, [2]int{x2, cy})
  } else {
    ret = append(ret, [2]int{cy, x2})
  }
  return ret
}

// TODO: Might be able to see to the other side of the map because we're just
// converting the points to vertices and not actually checking that they lie
// within the bounds of the map.
func (e *Entity) addVisibleAlongLine(vision int, line [][2]int) {
  for _, v := range line {
    e.visible[e.level.toVertex(v[0], v[1])] = true
    concealment := e.Concealment(e.level.grid[v[0]][v[1]].Terrain)
    if concealment < 0 {
      break
    }
    vision -= concealment + 1
    if vision <= 0 {
      break
    }
  }
}

func (e *Entity) figureVisibility() {
  vision := e.CurLosDist(e.level.GetCellAtPos(e.Pos).Terrain)
  ex := int(e.Pos.X)
  ey := int(e.Pos.Y)

  x := ex - vision
  if x < 0 {
    x = 0
  }
  y := ey - vision
  if y < 0 {
    y = 0
  }

  x2 := ex + vision
  if x2 >= len(e.level.grid) {
    x2 = len(e.level.grid) - 1
  }
  y2 := ey + vision
  if y2 >= len(e.level.grid[0]) {
    y2 = len(e.level.grid[0]) - 1
  }

  e.visible = make(map[int]bool, vision*vision)
  e.visible[e.level.toVertex(ex, ey)] = true
  for cx := x; cx <= x2; cx++ {
    e.addVisibleAlongLine(vision, bresenham(ex, ey, cx, y)[1:])
    e.addVisibleAlongLine(vision, bresenham(ex, ey, cx, y2)[1:])
  }
  for cy := y; cy <= y2; cy++ {
    e.addVisibleAlongLine(vision, bresenham(ex, ey, x, cy)[1:])
    e.addVisibleAlongLine(vision, bresenham(ex, ey, x2, cy)[1:])
  }
}

func (e *Entity) Coords() (x, y int) {
  return int(e.Pos.X), int(e.Pos.Y)
}

func (e *Entity) OnSetup() {
  e.Stats.Setup()
  e.figureVisibility()
}
func (e *Entity) OnRound() {
  e.Stats.Round()
}
func (e *Entity) AddEffect(effect stats.Effect) {
  e.Stats.AddEffect(effect, e.Side == e.level.Side)
}
func (e *Entity) CurAttack() int {
  return e.Stats.CurAttack(e.level.GetCellAtPos(e.Pos).Terrain)
}
func (e *Entity) CurDefense() int {
  return e.Stats.CurDefense(e.level.GetCellAtPos(e.Pos).Terrain)
}
// TODO: This is the method that should determine if something triggered as we
// moved into a cell.  It will also need to return this information to the
// caller, who can decide how to proceed.  There should be a very limited
// number of triggers that can happend.
func (e *Entity) OnEntry() {
  e.figureVisibility()
}

// Advances the entity up to max_dist towards position bx,by (in board
// coordinates).
// If an advance causes the entity to enter a new cell Advance will stop.
// This gives things that trigger on entering a cell a chance to respond
// before allowing the entity to contine moving.
// Returns the distance moved, will never be more than 1.0.
// This function may return 0 even when max_dist > 0, this indicates that
// the sprite was not prepared and the caller should wait before trying to
// Advance again.
// Also returns a bool indicating whether or not the target cell has been
// reached.
func (e *Entity) Advance(bp BoardPos, max_dist float32) (float32, bool) {
  if max_dist < 0 {
    panic("Tried to advance negative distance")
  }

  if max_dist == 0 {
    if e.s.CurState() != "ready" {
      e.s.Command("stop")
    }
    return 0, false
  }

  if e.s.CurState() != "walk" {
    e.s.Command("move")
  }

  // Wait until the sprite is actually walking to move it, otherwise it looks
  // like it's sliding
  if e.s.CurAnim() != "walk" {
    return 0, false
  }

  src := e.Pos
  dst := bp
  dst.Subtract(&src.Vec2)
  dist := dst.Length()

  // If we can reach the target cell then we can just set our coordinates and
  // return, the caller can decide whether or not to continue.
  if dist <= max_dist {
    e.Pos.Assign(&bp.Vec2)
    e.level.GetCellAtPos(e.prev_pos).ent = nil
    e.level.GetCellAtPos(e.Pos).ent = e
    e.prev_pos = e.Pos
    return dist, true
  }

  dst.Normalize()
  dst.Vec2.Scale(max_dist)
  src.Vec2.Add(&dst.Vec2)
  e.Pos.Assign(&src.Vec2)
  e.turnToFace(bp)
  return max_dist, false
}

func (e *Entity) turnToFace(dst BoardPos) {
  dst.Subtract(&e.Pos.Vec2)
  facing := math.Atan2(float64(dst.Y), float64(dst.X)) / (2 * math.Pi) * 360.0
  var face int
  if facing >= 22.5 || facing < -157.5 {
    face = 0
  } else {
    face = 1
  }
  if face != e.s.StateFacing() {
    e.s.Command("turn_left")
  }
}

func (e *Entity) Think(dt int64) {
  e.s.Think(dt)
}

func LoadAllUnits(dir string) ([]*UnitType, error) {
  var paths []string
  unit_dir := filepath.Join(dir, "units")
  err := filepath.Walk(unit_dir, func(path string, info os.FileInfo, err error) error {
    if !info.IsDir() && strings.HasSuffix(path, ".json") {
      paths = append(paths, path)
    }
    return nil
  })
  if err != nil {
    panic(fmt.Sprintf("Error reading directory %s: %s\n", dir, err.Error()))
  }

  var units []*UnitType
  for _, path := range paths {
    f, err := os.Open(path)
    if err != nil {
      return nil, err
    }
    defer f.Close()
    data, err := ioutil.ReadAll(f)
    if err != nil {
      return nil, err
    }
    var unit UnitType
    err = json.Unmarshal(data, &unit)
    if err != nil {
      return nil, err
    }
    units = append(units, &unit)
  }
  return units, nil
}
