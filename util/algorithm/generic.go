package algorithm

import (
  "fmt"
  "reflect"
)

func Choose(_a interface{}, choose func(interface{}) bool) interface{} {
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
