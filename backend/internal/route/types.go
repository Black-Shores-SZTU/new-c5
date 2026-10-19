package route

import (
	"errors"

	"indoor-navigation/backend/internal/graph"
)

// Mode 路线模式。
type Mode uint8

const (
	ModeCombined Mode = iota + 1
	ModeElevatorOnly
)

// DefaultPenaltyBase 跨层权重默认底数 y。
const DefaultPenaltyBase = 1.6

// String 返回模式名称。
func (m Mode) String() string {
	switch m {
	case ModeCombined:
		return "combined"
	case ModeElevatorOnly:
		return "elevator_only"
	default:
		return "unknown"
	}
}

// Endpoint 起终点教室。
type Endpoint struct {
	Building string
	Floor    string
	RoomNo   string
	CenterX  float64
	CenterZ  float64
}

// Request 路线请求。
type Request struct {
	From          Endpoint
	To            Endpoint
	Mode          Mode
	AllowedFloors graph.FloorBit
	PenaltyBase   float64
	TimeOptions   TimeOptions
}

// TimeOptions 预计时间参数。
type TimeOptions struct {
	WalkingSpeedMPS         float64
	StairSecondsPerFloor    float64
	ElevatorWaitSeconds     float64
	ElevatorSecondsPerFloor float64
}

// DefaultTimeOptions 返回默认时间参数。
func DefaultTimeOptions() TimeOptions {
	return TimeOptions{
		WalkingSpeedMPS:         1.2,
		StairSecondsPerFloor:    15,
		ElevatorWaitSeconds:     20,
		ElevatorSecondsPerFloor: 5,
	}
}

// Point 平面坐标点。
type Point struct {
	X float64
	Z float64
}

// Segment 同层折线段。
type Segment struct {
	Floor            string
	Points           []Point
	DistanceM        float64
	EstimatedSeconds float64
}

// Transition 跨层转换。
type Transition struct {
	FromFloor        string
	ToFloor          string
	Kind             graph.EdgeKind
	DistanceM        float64
	EstimatedSeconds float64
}

// Result 路线结果。
type Result struct {
	DistanceM        float64
	Cost             float64
	EstimatedSeconds int
	NodePath         []graph.NodeID
	Segments         []Segment
	Transitions      []Transition
}

var (
	ErrInvalidRequest   = errors.New("route: invalid request")
	ErrEndpointNotFound = errors.New("route: endpoint not found")
	ErrFloorUnavailable = errors.New("route: floor unavailable")
	ErrNoRoute          = errors.New("route: no route")
)
