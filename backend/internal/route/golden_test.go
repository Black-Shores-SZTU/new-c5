package route

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"indoor-navigation/backend/internal/graph"
)

var updateGolden = flag.Bool("update", false, "update golden files")

type goldenPoint struct {
	X float64 `json:"x"`
	Z float64 `json:"z"`
}

type goldenSegment struct {
	Floor            string        `json:"floor"`
	Points           []goldenPoint `json:"points"`
	DistanceM        float64       `json:"distance_m"`
	EstimatedSeconds float64       `json:"estimated_seconds"`
}

type goldenTransition struct {
	FromFloor        string  `json:"from_floor"`
	ToFloor          string  `json:"to_floor"`
	Kind             string  `json:"type"`
	DistanceM        float64 `json:"distance_m"`
	EstimatedSeconds float64 `json:"estimated_seconds"`
}

type goldenPlan struct {
	DistanceM        float64            `json:"distance_m"`
	EstimatedSeconds int                `json:"estimated_seconds"`
	Segments         []goldenSegment    `json:"segments"`
	Transitions      []goldenTransition `json:"transitions"`
}

type goldenNearestItem struct {
	Type              string  `json:"type"`
	Found             bool    `json:"found"`
	Code              string  `json:"code"`
	Floor             string  `json:"floor"`
	Hops              int     `json:"hops"`
	StraightDistanceM float64 `json:"straight_distance_m"`
}

type goldenNearest struct {
	Items []goldenNearestItem `json:"items"`
}

func round2(value float64) float64 {
	return float64(int64(value*100+0.5)) / 100
}

func planGolden(result Result) goldenPlan {
	segments := make([]goldenSegment, 0, len(result.Segments))
	for _, segment := range result.Segments {
		points := make([]goldenPoint, 0, len(segment.Points))
		for _, point := range segment.Points {
			points = append(points, goldenPoint{X: round2(point.X), Z: round2(point.Z)})
		}
		segments = append(segments, goldenSegment{
			Floor:            segment.Floor,
			Points:           points,
			DistanceM:        round2(segment.DistanceM),
			EstimatedSeconds: round2(segment.EstimatedSeconds),
		})
	}
	transitions := make([]goldenTransition, 0, len(result.Transitions))
	for _, transition := range result.Transitions {
		transitions = append(transitions, goldenTransition{
			FromFloor:        transition.FromFloor,
			ToFloor:          transition.ToFloor,
			Kind:             transition.Kind.String(),
			DistanceM:        round2(transition.DistanceM),
			EstimatedSeconds: round2(transition.EstimatedSeconds),
		})
	}
	return goldenPlan{
		DistanceM:        round2(result.DistanceM),
		EstimatedSeconds: result.EstimatedSeconds,
		Segments:         segments,
		Transitions:      transitions,
	}
}

func TestGoldenRouteCases(t *testing.T) {
	planner := NewPlanner(newTestStore(t, true, true))
	cases := []struct {
		name string
		req  Request
		want string
	}{
		{
			name: "combined_2floor",
			req: Request{
				From: testEndpoint("F1", "101", 0, 0),
				To:   testEndpoint("F2", "201", 0, 0),
			},
			want: "combined_2floor.json",
		},
		{
			name: "elevator_only_2floor",
			req: Request{
				From: testEndpoint("F1", "101", 0, 0),
				To:   testEndpoint("F2", "201", 0, 0),
				Mode: ModeElevatorOnly,
			},
			want: "elevator_only_2floor.json",
		},
		{
			name: "same_room",
			req: Request{
				From: testEndpoint("F1", "101", 0, 0),
				To:   testEndpoint("F1", "101", 0, 0),
			},
			want: "same_room.json",
		},
	}

	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			result, err := planner.Plan(context.Background(), item.req)
			if err != nil {
				t.Fatalf("plan failed: %v", err)
			}
			assertGolden(t, item.want, planGolden(result))
		})
	}
}

func TestGoldenNearest(t *testing.T) {
	planner := NewPlanner(newTestStore(t, true, true))
	result, err := planner.Nearest(context.Background(), NearestRequest{
		Origin: NearestOrigin{
			Building: "C5",
			Floor:    "F1",
			RoomNo:   "101",
			CenterX:  0,
			CenterZ:  0,
		},
	})
	if err != nil {
		t.Fatalf("nearest failed: %v", err)
	}

	items := make([]goldenNearestItem, 0, len(result.Items))
	for _, item := range result.Items {
		typeName, _ := graph.FacilityTypeName(item.Type)
		items = append(items, goldenNearestItem{
			Type:              typeName,
			Found:             item.Found,
			Code:              item.Code,
			Floor:             item.Floor,
			Hops:              item.Hops,
			StraightDistanceM: round2(item.StraightDistanceM),
		})
	}
	assertGolden(t, "nearest_2floor.json", goldenNearest{Items: items})
}

func assertGolden(t *testing.T, name string, value any) {
	t.Helper()
	path := filepath.Join("testdata", "golden", name)
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("marshal golden: %v", err)
	}
	payload = append(payload, '\n')

	if *updateGolden {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("create golden dir: %v", err)
		}
		if err := os.WriteFile(path, payload, 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		return
	}

	expected, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v (run go test ./internal/route -update)", name, err)
	}
	var gotValue, wantValue any
	if err := json.Unmarshal(payload, &gotValue); err != nil {
		t.Fatalf("decode golden output: %v", err)
	}
	if err := json.Unmarshal(expected, &wantValue); err != nil {
		t.Fatalf("decode golden file: %v", err)
	}
	if !reflect.DeepEqual(gotValue, wantValue) {
		t.Fatalf("golden mismatch %s\n--- got ---\n%s\n--- want ---\n%s", name, payload, expected)
	}
}
