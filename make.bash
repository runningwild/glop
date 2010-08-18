# Need to sudo for this line
cp glop.h /usr/local/include/glop.h

mkdir -p obj_glop
mkdir -p lib
gcc -arch x86_64 -m64 -fPIC -c -o obj_glop/glop_m.o glop.m
gcc -shared -W1 -o lib/libglop.so obj_glop/glop_m.o  -framework Cocoa -framework OpenGL
cp lib/libglop.so /usr/local/lib/libglop.so

make clean
make
make install

6g -o rawr.6 rawr.go
6l -o rawr rawr.6
mkdir -p rawr.app/Contents/MacOS
cp rawr rawr.app/Contents/MacOS/rawr

