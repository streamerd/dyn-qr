package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
	"golang.org/x/exp/rand"
	"golang.org/x/net/websocket"
)

type QRCodeData struct {
	ID   int64  `json:"id"`
	Data string `json:"data"`
}

type Bus struct {
	L string `json:"l"`
	M int    `json:"m"`
}

type BusStop struct {
	S int   `json:"s"`
	B []Bus `json:"b"`
}

// Add a mutex for the bus queue
var (
	dataMap     = make(map[string]string)
	busQueue    = make([]Bus, 0)
	queueMutex  sync.Mutex
	currentData string
)

// Constants for bus simulation
const (
	FIXED_BUS_STOP      = 4242
	MAX_BUSES           = 8   // Maximum number of buses in queue
	MIN_NEW_BUS_TIME    = 5   // Minimum minutes for a new bus
	MAX_NEW_BUS_TIME    = 30  // Maximum minutes for a new bus
	NEW_BUS_PROBABILITY = 0.3 // 30% chance of adding a new bus each tick
)

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("views/*")

	// Route to serve the HTML page
	r.GET("/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Dynamic QR Code Generator",
		})
	})

	// WebSocket route for publishing QR code data
	r.GET("/ws", func(c *gin.Context) {
		// Check if the request is a valid WebSocket handshake
		if strings.Contains(c.Request.Header.Get("Upgrade"), "websocket") {
			websocket.Handler(handleWebSocket).ServeHTTP(c.Writer, c.Request)
			return
		}

		// If not a valid WebSocket handshake, continue with regular HTTP handling
		c.Next()
	})

	// HTTP route for retrieving the QR code image
	r.GET("/qr/:id", func(c *gin.Context) {
		id := c.Param("id")

		queueMutex.Lock()
		data, exists := dataMap[id]
		queueMutex.Unlock()

		if !exists {
			fmt.Printf("Data not found for ID: %s\n", id)
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Header("Content-Type", "image/png")

		png, err := qrcode.Encode(data, qrcode.Low, 256)
		if err != nil {
			fmt.Printf("Error encoding QR: %v\n", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Data(http.StatusOK, "image/png", png)
	})

	r.Run()
}

func handleWebSocket(ws *websocket.Conn) {
	defer ws.Close()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		data, err := generateBusData()
		if err != nil {
			fmt.Printf("Error generating bus data: %v", err)
			return
		}

		id := time.Now().UnixMilli()
		idStr := fmt.Sprintf("%d", id)

		queueMutex.Lock()
		currentData = data
		dataMap[idStr] = data // Store in map with the ID
		queueMutex.Unlock()

		qrCodeData := QRCodeData{
			ID:   id,
			Data: data,
		}

		if err := websocket.JSON.Send(ws, qrCodeData); err != nil {
			fmt.Printf("Error sending QR code data: %v", err)
			return
		}
	}
}

func generateBusData() (string, error) {
	queueMutex.Lock()
	defer queueMutex.Unlock()

	// Update existing bus times
	updateBusQueue()

	// Maybe add a new bus
	if rand.Float64() < NEW_BUS_PROBABILITY && len(busQueue) < MAX_BUSES {
		addNewBus()
	}

	// Create the data structure
	data := BusStop{
		S: FIXED_BUS_STOP,
		B: busQueue,
	}

	// Marshal the data into a JSON string
	jsonData, err := json.Marshal(data) // Remove MarshalIndent for smaller payload
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

func updateBusQueue() {
	// No need for mutex here as it's called from generateBusData which has the lock
	updatedQueue := make([]Bus, 0, len(busQueue))

	for _, bus := range busQueue {
		// Create a new bus instance instead of modifying the existing one
		updatedBus := Bus{
			L: bus.L,
			M: bus.M - 1,
		}

		if updatedBus.M >= 0 {
			updatedQueue = append(updatedQueue, updatedBus)
		}
	}

	// Sort by arrival time
	sort.Slice(updatedQueue, func(i, j int) bool {
		return updatedQueue[i].M < updatedQueue[j].M
	})

	busQueue = updatedQueue
}

func addNewBus() {
	// No need for mutex here as it's called from generateBusData which has the lock
	lineNumber := rand.Intn(100) + 1
	minutes := rand.Intn(MAX_NEW_BUS_TIME-MIN_NEW_BUS_TIME) + MIN_NEW_BUS_TIME

	newBus := Bus{
		L: fmt.Sprintf("%d", lineNumber),
		M: minutes,
	}

	busQueue = append(busQueue, newBus)
}
