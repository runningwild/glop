package sprite

import (
  "fmt"
  "image"
  _ "image/png"
  "os"
  "path/filepath"
  "sort"
  "strconv"
  "strings"
  "github.com/runningwild/glop/util/algorithm"
  "github.com/runningwild/yedparse"
)

type sharedSprite struct {
  path string

  anim,state  *yed.Graph
  anim_start  *yed.Node
  state_start *yed.Node

  node_data map[*yed.Node]nodeData
  edge_data map[*yed.Edge]edgeData

  connector *sheet
  facings []*sheet

  manager *Manager
}

func loadSharedSprite(path string) (*sharedSprite, error) {
  state,err := yed.ParseFromFile(filepath.Join(path, "state.xgml"))
  if err != nil { return nil, err }

  err = verifyStateGraph(&state.Graph)
  if err != nil { return nil, err }

  anim,err := yed.ParseFromFile(filepath.Join(path, "anim.xgml"))
  if err != nil { return nil, err }

  err = verifyAnimGraph(&anim.Graph)
  if err != nil { return nil, err }

  // TODO: Verify both graphs at the same time - they both need to respond to
  // the same commands in the same way.

  num_facings, filenames, err := verifyDirectoryStructure(path, &anim.Graph)
  if err != nil { return nil, err }

  // If we've made it this far then the sprite is probably well formed so we
  // can start putting all of the data together
  var ss sharedSprite
  ss.path = path
  ss.anim = &anim.Graph
  ss.state = &state.Graph

  // Read through all of the files and figure out how much space we'll need
  // to arrange them all into one sprite sheet
  width := 0
  height := 0
  for facing := 0; facing < num_facings; facing++ {
    for _,filename := range filenames {
      file,err := os.Open(filepath.Join(path, fmt.Sprintf("%d", facing), filename))
      // if a file isn't there that's ok
      if err != nil { continue }

      config,_,err := image.DecodeConfig(file)
      file.Close()
      // if a file can't be read that is *not* ok
      if err != nil { return nil, err }

      if config.Height > height {
        height = config.Height
      }
      width += config.Width
    }
  }

  // Connectors are all frames that can be reached within a certain number of
  // milliseconds of any change in facing
  conn := figureConnectors(&anim.Graph, 150)

  // Arrange them all into one sprite sheet
  var fids []frameId
  for _,con := range conn {
    for facing := 0; facing < num_facings; facing++ {
      fids = append(fids, frameId{ facing: facing, node: con.Id() })
    }
  }
  sort.Sort(frameIdArray(fids))
  ss.connector,err = makeSheet(path, &anim.Graph, fids)
  if err != nil { return nil, err }

  // Now we make a sheet for each facing, but don't include any of the frames
  // that are in the connctor sheet
  used := make(map[*yed.Node]bool)
  for _,con := range conn {
    used[con] = true
  }
  for facing := 0; facing < num_facings; facing++ {
    var facing_fids []frameId
    for i := 0; i < anim.Graph.NumNodes(); i++ {
      node := anim.Graph.Node(i)
      if !used[node] {
        facing_fids = append(facing_fids, frameId{ facing: facing, node: node.Id() })
      }
    }
    sort.Sort(frameIdArray(facing_fids))
    sh,err := makeSheet(path, &anim.Graph, facing_fids)
    if err != nil { return nil, err }
    ss.facings = append(ss.facings, sh)
  }

  ss.connector.Load()
  ss.anim_start = getStartNode(ss.anim)
  ss.state_start = getStartNode(ss.state)

  ss.process()

  return &ss, nil
}

// Given the anim graph for a sprite, determines the frames that must always
// be loaded such that the remaining facings can be loaded only when the
// sprite facing changes, so long as the facings sprite sheet can be loaded
// in under limit milliseconds.  A higher value for limit will require more
// texture memory, but will reduce the chance that there will be any
// stuttering in the animation because a spritesheet couldn't be loaded in
// time.
func figureConnectors(anim *yed.Graph, limit int) []*yed.Node {
  var facing_edges []int
  for i := 0; i < anim.NumEdges(); i++ {
    edge := anim.Edge(i)
    if edge.Tag("facing") != "" {
      facing_edges = append(facing_edges, edge.Dst().Id())
    }
  }

  reachable := algorithm.ReachableWithinLimit(&animAlgoGraph{anim}, facing_edges, float64(limit))
  var ret []*yed.Node
  for _,reach := range reachable {
    ret = append(ret, anim.Node(reach))
  }
  return ret
}

func (ss *sharedSprite) markNodesWithState(node *yed.Node, state string) {
  used := make(map[*yed.Node]bool)
  unused := make(map[*yed.Node]bool)
  unused[node] = true
  for len(unused) > 0 {
    for node = range unused {
      break
    }
    delete(unused, node)
    used[node] = true
    data := ss.node_data[node]
    data.state = state
    ss.node_data[node] = data
    for i := 0; i < node.NumGroupOutputs(); i++ {
      edge := node.GroupOutput(i)
      if edge.Line(0) != "" {
        continue
      }
      n := edge.Dst()
      if ss.node_data[n].state != "" {
        continue
      }
      if !used[n] {
        unused[n] = true
      }
    }
  }
}

func (ss *sharedSprite) findCmdFromAnimNode(node *yed.Node, cmd string) *yed.Node {
  used := make(map[*yed.Node]bool)
  unused := make(map[*yed.Node]bool)
  unused[node] = true
  for len(unused) > 0 {
    for node = range unused {
      break
    }
    delete(unused, node)
    used[node] = true
    for i := 0; i < node.NumGroupOutputs(); i++ {
      edge := node.GroupOutput(i)
      if edge.NumLines() > 0 && edge.Line(0) == cmd {
        return edge.Dst()
      }
      n := edge.Dst()
      if !used[n] {
        unused[n] = true
      }
    }
  }
  return nil
}

func (ss *sharedSprite) markAnimFramesWithState(anim, state *yed.Node) {
  if ss.node_data[anim].state != "" {
    return
  }
  ss.markNodesWithState(anim, state.Line(0))
  for i := 0; i < state.NumGroupOutputs(); i++ {
    edge := state.GroupOutput(i)
    cmd := edge.Line(0)
    if cmd == "" {
      continue
    }
    next_anim := ss.findCmdFromAnimNode(anim, cmd)
    ss.markAnimFramesWithState(next_anim, edge.Dst())
  }
}

func (ss *sharedSprite) process() {
  ss.node_data = make(map[*yed.Node]nodeData)
  for i := 0; i < ss.anim.NumNodes(); i++ {
    node := ss.anim.Node(i)
    data := nodeData{ time: defaultFrameTime, sync_tag: node.Tag("sync") }
    t,err := strconv.ParseInt(node.Tag("time"), 10, 32)
    if err == nil {
      data.time = t
    }
    ss.node_data[node] = data
  }

  ss.edge_data = make(map[*yed.Edge]edgeData)
  proc_graph := func(graph *yed.Graph) {
    for i := 0; i < graph.NumEdges(); i++ {
      edge := graph.Edge(i)
      data := edgeData{ weight: 1.0 }

      f,err := strconv.ParseInt(edge.Tag("facing"), 10, 32)
      if err == nil {
        data.facing = int(f)
      }

      w,err := strconv.ParseFloat(edge.Tag("weight"), 64)
      if err == nil {
        data.weight = w
      }

      cmd := edge.Line(0)
      if !strings.Contains(cmd, ":") {
        data.cmd = cmd
      }

      ss.edge_data[edge] = data
    }
  }
  proc_graph(ss.anim)
  proc_graph(ss.state)

  ss.markAnimFramesWithState(ss.anim_start, ss.state_start)
  for i := 0; i < ss.anim.NumNodes(); i++ {
    node := ss.anim.Node(i)
    fmt.Printf("%s -> %s\n", node.Line(0), ss.node_data[node].state)
  }
}

