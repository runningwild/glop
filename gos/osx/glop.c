#include "glop.h"
#include <stdio.h>

int main() {
  int w;
  int c;
  int* pw = &w;
  int* pc = &c;
  Init();
  int foo = 0;
  int* pfoo = &foo;
  int bar= 0;
  int* pbar = &bar;
  CreateWindow((void**)&pw,(void**)&pc,300,100,200,200);
int i;
  for (i=0;;i++) {
if (i%100==0) {
  printf("i: %d\n", i);
}
    Think();
    GetInputEvents((void**)&pfoo, (void*)pbar);
    if (bar > 10) { return 0; }
    SwapBuffers((void*)pc);
  }
  return 0;
}

