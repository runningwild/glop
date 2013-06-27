package gos

// #cgo LDFLAGS: -Ldarwin/lib -lglop -framework Cocoa -framework IOKit -framework OpenGL -mmacosx-version-min=10.5
// #include "darwin/include/glop.h"
import "C"

import (
	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/system"
	"unsafe"
)

type osxSystemObject struct {
	window  uintptr // NSWindow*
	context uintptr // NSOpenGLContext*
	horizon int64
}

var (
	osx_system_object osxSystemObject
)

// Call after runtime.LockOSThread(), *NOT* in an init function
func (osx *osxSystemObject) Startup() {
	C.Init()
}

func GetSystemInterface() system.Os {
	return &osx_system_object
}

func (osx *osxSystemObject) Run() {
	C.Run()
}

func (osx *osxSystemObject) Quit() {
	C.Quit()
}

func (osx *osxSystemObject) CreateWindow(x, y, width, height int) {
	w := (*unsafe.Pointer)(unsafe.Pointer(&osx.window))
	c := (*unsafe.Pointer)(unsafe.Pointer(&osx.context))
	C.CreateWindow(w, c, C.int(x), C.int(y), C.int(width), C.int(height))
}

func (osx *osxSystemObject) SwapBuffers() {
	C.SwapBuffers(unsafe.Pointer(osx.context))
}

func (osx *osxSystemObject) Think() {
	C.Think()
}

func (osx *osxSystemObject) GetActiveDevices() map[gin.DeviceType][]gin.DeviceIndex {
	var first_device_id *C.DeviceId
	fdip := (*unsafe.Pointer)(unsafe.Pointer(&first_device_id))
	var length C.int
	C.GetActiveDevices(fdip, &length)
	c_ids := (*[1000]C.DeviceId)(unsafe.Pointer(first_device_id))[:length]
	ret := make(map[gin.DeviceType][]gin.DeviceIndex, length)
	for _, c_id := range c_ids {
		dt := gin.DeviceType(c_id.Type)
		di := gin.DeviceIndex(c_id.Index)
		ret[dt] = append(ret[dt], di)
	}
	return ret
}

// TODO: Make sure that events are given in sorted order (by timestamp)
// TODO: Adjust timestamp on events so that the oldest timestamp is newer than the
//       newest timestemp from the events from the previous call to GetInputEvents
//       Actually that should be in system
func (osx *osxSystemObject) GetInputEvents() ([]gin.OsEvent, int64) {
	var first_event *C.KeyEvent
	cp := (*unsafe.Pointer)(unsafe.Pointer(&first_event))
	var length C.int
	var horizon C.longlong
	C.GetInputEvents(cp, &length, &horizon)
	osx.horizon = int64(horizon)
	c_events := (*[1000]C.KeyEvent)(unsafe.Pointer(first_event))[:length]
	events := make([]gin.OsEvent, length)
	for i := range c_events {
		var device_type gin.DeviceType
		switch c_events[i].device_type {
		case C.deviceTypeKeyboard:
			device_type = gin.DeviceTypeKeyboard
		case C.deviceTypeMouse:
			device_type = gin.DeviceTypeMouse
		case C.deviceTypeController:
			device_type = gin.DeviceTypeController
		default:
			panic("Unknown device type")
		}
		events[i] = gin.OsEvent{
			KeyId: gin.KeyId{
				Device: gin.DeviceId{
					Index: gin.DeviceIndex(c_events[i].device_index),
					Type:  device_type,
				},
				Index: gin.KeyIndex(c_events[i].key_index),
			},
			Press_amt: float64(c_events[i].press_amt),
			Timestamp: int64(c_events[i].timestamp) / 1000000,
		}
	}
	return events, osx.horizon
}

func (osx *osxSystemObject) rawCursorToWindowCoords(x, y int) (int, int) {
	wx, wy, _, _ := osx.GetWindowDims()
	return x - wx, y - wy
}

func (osx *osxSystemObject) GetCursorPos() (int, int) {
	var x, y C.int
	C.GetMousePos(&x, &y)
	return osx.rawCursorToWindowCoords(int(x), int(y))
}

func (osx *osxSystemObject) HideCursor(hide bool) {
	if hide {
		C.LockCursor(1)
		C.HideCursor(1)
	} else {
		C.LockCursor(0)
		C.HideCursor(0)
	}
}

func (osx *osxSystemObject) GetWindowDims() (int, int, int, int) {
	var x, y, dx, dy C.int
	C.GetWindowDims(unsafe.Pointer(osx.window), &x, &y, &dx, &dy)
	return int(x), int(y), int(dx), int(dy)
}

func (osx *osxSystemObject) EnableVSync(enable bool) {
	var _enable C.int
	if enable {
		_enable = 1
	}
	C.EnableVSync(unsafe.Pointer(osx.context), _enable)
}

func (osx *osxSystemObject) HasFocus() bool {
	var has_focus C.int
	C.HasFocus(&has_focus)
	return bool(has_focus == 1)
}
