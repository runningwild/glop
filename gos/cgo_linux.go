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
	"strings"
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

	// This is read from an 8-bit value, but glop supports more values so it
	// gets read into a 32 bit int.
	Key uint32

	// Since we translate axis values into floats we want this value here for that.
	FValue float64

	Index int
}

const kControllerButton0 = 500
const kControllerAxis0Positive = 70000
const kControllerAxis0Negative = 80000
const kControllerHatSwitchUp = 90000
const kControllerHatSwitchUpRight = 90001
const kControllerHatSwitchRight = 90002
const kControllerHatSwitchDownRight = 90003
const kControllerHatSwitchDown = 90004
const kControllerHatSwitchDownLeft = 90005
const kControllerHatSwitchLeft = 90006
const kControllerHatSwitchUpLeft = 90007

func parsejsInput(b []byte) (jsInput, error) {
	var js jsInput
	if len(b) != 8 {
		return js, fmt.Errorf("Expected 8 bytes, got %d.", len(b))
	}
	js.TimestampMs = uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
	js.Value = int16(b[4]) | int16(b[5])<<8
	js.Type = b[6]
	js.Key = uint32(b[7])
	return js, nil
}

func pollJoysticks(name string, index int, jsCollect chan<- jsInput, onDeath func(string)) {
	defer onDeath(name)

	f, err := os.Open("/dev/input/by-path/" + name)
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
			switch js.Type {
			case 1: // A regular button
				js.Key += kControllerButton0
				js.FValue = float64(js.Value)

			case 2: // An Axis value
				if js.Value < 0 {
					js.Key += kControllerAxis0Negative
					js.FValue = float64(-js.Value) / 32768
				} else {
					js.Key += kControllerAxis0Positive
					js.FValue = float64(js.Value) / 32768
				}

			default:
				continue
			}
			js.Index = index
			jsCollect <- js
		}
	}
}

func trackJoysticks(jsCollect chan<- jsInput) error {
	defer close(jsCollect)
	var jsMutex sync.Mutex
	active := make(map[string]bool)
	onDeath := func(name string) {
		jsMutex.Lock()
		delete(active, name)
		jsMutex.Unlock()
	}
	index := 0
	for {
		f, err := os.Open("/dev/input/by-path")
		if err != nil {
			return err
		}
		names, err := f.Readdirnames(0)
		f.Close()
		if err != nil {
			return err
		}
		for _, name := range names {
			if strings.Contains(name, "event") {
				continue
			}
			jsMutex.Lock()
			if !active[name] {
				active[name] = true
				index++
				go pollJoysticks(name, index, jsCollect, onDeath)
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
	done := false
	for !done {
		select {
		case event := <-jsCollect:
			events = append(events, gin.OsEvent{
				KeyId: gin.KeyId{
					Device: gin.DeviceId{
						Index: gin.DeviceIndex(event.Index + 1),
						Type:  gin.DeviceTypeController,
					},
					Index: gin.KeyIndex(event.Key),
				},
				Press_amt: event.FValue,
				Timestamp: int64(event.TimestampMs),
			})
		default:
			done = true
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
