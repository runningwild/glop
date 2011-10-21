# Need to sudo for this line
cp glop-linux.h /usr/local/include/glop.h

mkdir -p obj_glop
mkdir -p lib
gcc -fPIC -m32 -c -o obj_glop/glop.o glop-linux.c
gcc -shared -W1 -o lib/libglop.so obj_glop/glop.o
cp lib/libglop.so /usr/local/lib/libglop.so

make clean
make
make install

#8g -o rawr.6 rawr.go
#8l -o rawr rawr.6

