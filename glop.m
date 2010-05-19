#import <Cocoa/Cocoa.h>

NSAutoreleasePool *pool;

void startEventListener() {
  [NSEvent
    addLocalMonitorForEventsMatchingMask:NSKeyDownMask
    handler:^(NSEvent *incomingEvent) {
      [NSApp stop:incomingEvent];
      incomingEvent = nil;
      return incomingEvent;
    }
  ];
}

void Init() {
  pool = [[NSAutoreleasePool alloc] init];
  [NSApplication sharedApplication];

  startEventListener();
}

void* CreateWindow() {
  NSRect windowRect = NSMakeRect(10.0f, 10.0f, 800.0f, 600.0f);
  NSWindow *window = [[NSWindow alloc] initWithContentRect:windowRect 
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
    return;
  }

  NSOpenGLContext* context = [[NSOpenGLContext alloc] initWithFormat:pixel_format shareContext:NO];
  [context setView:[window contentView]];

  return (void*)window;
}

void ShutDown() {
  [pool drain];
}

void Run() {
  [NSApp run];
}

