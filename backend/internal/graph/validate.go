package graph

import (
	"fmt"
	"math"
	"sort"
)

// Issue 单条校验问题。
type Issue struct {
	Rule    string `json:"rule"`
	Table   string `json:"table"`
	ID      string `json:"id,omitempty"`
	Message string `json:"message"`
}

// Counts 数据规模统计。
type Counts struct {
	Floors     int `json:"floors"`
	Rooms      int `json:"rooms"`
	Facilities int `json:"facilities"`
	Nodes      int `json:"nodes"`
	Edges      int `json:"edges"`
	Classrooms int `json:"classrooms"`
}

// Report 校验结果。
type Report struct {
	Valid    bool    `json:"valid"`
	Building string  `json:"building"`
	Counts   Counts  `json:"counts"`
	Errors   []Issue `json:"errors"`
	Warnings []Issue `json:"warnings"`
}

var roomCorridors = map[string]bool{
	"north": true,
	"midA":  true,
	"midB":  true,
	"south": true,
	"west":  true,
	"east":  true,
}

var coreCodes = map[string]bool{
	"NW":  true,
	"NE":  true,
	"MAW": true,
	"MAE": true,
	"MBW": true,
	"MBE": true,
	"SW":  true,
	"SE":  true,
}

var openSides = map[string]bool{
	"north": true,
	"south": true,
	"east":  true,
	"west":  true,
}

// Validate 校验原始数据。strict 为 true 时警告也视为错误。
func Validate(raw *RawData, strict bool) Report {
	report := Report{}
	if raw == nil {
		report.Errors = []Issue{{
			Rule:    "data",
			Table:   "-",
			Message: "原始数据为空",
		}}
		return report
	}

	report.Building = raw.Building.Code
	report.Counts = Counts{
		Floors:     len(raw.Floors),
		Rooms:      len(raw.Rooms),
		Facilities: len(raw.Facilities),
		Nodes:      len(raw.Nodes),
		Edges:      len(raw.Edges),
	}

	errors := make([]Issue, 0)
	warnings := make([]Issue, 0)
	addError := func(rule, table, id, format string, args ...any) {
		errors = append(errors, Issue{
			Rule:    rule,
			Table:   table,
			ID:      id,
			Message: fmt.Sprintf(format, args...),
		})
	}
	addWarning := func(rule, table, id, format string, args ...any) {
		warnings = append(warnings, Issue{
			Rule:    rule,
			Table:   table,
			ID:      id,
			Message: fmt.Sprintf(format, args...),
		})
	}

	if raw.Building.Code == "" {
		addError("required", "buildings", "0", "楼栋 code 不能为空")
	}

	floorByID := make(map[uint64]RawFloor, len(raw.Floors))
	floorCodes := make(map[string]bool, len(raw.Floors))
	floorLevels := make(map[int]bool, len(raw.Floors))
	for _, floor := range raw.Floors {
		floorByID[floor.ID] = floor
		id := fmt.Sprintf("%d", floor.ID)

		if floor.Code == "" {
			addError("required", "floors", id, "floor_code 不能为空")
		} else {
			if _, ok := FloorBitFromCode(floor.Code); !ok {
				addError("enum", "floors", id, "未知楼层编码 %q", floor.Code)
			}
			if floor.Code != fmt.Sprintf("F%d", floor.Level) {
				addError("consistency", "floors", id, "floor_code %q 与 level %d 不一致", floor.Code, floor.Level)
			}
			if floorCodes[floor.Code] {
				addError("unique", "floors", id, "楼层编码 %q 重复", floor.Code)
			}
			floorCodes[floor.Code] = true
		}

		if floor.Level < 1 || floor.Level > 5 {
			addError("enum", "floors", id, "level 超出 1–5：%d", floor.Level)
		}
		if floorLevels[floor.Level] {
			addError("unique", "floors", id, "楼层 level %d 重复", floor.Level)
		}
		floorLevels[floor.Level] = true

		if !finite(floor.MinX) || !finite(floor.MaxX) || !finite(floor.MinZ) || !finite(floor.MaxZ) {
			addError("coordinate", "floors", id, "楼层边界包含非有限数值")
		} else if floor.MinX >= floor.MaxX || floor.MinZ >= floor.MaxZ {
			addError("coordinate", "floors", id, "楼层边界必须满足 min < max")
		}
		if !finite(floor.FloorHeight) || floor.FloorHeight <= 0 {
			addError("distance", "floors", id, "floor_height 必须为正数")
		}
	}

	roomKeys := make(map[string]bool, len(raw.Rooms))
	roomByKey := make(map[refKey]RawRoom, len(raw.Rooms))
	classroomCount := 0
	for _, room := range raw.Rooms {
		id := room.RoomNo
		if id == "" {
			id = fmt.Sprintf("%d", room.ID)
		}
		if _, ok := floorByID[room.FloorID]; !ok {
			addError("fk", "rooms", id, "floor_id %d 不存在", room.FloorID)
		}
		if room.RoomNo == "" {
			addError("required", "rooms", id, "room_no 不能为空")
		}
		key := fmt.Sprintf("%d|%s", room.FloorID, room.RoomNo)
		if roomKeys[key] {
			addError("unique", "rooms", id, "同层房号 %q 重复", room.RoomNo)
		}
		roomKeys[key] = true
		roomByKey[refKey{floorID: room.FloorID, refType: "room", refCode: room.RoomNo}] = room

		if _, ok := RoomTypeMask(room.Type); !ok {
			addError("enum", "rooms", id, "未知房间类型 %q", room.Type)
		}
		if !roomCorridors[room.Corridor] {
			addError("enum", "rooms", id, "未知走廊编码 %q", room.Corridor)
		}
		if !finite(room.CenterX) || !finite(room.CenterZ) || !finite(room.DoorX) || !finite(room.DoorZ) {
			addError("coordinate", "rooms", id, "房间坐标包含非有限数值")
		}
		if !finite(room.SizeX) || !finite(room.SizeZ) || room.SizeX <= 0 || room.SizeZ <= 0 {
			addError("distance", "rooms", id, "房间尺寸必须为正数")
		}
		if !openSides[room.OpenSide] {
			addError("enum", "rooms", id, "未知门朝向 %q", room.OpenSide)
		}
		if room.Type == "classroom" {
			classroomCount++
		}
	}
	report.Counts.Classrooms = classroomCount

	facilityKeys := make(map[string]bool, len(raw.Facilities))
	facilityByKey := make(map[refKey]RawFacility, len(raw.Facilities))
	for _, facility := range raw.Facilities {
		id := facility.Code
		if id == "" {
			id = fmt.Sprintf("%d", facility.ID)
		}
		if _, ok := floorByID[facility.FloorID]; !ok {
			addError("fk", "facilities", id, "floor_id %d 不存在", facility.FloorID)
		}
		if facility.Code == "" {
			addError("required", "facilities", id, "设施 code 不能为空")
		}
		key := fmt.Sprintf("%d|%s", facility.FloorID, facility.Code)
		if facilityKeys[key] {
			addError("unique", "facilities", id, "同层设施 code %q 重复", facility.Code)
		}
		facilityKeys[key] = true
		facilityByKey[refKey{floorID: facility.FloorID, refType: "facility", refCode: facility.Code}] = facility

		if _, ok := FacilityTypeMask(facility.Type); !ok {
			addError("enum", "facilities", id, "未知设施类型 %q", facility.Type)
		}
		if !roomCorridors[facility.Corridor] && !coreCodes[facility.Corridor] {
			addError("enum", "facilities", id, "未知走廊/核心筒编码 %q", facility.Corridor)
		}
		if !finite(facility.CenterX) || !finite(facility.CenterZ) || !finite(facility.DoorX) || !finite(facility.DoorZ) {
			addError("coordinate", "facilities", id, "设施坐标包含非有限数值")
		}
		if !finite(facility.SizeX) || !finite(facility.SizeZ) || facility.SizeX <= 0 || facility.SizeZ <= 0 {
			addError("distance", "facilities", id, "设施尺寸必须为正数")
		}
	}

	nodeByID := make(map[uint64]RawNode, len(raw.Nodes))
	nodeKeySeen := make(map[string]bool, len(raw.Nodes))
	doorByRef := make(map[refKey]RawNode)
	facilityNodeByRef := make(map[refKey]RawNode)
	for _, node := range raw.Nodes {
		id := node.Code
		if id == "" {
			id = fmt.Sprintf("%d", node.ID)
		}
		nodeByID[node.ID] = node
		if _, ok := floorByID[node.FloorID]; !ok {
			addError("fk", "graph_nodes", id, "floor_id %d 不存在", node.FloorID)
		}
		if node.Code == "" {
			addError("required", "graph_nodes", id, "节点 code 不能为空")
		}
		key := fmt.Sprintf("%d|%s", node.FloorID, node.Code)
		if nodeKeySeen[key] {
			addError("unique", "graph_nodes", id, "同层节点 code %q 重复", node.Code)
		}
		nodeKeySeen[key] = true

		kind, kindOK := NodeKindFromCode(node.Type)
		if !kindOK {
			addError("enum", "graph_nodes", id, "未知节点类型 %q", node.Type)
		}
		if !finite(node.X) || !finite(node.Z) {
			addError("coordinate", "graph_nodes", id, "节点坐标包含非有限数值")
		}

		refType, refCode := deref(node.RefType), deref(node.RefCode)
		if refType != "" || refCode != "" {
			if refType == "" || refCode == "" {
				addError("required", "graph_nodes", id, "ref_type 与 ref_code 必须同时填写")
				continue
			}
			ref := refKey{floorID: node.FloorID, refType: refType, refCode: refCode}
			switch refType {
			case "room":
				if kindOK && kind != NodeDoor {
					addError("consistency", "graph_nodes", id, "只有 door 节点可以引用房间")
					break
				}
				if _, ok := roomByKey[ref]; !ok {
					addError("fk", "graph_nodes", id, "door 节点引用的房间 %q 不存在", refCode)
				}
				if _, exists := doorByRef[ref]; exists {
					addError("unique", "graph_nodes", id, "同一房间存在多个 door 节点")
				}
				doorByRef[ref] = node
			case "facility":
				if _, ok := facilityByKey[ref]; !ok {
					addError("fk", "graph_nodes", id, "节点引用的设施 %q 不存在", refCode)
				}
				if _, exists := facilityNodeByRef[ref]; exists {
					addError("unique", "graph_nodes", id, "同一设施存在多个接入节点")
				}
				facilityNodeByRef[ref] = node
			default:
				addError("enum", "graph_nodes", id, "未知 ref_type %q", refType)
			}
		} else if kindOK && kind == NodeDoor {
			addError("required", "graph_nodes", id, "door 节点必须填写 ref_type 与 ref_code")
		}
	}

	// 每个房间必须有且仅有一个 door 节点。
	for _, room := range raw.Rooms {
		ref := refKey{floorID: room.FloorID, refType: "room", refCode: room.RoomNo}
		if _, ok := doorByRef[ref]; !ok {
			addError("fk", "graph_nodes", room.RoomNo, "房间 %q 缺少 door 节点", room.RoomNo)
		}
	}

	// 每个设施必须有且仅有一个接入节点。
	for _, facility := range raw.Facilities {
		ref := refKey{floorID: facility.FloorID, refType: "facility", refCode: facility.Code}
		if _, ok := facilityNodeByRef[ref]; !ok {
			addError("fk", "graph_nodes", facility.Code, "设施 %q 缺少接入节点", facility.Code)
		}
	}

	edgeKeys := make(map[string]bool, len(raw.Edges))
	pairRows := make(map[string][]RawEdge)
	for _, edge := range raw.Edges {
		id := fmt.Sprintf("%d", edge.ID)
		fromNode, fromOK := nodeByID[edge.FromNodeID]
		toNode, toOK := nodeByID[edge.ToNodeID]
		if !fromOK || !toOK {
			addError("fk", "graph_edges", id, "边的端点节点不存在")
			continue
		}
		if !finite(edge.DistanceM) || edge.DistanceM <= 0 {
			addError("distance", "graph_edges", id, "distance_m 必须为正数")
		}
		kind, kindOK := EdgeKindFromCode(edge.EdgeType)
		if !kindOK {
			addError("enum", "graph_edges", id, "未知边类型 %q", edge.EdgeType)
			continue
		}
		if edge.FromNodeID == edge.ToNodeID {
			addError("self_loop", "graph_edges", id, "不允许自环")
		}

		edgeKey := fmt.Sprintf("%d|%d|%s", edge.FromNodeID, edge.ToNodeID, edge.EdgeType)
		if edgeKeys[edgeKey] {
			addError("unique", "graph_edges", id, "重复边 %s", edgeKey)
		}
		edgeKeys[edgeKey] = true

		left, right := edge.FromNodeID, edge.ToNodeID
		if left > right {
			left, right = right, left
		}
		pairKey := fmt.Sprintf("%d|%d|%s", left, right, edge.EdgeType)
		pairRows[pairKey] = append(pairRows[pairKey], edge)

		sameFloor := fromNode.FloorID == toNode.FloorID
		switch kind {
		case EdgeCorridor:
			if !sameFloor {
				addError("consistency", "graph_edges", id, "走廊边不能跨楼层")
			}
		case EdgeStairs, EdgeElevator:
			if sameFloor {
				addError("consistency", "graph_edges", id, "楼梯/电梯边必须跨楼层")
			}
			if fromKind, ok := NodeKindFromCode(fromNode.Type); !ok || fromKind != NodeCore {
				addError("consistency", "graph_edges", id, "跨层边起点必须是 core 节点")
			}
			if toKind, ok := NodeKindFromCode(toNode.Type); !ok || toKind != NodeCore {
				addError("consistency", "graph_edges", id, "跨层边终点必须是 core 节点")
			}
		}
	}

	// 无向边只允许一条记录；出现反向重复记录即报错。
	for pairKey, rows := range pairRows {
		if len(rows) < 2 {
			continue
		}
		for _, row := range rows {
			if row.Bidirectional {
				addError("bidirectional", "graph_edges", pairKey, "无向边存在重复反向记录")
				break
			}
		}
	}

	// 弱连通性 + 孤立节点检查。
	undirected := make(map[uint64][]uint64, len(raw.Nodes))
	degree := make(map[uint64]int, len(raw.Nodes))
	for _, edge := range raw.Edges {
		if _, ok := nodeByID[edge.FromNodeID]; !ok {
			continue
		}
		if _, ok := nodeByID[edge.ToNodeID]; !ok {
			continue
		}
		undirected[edge.FromNodeID] = append(undirected[edge.FromNodeID], edge.ToNodeID)
		undirected[edge.ToNodeID] = append(undirected[edge.ToNodeID], edge.FromNodeID)
		degree[edge.FromNodeID]++
		degree[edge.ToNodeID]++
	}

	nodeIDs := make([]uint64, 0, len(raw.Nodes))
	for id := range nodeByID {
		nodeIDs = append(nodeIDs, id)
	}
	sort.Slice(nodeIDs, func(i, j int) bool { return nodeIDs[i] < nodeIDs[j] })

	if len(nodeIDs) == 0 {
		addError("connectivity", "graph_nodes", "-", "没有可用的图节点")
	} else {
		visited := make(map[uint64]bool, len(nodeIDs))
		components := make([][]uint64, 0)
		for _, start := range nodeIDs {
			if visited[start] {
				continue
			}
			component := make([]uint64, 0)
			queue := []uint64{start}
			visited[start] = true
			for len(queue) > 0 {
				current := queue[0]
				queue = queue[1:]
				component = append(component, current)
				for _, next := range undirected[current] {
					if !visited[next] {
						visited[next] = true
						queue = append(queue, next)
					}
				}
			}
			components = append(components, component)
		}
		if len(components) > 1 {
			addError("connectivity", "graph_nodes", "-", "图存在 %d 个连通分量", len(components))
			for _, component := range components {
				if len(component) == 0 {
					continue
				}
				codes := make([]string, 0, minInt(len(component), 8))
				for index, nodeID := range component {
					if index >= 8 {
						codes = append(codes, "...")
						break
					}
					codes = append(codes, nodeByID[nodeID].Code)
				}
				sort.Strings(codes)
				addWarning("connectivity", "graph_nodes", fmt.Sprintf("%d", component[0]), "连通分量：%v", codes)
			}
		}
		for _, nodeID := range nodeIDs {
			if degree[nodeID] == 0 {
				addError("isolated", "graph_nodes", nodeByID[nodeID].Code, "节点没有任何边")
			}
		}
	}

	if strict && len(warnings) > 0 {
		errors = append(errors, warnings...)
		warnings = nil
	}

	sortIssues(errors)
	sortIssues(warnings)
	report.Errors = errors
	report.Warnings = warnings
	report.Valid = len(errors) == 0
	return report
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func sortIssues(items []Issue) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Rule != items[j].Rule {
			return items[i].Rule < items[j].Rule
		}
		if items[i].Table != items[j].Table {
			return items[i].Table < items[j].Table
		}
		return items[i].ID < items[j].ID
	})
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}
