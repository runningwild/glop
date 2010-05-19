package glop

// #include <glop.h>
import "C"

func Init() {
  C.Init()
}

func Run() {
  C.Run()
}

func ShutDown() {
  C.ShutDown()
}

func CreateWindow() {
  C.CreateWindow()
}
