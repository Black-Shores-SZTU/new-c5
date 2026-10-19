package route

import (
	"math"

	"indoor-navigation/backend/internal/graph"
)

// crossFloorFactor 计算跨层权重因子：min(y^(x-1), x)。
// x 为跨越楼层数，y 为 PenaltyBase。
func crossFloorFactor(floors int, base float64) float64 {
	if floors <= 1 {
		return 1
	}
	if base < 1 {
		base = 1
	}
	factor := math.Pow(base, float64(floors-1))
	if factor > float64(floors) {
		return float64(floors)
	}
	return factor
}

// edgeCost 返回边的加权成本；跨层边使用楼层惩罚。
func edgeCost(fromFloor, toFloor graph.FloorBit, distance float64, kind graph.EdgeKind, base float64) float64 {
	if kind == graph.EdgeCorridor || fromFloor == toFloor {
		return distance
	}
	fromLevel, fromOK := graph.FloorLevel(fromFloor)
	toLevel, toOK := graph.FloorLevel(toFloor)
	if !fromOK || !toOK {
		return distance
	}
	floors := toLevel - fromLevel
	if floors < 0 {
		floors = -floors
	}
	return distance * crossFloorFactor(floors, base)
}
