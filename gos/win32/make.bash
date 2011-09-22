g++ -o glop.o -m32 -c -I../include glop.cpp
g++ -o libglop.dll glop.o -shared -lopengl32 -lgdi32 -ldxguid -lwinmm -ldinput
rm glop.o

mkdir -p ../lib
mv libglop.dll ../lib/libglop.dll
