gcc -mmacosx-version-min=10.5 -arch x86_64 -m64 -fPIC -c -Iinclude -o glop.o glop.m
gcc -mmacosx-version-min=10.5 -arch x86_64 -m64 -weak_library /usr/lib/libSystem.B.dylib -install_name @executable_path/../lib/libglop.so -shared -dynamiclib -W1 -o libglop.so glop.o  -framework Cocoa -framework OpenGL

rm -f glop.o
mkdir -p lib
mv libglop.so lib/libglop.so

