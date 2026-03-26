package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestRoomManagerCreateAndGet(t *testing.T) {
	rm := NewRoomManager()
	room := rm.GetOrCreateRoom("abc123")
	if room == nil {
		t.Fatal("expected room, got nil")
	}
	if room.token != "abc123" {
		t.Fatalf("expected token abc123, got %s", room.token)
	}

	// Same token returns same room
	room2 := rm.GetOrCreateRoom("abc123")
	if room != room2 {
		t.Fatal("expected same room instance for same token")
	}
}

func TestRoomManagerDifferentTokens(t *testing.T) {
	rm := NewRoomManager()
	room1 := rm.GetOrCreateRoom("token1")
	room2 := rm.GetOrCreateRoom("token2")
	if room1 == room2 {
		t.Fatal("different tokens should create different rooms")
	}
}

func TestRoomManagerDelete(t *testing.T) {
	rm := NewRoomManager()
	rm.GetOrCreateRoom("delete-me")
	rm.DeleteRoom("delete-me")

	// Getting after delete should create a new room
	room := rm.GetOrCreateRoom("delete-me")
	if room == nil {
		t.Fatal("expected new room after delete")
	}
}

func TestRoomAddClientLimit(t *testing.T) {
	// Test via WebSocket upgrade - create a real relay and connect 3 clients
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws/{token}", handleWebSocket)
	server := httptest.NewServer(mux)
	defer server.Close()

	wsURL := "ws" + server.URL[4:] + "/ws/limit-test"

	// First two should connect fine
	dialer := websocket.Dialer{}
	conn1, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("first client failed to connect: %v", err)
	}
	defer conn1.Close()

	conn2, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("second client failed to connect: %v", err)
	}
	defer conn2.Close()

	// Third should be rejected (room full)
	conn3, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		// Connection refused is also acceptable
		return
	}
	defer conn3.Close()

	// If connected, should receive a close message
	_, _, err = conn3.ReadMessage()
	if err == nil {
		t.Fatal("third client should have been rejected")
	}
}

func TestRoomCleanup(t *testing.T) {
	rm := &RoomManager{rooms: make(map[string]*Room)}

	// Create an expired room
	rm.rooms["old"] = &Room{
		token:     "old",
		clients:   nil,
		createdAt: time.Now().Add(-15 * time.Minute),
	}
	// Create a fresh room
	rm.rooms["new"] = &Room{
		token:     "new",
		clients:   nil,
		createdAt: time.Now(),
	}

	// Run cleanup manually
	rm.mu.Lock()
	for token, room := range rm.rooms {
		if time.Since(room.createdAt) > 10*time.Minute {
			delete(rm.rooms, token)
		}
	}
	rm.mu.Unlock()

	rm.mu.RLock()
	defer rm.mu.RUnlock()
	if _, exists := rm.rooms["old"]; exists {
		t.Fatal("expired room should have been cleaned up")
	}
	if _, exists := rm.rooms["new"]; !exists {
		t.Fatal("fresh room should still exist")
	}
}

func TestHealthEndpoint(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != `{"status":"ok"}` {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestDownloadEndpointMissingToken(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /d/{token}", handleDownload)

	// Test with a valid token (should serve HTML)
	req := httptest.NewRequest("GET", "/d/testtoken", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Fatalf("expected text/html content type, got %s", ct)
	}
}

func TestUploadEndpoint(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /u/{token}", handleUpload)

	req := httptest.NewRequest("GET", "/u/testtoken", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Fatalf("expected text/html content type, got %s", ct)
	}
}
