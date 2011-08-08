// TODO(jwills): Sometimes joysticks can be left generating axis events even though it's centered
// TODO(jwills): Fill out the file system functions
// TODO(jwills): Either figure out a new model for joysticks, or make the joysticks not update
//               automatically
// TODO(jwills): Mouse wheel events aren't working for some reason.
// TODO(jwills): Event times, if the double value is more than an int32 can handle, asplode.
//#ifdef MACOSX
//
//#include "Os.h"
//#include "Input.h"
//#include <vector>
//#include <string>
//#include <set>
//#include <map>
//#include <algorithm>
//#include <mach-o/dyld.h>
//using namespace std;
//
#include <Carbon/Carbon.h>
//#include <OpenGL/OpenGL.h>
#include <OpenGL/gl.h>
#include <AGL/agl.h>
#include <stdio.h>

#include <ApplicationServices/ApplicationServices.h>
#include <IOKit/hid/IOHIDLib.h>
//
const int kEventClassGlop = 'Glop';
const int kEventGlopBreak = 0;
const int kEventGlopToggleFullScreen = 'Flsc';
//
//static set<OsWindowData*> all_windows;
//static HIPoint mouse_location;
//static HIPoint mouse_delta;
//static map<int,int> glop_key_map;
//
//const GlopKey kKeyGUI = 0;
//const GlopKey kKeyPadClear = 0;
//const GlopKey kKeyHelp = 0;
//
//
//const GlopKey key_map[] = {
//  'a', 's', 'd', 'f', 'h',
//  'g', 'z', 'x', 'c', 'v',
//    0, 'b', 'q', 'w', 'e',
//  'r', 'y', 't', '1', '2',
//  '3', '4', '6', '5', '!',
//  '9', '7', '-', '8', '0',
//  ']', 'o', 'u', '[', 'i',
//  'p', kKeyEnter, 'l', 'j', '\'',
//  'k', ';', '\\', ',', '/',
//  'n', 'm', '.', kKeyTab, ' ',
//  '`', kKeyBackspace, 0, kKeyEscape, 0,             // 50
//  kKeyGUI, 0, 0, 0, 0,
//  0, 0, 0, 0, 0,
//  kKeyPadDecimal, 0, kKeyPadMultiply, 0, kKeyPadAdd,
//  0, kKeyPadClear, 0, 0, 0,
//  kKeyPadDivide, 0, kKeyPadEnter, kKeyPadSubtract, 0,
//  0, 0, kKeyPad0, kKeyPad1, kKeyPad2,
//  kKeyPad3, kKeyPad4, kKeyPad5, kKeyPad6, kKeyPad7,
//  kKeyPad8, kKeyPad9, 0, 0, 0,
//  0, kKeyF5, kKeyF6, kKeyF7, kKeyF3,
//  kKeyF8, kKeyF9, 0, kKeyF11, 0,                    // 100
//  0, 0, 0, 0, kKeyF10,
//  0, kKeyF12, 0, 0, kKeyHelp,
//  kKeyHome, kKeyPageUp, kKeyDelete, kKeyF4, kKeyEnd,
//  kKeyF2, kKeyPageDown, kKeyF1, kKeyLeft, kKeyRight,
//  kKeyDown, kKeyUp,
//  -1, -1, -1, -1, -1};
//
//
//// BUG(jwills): With multiple windows open, you can get a gl context to draw onto the header bar of
//// a window by moving it around a bunch.  Why on earth does this happen?
//
//struct OsWindowData {
//  OsWindowData() :
//      window(NULL),
//      agl_context(NULL) {
//  }
//  ~OsWindowData() {
//    printf("destroyed\n");
//    if (window != NULL) {
//      printf("thundered\n");
//      DisposeWindow(window);
//    }
//    if (agl_context != NULL) {
//      aglDestroyContext(agl_context);
//    }
//  }
//  WindowRef window;
//  AGLContext agl_context;
//  Rect bounds;
//  Rect full_screen_dimensions;
//  string title;
//  bool full_screen;
//};
//
//string GetExecutablePath() {
//  char path[1024];
//  uint32_t path_size = 1024;
//  _NSGetExecutablePath(path, &path_size);
//  return string(path);
//}
//
//void GlopToggleFullScreen();
//
// HACK!
static bool ok_to_exit;
// END HACK!
//
static double last_time;
static bool lost_focus;
static bool caps_lock;
static bool num_lock;
//
//struct GlopOSXEvent {
//  double timestamp;
//  Os::KeyEvent event;
//  GlopOSXEvent() : event(0, 0, 0, false, false) {}
//  GlopOSXEvent(double _timestamp, int dx, int dy, int FIXME) :
//      timestamp(_timestamp),
//      event(dx, dy, (timestamp - 600000) * 1000, mouse_location.x, mouse_location.y, num_lock, caps_lock) {}
////  GlopOSXEvent(double _timestamp, GlopKey key, bool is_pressed) :
////      timestamp(_timestamp),
////      event(key, is_pressed, timestamp * 1000, mouse_location.x, mouse_location.y, false, false) {}
//  GlopOSXEvent(double _timestamp, GlopKey key, float pressed_amount) :
//      timestamp(_timestamp),
//      event(key, pressed_amount, (timestamp - 600000) * 1000, mouse_location.x, mouse_location.y, num_lock, caps_lock) {}
//  GlopOSXEvent(double _timestamp, Os::KeyEvent _event)
//      : timestamp((_timestamp - 600000) * 1000),
//        event(_event) {}
//};
//bool operator < (const GlopOSXEvent& a, const GlopOSXEvent& b) {
//  if (a.timestamp != b.timestamp)
//    return a.timestamp < b.timestamp;
//  if (a.event.key != b.event.key)
//    return a.event.key < b.event.key;
//  return 0;
//}
//
//
//vector<GlopOSXEvent> raw_events;
//

OSStatus glopEventHandler(EventHandlerCallRef next_handler, EventRef the_event, void* user_data) {
  OSStatus result = eventNotHandledErr;
  int event_class = GetEventClass(the_event);
  int event_kind = GetEventKind(the_event);
printf("--> %d %d\n", event_class, event_kind);
  if (event_class == kEventClassGlop && event_kind == kEventGlopBreak) {
    ok_to_exit = false;
    QuitApplicationEventLoop();
    last_time = GetEventTime(the_event);
    result = noErr;
  }
//  if (event_class == kEventClassApplication && event_kind == kEventAppQuit) {
//    if (ok_to_exit) {
//      exit(0);
//      result = noErr;
//    }
//  }
//  if (event_class == kEventClassWindow) {
//    if (event_kind == kEventWindowFocusRelinquish) {
//      lost_focus = true;
//      result = noErr;
//    }
//  }
//  if (event_class == kEventClassCommand) {
//    if (event_kind == kEventProcessCommand) {
//      HICommand command;
//    	GetEventParameter(
//    	    the_event,
//    	    kEventParamDirectObject,
//    	    kEventParamHICommand,
//    	    NULL,
//    	    sizeof(command),
//    	    NULL,
//          &command);
//      if (command.commandID == kEventGlopToggleFullScreen) {
//        printf("Toggle\n");
//        GlopToggleFullScreen();
//        result = noErr;
//      }
//    }
//  }
//  if (event_class == kEventClassKeyboard || event_class == kEventClassMouse) {
//    UInt32 modifier_keys;
//    GetEventParameter(
//        the_event,
//        kEventParamKeyModifiers,
//        typeUInt32,
//        NULL,
//        sizeof(modifier_keys),
//        NULL,
//        &modifier_keys);
//    caps_lock = (modifier_keys & 0x400);
//    num_lock = (modifier_keys & 0x10000);
//  }
//  if (event_class == kEventClassKeyboard) {
//    if (event_kind == kEventRawKeyDown) {
//      UInt32 key;
//    	GetEventParameter(
//    	    the_event,
//    	    kEventParamKeyCode,
//    	    typeUInt32,
//    	    NULL,
//    	    sizeof(key),
//    	    NULL,
//          &key);
//      printf("KeyDown: %d %x\n", key, key);
//      printf("Array: %d\n",key_map[key].index);
////      printf("%s\n", key_map[key].GetName().c_str());
//      if (key_map[key].index != 0)
//      raw_events.push_back(
//          GlopOSXEvent(GetEventTime(the_event), key_map[key], true));
//    }
//    if (event_kind == kEventRawKeyUp) {
//      UInt32 key;
//    	GetEventParameter(
//    	    the_event,
//    	    kEventParamKeyCode,
//    	    typeUInt32,
//    	    NULL,
//    	    sizeof(key),
//    	    NULL,
//          &key);
//      printf("KeyUp  : %d %x\n", key, key);
//      if (key_map[key].index != 0)
//      raw_events.push_back(
//          GlopOSXEvent(GetEventTime(the_event), key_map[key], false));
//    }
//    if (event_kind == kEventRawKeyModifiersChanged) {
//      UInt32 modifiers;
//    	GetEventParameter(
//    	    the_event,
//    	    kEventParamKeyModifiers,
//    	    typeUInt32,
//    	    NULL,
//    	    sizeof(modifiers),
//    	    NULL,
//          &modifiers);
//      printf("Modifiers: %d\t%x\n", modifiers, modifiers);
//      printf("GetCurren: %d\t%x\n", GetCurrentEventKeyModifiers(), GetCurrentEventKeyModifiers());
//      printf("KeyModifs: %d\t%x\n", GetCurrentKeyModifiers(), GetCurrentKeyModifiers());
//    }
//  }
//  if (event_class == kEventClassMouse) {
//    if (event_kind == kEventMouseWheelMoved) {
//      int axis;
//      GetEventParameter(
//        the_event,
//        kEventParamMouseWheelAxis,
//        typeMouseWheelAxis,
//        NULL,
//        sizeof(int),
//        NULL,
//        &axis);
//      SInt32 delta;
//      GetEventParameter(
//        the_event,
//        kEventParamMouseWheelDelta,
//        typeSInt32,
//        NULL,
//        sizeof(delta),
//        NULL,
//        &delta);
//      if (axis == kEventMouseWheelAxisY) {
//        if (delta < 0) {
//          raw_events.push_back(
//              GlopOSXEvent(GetEventTime(the_event), kMouseWheelDown, -delta / 1.0));
//          raw_events.push_back(
//              GlopOSXEvent(GetEventTime(the_event), kMouseWheelDown, 0));
//        } else {
//          raw_events.push_back(
//              GlopOSXEvent(GetEventTime(the_event), kMouseWheelUp, delta / 1.0));
//          raw_events.push_back(
//              GlopOSXEvent(GetEventTime(the_event), kMouseWheelUp, 0));
//        }
//      }
//    }
//    if (event_kind == kEventMouseMoved || event_kind == kEventMouseDragged) {
//      GetEventParameter(
//        the_event,
//        kEventParamMouseLocation,
//        typeHIPoint,
//        NULL,
//        sizeof(mouse_location),
//        NULL,
//        &mouse_location);
//      HIPoint current_delta;
//      GetEventParameter(
//        the_event,
//        kEventParamMouseDelta,
//        typeHIPoint,
//        NULL,
//        sizeof(current_delta),
//        NULL,
//        &current_delta);
//      mouse_delta.x += current_delta.x;
//      mouse_delta.y += current_delta.y;
//      printf("Mouse Moved: %f %f\n", mouse_delta.x, mouse_delta.y);
//      raw_events.push_back(
//          GlopOSXEvent(GetEventTime(the_event), current_delta.x, current_delta.y, 0));
//    }
//    if (event_kind == kEventMouseDown || event_kind == kEventMouseUp) {
//      HIPoint point;
//      EventMouseButton button;
//      GetEventParameter(
//          the_event,
//          kEventParamMouseLocation,
//          typeHIPoint,
//          NULL,
//          sizeof(point),
//          NULL,
//          &point);
//      GetEventParameter(
//          the_event,
//          kEventParamMouseButton,
//          typeMouseButton,
//          NULL,
//          sizeof(button),
//          NULL,
//          &button);
//      GlopKey glop_mouse_button(0);
//      if (button == kEventMouseButtonPrimary) {
//        glop_mouse_button = kMouseLButton;
//      } else 
//      if (button == kEventMouseButtonSecondary) {
//        glop_mouse_button = kMouseRButton;
//      } else
//      if (button == kEventMouseButtonTertiary) {
//        glop_mouse_button = kMouseMButton;
//      }
//      if (glop_mouse_button != GlopKey(0)) {
//        raw_events.push_back(
//            GlopOSXEvent(
//                GetEventTime(the_event),
//                glop_mouse_button,
//                event_kind == kEventMouseDown));
//      }
//    }
//  }
  return result;
}

OSStatus glopWindowHandler(EventHandlerCallRef next_handler, EventRef the_event, void* user_data) {
  OSStatus result = eventNotHandledErr;
  return result;
}
//  int event_class = GetEventClass(the_event);
//  int event_kind = GetEventKind(the_event);
//  OsWindowData* data = (OsWindowData*)user_data;
//  if (event_class == kEventClassWindow) {
//    if (event_kind == kEventWindowResizeCompleted || event_kind == kEventWindowBoundsChanged) {
//    	GetEventParameter(
//    	    the_event,
//    	    kEventParamCurrentBounds,
//    	    typeQDRectangle,
//    	    NULL,
//    	    sizeof(data->bounds),
//    	    NULL,
//          &data->bounds);
//      Os::SetCurrentContext(data);
//      result = noErr;
////      glViewport(0,50,data->bounds.right - data->bounds.left,
////      data->bounds.bottom - data->bounds.top+50);
//    }
//    if (event_kind == kEventWindowClosed) {
//      printf("Destroying window %s\n", data->title.c_str());
//      if (!data->full_screen) {
//// TODO(jwills): Look here, fixy fixy.
////        Os::DestroyWindow(data);
//      }
//      result = noErr;
//    }
//  }
//  return result;
//}
//
//struct GlopHIDEvent {
//  double timestamp;
//  UInt32 page;
//  UInt32 usage;
//  int value;
//  int queue;
//  GlopHIDEvent() {}
//  GlopHIDEvent(IOHIDValueRef value_ref, int _queue) :
//      page(IOHIDElementGetUsagePage(IOHIDValueGetElement(value_ref))),
//      usage(IOHIDElementGetUsage(IOHIDValueGetElement(value_ref))),
//      value(IOHIDValueGetIntegerValue(value_ref)),
//      queue(_queue) {
//    uint64_t t = IOHIDValueGetTimeStamp(value_ref);
//    Nanoseconds nano = AbsoluteToNanoseconds(*(AbsoluteTime*)&t);
//    long long time_in_nanoseconds = *(long long*)&nano;
//    timestamp = (double)time_in_nanoseconds / 1e9;
//  }
//};
//
//bool operator < (const GlopHIDEvent& a, const GlopHIDEvent& b) {
//  if (a.timestamp != b.timestamp)
//    return a.timestamp < b.timestamp;
//  if (a.page != b.page)
//    return a.page < b.page;
//  if (a.usage != b.usage)
//    return a.usage < b.usage;
//  return a.value < b.value;
//}
//
//static vector<IOHIDQueueRef> modifier_queues;
//static set<GlopHIDEvent> modifier_events;
//
//static vector<IOHIDQueueRef> joystick_queues;
//static set<GlopHIDEvent> joystick_events;
//
//static vector<IOHIDDeviceRef> joystick_devices;
//// This is a parallel vector to joystick_queues, this is so when we update the joysticks that we
//// don't move any around that were present from the last time we updated joysticks
//
//static void ExtractEvents(vector<IOHIDQueueRef>* queues, set<GlopHIDEvent>* events) {
//  for (int i = 0; i < queues->size(); i++) {
//    if ((*queues)[i] == NULL) continue;
//    IOHIDValueRef value = IOHIDQueueCopyNextValueWithTimeout((*queues)[i], 0);
//    while (value) {
//      events->insert(GlopHIDEvent(value, i));
//      value = IOHIDQueueCopyNextValueWithTimeout((*queues)[i], 0);
//    }
//  }
//}
//
//static void GlopProcessModifierHIDs(const void* value, void* context) {
//	IOHIDDeviceRef device_ref = (IOHIDDeviceRef)value;
//  // We match all devices and in here we check all elements.  This might be slow,
//  // but this also shouldn't happen very often, typically just once at startup.
//	if (device_ref) {
//    bool queue_created = false;
//    CFArrayRef elements = IOHIDDeviceCopyMatchingElements(device_ref, NULL, 0);
//    for (int i = 0; i < CFArrayGetCount(elements); i++) {
//      IOHIDElementRef element_ref = (IOHIDElementRef)CFArrayGetValueAtIndex(elements, i);
//      int page = IOHIDElementGetUsagePage(element_ref);
//      int usage = IOHIDElementGetUsage(element_ref);
//      if (page == 0x07 && ((usage >= 0xE0 && usage <= 0xE7) || usage == 0x39)) {
//        if (!queue_created) {
//          modifier_queues.push_back(IOHIDQueueCreate(NULL, device_ref, 100, kIOHIDOptionsTypeNone));
//          queue_created = true;
//        }
//        IOHIDQueueAddElement(modifier_queues.back(), element_ref);
////        printf("Added key: %x\n", usage);
//      }
//    }
//	}
//}
//
//// TODO(jwills): This function should probably undergo some serious testing
//static void GlopFindRemainingJoysticks(const void* value, void* context) {
//  IOHIDDeviceRef device_ref = (IOHIDDeviceRef)value;
//  if (device_ref) {
//    CFArrayRef elements = IOHIDDeviceCopyMatchingElements(device_ref, NULL, 0);
//    for (int i = 0; i < CFArrayGetCount(elements); i++) {
//      int index = -1;
//      for (int i = 0; i < CFArrayGetCount(elements); i++) {
//        IOHIDElementRef element_ref = (IOHIDElementRef)CFArrayGetValueAtIndex(elements, i);
//        int page = IOHIDElementGetUsagePage(element_ref);
//        int usage = IOHIDElementGetUsage(element_ref);
//        if (page == 0x01 && (usage == 0x04 || usage == 0x05)) {
//          index++;
//          while (index < joystick_devices.size() && joystick_devices[index] != device_ref) {
//            index++;
//          }
//          if (index < joystick_devices.size()) {
//            joystick_queues[index] = IOHIDQueueCreate(NULL, device_ref, 100, kIOHIDOptionsTypeNone);
//          }
//        }
//      }
//    }
//  }
//}
//
//static void GlopProcessJoysticks(const void* value, void* context) {
//	IOHIDDeviceRef device_ref = (IOHIDDeviceRef)value;
//  // We match all devices and in here we check all elements.  This might be slow,
//  // but this also shouldn't happen very often, typically just once at startup.
////  printf("new device:\n");
//	if (device_ref) {
//    CFArrayRef elements = IOHIDDeviceCopyMatchingElements(device_ref, NULL, 0);
//    int index = -1;
//    for (int i = 0; i < CFArrayGetCount(elements); i++) {
//      IOHIDElementRef element_ref = (IOHIDElementRef)CFArrayGetValueAtIndex(elements, i);
//      int page = IOHIDElementGetUsagePage(element_ref);
//      int usage = IOHIDElementGetUsage(element_ref);
////      printf("GlopProcessJoysticks(): %d %d\n", page, usage);
//      // TODO(jwills): Figure out if joysticks always indicate what they are with usage page 0x01
//      if (page == 0x01 && (usage == 0x04 || usage == 0x05)) {
//        index = index + 1;
//        while (index < joystick_devices.size() && joystick_devices[index] != device_ref) {
//          index++;
//        }
//        printf("1index: %d\n", index);
//        if (index == joystick_devices.size()) {
//          for (int j = 0; j < joystick_devices.size(); j++) {
//            if (joystick_devices[j] == NULL) {
//              index = j;
//              joystick_devices[j] = device_ref;
//              assert(joystick_queues[j] == NULL);
//              joystick_queues[j] = IOHIDQueueCreate(NULL, device_ref, 100, kIOHIDOptionsTypeNone);
//              break;
//            }
//          }
//          printf("2index: %d\n", index);
//          if (index == joystick_devices.size()) {
//            index = joystick_queues.size();
//            joystick_queues.push_back(
//                IOHIDQueueCreate(NULL, device_ref, 100, kIOHIDOptionsTypeNone));
//            joystick_devices.push_back(device_ref);
//          }
//          printf("3index: %d\n", index);
//        }
//      }
//      if (index != -1 &&
//          ((page == 0x01 && (usage >= 0x30 && usage <= 0x39)) ||
//           (page == 0x09 && (usage > 0)))) {
//        IOHIDQueueAddElement(joystick_queues[index], element_ref);
////        printf("Added joystick button: %x\n", usage);
//      }
//    }
//	}
//}
//
//static IOHIDManagerRef glop_manager;
//static void GlopHandleNewModifierHIDs() {
//	CFSetRef set_ref = IOHIDManagerCopyDevices(glop_manager);
//	if (set_ref) {
//    vector<IOHIDQueueRef> old_modifier_queues = modifier_queues;
//    modifier_queues.resize(0);
//		CFSetApplyFunction(set_ref, GlopProcessModifierHIDs, NULL);
//    ExtractEvents(&old_modifier_queues, &modifier_events);
//    for (int i = 0; i < old_modifier_queues.size(); i++) {
//      CFRelease(old_modifier_queues[i]);
//    }
//	} else {
//    printf("Error: Tried to handle new HIDs, but there are no HIDs.\n");
//	}
//  for (int i = 0; i < modifier_queues.size(); i++) {
//    IOHIDQueueStart(modifier_queues[i]);
//  }
//}
//static void GlopHandleNewJoysticks() {
//	CFSetRef set_ref = IOHIDManagerCopyDevices(glop_manager);
//	if (set_ref) {
//    vector<IOHIDQueueRef> old_joystick_queues = joystick_queues;
//    for (int i = 0; i < joystick_queues.size(); i++) {
//      joystick_queues[i] = NULL;
//    }
//		CFSetApplyFunction(set_ref, GlopFindRemainingJoysticks, NULL);
//    for (int i = 0; i < joystick_devices.size(); i++) {
//      if (joystick_queues[i] == NULL) {
//        if (joystick_devices[i] != NULL) {
//          printf("Removed joystick %d\n", i);
//        }
//        joystick_devices[i] = NULL;
//      }
//    }
//		CFSetApplyFunction(set_ref, GlopProcessJoysticks, NULL);
//    ExtractEvents(&old_joystick_queues, &joystick_events);
//	} else {
//    printf("Error: Tried to handle new HIDs, but there are no HIDs.\n");
//	}
//  for (int i = 0; i < joystick_queues.size(); i++) {
//    if (joystick_queues[i] != NULL) {
//      IOHIDQueueStart(joystick_queues[i]);
//    }
//  }
//  for (int i = 0; i < joystick_devices.size(); i++) {
////    printf("Device ID: %d\n", IOHIDDevice_GetProductID(joystick_devices[i]));
//  }
//}
//
//// This function is way less efficient than it could be, but it doesn't seem like there is really
//// any point.  This function should be called once every time a USB device is added or removed, so
//// if it's slow, it shouldn't matter.
//static void GlopUpdateDevices(void* context, IOReturn result, void* sender, IOHIDDeviceRef device) {
//  GlopHandleNewModifierHIDs();
//  GlopHandleNewJoysticks();
//}
//
static UnsignedWide glop_start_time;
void Init() {
  Microseconds(&glop_start_time);

  EventHandlerUPP handler_upp = NewEventHandlerUPP(glopEventHandler);
  EventTypeSpec event_types[11];
  event_types[0].eventClass = kEventClassGlop;
  event_types[0].eventKind  = kEventGlopBreak;
  event_types[1].eventClass = kEventClassApplication;
  event_types[1].eventKind  = kEventAppQuit;
  event_types[2].eventClass = kEventClassCommand;
  event_types[2].eventKind  = kEventProcessCommand;
  event_types[3].eventClass = kEventClassKeyboard;
  event_types[3].eventKind  = kEventRawKeyDown;
  event_types[4].eventClass = kEventClassKeyboard;
  event_types[4].eventKind  = kEventRawKeyUp;
  event_types[5].eventClass = kEventClassMouse;
  event_types[5].eventKind  = kEventMouseMoved;
  event_types[6].eventClass = kEventClassMouse;
  event_types[6].eventKind  = kEventMouseDragged;
  event_types[7].eventClass = kEventClassMouse;
  event_types[7].eventKind  = kEventMouseDown;
  event_types[8].eventClass = kEventClassMouse;
  event_types[8].eventKind  = kEventMouseUp;
  event_types[9].eventClass = kEventClassMouse;
  event_types[9].eventKind  = kEventMouseWheelMoved;
  event_types[10].eventClass = kEventClassWindow;
  event_types[10].eventKind  = kEventWindowFocusRelinquish;
  InstallApplicationEventHandler(handler_upp, 11, event_types, NULL, NULL);

//  // Handle HIDs
//  glop_manager = IOHIDManagerCreate(kCFAllocatorDefault, kIOHIDOptionsTypeNone);
//  if (glop_manager == NULL) {
//    printf("couldn't make a manager\n");
//  }
//  IOHIDManagerSetDeviceMatching(glop_manager, NULL);  // This will match all HIDs, but that's fine
//  IOReturn result = IOHIDManagerOpen(glop_manager, kIOHIDOptionsTypeNone);
//  if (kIOReturnSuccess != result) {
//    printf("Couldn't open IOHIDManager: %x\n", result);
//  }
//  IOHIDManagerRegisterDeviceMatchingCallback(glop_manager, GlopUpdateDevices, (void*)1);
//  IOHIDManagerRegisterDeviceRemovalCallback(glop_manager, GlopUpdateDevices, (void*)0);
//  CFRunLoopRef run_loop_ref = (CFRunLoopRef)GetCFRunLoopFromEventLoop(GetMainEventLoop());
//  IOHIDManagerScheduleWithRunLoop(glop_manager, run_loop_ref, kCFRunLoopDefaultMode);
//  GlopUpdateDevices(NULL, NULL, NULL, NULL);
//
//  OSStatus err;
//  IBNibRef nib_ref = NULL;
//  CFBundleRef bundle;
//  bundle = CFBundleGetBundleWithIdentifier(CFSTR("com.thunderproductions.glopframework"));
//  if (bundle == NULL) {
////    LOGF("Failed to create bundle reference.");
//  } else {
//    err = CreateNibReferenceWithCFBundle(bundle, CFSTR("main"), &nib_ref);
//    err = SetMenuBarFromNib(nib_ref, CFSTR("MainMenu"));
//    DisposeNibReference(nib_ref);
//  }
}

//void Os::ShutDown() {
//  // Booya! :-)
//}

//void Os::Think() {
//  // Static since we don't want to create and free this event every time we think
//  EventRef terminator;
//  CreateEvent(NULL, kEventClassGlop, kEventGlopBreak, 0, kEventAttributeNone, &terminator);
//  PostEventToQueue(GetMainEventQueue(), terminator, kEventPriorityLow);
//  ok_to_exit = true;
//  RunApplicationEventLoop();
//  ReleaseEvent(terminator);
//  // The application event loop does not get modifier keys for the keyboard, or joysticks, both of
//  // those have to be done separately.
//
//  ExtractEvents(&modifier_queues, &modifier_events);
//  set<GlopHIDEvent>::iterator event;
//  for (event = modifier_events.begin(); event != modifier_events.end(); event++) {
//    static map<UInt32, GlopKey> modifier_map;
//    if (modifier_map.size() == 0) {
//      modifier_map[0x39] = kKeyCapsLock;
//      modifier_map[0xE0] = kKeyLeftControl;
//      modifier_map[0xE1] = kKeyLeftShift;
//      modifier_map[0xE2] = kKeyLeftAlt;
//      modifier_map[0xE3] = kKeyLeftGui;
//      modifier_map[0xE4] = kKeyRightControl;
//      modifier_map[0xE5] = kKeyRightShift;
//      modifier_map[0xE6] = kKeyRightAlt;
//      modifier_map[0xE7] = kKeyRightGui;
//    }
//    printf("Modifier: %f %d %d %d\n", event->timestamp, event->page, event->usage, event->value);
//    raw_events.push_back(
//        GlopOSXEvent(event->timestamp,modifier_map[event->usage], (bool)event->value));
//  }
//  modifier_events.clear();
//  // modifier_events only gets new events inside RunApplicationEventLoop(), so we won't miss any
//  // events by clearing the whole thing here.
//
//  ExtractEvents(&joystick_queues, &joystick_events);
//  //  printf("Num Joystick Queues: %d\n", joystick_queues.size());
//
//  for (event = joystick_events.begin(); event != joystick_events.end(); event++) {
//    printf("Event: %f %d %d %d\n", event->timestamp, event->page, event->usage, event->value);
//    static map<pair<UInt32, UInt32>, int> joystick_map;
//    if (joystick_map.size() == 0) {
////      joystick_map[pair<UInt32, UInt32>(0x09,)] = 0;
//    }
//    printf("page: 0x%x\tusage: 0x%x\n", event->page, event->usage);
//    if (event->page == 0x09) {
//      raw_events.push_back(
//          GlopOSXEvent(
//              event->timestamp,
//              GetJoystickButton(event->usage - 1, event->queue),
//              (bool)event->value));
//    } else if (event->page == 0x01 && event->usage >= 0x30 && event->usage <= 0x35) {
//      printf("value: %d\n", event->value);
//      printf("queue: %d\n", event->queue);
//      if (event->value < 127) {
//        LOGF("Negative axis event %d", event->usage);
//        raw_events.push_back(
//            GlopOSXEvent(
//                event->timestamp,
//                GetJoystickAxisNeg(event->usage - 0x30, event->queue),
//                1.f - (event->value/128.f)));
//      } else
//      if (event->value > 128) {
//        LOGF("Positive axis event %d", event->usage);
//        raw_events.push_back(
//            GlopOSXEvent(
//                event->timestamp,
//                GetJoystickAxisPos(event->usage - 0x30, event->queue),
//                (event->value - 127) / 128.f));
//      } else {
//        LOGF("Zero axis event %d", event->usage);
//        raw_events.push_back(
//            GlopOSXEvent(
//                event->timestamp,
//                GetJoystickAxisNeg(event->usage - 0x30, event->queue),
//                0.f));
//        raw_events.push_back(
//            GlopOSXEvent(
//                event->timestamp,
//                GetJoystickAxisPos(event->usage - 0x30, event->queue),
//                0.f));
//      }
//    } else if (event->page == 0x01 && event->usage >= 0x39) {
//      printf("rawr!\n");
//      raw_events.push_back(
//          GlopOSXEvent(
//              event->timestamp,
//              GetJoystickHatUp(0, event->queue),
//              event->value == 0x07 || event->value <= 0x01));
//      raw_events.push_back(
//          GlopOSXEvent(
//              event->timestamp,
//              GetJoystickHatRight(0, event->queue),
//              event->value >= 0x01 && event->value <= 0x03));
//      raw_events.push_back(
//          GlopOSXEvent(
//              event->timestamp,
//              GetJoystickHatDown(0, event->queue),
//              event->value >= 0x03 && event->value <= 0x05));
//      raw_events.push_back(
//          GlopOSXEvent(
//              event->timestamp,
//              GetJoystickHatLeft(0, event->queue),
//              event->value >= 0x05 && event->value <= 0x07));
//    }
//  }
//  joystick_events.clear();
//
//  set<OsWindowData*>::iterator window;
//  for (window = all_windows.begin(); window != all_windows.end(); window++) {
//    WindowThink(*window);
//  }
//}
//
//
//OSStatus aglReportError(void) {
//	GLenum err = aglGetError();
//	if (AGL_NO_ERROR != err) {
//		char errStr[256];
//		printf(errStr, "AGL: %s",(char *) aglErrorString(err));
//	}
//	// ensure we are returning an OSStatus noErr if no error condition
//	if (err == AGL_NO_ERROR)
//		return noErr;
//	else
//		return (OSStatus) err;
//}

typedef struct {
  WindowRef window;
  AGLContext agl;
  Rect bounds;
  Rect full_bounds;
  //char* window_title;
  int full_screen;
} windowData;

void setCurrentContext(windowData* data) {
  if (data->agl == NULL) {
    //printf("No agl context, can't set context.\n");
    return;
  }
  if (!aglSetCurrentContext(data->agl)) {
    //aglReportError();
    //printf("Error!\n");
  }
  if (!aglUpdateContext(data->agl)) {
    //aglReportError();
    printf("Error!\n");
  }
}

void windowThink(windowData* data) {
  setCurrentContext(data);
}

void Think(windowData* window) {
  EventRef terminator;
  CreateEvent(NULL, kEventClassGlop, kEventGlopBreak, 0, kEventAttributeNone, &terminator);
  PostEventToQueue(GetMainEventQueue(), terminator, kEventPriorityLow);
  ok_to_exit = true;
  RunApplicationEventLoop();
  ReleaseEvent(terminator);
  // The application event loop does not get modifier keys for the keyboard, or joysticks, both of
  // those have to be done separately.

  windowThink(window);
}

bool createAGLContext(windowData* data) {
  // OSStatus err = noErr;
  // Use err = aglReportError(); to find out any problems with agl calls
  GLint attributes[] = {
    AGL_RGBA,
    AGL_DOUBLEBUFFER,
    AGL_DEPTH_SIZE, 32,
    (data->full_screen ? AGL_FULLSCREEN : AGL_NONE),
    AGL_NONE
  };
  AGLPixelFormat pixel_format;
  GDHandle gdhDisplay;
  DMGetGDeviceByDisplayID(CGMainDisplayID(), &gdhDisplay, false);
  if (data->full_screen) {
    CFDictionaryRef refDisplayMode = 0;
    boolean_t exact_match;
    refDisplayMode = CGDisplayBestModeForParametersAndRefreshRate(
        CGMainDisplayID(),
        32,
        data->full_bounds.right - data->full_bounds.left,
        data->full_bounds.bottom - data->full_bounds.top,
        65,
        &exact_match);
    if (!exact_match) {
      return false;
    }
    CGDisplaySwitchToMode(CGMainDisplayID(), refDisplayMode);
  }

  pixel_format = aglChoosePixelFormat(&gdhDisplay, 1, attributes);
  if (pixel_format) {
    data->agl = aglCreateContext(pixel_format, NULL);
    aglDestroyPixelFormat(pixel_format);
//    SetPort(GetWindowPort(data->window));
  } else {
    // No valid pixel format found
    return false;
  }
////	refDisplayMode = CGDisplayBestModeForParametersAndRefreshRate (pContextInfo->display, depth, width, height, refresh, NULL);

//  aglSetHIViewRef(data->agl_context, HIViewGetRoot(data->window));
  if (data->full_screen) {
    if (!aglSetCurrentContext(data->agl)) {
      return false;
    }
    aglSetFullScreen(data->agl, 0, 0, 0, 0);
//    aglSetDrawable(data->agl_context, GetWindowPort(data->window));
  } else {
    aglSetDrawable(data->agl, GetWindowPort(data->window));
    GLenum err = aglGetError();
    printf("aglerr: %s\n", aglErrorString(err));
  }
  return true;
}

void openWindow(windowData* data) {
  CreateNewWindow(
      kDocumentWindowClass,
          kWindowCollapseBoxAttribute |
          kWindowResizableAttribute |
          kWindowStandardHandlerAttribute |
          kWindowAsyncDragAttribute |
          kWindowLiveResizeAttribute,
  &(data->bounds),
  &(data->window));

  EventTypeSpec event_types[3];
  event_types[0].eventClass = kEventClassWindow;
  event_types[0].eventKind  = kEventWindowResizeCompleted;
  event_types[1].eventClass = kEventClassWindow;
  event_types[1].eventKind  = kEventWindowClosed;
  event_types[2].eventClass = kEventClassWindow;
  event_types[2].eventKind  = kEventWindowBoundsChanged;
  EventHandlerUPP handler_upp = NewEventHandlerUPP(glopWindowHandler);
  InstallWindowEventHandler(data->window, handler_upp, 3, event_types, data, NULL);
  createAGLContext(data);

  //Os::SetTitle(data, data->title);
  setCurrentContext(data);
  SelectWindow(data->window);
  ShowWindow(data->window);
}

void CreateWindow(windowData** data, int full_screen, int x, int y, int dx, int dy) {
  *data= (windowData*)malloc(sizeof(windowData));
  (*data)->full_screen = full_screen;
  if (full_screen) {
    // TODO: Do full screen garbage here
  } else {
    SetRect(&((*data)->bounds), x, y, x + dx, y + dy);
//    SetRect(&data->full_screen_dimensions, 0, 0, 1600, 1050);
    openWindow(*data);
  }
  // TODO: If we allow multiple windows then we need to add this window to a global list of windows
}


//
//bool GlopEnterFullScreen(OsWindowData* data) {
//  if (!GlopCreateAGLContext(data)) {
//    return false;
//  }
//  printf(
//      "Entering fullscreen with dimensions %d,%d\n",
//      data->full_screen_dimensions.right,
//      data->full_screen_dimensions.bottom);
//  Os::SetCurrentContext(data);
//  aglSetFullScreen(data->agl_context, 0, 0, 0, 0);
//  return true;
//}
//


//void GlopToggleFullScreen() {
//  return;
//  // First find the window with the focus
//  printf("GlopToggleFullScreen()\n");
//  OsWindowData* data = NULL;
//  OsWindowData* full_screen_data = NULL;
//  set<OsWindowData*>::iterator it;
//  for (it = all_windows.begin(); it != all_windows.end(); it++) {
//    if (IsWindowActive((*it)->window) || (*it)->full_screen) {
//      data = *it;
//      if (data->full_screen) {
//        full_screen_data = *it;
//      }
//    }
//  }
//  if (data == NULL) {
//    // We can't go into fullscreen if we don't know what window to apply this to
//    return;
//  }
//  if (full_screen_data != NULL) {
//    data = full_screen_data;
//  }
//  printf("GlopToggleFullScreen(%s)\n", data->title.c_str());
////  aglSetDrawable(NULL, NULL);
//  data->full_screen = !data->full_screen;
//  if (!data->full_screen) {
//    printf("Currently fullscreen, making windowed\n");
//    aglDestroyContext(data->agl_context);
//    data->agl_context = NULL;
//    GlopOpenWindow(data);
//  } else {
//    printf("Currently windowed, making fullscreen\n");
//    DisposeWindow(data->window);
//    aglDestroyContext(data->agl_context);
//    GlopEnterFullScreen(data);
//  }
//  if (data->full_screen)
//    printf("Now fullscreened\n");
//  else
//    printf("Now windowed\n");
//}
//  WindowRef window;
//  AGLContext agl_context;
//  Rect bounds;
//  Rect full_screen_dimensions;
//  string title;
//  bool full_screen;
//

//
//// Destroys a window that is completely or partially created.
//void Os::DestroyWindow(OsWindowData* data) {
//  all_windows.erase(data);
//  delete data;
//}
//
//bool Os::IsWindowMinimized(const OsWindowData* data) {
//  return IsWindowCollapsed(data->window);
//}
//
//void Os::GetWindowFocusState(OsWindowData* data, bool* is_in_focus, bool* focus_changed) {
//  if (data->full_screen) {
//    *is_in_focus = true;
//  } else {
//    *is_in_focus = IsWindowActive(data->window);
//  }
//  *focus_changed = lost_focus;
//  if (lost_focus)
//    LOGF("Lost focus");
//  lost_focus = false;
//}
//
//void Os::GetWindowPosition(const OsWindowData* data, int* x, int* y) {
//  *x = data->bounds.left;
//  *y = data->bounds.top;
//}
//
//void Os::GetWindowSize(const OsWindowData* data, int* width, int* height) {
//  *width  = data->bounds.right - data->bounds.left;
//  *height = data->bounds.bottom - data->bounds.top;
//}
//
//void Os::SetTitle(OsWindowData* data, const string& title) {
//  CFStringRef cf_title;
//  data->title = title;
//  cf_title = CFStringCreateWithCString(NULL, data->title.c_str(), kCFStringEncodingASCII);  
//  SetWindowTitleWithCFString(data->window, cf_title);
//  CFRelease(cf_title);
//}
//
//void Os::SetIcon(OsWindowData *window, const Image *icon) {
//  // TODO(jwills): Figure out exactly how this should work considering that it needs to run on PCs
//  // and on macs
//}
//
//void Os::SetWindowSize(OsWindowData *window, int width, int height) {
//  // TODO(jwills): Rawr!!!
//}
//
//// Input functions
//// ===============
//
//// See Os.h
//
//vector<Os::KeyEvent> Os::GetInputEvents(OsWindowData *window) {
//  vector<Os::KeyEvent> ret;
//  printf("RawEvents: %d\n", raw_events.size());
//  for (int i = 0; i < raw_events.size(); i++) {
//    LOGF("Event: %s : %f (%d)\n", raw_events[i].event.key.GetName().c_str(), raw_events[i].event.press_amount, raw_events[i].event.timestamp);
//  }
//  stable_sort(raw_events.begin(), raw_events.end());
//  ret.reserve(raw_events.size());
//  for (int i = 0; i < raw_events.size(); i++) {
//    ret.push_back(raw_events[i].event);
//  }
//  if (raw_events.size())
//  printf("------\n");
//  raw_events.resize(0);
//  mouse_delta.x = 0;
//  mouse_delta.y = 0;
////  if (ret.size() == 0)
//  ret.push_back(
//      Os::KeyEvent(
//          (int)((last_time - 600000) * 1000),
//          mouse_location.x,
//          mouse_location.y,
//          false,
//          false));
////  printf("%f\n", last_time);
//  for (int i = 0; i  < ret.size() - 1; i++) {
//    printf("Event: %s : %f\n", ret[i].key.GetName().c_str(), ret[i].press_amount);
//  }
//  return ret;
//}
//
//void Os::SetMousePosition(int x, int y) {
//  CGPoint point;
//  point.x = x;
//  point.y = y;
//  CGWarpMouseCursorPosition(point);
//}
//
//void Os::ShowMouseCursor(bool is_shown) {
//  if (is_shown) {
//    while (!CGCursorIsVisible()) {
//      CGDisplayShowCursor(kCGDirectMainDisplay);
//    }
//  } else {
//    while (CGCursorIsVisible()) {
//      CGDisplayHideCursor(kCGDirectMainDisplay);
//    }
//  }
//}
//
//void Os::RefreshJoysticks(OsWindowData *window) {
//}
//
//int Os::GetNumJoysticks(OsWindowData *window) {
//  int num_joysticks = 0;
//  for (int i = 0; i < joystick_devices.size(); i++) {
//    if (joystick_devices[i] != NULL) {
//      num_joysticks++;
//    }
//  }
//  return num_joysticks;
//}
//
//// Threading functions
//// ===================
//
//#include <pthread.h>
//
//void Os::StartThread(void(*thread_function)(void*), void* data) {
//  pthread_t thread;
//  if (pthread_create(&thread, NULL, (void*(*)(void*))thread_function, data) != 0) {
//    printf("Error forking thread\n");
//  }
//}
//
//struct OsMutex {
//  pthread_mutex_t mutex;
//};
//
//OsMutex* Os::NewMutex() {
//  OsMutex* mutex = new OsMutex;
//  pthread_mutex_init(&mutex->mutex, NULL);
//  return mutex;
//}
//
//void Os::DeleteMutex(OsMutex* mutex) {
//  pthread_mutex_destroy(&mutex->mutex);
//  delete mutex;
//}
//
//void Os::AcquireMutex(OsMutex* mutex) {
//  pthread_mutex_lock(&mutex->mutex);
//}
//
//void Os::ReleaseMutex(OsMutex* mutex) {
//  pthread_mutex_unlock(&mutex->mutex);
//}
//
//// Miscellaneous functions
//// =======================
//
//void Os::MessageBox(const string& title, const string& message) {
//  DialogRef the_item;
//  DialogItemIndex item_index;
//
//  CFStringRef cf_title;
//  cf_title = CFStringCreateWithCString(NULL, title.c_str(), kCFStringEncodingASCII);  
//  CFStringRef cf_message;
//  cf_message = CFStringCreateWithCString(NULL, message.c_str(), kCFStringEncodingASCII);  
//
//  CreateStandardAlert(kAlertStopAlert, cf_title, cf_message, NULL, &the_item);
//  RunStandardAlert(the_item, NULL, &item_index);
//
//  CFRelease(cf_title);
//  CFRelease(cf_message);
//}
//
//// TODO(jwills): Currently we only deal with the main display, decide whether or not we should be
//// able to display on other or multiple displays.
//vector<pair<int, int> > Os::GetFullScreenModes() {
//  CFArrayRef modes = CGDisplayAvailableModes(kCGDirectMainDisplay);
//  set<pair<int,int> > modes_set;
//  for (int i = 0; i < CFArrayGetCount(modes); i++) {
//    CFDictionaryRef attributes = (CFDictionaryRef)CFArrayGetValueAtIndex(modes, i);
//    CFNumberRef width_number  = (CFNumberRef)CFDictionaryGetValue(attributes, kCGDisplayWidth);
//    CFNumberRef height_number = (CFNumberRef)CFDictionaryGetValue(attributes, kCGDisplayHeight);
//    int width;
//    int height;
//    CFNumberGetValue(width_number,  kCFNumberIntType, &width);
//    CFNumberGetValue(height_number, kCFNumberIntType, &height);
//    modes_set.insert(pair<int,int>(width, height));
//  }
//  return vector<pair<int,int> >(modes_set.begin(), modes_set.end());
//}
//
//void Os::Sleep(int t) {
//  usleep(t*1000);
//}
//
//int Os::GetTime() {
//  UnsignedWide current_time;
//  Microseconds(&current_time);
//  unsigned long long start_time = (unsigned long long)glop_start_time.hi << 32 | glop_start_time.lo;
//  return (((unsigned long long)current_time.hi << 32 | current_time.lo) - start_time) / 1000.0;
//}
//
void SwapBuffers(void* data) {
  aglSwapBuffers(((windowData*)(data))->agl);
}
//
//int Os::GetRefreshRate() {
//  CFDictionaryRef mode_info;
//  int refresh_rate = 60; // Assume LCD screen
//  mode_info = CGDisplayCurrentMode(CGMainDisplayID());
//  if (mode_info) {
//    CFNumberRef value = (CFNumberRef)CFDictionaryGetValue(mode_info, kCGDisplayRefreshRate);
//    if (value) {
//      CFNumberGetValue(value, kCFNumberIntType, &refresh_rate);
//      if (refresh_rate == 0) {
//        refresh_rate = 60;
//      }
//    }
//  }
//  return refresh_rate;
//}
//
//void Os::EnableVSync(bool is_enable) {
//  set<OsWindowData*>::iterator it;
//  GLint enable = (is_enable ? 1 : 0);
//  for (it = all_windows.begin(); it != all_windows.end(); it++) {
//    aglSetInteger((*it)->agl_context, AGL_SWAP_INTERVAL, &enable);
//  }
//}
//
//
//vector<string> Os::ListFiles(const string &directory) {
//  return vector<string>();
//}
//
//vector<string> Os::ListSubdirectories(const string &directory) {
//  return vector<string>();
//}
//
//#endif // MACOSX
