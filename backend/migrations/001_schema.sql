-- C5 室内导航 MVP 表结构（MySQL 8.0，utf8mb4）
-- 由 Docker Compose 首次初始化时自动执行。

SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS buildings (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    code       VARCHAR(32)     NOT NULL,
    name       VARCHAR(64)     NOT NULL,
    full_name  VARCHAR(128)    NOT NULL,
    created_at DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_buildings_code (code)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS floors (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    building_id  BIGINT UNSIGNED NOT NULL,
    floor_code   VARCHAR(8)      NOT NULL,
    name         VARCHAR(32)     NOT NULL,
    level        INT             NOT NULL,
    min_x        DECIMAL(8, 2)   NOT NULL,
    max_x        DECIMAL(8, 2)   NOT NULL,
    min_z        DECIMAL(8, 2)   NOT NULL,
    max_z        DECIMAL(8, 2)   NOT NULL,
    floor_height DECIMAL(5, 2)   NOT NULL,
    created_at   DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at   DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_building_level (building_id, level),
    KEY idx_floors_building_id (building_id),
    CONSTRAINT fk_floors_building FOREIGN KEY (building_id) REFERENCES buildings (id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS rooms (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    floor_id   BIGINT UNSIGNED NOT NULL,
    room_no    VARCHAR(16)     NOT NULL,
    name       VARCHAR(64)     NOT NULL,
    type       VARCHAR(32)     NOT NULL,
    corridor   VARCHAR(16)     NOT NULL,
    center_x   DECIMAL(8, 2)   NOT NULL,
    center_z   DECIMAL(8, 2)   NOT NULL,
    size_x     DECIMAL(8, 2)   NOT NULL,
    size_z     DECIMAL(8, 2)   NOT NULL,
    door_x     DECIMAL(8, 2)   NOT NULL,
    door_z     DECIMAL(8, 2)   NOT NULL,
    open_side  VARCHAR(8)      NOT NULL,
    created_at DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_floor_room (floor_id, room_no),
    KEY idx_rooms_floor_id (floor_id),
    KEY idx_rooms_type (type),
    CONSTRAINT fk_rooms_floor FOREIGN KEY (floor_id) REFERENCES floors (id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS facilities (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    floor_id   BIGINT UNSIGNED NOT NULL,
    code       VARCHAR(32)     NOT NULL,
    name       VARCHAR(64)     NOT NULL,
    type       VARCHAR(32)     NOT NULL,
    corridor   VARCHAR(16)     NOT NULL,
    center_x   DECIMAL(8, 2)   NOT NULL,
    center_z   DECIMAL(8, 2)   NOT NULL,
    size_x     DECIMAL(8, 2)   NOT NULL,
    size_z     DECIMAL(8, 2)   NOT NULL,
    door_x     DECIMAL(8, 2)   NOT NULL,
    door_z     DECIMAL(8, 2)   NOT NULL,
    created_at DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_floor_facility (floor_id, code),
    KEY idx_facilities_floor_id (floor_id),
    KEY idx_facilities_type (type),
    CONSTRAINT fk_facilities_floor FOREIGN KEY (floor_id) REFERENCES floors (id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS graph_nodes (
    id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    floor_id   BIGINT UNSIGNED NOT NULL,
    code       VARCHAR(32)     NOT NULL,
    name       VARCHAR(64)     NOT NULL,
    type       VARCHAR(16)     NOT NULL,
    x          DECIMAL(8, 2)   NOT NULL,
    z          DECIMAL(8, 2)   NOT NULL,
    ref_type   VARCHAR(16)     NULL,
    ref_code   VARCHAR(32)     NULL,
    created_at DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_floor_node (floor_id, code),
    UNIQUE KEY uk_floor_ref (floor_id, ref_type, ref_code),
    KEY idx_graph_nodes_floor_id (floor_id),
    KEY idx_graph_nodes_type (type),
    CONSTRAINT fk_graph_nodes_floor FOREIGN KEY (floor_id) REFERENCES floors (id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS graph_edges (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    from_node_id  BIGINT UNSIGNED NOT NULL,
    to_node_id    BIGINT UNSIGNED NOT NULL,
    edge_type     VARCHAR(16)     NOT NULL,
    distance_m    DECIMAL(8, 2)   NOT NULL,
    bidirectional TINYINT(1)      NOT NULL DEFAULT 1,
    created_at    DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at    DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_edge_pair (from_node_id, to_node_id, edge_type),
    KEY idx_graph_edges_from (from_node_id),
    KEY idx_graph_edges_to (to_node_id),
    CONSTRAINT fk_graph_edges_from FOREIGN KEY (from_node_id) REFERENCES graph_nodes (id),
    CONSTRAINT fk_graph_edges_to FOREIGN KEY (to_node_id) REFERENCES graph_nodes (id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci;
