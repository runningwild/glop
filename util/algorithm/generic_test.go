package algo_test

import (
  . "gospec"
  "gospec"
  "glop/util/algorithm"
)

func GenericSpec(c gospec.Context) {
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
