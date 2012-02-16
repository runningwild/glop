package algorithm_test

import (
  . "github.com/orfjackal/gospec/src/gospec"
  "github.com/orfjackal/gospec/src/gospec"
  "github.com/runningwild/glop/util/algorithm"
  "fmt"
)

func ChooserSpec(c gospec.Context) {
  c.Specify("Choose on []int", func() {
    a := []int{0,1,2,3,4,5,6,7,8,9}
    var b []int
    b = algorithm.Choose(a, func(v interface{}) bool { return v.(int) % 2 == 0 }).([]int)
    c.Expect(b, ContainsInOrder, []int{0, 2, 4, 6, 8})

    b = algorithm.Choose(a, func(v interface{}) bool { return v.(int) % 2 == 1 }).([]int)
    c.Expect(b, ContainsInOrder, []int{1, 3, 5, 7, 9})

    b = algorithm.Choose(a, func(v interface{}) bool { return true }).([]int)
    c.Expect(b, ContainsInOrder, a)

    b = algorithm.Choose(a, func(v interface{}) bool { return false }).([]int)
    c.Expect(b, ContainsInOrder, []int{})

    b = algorithm.Choose([]int{}, func(v interface{}) bool { return false }).([]int)
    c.Expect(b, ContainsInOrder, []int{})
  })

  c.Specify("Choose on []string", func() {
    a := []string{"foo", "bar", "wing", "ding", "monkey", "machine"}
    var b []string
    b = algorithm.Choose(a, func(v interface{}) bool { return v.(string) > "foo" }).([]string)
    c.Expect(b, ContainsInOrder, []string{"wing", "monkey", "machine"})

    b = algorithm.Choose(a, func(v interface{}) bool { return v.(string) < "foo" }).([]string)
    c.Expect(b, ContainsInOrder, []string{"bar", "ding"})
  })
}

func MapperSpec(c gospec.Context) {
  c.Specify("Map from []int to []float64", func() {
    a := []int{0,1,2,3,4}
    var b []float64
    b = algorithm.Map(a, []float64{}, func(v interface{}) interface{} { return float64(v.(int)) }).([]float64)
    c.Expect(b, ContainsInOrder, []float64{0,1,2,3,4})
  })
  c.Specify("Map from []int to []string", func() {
    a := []int{0,1,2,3,4}
    var b []string
    b = algorithm.Map(a, []string{}, func(v interface{}) interface{} { return fmt.Sprintf("%d", v) }).([]string)
    c.Expect(b, ContainsInOrder, []string{"0", "1", "2", "3", "4"})
  })
}