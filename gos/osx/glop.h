#ifndef __GLOP_H__
#define __GLOP_H__

// GlopKey devices
#define glopDeviceKeyboard  -1
#define glopDeviceDerived  -2
#define glopMinDevice  -2

#define kAnyKey -1
#define kNoKey -2

#define kKeyBackspace  8
#define kKeyTab  9
#define kKeyEnter  13
#define kKeyReturn  13
#define kKeyEscape  27

#define kKeyF1  129
#define kKeyF2  130
#define kKeyF3  131
#define kKeyF4  132
#define kKeyF5  133
#define kKeyF6  134
#define kKeyF7  135
#define kKeyF8  136
#define kKeyF9  137
#define kKeyF10  138
#define kKeyF11  139
#define kKeyF12  140

#define kKeyCapsLock  150
#define kKeyNumLock  151
#define kKeyScrollLock  152
#define kKeyPrintScreen  153
#define kKeyPause  154
#define kKeyLeftShift  155
#define kKeyRightShift  156
#define kKeyLeftControl  157
#define kKeyRightControl  158
#define kKeyLeftAlt  159
#define kKeyRightAlt  160
#define kKeyLeftGui  161
#define kKeyRightGui  162

#define kKeyRight  166
#define kKeyLeft  167
#define kKeyUp  168
#define kKeyDown  169

#define kKeyPadDivide  170
#define kKeyPadMultiply  171
#define kKeyPadSubtract  172
#define kKeyPadAdd  173
#define kKeyPadEnter  174
#define kKeyPadDecimal  175
#define kKeyPadEquals  176
#define kKeyPad0  177
#define kKeyPad1  178
#define kKeyPad2  179
#define kKeyPad3  180
#define kKeyPad4  181
#define kKeyPad5  182
#define kKeyPad6  183
#define kKeyPad7  184
#define kKeyPad8  185
#define kKeyPad9  186

#define kKeyDelete  190
#define kKeyHome  191
#define kKeyInsert  192
#define kKeyEnd  193
#define kKeyPageUp  194
#define kKeyPageDown  195

// Key constants - mouse
#define kFirstMouseKeyIndex  300
#define kNumMouseButtons  8
#define kMouseUp  300
#define kMouseRight  301
#define kMouseDown  302
#define kMouseLeft  303
#define kMouseWheelUp  304
#define kMouseWheelDown  305
#define kMouseLButton  306
#define kMouseRButton  307
#define kMouseMButton  308




void Init();
void CreateWindow(void**, void**, int, int, int, int);


typedef struct {
  short index;
  short device;
  float press_amt;
  int mouse_dx;
  int mouse_dy;
  long long timestamp;
  int cursor_x;
  int cursor_y;
  int num_lock;
  int caps_lock;
} KeyEvent;
void GetInputEvents(void**, int*);
// GetInputEvents(KeyEvent**, length*);

void Run();
void SwapBuffers(void*);
void Think();
void Quit();

void CurrentMousePos(void*,void*,void*);

#endif
