package gos

// #cgo LDFLAGS: -Llinux/lib -lglop -lX11 -lGL
// #include "linux/include/glop.h"
import "C"

import (
	"fmt"
	"github.com/runningwild/glop/gin"
	"github.com/runningwild/glop/system"
	"os"
	"sort"
	"sync"
	"time"
	"unsafe"
)

type linuxSystemObject struct {
	horizon int64
}

var (
	linux_system_object linuxSystemObject
	jsCollect           chan jsInput
)

// Call after runtime.LockOSThread(), *NOT* in an init function
func (linux *linuxSystemObject) Startup() {
	C.GlopInit()
	jsCollect = make(chan jsInput, 100)
	go trackJoysticks(jsCollect)
}

func GetSystemInterface() system.Os {
	return &linux_system_object
}

func (linux *linuxSystemObject) Run() {
	panic("Not implemented on linux")
}

func (linux *linuxSystemObject) Quit() {
	panic("Not implemented on linux")
}

func (linux *linuxSystemObject) CreateWindow(x, y, width, height int) {
	C.GlopCreateWindow(unsafe.Pointer(&(([]byte("linux window"))[0])), C.int(x), C.int(y), C.int(width), C.int(height))
}

func (linux *linuxSystemObject) SwapBuffers() {
	C.GlopSwapBuffers()
}

func (linux *linuxSystemObject) Think() {
	C.GlopThink()
}

func (linux *linuxSystemObject) GetActiveDevices() map[gin.DeviceType][]gin.DeviceIndex {
	return nil
}

type jsInput struct {
	TimestampMs uint32
	Value       int16
	Type        uint8
	Key         uint8
	Index       int
}

func parsejsInput(b []byte) (jsInput, error) {
	var js jsInput
	if len(b) != 8 {
		return js, fmt.Errorf("Expected 8 bytes, got %d.", len(b))
	}
	js.TimestampMs = uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
	js.Value = int16(b[4]) | int16(b[5])<<8
	js.Type = b[6]
	js.Key = b[7]
	return js, nil
}

func pollJoysticks(index int, jsCollect chan<- jsInput, onDeath func(int)) {
	defer onDeath(index)

	f, err := os.Open(fmt.Sprintf("/dev/input/js%d", index))
	if err != nil {
		return
	}
	defer f.Close()

	buf := make([]byte, 1024)
	for {
		n, err := f.Read(buf)
		if err != nil {
			return
		}
		tmp := buf[0:n]
		for len(tmp) >= 8 {
			js, err := parsejsInput(tmp[0:8])
			if err != nil {
				continue
			}
			tmp = tmp[8:]
			js.Index = index
			jsCollect <- js
		}
	}
}

func trackJoysticks(jsCollect chan<- jsInput) error {
	defer close(jsCollect)
	var jsMutex sync.Mutex
	active := make(map[int]bool)
	onDeath := func(index int) {
		jsMutex.Lock()
		delete(active, index)
		jsMutex.Unlock()
	}
	for {
		f, err := os.Open("/dev/input")
		if err != nil {
			return err
		}
		names, err := f.Readdirnames(0)
		f.Close()
		if err != nil {
			return err
		}
		for _, name := range names {
			var index int
			_, err = fmt.Sscanf(name, "js%d", &index)
			if err != nil {
				continue
			}
			jsMutex.Lock()
			if !active[index] {
				active[index] = true
				go pollJoysticks(index, jsCollect, onDeath)
			}
			jsMutex.Unlock()
		}
		time.Sleep(time.Second * 5)
	}
	return nil
}

type osEventSlice []gin.OsEvent

func (oes osEventSlice) Len() int           { return len(oes) }
func (oes osEventSlice) Swap(i, j int)      { oes[i], oes[j] = oes[j], oes[i] }
func (oes osEventSlice) Less(i, j int) bool { return oes[i].Timestamp < oes[j].Timestamp }

// TODO: Make sure that events are given in sorted order (by timestamp)
// TODO: Adjust timestamp on events so that the oldest timestamp is newer than the
//       newest timestemp from the events from the previous call to GetInputEvents
//       Actually that should be in system
func (linux *linuxSystemObject) GetInputEvents() ([]gin.OsEvent, int64) {
	var first_event *C.GlopKeyEvent
	cp := (*unsafe.Pointer)(unsafe.Pointer(&first_event))
	var length C.int
	var horizon C.longlong
	C.GlopGetInputEvents(cp, unsafe.Pointer(&length), unsafe.Pointer(&horizon))
	linux.horizon = int64(horizon)
	c_events := (*[1000]C.GlopKeyEvent)(unsafe.Pointer(first_event))[:length]
	events := make([]gin.OsEvent, length)
	for i := range c_events {
		events[i] = gin.OsEvent{
			KeyId: gin.KeyId{
				Device: gin.DeviceId{
					Index: 5,
					Type:  gin.DeviceTypeKeyboard,
				},
				Index: gin.KeyIndex(c_events[i].index),
			},
			Press_amt: float64(c_events[i].press_amt),
			Timestamp: int64(c_events[i].timestamp),
		}
	}
	for {
		select {
		case event := <-jsCollect:
			events = append(events, gin.OsEvent{
				KeyId: gin.KeyId{
					Device: gin.DeviceId{
						Index: gin.DeviceIndex(event.Index),
						Type:  gin.DeviceTypeController,
					},
					Index: gin.KeyIndex(event.Key),
				},
				Press_amt: float64(event.Value),
				Timestamp: int64(event.TimestampMs),
			})
		default:
			break
		}
	}
	sort.Sort(osEventSlice(events))
	return events, linux.horizon
	// return nil, 0
}

func (linux *linuxSystemObject) HideCursor(hide bool) {
}

func (linux *linuxSystemObject) rawCursorToWindowCoords(x, y int) (int, int) {
	wx, wy, _, _ := linux.GetWindowDims()
	return x - wx, y - wy
}

func (linux *linuxSystemObject) GetCursorPos() (int, int) {
	var x, y C.int
	C.GlopGetMousePosition(&x, &y)
	return linux.rawCursorToWindowCoords(int(x), int(y))
}

func (linux *linuxSystemObject) GetWindowDims() (int, int, int, int) {
	var x, y, dx, dy C.int
	C.GlopGetWindowDims(&x, &y, &dx, &dy)
	return int(x), int(y), int(dx), int(dy)
}

func (linux *linuxSystemObject) EnableVSync(enable bool) {
	var _enable C.int
	if enable {
		_enable = 1
	}
	C.GlopEnableVSync(_enable)
}

func (linux *linuxSystemObject) HasFocus() bool {
	// TODO: Implement me!
	return true
}
