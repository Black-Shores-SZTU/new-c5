package graph

import "testing"

func TestFloorBitMapping(t *testing.T) {
	cases := []struct {
		code   string
		bit    FloorBit
		level  int
		wantOK bool
	}{
		{"F1", FloorF1, 1, true},
		{"F2", FloorF2, 2, true},
		{"F3", FloorF3, 3, true},
		{"F4", FloorF4, 4, true},
		{"F5", FloorF5, 5, true},
		{"F6", 0, 0, false},
		{"", 0, 0, false},
	}

	for _, item := range cases {
		bit, ok := FloorBitFromCode(item.code)
		if ok != item.wantOK || bit != item.bit {
			t.Fatalf("FloorBitFromCode(%q) = %v, %v", item.code, bit, ok)
		}
		if !item.wantOK {
			continue
		}
		if code, ok := FloorCode(item.bit); !ok || code != item.code {
			t.Fatalf("FloorCode(%v) = %q, %v", item.bit, code, ok)
		}
		if level, ok := FloorLevel(item.bit); !ok || level != item.level {
			t.Fatalf("FloorLevel(%v) = %d, %v", item.bit, level, ok)
		}
	}

	if _, ok := FloorCode(0); ok {
		t.Fatal("FloorCode(0) should fail")
	}
	if _, ok := FloorLevel(FloorF1 | FloorF2); ok {
		t.Fatal("combined floor bits should not map to a single level")
	}
}

func TestFloorMaskUnion(t *testing.T) {
	mask := FloorF1 | FloorF3 | FloorF5
	for _, bit := range []FloorBit{FloorF1, FloorF3, FloorF5} {
		if mask&bit == 0 {
			t.Fatalf("mask %05b missing bit %05b", mask, bit)
		}
	}
	for _, bit := range []FloorBit{FloorF2, FloorF4} {
		if mask&bit != 0 {
			t.Fatalf("mask %05b unexpectedly contains bit %05b", mask, bit)
		}
	}
}

func TestTypeMasksAndNames(t *testing.T) {
	if !(TypeClassroom | TypeDoor).Has(TypeClassroom) {
		t.Fatal("combined mask should contain classroom")
	}
	if TypeClassroom.Has(TypeDoor) {
		t.Fatal("classroom mask should not contain door")
	}

	roomCases := map[string]TypeMask{
		"classroom":    TypeClassroom,
		"teacher_rest": TypeTeacherRest,
		"auditorium":   TypeAuditorium,
	}
	for name, want := range roomCases {
		got, ok := RoomTypeMask(name)
		if !ok || got != want {
			t.Fatalf("RoomTypeMask(%q) = %v, %v", name, got, ok)
		}
	}
	if _, ok := RoomTypeMask("lab"); ok {
		t.Fatal("unknown room type should fail")
	}

	facilityCases := map[string]TypeMask{
		"elevator":   TypeElevator,
		"stairs":     TypeStairs,
		"restroom":   TypeRestroom,
		"entrance":   TypeEntrance,
		"water":      TypeWater,
		"auditorium": TypeAuditorium,
		"exit":       TypeExit,
	}
	for name, want := range facilityCases {
		got, ok := FacilityTypeMask(name)
		if !ok || got != want {
			t.Fatalf("FacilityTypeMask(%q) = %v, %v", name, got, ok)
		}
		back, ok := FacilityTypeName(got)
		if !ok || back != name {
			t.Fatalf("FacilityTypeName(%v) = %q, %v", got, back, ok)
		}
	}
	if _, ok := FacilityTypeMask("gate"); ok {
		t.Fatal("unknown facility type should fail")
	}
	if _, ok := FacilityTypeName(TypeDoor); ok {
		t.Fatal("door type should not have an API facility name")
	}
}

func TestNodeAndEdgeKindConversion(t *testing.T) {
	for _, item := range []struct {
		code string
		kind NodeKind
	}{
		{"corridor", NodeCorridor},
		{"core", NodeCore},
		{"door", NodeDoor},
	} {
		got, ok := NodeKindFromCode(item.code)
		if !ok || got != item.kind || got.String() != item.code {
			t.Fatalf("node kind %q -> %v, %v", item.code, got, ok)
		}
	}
	if _, ok := NodeKindFromCode("room"); ok {
		t.Fatal("unknown node kind should fail")
	}
	if NodeKind(99).String() != "unknown" {
		t.Fatal("unknown node kind should stringify as unknown")
	}

	for _, item := range []struct {
		code string
		kind EdgeKind
	}{
		{"corridor", EdgeCorridor},
		{"stairs", EdgeStairs},
		{"elevator", EdgeElevator},
	} {
		got, ok := EdgeKindFromCode(item.code)
		if !ok || got != item.kind || got.String() != item.code {
			t.Fatalf("edge kind %q -> %v, %v", item.code, got, ok)
		}
	}
	if _, ok := EdgeKindFromCode("ramp"); ok {
		t.Fatal("unknown edge kind should fail")
	}
	if EdgeKind(99).String() != "unknown" {
		t.Fatal("unknown edge kind should stringify as unknown")
	}
}
