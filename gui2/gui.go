package gui

import(
  "glop/gin"
)

type Point struct {
  X,Y int
}
func (p Point) Add(q Point) Point {
  return Point{
    X : p.X + q.X,
    Y : p.Y + q.Y,
  }
}
func (p Point) Inside(r Region) bool {
  if p.X < r.X { return false }
  if p.Y < r.Y { return false }
  if p.X > r.X + r.Dx { return false }
  if p.Y > r.Y + r.Dy { return false }
  return true
}
type Dims struct {
  Dx,Dy int
}
type Region struct {
  Point
  Dims
}
func (r Region) Add(p Point) Region {
  return Region{
    r.Point.Add(p),
    r.Dims,
  }
}
//func (r Region) setViewport() {
//  gl.Viewport(r.Point.X, r.Point.Y, r.Dims.Dx, r.Dims.Dy)
//}

type EventGroup struct {
  gin.EventGroup
  focus bool
}

type WidgetCore interface {
  Think(int64)

  // returns true if the EventGroup was consumed - i.e. it should not be passed on to its
  // children
  HandleEvents(EventGroup) bool
  Draw(Region)
  Children() []Widget
}

// Widgets are responsible for sending events to their children
// Widgets are responsible for laying out and drawing their children appropriately
type Widget interface {
  // A Timestep is sent along this channel on a regular basis.  Before this Widget reads values
  // from the think,events or kill channels it should update its dimensions
  ThinkChan() chan<- int64

  // Any events sent to this Widget will come along this channel.  This Widget is responsible
  // for forwarding these events along to its children, if appropriate.  Children should all
  // finish processing events before the next read on this channel.
  EventsChan() chan<- EventGroup

  // Sends the dimensions that this Widget would like to render itself with.  Typically called
  // between Think and Draw, but can be called other times as well.
  DimsChan() <-chan Dims

  // Draws the Widget to the specified region
  DrawChan() chan<- Region

  // After a Draw has been issued a single value should be sent along this channel when the
  // Draw is complete.
  DrawComplete() <-chan bool
}

type Drawer struct {
  draw_chan     chan Region
  draw_complete chan bool
}
func (w *Drawer) drawChan() chan Region {
  return w.draw_chan
}
func (w *Drawer) drawComplete() chan bool {
  return w.draw_complete
}
func (w *Drawer) drawSetup() {
  w.draw_chan = make(chan Region)
  w.draw_complete = make(chan bool)
}

type Dimser struct {
  dims_chan chan Dims
  Dims Dims
}
func (w *Dimser) dimsChan() chan Dims {
  return w.dims_chan
}
func (w *Dimser) dimsSetup() {
  w.dims_chan = make(chan Dims)
}
func (w *Dimser) getDims() Dims {
  return w.Dims
}

type NonThinker struct {}
func (w *NonThinker) thinkChan() chan int64 { return nil }
func (w *NonThinker) doThink(core WidgetCore, timestamp int64) {}
func (w *NonThinker) thinkSetup() {}

type Thinker struct {
  think_chan chan int64
}
func (w *Thinker) thinkChan() chan int64 {
  return w.think_chan
}
func (w *Thinker) doThink(core WidgetCore, timestamp int64) {
  for i := range core.Children() {
    c := core.Children()[i].ThinkChan()
    if c == nil { continue }
    c <- timestamp
  }
  core.Think(timestamp)
}
func (w *Thinker) thinkSetup() {
  w.think_chan = make(chan int64)
}

type NonResponder struct {}
func (w *NonResponder) eventsChan() chan EventGroup { return nil }
func (w *NonResponder) doEvents(core WidgetCore, events EventGroup) {}
func (w *NonResponder) eventsSetup() {}

type Responder struct {
  events_chan chan EventGroup
}
func (w *Responder) eventsChan() chan EventGroup {
  return w.events_chan
}
func (w *Responder) doEvents(core WidgetCore, events EventGroup) {
  if !core.HandleEvents(events) {
    for i := range core.Children() {
      c := core.Children()[i].EventsChan()
      if c == nil { continue }
      c <- events
    }
  }
}
func (w *Responder) eventsSetup() {
  w.events_chan = make(chan EventGroup)
}

type internalWidget interface {
  thinkChan() chan int64
  thinkSetup()
  doThink(WidgetCore, int64)

  eventsChan() chan EventGroup
  eventsSetup()
  doEvents(WidgetCore, EventGroup)

  dimsChan() chan Dims
  dimsSetup()
  getDims() Dims

  drawChan() chan Region
  drawComplete() chan bool
  drawSetup()
}
type externalWidget struct {
  internalWidget
}
func (w *externalWidget) ThinkChan() chan<- int64 {
  return w.thinkChan()
}
func (w *externalWidget) EventsChan() chan<- EventGroup {
  return w.eventsChan()
}
func (w *externalWidget) DimsChan() <-chan Dims {
  return w.dimsChan()
}
func (w *externalWidget) DrawChan() chan<- Region {
  return w.drawChan()
}
func (w *externalWidget) DrawComplete() <-chan bool {
  return w.drawComplete()
}

type ThinkerResponder struct {
  externalWidget
}
type thinkerResponder struct {
  Thinker
  Responder
  Dimser
  Drawer
}
func (w *ThinkerResponder) Startup(core WidgetCore) {
  w.externalWidget.internalWidget = &thinkerResponder{}
  startup(w, core)
}

func startup(w internalWidget, core WidgetCore) {
  // TODO: make sure we don't startup twice
  w.thinkSetup()
  w.eventsSetup()
  w.drawSetup()
  w.dimsSetup()
  go func() {
    for {
      select {
        case events := <-w.eventsChan():
          w.doEvents(core, events)

        case think := <-w.thinkChan():
          w.doThink(core, think)

        case region := <-w.drawChan():
          core.Draw(region)
          w.drawComplete() <- true

        case w.dimsChan() <- w.getDims():
      }
    }
  } ()
}
