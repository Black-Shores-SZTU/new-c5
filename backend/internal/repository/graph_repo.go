package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"indoor-navigation/backend/internal/graph"
	"indoor-navigation/backend/internal/model"
)

// GraphRepository 负责一次性批量读取构建内存图所需的六张表。
// 所有查询均为集合查询，避免逐行 N+1。
type GraphRepository struct {
	db *gorm.DB
}

func NewGraphRepository(db *gorm.DB) *GraphRepository {
	return &GraphRepository{db: db}
}

// Load 读取指定楼栋的楼层、房间、设施、节点与边。
func (r *GraphRepository) Load(ctx context.Context, buildingCode string) (*graph.RawData, error) {
	var building model.Building
	if err := r.db.WithContext(ctx).
		Where("code = ?", buildingCode).
		First(&building).Error; err != nil {
		return nil, fmt.Errorf("读取楼栋 %q 失败: %w", buildingCode, err)
	}

	var floors []model.Floor
	if err := r.db.WithContext(ctx).
		Where("building_id = ?", building.ID).
		Order("level ASC").
		Find(&floors).Error; err != nil {
		return nil, fmt.Errorf("读取楼层失败: %w", err)
	}

	floorIDs := make([]uint64, 0, len(floors))
	for _, floor := range floors {
		floorIDs = append(floorIDs, floor.ID)
	}

	var rooms []model.Room
	var facilities []model.Facility
	var nodes []model.GraphNode
	if len(floorIDs) > 0 {
		if err := r.db.WithContext(ctx).
			Where("floor_id IN ?", floorIDs).
			Order("floor_id ASC, room_no ASC").
			Find(&rooms).Error; err != nil {
			return nil, fmt.Errorf("读取房间失败: %w", err)
		}
		if err := r.db.WithContext(ctx).
			Where("floor_id IN ?", floorIDs).
			Order("floor_id ASC, code ASC").
			Find(&facilities).Error; err != nil {
			return nil, fmt.Errorf("读取设施失败: %w", err)
		}
		if err := r.db.WithContext(ctx).
			Where("floor_id IN ?", floorIDs).
			Order("floor_id ASC, code ASC").
			Find(&nodes).Error; err != nil {
			return nil, fmt.Errorf("读取图节点失败: %w", err)
		}
	}

	nodeIDs := make([]uint64, 0, len(nodes))
	for _, node := range nodes {
		nodeIDs = append(nodeIDs, node.ID)
	}

	var edges []model.GraphEdge
	if len(nodeIDs) > 0 {
		if err := r.db.WithContext(ctx).
			Where("from_node_id IN ? OR to_node_id IN ?", nodeIDs, nodeIDs).
			Order("from_node_id ASC, to_node_id ASC, edge_type ASC").
			Find(&edges).Error; err != nil {
			return nil, fmt.Errorf("读取图边失败: %w", err)
		}
	}

	raw := &graph.RawData{
		Building: graph.RawBuilding{
			ID:       building.ID,
			Code:     building.Code,
			Name:     building.Name,
			FullName: building.FullName,
		},
		Floors:     make([]graph.RawFloor, 0, len(floors)),
		Rooms:      make([]graph.RawRoom, 0, len(rooms)),
		Facilities: make([]graph.RawFacility, 0, len(facilities)),
		Nodes:      make([]graph.RawNode, 0, len(nodes)),
		Edges:      make([]graph.RawEdge, 0, len(edges)),
	}

	for _, floor := range floors {
		raw.Floors = append(raw.Floors, graph.RawFloor{
			ID:          floor.ID,
			BuildingID:  floor.BuildingID,
			Code:        floor.FloorCode,
			Name:        floor.Name,
			Level:       floor.Level,
			MinX:        floor.MinX,
			MaxX:        floor.MaxX,
			MinZ:        floor.MinZ,
			MaxZ:        floor.MaxZ,
			FloorHeight: floor.FloorHeight,
		})
	}
	for _, room := range rooms {
		raw.Rooms = append(raw.Rooms, graph.RawRoom{
			ID:       room.ID,
			FloorID:  room.FloorID,
			RoomNo:   room.RoomNo,
			Name:     room.Name,
			Type:     room.Type,
			Corridor: room.Corridor,
			CenterX:  room.CenterX,
			CenterZ:  room.CenterZ,
			SizeX:    room.SizeX,
			SizeZ:    room.SizeZ,
			DoorX:    room.DoorX,
			DoorZ:    room.DoorZ,
			OpenSide: room.OpenSide,
		})
	}
	for _, facility := range facilities {
		raw.Facilities = append(raw.Facilities, graph.RawFacility{
			ID:       facility.ID,
			FloorID:  facility.FloorID,
			Code:     facility.Code,
			Name:     facility.Name,
			Type:     facility.Type,
			Corridor: facility.Corridor,
			CenterX:  facility.CenterX,
			CenterZ:  facility.CenterZ,
			SizeX:    facility.SizeX,
			SizeZ:    facility.SizeZ,
			DoorX:    facility.DoorX,
			DoorZ:    facility.DoorZ,
		})
	}
	for _, node := range nodes {
		raw.Nodes = append(raw.Nodes, graph.RawNode{
			ID:      node.ID,
			FloorID: node.FloorID,
			Code:    node.Code,
			Name:    node.Name,
			Type:    node.Type,
			X:       node.X,
			Z:       node.Z,
			RefType: cloneStringPtr(node.RefType),
			RefCode: cloneStringPtr(node.RefCode),
		})
	}
	for _, edge := range edges {
		raw.Edges = append(raw.Edges, graph.RawEdge{
			ID:            edge.ID,
			FromNodeID:    edge.FromNodeID,
			ToNodeID:      edge.ToNodeID,
			EdgeType:      edge.EdgeType,
			DistanceM:     edge.DistanceM,
			Bidirectional: edge.Bidirectional,
		})
	}

	return raw, nil
}

func cloneStringPtr(value *string) *string {
	if value == nil || *value == "" {
		return nil
	}
	copied := *value
	return &copied
}
