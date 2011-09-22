8g -o example.8 example.go
8l -o example.exe example.8
mv example.exe buffer/example.exe
cp ../gos/lib/libglop.dll buffer/libglop.dll

#rm -rf example.app
#mkdir -p example.app/Contents/MacOS
#mkdir example.app/Contents/fonts
#mkdir example.app/Contents/lib
#mkdir example.app/Contents/sprites

#mv example example.app/Contents/MacOS/example
#cp fonts/* example.app/Contents/fonts/
#cp ../gos/osx/libglop.so example.app/Contents/lib/
#cp -r ../sprite/test_sprite example.app/Contents/sprites/
#rm -f example
#rm example.6

