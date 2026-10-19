# 只读内存图与数据校验设计

## 1. 目标

- 服务启动时批量读取 MySQL 中的 C5 数据，构建只读内存图。
- 运行期不修改图结构，可安全并发读取。
- 教室通过哈希直接映射到门节点与相邻关键节点，避免路线规划时做最近点搜索。
- 楼层与类型使用位掩码，便于后续搜索和筛选。
- 启动与 CLI 共用同一套校验规则。

## 2. 代码位置

```text
internal/graph/
├── types.go       # NodeID、位掩码、Node/Edge、RawData
├── builder.go     # 原始数据 → 邻接表、classroomToNodes、掩码
├── validate.go    # 数据校验规则
└── store.go       # 只读 GraphStore 与查询方法

internal/model/models.go       # 六张表的 GORM 实体
internal/repository/graph_repo.go  # 批量读取
cmd/validate/main.go           # 校验 CLI
```

## 3. 数据结构

- `NodeID`：楼栋内从 0 开始的连续编号。
- `FloorBit`：F1–F5 各占一位，可用于楼层过滤。
- `TypeMask`：教室、教师休息室、演播厅、电梯、楼梯、卫生间、入口、开水间、走廊、核心筒、门节点。
- `BuildingGraph`：节点数组、邻接表、`floor|code → NodeID` 索引、楼层/类型掩码并集。
- `GraphStore`：
  - `buildingGraphs`：楼栋编码 → 楼栋图。
  - `classroomToNodes`：`(楼栋, 楼层, 房号)` → 门节点与一跳邻居。
  - `stats`：图规模统计。
- `RawData`：repository 批量查询后的中间结构，builder 与 validator 共用。

只读约束：

- `BuildingGraph` 与 `GraphStore` 的字段不导出，构建完成后没有写方法。
- `Nodes()` / `Neighbors()` 返回副本，避免调用方修改内部切片。
- 启动时只构建一次；后续如需热更新，可再引入 `atomic.Pointer[GraphStore]`。

## 4. 构建流程

1. 读取楼栋、楼层、房间、设施、节点、边六张表。
2. 按楼层与节点 code 排序，分配稳定 `NodeID`。
3. 将无向边（`bidirectional=true`）展开为双向邻接；有向边只保留原方向。
4. 根据 door 节点的 `ref_type/ref_code` 建立 `classroomToNodes`：
   - `DoorNodeID` 为房间门节点。
   - `NodeIDs` 为门节点本身 + 一跳邻居，去重并排序。
5. 汇总楼层与类型位掩码。
6. 返回冻结的 `GraphStore`。

## 5. 校验规则

CLI 与启动流程共用 `graph.Validate`。

| 规则 | 内容 |
|---|---|
| fk | 楼层→楼栋；房间/设施/节点→楼层；边→节点；房间→door 节点；设施→door/core/corridor 接入节点 |
| unique | 楼层 level/code、同层房号、同层设施 code、同层节点 code、边 `(from,to,type)`、每个房间仅一个 door 节点、每个设施仅一个接入节点 |
| distance | `distance_m > 0`、尺寸 > 0、层高 > 0 |
| coordinate | 坐标与边界必须为有限数值，楼层 `min < max` |
| consistency | `floor_code` 与 `level` 一致；房间引用只能出现在 door 节点；设施引用可出现在 door/core/corridor 节点；走廊边同层；楼梯/电梯边跨层且两端为 core 节点 |
| bidirectional | 无向边只存一行；同一无向对出现反向重复记录时报错 |
| self_loop | 禁止 `from == to` |
| isolated | 任意节点度为 0 时报错 |
| connectivity | 按无向关系做 BFS；存在多个连通分量时报错并列出各分量节点 |
| classroom | 每个房间必须有 door 节点，否则 `classroomToNodes` 无法建立 |
| facility | 每个设施必须有且仅有一个接入节点；`exit` 作为合法设施类型 |

严格模式：

- 启动流程固定 `strict=true`，校验不通过即退出。
- CLI 默认 `strict=true`，可用 `-strict=false` 仅将警告保留为警告。

## 6. 校验 CLI

```powershell
cd "d:\Indoor Navigation\backend"
$env:MYSQL_DSN = "user:password@tcp(127.0.0.1:3306)/c5_navigation?charset=utf8mb4&parseTime=True&loc=Local"

go run ./cmd/validate
go run ./cmd/validate -building C5 -json
go run ./cmd/validate -strict=false
```

参数：

| 参数 | 默认 | 说明 |
|---|---|---|
| `-dsn` | `MYSQL_DSN` | MySQL 连接串 |
| `-building` | `C5` | 楼栋编码 |
| `-json` | `false` | JSON 报告 |
| `-strict` | `true` | 警告视为错误 |

退出码：

- `0`：校验通过。
- `1`：存在校验错误。
- `2`：配置、连接或读取失败。

## 7. 启动接入

`cmd/server/main.go` 启动顺序：

1. 读取 `MYSQL_DSN`，连接 MySQL 并 Ping。
2. 批量加载指定楼栋数据。
3. 执行严格校验；失败时输出错误并退出。
4. 构建 `GraphStore`，记录图规模。
5. 启动 HTTP 服务。

当前阶段已完成图加载与校验，并在 `internal/route` 中实现路线算法、最近设施查找与路径时间估算；HTTP 路线接口尚未接入。

## 8. 最近设施与路径后处理

- 最近设施：从教室门节点或指定节点出发做按层 BFS；跳过楼层掩码外节点；同层候选按二维直线距离升序；每类设施返回一个结果。
- 默认查询类型：卫生间、电梯、楼梯、安全出口、开水间。
- 路径后处理：按楼层切分 Segment、删除同向共线点、生成楼梯/电梯转层事件、汇总实际距离。
- 预计时间：步行按 1.2m/s；楼梯 15s/层；电梯 20s 等待 + 5s/层；结果四舍五入到整秒。
