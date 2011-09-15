6g -o rawr.6 rawr.go
6l -o rawr rawr.6

rm -rf rawr.app
mkdir -p rawr.app/Contents/MacOS
mkdir rawr.app/Contents/fonts
mkdir rawr.app/Contents/lib

mv rawr rawr.app/Contents/MacOS/rawr
cp fonts/* rawr.app/Contents/fonts/
cp ../gos/osx/libglop.so rawr.app/Contents/lib/
