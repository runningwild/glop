package game

func init() {
  registerActionType("linear shift", &ActionLinearShift{})
}
type ActionLinearShift struct {
  basicIcon
  nonInterrupt
  uninterruptable
  Ent       *Entity

  // Duh
  Cost  int

  // The maximum distance a unit can be and be affected by this ability
  Range int

  // The distance a unit can be pulled by this ability
  Pull  int

  // The distance a unit can be pushed by this ability
  Push  int

  // Entitys that can be targeted by this ability
  targets   []*Entity

  // Positions to which this entity can be moved by this ability
  shift_targets []BoardPos

  // MouseOver target, so we don't have to keep recalculating these lines
  hover_target *BoardPos

  // Selected target, if this is not nil it indicates that the user can now
  // select a shift target to commit to this action
  current_target *Entity

  // The shift target that was chosen to move this entity to
  shift_target BoardPos
}

func (a *ActionLinearShift) Prep() bool {
  if a.Ent.CurAp() < a.Cost {
    return false
  }

  a.targets = getEntsWithinRange(a.Ent, a.Range, a.Ent.level)
  if len(a.targets) == 0 {
    return false
  }

  for _,target := range a.targets {
    a.Ent.level.GetCellAtPos(target.pos).highlight |= Attackable
  }
  return true
}

func (a *ActionLinearShift) Cancel() {
  a.targets = nil
  a.hover_target = nil
  a.current_target = nil
  a.Ent.level.clearCache(Attackable | Reachable)
}

func (a *ActionLinearShift) MouseOver(bx,by float64) {
  if a.current_target != nil { return }
  var target *Entity
  target_pos := MakeBoardPos(int(bx), int(by))

  // If this is the same as our current target then we have already highlighted
  // the appropriate cells and cached the appropriate data
  if a.hover_target != nil && a.hover_target.IntEquals(target_pos) {
    return
  }
  a.Ent.level.clearCache(Reachable)

  for _,ent := range a.targets {
    if ent.pos.IntEquals(target_pos) {
      target = ent
      break
    }
  }
  if target == nil {
    return
  }

  dist := a.Ent.pos.Dist(target_pos)
  if dist > a.Range {
    return
  }

  a.figureShiftTargets(target_pos)
  a.hover_target = &target_pos
  for _,pos := range a.shift_targets {
    a.Ent.level.GetCellAtPos(pos).highlight |= Reachable
  }
}

func (a *ActionLinearShift) figureShiftTargets(target_pos BoardPos) {
  a.shift_targets = nil

  // For pulling we draw a line from the target to our position
  pull_line := bresenham(target_pos.Xi(), target_pos.Yi(), a.Ent.pos.Xi(), a.Ent.pos.Yi())

  // For pushing we draw a line from the target directly away from us
  far := target_pos.Sub(a.Ent.pos).Scale(a.Push)
  far = far.Add(a.Ent.pos)
  push_line := bresenham(target_pos.Xi(), target_pos.Yi(), far.Xi(), far.Yi())

  for _,line := range [][][2]int{ pull_line, push_line } {
    for i,pos := range line {
      if i == 0 { continue }
      if i > a.Pull { break }
      if pos[0] < 0 || pos[1] < 0 { break }
      if pos[0] >= len(a.Ent.level.grid) { break }
      if pos[1] >= len(a.Ent.level.grid[0]) { break }
      if a.Ent.level.grid[pos[0]][pos[1]].ent != nil { break }
      bp := MakeBoardPos(pos[0], pos[1])
      a.Ent.level.GetCellAtPos(bp).highlight |= Reachable
      a.shift_targets = append(a.shift_targets, bp)
    }
  }
}


func (a *ActionLinearShift) MouseClick(bx,by float64) ActionCommit {
  bp := MakeBoardPos(int(bx), int(by))
  // If current_target is nil then the user hasn't selected the target entity
  // for this action
  if a.current_target == nil {
    for _,target := range a.targets {
      if target.pos.IntEquals(bp) {
        a.current_target = target
        break
      }
    }
    if a.current_target == nil {
      return NoAction
    }
  } else {
    // If current_target isn't nil then the user might be selecting a cell to
    // shift current_target to
    for _,target := range a.shift_targets {
      if target.IntEquals(bp) {
        // Start the action
        a.shift_target = target
        if a.Ent.CurAp() < a.Cost {
          a.Cancel()
          return NoAction
        }
        a.Ent.SpendAp(a.Cost)
        return StandardAction
      }
    }

    // If we got here then maybe the user is selecting a new current_target
    ct := a.Ent.level.GetCellAtPos(bp).ent
    if ct != nil {
      a.current_target = ct
      a.Ent.level.clearCache(Attackable | Reachable)
      a.figureShiftTargets(a.current_target.pos)
      a.hover_target = &a.current_target.pos
      for _,pos := range a.shift_targets {
        a.Ent.level.GetCellAtPos(pos).highlight |= Reachable
      }
      a.Ent.level.GetCellAtPos(a.current_target.pos).highlight |= Attackable
    }
  }
  return NoAction
}

func (a *ActionLinearShift) Maintain(dt int64) MaintenanceStatus {
  path := []BoardPos{ a.shift_target }
  if AdvanceEntity(a.current_target, &path, dt) {
    a.Cancel()
    return Complete
  }
  return InProgress
}
