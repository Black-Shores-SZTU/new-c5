package router

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"indoor-navigation/backend/internal/dto"
	"indoor-navigation/backend/internal/graph"
	"indoor-navigation/backend/internal/handler"
	"indoor-navigation/backend/internal/route"
	"indoor-navigation/backend/internal/service"
)

func testStringPtr(value string) *string {
	return &value
}

func newTestRaw() *graph.RawData {
	return &graph.RawData{
		Building: graph.RawBuilding{ID: 1, Code: "C5", Name: "C5栋", FullName: "C5 公共教学楼"},
		Floors: []graph.RawFloor{
			{ID: 1, BuildingID: 1, Code: "F1", Name: "一层", Level: 1, MinX: -20, MaxX: 20, MinZ: -10, MaxZ: 20, FloorHeight: 4},
			{ID: 2, BuildingID: 1, Code: "F2", Name: "二层", Level: 2, MinX: -20, MaxX: 20, MinZ: -10, MaxZ: 20, FloorHeight: 4},
		},
		Rooms: []graph.RawRoom{
			{ID: 1, FloorID: 1, RoomNo: "101", Name: "101", Type: "classroom", Corridor: "north",
				CenterX: 0, CenterZ: 0, SizeX: 4, SizeZ: 4, DoorX: 0, DoorZ: 0, OpenSide: "south"},
			{ID: 2, FloorID: 2, RoomNo: "201", Name: "201", Type: "classroom", Corridor: "north",
				CenterX: 0, CenterZ: 0, SizeX: 4, SizeZ: 4, DoorX: 0, DoorZ: 0, OpenSide: "south"},
		},
		Facilities: []graph.RawFacility{
			{ID: 1, FloorID: 1, Code: "C5-F1-EXIT", Name: "安全出口", Type: "exit", Corridor: "north",
				CenterX: 0, CenterZ: 2, SizeX: 2, SizeZ: 1, DoorX: 0, DoorZ: 2},
			{ID: 2, FloorID: 1, Code: "C5-F1-RESTROOM", Name: "卫生间", Type: "restroom", Corridor: "north",
				CenterX: 0, CenterZ: 6, SizeX: 2, SizeZ: 2, DoorX: 0, DoorZ: 6},
			{ID: 3, FloorID: 1, Code: "C5-F1-STAIRS", Name: "楼梯", Type: "stairs", Corridor: "NW",
				CenterX: 0, CenterZ: 10, SizeX: 3, SizeZ: 3, DoorX: 0, DoorZ: 10},
			{ID: 4, FloorID: 1, Code: "C5-F1-ELEVATOR", Name: "电梯", Type: "elevator", Corridor: "NE",
				CenterX: 9, CenterZ: 2, SizeX: 2, SizeZ: 2, DoorX: 9, DoorZ: 2},
			{ID: 5, FloorID: 2, Code: "C5-F2-EXIT", Name: "安全出口", Type: "exit", Corridor: "north",
				CenterX: 0, CenterZ: 2, SizeX: 2, SizeZ: 1, DoorX: 0, DoorZ: 2},
			{ID: 6, FloorID: 2, Code: "C5-F2-RESTROOM", Name: "卫生间", Type: "restroom", Corridor: "north",
				CenterX: 0, CenterZ: 6, SizeX: 2, SizeZ: 2, DoorX: 0, DoorZ: 6},
			{ID: 7, FloorID: 2, Code: "C5-F2-STAIRS", Name: "楼梯", Type: "stairs", Corridor: "NW",
				CenterX: 0, CenterZ: 10, SizeX: 3, SizeZ: 3, DoorX: 0, DoorZ: 10},
			{ID: 8, FloorID: 2, Code: "C5-F2-ELEVATOR", Name: "电梯", Type: "elevator", Corridor: "NE",
				CenterX: 9, CenterZ: 2, SizeX: 2, SizeZ: 2, DoorX: 9, DoorZ: 2},
		},
		Nodes: []graph.RawNode{
			{ID: 1, FloorID: 1, Code: "F1-D101", Name: "101门", Type: "door", X: 0, Z: 0,
				RefType: testStringPtr("room"), RefCode: testStringPtr("101")},
			{ID: 2, FloorID: 1, Code: "F1-C1", Name: "F1走廊1", Type: "corridor", X: 0, Z: 2,
				RefType: testStringPtr("facility"), RefCode: testStringPtr("C5-F1-EXIT")},
			{ID: 3, FloorID: 1, Code: "F1-C2", Name: "F1走廊2", Type: "corridor", X: 0, Z: 6,
				RefType: testStringPtr("facility"), RefCode: testStringPtr("C5-F1-RESTROOM")},
			{ID: 4, FloorID: 1, Code: "F1-S", Name: "F1楼梯", Type: "core", X: 0, Z: 10,
				RefType: testStringPtr("facility"), RefCode: testStringPtr("C5-F1-STAIRS")},
			{ID: 5, FloorID: 1, Code: "F1-E", Name: "F1电梯", Type: "core", X: 9, Z: 2,
				RefType: testStringPtr("facility"), RefCode: testStringPtr("C5-F1-ELEVATOR")},
			{ID: 6, FloorID: 2, Code: "F2-D201", Name: "201门", Type: "door", X: 0, Z: 0,
				RefType: testStringPtr("room"), RefCode: testStringPtr("201")},
			{ID: 7, FloorID: 2, Code: "F2-C1", Name: "F2走廊1", Type: "corridor", X: 0, Z: 2,
				RefType: testStringPtr("facility"), RefCode: testStringPtr("C5-F2-EXIT")},
			{ID: 8, FloorID: 2, Code: "F2-C2", Name: "F2走廊2", Type: "corridor", X: 0, Z: 6,
				RefType: testStringPtr("facility"), RefCode: testStringPtr("C5-F2-RESTROOM")},
			{ID: 9, FloorID: 2, Code: "F2-S", Name: "F2楼梯", Type: "core", X: 0, Z: 10,
				RefType: testStringPtr("facility"), RefCode: testStringPtr("C5-F2-STAIRS")},
			{ID: 10, FloorID: 2, Code: "F2-E", Name: "F2电梯", Type: "core", X: 9, Z: 2,
				RefType: testStringPtr("facility"), RefCode: testStringPtr("C5-F2-ELEVATOR")},
		},
		Edges: []graph.RawEdge{
			{ID: 1, FromNodeID: 1, ToNodeID: 2, EdgeType: "corridor", DistanceM: 2, Bidirectional: true},
			{ID: 2, FromNodeID: 2, ToNodeID: 3, EdgeType: "corridor", DistanceM: 4, Bidirectional: true},
			{ID: 3, FromNodeID: 3, ToNodeID: 4, EdgeType: "corridor", DistanceM: 4, Bidirectional: true},
			{ID: 4, FromNodeID: 2, ToNodeID: 5, EdgeType: "corridor", DistanceM: 9, Bidirectional: true},
			{ID: 5, FromNodeID: 6, ToNodeID: 7, EdgeType: "corridor", DistanceM: 2, Bidirectional: true},
			{ID: 6, FromNodeID: 7, ToNodeID: 8, EdgeType: "corridor", DistanceM: 4, Bidirectional: true},
			{ID: 7, FromNodeID: 8, ToNodeID: 9, EdgeType: "corridor", DistanceM: 4, Bidirectional: true},
			{ID: 8, FromNodeID: 7, ToNodeID: 10, EdgeType: "corridor", DistanceM: 9, Bidirectional: true},
			{ID: 9, FromNodeID: 4, ToNodeID: 9, EdgeType: "stairs", DistanceM: 4, Bidirectional: true},
			{ID: 10, FromNodeID: 5, ToNodeID: 10, EdgeType: "elevator", DistanceM: 4, Bidirectional: true},
		},
	}
}

func newTestStoreFromRaw(t *testing.T, raw *graph.RawData) *graph.GraphStore {
	t.Helper()
	report := graph.Validate(raw, true)
	if !report.Valid {
		t.Fatalf("测试图数据不合法: %+v", report.Errors)
	}
	store, err := graph.Build(raw)
	if err != nil {
		t.Fatalf("构建测试图失败: %v", err)
	}
	return store
}

func newTestStore(t *testing.T) *graph.GraphStore {
	t.Helper()
	return newTestStoreFromRaw(t, newTestRaw())
}

func newTestEngine(t *testing.T) *gin.Engine {
	t.Helper()
	return newTestEngineWithRaw(t, newTestRaw(), nil)
}

func newTestEngineWithRaw(t *testing.T, raw *graph.RawData, pinger handler.Pinger) *gin.Engine {
	t.Helper()
	store := newTestStoreFromRaw(t, raw)
	catalog := service.NewCatalogService(store)
	planner := route.NewPlanner(store)
	return New(Dependencies{
		Route:   handler.NewRouteHandler(service.NewRouteService(catalog, planner, nil)),
		Nearby:  handler.NewNearbyHandler(service.NewNearbyService(catalog, planner)),
		Catalog: handler.NewCatalogHandler(catalog),
		Health:  handler.NewHealthHandler(pinger, store, "test"),
	})
}

type fakePinger struct {
	err error
}

func (f fakePinger) PingContext(context.Context) error {
	return f.err
}

func performRequest(engine *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body == "" {
		reader = bytes.NewReader(nil)
	} else {
		reader = bytes.NewReader([]byte(body))
	}
	request := httptest.NewRequest(method, path, reader)
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)
	return recorder
}

func TestCatalogEndpoints(t *testing.T) {
	engine := newTestEngine(t)

	recorder := performRequest(engine, http.MethodGet, "/buildings", "")
	if recorder.Code != 200 {
		t.Fatalf("buildings status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var buildings dto.BuildingsResponseEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &buildings); err != nil {
		t.Fatalf("decode buildings: %v", err)
	}
	if len(buildings.Data.Items) != 1 || buildings.Data.Items[0].Code != "C5" {
		t.Fatalf("unexpected buildings: %+v", buildings.Data.Items)
	}

	recorder = performRequest(engine, http.MethodGet, "/buildings/C5/floors", "")
	if recorder.Code != 200 {
		t.Fatalf("floors status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var floors dto.FloorsResponseEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &floors); err != nil {
		t.Fatalf("decode floors: %v", err)
	}
	if len(floors.Data.Items) != 2 || floors.Data.Items[0].Code != "F1" {
		t.Fatalf("unexpected floors: %+v", floors.Data.Items)
	}

	recorder = performRequest(engine, http.MethodGet, "/classrooms?keyword=101", "")
	if recorder.Code != 200 {
		t.Fatalf("classrooms status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var classrooms dto.ClassroomsResponseEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &classrooms); err != nil {
		t.Fatalf("decode classrooms: %v", err)
	}
	if classrooms.Data.Total != 1 || classrooms.Data.Items[0].RoomNo != "101" {
		t.Fatalf("unexpected classrooms: %+v", classrooms.Data)
	}
}

func TestRouteEndpoint(t *testing.T) {
	engine := newTestEngine(t)

	body := `{"from":{"building":"C5","floor":"F1","room_no":"101"},"to":{"building":"C5","floor":"F2","room_no":"201"}}`
	recorder := performRequest(engine, http.MethodPost, "/api/v1/route", body)
	if recorder.Code != 200 {
		t.Fatalf("route status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var routeResponse dto.RouteResponseEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &routeResponse); err != nil {
		t.Fatalf("decode route: %v", err)
	}
	if routeResponse.Data.DistanceM != 24 || routeResponse.Data.EstimatedSeconds != 32 {
		t.Fatalf("unexpected route: %+v", routeResponse.Data)
	}
	if len(routeResponse.Data.Transitions) != 1 || routeResponse.Data.Transitions[0].Type != "stairs" {
		t.Fatalf("unexpected transitions: %+v", routeResponse.Data.Transitions)
	}

	invalidBody := `{"from":{"building":"C5","floor":"F9","room_no":"101"},"to":{"building":"C5","floor":"F2","room_no":"201"}}`
	recorder = performRequest(engine, http.MethodPost, "/api/v1/route", invalidBody)
	if recorder.Code != 400 {
		t.Fatalf("invalid param status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var invalid dto.ErrorEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &invalid); err != nil {
		t.Fatalf("decode invalid param: %v", err)
	}
	if invalid.Code != 1001 {
		t.Fatalf("unexpected invalid param code: %+v", invalid)
	}

	missingBody := `{"from":{"building":"C5","floor":"F1","room_no":"999"},"to":{"building":"C5","floor":"F2","room_no":"201"}}`
	recorder = performRequest(engine, http.MethodPost, "/api/v1/route", missingBody)
	if recorder.Code != 404 {
		t.Fatalf("missing room status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var missing dto.ErrorEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &missing); err != nil {
		t.Fatalf("decode missing room: %v", err)
	}
	if missing.Code != 1002 {
		t.Fatalf("unexpected missing room code: %+v", missing)
	}
}

func TestNearbyEndpoint(t *testing.T) {
	engine := newTestEngine(t)
	recorder := performRequest(engine, http.MethodGet, "/api/v1/nearby?building=C5&floor=F1&room_no=101&types=restroom,exit", "")
	if recorder.Code != 200 {
		t.Fatalf("nearby status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var nearby dto.NearbyResponseEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &nearby); err != nil {
		t.Fatalf("decode nearby: %v", err)
	}
	if len(nearby.Data.Items) != 2 {
		t.Fatalf("unexpected nearby items: %+v", nearby.Data.Items)
	}
	for _, item := range nearby.Data.Items {
		if !item.Found {
			t.Fatalf("expected found facility: %+v", item)
		}
	}
}

func TestHealthEndpoint(t *testing.T) {
	engine := newTestEngine(t)
	recorder := performRequest(engine, http.MethodGet, "/health", "")
	if recorder.Code != 503 {
		t.Fatalf("health status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var health dto.HealthResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &health); err != nil {
		t.Fatalf("decode health: %v", err)
	}
	if health.Status != "degraded" || health.Checks["mysql"] != "error" {
		t.Fatalf("unexpected health: %+v", health)
	}
}

func TestSwaggerEndpoint(t *testing.T) {
	engine := newTestEngine(t)
	recorder := performRequest(engine, http.MethodGet, "/swagger/doc.json", "")
	if recorder.Code != 200 {
		t.Fatalf("swagger status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestHealthEndpointOK(t *testing.T) {
	engine := newTestEngineWithRaw(t, newTestRaw(), fakePinger{})
	recorder := performRequest(engine, http.MethodGet, "/health", "")
	if recorder.Code != 200 {
		t.Fatalf("health status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var health dto.HealthResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &health); err != nil {
		t.Fatalf("decode health: %v", err)
	}
	if health.Status != "ok" || health.Checks["mysql"] != "ok" || health.Checks["graph"] != "ok" {
		t.Fatalf("unexpected health: %+v", health)
	}
}

func TestRequestIDReuse(t *testing.T) {
	engine := newTestEngine(t)
	request := httptest.NewRequest(http.MethodGet, "/buildings", nil)
	request.Header.Set("X-Request-ID", "client-request-id")
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)

	if got := recorder.Header().Get("X-Request-ID"); got != "client-request-id" {
		t.Fatalf("unexpected response request id: %q", got)
	}
	var envelope struct {
		RequestID string `json:"request_id"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.RequestID != "client-request-id" {
		t.Fatalf("unexpected body request id: %q", envelope.RequestID)
	}
}

func TestRequestIDGenerated(t *testing.T) {
	engine := newTestEngine(t)
	recorder := performRequest(engine, http.MethodGet, "/buildings", "")
	generated := recorder.Header().Get("X-Request-ID")
	if len(generated) != 32 {
		t.Fatalf("unexpected generated request id: %q", generated)
	}
	var envelope struct {
		RequestID string `json:"request_id"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.RequestID != generated {
		t.Fatalf("body request id %q != header %q", envelope.RequestID, generated)
	}
}

func TestRouteErrorCodes(t *testing.T) {
	engine := newTestEngine(t)
	cases := []struct {
		name       string
		body       string
		wantStatus int
		wantCode   int
	}{
		{
			name:       "invalid floor enum",
			body:       `{"from":{"building":"C5","floor":"F9","room_no":"101"},"to":{"building":"C5","floor":"F2","room_no":"201"}}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   1001,
		},
		{
			name:       "invalid mode",
			body:       `{"from":{"building":"C5","floor":"F1","room_no":"101"},"to":{"building":"C5","floor":"F2","room_no":"201"},"mode":"walk"}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   1001,
		},
		{
			name:       "unknown room",
			body:       `{"from":{"building":"C5","floor":"F1","room_no":"999"},"to":{"building":"C5","floor":"F2","room_no":"201"}}`,
			wantStatus: http.StatusNotFound,
			wantCode:   1002,
		},
		{
			name:       "unsupported building",
			body:       `{"from":{"building":"C9","floor":"F1","room_no":"101"},"to":{"building":"C9","floor":"F2","room_no":"201"}}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   1001,
		},
		{
			name:       "allowed floors exclude target",
			body:       `{"from":{"building":"C5","floor":"F1","room_no":"101"},"to":{"building":"C5","floor":"F2","room_no":"201"},"allowed_floors":["F1"]}`,
			wantStatus: http.StatusBadRequest,
			wantCode:   1001,
		},
	}

	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			recorder := performRequest(engine, http.MethodPost, "/api/v1/route", item.body)
			if recorder.Code != item.wantStatus {
				t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
			}
			var envelope dto.ErrorEnvelope
			if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
				t.Fatalf("decode error: %v", err)
			}
			if envelope.Code != item.wantCode {
				t.Fatalf("code=%d, want %d body=%s", envelope.Code, item.wantCode, recorder.Body.String())
			}
		})
	}
}

func TestRouteNoRouteCode(t *testing.T) {
	raw := newTestRaw()
	corridors := make([]graph.RawEdge, 0, len(raw.Edges))
	for _, edge := range raw.Edges {
		if edge.EdgeType == "corridor" {
			corridors = append(corridors, edge)
		}
	}
	raw.Edges = corridors

	store, err := graph.Build(raw)
	if err != nil {
		t.Fatalf("build unchecked graph: %v", err)
	}
	catalog := service.NewCatalogService(store)
	planner := route.NewPlanner(store)
	engine := New(Dependencies{
		Route:   handler.NewRouteHandler(service.NewRouteService(catalog, planner, nil)),
		Nearby:  handler.NewNearbyHandler(service.NewNearbyService(catalog, planner)),
		Catalog: handler.NewCatalogHandler(catalog),
		Health:  handler.NewHealthHandler(nil, store, "test"),
	})

	body := `{"from":{"building":"C5","floor":"F1","room_no":"101"},"to":{"building":"C5","floor":"F2","room_no":"201"}}`
	recorder := performRequest(engine, http.MethodPost, "/api/v1/route", body)
	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var envelope dto.ErrorEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if envelope.Code != 2001 {
		t.Fatalf("unexpected no-route code: %+v", envelope)
	}
}

func TestCatalogNotFoundAndPagination(t *testing.T) {
	engine := newTestEngine(t)

	recorder := performRequest(engine, http.MethodGet, "/buildings/C9/floors", "")
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("floors status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var missing dto.ErrorEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &missing); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if missing.Code != 1003 {
		t.Fatalf("unexpected missing building code: %+v", missing)
	}

	recorder = performRequest(engine, http.MethodGet, "/classrooms?limit=1&offset=1", "")
	if recorder.Code != http.StatusOK {
		t.Fatalf("classrooms status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var classrooms dto.ClassroomsResponseEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &classrooms); err != nil {
		t.Fatalf("decode classrooms: %v", err)
	}
	if classrooms.Data.Total != 2 || len(classrooms.Data.Items) != 1 || classrooms.Data.Items[0].RoomNo != "201" {
		t.Fatalf("unexpected pagination: %+v", classrooms.Data)
	}
}

func TestPanicRecovery(t *testing.T) {
	engine := newTestEngine(t)
	engine.GET("/panic", func(*gin.Context) {
		panic("boom")
	})

	recorder := performRequest(engine, http.MethodGet, "/panic", "")
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	var envelope dto.ErrorEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if envelope.Code != 5000 || envelope.RequestID == "" {
		t.Fatalf("unexpected panic response: %+v", envelope)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	engine := newTestEngine(t)
	recorder := performRequest(engine, http.MethodGet, "/api/v1/route", "")
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("unexpected method status: %d", recorder.Code)
	}
}
