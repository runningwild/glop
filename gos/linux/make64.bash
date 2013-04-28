g++ -m64 -fPIC -o glop.o -c -Iinclude glop.cpp
g++ -m64 -fPIC -shared -dynamiclib -W1 -o libglop.so glop.o
rm -f glop.o
mkdir -p lib
mv libglop.so lib
