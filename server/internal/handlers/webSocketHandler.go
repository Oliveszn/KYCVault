package handlers

import (
	"kycvault/internal/middleware"
	ws "kycvault/internal/websocket"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type WSHandler struct {
	hub        *ws.Hub
	logger     *zap.Logger
	CORSOrigin string
}

func NewWSHandler(hub *ws.Hub, logger *zap.Logger, CORSOrigin string) *WSHandler {
	return &WSHandler{
		hub:        hub,
		logger:     logger,
		CORSOrigin: CORSOrigin,
	}
}

// GET /ws — client connects here on page load
func (h *WSHandler) Connect(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.Status(http.StatusUnauthorized)
		return
	}

	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			return origin == h.CORSOrigin
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("websocket upgrade failed", zap.Error(err))
		return
	}
	// defer conn.Close()

	h.hub.Register(userID, conn)
	// defer h.hub.Unregister(userID, conn)

	// // keep connection alive, read pings fmiddleware
	// for {
	// 	if _, _, err := conn.ReadMessage(); err != nil {
	// 		break // client disconnected
	// 	}
	// }
	go func() {
		defer h.hub.Unregister(userID, conn)
		defer conn.Close()

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				h.logger.Info("websocket disconnected",
					zap.String("user_id", userID.String()),
					zap.Error(err),
				)
				break
			}
		}
	}()
}
