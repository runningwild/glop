package base

import (
  "os"
  "json"
  "io/ioutil"
)

type Terrain string

func LoadJson(path string, target interface{}) error {
  f, err := os.Open(path)
  if err != nil {
    return err
  }
  data, err := ioutil.ReadAll(f)
  if err != nil {
    return err
  }
  err = json.Unmarshal(data, target)
  if err != nil {
    return err
  }
  return nil
}

func MaxNormi(x, y, x2, y2 int) int {
  dx := x2 - x
  if dx < 0 {
    dx = -dx
  }
  dy := y2 - y
  if dy < 0 {
    dy = -dy
  }
  if dx > dy {
    return dx
  }
  return dy
}
