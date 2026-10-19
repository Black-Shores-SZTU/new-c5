# C5 室内导航后端

## 1. 项目范围（MVP）

- 仅覆盖 C5 公共教学楼一栋楼、F1–F5 五个楼层。
- 不做跨楼导航、不做用户系统、不接实时定位。
- 数据来源为现场实测 CSV，后续阶段导入 MySQL。

## 2. 技术栈

- Go + Gin
- GORM + MySQL 8
- Redis 7

## 3. 当前阶段

本阶段在骨架上接入了只读内存图、路线算法与 HTTP API：

- 已有：目录结构、启动入口、六张表 GORM 模型、批量读取、只读内存图、校验 CLI。
- 已有：`internal/route` 路线算法包（二叉堆 Dijkstra、楼层掩码、CPath、同层压缩、仅电梯/结合模式、跨层权重）。
- 已有：最近设施 BFS（卫生间/电梯/楼梯/安全出口/开水间）与路径预计时间、转层事件。
- 已有：`/api/v1/route`、`/api/v1/nearby`、`/buildings`、`/buildings/{code}/floors`、`/classrooms`、`/health`。
- 已有：统一响应与错误码、参数校验、Request ID、Swagger 文档。
- 已有：Redis 路线缓存（map_version、TTL 抖动、singleflight、fail-open）。
- 已有：单元测试（graph/route/cache/router）、小图金标准、接口测试。
- 已有：Docker Compose（MySQL 8 / Redis 7 / API / Nginx）、日志落盘与轮转、MySQL 备份/恢复脚本、从零启动 README。
- 启动时读取 `MYSQL_DSN` 指定楼栋的数据，校验通过后构建 `GraphStore`；校验失败则退出。
- 未做：用户系统、实时定位。

## 4. 目录结构

```text
backend/
├── cmd/server/            # 启动入口
├── internal/
│   ├── cache/             # 路线缓存（Redis + singleflight）
│   ├── database/          # MySQL/Redis 连接
│   ├── dto/               # 请求/响应 DTO
│   ├── graph/             # 只读内存图与目录元数据
│   ├── handler/           # HTTP 入参/出参
│   ├── service/           # 业务编排
│   ├── repository/        # MySQL 数据访问
│   ├── route/             # 路线/设施算法包（含 golden）
│   ├── router/            # HTTP 路由注册
│   ├── middleware/        # RequestID / AccessLog / Recovery
│   ├── logging/           # slog + lumberjack 文件轮转
│   └── model/             # GORM 实体
├── pkg/
│   ├── response/          # 统一响应
│   └── errcode/           # 错误码
├── migrations/            # DDL 与演示种子 SQL
├── deploy/nginx/          # Nginx 反向代理配置
├── scripts/               # 备份/恢复/冒烟脚本
├── docs/
│   ├── README.md
│   ├── database-design.md
│   ├── graph-design.md
│   ├── swagger/           # Swagger UI 文档
│   └── measurement/       # 实测 CSV 模板与说明
├── logs/                  # 运行期日志（gitignore）
├── backups/               # 备份文件（gitignore）
├── Dockerfile
├── docker-compose.yml
├── README.md              # 从零启动
├── .env.example
├── .env.docker.example
├── .gitignore
├── go.mod
└── Makefile
```

## 5. 环境要求

| 依赖 | 版本 |
|---|---|
| Go | 1.26+ |
| MySQL | 8.0+ |
| Redis | 7.0+ |
| Make | 可选 |

## 6. 环境变量

复制 `.env.example` 为 `.env` 并按本机环境修改。当前代码读取 `APP_ENV`、`APP_PORT`、`BUILDING_CODE` 与 `MYSQL_DSN`。

| 变量 | 说明 | 示例 |
|---|---|---|
| `APP_ENV` | 运行环境 | `dev` / `prod` |
| `APP_PORT` | HTTP 端口 | `8080` |
| `BUILDING_CODE` | 启动时加载的楼栋 | `C5` |
| `MYSQL_DSN` | MySQL DSN（启动必需） | `user:password@tcp(127.0.0.1:3306)/c5_navigation?charset=utf8mb4&parseTime=True&loc=Local` |
| `REDIS_ADDR` | 路线缓存 Redis 地址 | `127.0.0.1:6379` |
| `REDIS_PASSWORD` | 路线缓存 Redis 密码 | 空 |
| `REDIS_DB` | 路线缓存 Redis 库号 | `0` |
| `ROUTE_CACHE_TTL_SECONDS` | 路线缓存基础 TTL（秒） | `900` |
| `ROUTE_CACHE_JITTER_SECONDS` | TTL 随机抖动上限（秒） | `300` |
| `LOG_LEVEL` | 日志级别 | `info` |
| `LOG_DIR` | 日志目录 | `logs` |
| `LOG_FILE` | 日志文件 | `app.log` |
| `LOG_MAX_SIZE_MB` | 单文件大小 | `100` |
| `LOG_MAX_BACKUPS` | 保留份数 | `7` |
| `LOG_MAX_AGE_DAYS` | 保留天数 | `14` |
| `LOG_COMPRESS` | 旧日志压缩 | `true` |

## 7. 启动

### 7.1 Docker Compose（推荐）

```powershell
cd "d:\Indoor Navigation\backend"
Copy-Item .env.docker.example .env
docker compose up -d --build
```

Nginx 入口：`http://localhost:8080`。健康检查、Swagger、备份与故障排查见根目录 [README.md](../README.md)。

### 7.2 本地 Go 进程

```powershell
cd "d:\Indoor Navigation\backend"
Copy-Item .env.example .env
go mod tidy
go run ./cmd/server
```

健康检查：

```text
GET http://localhost:8080/health
```

Swagger UI：

```text
http://localhost:8080/swagger/index.html
```

也可使用 `make run` / `make build` / `make fmt` / `make tidy`。

### 数据校验 CLI

```powershell
cd "d:\Indoor Navigation\backend"
$env:MYSQL_DSN = "user:password@tcp(127.0.0.1:3306)/c5_navigation?charset=utf8mb4&parseTime=True&loc=Local"

go run ./cmd/validate
go run ./cmd/validate -building C5 -json
go run ./cmd/validate -strict=false
```

退出码：`0` 通过，`1` 校验失败，`2` 配置/连接/读取失败。详细规则见 [graph-design.md](graph-design.md)。

启动服务时同样会执行严格校验，校验不通过不会进入 HTTP 服务。

### 路线算法自测

```powershell
cd "d:\Indoor Navigation\backend"
go test ./...
go test -race ./...
go test ./... -cover
```

测试使用 2 层 × 每层 5 节点的合成图，不依赖 MySQL；缓存测试使用 miniredis。
小图金标准位于 `internal/route/testdata/golden/`，修改算法后执行
`go test ./internal/route -run TestGolden -update` 重建。

### 路线缓存

Key 格式：

```text
route:{version}:{start}:{end}:{mode}:{avoid}
```

示例：

```text
route:1:C5-F1-101:C5-F3-305:combined:none
route:7:C5-F1-101:C5-F3-305:elevator_only:F4-F5
```

- `version`：Redis `map:version` 当前值。
- `start` / `end`：`楼栋-楼层-房号`。
- `mode`：`combined` / `elevator_only`。
- `avoid`：`allowed_floors` 的补集，升序用 `-` 连接；无排除为 `none`。
- TTL：基础 15 分钟 + 0–5 分钟随机抖动，总范围 15–20 分钟。
- 未命中时通过 `singleflight` 合并计算，并二次检查缓存。
- 数据更新后执行 `BumpMapVersion`（`INCR map:version`），旧 key 不再被读取，由 TTL 自然淘汰。
- Redis 不可用时 fail-open，直接计算路线并记录日志。

## 8. 接口总览

MVP 不需要鉴权。业务接口按冻结路径注册：

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/health` | 健康检查（MySQL + 内存图） |
| POST | `/api/v1/route` | 起终点教室路线规划 |
| GET | `/api/v1/nearby` | 最近设施查询 |
| GET | `/buildings` | 楼栋列表（MVP 仅 C5） |
| GET | `/buildings/{code}/floors` | 指定楼栋楼层列表 |
| GET | `/classrooms` | 教室搜索（楼栋/楼层/关键词/分页） |
| GET | `/swagger/index.html` | Swagger UI |

请求/响应字段、示例与错误码以 Swagger 文档为准；生成文件位于 `docs/swagger/`。

## 9. 统一响应

```json
{
  "code": 0,
  "message": "ok",
  "data": {},
  "request_id": "8f3c2a7e"
}
```

错误码：

| code | HTTP | 含义 |
|---|---|---|
| 0 | 200 | 成功 |
| 1001 | 400 | 参数校验失败 |
| 1002 | 404 | 房间不存在 |
| 1003 | 404 | 楼栋或楼层不存在 |
| 2001 | 422 | 未找到可行路线 |
| 5000 | 500 | 服务内部错误 |

## 10. 实测数据流程

1. 按 `docs/measurement/README.md` 填写六张 CSV。
2. 后续阶段实现 CSV 校验与导入命令。
3. 导入后生成 `graph_nodes` / `graph_edges`，供路线规划使用。

## 11. 文档索引

- [数据库表设计草案](database-design.md)
- [只读内存图与数据校验设计](graph-design.md)
- [实测 CSV 填写说明](measurement/README.md)
- [楼层模板](measurement/floors-template.csv)
- [房间模板](measurement/rooms-template.csv)
- [设施模板](measurement/facilities-template.csv)
- [路径边模板](measurement/edges-template.csv)
