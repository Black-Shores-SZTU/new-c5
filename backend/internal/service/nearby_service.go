package service

import (
	"context"
	"fmt"

	"indoor-navigation/backend/internal/dto"
	"indoor-navigation/backend/internal/graph"
	"indoor-navigation/backend/internal/route"
)

// NearbyService 处理最近设施查询。
type NearbyService struct {
	catalog *CatalogService
	planner *route.Planner
}

func NewNearbyService(catalog *CatalogService, planner *route.Planner) *NearbyService {
	return &NearbyService{catalog: catalog, planner: planner}
}

func (s *NearbyService) Nearby(ctx context.Context, query dto.NearbyQuery) (dto.NearbyResponse, error) {
	if err := validateBuilding(query.Building); err != nil {
		return dto.NearbyResponse{}, err
	}
	origin, ok := s.catalog.FindClassroom(query.Building, query.Floor, query.RoomNo)
	if !ok {
		return dto.NearbyResponse{}, fmt.Errorf("%w: %s %s", route.ErrEndpointNotFound, query.Floor, query.RoomNo)
	}

	facilityTypes, err := parseFacilityTypes(query.Types)
	if err != nil {
		return dto.NearbyResponse{}, err
	}
	allowedFloors, err := parseFloorList(splitCSV(query.AllowedFloors))
	if err != nil {
		return dto.NearbyResponse{}, err
	}

	result, err := s.planner.Nearest(ctx, route.NearestRequest{
		Origin: route.NearestOrigin{
			Building: origin.Building,
			Floor:    origin.Floor,
			RoomNo:   origin.RoomNo,
			CenterX:  origin.CenterX,
			CenterZ:  origin.CenterZ,
		},
		Types:         facilityTypes,
		AllowedFloors: allowedFloors,
	})
	if err != nil {
		return dto.NearbyResponse{}, err
	}

	items := make([]dto.NearbyItem, 0, len(result.Items))
	for _, item := range result.Items {
		typeName, _ := graph.FacilityTypeName(item.Type)
		dtoItem := dto.NearbyItem{
			Type:  typeName,
			Found: item.Found,
		}
		if item.Found {
			dtoItem.Code = item.Code
			dtoItem.Name = item.Name
			dtoItem.Floor = item.Floor
			dtoItem.X = item.X
			dtoItem.Z = item.Z
			dtoItem.Hops = item.Hops
			dtoItem.StraightDistanceM = item.StraightDistanceM
		}
		items = append(items, dtoItem)
	}

	return dto.NearbyResponse{
		Origin: dto.NearbyOrigin{
			Building: origin.Building,
			Floor:    origin.Floor,
			RoomNo:   origin.RoomNo,
		},
		Items: items,
	}, nil
}
