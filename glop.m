#import <Cocoa/Cocoa.h>
#import <OpenGL/gl.h>
#import <glop.h>

NSAutoreleasePool *pool;

void startEventListener() {
  [NSEvent
    addLocalMonitorForEventsMatchingMask:NSApplicationDefinedMask
    handler:^(NSEvent *incomingEvent) {
      if ([incomingEvent subtype] == 0) {
        [NSApp stop:incomingEvent];
      }
      incomingEvent = nil;
      return incomingEvent;
    }
  ];
  [NSEvent
    addLocalMonitorForEventsMatchingMask:NSKeyDownMask
    handler:^(NSEvent *incomingEvent) {
      [NSApp stop:incomingEvent];
      incomingEvent = nil;
      exit(0);
      return incomingEvent;
    }
  ];
}

void Init() {
  pool = [[NSAutoreleasePool alloc] init];
  [NSApplication sharedApplication];

  startEventListener();
}

void CreateWindow(void** _window, void** _context) {
  NSRect windowRect = NSMakeRect(10.0f, 10.0f, 512.0f, 512.0f);
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

void Think() {
  NSEvent* event = [NSEvent otherEventWithType:NSApplicationDefined location:NSZeroPoint modifierFlags:0 timestamp:[[NSProcessInfo processInfo] systemUptime] windowNumber:0 context:0 subtype:0 data1:0 data2:0];
  [NSApp postEvent:event atStart:FALSE];
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
