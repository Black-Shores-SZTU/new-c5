// ============================================================
// C5 公共教学楼 · 数据驱动楼层配置
// 依据 C-5 各楼层教室分布图整理；坐标为相对占位，待后端节点数据替换
// 仅服务于 P2/P3 的 Three.js 场景与 mock 路线，不影响 UI/DOM
// ============================================================

export const C5_BUILDING_ID = 'A'
export const C5_BUILDING_NAME = 'C5栋'
export const C5_BUILDING_FULL_NAME = 'C5 公共教学楼'

// TODO: 下列房间/走廊/核心筒坐标均为按图纸比例的相对占位值，
// 后端节点数据接入后需整体替换 center/door/x1..z2 等几何字段（保持 id 与邻接关系不变）。
export const C5_BOUNDS = {
  minX: -44,
  maxX: 44,
  minZ: -33,
  maxZ: 33
}

export const C5_COLORS = {
  classroomFloor: '#E8F1FF',
  classroomAccent: '#2B6BE5',
  teacherFloor: '#FFF1E2',
  teacherAccent: '#B7791F',
  elevatorFloor: '#FCE8E8',
  elevatorAccent: '#D92D20',
  restroomFloor: '#E6F7F4',
  restroomAccent: '#0E9F8F',
  auditoriumFloor: '#F1E9FF',
  auditoriumAccent: '#7C5CFC',
  stairsFloor: '#EEF2F7',
  stairsAccent: '#64748B',
  corridor: '#EDF2F7',
  slab: '#F4F7FB',
  wall: '#FFFFFF',
  edge: '#C7D4E4',
  text: '#334155',
  plate: 'rgba(255,255,255,.92)'
}

const FLOOR_LEVELS = {
  F1: 1,
  F2: 2,
  F3: 3,
  F4: 4,
  F5: 5
}

const range = (start, end) =>
  Array.from({ length: end - start + 1 }, (_, index) => String(start + index))

const TEACHER_REST = {
  F1: new Set(),
  F2: new Set(['203', '229']),
  F3: new Set(['303', '343', '363']),
  F4: new Set(['403', '421', '443', '463', '467']),
  F5: new Set(['503', '521', '540'])
}

// 走廊中心线与矩形范围
export const C5_CORRIDORS = {
  north: { id: 'north', name: '北长廊', axis: 'x', x1: -30, x2: 34, z1: -24.5, z2: -21.5, centerZ: -23 },
  midA: { id: 'midA', name: '中间长廊A', axis: 'x', x1: -30, x2: 34, z1: -9.5, z2: -6.5, centerZ: -7 },
  midB: { id: 'midB', name: '中间长廊B', axis: 'x', x1: -30, x2: 34, z1: 5.5, z2: 8.5, centerZ: 9 },
  south: { id: 'south', name: '南长廊', axis: 'x', x1: -30, x2: 34, z1: 20.5, z2: 23.5, centerZ: 25 },
  west: { id: 'west', name: '西长廊', axis: 'z', z1: -23, z2: 25, x1: -34, x2: -30, centerX: -32 },
  east: { id: 'east', name: '东长廊', axis: 'z', z1: -23, z2: 25, x1: 34, x2: 38, centerX: 36 }
}

// 房间行；z 范围已为背靠背房间留出结构缝
const ROWS = {
  northNorth: { z1: -33, z2: -24.5, corridor: 'north', openSide: 'south', anchorZ: -23 },
  northSouth: { z1: -21.5, z2: -16.2, corridor: 'north', openSide: 'north', anchorZ: -23 },
  midANorth: { z1: -14.8, z2: -9.5, corridor: 'midA', openSide: 'south', anchorZ: -7 },
  midASouth: { z1: -6.5, z2: -1.4, corridor: 'midA', openSide: 'north', anchorZ: -7 },
  midBNorth: { z1: 0.2, z2: 5.5, corridor: 'midB', openSide: 'south', anchorZ: 9 },
  midBSouth: { z1: 8.5, z2: 13.6, corridor: 'midB', openSide: 'north', anchorZ: 9, x2: 2 },
  southNorth: { z1: 14.6, z2: 20.5, corridor: 'south', openSide: 'south', anchorZ: 25, x2: 2 },
  southSouth: { z1: 23.5, z2: 29.5, corridor: 'south', openSide: 'north', anchorZ: 25 }
}

export const C5_CORE_NODES = [
  { id: 'NW', x: -32, z: -23, side: 'west' },
  { id: 'NE', x: 36, z: -23, side: 'east' },
  { id: 'MAW', x: -32, z: -7, side: 'west' },
  { id: 'MAE', x: 36, z: -7, side: 'east' },
  { id: 'MBW', x: -32, z: 9, side: 'west' },
  { id: 'MBE', x: 36, z: 9, side: 'east' },
  { id: 'SW', x: -32, z: 25, side: 'west' },
  { id: 'SE', x: 36, z: 25, side: 'east' }
]

const FLOOR_GROUPS = {
  F1: {
    north: { north: ['103', '102', '101'], south: ['105', '104'], reverse: false },
    west: { upper: ['106', '107'], lower: ['108', '109'] },
    midB: { north: ['110', '111A'], south: ['111B', '111'] },
    east: { upper: ['118', '117', '116', '115'], lower: ['114', '113', '112'], variant: 'one' },
    midA: { north: ['120', '119'], south: ['122', '121'], reverse: false },
    south: { north: ['123', '124', '125', '126', '127'], south: ['128', '129', '130', '131', '132'] }
  },
  F2: {
    north: { north: range(201, 207), south: range(208, 214), reverse: true },
    midB: { north: range(215, 222), south: range(223, 229) },
    east: { upper: ['231'], lower: ['230'] },
    midA: { north: range(232, 238), south: range(239, 244), reverse: true },
    south: { north: range(245, 247), south: range(248, 250) }
  },
  F3: {
    north: { north: range(301, 308), south: range(309, 315), reverse: true },
    west: { upper: range(316, 320), lower: range(321, 325) },
    midB: { north: range(326, 333), south: range(334, 340) },
    east: { upper: ['342'], lower: ['341'] },
    midA: { north: range(343, 349), south: range(350, 356), reverse: true },
    south: { north: range(357, 359), south: range(360, 369) }
  },
  F4: {
    north: { north: range(401, 408), south: range(409, 415), reverse: true },
    west: { upper: range(416, 420), lower: range(421, 425) },
    midB: { north: range(426, 433), south: range(434, 440) },
    east: { upper: ['442'], lower: ['441'] },
    midA: { north: range(443, 449), south: range(450, 456), reverse: true },
    south: { north: range(457, 459), south: range(460, 469) }
  },
  F5: {
    north: { north: range(501, 508), south: range(509, 515), reverse: true },
    west: { upper: range(516, 520), lower: range(521, 525) },
    midB: { north: range(526, 533), south: range(534, 540) },
    east: { upper: ['542'], lower: ['541'] },
    midA: { north: range(543, 549), south: range(550, 556), reverse: true },
    south: { north: range(557, 559), south: range(560, 572) }
  }
}

const EAST_ROWS = {
  one: {
    upper: { z1: -21.5, z2: -9.5, openSide: 'west', x1: 38, x2: 44, anchorX: 36 },
    lower: { z1: 11.5, z2: 20.5, openSide: 'west', x1: 38, x2: 44, anchorX: 36 }
  },
  default: {
    upper: { z1: -5, z2: 0, openSide: 'west', x1: 38, x2: 44, anchorX: 36 },
    lower: { z1: 0, z2: 5.5, openSide: 'west', x1: 38, x2: 44, anchorX: 36 }
  }
}

const WEST_ROWS = {
  upper: { z1: -21.5, z2: -9.5, openSide: 'east', x1: -44, x2: -34, anchorX: -32 },
  lower: { z1: 11.5, z2: 22.5, openSide: 'east', x1: -44, x2: -34, anchorX: -32 }
}

const GAP = 0.2
const floorDataCache = new Map()
const graphCache = new Map()

const makeRoom = (floorId, roomNo, type, corridor, center, size, door, anchor, openSide) => ({
  id: roomNo,
  roomNo,
  name: roomNo,
  type,
  building: C5_BUILDING_ID,
  floor: floorId,
  floorLevel: FLOOR_LEVELS[floorId],
  corridor,
  center,
  size,
  door,
  anchor,
  openSide,
  graphNodeId: null
})

const buildHorizontalRow = (floorId, corridorKey, rowKey, rowConfig, roomNos, reverse) => {
  if (!roomNos?.length) {
    return []
  }
  const row = ROWS[rowKey]
  const x1 = -30
  const x2 = row.x2 ?? 34
  const roomWidth = (x2 - x1 - GAP * (roomNos.length - 1)) / roomNos.length
  const depth = row.z2 - row.z1 - GAP
  const zCenter = (row.z1 + row.z2) / 2
  const teacherSet = TEACHER_REST[floorId]

  return roomNos.map((roomNo, index) => {
    const slot = reverse ? roomNos.length - 1 - index : index
    const x = x1 + roomWidth / 2 + slot * (roomWidth + GAP)
    const door =
      row.openSide === 'south'
        ? [x, 0.12, row.z2 - 0.02]
        : [x, 0.12, row.z1 + 0.02]
    const type = teacherSet.has(roomNo) ? 'teacher_rest' : 'classroom'
    return makeRoom(
      floorId,
      roomNo,
      type,
      corridorKey,
      [x, 0, zCenter],
      [roomWidth, depth],
      door,
      [x, 0.12, row.anchorZ],
      row.openSide
    )
  })
}

const buildVerticalRow = (floorId, corridorKey, rowConfig, roomNos, anchorX) => {
  if (!roomNos?.length) {
    return []
  }
  const roomDepth = rowConfig.x2 - rowConfig.x1 - GAP
  const zSpan = rowConfig.z2 - rowConfig.z1
  const roomWidth = (zSpan - GAP * (roomNos.length - 1)) / roomNos.length
  const xCenter = (rowConfig.x1 + rowConfig.x2) / 2
  const teacherSet = TEACHER_REST[floorId]

  return roomNos.map((roomNo, index) => {
    const z = rowConfig.z1 + roomWidth / 2 + index * (roomWidth + GAP)
    const door = [rowConfig.openSide === 'east' ? rowConfig.x2 - 0.02 : rowConfig.x1 + 0.02, 0.12, z]
    const type = teacherSet.has(roomNo) ? 'teacher_rest' : 'classroom'
    return makeRoom(
      floorId,
      roomNo,
      type,
      corridorKey,
      [xCenter, 0, z],
      [roomDepth, roomWidth],
      door,
      [anchorX, 0.12, z],
      rowConfig.openSide
    )
  })
}

const buildRooms = (floorId) => {
  const groups = FLOOR_GROUPS[floorId]
  const rooms = []

  if (groups.north) {
    rooms.push(
      ...buildHorizontalRow(floorId, 'north', 'northNorth', groups.north, groups.north.north, groups.north.reverse)
    )
    rooms.push(
      ...buildHorizontalRow(floorId, 'north', 'northSouth', groups.north, groups.north.south, groups.north.reverse)
    )
  }
  if (groups.midA) {
    rooms.push(
      ...buildHorizontalRow(floorId, 'midA', 'midANorth', groups.midA, groups.midA.north, groups.midA.reverse)
    )
    rooms.push(
      ...buildHorizontalRow(floorId, 'midA', 'midASouth', groups.midA, groups.midA.south, groups.midA.reverse)
    )
  }
  if (groups.midB) {
    rooms.push(
      ...buildHorizontalRow(floorId, 'midB', 'midBNorth', groups.midB, groups.midB.north, false)
    )
    rooms.push(
      ...buildHorizontalRow(floorId, 'midB', 'midBSouth', groups.midB, groups.midB.south, false)
    )
  }
  if (groups.south) {
    rooms.push(
      ...buildHorizontalRow(floorId, 'south', 'southNorth', groups.south, groups.south.north, false)
    )
    rooms.push(
      ...buildHorizontalRow(floorId, 'south', 'southSouth', groups.south, groups.south.south, false)
    )
  }
  if (groups.west) {
    rooms.push(
      ...buildVerticalRow(floorId, 'west', WEST_ROWS.upper, groups.west.upper, WEST_ROWS.upper.anchorX)
    )
    rooms.push(
      ...buildVerticalRow(floorId, 'west', WEST_ROWS.lower, groups.west.lower, WEST_ROWS.lower.anchorX)
    )
  }
  if (groups.east) {
    const eastRows = EAST_ROWS[groups.east.variant || 'default']
    rooms.push(
      ...buildVerticalRow(floorId, 'east', eastRows.upper, groups.east.upper, eastRows.upper.anchorX)
    )
    rooms.push(
      ...buildVerticalRow(floorId, 'east', eastRows.lower, groups.east.lower, eastRows.lower.anchorX)
    )
  }

  return rooms
}

const getCoreCenter = (node, type) => {
  if (type === 'restroom') {
    return node.side === 'west'
      ? [-36.5, 0, node.z]
      : [39.5, 0, node.z]
  }
  const x = node.side === 'west' ? -40 : 41
  const z = type === 'elevator' ? node.z - 2 : node.z + 2
  return [x, 0, z]
}

const makeFacility = (floorId, id, name, type, center, size, corridor, extra = {}) => ({
  id,
  name,
  type,
  building: C5_BUILDING_ID,
  floor: floorId,
  floorLevel: FLOOR_LEVELS[floorId],
  corridor,
  center,
  door: [...center],
  anchor: null,
  size,
  graphNodeId: null,
  ...extra
})

const buildFacilities = (floorId) => {
  const facilities = []
  C5_CORE_NODES.forEach((node) => {
    const base = `C5-${floorId}-${node.id}`
    ;['elevator', 'stairs', 'restroom'].forEach((type) => {
      const center = getCoreCenter(node, type)
      const name = type === 'elevator' ? '电梯' : type === 'stairs' ? '楼梯' : '卫生间'
      facilities.push(
        makeFacility(
          floorId,
          `${base}-${type}`,
          name,
          type,
          center,
          [3.5, 3],
          node.id,
          { nodeId: node.id, nodeX: node.x, nodeZ: node.z }
        )
      )
    })
  })

  if (floorId === 'F1') {
    facilities.push(
      makeFacility(
        'F1',
        'C5-F1-AUD',
        '演播厅',
        'auditorium',
        [17, 0, 14.5],
        [30, 12],
        'midB',
        {
          roomNo: '133-138',
          door: [17, 0.12, 8.6],
          anchor: [17, 0.12, 9],
          nodeId: 'MBE'
        }
      )
    )
    facilities.push(
      makeFacility(
        'F1',
        'C5-F1-ENTRANCE',
        'C5教学楼入口',
        'entrance',
        [41, 0, -29],
        [6, 6],
        'NE',
        { nodeId: 'NE' }
      )
    )
  }

  return facilities
}

export const getC5FloorData = (floorId) => {
  if (floorDataCache.has(floorId)) {
    return floorDataCache.get(floorId)
  }
  const rooms = buildRooms(floorId)
  const facilities = buildFacilities(floorId)
  const data = {
    floorId,
    floorLevel: FLOOR_LEVELS[floorId],
    rooms,
    facilities,
    locations: [...rooms, ...facilities],
    courtyard: floorId !== 'F1'
  }
  floorDataCache.set(floorId, data)
  return data
}

export const getC5Rooms = (floorId) => getC5FloorData(floorId).rooms
export const getC5Facilities = (floorId) => getC5FloorData(floorId).facilities
export const getC5Locations = (floorId) => getC5FloorData(floorId).locations

/* ============================================================
 * 走廊图：房间门 → 走廊锚点 → 交叉节点/核心筒
 * ============================================================ */

const addNode = (graph, id, x, z) => {
  if (!graph.nodes.has(id)) {
    graph.nodes.set(id, { id, x, z, edges: [] })
  }
  return graph.nodes.get(id)
}

const addEdge = (graph, a, b) => {
  if (!a || !b || a === b) {
    return
  }
  const first = graph.nodes.get(a)
  const second = graph.nodes.get(b)
  if (!first || !second) {
    return
  }
  const cost = Math.hypot(first.x - second.x, first.z - second.z)
  first.edges.push({ to: b, cost })
  second.edges.push({ to: a, cost })
}

const NODE_CONNECTIONS = {
  north: ['NW', 'NE'],
  midA: ['MAW', 'MAE'],
  midB: ['MBW', 'MBE'],
  south: ['SW', 'SE'],
  west: ['NW', 'MAW', 'MBW', 'SW'],
  east: ['NE', 'MAE', 'MBE', 'SE']
}

const buildGraph = (floorId) => {
  if (graphCache.has(floorId)) {
    return graphCache.get(floorId)
  }

  const data = getC5FloorData(floorId)
  const graph = {
    floorId,
    nodes: new Map(),
    coreCandidates: []
  }

  const intersectionIds = {}
  C5_CORE_NODES.forEach((node) => {
    const id = `F${data.floorLevel}-${node.id}`
    intersectionIds[node.id] = id
    addNode(graph, id, node.x, node.z)
  })

  const corridorPoints = {}
  Object.keys(NODE_CONNECTIONS).forEach((key) => {
    corridorPoints[key] = NODE_CONNECTIONS[key].map((nodeId) => intersectionIds[nodeId])
  })

  data.rooms.forEach((room) => {
    const doorId = `door-${room.id}`
    const anchorId = `anchor-${room.id}`
    addNode(graph, doorId, room.door[0], room.door[2])
    addNode(graph, anchorId, room.anchor[0], room.anchor[2])
    addEdge(graph, doorId, anchorId)
    room.graphNodeId = doorId
    corridorPoints[room.corridor].push(anchorId)
  })

  data.facilities.forEach((facility) => {
    if (facility.type === 'auditorium') {
      const doorId = `door-${facility.id}`
      const anchorId = `anchor-${facility.id}`
      addNode(graph, doorId, facility.door[0], facility.door[2])
      addNode(graph, anchorId, facility.anchor[0], facility.anchor[2])
      addEdge(graph, doorId, anchorId)
      facility.graphNodeId = doorId
      corridorPoints[facility.corridor].push(anchorId)
      return
    }

    const coreId = `core-${facility.id}`
    addNode(graph, coreId, facility.center[0], facility.center[2])
    const intersectionId = intersectionIds[facility.nodeId] || intersectionIds.NE
    addEdge(graph, coreId, intersectionId)
    facility.graphNodeId = coreId

    if (facility.type === 'elevator' || facility.type === 'stairs') {
      graph.coreCandidates.push({
        id: coreId,
        floorId,
        type: facility.type,
        nodeId: facility.nodeId,
        x: facility.center[0],
        z: facility.center[2]
      })
    }
  })

  Object.entries(corridorPoints).forEach(([key, ids]) => {
    const corridor = C5_CORRIDORS[key]
    const unique = [...new Set(ids)]
    unique.sort((firstId, secondId) => {
      const first = graph.nodes.get(firstId)
      const second = graph.nodes.get(secondId)
      return corridor.axis === 'x' ? first.x - second.x : first.z - second.z
    })
    for (let index = 1; index < unique.length; index += 1) {
      addEdge(graph, unique[index - 1], unique[index])
    }
  })

  graphCache.set(floorId, graph)
  return graph
}

const shortestPath = (graph, startId, endId) => {
  if (!graph.nodes.has(startId) || !graph.nodes.has(endId)) {
    return null
  }
  const distances = new Map()
  const previous = new Map()
  const unvisited = new Set(graph.nodes.keys())
  graph.nodes.forEach((_, id) => distances.set(id, Number.POSITIVE_INFINITY))
  distances.set(startId, 0)

  while (unvisited.size) {
    let current = null
    let best = Number.POSITIVE_INFINITY
    unvisited.forEach((id) => {
      const distance = distances.get(id)
      if (distance < best) {
        best = distance
        current = id
      }
    })
    if (!current || best === Number.POSITIVE_INFINITY) {
      break
    }
    if (current === endId) {
      break
    }
    unvisited.delete(current)
    const node = graph.nodes.get(current)
    node.edges.forEach((edge) => {
      if (!unvisited.has(edge.to)) {
        return
      }
      const next = best + edge.cost
      if (next < distances.get(edge.to)) {
        distances.set(edge.to, next)
        previous.set(edge.to, current)
      }
    })
  }

  if (!Number.isFinite(distances.get(endId))) {
    return null
  }

  const path = []
  let current = endId
  while (current) {
    path.unshift(current)
    current = previous.get(current)
  }
  return path
}

const graphPoints = (graph, ids) =>
  ids.map((id) => {
    const node = graph.nodes.get(id)
    return [Number(node.x.toFixed(2)), 0.12, Number(node.z.toFixed(2))]
  })

const withEndpoints = (points, from, to) => {
  const result = []
  const push = (point) => {
    const previous = result[result.length - 1]
    if (
      !previous ||
      Math.abs(previous[0] - point[0]) > 0.01 ||
      Math.abs(previous[2] - point[2]) > 0.01
    ) {
      result.push(point)
    }
  }
  push([from.center[0], 0.12, from.center[2]])
  points.forEach(push)
  push([to.center[0], 0.12, to.center[2]])
  return result
}

const fromCoreSegment = (points, from) => {
  const result = [[from.center[0], 0.12, from.center[2]]]
  points.forEach((point) => {
    const previous = result[result.length - 1]
    if (
      Math.abs(previous[0] - point[0]) > 0.01 ||
      Math.abs(previous[2] - point[2]) > 0.01
    ) {
      result.push(point)
    }
  })
  return result
}

const toCoreSegment = (points, to) => {
  const result = [...points]
  const previous = result[result.length - 1]
  if (
    !previous ||
    Math.abs(previous[0] - to.center[0]) > 0.01 ||
    Math.abs(previous[2] - to.center[2]) > 0.01
  ) {
    result.push([to.center[0], 0.12, to.center[2]])
  }
  return result
}

const routeDistance = (segments, from, to) => {
  let horizontal = 0
  Object.values(segments).forEach((points) => {
    for (let index = 1; index < points.length; index += 1) {
      horizontal += Math.hypot(
        points[index][0] - points[index - 1][0],
        points[index][2] - points[index - 1][2]
      )
    }
  })
  const vertical = Math.abs(to.floorLevel - from.floorLevel) * 3.6
  return Math.round((horizontal + vertical) * 2)
}

export const getC5Route = (from, to, elevatorOnly = false) => {
  if (!from || !to) {
    return { error: '请选择出发地和目的地' }
  }
  if (from.building !== C5_BUILDING_ID || to.building !== C5_BUILDING_ID) {
    return { error: '当前版本仅支持同一教学楼内导航' }
  }
  if (from.id === to.id) {
    return { error: '起点和终点不能相同' }
  }

  const fromGraph = buildGraph(from.floor)
  const toGraph = buildGraph(to.floor)
  const segments = {}
  let via = null

  if (from.floor === to.floor) {
    const path = shortestPath(fromGraph, from.graphNodeId, to.graphNodeId)
    if (!path) {
      return { error: '暂未找到可通行路线' }
    }
    segments[from.floor] = withEndpoints(graphPoints(fromGraph, path), from, to)
  } else {
    const types = elevatorOnly ? ['elevator'] : ['stairs', 'elevator']
    let best = null
    types.forEach((type) => {
      fromGraph.coreCandidates
        .filter((candidate) => candidate.type === type)
        .forEach((fromCore) => {
          const toCore = toGraph.coreCandidates.find(
            (candidate) =>
              candidate.type === type && candidate.nodeId === fromCore.nodeId
          )
          if (!toCore) {
            return
          }
          const fromPath = shortestPath(
            fromGraph,
            from.graphNodeId,
            fromCore.id
          )
          const toPath = shortestPath(
            toGraph,
            toCore.id,
            to.graphNodeId
          )
          if (!fromPath || !toPath) {
            return
          }
          const score =
            fromPath.length + toPath.length +
            Math.abs(fromCore.x - toCore.x) +
            Math.abs(fromCore.z - toCore.z)
          if (!best || score < best.score) {
            best = { type, fromPath, toPath, score }
          }
        })
    })

    if (!best) {
      return { error: '暂未找到可通行路线' }
    }

    via = best.type
    segments[from.floor] = fromCoreSegment(
      graphPoints(fromGraph, best.fromPath),
      from
    )
    segments[to.floor] = toCoreSegment(
      graphPoints(toGraph, best.toPath),
      to
    )
  }

  return {
    id: `${from.id}-${to.id}-${via || 'walk'}`,
    from,
    to,
    fromName: from.name,
    toName: to.name,
    fromFloor: from.floor,
    toFloor: to.floor,
    fromBuilding: from.building,
    toBuilding: to.building,
    via,
    viaLabel: via === 'elevator' ? '电梯' : via === 'stairs' ? '楼梯' : '步行',
    distance: routeDistance(segments, from, to),
    floorLabel:
      from.floor === to.floor ? `${from.floor} 同层` : `${from.floor} → ${to.floor}`,
    segments
  }
}
