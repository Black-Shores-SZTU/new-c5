package dto

// NearbyQuery 最近设施查询参数。
type NearbyQuery struct {
	Building      string `form:"building" binding:"required" example:"C5"`
	Floor         string `form:"floor" binding:"required,oneof=F1 F2 F3 F4 F5" example:"F1"`
	RoomNo        string `form:"room_no" binding:"required" example:"101"`
	Types         string `form:"types" example:"restroom,elevator,stairs,exit,water"`
	AllowedFloors string `form:"allowed_floors" example:"F1,F2"`
}

// NearbyOrigin 最近设施起点。
type NearbyOrigin struct {
	Building string `json:"building" example:"C5"`
	Floor    string `json:"floor" example:"F1"`
	RoomNo   string `json:"room_no" example:"101"`
}

// NearbyItem 单个类型的最近设施。
type NearbyItem struct {
	Type              string  `json:"type" example:"restroom"`
	Found             bool    `json:"found" example:"true"`
	Code              string  `json:"code,omitempty" example:"C5-F1-RESTROOM"`
	Name              string  `json:"name,omitempty" example:"卫生间"`
	Floor             string  `json:"floor,omitempty" example:"F1"`
	X                 float64 `json:"x" example:"0"`
	Z                 float64 `json:"z" example:"6"`
	Hops              int     `json:"hops" example:"2"`
	StraightDistanceM float64 `json:"straight_distance_m" example:"6"`
}

// NearbyResponse 最近设施响应。
type NearbyResponse struct {
	Origin NearbyOrigin `json:"origin"`
	Items  []NearbyItem `json:"items"`
}

// NearbyResponseEnvelope Swagger 最近设施成功响应。
type NearbyResponseEnvelope struct {
	Code      int            `json:"code" example:"0"`
	Message   string         `json:"message" example:"ok"`
	Data      NearbyResponse `json:"data"`
	RequestID string         `json:"request_id" example:"8f3c2a7e"`
}
