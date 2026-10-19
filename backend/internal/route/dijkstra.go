package route

import (
	"container/heap"
	"context"
	"math"

	"indoor-navigation/backend/internal/graph"
)

const invalidNode graph.NodeID = ^graph.NodeID(0)

type weightedEdge struct {
	To        graph.NodeID
	DistanceM float64
	Cost      float64
	Kind      graph.EdgeKind
}

type edgeRef struct {
	DistanceM float64
	Kind      graph.EdgeKind
}

type pathStep struct {
	From      graph.NodeID
	To        graph.NodeID
	DistanceM float64
	Kind      graph.EdgeKind
}

// overlay 保存 CPath 临时节点、边与坐标，不修改只读图。
type overlay struct {
	edges      map[graph.NodeID][]weightedEdge
	points     map[graph.NodeID]Point
	floors     map[graph.NodeID]graph.FloorBit
	baseNodeID graph.NodeID
	nodeCount  int
}

type searchOptions struct {
	allowedFloors graph.FloorBit
	mode          Mode
	penaltyBase   float64
}

// shortestPath 在只读图 + 临时 overlay 上执行二叉堆 Dijkstra。
func shortestPath(
	ctx context.Context,
	building *graph.BuildingGraph,
	over *overlay,
	start graph.NodeID,
	target graph.NodeID,
	opts searchOptions,
) ([]pathStep, float64, error) {
	dist := make([]float64, over.nodeCount)
	prev := make([]graph.NodeID, over.nodeCount)
	prevEdge := make([]edgeRef, over.nodeCount)
	for index := range dist {
		dist[index] = math.Inf(1)
		prev[index] = invalidNode
	}

	dist[start] = 0
	queue := &minHeap{{node: start, dist: 0}}
	heap.Init(queue)

	for queue.Len() > 0 {
		select {
		case <-ctx.Done():
			return nil, 0, ctx.Err()
		default:
		}

		item := heap.Pop(queue).(heapItem)
		if item.dist > dist[item.node] {
			continue
		}
		if item.node == target {
			break
		}

		from := item.node
		relax := func(to graph.NodeID, distanceM, cost float64, kind graph.EdgeKind) {
			if opts.mode == ModeElevatorOnly && kind == graph.EdgeStairs {
				return
			}
			if to < over.baseNodeID {
				node, ok := building.Node(to)
				if !ok {
					return
				}
				if opts.allowedFloors != 0 && node.Floor&opts.allowedFloors == 0 {
					return
				}
			}

			next := dist[from] + cost
			if next >= dist[to] {
				return
			}
			dist[to] = next
			prev[to] = from
			prevEdge[to] = edgeRef{DistanceM: distanceM, Kind: kind}
			heap.Push(queue, heapItem{node: to, dist: next})
		}

		if from < over.baseNodeID {
			fromNode, ok := building.Node(from)
			if ok {
				for _, edge := range building.Neighbors(from) {
					toNode, toOK := building.Node(edge.To)
					if !toOK {
						continue
					}
					cost := edgeCost(fromNode.Floor, toNode.Floor, edge.DistanceM, edge.Kind, opts.penaltyBase)
					relax(edge.To, edge.DistanceM, cost, edge.Kind)
				}
			}
		}

		for _, edge := range over.edges[from] {
			relax(edge.To, edge.DistanceM, edge.Cost, edge.Kind)
		}
	}

	if math.IsInf(dist[target], 1) {
		return nil, 0, ErrNoRoute
	}

	steps, err := reconstruct(prev, prevEdge, start, target)
	if err != nil {
		return nil, 0, err
	}
	return steps, dist[target], nil
}

func reconstruct(prev []graph.NodeID, prevEdge []edgeRef, start, target graph.NodeID) ([]pathStep, error) {
	steps := make([]pathStep, 0)
	for current := target; current != start; {
		parent := prev[current]
		if parent == invalidNode {
			return nil, ErrNoRoute
		}
		steps = append(steps, pathStep{
			From:      parent,
			To:        current,
			DistanceM: prevEdge[current].DistanceM,
			Kind:      prevEdge[current].Kind,
		})
		current = parent
	}
	for left, right := 0, len(steps)-1; left < right; left, right = left+1, right-1 {
		steps[left], steps[right] = steps[right], steps[left]
	}
	return steps, nil
}
