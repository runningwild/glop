6g -o example.6 example.go
6l -o example example.6

rm -rf example.app
mkdir -p example.app/Contents/MacOS
mkdir example.app/Contents/fonts
mkdir example.app/Contents/lib

mv example example.app/Contents/MacOS/example
cp fonts/* example.app/Contents/fonts/
cp ../gos/osx/libglop.so example.app/Contents/lib/
rm -f example
rm example.6

