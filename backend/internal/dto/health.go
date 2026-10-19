package dto

// HealthResponse 健康检查响应，不使用统一 envelope。
type HealthResponse struct {
	Status  string            `json:"status" example:"ok"`
	Service string            `json:"service" example:"c5-indoor-navigation-backend"`
	Version string            `json:"version" example:"dev"`
	Checks  map[string]string `json:"checks"`
	Time    string            `json:"time" example:"2026-10-19T12:00:00Z"`
}
