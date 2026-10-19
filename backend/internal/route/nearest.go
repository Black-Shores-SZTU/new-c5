package route

import (
	"context"
	"fmt"
	"math"

	"indoor-navigation/backend/internal/graph"
)

// DefaultFacilityTypes 默认查询的五类设施。
var DefaultFacilityTypes = []graph.TypeMask{
	graph.TypeRestroom,
	graph.TypeElevator,
	graph.TypeStairs,
	graph.TypeExit,
	graph.TypeWater,
}

// NearestOrigin 最近设施搜索的起点，可传教室或指定图节点。
type NearestOrigin struct {
	Building string
	Floor    string
	RoomNo   string
	NodeID   *graph.NodeID
	CenterX  float64
	CenterZ  float64
}

// NearestRequest 最近设施请求。
type NearestRequest struct {
	Origin        NearestOrigin
	Types         []graph.TypeMask
	AllowedFloors graph.FloorBit
}

// NearestItem 单个类型的最近设施结果。
type NearestItem struct {
	Found             bool
	Type              graph.TypeMask
	Code              string
	Name              string
	Floor             string
	NodeID            graph.NodeID
	X                 float64
	Z                 float64
	Hops              int
	StraightDistanceM float64
}

// NearestResult 最近设施结果集，按请求类型顺序返回。
type NearestResult struct {
	Items []NearestItem
}

// Nearest 按 BFS 层数优先、同层直线距离优先查找最近设施。
func (p *Planner) Nearest(ctx context.Context, req NearestRequest) (NearestResult, error) {
	if p == nil || p.store == nil {
		return NearestResult{}, ErrInvalidRequest
	}
	if ctx == nil {
		ctx = context.Background()
	}

	origin := req.Origin
	if origin.Building == "" {
		return NearestResult{}, fmt.Errorf("%w: 缺少楼栋编码", ErrInvalidRequest)
	}
	building, ok := p.store.Building(origin.Building)
	if !ok {
		return NearestResult{}, fmt.Errorf("%w: 楼栋 %q", ErrEndpointNotFound, origin.Building)
	}

	types := normalizeNearestTypes(req.Types)
	if len(types) == 0 {
		return NearestResult{}, fmt.Errorf("%w: 未指定设施类型", ErrInvalidRequest)
	}

	start, originX, originZ, originFloor, err := p.resolveNearestOrigin(building, origin)
	if err != nil {
		return NearestResult{}, err
	}

	allowed := req.AllowedFloors
	if allowed == 0 {
		allowed = building.FloorMask()
	}
	if originFloor&allowed == 0 {
		return NearestResult{}, fmt.Errorf("%w: 起点楼层不在允许范围", ErrFloorUnavailable)
	}

	facilityByNode := make(map[graph.NodeID][]graph.Facility)
	for _, facility := range p.store.FacilitiesByType(origin.Building, 0, allowed) {
		facilityByNode[facility.NodeID] = append(facilityByNode[facility.NodeID], facility)
	}

	pending := make(map[graph.TypeMask]bool, len(types))
	for _, facilityType := range types {
		pending[facilityType] = true
	}
	results := make(map[graph.TypeMask]NearestItem, len(types))
	visited := map[graph.NodeID]bool{start: true}
	queue := []graph.NodeID{start}
	hops := 0

	for len(queue) > 0 && len(pending) > 0 {
		select {
		case <-ctx.Done():
			return NearestResult{}, ctx.Err()
		default:
		}

		levelSize := len(queue)
		candidates := make(map[graph.TypeMask]NearestItem)

		for index := 0; index < levelSize; index++ {
			nodeID := queue[0]
			queue = queue[1:]

			for _, facility := range facilityByNode[nodeID] {
				if !pending[facility.Type] {
					continue
				}
				distance := math.Hypot(facility.X-originX, facility.Z-originZ)
				current, exists := candidates[facility.Type]
				if !exists || distance < current.StraightDistanceM {
					candidates[facility.Type] = NearestItem{
						Found:             true,
						Type:              facility.Type,
						Code:              facility.Code,
						Name:              facility.Name,
						Floor:             facility.FloorCode,
						NodeID:            facility.NodeID,
						X:                 facility.X,
						Z:                 facility.Z,
						Hops:              hops,
						StraightDistanceM: distance,
					}
				}
			}

			for _, edge := range building.Neighbors(nodeID) {
				if visited[edge.To] {
					continue
				}
				toNode, nodeOK := building.Node(edge.To)
				if !nodeOK {
					continue
				}
				if allowed != 0 && toNode.Floor&allowed == 0 {
					continue
				}
				visited[edge.To] = true
				queue = append(queue, edge.To)
			}
		}

		for facilityType, item := range candidates {
			results[facilityType] = item
			delete(pending, facilityType)
		}
		hops++
	}

	items := make([]NearestItem, 0, len(types))
	for _, facilityType := range types {
		if item, exists := results[facilityType]; exists {
			items = append(items, item)
			continue
		}
		items = append(items, NearestItem{Type: facilityType, Found: false})
	}
	return NearestResult{Items: items}, nil
}

func (p *Planner) resolveNearestOrigin(building *graph.BuildingGraph, origin NearestOrigin) (graph.NodeID, float64, float64, graph.FloorBit, error) {
	if origin.RoomNo != "" {
		if origin.Floor == "" {
			return 0, 0, 0, 0, fmt.Errorf("%w: 教室楼层不能为空", ErrInvalidRequest)
		}
		floorBit, ok := graph.FloorBitFromCode(origin.Floor)
		if !ok {
			return 0, 0, 0, 0, fmt.Errorf("%w: 未知楼层 %q", ErrInvalidRequest, origin.Floor)
		}
		adjacency, ok := p.store.ClassroomNodes(origin.Building, origin.Floor, origin.RoomNo)
		if !ok {
			return 0, 0, 0, 0, fmt.Errorf("%w: %s %s", ErrEndpointNotFound, origin.Floor, origin.RoomNo)
		}
		return adjacency.DoorNodeID, origin.CenterX, origin.CenterZ, floorBit, nil
	}

	if origin.NodeID != nil {
		node, ok := building.Node(*origin.NodeID)
		if !ok {
			return 0, 0, 0, 0, fmt.Errorf("%w: 节点 %d", ErrEndpointNotFound, *origin.NodeID)
		}
		return node.ID, node.X, node.Z, node.Floor, nil
	}

	return 0, 0, 0, 0, fmt.Errorf("%w: 未指定搜索起点", ErrInvalidRequest)
}

func normalizeNearestTypes(types []graph.TypeMask) []graph.TypeMask {
	source := types
	if len(source) == 0 {
		source = DefaultFacilityTypes
	}
	seen := make(map[graph.TypeMask]bool, len(source))
	result := make([]graph.TypeMask, 0, len(source))
	for _, facilityType := range source {
		if facilityType == 0 || seen[facilityType] {
			continue
		}
		seen[facilityType] = true
		result = append(result, facilityType)
	}
	return result
}
