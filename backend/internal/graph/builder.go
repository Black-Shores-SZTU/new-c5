package graph

import (
	"fmt"
	"sort"
)

type refKey struct {
	floorID uint64
	refType string
	refCode string
}

// Build 将校验通过的原始数据构建为只读图。
// 构建完成后不再修改任何图结构，调用方可安全并发读取。
func Build(raw *RawData) (*GraphStore, error) {
	if raw == nil {
		return nil, fmt.Errorf("graph: nil raw data")
	}

	floorByID := make(map[uint64]RawFloor, len(raw.Floors))
	floorCodeByID := make(map[uint64]string, len(raw.Floors))
	floorBitByID := make(map[uint64]FloorBit, len(raw.Floors))
	for _, floor := range raw.Floors {
		bit, ok := FloorBitFromCode(floor.Code)
		if !ok {
			return nil, fmt.Errorf("graph: unknown floor code %q", floor.Code)
		}
		floorByID[floor.ID] = floor
		floorCodeByID[floor.ID] = floor.Code
		floorBitByID[floor.ID] = bit
	}

	roomByKey := make(map[refKey]RawRoom, len(raw.Rooms))
	for _, room := range raw.Rooms {
		roomByKey[refKey{floorID: room.FloorID, refType: "room", refCode: room.RoomNo}] = room
	}
	facilityByKey := make(map[refKey]RawFacility, len(raw.Facilities))
	for _, facility := range raw.Facilities {
		facilityByKey[refKey{floorID: facility.FloorID, refType: "facility", refCode: facility.Code}] = facility
	}

	// 按楼层与编码排序，保证 NodeID 稳定、可复现。
	sortedNodes := make([]RawNode, len(raw.Nodes))
	copy(sortedNodes, raw.Nodes)
	sort.Slice(sortedNodes, func(i, j int) bool {
		left, right := sortedNodes[i], sortedNodes[j]
		leftFloor := floorByID[left.FloorID]
		rightFloor := floorByID[right.FloorID]
		if leftFloor.Level != rightFloor.Level {
			return leftFloor.Level < rightFloor.Level
		}
		return left.Code < right.Code
	})

	nodeIDByDBID := make(map[uint64]NodeID, len(sortedNodes))
	nodes := make([]Node, 0, len(sortedNodes))
	nodeByCode := make(map[string]NodeID, len(sortedNodes))
	doorByRef := make(map[refKey]NodeID)
	facilityNodeByRef := make(map[refKey]NodeID)
	var floorMask FloorBit
	var typeMask TypeMask

	for index, rawNode := range sortedNodes {
		floorBit, ok := floorBitByID[rawNode.FloorID]
		if !ok {
			return nil, fmt.Errorf("graph: node %q references unknown floor %d", rawNode.Code, rawNode.FloorID)
		}
		kind, ok := NodeKindFromCode(rawNode.Type)
		if !ok {
			return nil, fmt.Errorf("graph: node %q has unknown type %q", rawNode.Code, rawNode.Type)
		}

		types := TypeMask(0)
		switch kind {
		case NodeCorridor:
			types = TypeCorridor
		case NodeCore:
			types = TypeCore
		case NodeDoor:
			types = TypeDoor
		}

		refType, refCode := deref(rawNode.RefType), deref(rawNode.RefCode)
		if refType != "" || refCode != "" {
			if refType == "" || refCode == "" {
				return nil, fmt.Errorf("graph: node %q has incomplete ref fields", rawNode.Code)
			}
			key := refKey{floorID: rawNode.FloorID, refType: refType, refCode: refCode}
			switch refType {
			case "room":
				if kind != NodeDoor {
					return nil, fmt.Errorf("graph: node %q must be a door node to reference room %q", rawNode.Code, refCode)
				}
				room, exists := roomByKey[key]
				if !exists {
					return nil, fmt.Errorf("graph: door node %q references unknown room %q", rawNode.Code, refCode)
				}
				roomMask, valid := RoomTypeMask(room.Type)
				if !valid {
					return nil, fmt.Errorf("graph: room %q has unknown type %q", room.RoomNo, room.Type)
				}
				types |= roomMask
				if _, exists := doorByRef[key]; exists {
					return nil, fmt.Errorf("graph: duplicate door node for room %q", refCode)
				}
				doorByRef[key] = NodeID(index)
			case "facility":
				facility, exists := facilityByKey[key]
				if !exists {
					return nil, fmt.Errorf("graph: node %q references unknown facility %q", rawNode.Code, refCode)
				}
				facilityMask, valid := FacilityTypeMask(facility.Type)
				if !valid {
					return nil, fmt.Errorf("graph: facility %q has unknown type %q", facility.Code, facility.Type)
				}
				types |= facilityMask
				if _, exists := facilityNodeByRef[key]; exists {
					return nil, fmt.Errorf("graph: duplicate access node for facility %q", refCode)
				}
				facilityNodeByRef[key] = NodeID(index)
			default:
				return nil, fmt.Errorf("graph: node %q has invalid ref_type %q", rawNode.Code, refType)
			}
		} else if kind == NodeDoor {
			return nil, fmt.Errorf("graph: door node %q has no ref", rawNode.Code)
		}

		nodeIDByDBID[rawNode.ID] = NodeID(index)
		nodes = append(nodes, Node{
			ID:    NodeID(index),
			Code:  rawNode.Code,
			Name:  rawNode.Name,
			Floor: floorBit,
			Kind:  kind,
			Types: types,
			X:     rawNode.X,
			Z:     rawNode.Z,
		})
		nodeByCode[floorCodeByID[rawNode.FloorID]+"|"+rawNode.Code] = NodeID(index)
		floorMask |= floorBit
		typeMask |= types
	}

	facilities := make([]Facility, 0, len(raw.Facilities))
	for _, rawFacility := range raw.Facilities {
		key := refKey{floorID: rawFacility.FloorID, refType: "facility", refCode: rawFacility.Code}
		nodeID, ok := facilityNodeByRef[key]
		if !ok {
			return nil, fmt.Errorf("graph: facility %q has no access node", rawFacility.Code)
		}
		facilityMask, valid := FacilityTypeMask(rawFacility.Type)
		if !valid {
			return nil, fmt.Errorf("graph: facility %q has unknown type %q", rawFacility.Code, rawFacility.Type)
		}
		facilities = append(facilities, Facility{
			Code:      rawFacility.Code,
			Name:      rawFacility.Name,
			Type:      facilityMask,
			Floor:     floorBitByID[rawFacility.FloorID],
			FloorCode: floorCodeByID[rawFacility.FloorID],
			NodeID:    nodeID,
			X:         rawFacility.CenterX,
			Z:         rawFacility.CenterZ,
		})
	}

	adjacency := make([][]Edge, len(nodes))
	for _, rawEdge := range raw.Edges {
		from, fromOK := nodeIDByDBID[rawEdge.FromNodeID]
		to, toOK := nodeIDByDBID[rawEdge.ToNodeID]
		if !fromOK || !toOK {
			return nil, fmt.Errorf("graph: edge %d references unknown node", rawEdge.ID)
		}
		kind, ok := EdgeKindFromCode(rawEdge.EdgeType)
		if !ok {
			return nil, fmt.Errorf("graph: edge %d has unknown type %q", rawEdge.ID, rawEdge.EdgeType)
		}

		adjacency[from] = append(adjacency[from], Edge{
			To:        to,
			DistanceM: rawEdge.DistanceM,
			Kind:      kind,
		})
		if rawEdge.Bidirectional {
			adjacency[to] = append(adjacency[to], Edge{
				To:        from,
				DistanceM: rawEdge.DistanceM,
				Kind:      kind,
			})
		}
	}
	for index := range adjacency {
		sort.Slice(adjacency[index], func(i, j int) bool {
			left, right := adjacency[index][i], adjacency[index][j]
			if left.To != right.To {
				return left.To < right.To
			}
			return left.Kind < right.Kind
		})
	}

	buildingInfos := []BuildingInfo{{
		Code:       raw.Building.Code,
		Name:       raw.Building.Name,
		FullName:   raw.Building.FullName,
		FloorCount: len(raw.Floors),
	}}

	floorInfos := make([]FloorInfo, 0, len(raw.Floors))
	for _, floor := range raw.Floors {
		floorInfos = append(floorInfos, FloorInfo{
			Code:   floor.Code,
			Name:   floor.Name,
			Level:  floor.Level,
			MinX:   floor.MinX,
			MaxX:   floor.MaxX,
			MinZ:   floor.MinZ,
			MaxZ:   floor.MaxZ,
			Height: floor.FloorHeight,
		})
	}
	sort.Slice(floorInfos, func(i, j int) bool {
		return floorInfos[i].Level < floorInfos[j].Level
	})

	classroomInfos := make([]ClassroomInfo, 0, len(raw.Rooms))
	classroomByKey := make(map[string]ClassroomInfo, len(raw.Rooms))
	for _, room := range raw.Rooms {
		floor := floorByID[room.FloorID]
		info := ClassroomInfo{
			Building:   raw.Building.Code,
			Floor:      floor.Code,
			FloorLevel: floor.Level,
			RoomNo:     room.RoomNo,
			Name:       room.Name,
			Type:       room.Type,
			CenterX:    room.CenterX,
			CenterZ:    room.CenterZ,
		}
		classroomInfos = append(classroomInfos, info)
		classroomByKey[raw.Building.Code+"|"+floor.Code+"|"+room.RoomNo] = info
	}
	sort.Slice(classroomInfos, func(i, j int) bool {
		if classroomInfos[i].FloorLevel != classroomInfos[j].FloorLevel {
			return classroomInfos[i].FloorLevel < classroomInfos[j].FloorLevel
		}
		return classroomInfos[i].RoomNo < classroomInfos[j].RoomNo
	})

	classroomToNodes := make(map[ClassroomKey]ClassroomAdjacency, len(raw.Rooms))
	classroomCount := 0
	for _, room := range raw.Rooms {
		doorID, ok := doorByRef[refKey{floorID: room.FloorID, refType: "room", refCode: room.RoomNo}]
		if !ok {
			return nil, fmt.Errorf("graph: room %q has no door node", room.RoomNo)
		}

		nodeIDs := make([]NodeID, 0, 1+len(adjacency[doorID]))
		nodeIDs = append(nodeIDs, doorID)
		for _, edge := range adjacency[doorID] {
			nodeIDs = append(nodeIDs, edge.To)
		}

		floorCode := floorCodeByID[room.FloorID]
		classroomToNodes[ClassroomKey{
			Building: raw.Building.Code,
			Floor:    floorCode,
			RoomNo:   room.RoomNo,
		}] = ClassroomAdjacency{
			RoomNo:     room.RoomNo,
			RoomType:   room.Type,
			DoorNodeID: doorID,
			NodeIDs:    dedupeSorted(nodeIDs),
		}
		if room.Type == "classroom" {
			classroomCount++
		}
	}

	buildingGraph := &BuildingGraph{
		buildingCode: raw.Building.Code,
		nodes:        nodes,
		adjacency:    adjacency,
		nodeByCode:   nodeByCode,
		floorMask:    floorMask,
		typeMask:     typeMask,
	}

	return &GraphStore{
		buildingGraphs:   map[string]*BuildingGraph{raw.Building.Code: buildingGraph},
		classroomToNodes: classroomToNodes,
		facilities:       map[string][]Facility{raw.Building.Code: facilities},
		buildingInfos:    buildingInfos,
		floorInfos:       map[string][]FloorInfo{raw.Building.Code: floorInfos},
		classroomInfos:   map[string][]ClassroomInfo{raw.Building.Code: classroomInfos},
		classroomByKey:   classroomByKey,
		stats: Stats{
			Buildings:  1,
			Floors:     len(raw.Floors),
			Rooms:      len(raw.Rooms),
			Facilities: len(raw.Facilities),
			Nodes:      len(nodes),
			Edges:      len(raw.Edges),
			Classrooms: classroomCount,
		},
	}, nil
}

func deref(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func dedupeSorted(ids []NodeID) []NodeID {
	seen := make(map[NodeID]struct{}, len(ids))
	result := make([]NodeID, 0, len(ids))
	for _, id := range ids {
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})
	return result
}
