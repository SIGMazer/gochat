package tui

import (
	"testing"
	"time"
)

func TestTUIMessageHandling(t *testing.T) {
	// Create channels for testing
	outgoingChan := make(chan string, 10)
	incomingChan := make(chan Message, 10)
	
	// Create TUI model with channels
	model := InitModelWithChannels(outgoingChan, incomingChan)
	
	// Test that model was created with channels
	if model.outgoingChan == nil {
		t.Fatal("Outgoing channel not set")
	}
	if model.incomingChan == nil {
		t.Fatal("Incoming channel not set")
	}
	
	// Styles are automatically created in InitModelWithChannels
	
	// Test sending a message through outgoing channel
	testMsg := "Hello, World!"
	
	// Send message
	select {
	case outgoingChan <- testMsg:
		// Message sent successfully
	default:
		t.Fatal("Failed to send message to outgoing channel")
	}
	
	// Verify message was sent
	select {
	case receivedMsg := <-outgoingChan:
		if receivedMsg != testMsg {
			t.Errorf("Message mismatch. Expected: %s, Got: %s", testMsg, receivedMsg)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for outgoing message")
	}
	
	// Test receiving a message through incoming channel
	incomingMsg := Message{From: "TestUser", Text: "Test message"}
	
	// Send message to incoming channel
	select {
	case incomingChan <- incomingMsg:
		// Message sent successfully
	default:
		t.Fatal("Failed to send message to incoming channel")
	}
	
	// Verify message was sent
	select {
	case receivedMsg := <-incomingChan:
		if receivedMsg.From != incomingMsg.From || receivedMsg.Text != incomingMsg.Text {
			t.Errorf("Message mismatch. Expected: %+v, Got: %+v", incomingMsg, receivedMsg)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for incoming message")
	}
}

func TestMessageStruct(t *testing.T) {
	// Test Message struct creation
	msg := Message{From: "Alice", Text: "Hello, Bob!"}
	
	if msg.From != "Alice" {
		t.Errorf("Expected From: 'Alice', got: '%s'", msg.From)
	}
	
	if msg.Text != "Hello, Bob!" {
		t.Errorf("Expected Text: 'Hello, Bob!', got: '%s'", msg.Text)
	}
}
