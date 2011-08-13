#import <Cocoa/Cocoa.h>
#import <OpenGL/gl.h>
#import <glop.h>
#import <mach/mach_time.h>
#import <stdio.h>

NSAutoreleasePool* pool;
NSApplication* glop_app;

@interface GlopApplication : NSApplication {
}
- (void)sendEvent:(NSEvent*)event;
@end

@implementation GlopApplication
- (void)sendEvent:(NSEvent*)event {
  [glop_app stop:self];
  [super sendEvent: event];
}
@end

void Init() {
  glop_app = [GlopApplication sharedApplication];
  pool = [[NSAutoreleasePool alloc] init];
}

void Think() {
  uint64_t uptime = mach_absolute_time();
printf("uptime: %ull\n", uptime);
  NSEvent* event = [NSEvent otherEventWithType:NSApplicationDefined location:NSZeroPoint modifierFlags:0 timestamp:(NSTimeInterval)uptime windowNumber:0 context:0 subtype:0 data1:0 data2:0];
printf("made event\n");
  [glop_app postEvent:event atStart:FALSE];
printf("posted event\n");
  [glop_app run];
printf("ran\n");
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
