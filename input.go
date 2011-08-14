package main

type Key struct {
  index  uint16
  device uint16
}

func (k Key) String() string {
  return "make me!"
}

type KeyEvent struct {

}

