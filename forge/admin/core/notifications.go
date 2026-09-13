package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Notification represents a real-time event
type Notification struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Type      string    `json:"type"` // info, success, warning, error
	Timestamp time.Time `json:"timestamp"`
	ActionURL string    `json:"action_url,omitempty"`
}

// NotificationHub manages SSE connections and broadcasts
type NotificationHub struct {
	clients map[chan Notification]struct{}
	mu      sync.RWMutex
}

// NewNotificationHub creates a new hub
func NewNotificationHub() *NotificationHub {
	return &NotificationHub{
		clients: make(map[chan Notification]struct{}),
	}
}

// Subscribe adds a new client channel
func (h *NotificationHub) Subscribe() chan Notification {
	h.mu.Lock()
	defer h.mu.Unlock()
	ch := make(chan Notification, 20)
	h.clients[ch] = struct{}{}
	return ch
}

// Unsubscribe removes a client channel
func (h *NotificationHub) Unsubscribe(ch chan Notification) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[ch]; ok {
		delete(h.clients, ch)
		close(ch)
	}
}

// Notify broadcasts a notification to all subscribers
func (h *NotificationHub) Notify(n Notification) {
	if n.ID == "" {
		n.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	if n.Timestamp.IsZero() {
		n.Timestamp = time.Now()
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients {
		select {
		case ch <- n:
		default:
			// Client slow, skip to prevent blocking
		}
	}
}

// SSEHandler handles the browser SSE connection
func (h *NotificationHub) SSEHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := h.Subscribe()

	// Send initial heartbeat/ack
	fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"ok\"}\n\n")
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case n, ok := <-ch:
			if !ok {
				return
			}
			data, _ := json.Marshal(n)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-time.After(30 * time.Second):
			// Keep-alive heartbeat
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		case <-ctx.Done():
			h.Unsubscribe(ch)
			return
		}
	}
}
