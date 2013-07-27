g++ -mmacosx-version-min=10.5 -arch x86_64 -m64 -fPIC -c -Iinclude -o glop.o glop.mm
g++ -mmacosx-version-min=10.5 -arch x86_64 -m64 -weak_library /usr/lib/libSystem.B.dylib -install_name @executable_path/../lib/libglop.so -shared -dynamiclib -W1 -o libglop.so glop.o  -framework Cocoa -framework OpenGL -framework IOKit
g++ -mmacosx-version-min=10.5 -arch x86_64 -m64 -weak_library /usr/lib/libSystem.B.dylib -shared -dynamiclib -W1 -o libglopLOCAL.so glop.o  -framework Cocoa -framework OpenGL -framework IOKit

rm -f glop.o
mkdir -p lib
mv libglop.so lib/libglop.so
mv libglopLOCAL.so lib/libglopLOCAL.so
