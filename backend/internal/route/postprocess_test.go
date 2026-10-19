package route

import (
	"testing"

	"indoor-navigation/backend/internal/graph"
)

func TestRemoveCollinear(t *testing.T) {
	straight := []Point{{X: 0, Z: 0}, {X: 0, Z: 2}, {X: 0, Z: 4}, {X: 0, Z: 6}}
	result := removeCollinear(straight)
	if len(result) != 2 || result[0] != straight[0] || result[1] != straight[3] {
		t.Fatalf("straight points not compressed: %+v", result)
	}

	corner := []Point{{X: 0, Z: 0}, {X: 0, Z: 2}, {X: 2, Z: 2}}
	result = removeCollinear(corner)
	if len(result) != 3 {
		t.Fatalf("corner point should be kept: %+v", result)
	}

	duplicates := []Point{{X: 0, Z: 0}, {X: 0, Z: 0}, {X: 2, Z: 0}}
	result = removeCollinear(duplicates)
	if len(result) != 2 || result[0] != duplicates[0] || result[1] != duplicates[2] {
		t.Fatalf("duplicates not removed: %+v", result)
	}

	single := []Point{{X: 1, Z: 1}}
	if got := removeCollinear(single); len(got) != 1 {
		t.Fatalf("single point changed: %+v", got)
	}
}

func TestTransitionTime(t *testing.T) {
	options := DefaultTimeOptions()
	cases := []struct {
		kind graph.EdgeKind
		from graph.FloorBit
		to   graph.FloorBit
		want float64
	}{
		{graph.EdgeStairs, graph.FloorF1, graph.FloorF3, 30},
		{graph.EdgeStairs, graph.FloorF3, graph.FloorF1, 30},
		{graph.EdgeElevator, graph.FloorF1, graph.FloorF3, 30},
		{graph.EdgeElevator, graph.FloorF3, graph.FloorF1, 30},
		{graph.EdgeCorridor, graph.FloorF1, graph.FloorF2, 0},
		{graph.EdgeStairs, 0, graph.FloorF2, 0},
	}
	for _, item := range cases {
		got := transitionTime(item.kind, item.from, item.to, options)
		if !almostEqual(got, item.want) {
			t.Fatalf("transitionTime(%v, %v, %v) = %v, want %v", item.kind, item.from, item.to, got, item.want)
		}
	}
}

func TestNormalizeTimeOptions(t *testing.T) {
	defaults := normalizeTimeOptions(TimeOptions{})
	if defaults != DefaultTimeOptions() {
		t.Fatalf("unexpected defaults: %+v", defaults)
	}

	custom := normalizeTimeOptions(TimeOptions{WalkingSpeedMPS: 1.5, StairSecondsPerFloor: 20})
	if custom.WalkingSpeedMPS != 1.5 || custom.StairSecondsPerFloor != 20 {
		t.Fatalf("custom values overwritten: %+v", custom)
	}
	if custom.ElevatorWaitSeconds != 20 || custom.ElevatorSecondsPerFloor != 5 {
		t.Fatalf("missing defaults not filled: %+v", custom)
	}
}

func TestPostProcessSegments(t *testing.T) {
	store := newTestStore(t, true, true)
	building, _ := store.Building("C5")
	base := graph.NodeID(len(building.Nodes()))
	over := &overlay{
		edges:      map[graph.NodeID][]weightedEdge{},
		points:     map[graph.NodeID]Point{},
		floors:     map[graph.NodeID]graph.FloorBit{},
		baseNodeID: base,
		nodeCount:  int(base),
	}

	d101 := nodeID(t, building, "F1", "F1-D101")
	c1 := nodeID(t, building, "F1", "F1-C1")
	c2 := nodeID(t, building, "F1", "F1-C2")
	f1s := nodeID(t, building, "F1", "F1-S")
	f2s := nodeID(t, building, "F2", "F2-S")
	f2c2 := nodeID(t, building, "F2", "F2-C2")
	f2c1 := nodeID(t, building, "F2", "F2-C1")
	d201 := nodeID(t, building, "F2", "F2-D201")

	steps := []pathStep{
		{From: d101, To: c1, DistanceM: 2, Kind: graph.EdgeCorridor},
		{From: c1, To: c2, DistanceM: 4, Kind: graph.EdgeCorridor},
		{From: c2, To: f1s, DistanceM: 4, Kind: graph.EdgeCorridor},
		{From: f1s, To: f2s, DistanceM: 4, Kind: graph.EdgeStairs},
		{From: f2s, To: f2c2, DistanceM: 4, Kind: graph.EdgeCorridor},
		{From: f2c2, To: f2c1, DistanceM: 4, Kind: graph.EdgeCorridor},
		{From: f2c1, To: d201, DistanceM: 2, Kind: graph.EdgeCorridor},
	}

	segments, transitions, distance, seconds, err := postProcess(steps, building, over, TimeOptions{})
	if err != nil {
		t.Fatalf("post process failed: %v", err)
	}
	if !almostEqual(distance, 24) {
		t.Fatalf("unexpected distance: %v", distance)
	}
	if len(segments) != 2 || segments[0].Floor != "F1" || segments[1].Floor != "F2" {
		t.Fatalf("unexpected segments: %+v", segments)
	}
	if len(segments[0].Points) != 2 || len(segments[1].Points) != 2 {
		t.Fatalf("collinear points not compressed: %+v", segments)
	}
	if !almostEqual(segments[0].DistanceM, 10) || !almostEqual(segments[1].DistanceM, 10) {
		t.Fatalf("unexpected segment distances: %+v", segments)
	}
	if len(transitions) != 1 {
		t.Fatalf("unexpected transitions: %+v", transitions)
	}
	transition := transitions[0]
	if transition.FromFloor != "F1" || transition.ToFloor != "F2" || transition.Kind != graph.EdgeStairs {
		t.Fatalf("unexpected transition: %+v", transition)
	}
	if !almostEqual(transition.EstimatedSeconds, 15) {
		t.Fatalf("unexpected transition seconds: %v", transition.EstimatedSeconds)
	}
	if !almostEqual(seconds, 20.0/1.2+15) {
		t.Fatalf("unexpected total seconds: %v", seconds)
	}
}

func TestPostProcessEmpty(t *testing.T) {
	store := newTestStore(t, true, true)
	building, _ := store.Building("C5")
	segments, transitions, distance, seconds, err := postProcess(nil, building, emptyOverlay(building), TimeOptions{})
	if err != nil {
		t.Fatalf("empty post process failed: %v", err)
	}
	if len(segments) != 0 || len(transitions) != 0 || distance != 0 || seconds != 0 {
		t.Fatalf("unexpected empty result: %+v %+v %v %v", segments, transitions, distance, seconds)
	}
}
