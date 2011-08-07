mkdir -p obj_glop
mkdir -p lib
gcc -arch x86_64 -fPIC -c -o obj_glop/glop_c.o glop.c
gcc -arch x86_64 -install_name @executable_path/../lib/libglop.so -shared -W1 -o lib/libglop.so obj_glop/glop_c.o -framework Carbon -framework OpenGL -framework AGL -framework ApplicationServices -framework IOKit

