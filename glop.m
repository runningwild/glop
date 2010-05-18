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
  return (void*)window;
}

void ShutDown() {
  [pool drain];
}

void Run() {
  [NSApp run];
}

