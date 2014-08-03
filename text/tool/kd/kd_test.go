package kd_test

import (
	"github.com/orfjackal/gospec/src/gospec"
	"github.com/runningwild/glop/text/tool/kd"
	"image"
	"math"
)

func PartitionSpec(c gospec.Context) {
	c.Specify("Make sure partitioning works on cases with all unique values", func() {
		points := []image.Point{
			image.Point{3, 0},
			image.Point{4, 0},
			image.Point{2, 0},
			image.Point{7, 0},
			image.Point{8, 0},
			image.Point{6, 0},
			image.Point{5, 0},
			image.Point{1, 0},
			image.Point{9, 0},
			image.Point{0, 0},
		}
		c.Specify("Check partition with 0 on left", func() {
			c.Expect(kd.Partition(true, points, 0), gospec.Equals, 0)
		})
		c.Specify("Check partition with 1 on left", func() {
			c.Expect(kd.Partition(true, points, 1), gospec.Equals, 1)
		})
		c.Specify("Check partition with 2 on left", func() {
			c.Expect(kd.Partition(true, points, 2), gospec.Equals, 2)
		})
		c.Specify("Check partition with 3 on left", func() {
			c.Expect(kd.Partition(true, points, 3), gospec.Equals, 3)
		})
		c.Specify("Check partition with 4 on left", func() {
			c.Expect(kd.Partition(true, points, 4), gospec.Equals, 4)
		})
		c.Specify("Check partition with 5 on left", func() {
			c.Expect(kd.Partition(true, points, 5), gospec.Equals, 5)
		})
		c.Specify("Check partition with 6 on left", func() {
			c.Expect(kd.Partition(true, points, 6), gospec.Equals, 6)
		})
		c.Specify("Check partition with 7 on left", func() {
			c.Expect(kd.Partition(true, points, 7), gospec.Equals, 7)
		})
		c.Specify("Check partition with 8 on left", func() {
			c.Expect(kd.Partition(true, points, 8), gospec.Equals, 8)
		})
		c.Specify("Check partition with 9 on left", func() {
			c.Expect(kd.Partition(true, points, 9), gospec.Equals, 9)
		})
	})
	c.Specify("Simple checks to make sure that we can partition by y value", func() {
		points := []image.Point{
			image.Point{0, 3},
			image.Point{0, 4},
			image.Point{0, 2},
			image.Point{0, 7},
			image.Point{0, 8},
			image.Point{0, 6},
			image.Point{0, 5},
			image.Point{0, 1},
			image.Point{0, 9},
			image.Point{0, 0},
		}
		c.Specify("Check partition with 0 on left", func() {
			c.Expect(kd.Partition(false, points, 0), gospec.Equals, 0)
		})
		c.Specify("Check partition with 1 on left", func() {
			c.Expect(kd.Partition(false, points, 1), gospec.Equals, 1)
		})
		c.Specify("Check partition with 2 on left", func() {
			c.Expect(kd.Partition(false, points, 2), gospec.Equals, 2)
		})
		c.Specify("Check partition with 3 on left", func() {
			c.Expect(kd.Partition(false, points, 3), gospec.Equals, 3)
		})
		c.Specify("Check partition with 4 on left", func() {
			c.Expect(kd.Partition(false, points, 4), gospec.Equals, 4)
		})
		c.Specify("Check partition with 5 on left", func() {
			c.Expect(kd.Partition(false, points, 5), gospec.Equals, 5)
		})
		c.Specify("Check partition with 6 on left", func() {
			c.Expect(kd.Partition(false, points, 6), gospec.Equals, 6)
		})
		c.Specify("Check partition with 7 on left", func() {
			c.Expect(kd.Partition(false, points, 7), gospec.Equals, 7)
		})
		c.Specify("Check partition with 8 on left", func() {
			c.Expect(kd.Partition(false, points, 8), gospec.Equals, 8)
		})
		c.Specify("Check partition with 9 on left", func() {
			c.Expect(kd.Partition(false, points, 9), gospec.Equals, 9)
		})
	})
	c.Specify("Make sure partitioning works on cases with many duplicate values", func() {
		// 3 0s, 3 1s, and 3 2s
		points := []image.Point{
			image.Point{0, 0},
			image.Point{0, 0},
			image.Point{0, 0},
			image.Point{0, 0},
			image.Point{0, 0},
			image.Point{1, 0},
			image.Point{1, 0},
			image.Point{1, 0},
			image.Point{1, 0},
			image.Point{1, 0},
			image.Point{2, 0},
			image.Point{2, 0},
			image.Point{2, 0},
			image.Point{2, 0},
			image.Point{2, 0},
		}
		c.Specify("Check partition with 0 on left", func() {
			c.Expect(kd.Partition(true, points, 0), gospec.Equals, 0)
		})
		c.Specify("Check partition with 1 on left", func() {
			c.Expect(kd.Partition(true, points, 1), gospec.Equals, 5)
		})
		c.Specify("Check partition with 2 on left", func() {
			c.Expect(kd.Partition(true, points, 2), gospec.Equals, 5)
		})
		c.Specify("Check partition with 3 on left", func() {
			c.Expect(kd.Partition(true, points, 3), gospec.Equals, 5)
		})
		c.Specify("Check partition with 4 on left", func() {
			c.Expect(kd.Partition(true, points, 4), gospec.Equals, 5)
		})
		c.Specify("Check partition with 5 on left", func() {
			c.Expect(kd.Partition(true, points, 5), gospec.Equals, 5)
		})
		c.Specify("Check partition with 6 on left", func() {
			c.Expect(kd.Partition(true, points, 6), gospec.Equals, 10)
		})
		c.Specify("Check partition with 7 on left", func() {
			c.Expect(kd.Partition(true, points, 7), gospec.Equals, 10)
		})
		c.Specify("Check partition with 8 on left", func() {
			c.Expect(kd.Partition(true, points, 8), gospec.Equals, 10)
		})
		c.Specify("Check partition with 9 on left", func() {
			c.Expect(kd.Partition(true, points, 9), gospec.Equals, 10)
		})
		c.Specify("Check partition with 10 on left", func() {
			c.Expect(kd.Partition(true, points, 10), gospec.Equals, 10)
		})
		c.Specify("Check partition with 11 on left", func() {
			c.Expect(kd.Partition(true, points, 11), gospec.Equals, 15)
		})
		c.Specify("Check partition with 12 on left", func() {
			c.Expect(kd.Partition(true, points, 12), gospec.Equals, 15)
		})
		c.Specify("Check partition with 13 on left", func() {
			c.Expect(kd.Partition(true, points, 13), gospec.Equals, 15)
		})
		c.Specify("Check partition with 14 on left", func() {
			c.Expect(kd.Partition(true, points, 14), gospec.Equals, 15)
		})
		c.Specify("Check partition with 15 on left", func() {
			c.Expect(kd.Partition(true, points, 15), gospec.Equals, 15)
		})
	})
	c.Specify("Make sure large cases work", func() {
		var points []image.Point
		N := 1234
		for i := 0; i < 100000; i++ {
			N += 54577
			points = append(points, image.Point{(i + N) % 100000, 0})
		}
		c.Specify("Check partition with 50000 on left", func() {
			partition := kd.Partition(true, points, 50000)
			maxLeft := points[0].X
			minRight := points[len(points)-1].X
			for i := 0; i < partition; i++ {
				if points[i].X > maxLeft {
					maxLeft = points[i].X
				}
			}
			for i := partition; i < len(points); i++ {
				if points[i].X < minRight {
					minRight = points[i].X
				}
			}
			c.Expect(maxLeft < minRight, gospec.Equals, true)
			diff := 50000 - partition
			if diff < 0 {
				diff = -diff
			}
			c.Expect(diff < 10, gospec.Equals, true)
		})
	})
}

func TreeSpec(c gospec.Context) {
	c.Specify("Trees can count points on a disk.", func() {
		// Assemble a set of points that contain every unique integer coordinate point in (0,0) to (99,99)
		var points []image.Point
		for x := 0; x < 100; x++ {
			for y := 0; y < 100; y++ {
				points = append(points, image.Point{x, y})
			}
		}
		tree := kd.MakeKdTree(points)
		c.Expect(tree.NumPointsOnDisk(-100.1, 50, 100), gospec.Equals, 0)
		c.Expect(tree.NumPointsOnDisk(50, 50, 100), gospec.Equals, 10000)
		c.Expect(tree.NumPointsOnDisk(5, 5, 0.5), gospec.Equals, 1)
		c.Expect(tree.NumPointsOnDisk(5, 5, 1.1), gospec.Equals, 5)
	})
	c.Specify("Trees can tell distance to closest point.", func() {
		// Assemble a set of points that contain every unique integer coordinate point in (0,0) to (99,99)
		var points []image.Point
		for x := 0; x < 100; x++ {
			for y := 0; y < 100; y++ {
				points = append(points, image.Point{x, y})
			}
		}
		tree := kd.MakeKdTree(points)

		c.Expect(tree.DistToClosestPoint(-100, -100), gospec.IsWithin(1e-9), math.Sqrt(100*100+100*100))
		c.Expect(tree.DistToClosestPoint(0, 0), gospec.IsWithin(1e-9), 0.0)
		c.Expect(tree.DistToClosestPoint(0.1, 0.0), gospec.IsWithin(1e-9), 0.1)
	})
}
