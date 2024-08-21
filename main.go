package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
	"golang.org/x/net/websocket"
)

type QRCodeData struct {
	ID   int64  `json:"id"`
	Data string `json:"data"`
}

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
		id := c.GetInt64("id")

		png, err := qrcode.Encode(fmt.Sprintf("%d", id), qrcode.Medium, 256)
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
		data := fmt.Sprintf("Dynamic data %d", id)
		qrCodeData := QRCodeData{
			ID:   id,
			Data: data,
		}

		// Publish the QR code data over WebSocket
		if err := websocket.JSON.Send(ws, qrCodeData); err != nil {
			fmt.Printf("Error sending QR code data: %v", err)
			return
		}
	}
}
