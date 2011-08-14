#import <Cocoa/Cocoa.h>
#import <OpenGL/gl.h>
#import <glop.h>
#import <mach/mach_time.h>
#import <stdio.h>

NSAutoreleasePool* pool;
NSApplication* glop_app;

KeyEvent* glop_key_events;
int glop_key_events_cap;
int glop_key_events_len;

@interface GlopApplication : NSApplication {
}
- (void)sendEvent:(NSEvent*)event;
@end

@implementation GlopApplication
- (void)sendEvent:(NSEvent*)event {
  if ([event type] == NSApplicationDefined) {
    [glop_app stop:self];
  } else if ([event type] == NSLeftMouseDown) {
    KeyEvent* key_event = &glop_key_events[glop_key_events_len];
    glop_key_events_len++;
    key_event->index = 0;
    key_event->device = 1;
    key_event->press_amt = 1;
  } else {
    [super sendEvent: event];
  }
}
@end

void Init() {
  glop_app = [GlopApplication sharedApplication];
  pool = [[NSAutoreleasePool alloc] init];

  glop_key_events_len = 0;
  glop_key_events_cap = 1000;  // TODO: Should we actually double if we get more than 1000 events
  glop_key_events = (KeyEvent*)malloc(sizeof(KeyEvent) * glop_key_events_len);
}

void Think() {
  uint64_t uptime = mach_absolute_time();
  NSEvent* event = [NSEvent otherEventWithType:NSApplicationDefined location:NSZeroPoint modifierFlags:0 timestamp:(NSTimeInterval)uptime windowNumber:0 context:0 subtype:0 data1:0 data2:0];
  [glop_app postEvent:event atStart:FALSE];
  [glop_app run];
}

void GetInputEvents(void** _key_events, int* length) {
  *_key_events = glop_key_events;
  *length = glop_key_events_len;
  glop_key_events_len = 0;
}

void CreateWindow(void** _window, void** _context, int x, int y, int width, int height) {
  NSRect windowRect = NSMakeRect(x, y, width, height);
  NSWindow* window = [NSWindow alloc];
  *((NSWindow**)(_window)) = window;
  [window initWithContentRect:windowRect 
  styleMask:( NSResizableWindowMask | NSClosableWindowMask | NSTitledWindowMask) 
  backing:NSBackingStoreBuffered defer:NO];
  [window makeKeyAndOrderFront:nil];

  // Create and bind an OpenGL context
  NSOpenGLPixelFormatAttribute attributes[] = {
    NSOpenGLPFADoubleBuffer,
    NSOpenGLPFAAccelerated,
    NSOpenGLPFAColorSize, 32,
    NSOpenGLPFADepthSize, 32,
    //    NSOpenGLPFAFullScreen,
    0,
  };
  NSOpenGLPixelFormat* pixel_format = [[NSOpenGLPixelFormat alloc] initWithAttributes:attributes];
  if (pixel_format == nil) {
    // TODO: How do we signal this properly?
    exit(0);
    return;
  }
  NSOpenGLContext* context = [NSOpenGLContext alloc];
  *((NSOpenGLContext**)(_context)) = context;
  [context initWithFormat:pixel_format shareContext:NO];
  [context setView:[window contentView]];
  [context makeCurrentContext];
  glClear(GL_COLOR_BUFFER_BIT);
  [context flushBuffer];
}

void SwapBuffers(void* _context) {
  NSOpenGLContext* context = (NSOpenGLContext*)(_context);
  [context flushBuffer];
}

void ShutDown() {
  [pool drain];
}

void Run() {
  [NSApp run];
}

void CurrentMousePos(void* _window, void* _x, void* _y) {
  NSWindow* window = (NSWindow*)_window;
  int* x = (int*)_x;
  int* y = (int*)_y;
  NSPoint point = [window mouseLocationOutsideOfEventStream];
  *x = (int)point.x;
  *y = (int)point.y;
}
