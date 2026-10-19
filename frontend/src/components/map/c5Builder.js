import {
  BoxGeometry,
  CanvasTexture,
  Color,
  EdgesGeometry,
  Group,
  LineBasicMaterial,
  LineSegments,
  Mesh,
  MeshBasicMaterial,
  MeshStandardMaterial,
  PlaneGeometry,
  Shape,
  ShapeGeometry,
  Sprite,
  SpriteMaterial,
  SRGBColorSpace,
  Path,
  DoubleSide
} from 'three'
import { FontLoader } from 'three/examples/jsm/loaders/FontLoader.js'
import { TextGeometry } from 'three/examples/jsm/geometries/TextGeometry.js'
import fontJson from '../../assets/fonts/helvetiker_regular.typeface.json'
import {
  C5_BOUNDS,
  C5_COLORS,
  C5_CORRIDORS,
  C5_CORE_NODES,
  getC5FloorData
} from '../../mock/c5Data'

const FONT = new FontLoader().parse(fontJson)
const WALL_HEIGHT = 1.0
const WALL_THICKNESS = 0.18

const roundedRect = (context, x, y, width, height, radius) => {
  context.beginPath()
  context.moveTo(x + radius, y)
  context.lineTo(x + width - radius, y)
  context.quadraticCurveTo(x + width, y, x + width, y + radius)
  context.lineTo(x + width, y + height - radius)
  context.quadraticCurveTo(x + width, y + height, x + width - radius, y + height)
  context.lineTo(x + radius, y + height)
  context.quadraticCurveTo(x, y + height, x, y + height - radius)
  context.lineTo(x, y + radius)
  context.quadraticCurveTo(x, y, x + radius, y)
  context.closePath()
}

const makeLabelSprite = (text, { color = '#334155', scale = [2.8, 1.4] } = {}) => {
  const canvas = document.createElement('canvas')
  canvas.width = 256
  canvas.height = 128
  const context = canvas.getContext('2d')
  if (!context) {
    return new Sprite()
  }
  context.clearRect(0, 0, canvas.width, canvas.height)
  roundedRect(context, 4, 4, 248, 120, 22)
  context.fillStyle = 'rgba(255,255,255,.92)'
  context.fill()
  context.strokeStyle = C5_COLORS.edge
  context.lineWidth = 2
  context.stroke()
  context.fillStyle = color
  context.font =
    '600 40px "HarmonyOS Sans SC", "PingFang SC", "Microsoft YaHei", sans-serif'
  context.textAlign = 'center'
  context.textBaseline = 'middle'
  context.fillText(text, 128, 67)
  const texture = new CanvasTexture(canvas)
  texture.colorSpace = SRGBColorSpace
  texture.anisotropy = 4
  const material = new SpriteMaterial({
    map: texture,
    transparent: true,
    depthWrite: false
  })
  const sprite = new Sprite(material)
  sprite.scale.set(scale[0], scale[1], 1)
  return sprite
}

const addEdges = (parent, geometry, color = C5_COLORS.edge, opacity = 0.35) => {
  const edges = new LineSegments(
    new EdgesGeometry(geometry),
    new LineBasicMaterial({ color, transparent: true, opacity })
  )
  parent.add(edges)
  return edges
}

const makeTextLabel = (
  text,
  {
    widthLimit,
    depthLimit,
    color = C5_COLORS.text,
    y = 0.09,
    baseSize = 1.2
  }
) => {
  if (!text) {
    return null
  }
  const group = new Group()
  const geometry = new TextGeometry(text, {
    font: FONT,
    size: baseSize,
    depth: 0.03,
    curveSegments: 2,
    bevelEnabled: false
  })
  geometry.computeBoundingBox()
  geometry.center()
  const box = geometry.boundingBox
  const width = Math.max(0.1, box.max.x - box.min.x)
  const height = Math.max(0.1, box.max.y - box.min.y)
  const scaleX = Math.min(1, widthLimit / width)
  const scaleY = Math.min(1, depthLimit / height)

  const textMesh = new Mesh(
    geometry,
    new MeshBasicMaterial({
      color,
      polygonOffset: true,
      polygonOffsetFactor: -1,
      polygonOffsetUnits: -1
    })
  )
  textMesh.scale.set(scaleX, scaleY, 1)
  textMesh.rotation.x = -Math.PI / 2
  textMesh.position.y = y
  textMesh.renderOrder = 3
  group.add(textMesh)

  const plateWidth = width * scaleX + 0.24
  const plateHeight = height * scaleY + 0.18
  const plate = new Mesh(
    new PlaneGeometry(plateWidth, plateHeight),
    new MeshBasicMaterial({
      color: 0xffffff,
      transparent: true,
      opacity: 0.96,
      depthWrite: false,
      side: DoubleSide
    })
  )
  plate.rotation.x = -Math.PI / 2
  plate.position.y = y - 0.012
  plate.renderOrder = 2
  group.add(plate)

  return group
}

const roomFloorColor = (type) => {
  if (type === 'teacher_rest') {
    return C5_COLORS.teacherFloor
  }
  if (type === 'auditorium') {
    return C5_COLORS.auditoriumFloor
  }
  return C5_COLORS.classroomFloor
}

const roomTextColor = (type) => {
  if (type === 'teacher_rest') {
    return C5_COLORS.teacherAccent
  }
  if (type === 'auditorium') {
    return C5_COLORS.auditoriumAccent
  }
  return C5_COLORS.text
}

const addWalls = (group, room, wallMaterial, edgeMaterial) => {
  const [width, depth] = room.size
  const x1 = room.center[0] - width / 2
  const x2 = room.center[0] + width / 2
  const z1 = room.center[2] - depth / 2
  const z2 = room.center[2] + depth / 2
  const t = WALL_THICKNESS

  const addWall = (w, d, x, z) => {
    const geometry = new BoxGeometry(w, WALL_HEIGHT, d)
    const wall = new Mesh(geometry, wallMaterial)
    wall.position.set(x, WALL_HEIGHT / 2, z)
    wall.castShadow = true
    wall.receiveShadow = true
    group.add(wall)
    const edges = new LineSegments(new EdgesGeometry(geometry), edgeMaterial)
    edges.position.copy(wall.position)
    group.add(edges)
  }

  if (room.openSide === 'south') {
    addWall(width, t, room.center[0], z1 + t / 2)
    addWall(t, depth, x1 + t / 2, room.center[2])
    addWall(t, depth, x2 - t / 2, room.center[2])
  } else if (room.openSide === 'north') {
    addWall(width, t, room.center[0], z2 - t / 2)
    addWall(t, depth, x1 + t / 2, room.center[2])
    addWall(t, depth, x2 - t / 2, room.center[2])
  } else if (room.openSide === 'east') {
    addWall(t, depth, x1 + t / 2, room.center[2])
    addWall(width, t, room.center[0], z1 + t / 2)
    addWall(width, t, room.center[0], z2 - t / 2)
  } else {
    addWall(t, depth, x2 - t / 2, room.center[2])
    addWall(width, t, room.center[0], z1 + t / 2)
    addWall(width, t, room.center[0], z2 - t / 2)
  }
}

const addRoom = (parent, room, wallMaterial, edgeMaterial) => {
  const [width, depth] = room.size
  const floorGeometry = new PlaneGeometry(width, depth)
  const floor = new Mesh(
    floorGeometry,
    new MeshStandardMaterial({
      color: roomFloorColor(room.type),
      roughness: 0.94,
      metalness: 0
    })
  )
  floor.rotation.x = -Math.PI / 2
  floor.position.set(room.center[0], 0.055, room.center[2])
  floor.receiveShadow = true
  parent.add(floor)

  const floorEdges = new LineSegments(
    new EdgesGeometry(floorGeometry),
    edgeMaterial
  )
  floorEdges.rotation.x = -Math.PI / 2
  floorEdges.position.set(room.center[0], 0.057, room.center[2])
  parent.add(floorEdges)

  addWalls(parent, room, wallMaterial, edgeMaterial)

  const label = makeTextLabel(room.roomNo, {
    widthLimit: Math.max(0.8, width * 0.68),
    depthLimit: Math.max(0.55, depth * 0.42),
    color: roomTextColor(room.type)
  })
  if (label) {
    label.position.set(room.center[0], 0, room.center[2])
    parent.add(label)
  }
}

const addCorridors = (parent) => {
  Object.values(C5_CORRIDORS).forEach((corridor, index) => {
    const width = corridor.x2 - corridor.x1
    const depth = corridor.z2 - corridor.z1
    const geometry = new PlaneGeometry(width, depth)
    const mesh = new Mesh(
      geometry,
      new MeshStandardMaterial({
        color: C5_COLORS.corridor,
        roughness: 0.96,
        metalness: 0
      })
    )
    mesh.rotation.x = -Math.PI / 2
    mesh.position.set(
      (corridor.x1 + corridor.x2) / 2,
      0.012 + index * 0.0015,
      (corridor.z1 + corridor.z2) / 2
    )
    mesh.receiveShadow = true
    parent.add(mesh)

    const label = makeLabelSprite(corridor.name, {
      color: '#64748B',
      scale: corridor.axis === 'x' ? [4.2, 2.0] : [3.8, 1.9]
    })
    label.position.set(
      corridor.axis === 'x' ? 0 : (corridor.x1 + corridor.x2) / 2,
      1.9,
      corridor.axis === 'x' ? (corridor.z1 + corridor.z2) / 2 : 1
    )
    parent.add(label)
  })
}

const addSlab = (parent, floorId, floorLevel) => {
  const minX = C5_BOUNDS.minX
  const maxX = C5_BOUNDS.maxX
  const minZ = C5_BOUNDS.minZ
  const maxZ = C5_BOUNDS.maxZ
  const shape = new Shape()
  shape.moveTo(minX, -minZ)
  shape.lineTo(maxX, -minZ)
  shape.lineTo(maxX, -maxZ)
  shape.lineTo(minX, -maxZ)
  shape.lineTo(minX, -minZ)

  if (floorLevel > 1) {
    const hole = new Path()
    hole.moveTo(2, -8.5)
    hole.lineTo(32, -8.5)
    hole.lineTo(32, -20.5)
    hole.lineTo(2, -20.5)
    hole.lineTo(2, -8.5)
    shape.holes.push(hole)
  }

  const geometry = new ShapeGeometry(shape)
  const slab = new Mesh(
    geometry,
    new MeshStandardMaterial({
      color: C5_COLORS.slab,
      roughness: 0.96,
      metalness: 0
    })
  )
  slab.rotation.x = -Math.PI / 2
  slab.position.y = 0
  slab.receiveShadow = true
  parent.add(slab)

  const edges = new LineSegments(
    new EdgesGeometry(geometry),
    new LineBasicMaterial({
      color: C5_COLORS.edge,
      transparent: true,
      opacity: 0.7
    })
  )
  edges.rotation.x = -Math.PI / 2
  edges.position.y = 0.002
  parent.add(edges)

  if (floorLevel > 1) {
    // 中庭开洞处补一层浅色底面，避免俯视时露出场景背景形成纯黑矩形
    const voidFloor = new Mesh(
      new PlaneGeometry(30, 12),
      new MeshStandardMaterial({
        color: C5_COLORS.auditoriumFloor,
        roughness: 0.96,
        metalness: 0
      })
    )
    voidFloor.rotation.x = -Math.PI / 2
    voidFloor.position.set(17, -0.32, 14.5)
    voidFloor.receiveShadow = true
    parent.add(voidFloor)

    const voidEdges = new LineSegments(
      new EdgesGeometry(new PlaneGeometry(30, 12)),
      new LineBasicMaterial({
        color: C5_COLORS.edge,
        transparent: true,
        opacity: 0.6
      })
    )
    voidEdges.rotation.x = -Math.PI / 2
    voidEdges.position.set(17, -0.3, 14.5)
    parent.add(voidEdges)

    const courtyardLabel = makeLabelSprite('中庭', {
      color: '#94A1B2',
      scale: [3.4, 1.7]
    })
    courtyardLabel.position.set(17, 1.5, 14.5)
    parent.add(courtyardLabel)
  }
}

const addCoreBlocks = (parent, floorId) => {
  const data = getC5FloorData(floorId)
  data.facilities
    .filter((facility) => ['elevator', 'stairs', 'restroom'].includes(facility.type))
    .forEach((facility) => {
      const isElevator = facility.type === 'elevator'
      const isRestroom = facility.type === 'restroom'
      const fill = isElevator
        ? C5_COLORS.elevatorFloor
        : isRestroom
          ? C5_COLORS.restroomFloor
          : C5_COLORS.stairsFloor
      const accent = isElevator
        ? C5_COLORS.elevatorAccent
        : isRestroom
          ? C5_COLORS.restroomAccent
          : C5_COLORS.stairsAccent
      const size = isElevator || isRestroom ? [2.8, 2.0] : [2.6, 1.8]
      const group = new Group()
      group.position.set(facility.center[0], 0, facility.center[2])
      const geometry = new BoxGeometry(size[0], 1.0, size[1])
      const block = new Mesh(
        geometry,
        new MeshStandardMaterial({
          color: fill,
          roughness: 0.9,
          metalness: 0
        })
      )
      block.position.y = 0.5
      block.castShadow = true
      block.receiveShadow = true
      group.add(block)
      const edges = new LineSegments(
        new EdgesGeometry(geometry),
        new LineBasicMaterial({
          color: accent,
          transparent: true,
          opacity: 0.8
        })
      )
      edges.position.y = 0.5
      group.add(edges)

      if (facility.type === 'stairs') {
        for (let step = 0; step < 3; step += 1) {
          const stepGeometry = new BoxGeometry(size[0] * 0.78, 0.12, 0.28)
          const stepMesh = new Mesh(
            stepGeometry,
            new MeshStandardMaterial({
              color: C5_COLORS.stairsFloor,
              roughness: 0.9,
              metalness: 0
            })
          )
          stepMesh.position.set(0, 1.02 + step * 0.12, -0.4 + step * 0.34)
          stepMesh.castShadow = true
          group.add(stepMesh)
        }
      }

      const label = makeLabelSprite(facility.name, {
        color: accent,
        scale: [2.4, 1.2]
      })
      label.position.set(0, 2.0, 0)
      group.add(label)
      parent.add(group)
    })

  const entrance = data.facilities.find((facility) => facility.type === 'entrance')
  if (entrance) {
    const pad = new Mesh(
      new PlaneGeometry(entrance.size[0], entrance.size[1]),
      new MeshBasicMaterial({
        color: '#E6F7F4',
        transparent: true,
        opacity: 0.9,
        depthWrite: false
      })
    )
    pad.rotation.x = -Math.PI / 2
    pad.position.set(entrance.center[0], 0.04, entrance.center[2])
    parent.add(pad)
    const label = makeLabelSprite('C5 教学楼入口', {
      color: C5_COLORS.restroomAccent,
      scale: [5.2, 2.2]
    })
    label.position.set(entrance.center[0], 2.3, entrance.center[2])
    parent.add(label)
  }

  const auditorium = data.facilities.find(
    (facility) => facility.type === 'auditorium'
  )
  if (auditorium) {
    const group = new Group()
    group.position.set(auditorium.center[0], 0, auditorium.center[2])
    const floor = new Mesh(
      new PlaneGeometry(auditorium.size[0], auditorium.size[1]),
      new MeshStandardMaterial({
        color: C5_COLORS.auditoriumFloor,
        roughness: 0.94,
        metalness: 0
      })
    )
    floor.rotation.x = -Math.PI / 2
    floor.position.y = 0.055
    floor.receiveShadow = true
    group.add(floor)
    const wallMaterial = new MeshStandardMaterial({
      color: C5_COLORS.wall,
      roughness: 0.9,
      metalness: 0
    })
    const edgeMaterial = new LineBasicMaterial({
      color: C5_COLORS.edge,
      transparent: true,
      opacity: 0.4
    })
    // group 已平移到演播厅中心，墙体必须使用局部坐标，避免二次平移导致整层错位
    const pseudoRoom = {
      center: [0, 0, 0],
      size: auditorium.size,
      openSide: 'north'
    }
    addWalls(group, pseudoRoom, wallMaterial, edgeMaterial)
    const label = makeLabelSprite('演播厅', {
      color: C5_COLORS.auditoriumAccent,
      scale: [5.4, 2.4]
    })
    // 标签向房间北侧偏移，避开居中的「133-138」地面房号
    label.position.set(0, 2.2, -4.2)
    group.add(label)
    const number = makeTextLabel('133-138', {
      widthLimit: 18,
      depthLimit: 6,
      color: C5_COLORS.auditoriumAccent,
      baseSize: 1.6
    })
    if (number) {
      number.position.set(0, 0, 0)
      group.add(number)
    }
    parent.add(group)
  }
}

const addFloorLabel = (parent, floorLevel) => {
  const label = makeTextLabel(`${floorLevel}F`, {
    widthLimit: 7,
    depthLimit: 5,
    color: '#64748B',
    baseSize: 2.4
  })
  if (label) {
    label.position.set(-38, 0.06, 31)
    parent.add(label)
  }
}

export const buildC5FloorGroup = (floorId) => {
  const floorLevel = Number(floorId.replace('F', '')) || 1
  const data = getC5FloorData(floorId)
  const group = new Group()
  group.userData.floorId = floorId

  const wallMaterial = new MeshStandardMaterial({
    color: C5_COLORS.wall,
    roughness: 0.9,
    metalness: 0
  })
  const edgeMaterial = new LineBasicMaterial({
    color: C5_COLORS.edge,
    transparent: true,
    opacity: 0.35
  })

  addSlab(group, floorId, floorLevel)
  addCorridors(group)
  data.rooms.forEach((room) => {
    addRoom(group, room, wallMaterial, edgeMaterial)
  })
  addCoreBlocks(group, floorId)
  addFloorLabel(group, floorLevel)

  return group
}
