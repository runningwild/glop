#ifndef __GLOP_H__
#define __GLOP_H__

void Init();
void CreateWindow(void**, void**, int, int, int, int);


typedef struct {
  short index;
  short device;
  float press_amt;
  int mouse_dx;
  int mouse_dy;
  int timestamp;
  int cursor_x;
  int cursor_y;
  int num_lock;
  int  caps_lock;
} KeyEvent;
void GetInputEvents(void**, int*);
// GetInputEvents(KeyEvent**, length*);

void Run();
void SwapBuffers(void*);
void Think();

void CurrentMousePos(void*,void*,void*);

#endif
