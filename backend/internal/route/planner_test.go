package route

import (
	"context"
	"errors"
	"math"
	"testing"

	"indoor-navigation/backend/internal/graph"
)

func testStringPtr(value string) *string {
	return &value
}

func testRaw(includeCrossFloor bool) *graph.RawData {
	nodes := []graph.RawNode{
		{ID: 1, FloorID: 1, Code: "F1-D101", Name: "101门", Type: "door", X: 0, Z: 0,
			RefType: testStringPtr("room"), RefCode: testStringPtr("101")},
		{ID: 2, FloorID: 1, Code: "F1-C1", Name: "F1走廊1", Type: "corridor", X: 0, Z: 2,
			RefType: testStringPtr("facility"), RefCode: testStringPtr("C5-F1-EXIT")},
		{ID: 3, FloorID: 1, Code: "F1-C2", Name: "F1走廊2", Type: "corridor", X: 0, Z: 6,
			RefType: testStringPtr("facility"), RefCode: testStringPtr("C5-F1-RESTROOM")},
		{ID: 4, FloorID: 1, Code: "F1-S", Name: "F1楼梯", Type: "core", X: 0, Z: 10,
			RefType: testStringPtr("facility"), RefCode: testStringPtr("C5-F1-STAIRS")},
		{ID: 5, FloorID: 1, Code: "F1-E", Name: "F1电梯", Type: "core", X: 9, Z: 2,
			RefType: testStringPtr("facility"), RefCode: testStringPtr("C5-F1-ELEVATOR")},
		{ID: 6, FloorID: 2, Code: "F2-D201", Name: "201门", Type: "door", X: 0, Z: 0,
			RefType: testStringPtr("room"), RefCode: testStringPtr("201")},
		{ID: 7, FloorID: 2, Code: "F2-C1", Name: "F2走廊1", Type: "corridor", X: 0, Z: 2,
			RefType: testStringPtr("facility"), RefCode: testStringPtr("C5-F2-EXIT")},
		{ID: 8, FloorID: 2, Code: "F2-C2", Name: "F2走廊2", Type: "corridor", X: 0, Z: 6,
			RefType: testStringPtr("facility"), RefCode: testStringPtr("C5-F2-RESTROOM")},
		{ID: 9, FloorID: 2, Code: "F2-S", Name: "F2楼梯", Type: "core", X: 0, Z: 10,
			RefType: testStringPtr("facility"), RefCode: testStringPtr("C5-F2-STAIRS")},
		{ID: 10, FloorID: 2, Code: "F2-E", Name: "F2电梯", Type: "core", X: 9, Z: 2,
			RefType: testStringPtr("facility"), RefCode: testStringPtr("C5-F2-ELEVATOR")},
		{ID: 11, FloorID: 1, Code: "F1-W", Name: "F1开水间", Type: "door", X: -2, Z: 2,
			RefType: testStringPtr("facility"), RefCode: testStringPtr("C5-F1-WATER")},
		{ID: 12, FloorID: 2, Code: "F2-W", Name: "F2开水间", Type: "door", X: -2, Z: 2,
			RefType: testStringPtr("facility"), RefCode: testStringPtr("C5-F2-WATER")},
	}

	edges := []graph.RawEdge{
		{ID: 1, FromNodeID: 1, ToNodeID: 2, EdgeType: "corridor", DistanceM: 2, Bidirectional: true},
		{ID: 2, FromNodeID: 2, ToNodeID: 3, EdgeType: "corridor", DistanceM: 4, Bidirectional: true},
		{ID: 3, FromNodeID: 3, ToNodeID: 4, EdgeType: "corridor", DistanceM: 4, Bidirectional: true},
		{ID: 4, FromNodeID: 2, ToNodeID: 5, EdgeType: "corridor", DistanceM: 9, Bidirectional: true},
		{ID: 5, FromNodeID: 6, ToNodeID: 7, EdgeType: "corridor", DistanceM: 2, Bidirectional: true},
		{ID: 6, FromNodeID: 7, ToNodeID: 8, EdgeType: "corridor", DistanceM: 4, Bidirectional: true},
		{ID: 7, FromNodeID: 8, ToNodeID: 9, EdgeType: "corridor", DistanceM: 4, Bidirectional: true},
		{ID: 8, FromNodeID: 7, ToNodeID: 10, EdgeType: "corridor", DistanceM: 9, Bidirectional: true},
		{ID: 11, FromNodeID: 11, ToNodeID: 2, EdgeType: "corridor", DistanceM: 2, Bidirectional: true},
		{ID: 12, FromNodeID: 12, ToNodeID: 7, EdgeType: "corridor", DistanceM: 2, Bidirectional: true},
	}
	if includeCrossFloor {
		edges = append(edges,
			graph.RawEdge{ID: 9, FromNodeID: 4, ToNodeID: 9, EdgeType: "stairs", DistanceM: 4, Bidirectional: true},
			graph.RawEdge{ID: 10, FromNodeID: 5, ToNodeID: 10, EdgeType: "elevator", DistanceM: 4, Bidirectional: true},
		)
	}

	facilities := []graph.RawFacility{
		{ID: 1, FloorID: 1, Code: "C5-F1-EXIT", Name: "安全出口", Type: "exit", Corridor: "north",
			CenterX: 0, CenterZ: 2, SizeX: 2, SizeZ: 1, DoorX: 0, DoorZ: 2},
		{ID: 2, FloorID: 1, Code: "C5-F1-RESTROOM", Name: "卫生间", Type: "restroom", Corridor: "north",
			CenterX: 0, CenterZ: 6, SizeX: 2, SizeZ: 2, DoorX: 0, DoorZ: 6},
		{ID: 3, FloorID: 1, Code: "C5-F1-STAIRS", Name: "楼梯", Type: "stairs", Corridor: "NW",
			CenterX: 0, CenterZ: 10, SizeX: 3, SizeZ: 3, DoorX: 0, DoorZ: 10},
		{ID: 4, FloorID: 1, Code: "C5-F1-ELEVATOR", Name: "电梯", Type: "elevator", Corridor: "NE",
			CenterX: 9, CenterZ: 2, SizeX: 2, SizeZ: 2, DoorX: 9, DoorZ: 2},
		{ID: 5, FloorID: 1, Code: "C5-F1-WATER", Name: "开水间", Type: "water", Corridor: "north",
			CenterX: -2, CenterZ: 2, SizeX: 2, SizeZ: 2, DoorX: -2, DoorZ: 2},
		{ID: 6, FloorID: 2, Code: "C5-F2-EXIT", Name: "安全出口", Type: "exit", Corridor: "north",
			CenterX: 0, CenterZ: 2, SizeX: 2, SizeZ: 1, DoorX: 0, DoorZ: 2},
		{ID: 7, FloorID: 2, Code: "C5-F2-RESTROOM", Name: "卫生间", Type: "restroom", Corridor: "north",
			CenterX: 0, CenterZ: 6, SizeX: 2, SizeZ: 2, DoorX: 0, DoorZ: 6},
		{ID: 8, FloorID: 2, Code: "C5-F2-STAIRS", Name: "楼梯", Type: "stairs", Corridor: "NW",
			CenterX: 0, CenterZ: 10, SizeX: 3, SizeZ: 3, DoorX: 0, DoorZ: 10},
		{ID: 9, FloorID: 2, Code: "C5-F2-ELEVATOR", Name: "电梯", Type: "elevator", Corridor: "NE",
			CenterX: 9, CenterZ: 2, SizeX: 2, SizeZ: 2, DoorX: 9, DoorZ: 2},
		{ID: 10, FloorID: 2, Code: "C5-F2-WATER", Name: "开水间", Type: "water", Corridor: "north",
			CenterX: -2, CenterZ: 2, SizeX: 2, SizeZ: 2, DoorX: -2, DoorZ: 2},
	}

	return &graph.RawData{
		Building: graph.RawBuilding{ID: 1, Code: "C5", Name: "C5栋", FullName: "C5 公共教学楼"},
		Floors: []graph.RawFloor{
			{ID: 1, BuildingID: 1, Code: "F1", Name: "一层", Level: 1, MinX: -20, MaxX: 20, MinZ: -10, MaxZ: 20, FloorHeight: 4},
			{ID: 2, BuildingID: 1, Code: "F2", Name: "二层", Level: 2, MinX: -20, MaxX: 20, MinZ: -10, MaxZ: 20, FloorHeight: 4},
		},
		Rooms: []graph.RawRoom{
			{ID: 1, FloorID: 1, RoomNo: "101", Name: "101", Type: "classroom", Corridor: "north",
				CenterX: 0, CenterZ: 0, SizeX: 4, SizeZ: 4, DoorX: 0, DoorZ: 0, OpenSide: "south"},
			{ID: 2, FloorID: 2, RoomNo: "201", Name: "201", Type: "classroom", Corridor: "north",
				CenterX: 0, CenterZ: 0, SizeX: 4, SizeZ: 4, DoorX: 0, DoorZ: 0, OpenSide: "south"},
		},
		Facilities: facilities,
		Nodes:      nodes,
		Edges:      edges,
	}
}

func newTestStore(t *testing.T, includeCrossFloor, validate bool) *graph.GraphStore {
	t.Helper()
	raw := testRaw(includeCrossFloor)
	if validate {
		report := graph.Validate(raw, true)
		if !report.Valid {
			t.Fatalf("测试图数据不合法: %+v", report.Errors)
		}
	}
	store, err := graph.Build(raw)
	if err != nil {
		t.Fatalf("构建测试图失败: %v", err)
	}
	return store
}

func testEndpoint(floor, roomNo string, x, z float64) Endpoint {
	return Endpoint{Building: "C5", Floor: floor, RoomNo: roomNo, CenterX: x, CenterZ: z}
}

func almostEqual(left, right float64) bool {
	return math.Abs(left-right) < 1e-6
}

func TestPlanCombinedPrefersStairs(t *testing.T) {
	planner := NewPlanner(newTestStore(t, true, true))
	result, err := planner.Plan(context.Background(), Request{
		From: testEndpoint("F1", "101", 0, 0),
		To:   testEndpoint("F2", "201", 0, 0),
	})
	if err != nil {
		t.Fatalf("plan failed: %v", err)
	}
	if !almostEqual(result.DistanceM, 24) || !almostEqual(result.Cost, 24) {
		t.Fatalf("unexpected distance/cost: %+v", result)
	}
	if len(result.Transitions) != 1 || result.Transitions[0].Kind != graph.EdgeStairs {
		t.Fatalf("expected one stairs transition, got %+v", result.Transitions)
	}
	if len(result.Segments) != 2 {
		t.Fatalf("expected two floor segments, got %+v", result.Segments)
	}
	if result.Segments[0].Floor != "F1" || result.Segments[1].Floor != "F2" {
		t.Fatalf("unexpected segment floors: %+v", result.Segments)
	}
	if len(result.Segments[0].Points) != 2 {
		t.Fatalf("expected collinear compression to keep 2 points, got %+v", result.Segments[0].Points)
	}
	if !almostEqual(result.Segments[0].DistanceM, 10) || !almostEqual(result.Segments[1].DistanceM, 10) {
		t.Fatalf("unexpected segment distances: %+v", result.Segments)
	}
	if result.EstimatedSeconds != 32 {
		t.Fatalf("expected estimated 32s, got %d", result.EstimatedSeconds)
	}
	if !almostEqual(result.Transitions[0].EstimatedSeconds, 15) {
		t.Fatalf("unexpected stairs time: %+v", result.Transitions[0])
	}
	if !almostEqual(result.Segments[0].EstimatedSeconds, 10.0/1.2) {
		t.Fatalf("unexpected segment time: %+v", result.Segments[0])
	}
}

func TestPlanElevatorOnly(t *testing.T) {
	planner := NewPlanner(newTestStore(t, true, true))
	result, err := planner.Plan(context.Background(), Request{
		From: testEndpoint("F1", "101", 0, 0),
		To:   testEndpoint("F2", "201", 0, 0),
		Mode: ModeElevatorOnly,
	})
	if err != nil {
		t.Fatalf("plan failed: %v", err)
	}
	if !almostEqual(result.DistanceM, 26) {
		t.Fatalf("unexpected distance: %+v", result)
	}
	if len(result.Transitions) != 1 || result.Transitions[0].Kind != graph.EdgeElevator {
		t.Fatalf("expected one elevator transition, got %+v", result.Transitions)
	}
	if len(result.Segments) != 2 {
		t.Fatalf("expected two segments, got %+v", result.Segments)
	}
	if result.EstimatedSeconds != 43 {
		t.Fatalf("expected estimated 43s, got %d", result.EstimatedSeconds)
	}
	if !almostEqual(result.Transitions[0].EstimatedSeconds, 25) {
		t.Fatalf("unexpected elevator time: %+v", result.Transitions[0])
	}
	if !almostEqual(result.Segments[0].EstimatedSeconds, 11.0/1.2) {
		t.Fatalf("unexpected segment time: %+v", result.Segments[0])
	}
}

func TestPlanFloorMaskRejectsTarget(t *testing.T) {
	planner := NewPlanner(newTestStore(t, true, true))
	_, err := planner.Plan(context.Background(), Request{
		From:          testEndpoint("F1", "101", 0, 0),
		To:            testEndpoint("F2", "201", 0, 0),
		AllowedFloors: graph.FloorF1,
	})
	if !errors.Is(err, ErrFloorUnavailable) {
		t.Fatalf("expected ErrFloorUnavailable, got %v", err)
	}
}

func TestPlanSameRoom(t *testing.T) {
	planner := NewPlanner(newTestStore(t, true, true))
	result, err := planner.Plan(context.Background(), Request{
		From: testEndpoint("F1", "101", 0, 0),
		To:   testEndpoint("F1", "101", 0, 0),
	})
	if err != nil {
		t.Fatalf("plan failed: %v", err)
	}
	if result.DistanceM != 0 || len(result.Transitions) != 0 || len(result.Segments) != 1 {
		t.Fatalf("unexpected same-room result: %+v", result)
	}
}

func TestPlanNoRouteWithoutCrossFloor(t *testing.T) {
	store := newTestStore(t, false, false)
	planner := NewPlanner(store)
	_, err := planner.Plan(context.Background(), Request{
		From: testEndpoint("F1", "101", 0, 0),
		To:   testEndpoint("F2", "201", 0, 0),
	})
	if !errors.Is(err, ErrNoRoute) {
		t.Fatalf("expected ErrNoRoute, got %v", err)
	}
}

func TestPlanElevatorOnlyWithoutElevatorEdge(t *testing.T) {
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
		t.Fatalf("构建测试图失败: %v", err)
	}
	_, err = NewPlanner(store).Plan(context.Background(), Request{
		From: testEndpoint("F1", "101", 0, 0),
		To:   testEndpoint("F2", "201", 0, 0),
		Mode: ModeElevatorOnly,
	})
	if !errors.Is(err, ErrNoRoute) {
		t.Fatalf("expected ErrNoRoute when stairs excluded, got %v", err)
	}
}

func TestCrossFloorFactor(t *testing.T) {
	cases := []struct {
		floors int
		want   float64
	}{
		{1, 1},
		{2, 1.6},
		{3, 2.56},
		{4, 4},
		{5, 5},
	}
	for _, item := range cases {
		got := crossFloorFactor(item.floors, 1.6)
		if !almostEqual(got, item.want) {
			t.Fatalf("crossFloorFactor(%d, 1.6) = %v, want %v", item.floors, got, item.want)
		}
		if got > float64(item.floors) {
			t.Fatalf("factor %v exceeds %d floors", got, item.floors)
		}
	}
}

func TestPlanIncludesClassroomCenter(t *testing.T) {
	planner := NewPlanner(newTestStore(t, true, true))
	result, err := planner.Plan(context.Background(), Request{
		From: testEndpoint("F1", "101", 1, 0),
		To:   testEndpoint("F2", "201", 1, 0),
	})
	if err != nil {
		t.Fatalf("plan failed: %v", err)
	}
	if !almostEqual(result.DistanceM, 26) {
		t.Fatalf("expected distance 26 including CPath, got %+v", result)
	}
	if len(result.Segments) == 0 || len(result.Segments[0].Points) == 0 {
		t.Fatalf("missing segments: %+v", result)
	}
	start := result.Segments[0].Points[0]
	if !almostEqual(start.X, 1) || !almostEqual(start.Z, 0) {
		t.Fatalf("expected room center as first point, got %+v", start)
	}
}
