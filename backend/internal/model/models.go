package model

import "time"

// Building 楼栋表。
type Building struct {
	ID        uint64    `gorm:"primaryKey"`
	Code      string    `gorm:"size:32;not null;uniqueIndex"`
	Name      string    `gorm:"size:64;not null"`
	FullName  string    `gorm:"size:128;not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (Building) TableName() string { return "buildings" }

// Floor 楼层表。
type Floor struct {
	ID          uint64    `gorm:"primaryKey"`
	BuildingID  uint64    `gorm:"not null;index;uniqueIndex:uk_building_level"`
	FloorCode   string    `gorm:"size:8;not null"`
	Name        string    `gorm:"size:32;not null"`
	Level       int       `gorm:"not null;uniqueIndex:uk_building_level"`
	MinX        float64   `gorm:"type:decimal(8,2);not null"`
	MaxX        float64   `gorm:"type:decimal(8,2);not null"`
	MinZ        float64   `gorm:"type:decimal(8,2);not null"`
	MaxZ        float64   `gorm:"type:decimal(8,2);not null"`
	FloorHeight float64   `gorm:"type:decimal(5,2);not null"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

func (Floor) TableName() string { return "floors" }

// Room 房间表。
type Room struct {
	ID        uint64    `gorm:"primaryKey"`
	FloorID   uint64    `gorm:"not null;index;uniqueIndex:uk_floor_room"`
	RoomNo    string    `gorm:"size:16;not null;uniqueIndex:uk_floor_room"`
	Name      string    `gorm:"size:64;not null"`
	Type      string    `gorm:"size:32;not null;index"`
	Corridor  string    `gorm:"size:16;not null"`
	CenterX   float64   `gorm:"type:decimal(8,2);not null"`
	CenterZ   float64   `gorm:"type:decimal(8,2);not null"`
	SizeX     float64   `gorm:"type:decimal(8,2);not null"`
	SizeZ     float64   `gorm:"type:decimal(8,2);not null"`
	DoorX     float64   `gorm:"type:decimal(8,2);not null"`
	DoorZ     float64   `gorm:"type:decimal(8,2);not null"`
	OpenSide  string    `gorm:"size:8;not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (Room) TableName() string { return "rooms" }

// Facility 设施表。
type Facility struct {
	ID        uint64    `gorm:"primaryKey"`
	FloorID   uint64    `gorm:"not null;index;uniqueIndex:uk_floor_facility"`
	Code      string    `gorm:"size:32;not null;uniqueIndex:uk_floor_facility"`
	Name      string    `gorm:"size:64;not null"`
	Type      string    `gorm:"size:32;not null;index"`
	Corridor  string    `gorm:"size:16;not null"`
	CenterX   float64   `gorm:"type:decimal(8,2);not null"`
	CenterZ   float64   `gorm:"type:decimal(8,2);not null"`
	SizeX     float64   `gorm:"type:decimal(8,2);not null"`
	SizeZ     float64   `gorm:"type:decimal(8,2);not null"`
	DoorX     float64   `gorm:"type:decimal(8,2);not null"`
	DoorZ     float64   `gorm:"type:decimal(8,2);not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (Facility) TableName() string { return "facilities" }

// GraphNode 路径节点表。
// Door 节点通过 RefType/RefCode 关联房间或设施，用于建立 classroomToNodes。
type GraphNode struct {
	ID        uint64    `gorm:"primaryKey"`
	FloorID   uint64    `gorm:"not null;index;uniqueIndex:uk_floor_node;uniqueIndex:uk_floor_ref"`
	Code      string    `gorm:"size:32;not null;uniqueIndex:uk_floor_node"`
	Name      string    `gorm:"size:64;not null"`
	Type      string    `gorm:"size:16;not null;index"`
	X         float64   `gorm:"type:decimal(8,2);not null"`
	Z         float64   `gorm:"type:decimal(8,2);not null"`
	RefType   *string   `gorm:"size:16;uniqueIndex:uk_floor_ref"`
	RefCode   *string   `gorm:"size:32;uniqueIndex:uk_floor_ref"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (GraphNode) TableName() string { return "graph_nodes" }

// GraphEdge 路径边表。
// 无向边只存一行并标记 Bidirectional=true，内存图构建时展开为双向邻接。
type GraphEdge struct {
	ID            uint64    `gorm:"primaryKey"`
	FromNodeID    uint64    `gorm:"not null;index;uniqueIndex:uk_edge_pair"`
	ToNodeID      uint64    `gorm:"not null;index;uniqueIndex:uk_edge_pair"`
	EdgeType      string    `gorm:"size:16;not null;uniqueIndex:uk_edge_pair"`
	DistanceM     float64   `gorm:"type:decimal(8,2);not null"`
	Bidirectional bool      `gorm:"not null;default:true"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`
}

func (GraphEdge) TableName() string { return "graph_edges" }
