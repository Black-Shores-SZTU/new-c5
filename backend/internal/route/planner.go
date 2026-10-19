package route

import (
	"context"
	"errors"
	"fmt"
	"math"

	"indoor-navigation/backend/internal/graph"
)

// Planner 基于只读楼栋图执行路线规划。
type Planner struct {
	store *graph.GraphStore
}

func NewPlanner(store *graph.GraphStore) *Planner {
	return &Planner{store: store}
}

type resolvedEndpoint struct {
	adj      graph.ClassroomAdjacency
	center   Point
	floorBit graph.FloorBit
	floor    string
}

// Plan 规划两个教室之间的路线。
func (p *Planner) Plan(ctx context.Context, req Request) (Result, error) {
	if p == nil || p.store == nil {
		return Result{}, ErrInvalidRequest
	}
	if ctx == nil {
		ctx = context.Background()
	}

	if req.Mode == 0 {
		req.Mode = ModeCombined
	}
	if req.Mode != ModeCombined && req.Mode != ModeElevatorOnly {
		return Result{}, fmt.Errorf("%w: 未知路线模式", ErrInvalidRequest)
	}
	if req.PenaltyBase == 0 {
		req.PenaltyBase = DefaultPenaltyBase
	}
	if req.PenaltyBase < 1 {
		return Result{}, fmt.Errorf("%w: penalty_base 必须 >= 1", ErrInvalidRequest)
	}

	buildingCode := req.From.Building
	if buildingCode == "" {
		buildingCode = req.To.Building
	}
	if buildingCode == "" {
		return Result{}, fmt.Errorf("%w: 缺少楼栋编码", ErrInvalidRequest)
	}
	if req.From.Building != "" && req.To.Building != "" && req.From.Building != req.To.Building {
		return Result{}, fmt.Errorf("%w: 起终点不属于同一楼栋", ErrInvalidRequest)
	}

	building, ok := p.store.Building(buildingCode)
	if !ok {
		return Result{}, fmt.Errorf("%w: 楼栋 %q", ErrEndpointNotFound, buildingCode)
	}

	from, err := p.resolveEndpoint(buildingCode, req.From)
	if err != nil {
		return Result{}, err
	}
	to, err := p.resolveEndpoint(buildingCode, req.To)
	if err != nil {
		return Result{}, err
	}

	allowed := req.AllowedFloors
	if allowed == 0 {
		allowed = building.FloorMask()
	}
	if from.floorBit&allowed == 0 {
		return Result{}, fmt.Errorf("%w: 起点楼层 %s 不在允许范围", ErrFloorUnavailable, from.floor)
	}
	if to.floorBit&allowed == 0 {
		return Result{}, fmt.Errorf("%w: 终点楼层 %s 不在允许范围", ErrFloorUnavailable, to.floor)
	}

	if req.From.Floor == req.To.Floor && req.From.RoomNo == req.To.RoomNo {
		return Result{
			DistanceM: 0,
			Cost:      0,
			Segments: []Segment{{
				Floor:     req.From.Floor,
				Points:    []Point{{X: from.center.X, Z: from.center.Z}},
				DistanceM: 0,
			}},
		}, nil
	}

	over, err := buildOverlay(building, from, to)
	if err != nil {
		return Result{}, err
	}

	steps, cost, err := shortestPath(ctx, building, over, over.baseNodeID, over.baseNodeID+1, searchOptions{
		allowedFloors: allowed,
		mode:          req.Mode,
		penaltyBase:   req.PenaltyBase,
	})
	if err != nil {
		if errors.Is(err, ErrNoRoute) && req.Mode == ModeElevatorOnly {
			return Result{}, fmt.Errorf("%w: elevator-only 模式无可达路线", ErrNoRoute)
		}
		return Result{}, err
	}

	segments, transitions, distance, seconds, err := postProcess(steps, building, over, req.TimeOptions)
	if err != nil {
		return Result{}, err
	}

	// NodePath 只保留真实图节点；虚拟 CPath 节点仅用于搜索，不对外暴露。
	nodePath := make([]graph.NodeID, 0, len(steps))
	for _, step := range steps {
		if step.From < over.baseNodeID {
			if len(nodePath) == 0 || nodePath[len(nodePath)-1] != step.From {
				nodePath = append(nodePath, step.From)
			}
		}
		if step.To < over.baseNodeID {
			if len(nodePath) == 0 || nodePath[len(nodePath)-1] != step.To {
				nodePath = append(nodePath, step.To)
			}
		}
	}

	return Result{
		DistanceM:        distance,
		Cost:             cost,
		EstimatedSeconds: int(math.Round(seconds)),
		NodePath:         nodePath,
		Segments:         segments,
		Transitions:      transitions,
	}, nil
}

func (p *Planner) resolveEndpoint(buildingCode string, endpoint Endpoint) (resolvedEndpoint, error) {
	if endpoint.Floor == "" || endpoint.RoomNo == "" {
		return resolvedEndpoint{}, fmt.Errorf("%w: 楼层与房号不能为空", ErrInvalidRequest)
	}
	if !finite(endpoint.CenterX) || !finite(endpoint.CenterZ) {
		return resolvedEndpoint{}, fmt.Errorf("%w: 起终点坐标非法", ErrInvalidRequest)
	}

	floorBit, ok := graph.FloorBitFromCode(endpoint.Floor)
	if !ok {
		return resolvedEndpoint{}, fmt.Errorf("%w: 未知楼层 %q", ErrInvalidRequest, endpoint.Floor)
	}

	adjacency, ok := p.store.ClassroomNodes(buildingCode, endpoint.Floor, endpoint.RoomNo)
	if !ok {
		return resolvedEndpoint{}, fmt.Errorf("%w: %s %s", ErrEndpointNotFound, endpoint.Floor, endpoint.RoomNo)
	}
	if len(adjacency.NodeIDs) == 0 {
		return resolvedEndpoint{}, fmt.Errorf("%w: 教室 %s 没有相邻节点", ErrEndpointNotFound, endpoint.RoomNo)
	}

	return resolvedEndpoint{
		adj:      adjacency,
		center:   Point{X: endpoint.CenterX, Z: endpoint.CenterZ},
		floorBit: floorBit,
		floor:    endpoint.Floor,
	}, nil
}

func buildOverlay(building *graph.BuildingGraph, from, to resolvedEndpoint) (*overlay, error) {
	base := graph.NodeID(len(building.Nodes()))
	over := &overlay{
		edges:      make(map[graph.NodeID][]weightedEdge),
		points:     make(map[graph.NodeID]Point, 2),
		floors:     make(map[graph.NodeID]graph.FloorBit, 2),
		baseNodeID: base,
		nodeCount:  int(base) + 2,
	}

	add := func(virtual graph.NodeID, endpoint resolvedEndpoint) error {
		over.points[virtual] = endpoint.center
		over.floors[virtual] = endpoint.floorBit

		// CPath 通过教室门节点接入；门节点到相邻关键节点的连接沿用原图边。
		targets := make([]graph.NodeID, 0, 1)
		if _, ok := building.Node(endpoint.adj.DoorNodeID); ok {
			targets = append(targets, endpoint.adj.DoorNodeID)
		} else {
			targets = append(targets, endpoint.adj.NodeIDs...)
		}

		added := 0
		for _, nodeID := range targets {
			node, ok := building.Node(nodeID)
			if !ok {
				continue
			}
			distance := math.Hypot(node.X-endpoint.center.X, node.Z-endpoint.center.Z)
			over.edges[virtual] = append(over.edges[virtual], weightedEdge{
				To:        nodeID,
				DistanceM: distance,
				Cost:      distance,
				Kind:      graph.EdgeCorridor,
			})
			over.edges[nodeID] = append(over.edges[nodeID], weightedEdge{
				To:        virtual,
				DistanceM: distance,
				Cost:      distance,
				Kind:      graph.EdgeCorridor,
			})
			added++
		}
		if added == 0 {
			return fmt.Errorf("%w: 教室 %s 的相邻节点均不存在", ErrEndpointNotFound, endpoint.adj.RoomNo)
		}
		return nil
	}

	if err := add(base, from); err != nil {
		return nil, err
	}
	if err := add(base+1, to); err != nil {
		return nil, err
	}
	return over, nil
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
