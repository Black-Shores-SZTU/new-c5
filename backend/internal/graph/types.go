package graph

// NodeID 为楼栋内连续分配的节点编号。
type NodeID uint32

// FloorBit 楼层位掩码，F1–F5 分别占一位。
type FloorBit uint8

const (
	FloorF1 FloorBit = 1 << iota
	FloorF2
	FloorF3
	FloorF4
	FloorF5
)

// FloorBitFromCode 将 F1–F5 转为位掩码。
func FloorBitFromCode(code string) (FloorBit, bool) {
	switch code {
	case "F1":
		return FloorF1, true
	case "F2":
		return FloorF2, true
	case "F3":
		return FloorF3, true
	case "F4":
		return FloorF4, true
	case "F5":
		return FloorF5, true
	default:
		return 0, false
	}
}

// FloorCode 将楼层位掩码转为 F1–F5。
func FloorCode(bit FloorBit) (string, bool) {
	switch bit {
	case FloorF1:
		return "F1", true
	case FloorF2:
		return "F2", true
	case FloorF3:
		return "F3", true
	case FloorF4:
		return "F4", true
	case FloorF5:
		return "F5", true
	default:
		return "", false
	}
}

// FloorLevel 将楼层位掩码转为 1–5。
func FloorLevel(bit FloorBit) (int, bool) {
	switch bit {
	case FloorF1:
		return 1, true
	case FloorF2:
		return 2, true
	case FloorF3:
		return 3, true
	case FloorF4:
		return 4, true
	case FloorF5:
		return 5, true
	default:
		return 0, false
	}
}

// TypeMask 地点/节点类型位掩码。
type TypeMask uint16

const (
	TypeClassroom TypeMask = 1 << iota
	TypeTeacherRest
	TypeAuditorium
	TypeElevator
	TypeStairs
	TypeRestroom
	TypeEntrance
	TypeWater
	TypeCorridor
	TypeCore
	TypeDoor
	TypeExit
)

// Has 判断是否包含指定类型位。
func (m TypeMask) Has(bit TypeMask) bool {
	return m&bit != 0
}

// RoomTypeMask 将房间类型转为位掩码。
func RoomTypeMask(roomType string) (TypeMask, bool) {
	switch roomType {
	case "classroom":
		return TypeClassroom, true
	case "teacher_rest":
		return TypeTeacherRest, true
	case "auditorium":
		return TypeAuditorium, true
	default:
		return 0, false
	}
}

// FacilityTypeMask 将设施类型转为位掩码。
func FacilityTypeMask(facilityType string) (TypeMask, bool) {
	switch facilityType {
	case "elevator":
		return TypeElevator, true
	case "stairs":
		return TypeStairs, true
	case "restroom":
		return TypeRestroom, true
	case "entrance":
		return TypeEntrance, true
	case "water":
		return TypeWater, true
	case "auditorium":
		return TypeAuditorium, true
	case "exit":
		return TypeExit, true
	default:
		return 0, false
	}
}

// FacilityTypeName 将设施类型位掩码转为 API 名称。
func FacilityTypeName(mask TypeMask) (string, bool) {
	switch mask {
	case TypeElevator:
		return "elevator", true
	case TypeStairs:
		return "stairs", true
	case TypeRestroom:
		return "restroom", true
	case TypeEntrance:
		return "entrance", true
	case TypeWater:
		return "water", true
	case TypeAuditorium:
		return "auditorium", true
	case TypeExit:
		return "exit", true
	default:
		return "", false
	}
}

// NodeKind 图的节点类型。
type NodeKind uint8

const (
	NodeCorridor NodeKind = iota + 1
	NodeCore
	NodeDoor
)

// String 返回节点类型的可读名称。
func (k NodeKind) String() string {
	switch k {
	case NodeCorridor:
		return "corridor"
	case NodeCore:
		return "core"
	case NodeDoor:
		return "door"
	default:
		return "unknown"
	}
}

// NodeKindFromCode 将数据库节点类型转为 NodeKind。
func NodeKindFromCode(code string) (NodeKind, bool) {
	switch code {
	case "corridor":
		return NodeCorridor, true
	case "core":
		return NodeCore, true
	case "door":
		return NodeDoor, true
	default:
		return 0, false
	}
}

// EdgeKind 图的边类型。
type EdgeKind uint8

const (
	EdgeCorridor EdgeKind = iota + 1
	EdgeStairs
	EdgeElevator
)

// String 返回边类型的可读名称。
func (k EdgeKind) String() string {
	switch k {
	case EdgeCorridor:
		return "corridor"
	case EdgeStairs:
		return "stairs"
	case EdgeElevator:
		return "elevator"
	default:
		return "unknown"
	}
}

// EdgeKindFromCode 将数据库边类型转为 EdgeKind。
func EdgeKindFromCode(code string) (EdgeKind, bool) {
	switch code {
	case "corridor":
		return EdgeCorridor, true
	case "stairs":
		return EdgeStairs, true
	case "elevator":
		return EdgeElevator, true
	default:
		return 0, false
	}
}

// Node 只读图节点。
type Node struct {
	ID    NodeID
	Code  string
	Name  string
	Floor FloorBit
	Kind  NodeKind
	Types TypeMask
	X     float64
	Z     float64
}

// Edge 邻接边，DistanceM 单位为米。
type Edge struct {
	To        NodeID
	DistanceM float64
	Kind      EdgeKind
}

// ClassroomKey 教室映射的哈希键。
type ClassroomKey struct {
	Building string
	Floor    string
	RoomNo   string
}

// ClassroomAdjacency 教室到门节点及相邻关键节点的映射。
type ClassroomAdjacency struct {
	RoomNo     string
	RoomType   string
	DoorNodeID NodeID
	NodeIDs    []NodeID
}

// Facility 设施及其图接入节点。
type Facility struct {
	Code      string
	Name      string
	Type      TypeMask
	Floor     FloorBit
	FloorCode string
	NodeID    NodeID
	X         float64
	Z         float64
}

// BuildingInfo 目录接口使用的楼栋元数据。
type BuildingInfo struct {
	Code       string
	Name       string
	FullName   string
	FloorCount int
}

// FloorInfo 目录接口使用的楼层元数据。
type FloorInfo struct {
	Code   string
	Name   string
	Level  int
	MinX   float64
	MaxX   float64
	MinZ   float64
	MaxZ   float64
	Height float64
}

// ClassroomInfo 教室搜索与路线端点解析使用的元数据。
type ClassroomInfo struct {
	Building   string
	Floor      string
	FloorLevel int
	RoomNo     string
	Name       string
	Type       string
	CenterX    float64
	CenterZ    float64
}

// Stats 图规模统计。
type Stats struct {
	Buildings  int `json:"buildings"`
	Floors     int `json:"floors"`
	Rooms      int `json:"rooms"`
	Facilities int `json:"facilities"`
	Nodes      int `json:"nodes"`
	Edges      int `json:"edges"`
	Classrooms int `json:"classrooms"`
}

// RawData 为 repository 批量读取后交给 builder 的中间数据，字段与表结构一一对应。
type RawData struct {
	Building   RawBuilding
	Floors     []RawFloor
	Rooms      []RawRoom
	Facilities []RawFacility
	Nodes      []RawNode
	Edges      []RawEdge
}

type RawBuilding struct {
	ID       uint64
	Code     string
	Name     string
	FullName string
}

type RawFloor struct {
	ID          uint64
	BuildingID  uint64
	Code        string
	Name        string
	Level       int
	MinX        float64
	MaxX        float64
	MinZ        float64
	MaxZ        float64
	FloorHeight float64
}

type RawRoom struct {
	ID       uint64
	FloorID  uint64
	RoomNo   string
	Name     string
	Type     string
	Corridor string
	CenterX  float64
	CenterZ  float64
	SizeX    float64
	SizeZ    float64
	DoorX    float64
	DoorZ    float64
	OpenSide string
}

type RawFacility struct {
	ID       uint64
	FloorID  uint64
	Code     string
	Name     string
	Type     string
	Corridor string
	CenterX  float64
	CenterZ  float64
	SizeX    float64
	SizeZ    float64
	DoorX    float64
	DoorZ    float64
}

type RawNode struct {
	ID      uint64
	FloorID uint64
	Code    string
	Name    string
	Type    string
	X       float64
	Z       float64
	RefType *string
	RefCode *string
}

type RawEdge struct {
	ID            uint64
	FromNodeID    uint64
	ToNodeID      uint64
	EdgeType      string
	DistanceM     float64
	Bidirectional bool
}
