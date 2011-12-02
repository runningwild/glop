package sprite

import (
  "errors"
  "fmt"
  "os"
  "encoding/xml"
  "strconv"
  "strings"
  "regexp"
  "io"
)

var (
  tokenizer *regexp.Regexp
  splitter  *regexp.Regexp
)

type yedGraph struct {
  Section Section
}

type Section struct {
  Name       string      `xml:"attr"`
  Attributes []Attribute `xml:"attribute>"`
  Sections   []Section   `xml:"section>"`
}

func (s Section) Gids() []int {
  gids := make([]int, 0)
  for _, attr := range s.Attributes {
    if attr.Key == "gid" {
      gid, err := strconv.Atoi(attr.Data)
      if err != nil {
        panic(err.Error())
      }
      gids = append(gids, gid)
    }
  }
  return gids
}
func (s Section) GetAttribute(key string, do_or_die bool) string {
  for _, attr := range s.Attributes {
    if attr.Key == key {
      return attr.Data
    }
  }
  if do_or_die {
    panic(fmt.Sprintf("Unable to find attribute with key == '%s'", key))
  }
  return ""
}

type Attribute struct {
  Key  string `xml:"attr"`
  Type string `xml:"attr"`
  Data string `xml:"chardata"`
}

func init() {
  tokenizer = regexp.MustCompile("[^( |\t|\r|\n)]+")
  splitter = regexp.MustCompile("([^ \t\n\r\f]+)[ \t\r\f\n]*:[ \t\r\f\n]*([^ \t\n\r\f]+)")

}

func ParseXMLNode(section Section) Node {
  var node Node
  var err error
  id := section.GetAttribute("id", true)
  node.Id, err = strconv.Atoi(id)
  if err != nil {
    panic(err.Error())
  }
  node.Name = section.GetAttribute("label", true)
  node.Edges = make([]Edge, 0)
  node.Time = 100 // TODO: Check with Austin on this one

  lines := tokenizer.FindAllString(node.Name, -1)
  if len(lines) < 0 {
    panic("There is a node without a name")
  }
  node.Name = lines[0]
  for i := 1; i < len(lines); i++ {
    a := splitter.FindStringSubmatch(lines[i])
    if len(a) != 3 {
      panic(fmt.Sprintf("Error parsing line %d of node with name '%s'", i, node.Name))
    }
    switch strings.ToLower(a[1]) {
    case "time":
      t, err := strconv.Atoi(a[2])
      if err != nil {
        panic(fmt.Sprintf("Error reading time for node with name '%s'", node.Name))
      }
      node.Time = int64(t)
    case "mark":
      switch strings.ToLower(a[2]) {
      case "start":
        node.Start = true
      default:
        panic(fmt.Sprintf("Unknown mark: '%s'", a[2]))
      }
    default:
      panic(fmt.Sprintf("Unknown node key: '%s'", a[1]))
    }
  }
  return node
}

func ParseXMLEdge(section Section) (Edge, error) {
  edge := Edge{
    Weight: 1.0,
  }
  var err error

  source := section.GetAttribute("source", true)
  edge.Source, err = strconv.Atoi(source)
  if err != nil {
    return edge, err
  }

  target := section.GetAttribute("target", true)
  edge.Target, err = strconv.Atoi(target)
  if err != nil {
    return edge, err
  }

  label := section.GetAttribute("label", false)
  // It's ok for an edge to be unlabled 
  if err != nil {
    return edge, nil
  }

  lines := tokenizer.FindAllString(label, -1)
  if len(lines) == 0 {
    return edge, nil
  }

  for i := 0; i < len(lines); i++ {
    a := splitter.FindStringSubmatch(lines[i])
    if len(a) == 0 && i == 0 {
      edge.State = lines[0]
      continue
    }
    if len(a) != 3 {
      return edge, errors.New(fmt.Sprintf("Error parsing line %d of anim edge with label '%s'", i, label))
    }
    switch strings.ToLower(a[1]) {
    case "weight":
      weight, err := strconv.Atof64(a[2])
      if err != nil {
        return edge, errors.New(fmt.Sprintf("Error reading weight for anim edge with name '%s'", label))
      }
      edge.Weight = weight

    case "facing":
      facing, err := strconv.Atoi(a[2])
      if err != nil {
        return edge, errors.New(fmt.Sprintf("Error reading facing for anim edge with name '%s'", label))
      }
      edge.Facing = facing

    default:
      return edge, errors.New(fmt.Sprintf("Unknown animation edge key: '%s'", a[1]))
    }
  }

  return edge, nil
}

func ParseXMLGraph(section Section) *Graph {
  if section.Name != "graph" {
    panic(fmt.Sprintf("Expected section.Name == 'graph', found '%s'", section.Name))
  }
  var g Graph
  nodes := make(map[int]*Node)

  // map from gid to a list of ids that group contains
  groups := make(map[int][]int)
  for _, sub_section := range section.Sections {
    if sub_section.Name == "node" {
      for _, gid := range sub_section.Gids() {
        sid := sub_section.GetAttribute("id", true)
        id, err := strconv.Atoi(sid)
        if err != nil {
          panic(err.Error())
        }
        groups[gid] = append(groups[gid], id)
      }
    }
  }

  // Each section is either a node or an edge, call the appropriate functions to
  // populate the graph with their data
  var edges []Edge
  for _, sub_section := range section.Sections {
    switch sub_section.Name {
    case "node":
      if sub_section.GetAttribute("isGroup", false) != "" {
        continue
      }
      node := ParseXMLNode(sub_section)
      nodes[node.Id] = &node
    case "edge":
      edge, err := ParseXMLEdge(sub_section)
      if err != nil {
        panic(err.Error())
      }
      edges = append(edges, edge)
    default:
      panic(fmt.Sprintf("Expected section.Name == 'node' | 'edge', found '%s'", sub_section.Name))
    }
  }

  // Connect things up, taking into account that we can have edges come from groups
  for _, edge := range edges {
    sids, ok := groups[edge.Source]
    if !ok {
      sids = []int{edge.Source}
    }
    _, ok = groups[edge.Target]
    if ok {
      panic("Cannot have an edge point at a group")
    }
    for _, sid := range sids {
      var e2 Edge = edge
      e2.Source = sid
      source := nodes[sid]
      source.Edges = append(source.Edges, e2)
    }
  }

  // Normalize the graph, in case the nodes ids that we were given were not
  // consecutive intetegers from 0..n
  m := make(map[int]int)
  c := 0
  for k := range nodes {
    m[k] = c
    c++
  }
  g.nodes = make([]*Node, len(m))
  for k, n := range nodes {
    g.nodes[m[k]] = n
    n.Id = m[k]
  }
  for i := range edges {
    edges[i].Source = m[edges[i].Source]
    edges[i].Target = m[edges[i].Target]
  }
  for i := range g.nodes {
    for j := range g.nodes[i].Edges {
      g.nodes[i].Edges[j].Source = m[g.nodes[i].Edges[j].Source]
      g.nodes[i].Edges[j].Target = m[g.nodes[i].Edges[j].Target]
    }
  }
  return &g
}

func GraphFromYEd(section Section) (graph *Graph, err error) {
  defer func() {
    if r, ok := recover().(string); ok {
      graph = nil
      err = errors.New(r)
    }
  }()
  if section.Name != "xgml" {
    panic(fmt.Sprintf("Unable to parse Section, Section.Name is '%s' instead of xgml'", section.Name))
  }
  if len(section.Sections) != 1 {
    panic(fmt.Sprintf("Expected exactly 1 sections under 'xgml', found %d.", len(section.Sections)))
  }
  section = section.Sections[0]
  if section.Name != "graph" {
    panic(fmt.Sprintf("Expected a section named 'graph', found one named '%s'", section.Name))
  }
  graph = ParseXMLGraph(section)
  return
}

func ReadGraph(file io.Reader) (*Graph, error) {
  var yed Section
  err := xml.Unmarshal(file, &yed)
  if err != nil {
    return nil, err
  }
  return GraphFromYEd(yed)
}

func LoadGraph(filename string) (*Graph, error) {
  f, err := os.Open(filename)
  if err != nil {
    return nil, err
  }
  defer f.Close()
  return ReadGraph(f)
}
