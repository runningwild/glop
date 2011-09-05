cp glop.h ../include/glop.h
gcc -mmacosx-version-min=10.5 -arch x86_64 -m64 -fPIC -c -I../include -o glop.o glop.m
gcc -mmacosx-version-min=10.5 -arch x86_64 -m64 -weak_library /usr/lib/libSystem.B.dylib -install_name @executable_path/../lib/libglop.so -shared -dynamiclib -W1 -o libglop.so glop.o  -framework Cocoa -framework OpenGL

rm glop.o
cp libglop.so ../lib/libglop.so

gcc -o rawr -I../include/ -L../lib glop.c -lglop
rm -rf rawr.app
mkdir -p rawr.app/Contents/MacOS
mv rawr rawr.app/Contents/MacOS/rawr
mkdir -p rawr.app/Contents/lib
cp ../lib/libglop.so rawr.app/Contents/lib/libglop.so