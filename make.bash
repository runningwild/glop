# Need to sudo for this line
cp glop.h /usr/local/include/glop.h

mkdir -p obj_glop
mkdir -p lib
cp GlopView.h obj_glop/GlopView.h
gcc -c -o obj_glop/glop_m.o glop.m
#gcc -c -o obj_glop/glop_view_m.o GlopView.m
libtool -o lib/libglop.a obj_glop/glop_m.o #obj_glop/glop_view_m.h

make clean
make
make install


6g -o rawr.6 rawr.go
6l -o rawr rawr.6
mkdir -p rawr.app/Contents/MacOS
cp rawr rawr.app/Contents/MacOS/rawr
