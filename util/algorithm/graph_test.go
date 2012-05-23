package algorithm_test

import (
  . "github.com/orfjackal/gospec/src/gospec"
  "github.com/orfjackal/gospec/src/gospec"
  "github.com/runningwild/glop/util/algorithm"
)

type board [][]int

func (b board) NumVertex() int {
  return len(b) * len(b[0])
}
func (b board) Adjacent(n int) ([]int, []float64) {
  x := n % len(b[0])
  y := n / len(b[0])
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

func DijkstraSpec(c gospec.Context) {
  b := [][]int{
    []int{1, 2, 9, 4, 3, 2, 1}, // 0 - 6
    []int{9, 2, 9, 4, 3, 1, 1}, // 7 - 13
    []int{2, 1, 5, 5, 5, 2, 1}, // 14 - 20
    []int{1, 1, 1, 1, 1, 1, 1}, // 21 - 27
  }
  c.Specify("Check Dijkstra's gives the right path and weight", func() {
    weight, path := algorithm.Dijkstra(board(b), []int{0}, []int{11})
    c.Expect(weight, Equals, 16.0)
    c.Expect(path, ContainsInOrder, []int{0, 1, 8, 15, 22, 23, 24, 25, 26, 19, 12, 11})
  })
  c.Specify("Check multiple sources", func() {
    weight, path := algorithm.Dijkstra(board(b), []int{0, 1, 7, 2}, []int{11})
    c.Expect(weight, Equals, 10.0)
    c.Expect(path, ContainsInOrder, []int{2, 3, 4, 11})
  })
  c.Specify("Check multiple destinations", func() {
    weight, path := algorithm.Dijkstra(board(b), []int{0}, []int{6, 11, 21})
    c.Expect(weight, Equals, 7.0)
    c.Expect(path, ContainsInOrder, []int{0, 1, 8, 15, 22, 21})
  })
}

func ReachableSpec(c gospec.Context) {
  b := [][]int{
    []int{1, 2, 9, 4, 3, 2, 1}, // 0 - 6
    []int{9, 2, 9, 4, 3, 1, 1}, // 7 - 13
    []int{2, 1, 5, 5, 5, 2, 1}, // 14 - 20
    []int{1, 1, 1, 1, 1, 1, 1}, // 21 - 27
  }
  c.Specify("Check reachability", func() {
    reach := algorithm.ReachableWithinLimit(board(b), []int{3}, 5)
    c.Expect(reach, ContainsInOrder, []int{3, 4, 5, 10})
    reach = algorithm.ReachableWithinLimit(board(b), []int{3}, 10)
    c.Expect(reach, ContainsInOrder, []int{2, 3, 4, 5, 6, 10, 11, 12, 13, 17, 19, 20, 24, 25, 26, 27})
  })
  c.Specify("Check reachability with multiple sources", func() {
    reach := algorithm.ReachableWithinLimit(board(b), []int{0, 6}, 3)
    c.Expect(reach, ContainsInOrder, []int{0, 1, 5, 6, 12, 13, 20, 27})
    reach = algorithm.ReachableWithinLimit(board(b), []int{21, 27}, 2)
    c.Expect(reach, ContainsInOrder, []int{13, 14, 15, 20, 21, 22, 23, 25, 26, 27})
  })
  c.Specify("Check bounds with multiple sources", func() {
    reach := algorithm.ReachableWithinBounds(board(b), []int{0, 6}, 2, 4)
    c.Expect(reach, ContainsInOrder, []int{1, 5, 8, 12, 19, 20, 26, 27})
  })
}

type adag [][]int
func (a adag) NumVertex() int {
  return len(a)
}
func (a adag) Successors(n int) []int {
  return a[n]
}
func (a adag) allSuccessorsHelper(n int, m map[int]bool) {
  for _,s := range a[n] {
    m[s] = true
    a.allSuccessorsHelper(s, m)
  }
}
func (a adag) AllSuccessors(n int) map[int]bool {
  if len(a[n]) == 0 { return nil }
  m := make(map[int]bool)
  a.allSuccessorsHelper(n, m)
  return m
}
func checkOrder(c gospec.Context, a adag, order []int) {
  c.Expect(len(a), Equals, len(order))
  c.Specify("Ordering contains all vertices exactly once", func() {
    all := make(map[int]bool)
    for _,v := range order {
      all[v] = true
    }
    c.Expect(len(all), Equals, len(order))
    for i := 0; i < len(a); i++ {
      c.Expect(all[i], Equals, true)
    }
  })
  c.Specify("Successors of a vertex always occur later in the ordering", func() {
    for i := 0; i < len(order); i++ {
      all := a.AllSuccessors(order[i])
      for j := range order {
        if i == j { continue }
        succ,ok := all[order[j]]
        if j < i {
          c.Expect(!ok, Equals, true)
        } else {
          c.Expect(!ok || succ, Equals, true)
        }
      }
    }
  })
}
func TopoSpec(c gospec.Context) {
  c.Specify("Check toposort on linked list", func() {
    a := adag{
      []int{ 1 },
      []int{ 2 },
      []int{ 3 },
      []int{ 4 },
      []int{ 5 },
      []int{ 6 },
      []int{ },
    }
    order := algorithm.TopoSort(a)
    checkOrder(c, a, order)
  })

  c.Specify("multi-edges don't mess up toposort", func() {
    a := adag{
      []int{ 1, 1, 1 },
      []int{ },
    }
    order := algorithm.TopoSort(a)
    checkOrder(c, a, order)
  })

  c.Specify("Check toposort on a more complicated digraph", func() {
    a := adag{
      []int{ 8, 7, 4 },  // 0
      []int{ 5 },
      []int{ 0 },
      []int{ 9 },
      []int{ 14 },
      []int{ 15 },  // 5
      []int{ 1 },
      []int{  },
      []int{  },
      []int{ 13 },
      []int{ 3 },  // 10
      []int{ 12 },
      []int{ 18 },
      []int{ 16 },
      []int{  },
      []int{ 14 },  // 15
      []int{  },
      []int{  },
      []int{  },
      []int{  },
      []int{  },
    }
    order := algorithm.TopoSort(a)
    checkOrder(c, a, order)
  })

  c.Specify("A cyclic digraph returns nil", func() {
    a := adag{
      []int{ 8, 7, 4 },  // 0
      []int{ 5 },
      []int{ 0 },
      []int{ 9 },
      []int{ 14 },
      []int{ 15 },  // 5
      []int{ 1 },
      []int{  },
      []int{ 20 },
      []int{ 13 },
      []int{ 3 },  // 10
      []int{ 12 },
      []int{ 18 },
      []int{ 16 },
      []int{ 2 },
      []int{ 14 },  // 15
      []int{ 6 },
      []int{  },
      []int{  },
      []int{  },
      []int{  },
    }
    order := algorithm.TopoSort(a)
    c.Expect(len(order), Equals, 0)
  })
}

