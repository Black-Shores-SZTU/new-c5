package route

import (
	"container/heap"
	"testing"

	"indoor-navigation/backend/internal/graph"
)

func TestMinHeapOrder(t *testing.T) {
	queue := &minHeap{}
	heap.Init(queue)
	for _, item := range []heapItem{
		{node: 5, dist: 2},
		{node: 1, dist: 1},
		{node: 2, dist: 1},
		{node: 3, dist: 3},
		{node: 4, dist: 1},
	} {
		heap.Push(queue, item)
	}

	want := []graph.NodeID{1, 2, 4, 5, 3}
	for _, expected := range want {
		if queue.Len() == 0 {
			t.Fatalf("heap empty before node %d", expected)
		}
		item := heap.Pop(queue).(heapItem)
		if item.node != expected {
			t.Fatalf("pop %d, want %d", item.node, expected)
		}
	}
	if queue.Len() != 0 {
		t.Fatalf("heap should be empty, got %d", queue.Len())
	}
}
