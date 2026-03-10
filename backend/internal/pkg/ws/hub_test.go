package ws

import (
	"encoding/json"
	"testing"
)

func TestNewHub_InitializedFields(t *testing.T) {
	hub := NewHub()

	if hub == nil {
		t.Fatal("NewHub() returned nil")
	}

	if hub.clients == nil {
		t.Error("clients map is nil")
	}

	if hub.register == nil {
		t.Error("register channel is nil")
	}

	if hub.unregister == nil {
		t.Error("unregister channel is nil")
	}

	if hub.broadcast == nil {
		t.Error("broadcast channel is nil")
	}
}

func TestNewHub_BroadcastChannelBuffered(t *testing.T) {
	hub := NewHub()

	if cap(hub.broadcast) != 256 {
		t.Errorf("broadcast channel capacity = %d, want 256", cap(hub.broadcast))
	}
}

func TestNewHub_RegisterUnregisterUnbuffered(t *testing.T) {
	hub := NewHub()

	if cap(hub.register) != 0 {
		t.Errorf("register channel capacity = %d, want 0 (unbuffered)", cap(hub.register))
	}

	if cap(hub.unregister) != 0 {
		t.Errorf("unregister channel capacity = %d, want 0 (unbuffered)", cap(hub.unregister))
	}
}

func TestNewHub_ClientsMapEmpty(t *testing.T) {
	hub := NewHub()

	if len(hub.clients) != 0 {
		t.Errorf("clients map should be empty, has %d entries", len(hub.clients))
	}
}

func TestMessage_JSONMarshal(t *testing.T) {
	msg := Message{
		Type: "score_update",
		Data: map[string]interface{}{
			"user_id": "abc-123",
			"score":   42,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal Message: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if m["type"] != "score_update" {
		t.Errorf("type = %v, want %q", m["type"], "score_update")
	}

	dataField, ok := m["data"].(map[string]interface{})
	if !ok {
		t.Fatal("data field is not an object")
	}

	if dataField["user_id"] != "abc-123" {
		t.Errorf("data.user_id = %v, want %q", dataField["user_id"], "abc-123")
	}
}

func TestMessage_JSONMarshalNilData(t *testing.T) {
	msg := Message{
		Type: "ping",
		Data: nil,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal Message: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if m["type"] != "ping" {
		t.Errorf("type = %v, want %q", m["type"], "ping")
	}

	if m["data"] != nil {
		t.Errorf("data should be null, got %v", m["data"])
	}
}

func TestMessage_JSONRoundTrip(t *testing.T) {
	original := Message{
		Type: "leaderboard_update",
		Data: []interface{}{"entry1", "entry2"},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded Message
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Type != original.Type {
		t.Errorf("Type = %q, want %q", decoded.Type, original.Type)
	}

	if decoded.Data == nil {
		t.Error("Data should not be nil after round-trip")
	}
}
