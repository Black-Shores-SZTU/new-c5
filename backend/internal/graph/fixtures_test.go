package graph

import "fmt"

func strPtr(value string) *string {
	return &value
}

// validRaw 返回一张通过校验的两层小图，供 builder/validate 测试复用。
func validRaw() *RawData {
	return &RawData{
		Building: RawBuilding{ID: 1, Code: "C5", Name: "C5栋", FullName: "C5 公共教学楼"},
		Floors: []RawFloor{
			{ID: 1, BuildingID: 1, Code: "F1", Name: "一层", Level: 1, MinX: -20, MaxX: 20, MinZ: -10, MaxZ: 20, FloorHeight: 4},
			{ID: 2, BuildingID: 1, Code: "F2", Name: "二层", Level: 2, MinX: -20, MaxX: 20, MinZ: -10, MaxZ: 20, FloorHeight: 4},
		},
		Rooms: []RawRoom{
			{ID: 1, FloorID: 1, RoomNo: "101", Name: "101", Type: "classroom", Corridor: "north",
				CenterX: 0, CenterZ: 0, SizeX: 4, SizeZ: 4, DoorX: 0, DoorZ: 0, OpenSide: "south"},
			{ID: 2, FloorID: 2, RoomNo: "201", Name: "201", Type: "classroom", Corridor: "north",
				CenterX: 0, CenterZ: 0, SizeX: 4, SizeZ: 4, DoorX: 0, DoorZ: 0, OpenSide: "south"},
		},
		Facilities: []RawFacility{
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
		},
		Nodes: []RawNode{
			{ID: 1, FloorID: 1, Code: "F1-D101", Name: "101门", Type: "door", X: 0, Z: 0,
				RefType: strPtr("room"), RefCode: strPtr("101")},
			{ID: 2, FloorID: 1, Code: "F1-C1", Name: "F1走廊1", Type: "corridor", X: 0, Z: 2,
				RefType: strPtr("facility"), RefCode: strPtr("C5-F1-EXIT")},
			{ID: 3, FloorID: 1, Code: "F1-C2", Name: "F1走廊2", Type: "corridor", X: 0, Z: 6,
				RefType: strPtr("facility"), RefCode: strPtr("C5-F1-RESTROOM")},
			{ID: 4, FloorID: 1, Code: "F1-S", Name: "F1楼梯", Type: "core", X: 0, Z: 10,
				RefType: strPtr("facility"), RefCode: strPtr("C5-F1-STAIRS")},
			{ID: 5, FloorID: 1, Code: "F1-E", Name: "F1电梯", Type: "core", X: 9, Z: 2,
				RefType: strPtr("facility"), RefCode: strPtr("C5-F1-ELEVATOR")},
			{ID: 11, FloorID: 1, Code: "F1-W", Name: "F1开水间", Type: "door", X: -2, Z: 2,
				RefType: strPtr("facility"), RefCode: strPtr("C5-F1-WATER")},
			{ID: 6, FloorID: 2, Code: "F2-D201", Name: "201门", Type: "door", X: 0, Z: 0,
				RefType: strPtr("room"), RefCode: strPtr("201")},
			{ID: 7, FloorID: 2, Code: "F2-C1", Name: "F2走廊1", Type: "corridor", X: 0, Z: 2,
				RefType: strPtr("facility"), RefCode: strPtr("C5-F2-EXIT")},
			{ID: 8, FloorID: 2, Code: "F2-C2", Name: "F2走廊2", Type: "corridor", X: 0, Z: 6,
				RefType: strPtr("facility"), RefCode: strPtr("C5-F2-RESTROOM")},
			{ID: 9, FloorID: 2, Code: "F2-S", Name: "F2楼梯", Type: "core", X: 0, Z: 10,
				RefType: strPtr("facility"), RefCode: strPtr("C5-F2-STAIRS")},
			{ID: 10, FloorID: 2, Code: "F2-E", Name: "F2电梯", Type: "core", X: 9, Z: 2,
				RefType: strPtr("facility"), RefCode: strPtr("C5-F2-ELEVATOR")},
			{ID: 12, FloorID: 2, Code: "F2-W", Name: "F2开水间", Type: "door", X: -2, Z: 2,
				RefType: strPtr("facility"), RefCode: strPtr("C5-F2-WATER")},
		},
		Edges: []RawEdge{
			{ID: 1, FromNodeID: 1, ToNodeID: 2, EdgeType: "corridor", DistanceM: 2, Bidirectional: true},
			{ID: 2, FromNodeID: 2, ToNodeID: 3, EdgeType: "corridor", DistanceM: 4, Bidirectional: true},
			{ID: 3, FromNodeID: 3, ToNodeID: 4, EdgeType: "corridor", DistanceM: 4, Bidirectional: true},
			{ID: 4, FromNodeID: 2, ToNodeID: 5, EdgeType: "corridor", DistanceM: 9, Bidirectional: true},
			{ID: 5, FromNodeID: 6, ToNodeID: 7, EdgeType: "corridor", DistanceM: 2, Bidirectional: true},
			{ID: 6, FromNodeID: 7, ToNodeID: 8, EdgeType: "corridor", DistanceM: 4, Bidirectional: true},
			{ID: 7, FromNodeID: 8, ToNodeID: 9, EdgeType: "corridor", DistanceM: 4, Bidirectional: true},
			{ID: 8, FromNodeID: 7, ToNodeID: 10, EdgeType: "corridor", DistanceM: 9, Bidirectional: true},
			{ID: 9, FromNodeID: 4, ToNodeID: 9, EdgeType: "stairs", DistanceM: 4, Bidirectional: true},
			{ID: 10, FromNodeID: 5, ToNodeID: 10, EdgeType: "elevator", DistanceM: 4, Bidirectional: true},
			{ID: 11, FromNodeID: 11, ToNodeID: 2, EdgeType: "corridor", DistanceM: 2, Bidirectional: true},
			{ID: 12, FromNodeID: 12, ToNodeID: 7, EdgeType: "corridor", DistanceM: 2, Bidirectional: true},
		},
	}
}

func mustBuild(raw *RawData) (*GraphStore, error) {
	store, err := Build(raw)
	if err != nil {
		return nil, fmt.Errorf("build: %w", err)
	}
	return store, nil
}
