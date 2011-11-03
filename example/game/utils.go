package game

func maxNormi(x, y, x2, y2 int) int {
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
