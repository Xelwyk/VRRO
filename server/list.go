package main

import (
	"fmt"
	"log"
	"math"
	"sync"
	"time"
)

type Vertex struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
	//Confidence float64 `json:"c"`
}

type Point struct {
	Camera      Vertex `json:"camera"`
	Coordinates Vertex `json:"point"`
}

type ListNode struct {
	point Point
	next  *ListNode
}

type LinkedPointList struct {
	head      *ListNode
	tail      *ListNode
	size      uint
	sizeLimit uint
	buffer    chan Point
	Mu        sync.Mutex
}

func (list *LinkedPointList) add(point Point) {
	list.Mu.Lock()
	defer list.Mu.Unlock()
	node := &ListNode{point, nil}
	if list.head == nil {
		list.head = node
		list.tail = node
	} else {
		list.tail.next = node
		list.tail = node
	}
	list.size++

	//if list.size > list.sizeLimit {
	//	list.head = list.head.next
	//	list.size--
	//}
}

func (list *LinkedPointList) popHead() Point {
	if list.head == nil {
		return Point{}
	}
	point := list.head.point
	list.head = list.head.next
	list.size--
	return point
}

func (list *LinkedPointList) addFromChannel() {
	fmt.Println("Starting to add points from channel")
	for {
		for point := range list.buffer {
			if list.exorcise(point) {
				list.add(point)
				fmt.Println("Added point to list:", point)
			}
		}
	}
}

func (list *LinkedPointList) exorcise(point Point) bool {
	list.Mu.Lock()
	defer list.Mu.Unlock()
	if list.size > 0 {
		var wg sync.WaitGroup
		for node := list.head; node != nil; node = node.next {

			if samePoint(&point, &node.point) {
				return false
			}
			return true
			if !isInsideBrackets(&point, &node.point) {
				continue
			}
			x1 := point.Camera.X
			x2 := point.Coordinates.X
			y1 := point.Camera.Y
			y2 := point.Coordinates.Y
			z1 := point.Camera.Z
			z2 := point.Coordinates.Z
			x0 := node.point.Coordinates.X
			y0 := node.point.Coordinates.Y
			z0 := node.point.Coordinates.Z
			wg.Add(4)
			var r float64
			go func() {
				r = math.Sqrt(math.Pow(x1-x0, 2) + math.Pow(y1-y0, 2) + math.Pow(z1-z0, 2))
				wg.Done()
			}()
			var a1 float64
			go func() {
				a1 = (x0 * y1) + (x1 * y2) + (x2 * y0) - ((y0 * x1) + (y1 * x2) + (y2 * x0))
				wg.Done()
			}()
			var a2 float64
			go func() {
				a2 = (y0 * z1) + (y1 * z2) + (y2 * z0) - ((z0 * y1) + (z1 * y2) + (z2 * y0))
				wg.Done()
			}()
			var a3 float64
			go func() {
				a3 = (x0 * z1) + (x1 * z2) + (x2 * z0) - ((z0 * y1) + (z1 * x2) + (z2 * x0))
				wg.Done()
			}()
			wg.Wait()
			A := 0.5 * math.Sqrt(math.Pow(a1, 2)+math.Pow(a2, 2)+math.Pow(a3, 2))
			h := (2 * A) / r
			if h < 0.05 {
				if node.next == nil {
					list.tail = node
				} else {
					node.point = node.next.point
				}
				list.size--
			}
		}
	}
	return true
}

func isInsideBrackets(new, old *Point) bool {
	if old.Coordinates.X < min(new.Camera.X, new.Coordinates.X) {
		return false
	}
	if old.Coordinates.X > max(new.Camera.X, new.Coordinates.X) {
		return false
	}
	if old.Coordinates.Y < min(new.Camera.Y, new.Coordinates.Y) {
		return false
	}
	if old.Coordinates.Y > max(new.Camera.Y, new.Coordinates.Y) {
		return false
	}
	if old.Coordinates.Z < min(new.Camera.Z, new.Coordinates.Z) {
		return false
	}
	if old.Coordinates.Z > max(new.Camera.Z, new.Coordinates.Z) {
		return false
	}
	return true
}

func samePoint(new, old *Point) bool {
	var xdiff float64
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		xdiff = math.Abs(new.Coordinates.X - old.Coordinates.X)
		wg.Done()
	}()
	var ydiff float64
	go func() {
		ydiff = math.Abs(new.Coordinates.Y - old.Coordinates.Y)
		wg.Done()
	}()
	var zdiff float64
	go func() {
		zdiff = math.Abs(new.Coordinates.Z - old.Coordinates.Z)
		wg.Done()
	}()
	wg.Wait()
	diff := 0.05
	if xdiff < diff && ydiff < diff && zdiff < diff {
		return true
	}
	return false
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}
