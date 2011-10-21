#ifndef __GLOP_H__
#define __GLOP_H__

void GlopInit();
void* GlopCreateWindow(
    void* title,
    int x,
    int y,
    int width,
    int height);
void GlopThink();
void GlopSwapBuffers();

void GlopGetMousePosition(int* x, int* y);
void GlopGetWindowDims(int* x, int* y, int* dx, int* dy);



/*

//void CreateWindow(void**, void**, int, int, int, int);

void GlopSwapBuffers(void*);

void GlopThink();

typedef struct {
  short index;
  short device;
  float press_amt;
  long long timestamp;
  int cursor_x;
  int cursor_y;
  int num_lock;
  int caps_lock;
} GlopKeyEvent;
void GlopClearKeyEvent(GlopKeyEvent* event) {
  event->index = 0;
  event->device = 0;
  event->press_amt = 0;
  event->timestamp = 0;
  event->cursor_x = 0;
  event->cursor_y = 0;
  event->num_lock = 0;
  event->caps_lock = 0;
}

void GlopGetInputEvents(void* _window, void** _events_ret, void* _num_events, void* _horizon);

void GlopGetMousePosition(int* x,int* y);
void GlopGetWindowDims(void* _window, int* x, int* y, int* dx, int* dy);

void GlopEnableVSync(int);

*/
#endif
