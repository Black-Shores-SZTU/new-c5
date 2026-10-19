package route

import (
	"context"
	"errors"
	"testing"

	"indoor-navigation/backend/internal/graph"
)

func emptyOverlay(building *graph.BuildingGraph) *overlay {
	base := graph.NodeID(len(building.Nodes()))
	return &overlay{
		edges:      map[graph.NodeID][]weightedEdge{},
		points:     map[graph.NodeID]Point{},
		floors:     map[graph.NodeID]graph.FloorBit{},
		baseNodeID: base,
		nodeCount:  int(base),
	}
}

func nodeID(t *testing.T, building *graph.BuildingGraph, floor, code string) graph.NodeID {
	t.Helper()
	id, ok := building.NodeIDByCode(floor, code)
	if !ok {
		t.Fatalf("node %s|%s missing", floor, code)
	}
	return id
}

func TestShortestPathDirect(t *testing.T) {
	store := newTestStore(t, true, true)
	building, _ := store.Building("C5")
	start := nodeID(t, building, "F1", "F1-D101")
	target := nodeID(t, building, "F1", "F1-C2")

	steps, cost, err := shortestPath(context.Background(), building, emptyOverlay(building), start, target, searchOptions{
		allowedFloors: building.FloorMask(),
		mode:          ModeCombined,
		penaltyBase:   DefaultPenaltyBase,
	})
	if err != nil {
		t.Fatalf("shortest path failed: %v", err)
	}
	if !almostEqual(cost, 6) {
		t.Fatalf("unexpected cost: %v", cost)
	}
	if len(steps) != 2 || steps[0].From != start || steps[1].To != target {
		t.Fatalf("unexpected steps: %+v", steps)
	}
	distance := 0.0
	for _, step := range steps {
		distance += step.DistanceM
	}
	if !almostEqual(distance, 6) {
		t.Fatalf("unexpected distance: %v", distance)
	}
}

func TestShortestPathFloorMask(t *testing.T) {
	store := newTestStore(t, true, true)
	building, _ := store.Building("C5")
	start := nodeID(t, building, "F1", "F1-D101")
	target := nodeID(t, building, "F2", "F2-D201")

	_, _, err := shortestPath(context.Background(), building, emptyOverlay(building), start, target, searchOptions{
		allowedFloors: graph.FloorF1,
		mode:          ModeCombined,
		penaltyBase:   DefaultPenaltyBase,
	})
	if !errors.Is(err, ErrNoRoute) {
		t.Fatalf("expected ErrNoRoute, got %v", err)
	}
}

func TestShortestPathContextCanceled(t *testing.T) {
	store := newTestStore(t, true, true)
	building, _ := store.Building("C5")
	start := nodeID(t, building, "F1", "F1-D101")
	target := nodeID(t, building, "F1", "F1-C2")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err := shortestPath(ctx, building, emptyOverlay(building), start, target, searchOptions{
		allowedFloors: building.FloorMask(),
		mode:          ModeCombined,
		penaltyBase:   DefaultPenaltyBase,
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestShortestPathElevatorOnly(t *testing.T) {
	store := newTestStore(t, true, true)
	building, _ := store.Building("C5")
	start := nodeID(t, building, "F1", "F1-D101")
	target := nodeID(t, building, "F2", "F2-D201")

	steps, _, err := shortestPath(context.Background(), building, emptyOverlay(building), start, target, searchOptions{
		allowedFloors: building.FloorMask(),
		mode:          ModeElevatorOnly,
		penaltyBase:   DefaultPenaltyBase,
	})
	if err != nil {
		t.Fatalf("elevator-only path failed: %v", err)
	}
	elevator, stairs := 0, 0
	for _, step := range steps {
		switch step.Kind {
		case graph.EdgeElevator:
			elevator++
		case graph.EdgeStairs:
			stairs++
		}
	}
	if elevator != 1 || stairs != 0 {
		t.Fatalf("unexpected edge kinds: elevator=%d stairs=%d steps=%+v", elevator, stairs, steps)
	}
}

func TestShortestPathElevatorOnlyNoRoute(t *testing.T) {
	raw := testRaw(true)
	edges := make([]graph.RawEdge, 0, len(raw.Edges))
	for _, edge := range raw.Edges {
		if edge.EdgeType != "elevator" {
			edges = append(edges, edge)
		}
	}
	raw.Edges = edges

	store, err := graph.Build(raw)
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	building, _ := store.Building("C5")
	start := nodeID(t, building, "F1", "F1-D101")
	target := nodeID(t, building, "F2", "F2-D201")

	_, _, err = shortestPath(context.Background(), building, emptyOverlay(building), start, target, searchOptions{
		allowedFloors: building.FloorMask(),
		mode:          ModeElevatorOnly,
		penaltyBase:   DefaultPenaltyBase,
	})
	if !errors.Is(err, ErrNoRoute) {
		t.Fatalf("expected ErrNoRoute, got %v", err)
	}
}
