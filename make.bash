mkdir -p obj_glop
mkdir -p lib
gcc -c -o obj_glop/glop_m.o glop.m
libtool -o lib/libglop.a obj_glop/glop_m.o


# Need to sudo for this line
cp glop.h /usr/local/include/glop.h

6g -o rawr.6 rawr.go
6l -o rawr rawr.6
mkdir -p rawr.app/Contents/MacOS
cp rawr rawr.app/Contents/MacOS/rawr
