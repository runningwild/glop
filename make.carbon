mkdir -p obj_glop
mkdir -p lib
gcc -arch i386 -fPIC -c -o obj_glop/glop_c.o glop.c
gcc -arch i386 -install_name @executable_path/../lib/libglop.so -shared -W1 -o lib/libglop.so obj_glop/glop_c.o -framework Carbon -framework OpenGL -framework AGL -framework ApplicationServices -framework IOKit

make clean
make
make install

mkdir -p rawr.app/Contents/MacOS/
mkdir -p rawr.app/Contents/lib/
8g -o rawr.8 rawr.go
8l -o rawr rawr.8
cp rawr rawr.app/Contents/MacOS/rawr
cp lib/libglop.so rawr.app/Contents/lib

