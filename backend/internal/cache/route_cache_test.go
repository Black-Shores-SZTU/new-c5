package cache

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"indoor-navigation/backend/internal/dto"
	"indoor-navigation/backend/internal/graph"
)

func newTestCache(t *testing.T) (*RouteCache, *miniredis.Miniredis) {
	t.Helper()
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr(), MaxRetries: 0})
	cache := NewRouteCache(client, Options{
		BaseTTL: 15 * time.Minute,
		Jitter:  5 * time.Minute,
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := cache.EnsureMapVersion(ctx); err != nil {
		t.Fatalf("ensure map version: %v", err)
	}
	return cache, server
}

func TestRouteKeyAndAvoid(t *testing.T) {
	info := graph.ClassroomInfo{Building: "C5", Floor: "F1", RoomNo: "101"}
	if got := EndpointKey(info); got != "C5-F1-101" {
		t.Fatalf("unexpected endpoint key: %s", got)
	}

	key := RouteKey("1", "C5-F1-101", "C5-F3-305", "combined", "none")
	if key != "route:1:C5-F1-101:C5-F3-305:combined:none" {
		t.Fatalf("unexpected route key: %s", key)
	}

	if got := AvoidKey(0); got != "none" {
		t.Fatalf("empty allowed floors should be none, got %s", got)
	}
	if got := AvoidKey(graph.FloorF1 | graph.FloorF2 | graph.FloorF3); got != "F4-F5" {
		t.Fatalf("unexpected avoid key: %s", got)
	}
	all := graph.FloorF1 | graph.FloorF2 | graph.FloorF3 | graph.FloorF4 | graph.FloorF5
	if got := AvoidKey(all); got != "none" {
		t.Fatalf("all floors should be none, got %s", got)
	}
}

func TestGetOrComputeAndTTL(t *testing.T) {
	cache, server := newTestCache(t)
	ctx := context.Background()
	key := RouteKey("1", "C5-F1-101", "C5-F3-305", "combined", "none")

	var calls atomic.Int32
	compute := func(context.Context) (dto.RouteResponse, error) {
		calls.Add(1)
		return dto.RouteResponse{DistanceM: 42}, nil
	}

	first, cached, err := cache.GetOrCompute(ctx, key, compute)
	if err != nil {
		t.Fatalf("first compute: %v", err)
	}
	if cached || first.DistanceM != 42 || calls.Load() != 1 {
		t.Fatalf("unexpected first result: cached=%v calls=%d value=%+v", cached, calls.Load(), first)
	}

	second, cached, err := cache.GetOrCompute(ctx, key, compute)
	if err != nil {
		t.Fatalf("second compute: %v", err)
	}
	if !cached || second.DistanceM != 42 || calls.Load() != 1 {
		t.Fatalf("unexpected second result: cached=%v calls=%d value=%+v", cached, calls.Load(), second)
	}

	ttl := server.TTL(key)
	if ttl < 15*time.Minute || ttl > 20*time.Minute {
		t.Fatalf("ttl out of range: %v", ttl)
	}
}

func TestGetOrComputeSingleflight(t *testing.T) {
	cache, _ := newTestCache(t)
	ctx := context.Background()
	key := RouteKey("1", "C5-F1-101", "C5-F3-305", "combined", "none")

	var calls atomic.Int32
	started := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	compute := func(context.Context) (dto.RouteResponse, error) {
		calls.Add(1)
		once.Do(func() { close(started) })
		<-release
		return dto.RouteResponse{DistanceM: 1}, nil
	}

	const workers = 8
	var wait sync.WaitGroup
	errs := make(chan error, workers)
	for index := 0; index < workers; index++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			if _, _, err := cache.GetOrCompute(ctx, key, compute); err != nil {
				errs <- err
			}
		}()
	}

	<-started
	time.Sleep(50 * time.Millisecond)
	close(release)
	wait.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent compute: %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected single computation, got %d", calls.Load())
	}
}

func TestMapVersionBump(t *testing.T) {
	cache, _ := newTestCache(t)
	ctx := context.Background()

	version, err := cache.MapVersion(ctx)
	if err != nil || version != "1" {
		t.Fatalf("unexpected initial version: %s err=%v", version, err)
	}

	next, err := cache.BumpMapVersion(ctx)
	if err != nil || next != 2 {
		t.Fatalf("unexpected bump result: %d err=%v", next, err)
	}

	version, err = cache.MapVersion(ctx)
	if err != nil || version != "2" {
		t.Fatalf("unexpected bumped version: %s err=%v", version, err)
	}

	oldKey := RouteKey("1", "C5-F1-101", "C5-F3-305", "combined", "none")
	newKey := RouteKey(version, "C5-F1-101", "C5-F3-305", "combined", "none")
	if oldKey == newKey {
		t.Fatal("version bump did not change cache key")
	}
}

func TestGetOrComputeWithoutRedis(t *testing.T) {
	cache := NewRouteCache(nil, Options{})
	called := false
	value, cached, err := cache.GetOrCompute(context.Background(), "route:1:test", func(context.Context) (dto.RouteResponse, error) {
		called = true
		return dto.RouteResponse{DistanceM: 7}, nil
	})
	if err != nil {
		t.Fatalf("compute without redis: %v", err)
	}
	if cached || !called || value.DistanceM != 7 {
		t.Fatalf("unexpected fail-open result: cached=%v called=%v value=%+v", cached, called, value)
	}
}

func TestGetOrComputeRedisDown(t *testing.T) {
	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer server.Close()

	client := redis.NewClient(&redis.Options{
		Addr:         server.Addr(),
		MaxRetries:   0,
		DialTimeout:  100 * time.Millisecond,
		ReadTimeout:  100 * time.Millisecond,
		WriteTimeout: 100 * time.Millisecond,
	})
	cache := NewRouteCache(client, Options{BaseTTL: 15 * time.Minute})
	server.Close()

	value, cached, err := cache.GetOrCompute(context.Background(), "route:1:test", func(context.Context) (dto.RouteResponse, error) {
		return dto.RouteResponse{DistanceM: 9}, nil
	})
	if err != nil {
		t.Fatalf("compute with redis down: %v", err)
	}
	if cached || value.DistanceM != 9 {
		t.Fatalf("unexpected fail-open result: cached=%v value=%+v", cached, value)
	}
}

func TestAvoidKeyOrder(t *testing.T) {
	cases := []struct {
		allowed graph.FloorBit
		want    string
	}{
		{0, "none"},
		{graph.FloorF1, "F2-F3-F4-F5"},
		{graph.FloorF2 | graph.FloorF4, "F1-F3-F5"},
		{graph.FloorF1 | graph.FloorF2 | graph.FloorF3 | graph.FloorF4 | graph.FloorF5, "none"},
	}
	for _, item := range cases {
		if got := AvoidKey(item.allowed); got != item.want {
			t.Fatalf("AvoidKey(%05b) = %q, want %q", item.allowed, got, item.want)
		}
	}

	info := graph.ClassroomInfo{Building: "C5", Floor: "F3", RoomNo: "305"}
	if got := EndpointKey(info); got != "C5-F3-305" {
		t.Fatalf("unexpected endpoint key: %s", got)
	}
}

func TestTTLClamp(t *testing.T) {
	short := NewRouteCache(nil, Options{BaseTTL: time.Minute, Jitter: time.Hour})
	if short.baseTTL != minTTL {
		t.Fatalf("base ttl not clamped up: %v", short.baseTTL)
	}
	if short.baseTTL+short.jitter > maxTTL {
		t.Fatalf("ttl window exceeds max: %v", short.baseTTL+short.jitter)
	}
	for index := 0; index < 100; index++ {
		ttl := short.ttl()
		if ttl < minTTL || ttl > maxTTL {
			t.Fatalf("ttl out of range: %v", ttl)
		}
	}

	long := NewRouteCache(nil, Options{BaseTTL: time.Hour, Jitter: time.Hour})
	if long.baseTTL != maxTTL || long.jitter != 0 {
		t.Fatalf("long ttl not clamped: base=%v jitter=%v", long.baseTTL, long.jitter)
	}
}

func TestComputeErrorNotCached(t *testing.T) {
	cache, server := newTestCache(t)
	ctx := context.Background()
	key := RouteKey("1", "C5-F1-101", "C5-F2-201", "combined", "none")

	var calls atomic.Int32
	compute := func(context.Context) (dto.RouteResponse, error) {
		calls.Add(1)
		return dto.RouteResponse{}, context.DeadlineExceeded
	}
	for index := 0; index < 2; index++ {
		if _, _, err := cache.GetOrCompute(ctx, key, compute); err == nil {
			t.Fatal("expected compute error")
		}
	}
	if calls.Load() != 2 {
		t.Fatalf("failed result should not be cached, calls=%d", calls.Load())
	}
	if server.Exists(key) {
		t.Fatal("failed result unexpectedly written to redis")
	}
}

func TestCorruptCacheRecomputed(t *testing.T) {
	cache, server := newTestCache(t)
	ctx := context.Background()
	key := RouteKey("1", "C5-F1-101", "C5-F2-201", "combined", "none")
	if err := cache.client.Set(ctx, key, "not-json", time.Minute).Err(); err != nil {
		t.Fatalf("seed corrupt cache: %v", err)
	}

	var calls atomic.Int32
	compute := func(context.Context) (dto.RouteResponse, error) {
		calls.Add(1)
		return dto.RouteResponse{DistanceM: 12}, nil
	}
	value, cached, err := cache.GetOrCompute(ctx, key, compute)
	if err != nil {
		t.Fatalf("recompute after corrupt cache: %v", err)
	}
	if cached || value.DistanceM != 12 || calls.Load() != 1 {
		t.Fatalf("unexpected recompute result: cached=%v calls=%d value=%+v", cached, calls.Load(), value)
	}
	if !server.Exists(key) {
		t.Fatal("recomputed result should be cached again")
	}

	if _, cached, err := cache.GetOrCompute(ctx, key, compute); err != nil || !cached {
		t.Fatalf("second read should hit cache: cached=%v err=%v", cached, err)
	}
}

func TestMapVersionAutoInit(t *testing.T) {
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	cache := NewRouteCache(client, Options{})

	version, err := cache.MapVersion(context.Background())
	if err != nil || version != "1" {
		t.Fatalf("unexpected auto init version: %q err=%v", version, err)
	}
}
