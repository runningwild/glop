package game

type basicIcon struct {
  path string
}
func (b basicIcon) Icon() string {
  return b.path
}

// An Action represents anything that a unit can spend AP to do, move, attack,
// charge, anything where the user will have to use some sort of GUI to make
// it explicit how to perform the action, etc...
type Action interface {
  // Path to an icon that can be used to represent this action, except for movement
  // this icon will be shown in a unit's stat window
  Icon() string

  // Run when the player selects this action but before he commits to it.
  // Performs any pre-processing desired and highlights anything necessary or
  // creates any GUI necessary for the user to decide how to proceed.  Returns
  // true if this action could be prepped, in which case it is available for 
  // an immediate call to Do() or Maintain().
  Prep() bool

  // Run if Prep() was selected but the user decided not to perform the action.
  // Should tear down any GUI created and undo anything else no long necessary
  // to keep around.  Can be called redundantly.
  Cancel()

  // Called regularly after Prep() but before Do() and gives the action a
  // chance to modify any UI in response to mouse location.
  // bx and by are board coordinates
  MouseOver(bx,by float64)

  // Called after Prep() and indicates that the user clicked at a particular
  // location.  This function should return true if this click means that the
  // user has chosen to commit to this action.
  // bx and by are board coordinates
  MouseClick(bx,by float64) bool

  // Called periodically after Prep() and after the action has been committed.
  // The method should return false until the action is complete, at which
  // point it should return true.  After this method returns true the Mouse*()
  // methods and Maintain() will not be called again until the action has been
  // prepped again.
  Maintain(dt int64) bool
}
