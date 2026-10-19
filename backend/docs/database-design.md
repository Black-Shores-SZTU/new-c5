# 数据库表设计草案（C5 MVP）

> 状态：草案，后续阶段按此生成 GORM 模型与迁移脚本。
> 数据库：MySQL 8.0+，字符集 `utf8mb4`，排序规则 `utf8mb4_0900_ai_ci`。

## 1. 统一约定

- 主键：`id BIGINT UNSIGNED AUTO_INCREMENT`。
- 时间字段：`created_at`、`updated_at`，由 GORM 自动维护。
- MVP 不启用软删除，不增加 `deleted_at`；需要时再补。
- 坐标单位：米，保留两位小数。
- 坐标系：原点在 C5 建筑中心；X 向东为正，Z 向南为正，Y 向上为正。
- 枚举值统一使用小写英文，避免中文入库。

## 2. 表结构

### 2.1 buildings（楼栋）

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | BIGINT UNSIGNED | PK | 主键 |
| code | VARCHAR(32) | NOT NULL, UNIQUE | 楼栋编码，MVP 为 `C5` |
| name | VARCHAR(64) | NOT NULL | 显示名称，如 `C5栋` |
| full_name | VARCHAR(128) | NOT NULL | 全称，如 `C5 公共教学楼` |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |

### 2.2 floors（楼层）

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | BIGINT UNSIGNED | PK | 主键 |
| building_id | BIGINT UNSIGNED | NOT NULL, FK | 所属楼栋 |
| floor_code | VARCHAR(8) | NOT NULL | `F1`–`F5` |
| name | VARCHAR(32) | NOT NULL | `一层`–`五层` |
| level | TINYINT UNSIGNED | NOT NULL | 1–5 |
| min_x | DECIMAL(8,2) | NOT NULL | 楼层边界最小 X |
| max_x | DECIMAL(8,2) | NOT NULL | 楼层边界最大 X |
| min_z | DECIMAL(8,2) | NOT NULL | 楼层边界最小 Z |
| max_z | DECIMAL(8,2) | NOT NULL | 楼层边界最大 Z |
| floor_height | DECIMAL(5,2) | NOT NULL | 层高（米） |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |

索引：`UNIQUE (building_id, level)`，`INDEX (building_id)`。

### 2.3 rooms（房间）

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | BIGINT UNSIGNED | PK | 主键 |
| floor_id | BIGINT UNSIGNED | NOT NULL, FK | 所属楼层 |
| room_no | VARCHAR(16) | NOT NULL | 房号，如 `101`、`111A` |
| name | VARCHAR(64) | NOT NULL | 显示名称，默认同房号 |
| type | VARCHAR(32) | NOT NULL | `classroom` / `teacher_rest` / `auditorium` |
| corridor | VARCHAR(16) | NOT NULL | 临近走廊编码 |
| center_x | DECIMAL(8,2) | NOT NULL | 房间中心 X |
| center_z | DECIMAL(8,2) | NOT NULL | 房间中心 Z |
| size_x | DECIMAL(8,2) | NOT NULL | X 方向尺寸 |
| size_z | DECIMAL(8,2) | NOT NULL | Z 方向尺寸 |
| door_x | DECIMAL(8,2) | NOT NULL | 门中心 X |
| door_z | DECIMAL(8,2) | NOT NULL | 门中心 Z |
| open_side | VARCHAR(8) | NOT NULL | 门朝向：`north` / `south` / `east` / `west` |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |

索引：`UNIQUE (floor_id, room_no)`，`INDEX (floor_id)`，`INDEX (type)`。

### 2.4 facilities（设施）

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | BIGINT UNSIGNED | PK | 主键 |
| floor_id | BIGINT UNSIGNED | NOT NULL, FK | 所属楼层 |
| code | VARCHAR(32) | NOT NULL | 设施编码，如 `C5-F1-ENTRANCE` |
| name | VARCHAR(64) | NOT NULL | 显示名称 |
| type | VARCHAR(32) | NOT NULL | `elevator` / `stairs` / `restroom` / `entrance` / `water` / `auditorium` / `exit` |
| corridor | VARCHAR(16) | NOT NULL | 临近走廊或核心筒编码 |
| center_x | DECIMAL(8,2) | NOT NULL | 中心 X |
| center_z | DECIMAL(8,2) | NOT NULL | 中心 Z |
| size_x | DECIMAL(8,2) | NOT NULL | X 方向尺寸 |
| size_z | DECIMAL(8,2) | NOT NULL | Z 方向尺寸 |
| door_x | DECIMAL(8,2) | NOT NULL | 门/接入点 X |
| door_z | DECIMAL(8,2) | NOT NULL | 门/接入点 Z |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |

索引：`UNIQUE (floor_id, code)`，`INDEX (floor_id)`，`INDEX (type)`。

### 2.5 graph_nodes（路径节点）

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | BIGINT UNSIGNED | PK | 主键 |
| floor_id | BIGINT UNSIGNED | NOT NULL, FK | 所属楼层 |
| code | VARCHAR(32) | NOT NULL | 节点编码，如 `F1-NW`、`F1-101` |
| name | VARCHAR(64) | NOT NULL | 节点名称 |
| type | VARCHAR(16) | NOT NULL | `corridor` / `core` / `door` |
| x | DECIMAL(8,2) | NOT NULL | 节点 X |
| z | DECIMAL(8,2) | NOT NULL | 节点 Z |
| ref_type | VARCHAR(16) | NULL | 引用类型：`room`（仅 door）/ `facility`（door/core/corridor） |
| ref_code | VARCHAR(32) | NULL | 引用的房号或设施编码 |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |

索引：`UNIQUE (floor_id, code)`，`UNIQUE (floor_id, ref_type, ref_code)`，`INDEX (floor_id)`，`INDEX (type)`。
说明：房间引用只能出现在 door 节点；设施引用可出现在 door/core/corridor 节点，作为该设施的接入点。每个设施必须有且仅有一个接入节点；无引用的节点 `ref_type/ref_code` 必须为 NULL。MySQL 唯一索引允许多个 NULL，因此不会冲突。

### 2.6 graph_edges（路径边）

| 字段 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | BIGINT UNSIGNED | PK | 主键 |
| from_node_id | BIGINT UNSIGNED | NOT NULL, FK | 起点节点 |
| to_node_id | BIGINT UNSIGNED | NOT NULL, FK | 终点节点 |
| distance_m | DECIMAL(8,2) | NOT NULL | 实测距离（米） |
| edge_type | VARCHAR(16) | NOT NULL | `corridor` / `stairs` / `elevator` |
| bidirectional | TINYINT(1) | NOT NULL DEFAULT 1 | 是否双向 |
| created_at | DATETIME | NOT NULL | 创建时间 |
| updated_at | DATETIME | NOT NULL | 更新时间 |

索引：`UNIQUE (from_node_id, to_node_id, edge_type)`，`INDEX (from_node_id)`，`INDEX (to_node_id)`，`INDEX (edge_type)`。

## 3. 关系与约束

- `floors.building_id` → `buildings.id`
- `rooms.floor_id` → `floors.id`
- `facilities.floor_id` → `floors.id`
- `graph_nodes.floor_id` → `floors.id`
- `graph_edges.from_node_id` / `to_node_id` → `graph_nodes.id`
- 房间与设施不直接外键到节点；door 节点通过 `ref_type + ref_code` 关联同层房间，设施通过 door/core/corridor 节点上的 `ref_type=facility` 关联接入节点，导入阶段负责生成并校验唯一性。

## 4. Redis 使用设计（后续阶段实现）

| Key | 类型 | TTL | 用途 |
|---|---|---|---|
| `c5:floor:{floor_id}:layout` | String(JSON) | 600s | 楼层房间、设施与走廊布局 |
| `c5:search:{sha1(keyword+floor+type)}` | String(JSON) | 300s | 搜索结果缓存 |
| `c5:route:{from}:{to}:{pref}` | String(JSON) | 600s | 路线结果缓存 |

数据导入或后台更新后，按前缀清理 `c5:` 相关缓存。

## 5. 待确认

- 是否需要为房间/设施增加 `floor_label` 等冗余展示字段；MVP 暂不需要。
- `distance_m` 是否允许 0（同节点）由导入校验阶段决定，数据库暂不限制。
