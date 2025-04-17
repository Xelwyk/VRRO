package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/julienschmidt/httprouter"
)

func main() {
	fmt.Println("Starting server")
	const bufferSize = 500000000
	points := LinkedPointList{nil, nil, 0, bufferSize, make(chan Point, bufferSize), sync.Mutex{}}
	camera := Vertex{}
	go points.addFromChannel()
	go intakeCloudPoints(&points, &camera)
	serveCloudPoints(&points, &camera)
}

func serveCloudPoints(points *LinkedPointList, camera *Vertex) {
	fmt.Println("Starting HTTP server")
	router := httprouter.New()
	router.GET("/cloudpoints", handleCPServing(points))
	router.GET("/camera", handleCmanServing(camera))
	log.Fatal(http.ListenAndServe(":8081", router))
}

func handleCPServing(points *LinkedPointList) httprouter.Handle {
	fmt.Println("Setting up handler")
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		//fmt.Println("Handling request")
		points.Mu.Lock()
		defer points.Mu.Unlock()

		var readyPoints []Vertex

		for node := points.head; node != nil; node = node.next {
			readyPoints = append(readyPoints, node.point.Coordinates)
		}

		jsonResponse, err := json.Marshal(readyPoints)
		if err != nil {
			fmt.Println("Error marshalling point to JSON:", err)
		}
		fmt.Fprintf(w, "%s", jsonResponse)
		//fmt.Println("Finished handling request")
	}
}

func handleCmanServing(camera *Vertex) httprouter.Handle {
	fmt.Println("Setting up camera handler")
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		//fmt.Println("Handling camera request")
		cameraJSON, err := json.Marshal(camera)
		if err != nil {
			fmt.Println("Error marshalling camera to JSON:", err)
		}
		fmt.Fprintf(w, "%s", cameraJSON)
		//fmt.Println("Finished handling camera request")
	}
}

func intakeCloudPoints(points *LinkedPointList, camera *Vertex) {
	fmt.Println("Starting UDP server")
	// Resolve the address to listen on
	addr, err := net.ResolveUDPAddr("udp", ":8080")
	if err != nil {
		fmt.Println("Error resolving address:", err)
		os.Exit(1)
	}
	fmt.Println("address:", addr.IP.String())

	// Create a UDP connection
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error creating UDP connection:", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("Listening on", addr)

	// Buffer to store received data
	buffer := make([]byte, 5*1024*1024) // 5MB buffer
	var stringedBuffer string

	for {
		// Read data from the connection
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading from UDP connection:", err)
			stringedBuffer = ""
			continue
		}
		stringedBuffer += string(buffer[:n])
		if stringedBuffer[len(stringedBuffer)-2] != ']' {
			continue
		}

		// Unmarshal the received data into a slice of Point structs
		var newData map[string]any
		//fmt.Println("Received data:", stringedBuffer)
		err = json.Unmarshal([]byte(stringedBuffer), &newData)
		if err != nil {
			fmt.Println("Error unmarshalling JSON:", err)
			stringedBuffer = ""
			continue
		}

		camera.X = newData["camera"].(map[string]any)["x"].(float64)
		camera.Y = newData["camera"].(map[string]any)["y"].(float64)
		camera.Z = newData["camera"].(map[string]any)["z"].(float64)
		//camera.Confidence = 1.0

		for _, newPointMap := range newData["points"].([]any) {
			newPoint := Vertex{
				X: newPointMap.(map[string]any)["x"].(float64),
				Y: newPointMap.(map[string]any)["y"].(float64),
				Z: newPointMap.(map[string]any)["z"].(float64),
				//Confidence: newPointMap.(map[string]any)["c"].(float64),
			}
			points.buffer <- Point{
				Camera:      *camera,
				Coordinates: newPoint,
			}
		}
		stringedBuffer = ""
	}
}
