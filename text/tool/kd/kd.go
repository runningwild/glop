package kd

import (
	"image"
	"math/rand"
)

type k2Tree struct {
	root *k2Node
}

type k2Node struct {
	xaxis  bool
	bounds image.Rectangle
	points []image.Point
	kids   [2]*k2Node
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

func makeKdTree(points []image.Point) *k2Tree {
	if len(points) == 0 {
		return &k2Tree{}
	}
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
	return &k2Tree{root: &k2Node{bounds: bounds, points: points}}
}
