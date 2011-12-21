package render

import (
  "runtime"
)

var (
  render_funcs chan func()
  purge chan bool
)

func init() {
  render_funcs = make(chan func(), 1000)
  purge = make(chan bool)
}

// Queues a function to run on the render thread
func Queue(f func()) {
  render_funcs <- f
}

// Waits until all render thread functions have been run
func Purge() {
  purge <- true
  <-purge
}

func Init() {
  go func() {
    runtime.LockOSThread()
    for {
      select {
        case f := <-render_funcs:
          f()
        case <-purge:
          for {
            select {
              case f := <-render_funcs:
                f()
              default:
              goto purged
            }
          }
          purged:
          purge <- true
      }
    }
  } ()
}
