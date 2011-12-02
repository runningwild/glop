package sprite

import (
  "errors"
  "fmt"
  "math/rand"
  "strings"
)

type Edge struct {
  // Node.Id of the source and target for this edge
  Source int
  Target int

  // The command associated with traversing this edge
  State string

  // When there are multiple edges available to traverse this is the weight given to this edge
  // when deciding between them
  Weight float64

  // The change in facing associate with traversing this edge
  Facing int
}

// A commandGraph is a simple wrapper around a graph that allows you to specify a
// single command.  When algorithm.Dijkstra is called on this graph it will not
// follow command edges unless those edges match commandGraph.cmd
type commandGraph struct {
  *Graph
  cmd string
}

func (g *commandGraph) Adjacent(n int) ([]int, []float64) {
  var adj []int
  var times []float64
  for _, edge := range g.nodes[n].Edges {
    if edge.State != "" && edge.State != g.cmd {
      continue
    }
    adj = append(adj, edge.Target)
    times = append(times, float64(g.nodes[edge.Target].Time))
  }
  return adj, times
}

type Graph struct {
  nodes []*Node
}

func (g *Graph) NumVertex() int {
  return len(g.nodes)
}
func (g *Graph) Adjacent(n int) ([]int, []float64) {
  adj := make([]int, len(g.nodes[n].Edges))
  times := make([]float64, len(g.nodes[n].Edges))
  for i := range g.nodes[n].Edges {
    adj[i] = g.nodes[n].Edges[i].Target
    times[i] = float64(g.nodes[g.nodes[n].Edges[i].Target].Time)
  }
  return adj, times
}

func (g *Graph) IdFromName(name string) int {
  for id, n := range g.nodes {
    if name == n.Name {
      return id
    }
  }
  return -1
}
func (g *Graph) AllCmds() map[string]bool {
  cmds := make(map[string]bool)
  for _, node := range g.nodes {
    for _, edge := range node.Edges {
      if edge.State == "" {
        continue
      }
      cmds[edge.State] = true
    }
  }
  return cmds
}
func (g *Graph) StartNode() *Node {
  for _, node := range g.nodes {
    if node.Start {
      return node
    }
  }
  return nil
}

type Node struct {
  Name  string
  Id    int
  Edges []Edge
  Start bool

  // anim graph values
  Time  int64  // ms for this frame
  State string // If this is an anim node then this is the state to which it belongs
  // if this is a state node then it is ""
}

func (node *Node) IsCore() bool {
  return strings.HasPrefix(node.Name, node.State)
}

// Returns an edge from node with the name cmd.
// If multiple such edges exist one will be chosen at random using the weights
// specified in the matching edges.  If no edges match this function returns nil.
func (node *Node) FindEdge(cmd string) *Edge {
  var matches []*Edge
  weight := 0.0
  for i := range node.Edges {
    if node.Edges[i].State == cmd {
      matches = append(matches, &node.Edges[i])
      weight += node.Edges[i].Weight
    }
  }
  if len(matches) == 0 {
    return nil
  }
  hit := rand.Float64() * weight
  sum := 0.0
  for _, edge := range matches {
    sum += edge.Weight
    if hit < sum {
      return edge
    }
  }
  // TODO: should probably log a warning here, this shouldn't ever happen
  panic("WHAT")
  return matches[len(matches)-1]
}

func ProcessGraph(graph_name string, g *Graph) error {
  start_count := 0
  for _, node := range g.nodes {
    if node.Start {
      start_count++
    }
  }
  if start_count != 1 {
    return errors.New(fmt.Sprintf("Must be exactly one node marked as a start node with 'mark:start', but found %d", start_count))
  }
  return nil
}

// TODO: Make sure that recoveries don't overlap by first filling out states through
// core animations only, and only once that has been done over the whole graph, then
// go through and fill out states for recovery (i.e. post-core) animations.
func dfsState(anim *Graph, node int, state string) {
  if anim.nodes[node].State != "" {
    return
  }
  anim.nodes[node].State = state
  for _, edge := range anim.nodes[node].Edges {
    if edge.State != "" {
      continue
    }
    dfsState(anim, edge.Target, state)
  }
}
func ProcessTopology(anim, state *Graph, anim_node, state_node int, used map[int]bool) error {
  if _, ok := used[state_node]; ok {
    return nil
  }
  used[state_node] = true
  state_name := state.nodes[state_node].Name
  dfsState(anim, anim_node, state_name)
  for _, edge := range state.nodes[state_node].Edges {
    new_anim_nodes := make([]int, 0)
    for _, node := range anim.nodes {
      if node.State != state_name {
        continue
      }
      for _, anim_edge := range node.Edges {
        if anim_edge.State == edge.State {
          new_anim_nodes = append(new_anim_nodes, anim_edge.Target)
        }
      }
    }
    if len(new_anim_nodes) == 0 {
      return errors.New(fmt.Sprintf("Unable to find the command %s from animation frame %s.", anim.nodes[anim_node].Name, edge.State))
    }
    for _, new_anim_node := range new_anim_nodes {
      ProcessTopology(anim, state, new_anim_node, state.nodes[edge.Target].Id, used)
    }
  }
  return nil
}

func ProcessAnimWithState(anim, state *Graph) error {
  state_names := make(map[string]bool, len(state.nodes))
  for _, node := range state.nodes {
    state_names[node.Name] = true
  }

  if len(state_names) != len(state.nodes) {
    return errors.New(fmt.Sprintf("%d nodes, but found %d distinct state names.", len(state.nodes), len(state_names)))
  }

  for k1 := range state_names {
    for k2 := range state_names {
      if k1 == k2 {
        continue
      }
      if strings.HasPrefix(k1, k2) || strings.HasPrefix(k2, k1) {
        return errors.New(fmt.Sprintf("Cannot have a state name be a prefix of another state name: '%s' '%s'", k1, k2))
      }
    }
  }

  used_states := make(map[string]bool)
  for i := range anim.nodes {
    for state_name := range state_names {
      if strings.HasPrefix(anim.nodes[i].Name, state_name) {
        //        anim.nodes[i].State = state_name
        used_states[state_name] = true
      }
    }
  }
  if len(used_states) != len(state_names) {
    unused := make([]string, 0)
    for state_name := range state_names {
      if _, ok := used_states[state_name]; !ok {
        unused = append(unused, state_name)
      }
    }
    return errors.New(fmt.Sprintf("The following states were not accounted for in the animation: %v", unused))
  }

  {
    ac := anim.AllCmds()
    sc := state.AllCmds()
    a_not_s := make([]string, 0)
    for cmd := range ac {
      if _, ok := sc[cmd]; !ok {
        a_not_s = append(a_not_s, cmd)
      }
    }
    if len(a_not_s) > 0 {
      return errors.New(fmt.Sprintf("The following commands were found in the animation graph but not in the state graph: %v", a_not_s))
    }
    s_not_a := make([]string, 0)
    for cmd := range sc {
      if _, ok := ac[cmd]; !ok {
        s_not_a = append(s_not_a, cmd)
      }
    }
    if len(s_not_a) > 0 {
      return errors.New(fmt.Sprintf("The following commands were found in the state graph but not in the animation graph: %v", s_not_a))
    }
  }

  var err error
  if err = ProcessGraph("anim", anim); err != nil {
    return err
  }
  if err = ProcessGraph("state", state); err != nil {
    return err
  }

  if err = ProcessTopology(anim, state, anim.StartNode().Id, state.StartNode().Id, make(map[int]bool)); err != nil {
    return err
  }

  return nil
}
