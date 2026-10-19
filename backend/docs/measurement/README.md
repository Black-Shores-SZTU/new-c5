# C5 实测 CSV 填写说明

## 1. 范围

本目录用于 C5 教学楼 F1–F5 的真实数据采集。MVP 只做 C5 一栋楼，不测楼间路网，不做用户系统，不接实时定位。

填写完成后，后续阶段统一导入 MySQL，生成 `buildings`、`floors`、`rooms`、`facilities`、`graph_nodes`、`graph_edges`。

## 2. 文件清单

| 文件 | 内容 | 目标表 |
|---|---|---|
| `building-attributes-template.csv` | 楼栋、楼层、层高、边界、原点、朝向、出入口与服务楼层汇总 | `buildings` + `floors` |
| `facilities-template.csv` | 电梯、楼梯、卫生间、出入口、开水间、安全出口、演播厅 | `facilities` |
| `bpoints-template.csv` | 走廊节点与设施接入节点 | `graph_nodes` |
| `bpaths-template.csv` | 走廊、楼梯、电梯之间的通行边 | `graph_edges` |
| `classrooms-template.csv` | 教室、教师休息室、演播厅 | `rooms` |
| `cpaths-template.csv` | 教室门到相邻走廊节点的连接 | `graph_nodes`（door）+ `graph_edges` |

所有 CSV 均为 UTF-8 with BOM，第一行为表头，第二行为示例行；示例行中的测量值留空，仅用于说明格式。

## 3. 坐标、编码与精度

### 3.1 坐标原点与方向

- 原点：C5 建筑中心，局部坐标 `(0,0)`。
- `x_m`：向东为正。
- `y_m`：向南为正，与数据库 `z` 一致；北长廊 `y_m` 为负。
- `origin_x_m` / `origin_y_m`：局部原点在校园或场地参考系中的位置；MVP 无校园坐标系时填 `0.00`，仅作为元数据。
- `orientation_deg`：图纸上方相对真北的顺时针角度；MVP 仅记录，不参与寻路。

### 3.2 精度

- 坐标：米，保留两位小数。
- 距离：米，保留一位小数。
- 所有距离必须大于 0；同节点距离 0 不允许。

### 3.3 编码

- 节点编号一经确定不得更改，不得复用；新增节点使用新序号。
- 走廊节点：`C5-F{楼层}-{ZONE}-{序号}`，如 `C5-F1-N-01`。
- 设施接入节点：`C5-F{楼层}-F-{TYPE}-{序号}`，如 `C5-F1-F-ELEV-01`。
- 房间门节点：`C5-F{楼层}-D-{房号}`，如 `C5-F1-D-101`。
- `ZONE` 取值：`N`、`MA`、`MB`、`S`、`W`、`E`、`NW`、`NE`、`MAW`、`MAE`、`MBW`、`MBE`、`SW`、`SE`。
- 枚举值一律使用小写英文。
- 多个值用 `|` 分隔，不要使用逗号。

## 4. 现场测量顺序

1. 先测 F1：建筑属性、设施、BPoint、BPath、Classroom、CPath。
2. 用 F1 数据跑通一次导入与路线规划。
3. 确认编号、坐标方向和距离口径无误后，再铺 F2–F5。
4. 每层完成后做一次自检，避免最后集中返工。

## 5. building-attributes-template.csv

```csv
building_code,building_name,full_name,floor_count,floor_code,floor_name,level,floor_height_m,origin_x_m,origin_y_m,orientation_deg,min_x_m,max_x_m,min_y_m,max_y_m,entrance_codes,elevator_served_floors,stairs_served_floors,note
```

| 字段 | 必填 | 说明 |
|---|---|---|
| building_code | 是 | 固定 `C5` |
| building_name | 是 | `C5栋` |
| full_name | 是 | `C5 公共教学楼` |
| floor_count | 是 | 地上层数，MVP 为 5 |
| floor_code | 是 | `F1`–`F5` |
| floor_name | 是 | `一层`–`五层` |
| level | 是 | 1–5 |
| floor_height_m | 是 | 层高，米 |
| origin_x_m / origin_y_m | 是 | 局部原点在场地参考系的位置；无参考系填 0.00 |
| orientation_deg | 是 | 图纸上方相对真北的顺时针角度 |
| min_x_m / max_x_m | 是 | 楼层 X 方向边界 |
| min_y_m / max_y_m | 是 | 楼层 Y 方向边界，入库映射为 `min_z/max_z` |
| entrance_codes | 否 | 出入口设施编码，多个用 `|` |
| elevator_served_floors | 否 | 电梯服务楼层，多个用 `|` |
| stairs_served_floors | 否 | 楼梯服务楼层，多个用 `|` |
| note | 否 | 备注 |

一行一层；建筑级字段在每一行重复填写。

## 6. facilities-template.csv

```csv
floor,facility_code,facility_name,facility_type,corridor,center_x_m,center_y_m,size_x_m,size_y_m,access_x_m,access_y_m,served_floors,note
```

| 字段 | 必填 | 说明 |
|---|---|---|
| floor | 是 | `F1`–`F5` |
| facility_code | 是 | 同层唯一，如 `C5-F1-F-ELEV-01` |
| facility_name | 是 | 显示名称 |
| facility_type | 是 | `elevator` / `stairs` / `restroom` / `entrance` / `water` / `auditorium` / `exit` |
| corridor | 是 | 临近走廊或核心筒编码 |
| center_x_m / center_y_m | 是 | 设施几何中心 |
| size_x_m / size_y_m | 是 | 设施尺寸 |
| access_x_m / access_y_m | 是 | 主接入点/门中心；BPoint 设施节点必须使用同一坐标 |
| served_floors | 否 | 电梯/楼梯服务楼层，多个用 `|`；其他设施留空 |
| note | 否 | 备注 |

每个设施在 MVP 只保留一个主接入点；多层电梯/楼梯每一层都要有一条设施记录。

## 7. bpoints-template.csv

```csv
node_code,node_name,node_type,floor,x_m,y_m,served_floors,ref_type,ref_code,note
```

| 字段 | 必填 | 说明 |
|---|---|---|
| node_code | 是 | 节点编号，按第 3.3 节约定 |
| node_name | 是 | 节点名称 |
| node_type | 是 | `corridor_turn` / `corridor_junction` / `corridor_end` / `elevator` / `stairs` / `entrance` / `restroom` / `water` / `exit` |
| floor | 是 | `F1`–`F5` |
| x_m / y_m | 是 | 沿人走中线的节点坐标 |
| served_floors | 否 | 电梯/楼梯节点填写，多个用 `|` |
| ref_type | 否 | 设施接入节点填 `facility`；走廊节点留空 |
| ref_code | 否 | 对应 `facility_code` |
| note | 否 | 备注 |

类型入库映射：

- `corridor_turn` / `corridor_junction` / `corridor_end` → `graph_nodes.type=corridor`
- `elevator` / `stairs` → `graph_nodes.type=core`
- `entrance` / `restroom` / `water` / `exit` → `graph_nodes.type=corridor`，并填 `ref_type=facility`
- 房间门节点不在此表填写，由 CPath 生成。

## 8. bpaths-template.csv

```csv
floor_from,floor_to,from_node_code,to_node_code,distance_m,edge_type,bidirectional,note
```

| 字段 | 必填 | 说明 |
|---|---|---|
| floor_from / floor_to | 是 | 同层填相同楼层；跨层填起止楼层 |
| from_node_code / to_node_code | 是 | BPoint 中已存在的节点编码 |
| distance_m | 是 | 沿人走中线实测，0.1 米精度 |
| edge_type | 是 | `corridor` / `stairs` / `elevator` |
| bidirectional | 是 | `TRUE` / `FALSE` |
| note | 否 | 备注 |

规则：

- 双向边只填一行，`bidirectional=TRUE`；不要额外填反向行，否则会触发重复反向边校验。
- 现场需要正向、反向各走一次确认可通，但只记录一条边。
- 同层走廊边 `floor_from = floor_to`。
- 跨层只允许楼梯/电梯，且两端节点必须在对应楼层存在。
- 楼梯距离沿梯段中线测量；电梯距离按层高×层数或实测井道行程填写。
- 不测跨楼边。

## 9. classrooms-template.csv

```csv
floor,room_no,room_name,room_type,corridor,center_x_m,center_y_m,size_x_m,size_y_m,door_x_m,door_y_m,open_side,note
```

| 字段 | 必填 | 说明 |
|---|---|---|
| floor | 是 | `F1`–`F5` |
| room_no | 是 | 房号，同层唯一 |
| room_name | 是 | 显示名称，默认同房号 |
| room_type | 是 | `classroom` / `teacher_rest` / `auditorium` |
| corridor | 是 | 临近走廊编码 |
| center_x_m / center_y_m | 是 | 房间中心坐标 |
| size_x_m / size_y_m | 是 | 房间尺寸 |
| door_x_m / door_y_m | 是 | 门洞中心坐标 |
| open_side | 是 | `north` / `south` / `east` / `west` |
| note | 否 | 备注 |

## 10. cpaths-template.csv

```csv
floor,room_no,door_node_code,adjacent_node_code,distance_m,note
```

| 字段 | 必填 | 说明 |
|---|---|---|
| floor | 是 | `F1`–`F5` |
| room_no | 是 | 对应 Classroom 房号 |
| door_node_code | 是 | 门节点编号，按 `C5-F1-D-101` 约定 |
| adjacent_node_code | 是 | BPoint 中的走廊节点编码 |
| distance_m | 是 | 门中心到相邻节点，沿人走线，0.1 米精度 |
| note | 否 | 备注 |

规则：

- 门节点坐标直接取 Classroom 的 `door_x_m/door_y_m`，不重复测量。
- 导入时由 CPath 生成 `type=door` 的 `graph_nodes`，并生成门到相邻节点的 `graph_edges`。
- 一个房间有多个门或多个相邻节点时，填多行。
- `adjacent_node_code` 必须是同层 BPoint 走廊节点。

## 11. 导入映射

| CSV | 目标表/字段 |
|---|---|
| building-attributes | `buildings` + `floors`；原点、朝向仅作为元数据 |
| facilities | `facilities` |
| bpoints | `graph_nodes`；语义类型映射为 `corridor/core/door` |
| bpaths | `graph_edges` |
| classrooms | `rooms` |
| cpaths | 生成 `graph_nodes`（door）+ `graph_edges`（door→相邻节点） |

现有表没有原点、朝向、`served_floors` 字段；这三类信息在 CSV 中保留，不直接入库。

## 12. 填写后自检

1. 表头未修改，每行列数与表头一致。
2. 所有楼层在 `F1`–`F5` 内，不出现楼间数据。
3. 同层房号、设施编码、节点编号无重复。
4. BPath 的节点都能在 BPoint 中找到；CPath 的相邻节点都能在 BPoint 中找到。
5. 每个教室至少有一条 CPath；每个设施恰好有一个 BPoint 接入节点。
6. 双向边只有一行且 `bidirectional=TRUE`，没有重复反向行。
7. 跨层边只有 `stairs/elevator`，且 `served_floors` 与跨层边一致。
8. 所有距离大于 0；坐标在楼层边界内。
9. BPoint 设施节点坐标与 facilities 的 `access_x_m/access_y_m` 一致。
10. 北长廊 `y_m` 为负；没有混用“北为正”的坐标。

## 13. 常见错误

- 沿墙量距离，而不是沿人走中线。
- 双向边填了两行反向记录。
- 把房间门节点也填进 BPoint。
- 电梯边距离填 0。
- 节点编号中途改号或复用。
- 坐标 y 轴方向与数据库 Z 不一致。
- 备注中使用英文逗号导致列错位；备注如含英文逗号，需要用双引号包裹。
