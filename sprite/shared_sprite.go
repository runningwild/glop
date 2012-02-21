package sprite

import (
  "fmt"
  "image"
  _ "image/png"
  "os"
  "path/filepath"
  "strconv"
  "strings"
	"github.com/runningwild/yedparse"
)

type sharedSprite struct {
  anim,state  *yed.Graph
  anim_start  *yed.Node
  state_start *yed.Node

  node_data map[*yed.Node]nodeData
  edge_data map[*yed.Edge]edgeData

  connector *sheet
  facings []*sheet
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

func (ss *sharedSprite) process() {
  ss.node_data = make(map[*yed.Node]nodeData)
  for i := 0; i < ss.anim.NumNodes(); i++ {
    node := ss.anim.Node(i)
    data := nodeData{ time: defaultFrameTime }
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
}

