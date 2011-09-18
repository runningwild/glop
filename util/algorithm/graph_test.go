package algo_test

import (
  . "gospec"
  "gospec"
  "glop/util/algorithm"
)

type Graph interface {
  NumVertex() int
  Adjacent(int) ([]int, []float64)
}

type board [][]int
func (b board) NumVertex() int {
  return len(b) * len(b[0])
}
func (b board) Adjacent(n int) ([]int, []float64) {
  x := n%len(b[0])
  y := n/len(b[0])
  var adj []int
  var weight []float64
  if x > 0 {
    adj = append(adj, n-1)
    weight = append(weight, float64(b[y][x-1]))
  }
  if y > 0 {
    adj = append(adj, n-len(b[0]))
    weight = append(weight, float64(b[y-1][x]))
  }
  if x < len(b[0])-1 {
    adj = append(adj, n+1)
    weight = append(weight, float64(b[y][x+1]))
  }
  if y < len(b)-1 {
    adj = append(adj, n+len(b[0]))
    weight = append(weight, float64(b[y+1][x]))
  }
  return adj, weight
}


func GraphSpec(c gospec.Context) {
  b := [][]int{
    []int{ 1,2,9,4,3,2,1 },  // 0 - 6
    []int{ 9,2,9,4,3,1,1 },  // 7 - 13
    []int{ 2,1,5,5,5,2,1 },  // 14 - 20
    []int{ 1,1,1,1,1,1,1 },  // 21 - 27
  }
  c.Specify("Check Dijkstra's gives the right path and weight", func() {
    weight,path := algorithm.Dijkstra(board(b), []int{0}, []int{11})
    c.Expect(weight, Equals, 16.0)
    c.Expect(path, ContainsInOrder, []int{ 0, 1, 8, 15, 22, 23, 24, 25, 26, 19, 12, 11 })
  })
  c.Specify("Check multiple sources", func() {
    weight,path := algorithm.Dijkstra(board(b), []int{ 0, 1, 7, 2 }, []int{11})
    c.Expect(weight, Equals, 10.0)
    c.Expect(path, ContainsInOrder, []int{ 2, 3, 4, 11 })
  })
  c.Specify("Check multiple destinations", func() {
    weight,path := algorithm.Dijkstra(board(b), []int{ 0 }, []int{ 6, 11, 21 })
    c.Expect(weight, Equals, 7.0)
    c.Expect(path, ContainsInOrder, []int{ 0, 1, 8, 15, 22, 21 })
  })
}