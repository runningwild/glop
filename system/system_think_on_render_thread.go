// Windows requires that event handling be done on the same thread as opengl,
// so on windows sys.Think() actually queues up that logic to run on the render
// thread.
// +build windows

package system

import (
	"github.com/runningwild/glop/render"
)

func (sys *sysObj) Think() {
	render.Queue(func() {
		sys.thinkInternal()
	})
}
