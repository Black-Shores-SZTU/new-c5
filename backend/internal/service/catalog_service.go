package service

import (
	"strings"

	"indoor-navigation/backend/internal/dto"
	"indoor-navigation/backend/internal/graph"
)

// CatalogService 提供楼栋、楼层与教室目录查询。
type CatalogService struct {
	store *graph.GraphStore
}

func NewCatalogService(store *graph.GraphStore) *CatalogService {
	return &CatalogService{store: store}
}

func (s *CatalogService) Buildings() []dto.BuildingItem {
	if s == nil || s.store == nil {
		return []dto.BuildingItem{}
	}
	source := s.store.Buildings()
	items := make([]dto.BuildingItem, 0, len(source))
	for _, building := range source {
		items = append(items, dto.BuildingItem{
			Code:     building.Code,
			Name:     building.Name,
			FullName: building.FullName,
			Floors:   building.FloorCount,
		})
	}
	return items
}

func (s *CatalogService) Floors(building string) ([]dto.FloorItem, bool) {
	if s == nil || s.store == nil {
		return nil, false
	}
	building = normalizeBuilding(building)
	source := s.store.Floors(building)
	if len(source) == 0 {
		return nil, false
	}
	items := make([]dto.FloorItem, 0, len(source))
	for _, floor := range source {
		items = append(items, dto.FloorItem{
			Code:  floor.Code,
			Name:  floor.Name,
			Level: floor.Level,
			Bounds: dto.FloorBounds{
				MinX:   floor.MinX,
				MaxX:   floor.MaxX,
				MinZ:   floor.MinZ,
				MaxZ:   floor.MaxZ,
				Height: floor.Height,
			},
		})
	}
	return items, true
}

func (s *CatalogService) Classrooms(building, floor, keyword string, limit, offset int) (int, []dto.ClassroomItem) {
	if s == nil || s.store == nil {
		return 0, []dto.ClassroomItem{}
	}
	building = normalizeBuilding(building)
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	needle := strings.ToLower(strings.TrimSpace(keyword))
	matched := make([]graph.ClassroomInfo, 0)
	for _, info := range s.store.ClassroomInfos(building, floor) {
		if needle != "" &&
			!strings.Contains(strings.ToLower(info.RoomNo), needle) &&
			!strings.Contains(strings.ToLower(info.Name), needle) {
			continue
		}
		matched = append(matched, info)
	}

	total := len(matched)
	if offset >= total {
		return total, []dto.ClassroomItem{}
	}
	end := offset + limit
	if end > total {
		end = total
	}

	items := make([]dto.ClassroomItem, 0, end-offset)
	for _, info := range matched[offset:end] {
		items = append(items, dto.ClassroomItem{
			Building: info.Building,
			Floor:    info.Floor,
			RoomNo:   info.RoomNo,
			Name:     info.Name,
			Type:     info.Type,
			Center:   dto.ClassroomCenter{X: info.CenterX, Z: info.CenterZ},
		})
	}
	return total, items
}

func (s *CatalogService) FindClassroom(building, floor, roomNo string) (graph.ClassroomInfo, bool) {
	if s == nil || s.store == nil {
		return graph.ClassroomInfo{}, false
	}
	return s.store.ClassroomInfo(normalizeBuilding(building), floor, roomNo)
}

func normalizeBuilding(building string) string {
	building = strings.TrimSpace(building)
	if building == "" {
		return "C5"
	}
	return building
}
