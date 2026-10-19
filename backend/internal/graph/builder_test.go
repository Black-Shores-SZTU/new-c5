package graph

import "testing"

func TestBuildGraph(t *testing.T) {
	raw := validRaw()
	if report := Validate(raw, true); !report.Valid {
		t.Fatalf("fixture should be valid: %+v", report.Errors)
	}

	store, err := Build(raw)
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	stats := store.Stats()
	wantStats := Stats{Buildings: 1, Floors: 2, Rooms: 2, Facilities: 10, Nodes: 12, Edges: 12, Classrooms: 2}
	if stats != wantStats {
		t.Fatalf("unexpected stats: %+v", stats)
	}

	building, ok := store.Building("C5")
	if !ok {
		t.Fatal("building C5 missing")
	}
	if building.BuildingCode() != "C5" {
		t.Fatalf("unexpected building code: %s", building.BuildingCode())
	}
	if building.FloorMask() != FloorF1|FloorF2 {
		t.Fatalf("unexpected floor mask: %05b", building.FloorMask())
	}
	for _, mask := range []TypeMask{TypeClassroom, TypeDoor, TypeCorridor, TypeCore, TypeExit, TypeRestroom, TypeStairs, TypeElevator, TypeWater} {
		if !building.TypeMask().Has(mask) {
			t.Fatalf("type mask missing %v: %05b", mask, building.TypeMask())
		}
	}

	// 节点按楼层 level、编码升序稳定分配 NodeID。
	order := []struct {
		id    NodeID
		code  string
		floor FloorBit
	}{
		{0, "F1-C1", FloorF1},
		{1, "F1-C2", FloorF1},
		{2, "F1-D101", FloorF1},
		{3, "F1-E", FloorF1},
		{4, "F1-S", FloorF1},
		{5, "F1-W", FloorF1},
		{6, "F2-C1", FloorF2},
		{7, "F2-C2", FloorF2},
		{8, "F2-D201", FloorF2},
		{9, "F2-E", FloorF2},
		{10, "F2-S", FloorF2},
		{11, "F2-W", FloorF2},
	}
	for _, item := range order {
		node, ok := building.Node(item.id)
		if !ok || node.Code != item.code || node.Floor != item.floor {
			t.Fatalf("node %d = %+v", item.id, node)
		}
		floorCode, _ := FloorCode(item.floor)
		got, ok := building.NodeIDByCode(floorCode, item.code)
		if !ok || got != item.id {
			t.Fatalf("NodeIDByCode(%s,%s) = %d, %v", floorCode, item.code, got, ok)
		}
	}

	door, _ := building.Node(2)
	if door.Kind != NodeDoor {
		t.Fatalf("expected door node, got %v", door.Kind)
	}
	if !door.Types.Has(TypeDoor) || !door.Types.Has(TypeClassroom) {
		t.Fatalf("door types missing room type: %05b", door.Types)
	}
	exitNode, _ := building.Node(0)
	if !exitNode.Types.Has(TypeCorridor) || !exitNode.Types.Has(TypeExit) {
		t.Fatalf("corridor types missing facility type: %05b", exitNode.Types)
	}

	// 无向边在构建时展开为双向。
	neighbors := building.Neighbors(0)
	wantNeighbors := []NodeID{1, 2, 3, 5}
	if len(neighbors) != len(wantNeighbors) {
		t.Fatalf("unexpected neighbor count: %+v", neighbors)
	}
	for index, want := range wantNeighbors {
		if neighbors[index].To != want {
			t.Fatalf("neighbor[%d] = %v, want %v", index, neighbors[index].To, want)
		}
	}
	stairsNeighbors := building.Neighbors(4)
	if len(stairsNeighbors) != 2 || stairsNeighbors[0].To != 1 || stairsNeighbors[1].To != 10 {
		t.Fatalf("stairs adjacency not expanded: %+v", stairsNeighbors)
	}
	if stairsNeighbors[1].Kind != EdgeStairs {
		t.Fatalf("cross-floor edge kind = %v", stairsNeighbors[1].Kind)
	}
}

func TestBuildClassroomAndFacilityMappings(t *testing.T) {
	store, err := Build(validRaw())
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	adjacency, ok := store.ClassroomNodes("C5", "F1", "101")
	if !ok {
		t.Fatal("classroom 101 missing")
	}
	if adjacency.DoorNodeID != 2 || adjacency.RoomType != "classroom" {
		t.Fatalf("unexpected classroom adjacency: %+v", adjacency)
	}
	if len(adjacency.NodeIDs) != 2 || adjacency.NodeIDs[0] != 0 || adjacency.NodeIDs[1] != 2 {
		t.Fatalf("classroom node ids = %v", adjacency.NodeIDs)
	}

	all := store.Classrooms("C5", "")
	if len(all) != 2 || all[0].RoomNo != "101" || all[1].RoomNo != "201" {
		t.Fatalf("unexpected classrooms: %+v", all)
	}
	if floorOnly := store.Classrooms("C5", "F2"); len(floorOnly) != 1 || floorOnly[0].RoomNo != "201" {
		t.Fatalf("unexpected F2 classrooms: %+v", floorOnly)
	}

	facilities := store.Facilities("C5")
	if len(facilities) != 10 {
		t.Fatalf("unexpected facility count: %d", len(facilities))
	}
	byType := store.FacilitiesByType("C5", TypeRestroom, FloorF1)
	if len(byType) != 1 || byType[0].Code != "C5-F1-RESTROOM" {
		t.Fatalf("unexpected restrooms: %+v", byType)
	}
	if byType[0].NodeID != 1 || byType[0].FloorCode != "F1" {
		t.Fatalf("unexpected restroom mapping: %+v", byType[0])
	}
	if allFloors := store.FacilitiesByType("C5", TypeExit, 0); len(allFloors) != 2 {
		t.Fatalf("unexpected exits: %+v", allFloors)
	}
}

func TestBuildCatalogMetadata(t *testing.T) {
	store, err := Build(validRaw())
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	buildings := store.Buildings()
	if len(buildings) != 1 || buildings[0].Code != "C5" || buildings[0].FloorCount != 2 {
		t.Fatalf("unexpected buildings: %+v", buildings)
	}
	floors := store.Floors("C5")
	if len(floors) != 2 || floors[0].Level != 1 || floors[1].Level != 2 {
		t.Fatalf("floors not sorted by level: %+v", floors)
	}
	infos := store.ClassroomInfos("C5", "")
	if len(infos) != 2 || infos[0].RoomNo != "101" || infos[1].RoomNo != "201" {
		t.Fatalf("unexpected classroom infos: %+v", infos)
	}
	info, ok := store.ClassroomInfo("C5", "F2", "201")
	if !ok || info.FloorLevel != 2 || info.Building != "C5" {
		t.Fatalf("unexpected classroom info: %+v", info)
	}
}

func TestStoreReturnsCopies(t *testing.T) {
	store, err := Build(validRaw())
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	building, _ := store.Building("C5")

	nodes := building.Nodes()
	nodes[0].Code = "MUTATED"
	if node, _ := building.Node(0); node.Code != "F1-C1" {
		t.Fatalf("internal nodes mutated: %+v", node)
	}

	neighbors := building.Neighbors(0)
	neighbors[0].DistanceM = 999
	if current := building.Neighbors(0)[0]; current.DistanceM == 999 {
		t.Fatal("internal adjacency mutated")
	}

	facilities := store.Facilities("C5")
	facilities[0].Name = "MUTATED"
	if current := store.Facilities("C5")[0]; current.Name == "MUTATED" {
		t.Fatal("internal facilities mutated")
	}
}

func TestBuildErrors(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*RawData)
	}{
		{
			name: "unknown floor code",
			mutate: func(raw *RawData) {
				raw.Floors[0].Code = "F9"
			},
		},
		{
			name: "door without ref",
			mutate: func(raw *RawData) {
				raw.Nodes[0].RefType = nil
				raw.Nodes[0].RefCode = nil
			},
		},
		{
			name: "facility without access node",
			mutate: func(raw *RawData) {
				raw.Nodes = removeNode(raw.Nodes, 11)
				raw.Edges = removeEdge(raw.Edges, 11)
			},
		},
		{
			name: "duplicate facility access node",
			mutate: func(raw *RawData) {
				raw.Nodes = append(raw.Nodes, RawNode{
					ID: 13, FloorID: 1, Code: "F1-W2", Name: "重复开水间", Type: "corridor", X: -2, Z: 2,
					RefType: strPtr("facility"), RefCode: strPtr("C5-F1-WATER"),
				})
			},
		},
		{
			name: "room without door",
			mutate: func(raw *RawData) {
				raw.Nodes = removeNode(raw.Nodes, 1)
				raw.Edges = removeEdge(raw.Edges, 1)
			},
		},
	}

	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			raw := cloneRaw(validRaw())
			item.mutate(raw)
			if _, err := Build(raw); err == nil {
				t.Fatalf("expected build error for %s", item.name)
			}
		})
	}

	if _, err := Build(nil); err == nil {
		t.Fatal("nil raw data should fail")
	}
}

func cloneRaw(source *RawData) *RawData {
	clone := *source
	clone.Floors = append([]RawFloor(nil), source.Floors...)
	clone.Rooms = append([]RawRoom(nil), source.Rooms...)
	clone.Facilities = append([]RawFacility(nil), source.Facilities...)
	clone.Nodes = append([]RawNode(nil), source.Nodes...)
	clone.Edges = append([]RawEdge(nil), source.Edges...)
	return &clone
}

func removeNode(nodes []RawNode, id uint64) []RawNode {
	result := make([]RawNode, 0, len(nodes))
	for _, node := range nodes {
		if node.ID != id {
			result = append(result, node)
		}
	}
	return result
}

func removeEdge(edges []RawEdge, id uint64) []RawEdge {
	result := make([]RawEdge, 0, len(edges))
	for _, edge := range edges {
		if edge.ID != id {
			result = append(result, edge)
		}
	}
	return result
}
