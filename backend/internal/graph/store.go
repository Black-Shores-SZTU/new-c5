package graph

import "sort"

// BuildingGraph 只读的楼栋图。字段不导出，构建完成后不再修改。
type BuildingGraph struct {
	buildingCode string
	nodes        []Node
	adjacency    [][]Edge
	nodeByCode   map[string]NodeID
	floorMask    FloorBit
	typeMask     TypeMask
}

// GraphStore 启动时构建的只读图集合，可安全并发读取。
type GraphStore struct {
	buildingGraphs   map[string]*BuildingGraph
	classroomToNodes map[ClassroomKey]ClassroomAdjacency
	facilities       map[string][]Facility
	buildingInfos    []BuildingInfo
	floorInfos       map[string][]FloorInfo
	classroomInfos   map[string][]ClassroomInfo
	classroomByKey   map[string]ClassroomInfo
	stats            Stats
}

// Building 返回指定楼栋的只读图。
func (s *GraphStore) Building(code string) (*BuildingGraph, bool) {
	if s == nil {
		return nil, false
	}
	building, ok := s.buildingGraphs[code]
	return building, ok
}

// ClassroomNodes 返回指定教室的门节点及相邻关键节点。
func (s *GraphStore) ClassroomNodes(building, floor, roomNo string) (ClassroomAdjacency, bool) {
	if s == nil {
		return ClassroomAdjacency{}, false
	}
	value, ok := s.classroomToNodes[ClassroomKey{
		Building: building,
		Floor:    floor,
		RoomNo:   roomNo,
	}]
	return value, ok
}

// Classrooms 返回指定楼栋/楼层的教室映射；floor 为空时返回整栋楼。
func (s *GraphStore) Classrooms(building, floor string) []ClassroomAdjacency {
	if s == nil {
		return nil
	}
	result := make([]ClassroomAdjacency, 0)
	for key, value := range s.classroomToNodes {
		if key.Building != building {
			continue
		}
		if floor != "" && key.Floor != floor {
			continue
		}
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].RoomNo < result[j].RoomNo
	})
	return result
}

// Stats 返回图规模统计。
func (s *GraphStore) Stats() Stats {
	if s == nil {
		return Stats{}
	}
	return s.stats
}

// Facilities 返回指定楼栋的全部设施副本。
func (s *GraphStore) Facilities(building string) []Facility {
	if s == nil {
		return nil
	}
	source := s.facilities[building]
	result := make([]Facility, len(source))
	copy(result, source)
	return result
}

// FacilitiesByType 按类型与楼层掩码过滤设施；typeMask 或 floorMask 为 0 表示不过滤。
func (s *GraphStore) FacilitiesByType(building string, typeMask TypeMask, floorMask FloorBit) []Facility {
	if s == nil {
		return nil
	}
	source := s.facilities[building]
	result := make([]Facility, 0, len(source))
	for _, facility := range source {
		if typeMask != 0 && !facility.Type.Has(typeMask) {
			continue
		}
		if floorMask != 0 && facility.Floor&floorMask == 0 {
			continue
		}
		result = append(result, facility)
	}
	return result
}

// Buildings 返回楼栋元数据副本。
func (s *GraphStore) Buildings() []BuildingInfo {
	if s == nil {
		return nil
	}
	result := make([]BuildingInfo, len(s.buildingInfos))
	copy(result, s.buildingInfos)
	return result
}

// Floors 返回指定楼栋的楼层元数据副本。
func (s *GraphStore) Floors(building string) []FloorInfo {
	if s == nil {
		return nil
	}
	source := s.floorInfos[building]
	result := make([]FloorInfo, len(source))
	copy(result, source)
	return result
}

// ClassroomInfos 返回指定楼栋/楼层的教室元数据；floor 为空时返回整栋楼。
func (s *GraphStore) ClassroomInfos(building, floor string) []ClassroomInfo {
	if s == nil {
		return nil
	}
	source := s.classroomInfos[building]
	result := make([]ClassroomInfo, 0, len(source))
	for _, info := range source {
		if floor != "" && info.Floor != floor {
			continue
		}
		result = append(result, info)
	}
	return result
}

// ClassroomInfo 按楼栋、楼层、房号查找教室。
func (s *GraphStore) ClassroomInfo(building, floor, roomNo string) (ClassroomInfo, bool) {
	if s == nil {
		return ClassroomInfo{}, false
	}
	info, ok := s.classroomByKey[building+"|"+floor+"|"+roomNo]
	return info, ok
}

// BuildingCode 返回楼栋编码。
func (g *BuildingGraph) BuildingCode() string {
	return g.buildingCode
}

// FloorMask 返回楼栋内节点楼层位掩码的并集。
func (g *BuildingGraph) FloorMask() FloorBit {
	return g.floorMask
}

// TypeMask 返回楼栋内节点类型位掩码的并集。
func (g *BuildingGraph) TypeMask() TypeMask {
	return g.typeMask
}

// Node 按节点编号返回节点。
func (g *BuildingGraph) Node(id NodeID) (Node, bool) {
	if g == nil || int(id) >= len(g.nodes) {
		return Node{}, false
	}
	return g.nodes[id], true
}

// NodeIDByCode 按楼层与节点编码查找节点编号。
func (g *BuildingGraph) NodeIDByCode(floor, code string) (NodeID, bool) {
	if g == nil {
		return 0, false
	}
	id, ok := g.nodeByCode[floor+"|"+code]
	return id, ok
}

// Nodes 返回节点副本，调用方修改不会影响内部图。
func (g *BuildingGraph) Nodes() []Node {
	if g == nil {
		return nil
	}
	result := make([]Node, len(g.nodes))
	copy(result, g.nodes)
	return result
}

// Neighbors 返回节点出边副本，调用方修改不会影响内部邻接表。
func (g *BuildingGraph) Neighbors(id NodeID) []Edge {
	if g == nil || int(id) >= len(g.adjacency) {
		return nil
	}
	source := g.adjacency[id]
	result := make([]Edge, len(source))
	copy(result, source)
	return result
}
