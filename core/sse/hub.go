package sse

import (
	"encoding/json"
	"sync"
)

type Event struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type MCPQueryEvent struct {
	Page  string `json:"page"`
	Query string `json:"query"`
}

type Client struct {
	ID     string
	UserID string
	Send   chan []byte
}

type Hub struct {
	clients    map[string]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *userMessage
	mu         sync.RWMutex
}

type userMessage struct {
	userID string
	data   []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *userMessage, 256),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.UserID] == nil {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.UserID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.clients, client.UserID)
					}
				}
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			if clients, ok := h.clients[msg.userID]; ok {
				for client := range clients {
					select {
					case client.Send <- msg.data:
					default:
						close(client.Send)
						delete(clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) BroadcastToUser(userID string, event Event) error {
	data, err := json.Marshal(event.Data)
	if err != nil {
		return err
	}

	sseData := formatSSEMessage(event.Type, data)

	h.broadcast <- &userMessage{
		userID: userID,
		data:   sseData,
	}
	return nil
}

func (h *Hub) BroadcastMCPQuery(userID string, page string, query string) error {
	return h.BroadcastToUser(userID, Event{
		Type: "mcp-query",
		Data: MCPQueryEvent{
			Page:  page,
			Query: query,
		},
	})
}

func formatSSEMessage(eventType string, data []byte) []byte {
	msg := []byte("event: " + eventType + "\ndata: ")
	msg = append(msg, data...)
	msg = append(msg, '\n', '\n')
	return msg
}

func (h *Hub) ClientCount(userID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if clients, ok := h.clients[userID]; ok {
		return len(clients)
	}
	return 0
}

func (h *Hub) TotalClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	total := 0
	for _, clients := range h.clients {
		total += len(clients)
	}
	return total
}
