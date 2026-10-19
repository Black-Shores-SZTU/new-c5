package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"indoor-navigation/backend/internal/cache"
	"indoor-navigation/backend/internal/dto"
	"indoor-navigation/backend/internal/graph"
	"indoor-navigation/backend/internal/route"
)

// RouteService 将 API DTO 转换为领域路线请求，并在可用时使用 Redis 缓存。
type RouteService struct {
	catalog *CatalogService
	planner *route.Planner
	cache   *cache.RouteCache
}

func NewRouteService(catalog *CatalogService, planner *route.Planner, routeCache *cache.RouteCache) *RouteService {
	return &RouteService{catalog: catalog, planner: planner, cache: routeCache}
}

func (s *RouteService) Plan(ctx context.Context, request dto.RouteRequest) (dto.RouteResponse, error) {
	if request.From == nil || request.To == nil {
		return dto.RouteResponse{}, fmt.Errorf("%w: from/to 不能为空", route.ErrInvalidRequest)
	}
	if err := validateBuilding(request.From.Building); err != nil {
		return dto.RouteResponse{}, err
	}
	if err := validateBuilding(request.To.Building); err != nil {
		return dto.RouteResponse{}, err
	}

	from, ok := s.catalog.FindClassroom(request.From.Building, request.From.Floor, request.From.RoomNo)
	if !ok {
		return dto.RouteResponse{}, fmt.Errorf("%w: %s %s", route.ErrEndpointNotFound, request.From.Floor, request.From.RoomNo)
	}
	to, ok := s.catalog.FindClassroom(request.To.Building, request.To.Floor, request.To.RoomNo)
	if !ok {
		return dto.RouteResponse{}, fmt.Errorf("%w: %s %s", route.ErrEndpointNotFound, request.To.Floor, request.To.RoomNo)
	}

	mode, err := parseRouteMode(request.Mode)
	if err != nil {
		return dto.RouteResponse{}, err
	}
	allowedFloors, err := parseFloorList(request.AllowedFloors)
	if err != nil {
		return dto.RouteResponse{}, err
	}

	compute := func(computeCtx context.Context) (dto.RouteResponse, error) {
		return s.computeRoute(computeCtx, from, to, mode, allowedFloors)
	}

	if s.cache == nil {
		return compute(ctx)
	}

	version, err := s.cache.MapVersion(ctx)
	if err != nil {
		log.Printf("route cache version unavailable, bypass: %v", err)
		return compute(ctx)
	}

	key := cache.RouteKey(
		version,
		cache.EndpointKey(from),
		cache.EndpointKey(to),
		mode.String(),
		cache.AvoidKey(allowedFloors),
	)
	response, _, err := s.cache.GetOrCompute(ctx, key, compute)
	if err != nil {
		return dto.RouteResponse{}, err
	}
	return response, nil
}

func (s *RouteService) computeRoute(
	ctx context.Context,
	from graph.ClassroomInfo,
	to graph.ClassroomInfo,
	mode route.Mode,
	allowedFloors graph.FloorBit,
) (dto.RouteResponse, error) {
	result, err := s.planner.Plan(ctx, route.Request{
		From: route.Endpoint{
			Building: from.Building,
			Floor:    from.Floor,
			RoomNo:   from.RoomNo,
			CenterX:  from.CenterX,
			CenterZ:  from.CenterZ,
		},
		To: route.Endpoint{
			Building: to.Building,
			Floor:    to.Floor,
			RoomNo:   to.RoomNo,
			CenterX:  to.CenterX,
			CenterZ:  to.CenterZ,
		},
		Mode:          mode,
		AllowedFloors: allowedFloors,
	})
	if err != nil {
		return dto.RouteResponse{}, err
	}

	segments := make([]dto.RouteSegment, 0, len(result.Segments))
	for _, segment := range result.Segments {
		points := make([]dto.RoutePoint, 0, len(segment.Points))
		for _, point := range segment.Points {
			points = append(points, dto.RoutePoint{X: point.X, Z: point.Z})
		}
		segments = append(segments, dto.RouteSegment{
			Floor:            segment.Floor,
			Points:           points,
			DistanceM:        segment.DistanceM,
			EstimatedSeconds: segment.EstimatedSeconds,
		})
	}

	transitions := make([]dto.RouteTransition, 0, len(result.Transitions))
	for _, transition := range result.Transitions {
		transitions = append(transitions, dto.RouteTransition{
			FromFloor:        transition.FromFloor,
			ToFloor:          transition.ToFloor,
			Type:             transition.Kind.String(),
			DistanceM:        transition.DistanceM,
			EstimatedSeconds: transition.EstimatedSeconds,
		})
	}

	return dto.RouteResponse{
		Mode:             mode.String(),
		From:             dto.RoomRef{Building: from.Building, Floor: from.Floor, RoomNo: from.RoomNo},
		To:               dto.RoomRef{Building: to.Building, Floor: to.Floor, RoomNo: to.RoomNo},
		DistanceM:        result.DistanceM,
		EstimatedSeconds: result.EstimatedSeconds,
		Segments:         segments,
		Transitions:      transitions,
	}, nil
}

func validateBuilding(building string) error {
	if strings.TrimSpace(building) != "C5" {
		return fmt.Errorf("%w: MVP 仅支持 C5", route.ErrInvalidRequest)
	}
	return nil
}

func parseRouteMode(value string) (route.Mode, error) {
	switch strings.TrimSpace(value) {
	case "", "combined":
		return route.ModeCombined, nil
	case "elevator_only":
		return route.ModeElevatorOnly, nil
	default:
		return 0, fmt.Errorf("%w: 未知路线模式 %q", route.ErrInvalidRequest, value)
	}
}
