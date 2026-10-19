package route

import (
	"testing"

	"indoor-navigation/backend/internal/graph"
)

func TestCrossFloorFactorTable(t *testing.T) {
	cases := []struct {
		floors int
		base   float64
		want   float64
	}{
		{0, 1.6, 1},
		{1, 1.6, 1},
		{2, 1.6, 1.6},
		{3, 1.6, 2.56},
		{4, 1.6, 4},
		{5, 1.6, 5},
		{2, 1, 1},
		{3, 1, 1},
		{2, 2, 2},
		{3, 2, 3},
		{4, 2, 4},
		{5, 2, 5},
		{2, 0.5, 1},
	}
	for _, item := range cases {
		got := crossFloorFactor(item.floors, item.base)
		if !almostEqual(got, item.want) {
			t.Fatalf("crossFloorFactor(%d, %v) = %v, want %v", item.floors, item.base, got, item.want)
		}
	}
}

func TestEdgeCost(t *testing.T) {
	cases := []struct {
		name     string
		from     graph.FloorBit
		to       graph.FloorBit
		distance float64
		kind     graph.EdgeKind
		want     float64
	}{
		{"corridor ignores penalty", graph.FloorF1, graph.FloorF2, 10, graph.EdgeCorridor, 10},
		{"same floor ignores penalty", graph.FloorF1, graph.FloorF1, 10, graph.EdgeStairs, 10},
		{"stairs two floors", graph.FloorF1, graph.FloorF3, 4, graph.EdgeStairs, 6.4},
		{"elevator downward", graph.FloorF3, graph.FloorF1, 4, graph.EdgeElevator, 6.4},
		{"invalid floor falls back to distance", 0, graph.FloorF2, 7, graph.EdgeStairs, 7},
	}
	for _, item := range cases {
		got := edgeCost(item.from, item.to, item.distance, item.kind, 1.6)
		if !almostEqual(got, item.want) {
			t.Fatalf("%s: got %v, want %v", item.name, got, item.want)
		}
	}
}
