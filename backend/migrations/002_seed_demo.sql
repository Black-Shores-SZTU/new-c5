-- 仅用于本地 Compose 启动与冒烟测试的最小演示数据（F1–F2）。
-- 不是 C5 实测数据；真实数据请后续通过实测 CSV 导入流程覆盖。

SET NAMES utf8mb4;

INSERT INTO buildings (id, code, name, full_name) VALUES
    (1, 'C5', 'C5栋', 'C5 公共教学楼');

INSERT INTO floors (id, building_id, floor_code, name, level, min_x, max_x, min_z, max_z, floor_height) VALUES
    (1, 1, 'F1', '一层', 1, -20.00, 20.00, -10.00, 20.00, 4.00),
    (2, 1, 'F2', '二层', 2, -20.00, 20.00, -10.00, 20.00, 4.00);

INSERT INTO rooms (id, floor_id, room_no, name, type, corridor, center_x, center_z, size_x, size_z, door_x, door_z, open_side) VALUES
    (1, 1, '101', '101', 'classroom', 'north', 0.00, 0.00, 4.00, 4.00, 0.00, 0.00, 'south'),
    (2, 2, '201', '201', 'classroom', 'north', 0.00, 0.00, 4.00, 4.00, 0.00, 0.00, 'south');

INSERT INTO facilities (id, floor_id, code, name, type, corridor, center_x, center_z, size_x, size_z, door_x, door_z) VALUES
    (1, 1, 'C5-F1-EXIT', '安全出口', 'exit', 'north', 0.00, 2.00, 2.00, 1.00, 0.00, 2.00),
    (2, 1, 'C5-F1-RESTROOM', '卫生间', 'restroom', 'north', 0.00, 6.00, 2.00, 2.00, 0.00, 6.00),
    (3, 1, 'C5-F1-STAIRS', '楼梯', 'stairs', 'NW', 0.00, 10.00, 3.00, 3.00, 0.00, 10.00),
    (4, 1, 'C5-F1-ELEVATOR', '电梯', 'elevator', 'NE', 9.00, 2.00, 2.00, 2.00, 9.00, 2.00),
    (5, 1, 'C5-F1-WATER', '开水间', 'water', 'north', -2.00, 2.00, 2.00, 2.00, -2.00, 2.00),
    (6, 2, 'C5-F2-EXIT', '安全出口', 'exit', 'north', 0.00, 2.00, 2.00, 1.00, 0.00, 2.00),
    (7, 2, 'C5-F2-RESTROOM', '卫生间', 'restroom', 'north', 0.00, 6.00, 2.00, 2.00, 0.00, 6.00),
    (8, 2, 'C5-F2-STAIRS', '楼梯', 'stairs', 'NW', 0.00, 10.00, 3.00, 3.00, 0.00, 10.00),
    (9, 2, 'C5-F2-ELEVATOR', '电梯', 'elevator', 'NE', 9.00, 2.00, 2.00, 2.00, 9.00, 2.00),
    (10, 2, 'C5-F2-WATER', '开水间', 'water', 'north', -2.00, 2.00, 2.00, 2.00, -2.00, 2.00);

INSERT INTO graph_nodes (id, floor_id, code, name, type, x, z, ref_type, ref_code) VALUES
    (1, 1, 'F1-D101', '101门', 'door', 0.00, 0.00, 'room', '101'),
    (2, 1, 'F1-C1', 'F1走廊1', 'corridor', 0.00, 2.00, 'facility', 'C5-F1-EXIT'),
    (3, 1, 'F1-C2', 'F1走廊2', 'corridor', 0.00, 6.00, 'facility', 'C5-F1-RESTROOM'),
    (4, 1, 'F1-S', 'F1楼梯', 'core', 0.00, 10.00, 'facility', 'C5-F1-STAIRS'),
    (5, 1, 'F1-E', 'F1电梯', 'core', 9.00, 2.00, 'facility', 'C5-F1-ELEVATOR'),
    (6, 2, 'F2-D201', '201门', 'door', 0.00, 0.00, 'room', '201'),
    (7, 2, 'F2-C1', 'F2走廊1', 'corridor', 0.00, 2.00, 'facility', 'C5-F2-EXIT'),
    (8, 2, 'F2-C2', 'F2走廊2', 'corridor', 0.00, 6.00, 'facility', 'C5-F2-RESTROOM'),
    (9, 2, 'F2-S', 'F2楼梯', 'core', 0.00, 10.00, 'facility', 'C5-F2-STAIRS'),
    (10, 2, 'F2-E', 'F2电梯', 'core', 9.00, 2.00, 'facility', 'C5-F2-ELEVATOR'),
    (11, 1, 'F1-W', 'F1开水间', 'door', -2.00, 2.00, 'facility', 'C5-F1-WATER'),
    (12, 2, 'F2-W', 'F2开水间', 'door', -2.00, 2.00, 'facility', 'C5-F2-WATER');

INSERT INTO graph_edges (id, from_node_id, to_node_id, edge_type, distance_m, bidirectional) VALUES
    (1, 1, 2, 'corridor', 2.00, 1),
    (2, 2, 3, 'corridor', 4.00, 1),
    (3, 3, 4, 'corridor', 4.00, 1),
    (4, 2, 5, 'corridor', 9.00, 1),
    (5, 6, 7, 'corridor', 2.00, 1),
    (6, 7, 8, 'corridor', 4.00, 1),
    (7, 8, 9, 'corridor', 4.00, 1),
    (8, 7, 10, 'corridor', 9.00, 1),
    (9, 4, 9, 'stairs', 4.00, 1),
    (10, 5, 10, 'elevator', 4.00, 1),
    (11, 11, 2, 'corridor', 2.00, 1),
    (12, 12, 7, 'corridor', 2.00, 1);
