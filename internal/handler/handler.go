package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/s2ar/swagat/internal/service"
)

type Handler struct {
	services *service.Service
}

func New(services *service.Service) *Handler {
	return &Handler{services: services}
}

func (h *Handler) InitRouters() *gin.Engine {
	router := gin.New()
	router.LoadHTMLFiles("web/client.html")

	router.GET("/", h.getClientPage)

	router.GET("/ws", func(c *gin.Context) {
		h.serveWS(c.Writer, c.Request)
	})
	return router
}

func (h *Handler) getClientPage(c *gin.Context) {
	c.HTML(http.StatusOK, "client.html", nil)
}

func (h *Handler) serveWS(w http.ResponseWriter, r *http.Request) {
	hub := h.services.Hub
	conn, err := service.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &service.Client{
		Hub:  hub,
		Conn: conn,
		Send: make(chan []byte, 256),
	}
	hub.Register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.WritePump()
	go client.ReadPump()
}
