package algorithm

import (
  "fmt"
  "reflect"
)

type Chooser func(interface{}) bool

// Given a slice and a Chooser, returns a slice of the same type as the input
// slice that contains only those elements of the input slice for which
// choose() returns true.  The elements of the returned slice will be in the
// same order that they were in in the input slice.
func Choose(_a interface{}, choose Chooser) interface{} {
  a := reflect.ValueOf(_a)
  if a.Kind() != reflect.Slice {
    panic(fmt.Sprintf("Can only Choose from a slice, not a %v", a))
  }
  count := 0
  for i := 0; i < a.Len(); i++ {
    if choose(a.Index(i).Interface()) {
      count++
    }
  }
  ret := reflect.MakeSlice(a.Type(), count, count)
  cur := 0
  for i := 0; i < a.Len(); i++ {
    if choose(a.Index(i).Interface()) {
      ret.Index(cur).Set(a.Index(i))
      cur++
    }
  }
  return ret.Interface()
}

type Mapper func(a interface{}) interface{}

func Map(_a interface{}, _b interface{}, mapper Mapper) interface{} {
  a := reflect.ValueOf(_a)
  if a.Kind() != reflect.Slice {
    panic(fmt.Sprintf("Can only Map from a slice, not a %v", a))
  }

  b := reflect.ValueOf(_b)
  if b.Kind() != reflect.Slice {
    panic(fmt.Sprintf("Can only Map to a slice, not a %v", b))
  }

  ret := reflect.MakeSlice(b.Type(), a.Len(), a.Len())
  for i := 0; i < a.Len(); i++ {
    el := reflect.ValueOf(mapper(a.Index(i).Interface()))
    ret.Index(i).Set(el)
  }

  return ret.Interface()
}
