package gui

import (
  "glop/gin"
)

type rootWidget struct {
  BaseWidget
}

type Gui struct {
  root  rootWidget
  focus Focus
}
func (g *Gui) HandleEventGroup(event_group gin.EventGroup) {  
  g.root.HandleEventGroup(event_group)
}
func (g *Gui) Think(int64) {  
}
func (g *Gui) AddWidget(w FocusWidget) {
  g.root.AddChild(w.this())
}

// Creates the gui object and specifies the size of the screen that it will render to.  The gui
// naturally renders to the rectangular region from (0,0) to (dx,dy)
func Make(input *gin.Input, dx,dy int) *Gui {
  g := new(Gui)
  input.RegisterEventListener(g)
  return g
}


