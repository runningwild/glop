package ai

import (
  "github.com/runningwild/yedparse"
  "github.com/runningwild/polish"
  "math/rand"
)

type Error struct {
  ErrorString string
}
func (e *Error) Error() string {
  return e.ErrorString
}

var InterruptError error = &Error{ "Evaluation was terminated due to an interrupt." }
var TermError error = &Error{ "Evaluation was terminated early." }
var StartError error = &Error{ "No start node was found." }

type AiGraph struct {
  Graph   *yed.Graph
  Context *polish.Context

  // If a signal is sent along this channel it will terminate evaluation with
  // the error that was sent
  term chan error
}

func NewGraph() *AiGraph {
  return &AiGraph{
    term: make(chan error, 1),
  }
}

func (aig *AiGraph) Term() chan<- error {
  return aig.term
}

func (aig *AiGraph) subEval(labels *[]string, node *yed.Node) error {
  select {
    case err := <-aig.term:
    return err

    default:
  }
  *labels = append(*labels, node.Label())
  res, err := aig.Context.Eval(node.Label())
  if err != nil {
    return err
  }
  var red,green,black []*yed.Edge
  for i := 0; i < node.NumOutputs(); i++ {
    edge := node.Output(i)
    r,g,b,_ := edge.RGBA()
    if r > 200 && g < 100 && b < 100 {
      red = append(red, edge)
    } else if g > 200 && r < 100 && b < 100 {
      green = append(green, edge)
    } else {
      black = append(black, edge)
    }
  }
  if (len(red) == 0) != (len(green) == 0) {
    panic("A node cannot have red edges without green edges or vice versa.")
  }

  // A node can have green, red, and black edges.  In this case the condition
  // will be evaluated and if true either a green or black edge will be
  // traversed, and if false a red or black edge will be traversed.

  var edges []*yed.Edge
  if len(green) > 0 {
    for _,edge := range black {
      red = append(red, edge)
      green = append(green, edge)
    }
    if len(res) != 1 {
      panic("Needed to evaluate a node, but it didn't leave exactly one value after evalutation.")
    }
    if res[0].Bool() {
      edges = green
    } else {
      edges = red
    }
  } else {
    edges = black
  }

  if len(edges) == 0 {
    return nil
  }
  follow := edges[rand.Intn(len(edges))]
  return aig.subEval(labels, follow.Dst())
}

func (aig *AiGraph) Eval() ([]string, error) {
  var labels []string
  for i := 0; i < aig.Graph.NumNodes(); i++ {
    node := aig.Graph.Node(i)
    if node.Label() == "start" {
      err := aig.subEval(&labels, node.Output(0).Dst())
      return labels, err
    }
  }
  return labels, StartError
}
