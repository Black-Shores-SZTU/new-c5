package route

import (
	"context"
	"errors"
	"math"
	"testing"

	"indoor-navigation/backend/internal/graph"
)

func nearestItemByType(items []NearestItem, facilityType graph.TypeMask) (NearestItem, bool) {
	for _, item := range items {
		if item.Type == facilityType {
			return item, true
		}
	}
	return NearestItem{}, false
}

func TestNearestFacilities(t *testing.T) {
	planner := NewPlanner(newTestStore(t, true, true))
	result, err := planner.Nearest(context.Background(), NearestRequest{
		Origin: NearestOrigin{
			Building: "C5",
			Floor:    "F1",
			RoomNo:   "101",
			CenterX:  0,
			CenterZ:  0,
		},
	})
	if err != nil {
		t.Fatalf("nearest failed: %v", err)
	}
	if len(result.Items) != 5 {
		t.Fatalf("expected 5 items, got %+v", result.Items)
	}

	cases := []struct {
		facilityType graph.TypeMask
		code         string
		floor        string
		hops         int
		distance     float64
	}{
		{graph.TypeExit, "C5-F1-EXIT", "F1", 1, 2},
		{graph.TypeRestroom, "C5-F1-RESTROOM", "F1", 2, 6},
		{graph.TypeElevator, "C5-F1-ELEVATOR", "F1", 2, math.Hypot(9, 2)},
		{graph.TypeWater, "C5-F1-WATER", "F1", 2, math.Hypot(2, 2)},
		{graph.TypeStairs, "C5-F1-STAIRS", "F1", 3, 10},
	}

	for _, item := range cases {
		actual, ok := nearestItemByType(result.Items, item.facilityType)
		if !ok || !actual.Found {
			t.Fatalf("facility type %v not found: %+v", item.facilityType, result.Items)
		}
		if actual.Code != item.code || actual.Floor != item.floor || actual.Hops != item.hops {
			t.Fatalf("unexpected facility: %+v", actual)
		}
		if !almostEqual(actual.StraightDistanceM, item.distance) {
			t.Fatalf("unexpected straight distance for %s: %v", actual.Code, actual.StraightDistanceM)
		}
	}
}

func TestNearestFloorMask(t *testing.T) {
	planner := NewPlanner(newTestStore(t, true, true))
	result, err := planner.Nearest(context.Background(), NearestRequest{
		Origin: NearestOrigin{
			Building: "C5",
			Floor:    "F1",
			RoomNo:   "101",
			CenterX:  0,
			CenterZ:  0,
		},
		AllowedFloors: graph.FloorF1,
	})
	if err != nil {
		t.Fatalf("nearest failed: %v", err)
	}
	for _, item := range result.Items {
		if item.Found && item.Floor != "F1" {
			t.Fatalf("floor mask returned non-F1 facility: %+v", item)
		}
	}

	_, err = planner.Nearest(context.Background(), NearestRequest{
		Origin: NearestOrigin{
			Building: "C5",
			Floor:    "F1",
			RoomNo:   "101",
			CenterX:  0,
			CenterZ:  0,
		},
		AllowedFloors: graph.FloorF2,
	})
	if !errors.Is(err, ErrFloorUnavailable) {
		t.Fatalf("expected ErrFloorUnavailable, got %v", err)
	}
}

func TestNearestMissingType(t *testing.T) {
	planner := NewPlanner(newTestStore(t, true, true))
	result, err := planner.Nearest(context.Background(), NearestRequest{
		Origin: NearestOrigin{
			Building: "C5",
			Floor:    "F1",
			RoomNo:   "101",
			CenterX:  0,
			CenterZ:  0,
		},
		Types: []graph.TypeMask{graph.TypeAuditorium},
	})
	if err != nil {
		t.Fatalf("nearest failed: %v", err)
	}
	if len(result.Items) != 1 || result.Items[0].Found {
		t.Fatalf("expected missing auditorium, got %+v", result.Items)
	}
}

func TestNearestTieBreakByStraightDistance(t *testing.T) {
	raw := &graph.RawData{
		Building: graph.RawBuilding{ID: 1, Code: "C5", Name: "C5栋", FullName: "C5 公共教学楼"},
		Floors: []graph.RawFloor{
			{ID: 1, BuildingID: 1, Code: "F1", Name: "一层", Level: 1, MinX: -10, MaxX: 10, MinZ: -10, MaxZ: 10, FloorHeight: 4},
		},
		Nodes: []graph.RawNode{
			{ID: 1, FloorID: 1, Code: "F1-C", Name: "走廊", Type: "corridor", X: 0, Z: 2},
			{ID: 2, FloorID: 1, Code: "F1-O", Name: "起点", Type: "corridor", X: 0, Z: 0},
			{ID: 3, FloorID: 1, Code: "F1-R1", Name: "卫生间1", Type: "door", X: 1, Z: 2,
				RefType: testStringPtr("facility"), RefCode: testStringPtr("C5-F1-R1")},
			{ID: 4, FloorID: 1, Code: "F1-R2", Name: "卫生间2", Type: "door", X: 5, Z: 2,
				RefType: testStringPtr("facility"), RefCode: testStringPtr("C5-F1-R2")},
		},
		Facilities: []graph.RawFacility{
			{ID: 1, FloorID: 1, Code: "C5-F1-R1", Name: "卫生间1", Type: "restroom", Corridor: "north",
				CenterX: 1, CenterZ: 2, SizeX: 2, SizeZ: 2, DoorX: 1, DoorZ: 2},
			{ID: 2, FloorID: 1, Code: "C5-F1-R2", Name: "卫生间2", Type: "restroom", Corridor: "north",
				CenterX: 5, CenterZ: 2, SizeX: 2, SizeZ: 2, DoorX: 5, DoorZ: 2},
		},
		Edges: []graph.RawEdge{
			{ID: 1, FromNodeID: 1, ToNodeID: 2, EdgeType: "corridor", DistanceM: 2, Bidirectional: true},
			{ID: 2, FromNodeID: 1, ToNodeID: 3, EdgeType: "corridor", DistanceM: 1, Bidirectional: true},
			{ID: 3, FromNodeID: 1, ToNodeID: 4, EdgeType: "corridor", DistanceM: 5, Bidirectional: true},
		},
	}

	report := graph.Validate(raw, true)
	if !report.Valid {
		t.Fatalf("tie-break graph invalid: %+v", report.Errors)
	}
	store, err := graph.Build(raw)
	if err != nil {
		t.Fatalf("build tie-break graph: %v", err)
	}
	building, ok := store.Building("C5")
	if !ok {
		t.Fatal("building missing")
	}
	originID, ok := building.NodeIDByCode("F1", "F1-O")
	if !ok {
		t.Fatal("origin node missing")
	}

	result, err := NewPlanner(store).Nearest(context.Background(), NearestRequest{
		Origin: NearestOrigin{Building: "C5", NodeID: &originID},
		Types:  []graph.TypeMask{graph.TypeRestroom},
	})
	if err != nil {
		t.Fatalf("nearest failed: %v", err)
	}
	if len(result.Items) != 1 || !result.Items[0].Found {
		t.Fatalf("restroom not found: %+v", result.Items)
	}
	if result.Items[0].Code != "C5-F1-R1" || result.Items[0].Hops != 2 {
		t.Fatalf("unexpected tie-break result: %+v", result.Items[0])
	}
}
