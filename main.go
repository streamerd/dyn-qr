package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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

var dataMap = make(map[string]string)

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
		id := c.Param("id") // Get the ID from the URL parameter

		// Retrieve the data associated with the ID (this is an example, you need to implement this)
		data, err := RetrieveDataByID(id)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		png, err := qrcode.Encode(data, qrcode.Medium, 256)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.Data(http.StatusOK, "image/png", png)
	})

	r.Run()
}

func handleWebSocket(ws *websocket.Conn) {
	defer ws.Close()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		id := time.Now().UnixNano()

		data, err := RandomJSON()
		if err != nil {
			fmt.Printf("Error generating random JSON: %v", err)
			return
		}

		qrCodeData := QRCodeData{
			ID:   id,
			Data: data,
		}

		// Store the data in the map
		dataMap[fmt.Sprintf("%d", id)] = data

		// Publish the QR code data over WebSocket
		if err := websocket.JSON.Send(ws, qrCodeData); err != nil {
			fmt.Printf("Error sending QR code data: %v", err)
			return
		}
	}
}

func RandomJSON() (string, error) {
	rand.Seed(uint64((time.Now().UnixNano() / 1000000) % 1000000))

	// Generate a random stop number between 1 and 10000.
	stop := rand.Intn(10000) + 1

	// Generate random bus data
	numBuses := rand.Intn(5) + 1 // Random number of buses between 1 and 5
	buses := make([]map[string]int, numBuses)

	for i := 0; i < numBuses; i++ {
		busLine := fmt.Sprintf("Bus %d", rand.Intn(100)+1) // Random bus line between 1 and 100
		minutesLeft := rand.Intn(30) + 1                   // Random minutes left between 1 and 30
		buses[i] = map[string]int{busLine: minutesLeft}
	}

	// Create the data structure
	data := map[string]interface{}{
		"stop":     stop,
		"incoming": buses,
	}

	// Marshal the data into a JSON string
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

func RetrieveDataByID(id string) (string, error) {
	data, exists := dataMap[id]
	if !exists {
		return "", fmt.Errorf("data not found for ID: %s", id)
	}
	return data, nil
}
