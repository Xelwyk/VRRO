package main

import (
	"log"
	"math"
	"sync"
	"time"

	"github.com/kyroy/kdtree"
)

const threshold float64 = 0.018

type Vertex struct {
	X, Y, Z float64
}

type IncomingVertex struct {
	Camera      Vertex
	Coordinates Vertex
}

func (v Vertex) Dimensions() int {
	return 3
}

func (v Vertex) Dimension(i int) float64 {
	return []float64{v.X, v.Y, v.Z}[i]
}

type PointCloud struct {
	buffer chan IncomingVertex
	tree   *kdtree.KDTree
	mutex  sync.Mutex
}

func (pc *PointCloud) init() {
	pc.tree = kdtree.New([]kdtree.Point{Vertex{0, 0, 0}})
}

func (pc *PointCloud) insert(point Vertex) {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	nearestPoint := pc.tree.KNN(point, 1)
	if distance(point, nearestPoint[0].(Vertex)) < threshold {
		return
	}
	pc.tree.Insert(point)
}

func (pc *PointCloud) adderLoop() {
	for {
		for point := range pc.buffer {
			pc.insert(point.Coordinates)
		}
	}
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func distance(p1, p2 Vertex) float64 {
	return math.Sqrt(math.Pow(p2.X-p1.X, 2) + math.Pow(p2.Y-p1.Y, 2) + math.Pow(p2.Z-p1.Z, 2))
}
