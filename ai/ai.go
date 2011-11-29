package ai

import (
  "yed"
  "polish"
  "rand"
)

type Error struct {
  ErrorString string
}
func (e *Error) Error() string {
  return e.ErrorString
}
var TermError error = &Error{ "Evaluation was terminated early." }
var StartError error = &Error{ "No start node was found." }

type AiGraph struct {
  Graph   *yed.Graph
  Context *polish.Context

  // If a signal is sent along this channel it will terminate evaluation
  term chan bool
}

func NewGraph() *AiGraph {
  return &AiGraph{
    term: make(chan bool, 1),
  }
}

func (aig *AiGraph) Term() chan<- bool {
  return aig.term
}

func (aig *AiGraph) subEval(node *yed.Node) error {
  select {
    case <-aig.term:
    return TermError

    default:
  }
  res, err := aig.Context.Eval(node.Label)
  if err != nil {
    return err
  }
  var red,green,black []*yed.Edge
  for _,edge := range node.Outputs {
    if edge.R > 200 && edge.G < 100 && edge.B < 100 {
      red = append(red, edge)
    } else if edge.G > 200 && edge.R < 100 && edge.B < 100 {
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
  return aig.subEval(aig.Graph.Nodes[follow.Dst])
}

func (aig *AiGraph) Eval() error {
  for _,node := range aig.Graph.Nodes {
    if node.Label == "start" {
      err := aig.subEval(aig.Graph.Nodes[node.Outputs[0].Dst])
      if err != nil {
        return err
      }
    }
  }
  return StartError
}
