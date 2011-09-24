package algorithm

import (
  "container/heap"
  "sort"
)

type Graph interface {
  NumVertex() int
  Adjacent(int) ([]int, []float64)
}

type dNode struct {
  v int  // current vertex
  p int  // previous vertex (for extracting path)
  weight float64
}

type dArray []dNode
func (da *dArray) Len() int {
  return len(*da)
}
func (da *dArray) Swap(i,j int) {
  (*da)[i],(*da)[j] = (*da)[j],(*da)[i]
}
func (da *dArray) Less(i,j int) bool {
  return (*da)[i].weight < (*da)[j].weight
}
func (da *dArray) Push(x interface{}) {
  *da = append(*da, x.(dNode))
}
func (da *dArray) Pop() interface{} {
  val := (*da)[len(*da)-1]
  *da = (*da)[0:len(*da)-1]
  return val
}

// Returns the list of vertices that can be reached from the vertices in src with total
// path weight <= limit.
func ReachableWithinLimit(g Graph, src []int, limit float64) []int {
  used := make([]bool, g.NumVertex())
  h := make(dArray, len(src))
  for i,s := range src {
    h[i] = dNode{ v:s, weight:0 }
  }

  for len(h) > 0 {
    cur := heap.Pop(&h).(dNode)
    if used[cur.v] { continue }
    if cur.weight > limit { break }
    used[cur.v] = true
    adj,weights := g.Adjacent(cur.v)
    for i := range adj {
      heap.Push(&h, dNode{ v:adj[i], weight:weights[i]+cur.weight })
    }
  }

  var ret []int
  for v := range used {
    if used[v] {
      ret = append(ret, v)
    }
  }

  sort.Ints(ret)
  return ret
}

func Dijkstra(g Graph, src []int, dst []int) (float64, []int) {
  used := make([]bool, g.NumVertex())
  conn := make([]int, g.NumVertex())
  h := make(dArray, len(src))
  for i,s := range src {
    h[i] = dNode{ v:s, p:-1, weight:0 }
  }
  target := make(map[int]bool, len(dst))
  for _,d := range dst {
    target[d] = true
  }

  for len(h) > 0 {
    cur := heap.Pop(&h).(dNode)
    if used[cur.v] { continue }
    used[cur.v] = true
    conn[cur.v] = cur.p
    if _,ok := target[cur.v]; ok {
      // Extract the path
      var path []int
      c := cur.v
      for c != -1 {
        path = append(path, c)
        c = conn[c]
      }
      // The path comes out backwards, so reverse it
      for i := 0; i < len(path)/2; i++ {
        path[i],path[len(path) - i - 1] = path[len(path) - i - 1],path[i]
      }
      return cur.weight, path
    }
    adj,weights := g.Adjacent(cur.v)
    for i := range adj {
      heap.Push(&h, dNode{ v:adj[i], p:cur.v, weight:weights[i]+cur.weight })
    }
  }
  return -1, nil
}

