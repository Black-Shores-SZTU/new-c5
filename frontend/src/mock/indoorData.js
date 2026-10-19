// ============================================================
// P2 本地 mock 数据：楼层 / 教学楼 / 教室 / 设施 / 路线
// 后端未联调，全部为本地数据；真实接口位置见 src/api/indoor.js
// ============================================================

import {
  C5_BUILDING_FULL_NAME,
  C5_BUILDING_ID,
  C5_BUILDING_NAME,
  getC5Locations,
  getC5Rooms,
  getC5Route
} from './c5Data'

export const FLOOR_HEIGHT = 3.6
const DISTANCE_SCALE = 2

export const floors = [
  { id: 'F1', label: 'F1', name: '一层', level: 1 },
  { id: 'F2', label: 'F2', name: '二层', level: 2 },
  { id: 'F3', label: 'F3', name: '三层', level: 3 },
  { id: 'F4', label: 'F4', name: '四层', level: 4 },
  { id: 'F5', label: 'F5', name: '五层', level: 5 }
]

export const buildings = [
  {
    id: C5_BUILDING_ID,
    name: C5_BUILDING_NAME,
    fullName: C5_BUILDING_FULL_NAME,
    position: [-31, 0, 4],
    size: [18, 9, 12]
  },
  {
    id: 'B',
    name: 'B栋',
    fullName: '教学楼 B',
    position: [0, 0, -10],
    size: [16, 8, 11]
  },
  {
    id: 'C',
    name: 'C栋',
    fullName: '教学楼 C',
    position: [31, 0, 6],
    size: [15, 10, 10]
  }
]

// 单层布局：中间走廊，南北各 4 个房间，两端为楼梯/电梯核心
const ROOM_LAYOUT = [
  { index: 1, x: -21, z: 11, type: 'classroom' },
  { index: 2, x: -7, z: 11, type: 'classroom' },
  { index: 3, x: 7, z: 11, type: 'classroom' },
  { index: 4, x: 21, z: 11, type: 'classroom' },
  { index: 5, x: -21, z: -11, type: 'classroom' },
  { index: 6, x: -7, z: -11, type: 'classroom' },
  { index: 7, x: 7, z: -11, type: 'restroom', name: '卫生间' },
  { index: 8, x: 21, z: -11, type: 'water', name: '开水间' }
]

const roomCache = new Map()
const locationCache = new Map()

const getFloorMeta = (floorId) => floors.find((floor) => floor.id === floorId)

export function getFloorRooms(buildingId, floorId) {
  if (buildingId === C5_BUILDING_ID) {
    return getC5Rooms(floorId)
  }

  const cacheKey = `${buildingId}-${floorId}`
  if (roomCache.has(cacheKey)) {
    return roomCache.get(cacheKey)
  }

  const floor = getFloorMeta(floorId)
  if (!floor) {
    return []
  }

  const rooms = ROOM_LAYOUT.map((layout) => {
    const roomId = `${buildingId}${floor.level}${String(layout.index).padStart(2, '0')}`
    return {
      id: roomId,
      name: layout.name || roomId,
      type: layout.type,
      building: buildingId,
      floor: floorId,
      floorLevel: floor.level,
      center: [layout.x, 0, layout.z],
      door: [layout.x, 0, layout.z > 0 ? 5.6 : -5.6],
      size: [12, 10]
    }
  })

  roomCache.set(cacheKey, rooms)
  return rooms
}

const CORE_LAYOUT = [
  { type: 'elevator', name: '电梯', center: [27, 0, 0] },
  { type: 'stairs', name: '楼梯', center: [-27, 0, 0] },
  { type: 'exit', name: '安全出口', center: [31.2, 0, 0] }
]

export function getFloorLocations(buildingId, floorId) {
  if (buildingId === C5_BUILDING_ID) {
    return getC5Locations(floorId)
  }

  const cacheKey = `${buildingId}-${floorId}`
  if (locationCache.has(cacheKey)) {
    return locationCache.get(cacheKey)
  }

  const floor = getFloorMeta(floorId)
  if (!floor) {
    return []
  }

  const cores = CORE_LAYOUT.map((core) => ({
    id: `${buildingId}-${floorId}-${core.type}`,
    name: core.name,
    type: core.type,
    building: buildingId,
    floor: floorId,
    floorLevel: floor.level,
    center: [...core.center],
    door: [...core.center],
    size: [4, 8]
  }))

  const locations = [...getFloorRooms(buildingId, floorId), ...cores]
  locationCache.set(cacheKey, locations)
  return locations
}

export const allLocations = buildings.flatMap((building) =>
  floors.flatMap((floor) => getFloorLocations(building.id, floor.id))
)

// 支持教室号（A101）与设施名（卫生间/电梯）两种输入
export function findLocation(input, buildingId, floorId) {
  const keyword = String(input || '').trim().toUpperCase()
  if (!keyword) {
    return null
  }

  const exact = allLocations.find((location) => location.id.toUpperCase() === keyword)
  if (exact) {
    return exact
  }

  // C5 房号为纯数字；兼容旧输入 A101 / A305
  const legacyKeyword = keyword.replace(/^A(?=\d)/, '')
  const legacy = allLocations.find(
    (location) =>
      location.building === C5_BUILDING_ID &&
      location.id.toUpperCase() === legacyKeyword
  )
  if (legacy) {
    return legacy
  }

  const named = allLocations.filter(
    (location) => location.name.toUpperCase() === keyword
  )
  if (!named.length) {
    return null
  }

  return (
    named.find(
      (location) => location.building === buildingId && location.floor === floorId
    ) ||
    named.find((location) => location.building === buildingId) ||
    named[0]
  )
}

export function getNearestLocation(type, buildingId, floorId) {
  return getFloorLocations(buildingId, floorId).find(
    (location) => location.type === type
  )
}

function getCore(buildingId, floorId, type) {
  return getFloorLocations(buildingId, floorId).find(
    (location) => location.type === type
  )
}

const distance2D = (a, b) => Math.hypot(a[0] - b[0], a[2] - b[2])

const polylineDistance = (points) => {
  let total = 0
  for (let i = 1; i < points.length; i += 1) {
    total += distance2D(points[i - 1], points[i])
  }
  return total
}

const normalizePoints = (points) => {
  const result = []
  points.forEach((point) => {
    const next = [Number(point[0].toFixed(2)), 0.12, Number(point[2].toFixed(2))]
    const previous = result[result.length - 1]
    if (
      !previous ||
      Math.abs(previous[0] - next[0]) > 0.01 ||
      Math.abs(previous[2] - next[2]) > 0.01
    ) {
      result.push(next)
    }
  })
  return result
}

const chooseVerticalCore = (from, to, elevatorOnly) => {
  if (elevatorOnly) {
    return 'elevator'
  }

  const candidates = ['stairs', 'elevator']
  let bestType = 'stairs'
  let bestDistance = Number.POSITIVE_INFINITY

  candidates.forEach((type) => {
    const fromCore = getCore(from.building, from.floor, type)
    const toCore = getCore(to.building, to.floor, type)
    if (!fromCore || !toCore) {
      return
    }
    const total =
      distance2D(from.center, fromCore.center) +
      distance2D(to.center, toCore.center)
    if (total < bestDistance) {
      bestDistance = total
      bestType = type
    }
  })

  return bestType
}

export function buildRoute(from, to, elevatorOnly = false) {
  if (!from || !to) {
    return { error: '请选择出发地和目的地' }
  }
  if (from.building !== to.building) {
    return { error: '当前版本仅支持同一教学楼内导航' }
  }
  if (from.id === to.id && from.floor === to.floor) {
    return { error: '起点和终点不能相同' }
  }

  if (from.building === C5_BUILDING_ID) {
    return getC5Route(from, to, elevatorOnly)
  }

  const segments = {}
  let verticalType = null

  if (from.floor === to.floor) {
    segments[from.floor] = normalizePoints([
      from.center,
      from.door,
      [from.door[0], 0, 0],
      [to.door[0], 0, 0],
      to.door,
      to.center
    ])
  } else {
    verticalType = chooseVerticalCore(from, to, elevatorOnly)
    const fromCore = getCore(from.building, from.floor, verticalType)
    const toCore = getCore(to.building, to.floor, verticalType)

    segments[from.floor] = normalizePoints([
      from.center,
      from.door,
      [from.door[0], 0, 0],
      [fromCore.center[0], 0, 0],
      fromCore.center
    ])
    segments[to.floor] = normalizePoints([
      toCore.center,
      [toCore.center[0], 0, 0],
      [to.door[0], 0, 0],
      to.door,
      to.center
    ])
  }

  const horizontalDistance = Object.values(segments).reduce(
    (total, points) => total + polylineDistance(points),
    0
  )
  const verticalDistance =
    Math.abs(to.floorLevel - from.floorLevel) * FLOOR_HEIGHT

  return {
    id: `${from.id}-${to.id}-${verticalType || 'walk'}`,
    from,
    to,
    fromName: from.name,
    toName: to.name,
    fromFloor: from.floor,
    toFloor: to.floor,
    fromBuilding: from.building,
    toBuilding: to.building,
    via: verticalType,
    viaLabel:
      verticalType === 'elevator'
        ? '电梯'
        : verticalType === 'stairs'
          ? '楼梯'
          : '步行',
    distance: Math.round((horizontalDistance + verticalDistance) * DISTANCE_SCALE),
    floorLabel:
      from.floor === to.floor ? `${from.floor} 同层` : `${from.floor} → ${to.floor}`,
    segments
  }
}
