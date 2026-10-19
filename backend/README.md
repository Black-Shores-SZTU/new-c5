# C5 室内导航后端 · 从零启动

C5 公共教学楼室内导航 MVP 后端：Go + Gin + GORM + MySQL 8 + Redis 7。
仅覆盖 C5 一栋楼 F1–F5，不做跨楼、用户系统与实时定位。

## 1. 架构

```text
浏览器 / 前端
    │  http://localhost:8080
    ▼
Nginx（反向代理、gzip、访问日志）
    │  http://api:8080
    ▼
Go API（Gin）
    ├── MySQL 8：楼栋/楼层/房间/设施/节点/边
    └── Redis 7：路线缓存（map:version + singleflight）
```

- 启动时一次性读取六张表，校验后构建**只读内存图**；校验失败直接退出。
- 业务接口：`/api/v1/route`、`/api/v1/nearby`、`/buildings`、`/buildings/{code}/floors`、`/classrooms`、`/health`、`/swagger/index.html`。
- API 容器不对外发布端口，只通过 Nginx 访问。

## 2. 前置条件

| 依赖 | 版本 | 说明 |
|---|---|---|
| Docker Desktop | 29+ / Compose v2 | 需启用 WSL2（虚拟机平台）并可运行 Linux 容器 |
| Go | 1.26+ | 仅本地开发/测试需要 |
| 端口 | 8080 / 3306 / 6379 | 可在 `.env` 中改宿主机端口 |

## 3. 一键启动（Docker Compose）

```powershell
cd "d:\Indoor Navigation\backend"
Copy-Item .env.docker.example .env
docker compose up -d --build
docker compose ps
```

等待 `c5-api` 与 `c5-mysql` 变为 healthy（首次构建与 MySQL 初始化需要一点时间）。

验证：

```powershell
Invoke-RestMethod http://localhost:8080/health
powershell -ExecutionPolicy Bypass -File .\scripts\smoke.ps1
```

- 健康检查：<http://localhost:8080/health>
- Swagger UI：<http://localhost:8080/swagger/index.html>
- 路线示例：

```powershell
$body = '{"from":{"building":"C5","floor":"F1","room_no":"101"},"to":{"building":"C5","floor":"F2","room_no":"201"}}'
Invoke-RestMethod -Method Post -Uri http://localhost:8080/api/v1/route -ContentType application/json -Body $body
```

演示种子数据为 F1–F2 最小图，上述请求应返回约 24m / 32s、stairs 转层。

停止与重置：

```powershell
docker compose down          # 停止，保留数据
docker compose down -v       # 停止并删除 MySQL/Redis 数据卷（下次重新初始化）
```

> MySQL 初始化脚本只在**数据卷为空**时执行。修改 `migrations/` 后需 `down -v` 再启动，或手动执行 SQL。

## 4. 本地开发（API 跑在宿主机）

只启动依赖：

```powershell
cd "d:\Indoor Navigation\backend"
docker compose up -d mysql redis

$env:APP_ENV = "dev"
$env:APP_PORT = "8080"
$env:MYSQL_DSN = "c5:change-me@tcp(127.0.0.1:3306)/c5_navigation?charset=utf8mb4&parseTime=True&loc=Local"
$env:REDIS_ADDR = "127.0.0.1:6379"
$env:LOG_DIR = "logs"
go run ./cmd/server
```

密码以 `.env` 中的 `MYSQL_USER` / `MYSQL_PASSWORD` 为准。Go 程序不会自动读取 `.env`，需要显式设置环境变量。

常用命令：

```powershell
go build ./...
go run ./cmd/validate -building C5      # 数据校验 CLI，退出码 0/1/2
go run ./cmd/validate -building C5 -json
```

## 5. 测试

```powershell
go test ./...
go test -race ./...
go test ./... -cover
go test ./... -coverprofile=coverage.out
```

> `-race` 在 Windows 需要 `CGO_ENABLED=1` 与 gcc（MinGW-w64 / MSYS2）。未安装 gcc 时先运行普通测试。

覆盖范围：

| 包 | 内容 |
|---|---|
| `internal/graph` | 位掩码、图构建、教室/设施映射、校验规则、只读副本 |
| `internal/route` | 二叉堆、跨层权重、Dijkstra、最近设施、路径后处理 |
| `internal/cache` | key/avoid、TTL 夹取、singleflight、版本 bump、fail-open、脏缓存 |
| `internal/router` | 全部接口成功/错误码、Request ID、panic 恢复、健康检查、Swagger |

小图金标准（2 层 × 5 节点）位于 `internal/route/testdata/golden/`：

```powershell
go test ./internal/route -run TestGolden -update   # 修改算法后重建基线
go test ./internal/route -run TestGolden            # 校验基线
```

金标准覆盖：结合模式 24m/32s、仅电梯 26m/43s、同房间零距离、五类最近设施。

## 6. 日志

- API 日志：`logs/app.log`，同时输出 stdout。
- 轮转：默认单文件 100MB、保留 7 份、14 天、gzip 压缩。
- Nginx 日志：`logs/nginx/access.log`、`logs/nginx/error.log`。
- 访问日志字段：`request_id / method / path / query / status / latency_ms / client_ip / bytes`；5xx 用 error 级别，panic 堆栈同样落盘。

常用命令：

```powershell
Get-Content .\logs\app.log -Tail 50 -Wait
Get-Content .\logs\nginx\access.log -Tail 50
docker compose logs -f api
```

环境变量：

| 变量 | 默认 | 说明 |
|---|---|---|
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
| `LOG_DIR` | `logs` | 日志目录 |
| `LOG_FILE` | `app.log` | 文件名 |
| `LOG_MAX_SIZE_MB` | `100` | 单文件大小 |
| `LOG_MAX_BACKUPS` | `7` | 保留份数 |
| `LOG_MAX_AGE_DAYS` | `14` | 保留天数 |
| `LOG_COMPRESS` | `true` | 是否压缩旧文件 |

> Linux 下使用 bind mount 时，宿主机 `logs/` 需要对容器内 `app` 用户可写：
> `mkdir -p logs && chmod 0777 logs`。

## 7. 备份与恢复

MySQL 逻辑备份（`--single-transaction`，不影响在线读）：

```powershell
# Windows
powershell -ExecutionPolicy Bypass -File .\scripts\backup.ps1

# Linux / macOS
bash scripts/backup.sh
```

- 输出：`backups/mysql_YYYYMMDD_HHMMSS.sql.gz`
- 默认保留 14 天，可用 `-RetentionDays` 或 `RETENTION_DAYS` 覆盖。
- Redis 只存路线缓存，不纳入备份；`map:version` 丢失会自动重建为 1。

恢复到临时库验证：

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\restore.ps1 `
  -File .\backups\mysql_20261019_120000.sql.gz `
  -Database c5_navigation_restore
```

恢复后核对：

```powershell
docker compose exec mysql mysql -uroot -p -e "USE c5_navigation_restore; SHOW TABLES;"
```

定时备份：Windows 任务计划程序调用 `backup.ps1`；Linux 使用 cron，例如每天 02:00：

```cron
0 2 * * * cd /path/to/backend && bash scripts/backup.sh >> logs/backup.log 2>&1
```

## 8. 数据说明

- `migrations/001_schema.sql`：六张表 DDL。
- `migrations/002_seed_demo.sql`：F1–F2 最小演示数据，**不是 C5 实测数据**，仅用于 Compose 启动与冒烟。
- 真实数据按 `docs/measurement/README.md` 填写 CSV，后续阶段导入；导入后执行一次缓存版本自增：

```powershell
docker compose exec redis redis-cli INCR map:version
```

或调用代码中的 `BumpMapVersion`。

## 9. 常见问题

| 现象 | 处理 |
|---|---|
| `c5-api` 反复重启，日志出现“图校验未通过” | 数据库数据不合法；先跑 `go run ./cmd/validate -building C5` 定位 |
| `8080` 被占用 | 修改 `.env` 的 `NGINX_HOST_PORT`，例如 `18080` |
| `3306` / `6379` 被占用 | 修改 `.env` 的 `MYSQL_HOST_PORT` / `REDIS_HOST_PORT` |
| 修改了 SQL 但没生效 | 初始化脚本只执行一次；`docker compose down -v` 后重启 |
| Nginx 502 | `docker compose ps` 看 api 是否 healthy，`docker compose logs api` 查原因 |
| Redis 不可用 | 缓存自动关闭（fail-open），路线接口仍可用 |
| Linux 下 api 无法写日志 | `chmod 0777 logs`，或改为 Docker 命名卷 |

## 10. 目录

```text
backend/
├── cmd/server/            # 启动入口
├── cmd/validate/          # 数据校验 CLI
├── internal/
│   ├── cache/             # 路线缓存
│   ├── database/          # MySQL/Redis 连接
│   ├── dto/               # 请求/响应结构
│   ├── graph/             # 只读内存图 + 校验
│   ├── handler/           # HTTP 入口
│   ├── logging/           # slog + 文件轮转
│   ├── middleware/        # RequestID / AccessLog / Recovery
│   ├── model/             # GORM 实体
│   ├── repository/        # 批量读表
│   ├── route/             # 路线/最近设施算法
│   ├── router/            # 路由注册
│   └── service/           # 业务编排
├── pkg/                   # 统一响应、错误码
├── migrations/            # MySQL DDL 与演示种子
├── deploy/nginx/          # Nginx 配置
├── scripts/               # 备份、恢复、冒烟
├── docs/                  # 设计文档与 Swagger
├── logs/                  # 运行期日志（gitignore）
├── backups/               # 备份文件（gitignore）
├── Dockerfile
├── docker-compose.yml
├── .env.docker.example
└── Makefile
```

## 11. 安全提示

- 上线前必须修改 `.env` 中的 MySQL root/业务密码，不要把 `.env` 提交到仓库。
- MySQL/Redis 只绑定 `127.0.0.1`；API 不直接暴露端口。
- Swagger 默认开启；生产如需关闭，移除 `internal/router/router.go` 中的 Swagger 路由。
- 本 MVP 无鉴权，不要直接暴露到公网。
