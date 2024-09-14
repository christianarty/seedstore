package util

import (
	"seedstore/types"
	"testing"
)

func TestConcurrentQueue(t *testing.T) {
	queue := ConcurrentQueue[types.MQTTMessage]{}

	// Test Enqueue and Size
	queue.Enqueue(types.MQTTMessage{Name: "Test1"})
	queue.Enqueue(types.MQTTMessage{Name: "Test2"})

	if queue.Size() != 2 {
		t.Errorf("Expected queue size 2, got %d", queue.Size())
	}

	// Test Dequeue
	item := queue.Dequeue()
	if item.Name != "Test1" {
		t.Errorf("Expected 'Test1', got %s", item.Name)
	}

	if queue.Size() != 1 {
		t.Errorf("Expected queue size 1, got %d", queue.Size())
	}

	// Test IsEmpty
	queue.Dequeue()
	if !queue.IsEmpty() {
		t.Error("Expected queue to be empty")
	}
}
