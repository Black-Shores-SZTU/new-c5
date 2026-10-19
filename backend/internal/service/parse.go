package service

import (
	"fmt"
	"strings"

	"indoor-navigation/backend/internal/graph"
	"indoor-navigation/backend/internal/route"
)

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.Split(value, ",")
}

func parseFloorList(values []string) (graph.FloorBit, error) {
	var mask graph.FloorBit
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		bit, ok := graph.FloorBitFromCode(value)
		if !ok {
			return 0, fmt.Errorf("%w: 未知楼层 %q", route.ErrInvalidRequest, value)
		}
		mask |= bit
	}
	return mask, nil
}

func parseFacilityTypes(value string) ([]graph.TypeMask, error) {
	parts := splitCSV(value)
	if len(parts) == 0 {
		result := make([]graph.TypeMask, len(route.DefaultFacilityTypes))
		copy(result, route.DefaultFacilityTypes)
		return result, nil
	}

	allowed := map[string]graph.TypeMask{
		"restroom": graph.TypeRestroom,
		"elevator": graph.TypeElevator,
		"stairs":   graph.TypeStairs,
		"exit":     graph.TypeExit,
		"water":    graph.TypeWater,
	}
	seen := make(map[graph.TypeMask]bool, len(parts))
	result := make([]graph.TypeMask, 0, len(parts))
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		mask, ok := allowed[name]
		if !ok {
			return nil, fmt.Errorf("%w: 未知设施类型 %q", route.ErrInvalidRequest, name)
		}
		if seen[mask] {
			continue
		}
		seen[mask] = true
		result = append(result, mask)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("%w: 未指定设施类型", route.ErrInvalidRequest)
	}
	return result, nil
}
