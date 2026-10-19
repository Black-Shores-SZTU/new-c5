package dto

// BuildingItem 楼栋列表项。
type BuildingItem struct {
	Code     string `json:"code" example:"C5"`
	Name     string `json:"name" example:"C5栋"`
	FullName string `json:"full_name" example:"C5 公共教学楼"`
	Floors   int    `json:"floors" example:"5"`
}

// BuildingsResponse 楼栋列表响应。
type BuildingsResponse struct {
	Items []BuildingItem `json:"items"`
}

// BuildingsResponseEnvelope Swagger 楼栋列表成功响应。
type BuildingsResponseEnvelope struct {
	Code      int               `json:"code" example:"0"`
	Message   string            `json:"message" example:"ok"`
	Data      BuildingsResponse `json:"data"`
	RequestID string            `json:"request_id" example:"8f3c2a7e"`
}

// FloorBounds 楼层边界。
type FloorBounds struct {
	MinX   float64 `json:"min_x" example:"-44"`
	MaxX   float64 `json:"max_x" example:"44"`
	MinZ   float64 `json:"min_z" example:"-33"`
	MaxZ   float64 `json:"max_z" example:"33"`
	Height float64 `json:"height" example:"4.2"`
}

// FloorItem 楼层列表项。
type FloorItem struct {
	Code   string      `json:"code" example:"F1"`
	Name   string      `json:"name" example:"一层"`
	Level  int         `json:"level" example:"1"`
	Bounds FloorBounds `json:"bounds"`
}

// FloorsResponse 楼层列表响应。
type FloorsResponse struct {
	Building string      `json:"building" example:"C5"`
	Items    []FloorItem `json:"items"`
}

// FloorsResponseEnvelope Swagger 楼层列表成功响应。
type FloorsResponseEnvelope struct {
	Code      int            `json:"code" example:"0"`
	Message   string         `json:"message" example:"ok"`
	Data      FloorsResponse `json:"data"`
	RequestID string         `json:"request_id" example:"8f3c2a7e"`
}

// ClassroomQuery 教室搜索参数。
type ClassroomQuery struct {
	Building string `form:"building" example:"C5"`
	Floor    string `form:"floor" binding:"omitempty,oneof=F1 F2 F3 F4 F5" example:"F1"`
	Keyword  string `form:"keyword" example:"10"`
	Limit    int    `form:"limit" binding:"omitempty,min=1,max=100" example:"20"`
	Offset   int    `form:"offset" binding:"omitempty,min=0" example:"0"`
}

// ClassroomCenter 教室中心坐标。
type ClassroomCenter struct {
	X float64 `json:"x" example:"0"`
	Z float64 `json:"z" example:"0"`
}

// ClassroomItem 教室搜索结果项。
type ClassroomItem struct {
	Building string          `json:"building" example:"C5"`
	Floor    string          `json:"floor" example:"F1"`
	RoomNo   string          `json:"room_no" example:"101"`
	Name     string          `json:"name" example:"101"`
	Type     string          `json:"type" example:"classroom"`
	Center   ClassroomCenter `json:"center"`
}

// ClassroomsResponse 教室搜索响应。
type ClassroomsResponse struct {
	Total int             `json:"total" example:"2"`
	Items []ClassroomItem `json:"items"`
}

// ClassroomsResponseEnvelope Swagger 教室搜索成功响应。
type ClassroomsResponseEnvelope struct {
	Code      int                `json:"code" example:"0"`
	Message   string             `json:"message" example:"ok"`
	Data      ClassroomsResponse `json:"data"`
	RequestID string             `json:"request_id" example:"8f3c2a7e"`
}
