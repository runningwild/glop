// Only windows actually requires sys.Think() to be on the render thread, and
// darwin will hang it this is done, so for everything other than windows the
// system thinking logic is done when sys.Think() is called.
// +build !windows

package system

func (sys *sysObj) Think() {
	sys.thinkInternal()
}
