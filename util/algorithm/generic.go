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

// //  Given a pointer to a slice and a Chooser,  type as the input
// // slice that contains only those elements of the input slice for which
// // choose() returns true.  The elements of the returned slice will be in the
// // same order that they were in in the input slice.
func Choose2(_a interface{}, chooser interface{}) {
  a := reflect.ValueOf(_a)
  if a.Kind() != reflect.Ptr || a.Elem().Kind() != reflect.Slice {
    panic(fmt.Sprintf("Can only Choose from a pointer to a slice, not a %v", a))
  }

  c := reflect.ValueOf(chooser)
  if c.Kind() != reflect.Func {
    panic(fmt.Sprintf("chooser must be a func, not a %v", c))
  }
  if c.Type().NumIn() != 1 {
    panic("chooser must take exactly 1 input parameter")
  }
  if c.Type().In(0).Kind() != a.Elem().Type().Elem().Kind() {
    panic(fmt.Sprintf("chooser's parameter must be %v, not %v", a.Addr().Elem(), c.Type().In(0)))
  }
  if c.Type().NumOut() != 1 || c.Type().Out(0).Kind() != reflect.Bool {
    panic("chooser must have exactly 1 output parameter, a bool")
  }

  count := 0
  in := make([]reflect.Value, 1)
  var out []reflect.Value
  slice := a.Elem()
  for i := 0; i < slice.Len(); i++ {
    in[0] = slice.Index(i)
    out = c.Call(in)
    if out[0].Bool() {
      if count > 0 {
        slice.Index(i-count).Set(slice.Index(i))
      }
    } else {
      count++
    }
  }
  slice.Set(slice.Slice(0, slice.Len() - count))
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

func Map2(_in interface{}, _out interface{}, mapper interface{}) {
  in := reflect.ValueOf(_in)
  if in.Kind() != reflect.Slice {
    panic(fmt.Sprintf("Can only Map from a slice, not a %v", in))
  }

  out := reflect.ValueOf(_out)
  if out.Kind() != reflect.Ptr || out.Elem().Kind() != reflect.Slice {
    panic(fmt.Sprintf("Can only Map to a pointer to a slice, not a %v", out))
  }

  m := reflect.ValueOf(mapper)
  if m.Kind() != reflect.Func {
    panic(fmt.Sprintf("mapper must be a func, not a %v", m))
  }
  if m.Type().NumIn() != 1 {
    panic("chooser must take exactly 1 input parameter")
  }
  if m.Type().In(0).Kind() != in.Type().Elem().Kind() {
    panic(fmt.Sprintf("mapper's parameter must be %v, not %v", in.Addr().Elem(), m.Type().In(0)))
  }
  if m.Type().NumOut() != 1 {
    panic("chooser must have exactly 1 output parameter")
  }
  if m.Type().Out(0).Kind() != out.Elem().Type().Elem().Kind() {
    panic(fmt.Sprintf("mapper's output parameter must be %v, not %v", out.Type().Elem().Kind(), m.Type().Out(0)))
  }

  if out.Elem().Len() < in.Len() {
    slice := reflect.MakeSlice(out.Elem().Type(), in.Len(), in.Len())
    out.Elem().Set(slice)
  }

  out.Elem().SetLen(in.Len())
  for i := 0; i < out.Elem().Len(); i++ {
    v := m.Call([]reflect.Value{in.Index(i)})
    out.Elem().Index(i).Set(v[0])
  }
}








