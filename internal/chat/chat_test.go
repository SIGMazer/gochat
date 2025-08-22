package chat

import (
	"testing"
	"time"
	"net"
	"gochat/internal/tui"
)

// Mock connection for testing
type mockConn struct {
	net.Conn
}

func (m *mockConn) Close() error { return nil }
func (m *mockConn) Read(b []byte) (n int, err error) { return 0, nil }
func (m *mockConn) Write(b []byte) (n int, err error) { return 0, nil }
func (m *mockConn) RemoteAddr() net.Addr { return nil }
func (m *mockConn) LocalAddr() net.Addr { return nil }

func TestChatRoomMessageChannel(t *testing.T) {
	// Create a chat room
	room := NewRoom()
	
	// Create a channel to receive messages
	msgChan := make(chan tui.Message, 10)
	
	// Set the TUI message channel
	room.SetTUIMessageChannel(msgChan)
	
	// Test sending a message through the channel
	testMsg := tui.Message{From: "TestUser", Text: "Hello, World!"}
	
	// Send message through the channel
	select {
	case msgChan <- testMsg:
		// Message sent successfully
	default:
		t.Fatal("Failed to send message to channel")
	}
	
	// Verify the message was sent
	select {
	case receivedMsg := <-msgChan:
		if receivedMsg.From != testMsg.From || receivedMsg.Text != testMsg.Text {
			t.Errorf("Message mismatch. Expected: %+v, Got: %+v", testMsg, receivedMsg)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for message")
	}
}

func TestChatRoomAddPeer(t *testing.T) {
	room := NewRoom()
	
	// Mock connection
	conn := &mockConn{}
	
	// Add a peer
	room.AddPeer("TestPeer", conn)
	
	// Verify peer was added
	if len(room.Peers) != 1 {
		t.Errorf("Expected 1 peer, got %d", len(room.Peers))
	}
	
	if room.Peers[0].Name != "TestPeer" {
		t.Errorf("Expected peer name 'TestPeer', got '%s'", room.Peers[0].Name)
	}
}

func TestChatRoomRemovePeer(t *testing.T) {
	room := NewRoom()
	
	// Add a peer
	conn := &mockConn{}
	room.AddPeer("TestPeer", conn)
	
	// Get the peer's UUID
	peerUUID := room.Peers[0].uuid
	
	// Remove the peer
	room.RemovePeer(peerUUID)
	
	// Verify peer was removed
	if len(room.Peers) != 0 {
		t.Errorf("Expected 0 peers after removal, got %d", len(room.Peers))
	}
}
