package dto

// RoomRef 起终点教室引用。
type RoomRef struct {
	Building string `json:"building" binding:"required" example:"C5"`
	Floor    string `json:"floor" binding:"required,oneof=F1 F2 F3 F4 F5" example:"F1"`
	RoomNo   string `json:"room_no" binding:"required" example:"101"`
}

// RouteRequest 路线请求。
type RouteRequest struct {
	From          *RoomRef `json:"from" binding:"required"`
	To            *RoomRef `json:"to" binding:"required"`
	Mode          string   `json:"mode" binding:"omitempty,oneof=combined elevator_only" example:"combined"`
	AllowedFloors []string `json:"allowed_floors" binding:"omitempty,dive,oneof=F1 F2 F3 F4 F5" example:"F1,F2,F3"`
}

// RoutePoint 折线点，单位米，原点为楼栋中心。
type RoutePoint struct {
	X float64 `json:"x" example:"0"`
	Z float64 `json:"z" example:"10"`
}

// RouteSegment 同层折线段。
type RouteSegment struct {
	Floor            string       `json:"floor" example:"F1"`
	Points           []RoutePoint `json:"points"`
	DistanceM        float64      `json:"distance_m" example:"10"`
	EstimatedSeconds float64      `json:"estimated_seconds" example:"8.33"`
}

// RouteTransition 跨层事件。
type RouteTransition struct {
	FromFloor        string  `json:"from_floor" example:"F1"`
	ToFloor          string  `json:"to_floor" example:"F2"`
	Type             string  `json:"type" example:"stairs"`
	DistanceM        float64 `json:"distance_m" example:"4"`
	EstimatedSeconds float64 `json:"estimated_seconds" example:"15"`
}

// RouteResponse 路线响应。
type RouteResponse struct {
	Mode             string            `json:"mode" example:"combined"`
	From             RoomRef           `json:"from"`
	To               RoomRef           `json:"to"`
	DistanceM        float64           `json:"distance_m" example:"24"`
	EstimatedSeconds int               `json:"estimated_seconds" example:"32"`
	Segments         []RouteSegment    `json:"segments"`
	Transitions      []RouteTransition `json:"transitions"`
}

// RouteResponseEnvelope Swagger 路线成功响应。
type RouteResponseEnvelope struct {
	Code      int           `json:"code" example:"0"`
	Message   string        `json:"message" example:"ok"`
	Data      RouteResponse `json:"data"`
	RequestID string        `json:"request_id" example:"8f3c2a7e"`
}
