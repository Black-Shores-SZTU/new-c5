<script setup>
import {
  BufferGeometry,
  CanvasTexture,
  CircleGeometry,
  Color,
  ConeGeometry,
  DirectionalLight,
  DoubleSide,
  EdgesGeometry,
  Float32BufferAttribute,
  GridHelper,
  Group,
  HemisphereLight,
  LineBasicMaterial,
  LineSegments,
  Mesh,
  MeshBasicMaterial,
  MeshStandardMaterial,
  NoToneMapping,
  OrthographicCamera,
  PCFShadowMap,
  PerspectiveCamera,
  PlaneGeometry,
  Raycaster,
  RingGeometry,
  SRGBColorSpace,
  Scene,
  SphereGeometry,
  Sprite,
  SpriteMaterial,
  Vector2,
  Vector3,
  WebGLRenderer,
  BoxGeometry
} from 'three'
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js'
import { Line2 } from 'three/examples/jsm/lines/Line2.js'
import { LineGeometry } from 'three/examples/jsm/lines/LineGeometry.js'
import { LineMaterial } from 'three/examples/jsm/lines/LineMaterial.js'
import gsap from 'gsap'
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { buildC5FloorGroup } from './c5Builder'
import { C5_BUILDING_ID } from '../../mock/c5Data'
import {
  buildings,
  floors,
  getFloorLocations,
  getFloorRooms
} from '../../mock/indoorData'

const props = defineProps({
  viewMode: {
    type: String,
    default: '3d'
  },
  currentFloor: {
    type: String,
    default: 'F1'
  },
  building: {
    type: String,
    default: 'A'
  },
  route: {
    type: Object,
    default: null
  },
  overview: {
    type: Boolean,
    default: false
  },
  navigation: {
    type: Object,
    default: () => ({
      active: false,
      paused: false,
      following: true,
      position: [0, 0, 0],
      direction: [0, 0, 1]
    })
  },
  selectedLocation: {
    type: Object,
    default: null
  }
})

const emit = defineEmits(['building-select', 'camera-interaction'])

const containerRef = ref(null)

const COLORS = {
  bg3d: '#101B2C',
  bg2d: '#E4EAF2',
  floor: '#EDF2F7',
  room: '#F7FAFD',
  roomRestroom: '#FFF6E8',
  roomWater: '#E6F7F4',
  wall: '#F7FAFD',
  edge: '#C7D4E4',
  corridor: '#EDF2F7',
  label: '#475569',
  route: '#0A84E8',
  casing: '#FFFFFF',
  start: '#12B76A',
  end: '#E5484D',
  primary: '#2B6BE5',
  primary50: '#EDF3FE',
  primary100: '#D9E6FD',
  primary600: '#2258C8',
  accent: '#0E9F8F',
  accent50: '#E6F7F4',
  accent600: '#0A8175',
  warning: '#E8A33D',
  warning50: '#FFF6E8',
  warning600: '#9A6700',
  neutral200: '#E2E8F0',
  neutral300: '#C9D3E0',
  neutral500: '#64748B',
  neutral700: '#334155',
  neutral900: '#0F172A'
}

let scene = null
let renderer = null
let perspectiveCamera = null
let orthographicCamera = null
let activeCamera = null
let controls = null
let animationFrameId = 0
let resizeObserver = null
let indoorRoot = null
let campusRoot = null
let routeRoot = null
let routeAnim = null
let cameraTween = null
let targetTween = null
let zoomTween = null
let isTransitioning = false
let isCameraAnimating = false
let navActive = false
let navArrow = null
let navPointerStart = null
let navInteractionEmitted = false
let selectionRoot = null
let is2dPanning = false
let last2dPointer = null
let currentMode = props.viewMode
let currentOverview = props.overview
let currentFloorId = props.currentFloor
let currentBuildingId = props.building
let viewportWidth = 1
let viewportHeight = 1

const floorGroups = new Map()
const campusBuildings = []
const raycaster = new Raycaster()
const pointer = new Vector2()
let pointerDown = null
let hoveredBuilding = null

const FOLLOW_OFFSET = new Vector3(18, 26, 22)
const FOLLOW_LOOKAHEAD = 7
const FOLLOW_LERP = 0.1

const getFollowPreset = () => {
  const position = props.navigation?.position || [0, 0, 0]
  const directionValue = props.navigation?.direction || [0, 0, 1]
  const avatar = new Vector3(position[0], 0, position[2])
  const direction = new Vector3(
    directionValue[0],
    0,
    directionValue[2]
  )

  if (direction.lengthSq() < 0.0001) {
    direction.set(0, 0, 1)
  }
  direction.normalize()

  return {
    position: avatar.clone().add(FOLLOW_OFFSET),
    target: avatar
      .clone()
      .add(direction.multiplyScalar(FOLLOW_LOOKAHEAD)),
    orthoSize: 52
  }
}

const updateFollowCamera = () => {
  if (!controls || !activeCamera?.isPerspectiveCamera) {
    return
  }
  const position = props.navigation?.position
  if (!position) {
    return
  }

  const directionValue = props.navigation?.direction || [0, 0, 1]
  const avatar = new Vector3(position[0], 0, position[2])
  const direction = new Vector3(
    directionValue[0],
    0,
    directionValue[2]
  )
  if (direction.lengthSq() < 0.0001) {
    direction.set(0, 0, 1)
  }
  direction.normalize()

  const desiredPosition = avatar.clone().add(FOLLOW_OFFSET)
  const desiredTarget = avatar
    .clone()
    .add(direction.multiplyScalar(FOLLOW_LOOKAHEAD))

  activeCamera.position.lerp(desiredPosition, FOLLOW_LERP)
  controls.target.lerp(desiredTarget, FOLLOW_LERP)
  activeCamera.lookAt(controls.target)
}

const updateNavArrow = () => {
  if (!navArrow || !props.navigation?.position) {
    return
  }
  const position = props.navigation.position
  const direction = props.navigation.direction || [0, 0, 1]
  navArrow.group.position.set(position[0], 0, position[2])
  navArrow.arrow.rotation.y = Math.atan2(direction[0], direction[2])
}

const updateSelectionMarker = () => {
  if (selectionRoot) {
    scene.remove(selectionRoot)
    disposeObject(selectionRoot)
    selectionRoot = null
  }

  const location = props.selectedLocation
  if (
    !location ||
    location.floor !== currentFloorId ||
    currentOverview ||
    navActive
  ) {
    return
  }

  selectionRoot = new Group()
  const group = new Group()
  group.position.set(location.center[0], 0, location.center[2])

  const baseRing = new Mesh(
    new RingGeometry(0.95, 1.1, 48),
    new MeshBasicMaterial({
      color: COLORS.casing,
      transparent: true,
      opacity: 1,
      depthTest: false
    })
  )
  baseRing.rotation.x = -Math.PI / 2
  baseRing.position.y = 0.03
  baseRing.renderOrder = 10
  group.add(baseRing)

  const ring = new Mesh(
    new RingGeometry(0.74, 0.85, 48),
    new MeshBasicMaterial({
      color: COLORS.primary,
      transparent: true,
      opacity: 1,
      depthTest: false
    })
  )
  ring.rotation.x = -Math.PI / 2
  ring.position.y = 0.04
  ring.renderOrder = 10
  group.add(ring)

  const head = new Mesh(
    new SphereGeometry(0.62, 16, 16),
    new MeshStandardMaterial({
      color: COLORS.primary,
      roughness: 0.35,
      metalness: 0.05
    })
  )
  head.position.y = 1.75
  head.castShadow = true
  group.add(head)

  const tip = new Mesh(
    new ConeGeometry(0.5, 1.2, 16),
    new MeshStandardMaterial({
      color: COLORS.primary,
      roughness: 0.35,
      metalness: 0.05
    })
  )
  tip.position.y = 0.95
  tip.rotation.x = Math.PI
  tip.castShadow = true
  group.add(tip)

  const dot = new Mesh(
    new SphereGeometry(0.18, 12, 12),
    new MeshBasicMaterial({ color: COLORS.casing })
  )
  dot.position.set(0, 1.8, 0.55)
  group.add(dot)

  // 房间/设施自身已有名称标签，选中标记不再重复名称，避免双层文字重叠

  selectionRoot.add(group)
  selectionRoot.visible = !currentOverview
  scene.add(selectionRoot)
}

const focusLocation = (location) => {
  if (!controls || !location) {
    return
  }
  if (isTransitioning) {
    window.setTimeout(() => focusLocation(location), 750)
    return
  }

  isCameraAnimating = false
  gsap.killTweensOf(activeCamera.position)
  gsap.killTweensOf(controls.target)

  const target = new Vector3(location.center[0], 0, location.center[2])
  const is2d = activeCamera.isOrthographicCamera
  const nextPosition = is2d
    ? new Vector3(target.x, activeCamera.position.y, target.z)
    : target.clone().add(new Vector3(18, 24, 22))

  isCameraAnimating = true
  controls.enabled = false

  const finish = () => {
    isCameraAnimating = false
    applyControlsEnabled()
    if (currentMode === '3d') {
      controls.update()
    }
  }

  gsap.to(controls.target, {
    x: target.x,
    y: target.y,
    z: target.z,
    duration: 0.6,
    ease: 'power2.inOut'
  })

  gsap.to(activeCamera.position, {
    x: nextPosition.x,
    y: nextPosition.y,
    z: nextPosition.z,
    duration: 0.6,
    ease: 'power2.inOut',
    onUpdate: () => activeCamera.lookAt(controls.target),
    onComplete: finish
  })

  if (is2d) {
    gsap.to(activeCamera, {
      zoom: Math.max(activeCamera.zoom, 1.1),
      duration: 0.6,
      ease: 'power2.inOut',
      onUpdate: () => activeCamera.updateProjectionMatrix()
    })
  }
}

const getPreset = (mode, overview) => {
  if (overview) {
    return mode === '2d'
      ? {
          position: new Vector3(0, 120, 0),
          target: new Vector3(0, 0, 0),
          orthoSize: 96
        }
      : {
          position: new Vector3(62, 56, 76),
          target: new Vector3(0, 0, 0),
          orthoSize: 96
        }
  }

  // 桌面端左侧有搜索面板，2D 取景略向左让出遮挡；窄屏无侧栏，改为居中避免东侧被裁切
  const indoorCenterX = viewportWidth < 720 ? 0 : -12

  return mode === '2d'
    ? {
        position: new Vector3(indoorCenterX, 118, 0),
        target: new Vector3(indoorCenterX, 0, 0),
        orthoSize: 84
      }
    : {
        position: new Vector3(56, 72, 78),
        target: new Vector3(-6, 0, 0),
        orthoSize: 84
      }
}

// 2D 由自定义平移/缩放接管，OrbitControls 完全关闭，避免其极角约束把相机拉成侧视
const applyControlsEnabled = () => {
  if (!controls) {
    return
  }
  const followOwnsCamera = Boolean(navActive && props.navigation?.following)
  controls.enabled = currentMode === '3d' && !followOwnsCamera
}

const configureControls = (mode) => {
  if (!controls) {
    return
  }
  controls.enableRotate = mode === '3d'
  controls.screenSpacePanning = mode === '2d'
  controls.minPolarAngle = mode === '3d' ? 0.16 : 0
  controls.maxPolarAngle = mode === '3d' ? Math.PI / 2.2 : Math.PI
  applyControlsEnabled()
}

const setCameraUpForMode = (camera, mode) => {
  if (!camera) {
    return
  }
  if (mode === '2d') {
    camera.up.set(0, 0, -1)
  } else {
    camera.up.set(0, 1, 0)
  }
}

const updateOrthoFrustum = (camera, baseSize) => {
  if (!camera || !camera.isOrthographicCamera) {
    return
  }
  const aspect = viewportWidth / Math.max(viewportHeight, 1)
  const neededWidth = currentOverview ? 126 : 94
  const neededDepth = currentOverview ? 86 : 72
  const verticalSize = Math.max(baseSize, neededDepth, neededWidth / Math.max(aspect, 0.1))
  camera.left = (-verticalSize * aspect) / 2
  camera.right = (verticalSize * aspect) / 2
  camera.top = verticalSize / 2
  camera.bottom = -verticalSize / 2
  camera.updateProjectionMatrix()
}

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

const makeLabelSprite = (text, options = {}) => {
  const {
    color = COLORS.label,
    background = 'rgba(255,255,255,.9)',
    border = COLORS.neutral200,
    scale = [5.2, 2.6]
  } = options

  const canvas = document.createElement('canvas')
  canvas.width = 256
  canvas.height = 128
  const context = canvas.getContext('2d')

  context.clearRect(0, 0, canvas.width, canvas.height)
  roundedRect(context, 4, 4, 248, 120, 24)
  context.fillStyle = background
  context.fill()
  context.strokeStyle = border
  context.lineWidth = 2
  context.stroke()

  context.fillStyle = color
  context.font =
    '600 44px "HarmonyOS Sans SC", "PingFang SC", "Microsoft YaHei", sans-serif'
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

const addEdges = (parent, geometry, color = COLORS.edge, opacity = 0.55) => {
  const edges = new LineSegments(
    new EdgesGeometry(geometry),
    new LineBasicMaterial({
      color,
      transparent: true,
      opacity
    })
  )
  parent.add(edges)
  return edges
}

const buildRoom = (room) => {
  const group = new Group()
  const [width, depth] = room.size
  const floorColor =
    room.type === 'restroom'
      ? COLORS.roomRestroom
      : room.type === 'water'
        ? COLORS.roomWater
        : COLORS.room

  const floorGeometry = new BoxGeometry(width, 0.08, depth)
  const floorMesh = new Mesh(
    floorGeometry,
    new MeshStandardMaterial({
      color: floorColor,
      roughness: 0.95,
      metalness: 0
    })
  )
  floorMesh.position.set(room.center[0], 0.04, room.center[2])
  floorMesh.receiveShadow = true
  group.add(floorMesh)

  const outline = addEdges(group, floorGeometry, COLORS.edge, 0.5)
  outline.position.copy(floorMesh.position)

  const wallHeight = 1.1
  const thickness = 0.24
  const wallMaterial = new MeshStandardMaterial({
    color: COLORS.wall,
    roughness: 0.9,
    metalness: 0
  })

  const addWall = (wallWidth, wallDepth, x, z) => {
    const geometry = new BoxGeometry(wallWidth, wallHeight, wallDepth)
    const wall = new Mesh(geometry, wallMaterial)
    wall.position.set(x, wallHeight / 2, z)
    wall.castShadow = true
    wall.receiveShadow = true
    group.add(wall)
    const wallEdges = addEdges(group, geometry, COLORS.edge, 0.4)
    wallEdges.position.copy(wall.position)
  }

  const backZ =
    room.center[2] > 0
      ? room.center[2] + depth / 2 - thickness / 2
      : room.center[2] - depth / 2 + thickness / 2
  addWall(width, thickness, room.center[0], backZ)
  addWall(thickness, depth, room.center[0] - width / 2 + thickness / 2, room.center[2])
  addWall(thickness, depth, room.center[0] + width / 2 - thickness / 2, room.center[2])

  const labelColor =
    room.type === 'restroom'
      ? COLORS.warning600
      : room.type === 'water'
        ? COLORS.accent600
        : COLORS.label
  const label = makeLabelSprite(room.name, {
    color: labelColor,
    scale: [5.2, 2.6]
  })
  // 标签向房间内侧偏移，避免遮住位于房间中心的起终点标记
  const labelOffsetZ = room.center[2] > 0 ? 3.2 : -3.2
  label.position.set(room.center[0], 2.3, room.center[2] + labelOffsetZ)
  group.add(label)

  return group
}

const buildCore = (location) => {
  const group = new Group()
  const isExit = location.type === 'exit'
  const size = isExit ? [2.2, 0.7, 2.2] : [4, 1.1, 8]
  const color =
    location.type === 'elevator'
      ? COLORS.primary50
      : location.type === 'stairs'
        ? COLORS.accent50
        : COLORS.warning50
  const edgeColor =
    location.type === 'elevator'
      ? COLORS.primary
      : location.type === 'stairs'
        ? COLORS.accent
        : COLORS.warning
  const textColor =
    location.type === 'elevator'
      ? COLORS.primary600
      : location.type === 'stairs'
        ? COLORS.accent600
        : COLORS.warning600

  const geometry = new BoxGeometry(size[0], size[1], size[2])
  const mesh = new Mesh(
    geometry,
    new MeshStandardMaterial({
      color,
      roughness: 0.88,
      metalness: 0
    })
  )
  mesh.position.set(location.center[0], size[1] / 2, location.center[2])
  mesh.castShadow = true
  mesh.receiveShadow = true
  group.add(mesh)

  const edges = addEdges(group, geometry, edgeColor, 0.75)
  edges.position.copy(mesh.position)

  const label = makeLabelSprite(location.name, {
    color: textColor,
    scale: [4.4, 2.2]
  })
  const labelOffsetZ = isExit ? 1.6 : 3.4
  label.position.set(
    location.center[0],
    isExit ? 1.8 : 2.3,
    location.center[2] + labelOffsetZ
  )
  group.add(label)

  return group
}

const buildFloorGroup = (buildingId, floor) => {
  if (buildingId === C5_BUILDING_ID) {
    const group = buildC5FloorGroup(floor.id)
    group.userData.floorId = floor.id
    return group
  }

  const group = new Group()
  group.userData.floorId = floor.id

  const slabGeometry = new BoxGeometry(64, 0.4, 36)
  const slab = new Mesh(
    slabGeometry,
    new MeshStandardMaterial({
      color: COLORS.floor,
      roughness: 0.96,
      metalness: 0
    })
  )
  slab.position.y = -0.2
  slab.receiveShadow = true
  group.add(slab)

  const slabEdges = addEdges(group, slabGeometry, COLORS.edge, 0.65)
  slabEdges.position.copy(slab.position)

  const corridorGeometry = new PlaneGeometry(64, 8)
  const corridor = new Mesh(
    corridorGeometry,
    new MeshStandardMaterial({
      color: COLORS.corridor,
      roughness: 0.96,
      metalness: 0
    })
  )
  corridor.rotation.x = -Math.PI / 2
  corridor.position.y = 0.012
  corridor.receiveShadow = true
  group.add(corridor)

  getFloorRooms(buildingId, floor.id).forEach((room) => {
    group.add(buildRoom(room))
  })

  getFloorLocations(buildingId, floor.id)
    .filter((location) => ['elevator', 'stairs', 'exit'].includes(location.type))
    .forEach((location) => {
      group.add(buildCore(location))
    })

  return group
}

const buildIndoorWorld = (buildingId) => {
  if (indoorRoot) {
    scene.remove(indoorRoot)
    disposeObject(indoorRoot)
    indoorRoot = null
  }

  floorGroups.clear()
  indoorRoot = new Group()
  indoorRoot.visible = !currentOverview
  scene.add(indoorRoot)

  floors.forEach((floor) => {
    const group = buildFloorGroup(buildingId, floor)
    group.visible = floor.id === currentFloorId
    indoorRoot.add(group)
    floorGroups.set(floor.id, group)
  })

  updateRouteVisuals()
}

const buildCampus = () => {
  campusRoot = new Group()
  campusRoot.visible = currentOverview
  scene.add(campusRoot)

  const groundGeometry = new PlaneGeometry(160, 100)
  const ground = new Mesh(
    groundGeometry,
    new MeshStandardMaterial({
      color: COLORS.bg2d,
      roughness: 0.98,
      metalness: 0
    })
  )
  ground.rotation.x = -Math.PI / 2
  ground.position.y = -1
  ground.receiveShadow = true
  campusRoot.add(ground)

  const grid = new GridHelper(160, 80, 0xc7d4e4, 0xd3dde9)
  grid.position.y = -0.98
  grid.material.transparent = true
  grid.material.opacity = 0.35
  grid.material.depthWrite = false
  campusRoot.add(grid)

  buildings.forEach((building) => {
    const group = new Group()
    group.position.set(building.position[0], 0, building.position[2])

    const geometry = new BoxGeometry(
      building.size[0],
      building.size[1],
      building.size[2]
    )
    const mesh = new Mesh(
      geometry,
      new MeshStandardMaterial({
        color: COLORS.floor,
        roughness: 0.86,
        metalness: 0
      })
    )
    mesh.position.y = building.size[1] / 2
    mesh.castShadow = true
    mesh.receiveShadow = true
    mesh.userData.buildingId = building.id
    group.add(mesh)

    const edges = addEdges(group, geometry, COLORS.edge, 0.8)
    edges.position.copy(mesh.position)

    const roof = new Mesh(
      new BoxGeometry(building.size[0] * 0.35, 0.5, building.size[2] * 0.35),
      new MeshStandardMaterial({
        color: COLORS.room,
        roughness: 0.9,
        metalness: 0
      })
    )
    roof.position.y = building.size[1] + 0.25
    roof.castShadow = true
    group.add(roof)

    const label = makeLabelSprite(building.name, {
      color: COLORS.neutral700,
      scale: [7, 3.5]
    })
    label.position.set(0, building.size[1] + 2.4, 0)
    group.add(label)

    campusRoot.add(group)
    campusBuildings.push({ id: building.id, mesh })
  })
}

// GridHelper 使用全局令牌中的结构描边色，透明度压低，仅作校园总览的轻网格
const createArrowGeometry = () => {
  const geometry = new BufferGeometry()
  geometry.setAttribute(
    'position',
    new Float32BufferAttribute(
      [0, 0, 1.15, -0.62, 0, -0.78, 0.62, 0, -0.78],
      3
    )
  )
  geometry.computeVertexNormals()
  return geometry
}

const createMarker = (position, color) => {
  const group = new Group()
  group.position.set(position[0], 0.1, position[2])

  const base = new Mesh(
    new RingGeometry(0.72, 0.86, 48),
    new MeshBasicMaterial({
      color,
      transparent: true,
      opacity: 0.9,
      depthTest: false
    })
  )
  base.rotation.x = -Math.PI / 2
  base.position.y = 0.02
  base.renderOrder = 7
  group.add(base)

  const pyramid = new Mesh(
    new ConeGeometry(0.85, 1.8, 4),
    new MeshStandardMaterial({
      color,
      roughness: 0.35,
      metalness: 0.05,
      flatShading: true
    })
  )
  pyramid.position.y = 1.05
  pyramid.rotation.y = Math.PI / 4
  pyramid.castShadow = true
  group.add(pyramid)

  const rings = []
  for (let index = 0; index < 2; index += 1) {
    const ring = new Mesh(
      new RingGeometry(0.92, 1.04, 48),
      new MeshBasicMaterial({
        color,
        transparent: true,
        opacity: 0.4,
        depthTest: false,
        depthWrite: false
      })
    )
    ring.rotation.x = -Math.PI / 2
    ring.position.y = 0.03
    ring.userData.phase = index * 0.5
    ring.renderOrder = 7
    group.add(ring)
    rings.push(ring)
  }

  routeRoot.add(group)
  return { group, pyramid, rings }
}

const updateRouteVisuals = () => {
  if (routeRoot) {
    scene.remove(routeRoot)
    disposeObject(routeRoot)
    routeRoot = null
  }
  routeAnim = null
  navArrow = null

  if (
    !props.route ||
    currentOverview ||
    props.route.fromBuilding !== currentBuildingId
  ) {
    return
  }

  const points = props.route.segments[currentFloorId]
  if (!points || points.length < 2) {
    return
  }

  routeRoot = new Group()
  routeRoot.renderOrder = 4
  scene.add(routeRoot)

  const positions = points.flatMap((point) => [point[0], point[1], point[2]])

  const casingGeometry = new LineGeometry()
  casingGeometry.setPositions(positions)
  const casingMaterial = new LineMaterial({
    color: COLORS.casing,
    linewidth: 10,
    depthTest: false
  })
  casingMaterial.resolution.set(viewportWidth, viewportHeight)
  const casing = new Line2(casingGeometry, casingMaterial)
  casing.computeLineDistances()
  casing.renderOrder = 4
  routeRoot.add(casing)

  const coreGeometry = new LineGeometry()
  coreGeometry.setPositions(positions)
  const coreMaterial = new LineMaterial({
    color: COLORS.route,
    linewidth: 6,
    depthTest: false
  })
  coreMaterial.resolution.set(viewportWidth, viewportHeight)
  const core = new Line2(coreGeometry, coreMaterial)
  core.computeLineDistances()
  core.renderOrder = 5
  routeRoot.add(core)

  const arrowGeometry = createArrowGeometry()
  const arrowMaterial = new MeshBasicMaterial({
    color: COLORS.casing,
    transparent: true,
    opacity: 0.95,
    depthTest: false,
    side: DoubleSide
  })
  const arrows = []
  for (let index = 0; index < 3; index += 1) {
    const arrow = new Mesh(arrowGeometry, arrowMaterial)
    arrow.position.y = 0.15
    arrow.renderOrder = 6
    arrow.visible = !navActive
    routeRoot.add(arrow)
    arrows.push(arrow)
  }

  // P3 跟随箭头：主色箭头 + 白色圆盘 + 主色圆环，保证浅色地面上也清晰
  if (navActive) {
    const navGroup = new Group()

    const disc = new Mesh(
      new CircleGeometry(1.05, 32),
      new MeshBasicMaterial({
        color: COLORS.casing,
        transparent: true,
        opacity: 1,
        depthTest: false
      })
    )
    disc.rotation.x = -Math.PI / 2
    disc.position.y = 0.14
    disc.renderOrder = 12
    navGroup.add(disc)

    const ring = new Mesh(
      new RingGeometry(1.05, 1.18, 40),
      new MeshBasicMaterial({
        color: COLORS.primary,
        transparent: true,
        opacity: 1,
        depthTest: false
      })
    )
    ring.rotation.x = -Math.PI / 2
    ring.position.y = 0.145
    ring.renderOrder = 12
    navGroup.add(ring)

    const navArrowMesh = new Mesh(
      createArrowGeometry(),
      new MeshBasicMaterial({
        color: COLORS.primary,
        transparent: true,
        opacity: 1,
        depthTest: false,
        side: DoubleSide
      })
    )
    navArrowMesh.scale.setScalar(1.45)
    navArrowMesh.position.y = 0.17
    navArrowMesh.renderOrder = 13
    navGroup.add(navArrowMesh)

    routeRoot.add(navGroup)
    navArrow = { group: navGroup, arrow: navArrowMesh }
    updateNavArrow()
  }

  const markers = []
  if (props.route.from.floor === currentFloorId) {
    markers.push(createMarker(props.route.from.center, COLORS.start))
  }
  if (props.route.to.floor === currentFloorId) {
    markers.push(createMarker(props.route.to.center, COLORS.end))
  }

  const cumulative = [0]
  for (let index = 1; index < points.length; index += 1) {
    const previous = points[index - 1]
    const current = points[index]
    cumulative.push(
      cumulative[index - 1] +
        Math.hypot(current[0] - previous[0], current[2] - previous[2])
    )
  }

  routeAnim = {
    points,
    cumulative,
    total: cumulative[cumulative.length - 1],
    arrows,
    markers
  }
}

const updateRouteAnimation = (elapsed) => {
  if (!routeAnim) {
    return
  }

  routeAnim.markers.forEach((marker, markerIndex) => {
    marker.pyramid.rotation.y = elapsed * 1.15 + markerIndex * Math.PI * 0.25
    marker.rings.forEach((ring) => {
      const progress = (elapsed * 0.55 + ring.userData.phase) % 1
      const scale = 1 + progress * 1.9
      ring.scale.setScalar(scale)
      ring.material.opacity = 0.38 * (1 - progress)
    })
  })

  if (navActive) {
    routeAnim.arrows.forEach((arrow) => {
      arrow.visible = false
    })
    return
  }

  if (routeAnim.total < 0.2) {
    return
  }

  routeAnim.arrows.forEach((arrow, index) => {
    arrow.visible = true
    const progress = (elapsed / 2.8 + index / 3) % 1
    const distance = progress * routeAnim.total
    let segmentIndex = 0
    while (
      segmentIndex < routeAnim.cumulative.length - 2 &&
      routeAnim.cumulative[segmentIndex + 1] < distance
    ) {
      segmentIndex += 1
    }

    const start = routeAnim.points[segmentIndex]
    const end = routeAnim.points[segmentIndex + 1] || start
    const segmentLength =
      routeAnim.cumulative[segmentIndex + 1] -
        routeAnim.cumulative[segmentIndex] || 1
    const ratio =
      (distance - routeAnim.cumulative[segmentIndex]) / segmentLength

    const x = start[0] + (end[0] - start[0]) * ratio
    const z = start[2] + (end[2] - start[2]) * ratio
    arrow.position.set(x, 0.15, z)
    arrow.rotation.y = Math.atan2(end[0] - start[0], end[2] - start[2])
  })
}

const applyWorldVisibility = () => {
  if (campusRoot) {
    campusRoot.visible = currentOverview
  }
  if (indoorRoot) {
    indoorRoot.visible = !currentOverview
  }
  if (routeRoot) {
    routeRoot.visible = !currentOverview
  }
  if (selectionRoot) {
    selectionRoot.visible = !currentOverview && !navActive
  }
  if (!currentOverview && hoveredBuilding) {
    hoveredBuilding.material.color.set(COLORS.floor)
    hoveredBuilding = null
  }
  if (renderer) {
    renderer.domElement.style.cursor = currentOverview ? 'default' : 'grab'
  }
}

const applyCamera = (mode, overview, animate = true) => {
  if (!controls) {
    return
  }

  // 若上一次缩放/定位动画仍在进行，先终止，避免 controls 一直处于禁用状态
  isCameraAnimating = false
  gsap.killTweensOf(activeCamera.position)
  gsap.killTweensOf(controls.target)

  currentMode = mode
  currentOverview = overview
  const wantsFollow =
    mode === '3d' &&
    !overview &&
    Boolean(props.navigation?.active && props.navigation?.following)
  const preset = wantsFollow ? getFollowPreset() : getPreset(mode, overview)
  const nextCamera = mode === '2d' ? orthographicCamera : perspectiveCamera
  const previousCamera = activeCamera

  if (nextCamera.isOrthographicCamera) {
    updateOrthoFrustum(nextCamera, preset.orthoSize)
    nextCamera.zoom = 1
    nextCamera.updateProjectionMatrix()
  }

  if (!animate) {
    activeCamera = nextCamera
    nextCamera.position.copy(preset.position)
    setCameraUpForMode(nextCamera, mode)
    controls.object = nextCamera
    controls.target.copy(preset.target)
    nextCamera.lookAt(controls.target)
    configureControls(mode)
    if (currentMode === '3d') {
      controls.update()
    }
    return
  }

  const startTarget = controls.target.clone()
  let startPosition = previousCamera.position.clone()

  // 从 2D 正俯视切回 3D 时，避免相机与目标共线导致朝向退化
  if (
    mode === '3d' &&
    Math.abs(startPosition.x - startTarget.x) < 0.01 &&
    Math.abs(startPosition.z - startTarget.z) < 0.01
  ) {
    startPosition = new Vector3(
      startTarget.x,
      startTarget.y + 58,
      startTarget.z + 36
    )
  }

  nextCamera.position.copy(startPosition)
  setCameraUpForMode(nextCamera, mode)
  nextCamera.lookAt(startTarget)
  activeCamera = nextCamera
  controls.object = nextCamera
  controls.enabled = false
  isTransitioning = true

  if (cameraTween) {
    cameraTween.kill()
  }
  if (targetTween) {
    targetTween.kill()
  }
  if (zoomTween) {
    zoomTween.kill()
  }

  const finish = () => {
    isTransitioning = false
    configureControls(mode)
    if (currentMode === '3d') {
      controls.update()
    }
    renderer.shadowMap.needsUpdate = true
  }

  cameraTween = gsap.to(nextCamera.position, {
    x: preset.position.x,
    y: preset.position.y,
    z: preset.position.z,
    duration: 0.7,
    ease: 'power2.inOut',
    onUpdate: () => nextCamera.lookAt(controls.target),
    onComplete: finish
  })

  targetTween = gsap.to(controls.target, {
    x: preset.target.x,
    y: preset.target.y,
    z: preset.target.z,
    duration: 0.7,
    ease: 'power2.inOut'
  })

  if (nextCamera.isOrthographicCamera) {
    zoomTween = gsap.fromTo(
      nextCamera,
      { zoom: 1 },
      {
        zoom: preset.zoom || 1,
        duration: 0.7,
        ease: 'power2.inOut',
        onUpdate: () => nextCamera.updateProjectionMatrix()
      }
    )
  }

  const targetBackground = new Color(
    mode === '2d' ? COLORS.bg2d : COLORS.bg3d
  )
  gsap.to(scene.background, {
    r: targetBackground.r,
    g: targetBackground.g,
    b: targetBackground.b,
    duration: 0.55,
    ease: 'power2.inOut'
  })

  renderer.shadowMap.enabled = mode === '3d'
}

const zoomBy = (factor) => {
  if (!controls || isTransitioning || isCameraAnimating) {
    return
  }

  isCameraAnimating = true
  controls.enabled = false
  const finish = () => {
    isCameraAnimating = false
    applyControlsEnabled()
    if (currentMode === '3d') {
      controls.update()
    }
  }

  if (activeCamera.isOrthographicCamera) {
    const nextZoom = Math.min(
      4,
      Math.max(0.55, activeCamera.zoom * factor)
    )
    if (zoomTween) {
      zoomTween.kill()
    }
    zoomTween = gsap.to(activeCamera, {
      zoom: nextZoom,
      duration: 0.22,
      ease: 'power2.out',
      onUpdate: () => activeCamera.updateProjectionMatrix(),
      onComplete: finish
    })
    return
  }

  const target = controls.target.clone()
  const direction = activeCamera.position.clone().sub(target)
  const currentDistance = direction.length()
  const nextDistance = Math.min(
    controls.maxDistance,
    Math.max(controls.minDistance, currentDistance / factor)
  )
  const nextPosition = target.clone().add(direction.setLength(nextDistance))

  gsap.to(activeCamera.position, {
    x: nextPosition.x,
    y: nextPosition.y,
    z: nextPosition.z,
    duration: 0.22,
    ease: 'power2.out',
    onUpdate: () => activeCamera.lookAt(target),
    onComplete: finish
  })
}

const zoomIn = () => zoomBy(1.2)
const zoomOut = () => zoomBy(1 / 1.2)

const resetCamera = () => {
  applyCamera(currentMode, currentOverview, true)
}

const locate = () => {
  if (!controls || isTransitioning || isCameraAnimating) {
    return
  }

  isCameraAnimating = true
  controls.enabled = false
  const target = props.route?.from?.center || [0, 0, 0]
  const nextTarget = new Vector3(target[0], 0, target[2])
  const offset = activeCamera.position.clone().sub(controls.target)
  const nextPosition = nextTarget.clone().add(offset)

  const finish = () => {
    isCameraAnimating = false
    applyControlsEnabled()
    if (currentMode === '3d') {
      controls.update()
    }
  }

  gsap.to(controls.target, {
    x: nextTarget.x,
    y: nextTarget.y,
    z: nextTarget.z,
    duration: 0.45,
    ease: 'power2.inOut'
  })
  gsap.to(activeCamera.position, {
    x: nextPosition.x,
    y: activeCamera.isOrthographicCamera
      ? activeCamera.position.y
      : nextPosition.y,
    z: nextPosition.z,
    duration: 0.45,
    ease: 'power2.inOut',
    onUpdate: () => activeCamera.lookAt(controls.target),
    onComplete: finish
  })
}

const updatePointer = (event) => {
  const rect = renderer.domElement.getBoundingClientRect()
  pointer.x = ((event.clientX - rect.left) / rect.width) * 2 - 1
  pointer.y = -((event.clientY - rect.top) / rect.height) * 2 + 1
  raycaster.setFromCamera(pointer, activeCamera)
}

const handlePointerMove = (event) => {
  if (!renderer) {
    return
  }

  if (is2dPanning && currentMode === '2d' && !isTransitioning) {
    const dx = event.clientX - last2dPointer.x
    const dy = event.clientY - last2dPointer.y
    last2dPointer = { x: event.clientX, y: event.clientY }
    const worldPerPixel =
      (orthographicCamera.top - orthographicCamera.bottom) /
      Math.max(orthographicCamera.zoom, 0.1) /
      Math.max(viewportHeight, 1)
    orthographicCamera.position.x -= dx * worldPerPixel
    orthographicCamera.position.z -= dy * worldPerPixel
    controls.target.x -= dx * worldPerPixel
    controls.target.z -= dy * worldPerPixel
    orthographicCamera.lookAt(controls.target)
    return
  }

  if (
    navPointerStart &&
    navActive &&
    props.navigation?.following &&
    !props.navigation?.paused &&
    !navInteractionEmitted
  ) {
    const moved = Math.hypot(
      event.clientX - navPointerStart.x,
      event.clientY - navPointerStart.y
    )
    if (moved > 6) {
      navInteractionEmitted = true
      emit('camera-interaction')
    }
  }

  if (!currentOverview) {
    return
  }

  updatePointer(event)
  const hits = raycaster.intersectObjects(
    campusBuildings.map((item) => item.mesh),
    false
  )
  const nextHovered = hits.length ? hits[0].object : null

  if (nextHovered !== hoveredBuilding) {
    if (hoveredBuilding) {
      hoveredBuilding.material.color.set(COLORS.floor)
    }
    hoveredBuilding = nextHovered
    if (hoveredBuilding) {
      hoveredBuilding.material.color.set(COLORS.primary100)
    }
    renderer.domElement.style.cursor = hoveredBuilding ? 'pointer' : 'default'
  }
}

const handlePointerDown = (event) => {
  if (currentMode === '2d' && !isTransitioning) {
    is2dPanning = true
    last2dPointer = { x: event.clientX, y: event.clientY }
  }
  pointerDown = { x: event.clientX, y: event.clientY }
  navPointerStart =
    navActive && props.navigation?.following && !props.navigation?.paused
      ? { x: event.clientX, y: event.clientY }
      : null
  navInteractionEmitted = false
  if (renderer) {
    renderer.domElement.style.cursor = currentOverview ? 'pointer' : 'grabbing'
  }
}

const handlePointerUp = (event) => {
  is2dPanning = false
  last2dPointer = null
  navPointerStart = null
  navInteractionEmitted = false

  if (renderer) {
    renderer.domElement.style.cursor = currentOverview
      ? hoveredBuilding
        ? 'pointer'
        : 'default'
      : 'grab'
  }

  if (!currentOverview || !pointerDown) {
    pointerDown = null
    return
  }

  const moved = Math.hypot(
    event.clientX - pointerDown.x,
    event.clientY - pointerDown.y
  )
  pointerDown = null
  if (moved > 6) {
    return
  }

  updatePointer(event)
  const hits = raycaster.intersectObjects(
    campusBuildings.map((item) => item.mesh),
    false
  )
  if (hits.length) {
    emit('building-select', hits[0].object.userData.buildingId)
  }
}

const resize = () => {
  if (!containerRef.value || !renderer) {
    return
  }
  const rect = containerRef.value.getBoundingClientRect()
  viewportWidth = Math.max(1, rect.width)
  viewportHeight = Math.max(1, rect.height)

  renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2))
  renderer.setSize(viewportWidth, viewportHeight, false)

  perspectiveCamera.aspect = viewportWidth / viewportHeight
  perspectiveCamera.updateProjectionMatrix()
  updateOrthoFrustum(
    orthographicCamera,
    currentOverview ? 96 : 84
  )

  if (routeRoot) {
    routeRoot.traverse((child) => {
      if (child.material?.isLineMaterial) {
        child.material.resolution.set(viewportWidth, viewportHeight)
        child.material.needsUpdate = true
      }
    })
  }
}

const disposeObject = (object) => {
  object.traverse((child) => {
    if (child.geometry) {
      child.geometry.dispose()
    }
    if (child.material) {
      const materials = Array.isArray(child.material)
        ? child.material
        : [child.material]
      materials.forEach((material) => {
        if (material.map) {
          material.map.dispose()
        }
        material.dispose()
      })
    }
  })
  object.clear()
}

const handleWheel = (event) => {
  if (currentMode === '2d' && !isTransitioning) {
    const factor = event.deltaY > 0 ? 1 / 1.12 : 1.12
    orthographicCamera.zoom = Math.min(
      4,
      Math.max(0.55, orthographicCamera.zoom * factor)
    )
    orthographicCamera.updateProjectionMatrix()
    return
  }
  if (
    navActive &&
    props.navigation?.following &&
    !props.navigation?.paused
  ) {
    emit('camera-interaction')
  }
}

const animate = () => {
  animationFrameId = requestAnimationFrame(animate)
  const elapsed = performance.now() / 1000

  const followOwnsCamera = Boolean(
    navActive && props.navigation?.following
  )

  if (
    !isTransitioning &&
    !isCameraAnimating &&
    !followOwnsCamera &&
    currentMode === '3d'
  ) {
    controls.update()
  }

  if (
    navActive &&
    props.navigation?.following &&
    !props.navigation?.paused &&
    !isTransitioning &&
    activeCamera?.isPerspectiveCamera
  ) {
    updateFollowCamera()
  }

  updateRouteAnimation(elapsed)
  updateNavArrow()
  renderer.render(scene, activeCamera)
}

const initialize = () => {
  const rect = containerRef.value.getBoundingClientRect()
  viewportWidth = Math.max(1, rect.width)
  viewportHeight = Math.max(1, rect.height)

  scene = new Scene()
  scene.background = new Color(
    currentMode === '2d' ? COLORS.bg2d : COLORS.bg3d
  )

  renderer = new WebGLRenderer({ antialias: true })
  renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2))
  renderer.setSize(viewportWidth, viewportHeight, false)
  renderer.outputColorSpace = SRGBColorSpace
  renderer.toneMapping = NoToneMapping
  renderer.shadowMap.enabled = currentMode === '3d'
  renderer.shadowMap.type = PCFShadowMap
  containerRef.value.appendChild(renderer.domElement)

  perspectiveCamera = new PerspectiveCamera(
    45,
    viewportWidth / viewportHeight,
    0.1,
    800
  )
  orthographicCamera = new OrthographicCamera(-1, 1, 1, -1, 0.1, 800)
  activeCamera = currentMode === '2d' ? orthographicCamera : perspectiveCamera

  const preset = getPreset(currentMode, currentOverview)
  updateOrthoFrustum(orthographicCamera, preset.orthoSize)

  controls = new OrbitControls(activeCamera, renderer.domElement)
  controls.enableDamping = true
  controls.dampingFactor = 0.08
  controls.minDistance = 16
  controls.maxDistance = 190
  controls.target.copy(preset.target)
  activeCamera.position.copy(preset.position)
  setCameraUpForMode(activeCamera, currentMode)
  activeCamera.lookAt(controls.target)
  configureControls(currentMode)
  if (currentMode === '3d') {
    controls.update()
  }

  const hemisphere = new HemisphereLight(0xffffff, 0xc7d4e4, 1.5)
  scene.add(hemisphere)

  const directional = new DirectionalLight(0xffffff, 2.1)
  directional.position.set(42, 62, 28)
  directional.castShadow = true
  directional.shadow.mapSize.set(2048, 2048)
  directional.shadow.camera.near = 1
  directional.shadow.camera.far = 220
  directional.shadow.camera.left = -76
  directional.shadow.camera.right = 76
  directional.shadow.camera.top = 76
  directional.shadow.camera.bottom = -76
  directional.shadow.bias = -0.0004
  directional.shadow.normalBias = 0.02
  scene.add(directional)

  const fill = new DirectionalLight(0xffffff, 0.65)
  fill.position.set(-36, 28, -42)
  scene.add(fill)

  buildCampus()
  buildIndoorWorld(currentBuildingId)
  applyWorldVisibility()
  updateRouteVisuals()

  renderer.domElement.addEventListener('pointermove', handlePointerMove)
  renderer.domElement.addEventListener('pointerdown', handlePointerDown)
  renderer.domElement.addEventListener('pointerup', handlePointerUp)
  renderer.domElement.addEventListener('wheel', handleWheel, { passive: true })

  resizeObserver = new ResizeObserver(resize)
  resizeObserver.observe(containerRef.value)

  animate()
}

const dispose = () => {
  cancelAnimationFrame(animationFrameId)

  if (cameraTween) {
    cameraTween.kill()
  }
  if (targetTween) {
    targetTween.kill()
  }
  if (zoomTween) {
    zoomTween.kill()
  }

  if (resizeObserver) {
    resizeObserver.disconnect()
    resizeObserver = null
  }

  if (renderer) {
    renderer.domElement.removeEventListener('pointermove', handlePointerMove)
    renderer.domElement.removeEventListener('pointerdown', handlePointerDown)
    renderer.domElement.removeEventListener('pointerup', handlePointerUp)
    renderer.domElement.removeEventListener('wheel', handleWheel)
  }

  if (controls) {
    controls.dispose()
    controls = null
  }

  if (scene) {
    disposeObject(scene)
    scene = null
  }

  if (renderer) {
    renderer.dispose()
    if (renderer.domElement.parentNode) {
      renderer.domElement.parentNode.removeChild(renderer.domElement)
    }
    renderer = null
  }

  indoorRoot = null
  campusRoot = null
  routeRoot = null
  routeAnim = null
  floorGroups.clear()
  campusBuildings.length = 0
}

onMounted(initialize)
onBeforeUnmount(dispose)

watch(
  () => props.viewMode,
  (mode) => {
    applyCamera(mode, currentOverview, true)
  }
)

watch(
  () => props.overview,
  (overview) => {
    currentOverview = overview
    applyWorldVisibility()
    updateRouteVisuals()
    updateSelectionMarker()
    applyCamera(currentMode, overview, true)
  }
)

watch(
  () => props.currentFloor,
  (floorId) => {
    currentFloorId = floorId
    floorGroups.forEach((group, id) => {
      group.visible = id === floorId
    })
    updateRouteVisuals()
    updateSelectionMarker()
  }
)

watch(
  () => props.building,
  (buildingId) => {
    currentBuildingId = buildingId
    buildIndoorWorld(buildingId)
    applyWorldVisibility()
    updateSelectionMarker()
  }
)

watch(
  () => props.route,
  () => {
    updateRouteVisuals()
    applyWorldVisibility()
  },
  { deep: true }
)

watch(
  () => props.navigation?.active,
  (active) => {
    navActive = Boolean(active)
    if (!controls) {
      return
    }

    if (navActive) {
      controls.enabled = false
      updateRouteVisuals()
      updateSelectionMarker()
      updateNavArrow()
      return
    }

    applyControlsEnabled()
    updateRouteVisuals()
    updateSelectionMarker()
    applyCamera(currentMode, currentOverview, true)
  }
)

watch(
  () => props.navigation?.following,
  (following) => {
    if (!navActive || !controls) {
      return
    }
    applyControlsEnabled()
  }
)

watch(
  () => props.selectedLocation,
  () => {
    updateSelectionMarker()
  },
  { deep: true }
)

defineExpose({
  zoomIn,
  zoomOut,
  resetCamera,
  locate,
  focusLocation
})
</script>

<template>
  <div ref="containerRef" class="map-canvas"></div>
</template>

<style scoped lang="less">
.map-canvas {
  position: absolute;
  inset: 0;
  overflow: hidden;
  background: var(--color-map-3d-bg);

  :deep(canvas) {
    display: block;
    width: 100%;
    height: 100%;
  }
}
</style>
