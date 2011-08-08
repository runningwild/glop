#ifndef __GLOP_H__
#define __GLOP_H__

void CreateWindow(void**, int, int, int, int, int);
// windowData**, full_screen, x, y, dx, dy

void SwapBuffers(void*);
// windowData*

void Think(void*);
// windowData*

void Init();

//void CreateWindow(void**, void**, int, int, int, int);
//void Run();

//void CurrentMousePos(void*,void*,void*);

#endif
