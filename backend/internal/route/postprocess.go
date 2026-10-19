package route

import (
	"fmt"
	"math"

	"indoor-navigation/backend/internal/graph"
)

func normalizeTimeOptions(options TimeOptions) TimeOptions {
	defaults := DefaultTimeOptions()
	if options.WalkingSpeedMPS <= 0 {
		options.WalkingSpeedMPS = defaults.WalkingSpeedMPS
	}
	if options.StairSecondsPerFloor <= 0 {
		options.StairSecondsPerFloor = defaults.StairSecondsPerFloor
	}
	if options.ElevatorWaitSeconds <= 0 {
		options.ElevatorWaitSeconds = defaults.ElevatorWaitSeconds
	}
	if options.ElevatorSecondsPerFloor <= 0 {
		options.ElevatorSecondsPerFloor = defaults.ElevatorSecondsPerFloor
	}
	return options
}

// postProcess 将节点路径按楼层切分为 Segment，去除共线点，并计算转层事件与预计时间。
func postProcess(
	steps []pathStep,
	building *graph.BuildingGraph,
	over *overlay,
	options TimeOptions,
) ([]Segment, []Transition, float64, float64, error) {
	if len(steps) == 0 {
		return nil, nil, 0, 0, nil
	}
	options = normalizeTimeOptions(options)

	firstPoint, ok := nodePoint(building, over, steps[0].From)
	if !ok {
		return nil, nil, 0, 0, fmt.Errorf("%w: 起点节点不存在", ErrNoRoute)
	}
	currentFloor, ok := nodeFloor(building, over, steps[0].From)
	if !ok {
		return nil, nil, 0, 0, fmt.Errorf("%w: 起点楼层不可用", ErrNoRoute)
	}
	currentCode, ok := graph.FloorCode(currentFloor)
	if !ok {
		return nil, nil, 0, 0, fmt.Errorf("%w: 起点楼层无效", ErrNoRoute)
	}

	segments := make([]Segment, 0)
	transitions := make([]Transition, 0)
	current := Segment{Floor: currentCode, Points: []Point{firstPoint}}
	currentDistance := 0.0
	totalDistance := 0.0
	totalSeconds := 0.0

	closeSegment := func() {
		current.Points = removeCollinear(current.Points)
		current.DistanceM = currentDistance
		current.EstimatedSeconds = currentDistance / options.WalkingSpeedMPS
		if len(current.Points) >= 2 {
			segments = append(segments, current)
			totalSeconds += current.EstimatedSeconds
		}
	}

	for _, step := range steps {
		toPoint, pointOK := nodePoint(building, over, step.To)
		if !pointOK {
			return nil, nil, 0, 0, fmt.Errorf("%w: 路径节点不存在", ErrNoRoute)
		}
		toFloor, floorOK := nodeFloor(building, over, step.To)
		if !floorOK {
			return nil, nil, 0, 0, fmt.Errorf("%w: 路径楼层不可用", ErrNoRoute)
		}

		totalDistance += step.DistanceM

		if toFloor == currentFloor {
			current.Points = append(current.Points, toPoint)
			currentDistance += step.DistanceM
			continue
		}

		closeSegment()

		toCode, codeOK := graph.FloorCode(toFloor)
		if !codeOK {
			return nil, nil, 0, 0, fmt.Errorf("%w: 目标楼层无效", ErrNoRoute)
		}
		transitionSeconds := transitionTime(step.Kind, currentFloor, toFloor, options)
		transitions = append(transitions, Transition{
			FromFloor:        currentCode,
			ToFloor:          toCode,
			Kind:             step.Kind,
			DistanceM:        step.DistanceM,
			EstimatedSeconds: transitionSeconds,
		})
		totalSeconds += transitionSeconds

		current = Segment{Floor: toCode, Points: []Point{toPoint}}
		currentFloor = toFloor
		currentCode = toCode
		currentDistance = 0
	}

	closeSegment()
	return segments, transitions, totalDistance, totalSeconds, nil
}

func transitionTime(kind graph.EdgeKind, fromFloor, toFloor graph.FloorBit, options TimeOptions) float64 {
	fromLevel, fromOK := graph.FloorLevel(fromFloor)
	toLevel, toOK := graph.FloorLevel(toFloor)
	if !fromOK || !toOK {
		return 0
	}
	floors := toLevel - fromLevel
	if floors < 0 {
		floors = -floors
	}
	switch kind {
	case graph.EdgeStairs:
		return float64(floors) * options.StairSecondsPerFloor
	case graph.EdgeElevator:
		return options.ElevatorWaitSeconds + float64(floors)*options.ElevatorSecondsPerFloor
	default:
		return 0
	}
}

func nodePoint(building *graph.BuildingGraph, over *overlay, id graph.NodeID) (Point, bool) {
	if point, ok := over.points[id]; ok {
		return point, true
	}
	node, ok := building.Node(id)
	if !ok {
		return Point{}, false
	}
	return Point{X: node.X, Z: node.Z}, true
}

func nodeFloor(building *graph.BuildingGraph, over *overlay, id graph.NodeID) (graph.FloorBit, bool) {
	if floor, ok := over.floors[id]; ok {
		return floor, true
	}
	node, ok := building.Node(id)
	if !ok {
		return 0, false
	}
	return node.Floor, true
}

func removeCollinear(points []Point) []Point {
	if len(points) <= 2 {
		return points
	}

	result := make([]Point, 0, len(points))
	result = append(result, points[0])
	for index := 1; index < len(points)-1; index++ {
		previous := result[len(result)-1]
		current := points[index]
		next := points[index+1]

		if samePoint(previous, current) {
			continue
		}

		cross := (current.X-previous.X)*(next.Z-previous.Z) - (current.Z-previous.Z)*(next.X-previous.X)
		dot := (current.X-previous.X)*(next.X-current.X) + (current.Z-previous.Z)*(next.Z-current.Z)
		if math.Abs(cross) < 1e-9 && dot >= 0 {
			continue
		}
		result = append(result, current)
	}

	last := points[len(points)-1]
	if !samePoint(result[len(result)-1], last) {
		result = append(result, last)
	}
	return result
}

func samePoint(left, right Point) bool {
	return math.Abs(left.X-right.X) < 1e-9 && math.Abs(left.Z-right.Z) < 1e-9
}
