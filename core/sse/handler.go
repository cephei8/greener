package sse

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func NewHandler(hub *Hub) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("session", c)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
		}

		auth, ok := sess.Values["authenticated"].(bool)
		if !ok || !auth {
			return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
		}

		userIDStr, ok := sess.Values["user_id"].(string)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid session")
		}

		w := c.Response()
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")

		client := &Client{
			ID:     uuid.New().String(),
			UserID: userIDStr,
			Send:   make(chan []byte, 256),
		}

		hub.Register(client)
		defer hub.Unregister(client)

		// Send connected event with client ID
		connectedData, _ := json.Marshal(map[string]string{
			"status":    "connected",
			"client_id": client.ID,
		})
		initialEvent := formatSSEMessage("connected", connectedData)
		if _, err := w.Write(initialEvent); err != nil {
			return err
		}
		w.Flush()

		ctx := c.Request().Context()
		for {
			select {
			case <-ctx.Done():
				return nil
			case msg, ok := <-client.Send:
				if !ok {
					return nil
				}
				if _, err := w.Write(msg); err != nil {
					return err
				}
				w.Flush()
			}
		}
	}
}

func NewSetPrimaryHandler(hub *Hub) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get("session", c)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
		}

		auth, ok := sess.Values["authenticated"].(bool)
		if !ok || !auth {
			return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
		}

		userIDStr, ok := sess.Values["user_id"].(string)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid session")
		}

		var req struct {
			ClientID string `json:"client_id"`
		}
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}

		if req.ClientID == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "client_id is required")
		}

		if !hub.SetPrimary(userIDStr, req.ClientID) {
			return echo.NewHTTPError(http.StatusNotFound, "client not found")
		}

		return c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
		})
	}
}
