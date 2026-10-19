package route

import (
	"container/heap"

	"indoor-navigation/backend/internal/graph"
)

type heapItem struct {
	node graph.NodeID
	dist float64
}

type minHeap []heapItem

func (h minHeap) Len() int {
	return len(h)
}

func (h minHeap) Less(i, j int) bool {
	if h[i].dist != h[j].dist {
		return h[i].dist < h[j].dist
	}
	return h[i].node < h[j].node
}

func (h minHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *minHeap) Push(value any) {
	*h = append(*h, value.(heapItem))
}

func (h *minHeap) Pop() any {
	old := *h
	last := len(old) - 1
	item := old[last]
	*h = old[:last]
	return item
}

var _ heap.Interface = (*minHeap)(nil)
