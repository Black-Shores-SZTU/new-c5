package graph

import "testing"

func hasIssue(issues []Issue, rule string) bool {
	for _, issue := range issues {
		if issue.Rule == rule {
			return true
		}
	}
	return false
}

func TestValidateValidData(t *testing.T) {
	report := Validate(validRaw(), true)
	if !report.Valid {
		t.Fatalf("expected valid report, got %+v", report.Errors)
	}
	if report.Building != "C5" {
		t.Fatalf("unexpected building: %s", report.Building)
	}
	want := Counts{Floors: 2, Rooms: 2, Facilities: 10, Nodes: 12, Edges: 12, Classrooms: 2}
	if report.Counts != want {
		t.Fatalf("unexpected counts: %+v", report.Counts)
	}
}

func TestValidateNilData(t *testing.T) {
	report := Validate(nil, true)
	if report.Valid || !hasIssue(report.Errors, "data") {
		t.Fatalf("unexpected nil report: %+v", report)
	}
}

func TestValidateRules(t *testing.T) {
	cases := []struct {
		name   string
		rule   string
		mutate func(*RawData)
	}{
		{
			name: "building code required",
			rule: "required",
			mutate: func(raw *RawData) {
				raw.Building.Code = ""
			},
		},
		{
			name: "unknown floor code",
			rule: "enum",
			mutate: func(raw *RawData) {
				raw.Floors[0].Code = "F9"
			},
		},
		{
			name: "floor code level mismatch",
			rule: "consistency",
			mutate: func(raw *RawData) {
				raw.Floors[0].Level = 3
			},
		},
		{
			name: "duplicate floor code",
			rule: "unique",
			mutate: func(raw *RawData) {
				raw.Floors = append(raw.Floors, RawFloor{
					ID: 3, BuildingID: 1, Code: "F1", Name: "重复一层", Level: 3,
					MinX: -20, MaxX: 20, MinZ: -10, MaxZ: 20, FloorHeight: 4,
				})
			},
		},
		{
			name: "room floor foreign key",
			rule: "fk",
			mutate: func(raw *RawData) {
				raw.Rooms[0].FloorID = 99
			},
		},
		{
			name: "unknown room type",
			rule: "enum",
			mutate: func(raw *RawData) {
				raw.Rooms[0].Type = "lab"
			},
		},
		{
			name: "unknown facility type",
			rule: "enum",
			mutate: func(raw *RawData) {
				raw.Facilities[0].Type = "gate"
			},
		},
		{
			name: "duplicate facility code",
			rule: "unique",
			mutate: func(raw *RawData) {
				duplicate := raw.Facilities[0]
				duplicate.ID = 11
				raw.Facilities = append(raw.Facilities, duplicate)
			},
		},
		{
			name: "node floor foreign key",
			rule: "fk",
			mutate: func(raw *RawData) {
				raw.Nodes[0].FloorID = 99
			},
		},
		{
			name: "duplicate node code",
			rule: "unique",
			mutate: func(raw *RawData) {
				duplicate := raw.Nodes[0]
				duplicate.ID = 13
				raw.Nodes = append(raw.Nodes, duplicate)
			},
		},
		{
			name: "door references missing room",
			rule: "fk",
			mutate: func(raw *RawData) {
				raw.Nodes[0].RefCode = strPtr("999")
			},
		},
		{
			name: "room ref on non-door node",
			rule: "consistency",
			mutate: func(raw *RawData) {
				raw.Nodes[1].RefType = strPtr("room")
				raw.Nodes[1].RefCode = strPtr("101")
			},
		},
		{
			name: "incomplete ref fields",
			rule: "required",
			mutate: func(raw *RawData) {
				raw.Nodes[1].RefCode = nil
			},
		},
		{
			name: "facility missing access node",
			rule: "fk",
			mutate: func(raw *RawData) {
				raw.Nodes = removeNode(raw.Nodes, 11)
			},
		},
		{
			name: "room missing door node",
			rule: "fk",
			mutate: func(raw *RawData) {
				raw.Nodes = removeNode(raw.Nodes, 1)
			},
		},
		{
			name: "edge unknown node",
			rule: "fk",
			mutate: func(raw *RawData) {
				raw.Edges[0].ToNodeID = 99
			},
		},
		{
			name: "edge non-positive distance",
			rule: "distance",
			mutate: func(raw *RawData) {
				raw.Edges[0].DistanceM = 0
			},
		},
		{
			name: "self loop",
			rule: "self_loop",
			mutate: func(raw *RawData) {
				raw.Edges[0].ToNodeID = raw.Edges[0].FromNodeID
			},
		},
		{
			name: "corridor edge across floors",
			rule: "consistency",
			mutate: func(raw *RawData) {
				raw.Edges[0].ToNodeID = 6
			},
		},
		{
			name: "stairs edge on same floor",
			rule: "consistency",
			mutate: func(raw *RawData) {
				raw.Edges[8].ToNodeID = 3
			},
		},
		{
			name: "duplicate edge",
			rule: "unique",
			mutate: func(raw *RawData) {
				duplicate := raw.Edges[0]
				duplicate.ID = 13
				raw.Edges = append(raw.Edges, duplicate)
			},
		},
		{
			name: "reverse duplicate bidirectional edge",
			rule: "bidirectional",
			mutate: func(raw *RawData) {
				raw.Edges = append(raw.Edges, RawEdge{
					ID: 13, FromNodeID: 2, ToNodeID: 1, EdgeType: "corridor",
					DistanceM: 2, Bidirectional: true,
				})
			},
		},
		{
			name: "isolated node",
			rule: "isolated",
			mutate: func(raw *RawData) {
				raw.Nodes = append(raw.Nodes, RawNode{
					ID: 13, FloorID: 1, Code: "F1-ISO", Name: "孤立节点", Type: "corridor", X: 5, Z: 5,
				})
			},
		},
		{
			name: "disconnected graph",
			rule: "connectivity",
			mutate: func(raw *RawData) {
				raw.Edges = removeEdge(raw.Edges, 9)
				raw.Edges = removeEdge(raw.Edges, 10)
			},
		},
	}

	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			raw := cloneRaw(validRaw())
			item.mutate(raw)
			report := Validate(raw, false)
			if report.Valid {
				t.Fatalf("expected invalid report for %s", item.name)
			}
			if !hasIssue(report.Errors, item.rule) {
				t.Fatalf("expected rule %q, got %+v", item.rule, report.Errors)
			}
		})
	}
}

func TestValidateStrictPromotesWarnings(t *testing.T) {
	raw := cloneRaw(validRaw())
	raw.Edges = removeEdge(raw.Edges, 9)
	raw.Edges = removeEdge(raw.Edges, 10)

	report := Validate(raw, false)
	if !hasIssue(report.Errors, "connectivity") {
		t.Fatalf("expected connectivity error, got %+v", report.Errors)
	}
	if len(report.Warnings) == 0 {
		t.Fatal("expected connectivity component warnings")
	}

	strict := Validate(raw, true)
	if strict.Valid || !hasIssue(strict.Errors, "connectivity") {
		t.Fatalf("strict report should keep connectivity errors: %+v", strict.Errors)
	}
	if len(strict.Warnings) != 0 {
		t.Fatalf("strict mode should promote warnings, got %+v", strict.Warnings)
	}
}
