package kd

import (
	"image"
	"math/rand"
)

const maxPointsPerNode = 10

type Tree struct {
	*k2Node
}

type k2Node struct {
	xaxis  bool
	count  int
	bounds image.Rectangle
	points []image.Point
	kids   [2]*k2Node
}

func distSqFromPointToPoint(x0, y0, x1, y1 float64) float64 {
	return (x0-x1)*(x0-x1) + (y0-y1)*(y0-y1)
}

// distFromPointToRect returns the manhattan distance from p to r.
func distFromPointToRect(p image.Point, r image.Rectangle) int {
	inX := p.X >= r.Min.X && p.X <= r.Max.X
	inY := p.Y >= r.Min.Y && p.Y <= r.Max.Y
	if inX && inY {
		return 0
	}
	var dX, dY int
	if inX {
		if p.Y < r.Min.Y {
			dY = r.Min.Y - p.Y
		} else {
			dY = p.Y - r.Max.Y
		}
	}
	if inY {
		if p.X < r.Min.X {
			dX = r.Min.X - p.X
		} else {
			dX = p.X - r.Max.X
		}
	}
	return dX + dY
}

func (k2 *k2Node) NumPointsOnDisk(x, y, radius float64) int {
	rsq := radius * radius

	if distFromPointToRect(image.Point{int(x), int(y)}, k2.bounds)+2 > int(radius) {
		return 0
	}

	// If we have no kids, then count each point that we have that's on the disk and return that.
	if k2.points != nil {
		count := 0
		for _, point := range k2.points {
			if distSqFromPointToPoint(float64(point.X), float64(point.Y), x, y) < rsq {
				count++
			}
		}
		return count
	}

	// Check if the disk completely contains this node, if it does just return the total count.
	if distSqFromPointToPoint(float64(k2.bounds.Min.X), float64(k2.bounds.Min.Y), x, y) < rsq &&
		distSqFromPointToPoint(float64(k2.bounds.Min.X), float64(k2.bounds.Max.Y), x, y) < rsq &&
		distSqFromPointToPoint(float64(k2.bounds.Max.X), float64(k2.bounds.Min.Y), x, y) < rsq &&
		distSqFromPointToPoint(float64(k2.bounds.Max.X), float64(k2.bounds.Max.Y), x, y) < rsq {
		return k2.count
	}

	// If we've gotten here then we just let our kids do the work.
	return k2.kids[0].NumPointsOnDisk(x, y, radius) + k2.kids[1].NumPointsOnDisk(x, y, radius)
}

func (k2 *k2Node) DistToClosestPoint(x, y float64) float64 {
	// Starting with dist == 1, do repeated doubling until we find at least one point, then binary
	// search to find the closest point.
	startingDist := 100.0
	dist := startingDist
	for k2.NumPointsOnDisk(x, y, dist) == 0 {
		dist *= 2
	}
	var min, max float64
	if dist == startingDist {
		// special case if there was already one point within range 1.0.
		min = 0.0
		max = startingDist
	} else {
		min = dist / 2
		max = dist
	}

	for max-min > 1e-2 {
		mid := (max + min) / 2
		if k2.NumPointsOnDisk(x, y, mid) == 0 {
			min = mid
		} else {
			max = mid
		}
	}
	return (max + min) / 2
}

func Partition(x bool, points []image.Point, leftCount int) int {
	if leftCount == 0 {
		return 0
	}
	if len(points) == 0 {
		return -1
	}
	divPoint := points[rand.Intn(len(points))]
	var div int
	if x {
		div = divPoint.X
	} else {
		div = divPoint.Y
	}
	leftIndex := 0
	rightIndex := len(points)
	onDiv := 0
	pos := 0
	for pos < rightIndex {
		if (x && points[pos].X <= div) || (!x && points[pos].Y <= div) {
			if (x && points[pos].X == div) || (!x && points[pos].Y == div) {
				onDiv++
			}
			points[pos], points[leftIndex] = points[leftIndex], points[pos]
			leftIndex++
			pos++
		} else {
			rightIndex--
			points[pos], points[rightIndex] = points[rightIndex], points[pos]
		}
	}
	if leftIndex == leftCount {
		return leftIndex
	}
	if leftIndex >= leftCount && onDiv >= 2 && leftIndex-(onDiv-1) <= leftCount {
		return leftIndex
	}
	if leftIndex < leftCount {
		return leftIndex + Partition(x, points[leftIndex:], leftCount-leftIndex)
	} else {
		return Partition(x, points[0:leftIndex], leftCount)
	}
}

func boundingBoxFromPoints(points []image.Point) image.Rectangle {
	var bounds image.Rectangle
	bounds.Min = points[0]
	bounds.Max = points[0]
	for _, point := range points {
		if point.X < bounds.Min.X {
			bounds.Min.X = point.X
		}
		if point.Y < bounds.Min.Y {
			bounds.Min.Y = point.Y
		}
		if point.X > bounds.Max.X {
			bounds.Max.X = point.X
		}
		if point.Y > bounds.Max.Y {
			bounds.Max.Y = point.Y
		}
	}
	return bounds
}

func makek2Node(points []image.Point, x bool) *k2Node {
	node := k2Node{
		bounds: boundingBoxFromPoints(points),
		count:  len(points),
		xaxis:  x,
	}
	if len(points) <= maxPointsPerNode {
		node.points = points
		return &node
	}
	leftIndex := Partition(x, points, len(points)/2)
	if leftIndex == 0 || leftIndex == len(points) {
		node.points = points
		return &node
	}
	node.kids[0] = makek2Node(points[0:leftIndex], !x)
	node.kids[1] = makek2Node(points[leftIndex:], !x)
	return &node
}

func MakeKdTree(points []image.Point) *Tree {
	if len(points) == 0 {
		return &Tree{}
	}
	return &Tree{k2Node: makek2Node(points, true)}
}
