package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	rand "math/rand/v2"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"

	"indoor-navigation/backend/internal/dto"
	"indoor-navigation/backend/internal/graph"
)

const (
	DefaultMapVersionKey = "map:version"
	DefaultBaseTTL       = 15 * time.Minute
	DefaultJitter        = 5 * time.Minute
	minTTL               = 5 * time.Minute
	maxTTL               = 30 * time.Minute
)

var (
	ErrDisabled  = errors.New("cache: disabled")
	allFloorBits = graph.FloorF1 | graph.FloorF2 | graph.FloorF3 | graph.FloorF4 | graph.FloorF5
)

// Options 路线缓存配置。
type Options struct {
	MapVersionKey string
	BaseTTL       time.Duration
	Jitter        time.Duration
}

// RouteCache 路线结果缓存，Redis 不可用时由调用方降级为直接计算。
type RouteCache struct {
	client        redis.UniversalClient
	group         singleflight.Group
	mapVersionKey string
	baseTTL       time.Duration
	jitter        time.Duration
}

func NewRouteCache(client redis.UniversalClient, options Options) *RouteCache {
	if options.MapVersionKey == "" {
		options.MapVersionKey = DefaultMapVersionKey
	}
	if options.BaseTTL <= 0 {
		options.BaseTTL = DefaultBaseTTL
	}
	if options.BaseTTL < minTTL {
		options.BaseTTL = minTTL
	}
	if options.BaseTTL > maxTTL {
		options.BaseTTL = maxTTL
	}
	if options.Jitter < 0 {
		options.Jitter = 0
	}
	if options.BaseTTL+options.Jitter > maxTTL {
		options.Jitter = maxTTL - options.BaseTTL
	}
	return &RouteCache{
		client:        client,
		mapVersionKey: options.MapVersionKey,
		baseTTL:       options.BaseTTL,
		jitter:        options.Jitter,
	}
}

// EndpointKey 将教室转为缓存 key 中的端点片段，如 C5-F1-101。
func EndpointKey(info graph.ClassroomInfo) string {
	return strings.Join([]string{info.Building, info.Floor, info.RoomNo}, "-")
}

// AvoidKey 将 allowed_floors 转成排除楼层片段；无排除时返回 none。
func AvoidKey(allowed graph.FloorBit) string {
	if allowed == 0 || allowed&allFloorBits == allFloorBits {
		return "none"
	}
	excluded := allFloorBits &^ allowed
	codes := make([]string, 0, 5)
	for _, bit := range []graph.FloorBit{
		graph.FloorF1,
		graph.FloorF2,
		graph.FloorF3,
		graph.FloorF4,
		graph.FloorF5,
	} {
		if excluded&bit == 0 {
			continue
		}
		if code, ok := graph.FloorCode(bit); ok {
			codes = append(codes, code)
		}
	}
	if len(codes) == 0 {
		return "none"
	}
	return strings.Join(codes, "-")
}

// RouteKey 生成 route:{version}:{start}:{end}:{mode}:{avoid}。
func RouteKey(version, start, end, mode, avoid string) string {
	return fmt.Sprintf("route:%s:%s:%s:%s:%s", version, start, end, mode, avoid)
}

// EnsureMapVersion 初始化 map_version；已存在时保持原值。
func (c *RouteCache) EnsureMapVersion(ctx context.Context) error {
	if c == nil || c.client == nil {
		return ErrDisabled
	}
	return c.client.SetNX(ctx, c.mapVersionKey, "1", 0).Err()
}

// MapVersion 读取当前 map_version。
func (c *RouteCache) MapVersion(ctx context.Context) (string, error) {
	if c == nil || c.client == nil {
		return "", ErrDisabled
	}
	value, err := c.client.Get(ctx, c.mapVersionKey).Result()
	if err == nil {
		return value, nil
	}
	if !errors.Is(err, redis.Nil) {
		return "", err
	}

	created, err := c.client.SetNX(ctx, c.mapVersionKey, "1", 0).Result()
	if err != nil {
		return "", err
	}
	if created {
		return "1", nil
	}
	return c.client.Get(ctx, c.mapVersionKey).Result()
}

// BumpMapVersion 自增 map_version，使旧缓存 key 失效。
func (c *RouteCache) BumpMapVersion(ctx context.Context) (int64, error) {
	if c == nil || c.client == nil {
		return 0, ErrDisabled
	}
	return c.client.Incr(ctx, c.mapVersionKey).Result()
}

// GetOrCompute 读取缓存；未命中时通过 singleflight 合并计算并回填。
// Redis 不可用时直接计算，保证路线请求可用。
func (c *RouteCache) GetOrCompute(
	ctx context.Context,
	key string,
	compute func(context.Context) (dto.RouteResponse, error),
) (dto.RouteResponse, bool, error) {
	if c == nil || c.client == nil {
		value, err := compute(ctx)
		return value, false, err
	}

	if value, ok := c.get(ctx, key); ok {
		return value, true, nil
	}

	result, err, _ := c.group.Do(key, func() (any, error) {
		if value, ok := c.get(ctx, key); ok {
			return value, nil
		}
		value, err := compute(ctx)
		if err != nil {
			return nil, err
		}
		c.set(ctx, key, value)
		return value, nil
	})
	if err != nil {
		return dto.RouteResponse{}, false, err
	}

	value, ok := result.(dto.RouteResponse)
	if !ok {
		return dto.RouteResponse{}, false, fmt.Errorf("cache: unexpected result type")
	}
	return value, false, nil
}

func (c *RouteCache) get(ctx context.Context, key string) (dto.RouteResponse, bool) {
	raw, err := c.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return dto.RouteResponse{}, false
	}
	if err != nil {
		log.Printf("route cache get %s: %v", key, err)
		return dto.RouteResponse{}, false
	}

	var value dto.RouteResponse
	if err := json.Unmarshal(raw, &value); err != nil {
		log.Printf("route cache decode %s: %v", key, err)
		_ = c.client.Del(ctx, key).Err()
		return dto.RouteResponse{}, false
	}
	return value, true
}

func (c *RouteCache) set(ctx context.Context, key string, value dto.RouteResponse) {
	payload, err := json.Marshal(value)
	if err != nil {
		log.Printf("route cache encode %s: %v", key, err)
		return
	}
	if err := c.client.Set(ctx, key, payload, c.ttl()).Err(); err != nil {
		log.Printf("route cache set %s: %v", key, err)
	}
}

func (c *RouteCache) ttl() time.Duration {
	ttl := c.baseTTL
	if c.jitter > 0 {
		ttl += time.Duration(rand.Int64N(int64(c.jitter)))
	}
	if ttl > maxTTL {
		return maxTTL
	}
	if ttl < minTTL {
		return minTTL
	}
	return ttl
}
