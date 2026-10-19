import { onBeforeUnmount, reactive, watch } from 'vue'

const WALK_SPEED = 3.2
const TRANSFER_DURATION = 2500
const HINT_DURATION = 2400
const TURN_ANGLE = 25
const TURN_LOOKAHEAD = 8
const ARRIVAL_LOOKAHEAD = 5

const clamp = (value, min, max) => Math.min(max, Math.max(min, value))

const distance2D = (a, b) => Math.hypot(a[0] - b[0], a[2] - b[2])

const polylineLength = (points) => {
  let total = 0
  for (let index = 1; index < points.length; index += 1) {
    total += distance2D(points[index - 1], points[index])
  }
  return total
}

const buildCumulative = (points) => {
  const cumulative = [0]
  for (let index = 1; index < points.length; index += 1) {
    cumulative.push(cumulative[index - 1] + distance2D(points[index - 1], points[index]))
  }
  return cumulative
}

const sampleWalk = (phase, distance) => {
  const { points, cumulative } = phase
  let segmentIndex = 0
  while (
    segmentIndex < cumulative.length - 2 &&
    cumulative[segmentIndex + 1] < distance
  ) {
    segmentIndex += 1
  }

  const start = points[segmentIndex]
  const end = points[segmentIndex + 1] || start
  const segmentLength =
    cumulative[segmentIndex + 1] - cumulative[segmentIndex] || 1
  const progress = clamp(
    (distance - cumulative[segmentIndex]) / segmentLength,
    0,
    1
  )

  const x = start[0] + (end[0] - start[0]) * progress
  const z = start[2] + (end[2] - start[2]) * progress
  const dx = end[0] - start[0]
  const dz = end[2] - start[2]
  const length = Math.hypot(dx, dz) || 1

  return {
    position: [x, 0, z],
    direction: [dx / length, 0, dz / length],
    segmentIndex,
    remainingInSegment: segmentLength * (1 - progress)
  }
}

const buildWalkAction = (phase, distance, isLastPhase) => {
  const { points, total } = phase
  if (isLastPhase && total - distance <= ARRIVAL_LOOKAHEAD) {
    return { text: '即将到达', icon: 'flag' }
  }

  const sample = sampleWalk(phase, distance)
  const segmentIndex = sample.segmentIndex
  if (sample.remainingInSegment > TURN_LOOKAHEAD) {
    return { text: '直行', icon: 'straight' }
  }

  const p0 = points[segmentIndex]
  const p1 = points[segmentIndex + 1]
  const p2 = points[segmentIndex + 2]
  if (!p0 || !p1 || !p2) {
    return { text: '直行', icon: 'straight' }
  }

  const d1x = p1[0] - p0[0]
  const d1z = p1[2] - p0[2]
  const d2x = p2[0] - p1[0]
  const d2z = p2[2] - p1[2]
  const length1 = Math.hypot(d1x, d1z)
  const length2 = Math.hypot(d2x, d2z)
  if (!length1 || !length2) {
    return { text: '直行', icon: 'straight' }
  }

  const cos = clamp((d1x * d2x + d1z * d2z) / (length1 * length2), -1, 1)
  const angle = (Math.acos(cos) * 180) / Math.PI
  if (angle < TURN_ANGLE) {
    return { text: '直行', icon: 'straight' }
  }

  // XZ 平面叉积：正值对应左转
  const cross = d1z * d2x - d1x * d2z
  return cross > 0
    ? { text: '前方左转', icon: 'turn-left' }
    : { text: '前方右转', icon: 'turn-right' }
}

export function useNavigationSession({ route, viewMode, currentFloor }) {
  const navigation = reactive({
    active: false,
    paused: false,
    following: true,
    arrived: false,
    transfer: false,
    floor: 'F1',
    floorLabel: '',
    action: '直行',
    actionIcon: 'straight',
    remaining: 0,
    progress: 0,
    position: [0, 0, 0],
    direction: [0, 0, 1],
    hint: '',
    hintType: 'primary'
  })

  let phases = []
  let phaseIndex = 0
  let phaseElapsed = 0
  let totalDuration = 0
  let rafId = 0
  let lastTimestamp = 0
  let uiAccumulator = 0
  let hintTimer = null

  const clearHintTimer = () => {
    if (hintTimer) {
      clearTimeout(hintTimer)
      hintTimer = null
    }
  }

  const showHint = (text, type = 'primary') => {
    navigation.hint = text
    navigation.hintType = type
    clearHintTimer()
    hintTimer = setTimeout(() => {
      navigation.hint = ''
    }, HINT_DURATION)
  }

  const syncFloorLabel = () => {
    const currentRoute = route.value
    if (!currentRoute) {
      navigation.floorLabel = ''
      return
    }
    if (currentRoute.fromFloor === currentRoute.toFloor) {
      navigation.floorLabel = `${navigation.floor} 同层`
    } else if (navigation.floor === currentRoute.toFloor) {
      navigation.floorLabel = `${navigation.floor} 目标层`
    } else {
      navigation.floorLabel = `${navigation.floor} → ${currentRoute.toFloor}`
    }
  }

  const buildPhases = (currentRoute) => {
    const result = []
    const floorKeys = Object.keys(currentRoute.segments || {}).filter(
      (floor) =>
        Array.isArray(currentRoute.segments[floor]) &&
        currentRoute.segments[floor].length >= 2
    )

    floorKeys.forEach((floor, index) => {
      const points = currentRoute.segments[floor]
      const total = polylineLength(points)
      result.push({
        type: 'walk',
        floor,
        points,
        total,
        cumulative: buildCumulative(points),
        duration: Math.max(4500, (total / WALK_SPEED) * 1000)
      })

      if (index < floorKeys.length - 1) {
        result.push({
          type: 'transfer',
          floor,
          nextFloor: floorKeys[index + 1],
          via: currentRoute.via,
          duration: TRANSFER_DURATION
        })
      }
    })

    return result
  }

  const updatePosition = () => {
    if (!navigation.active || !phases.length) {
      return
    }
    const phase = phases[phaseIndex]
    if (!phase || phase.type !== 'walk') {
      return
    }

    const distance = phase.total * clamp(phaseElapsed / phase.duration, 0, 1)
    const sample = sampleWalk(phase, distance)
    navigation.position = sample.position
    navigation.direction = sample.direction
  }

  const updateUi = () => {
    if (!navigation.active) {
      return
    }
    const currentRoute = route.value
    if (!currentRoute) {
      return
    }

    const phase = phases[phaseIndex]
    const completedDuration = phases
      .slice(0, phaseIndex)
      .reduce((total, item) => total + item.duration, 0)
    const currentDuration = phase ? phase.duration : 0
    const currentFraction = phase
      ? clamp(phaseElapsed / phase.duration, 0, 1)
      : 0
    const progress =
      totalDuration > 0
        ? clamp(
            (completedDuration + currentDuration * currentFraction) /
              totalDuration,
            0,
            1
          )
        : 0

    navigation.progress = progress
    if (!navigation.arrived) {
      navigation.remaining = Math.max(
        0,
        Math.round(currentRoute.distance * (1 - progress))
      )
    }

    syncFloorLabel()

    if (
      !navigation.transfer &&
      !navigation.arrived &&
      phase &&
      phase.type === 'walk'
    ) {
      const distance = phase.total * currentFraction
      const action = buildWalkAction(
        phase,
        distance,
        phaseIndex === phases.length - 1
      )
      navigation.action = action.text
      navigation.actionIcon = action.icon
    }
  }

  const finishArrival = () => {
    navigation.arrived = true
    navigation.paused = true
    navigation.transfer = false
    navigation.progress = 1
    navigation.remaining = 0
    navigation.action = `已到达 ${route.value?.toName || ''}`
    navigation.actionIcon = 'flag'
    syncFloorLabel()
    showHint('已到达目的地 · 可结束导航', 'primary')
    updateUi()
  }

  const startTransfer = (nextPhase) => {
    const viaLabel = nextPhase.via === 'elevator' ? '电梯' : '楼梯'
    navigation.transfer = true
    navigation.action = `乘${viaLabel}到 ${nextPhase.nextFloor}`
    navigation.actionIcon =
      nextPhase.via === 'elevator' ? 'elevator' : 'stairs'
    showHint(
      `已到达换乘点 · 请乘${viaLabel}前往 ${nextPhase.nextFloor}`,
      'warning'
    )
    updateUi()
  }

  const advance = (delta) => {
    if (!phases.length) {
      return
    }

    phaseElapsed += delta
    const phase = phases[phaseIndex]
    if (!phase || phaseElapsed < phase.duration) {
      return
    }

    if (phase.type === 'walk') {
      const nextPhase = phases[phaseIndex + 1]
      if (!nextPhase) {
        finishArrival()
        return
      }

      phaseIndex += 1
      phaseElapsed = 0

      if (nextPhase.type === 'transfer') {
        startTransfer(nextPhase)
        return
      }

      navigation.transfer = false
      navigation.floor = nextPhase.floor
      currentFloor.value = nextPhase.floor
      navigation.action = '直行'
      navigation.actionIcon = 'straight'
      updatePosition()
      updateUi()
      return
    }

    if (phase.type === 'transfer') {
      const nextPhase = phases[phaseIndex + 1]
      phaseIndex += 1
      phaseElapsed = 0
      navigation.transfer = false

      if (!nextPhase) {
        finishArrival()
        return
      }

      navigation.floor = nextPhase.floor
      currentFloor.value = nextPhase.floor
      navigation.action = '直行'
      navigation.actionIcon = 'straight'
      updatePosition()
      updateUi()
    }
  }

  const tick = (timestamp) => {
    if (!navigation.active) {
      return
    }
    rafId = requestAnimationFrame(tick)

    if (!lastTimestamp) {
      lastTimestamp = timestamp
    }
    const delta = Math.min(64, timestamp - lastTimestamp)
    lastTimestamp = timestamp

    if (!navigation.paused && navigation.following && !navigation.arrived) {
      advance(delta)
    }

    updatePosition()

    uiAccumulator += delta
    if (uiAccumulator >= 100) {
      updateUi()
      uiAccumulator = 0
    }
  }

  const resetInternal = () => {
    if (rafId) {
      cancelAnimationFrame(rafId)
      rafId = 0
    }
    clearHintTimer()
    phases = []
    phaseIndex = 0
    phaseElapsed = 0
    totalDuration = 0
    lastTimestamp = 0
    uiAccumulator = 0
  }

  const start = (currentRoute) => {
    resetInternal()
    if (!currentRoute) {
      return
    }

    const nextPhases = buildPhases(currentRoute)
    if (!nextPhases.length) {
      return
    }

    phases = nextPhases
    phaseIndex = 0
    phaseElapsed = 0
    totalDuration = phases.reduce((total, phase) => total + phase.duration, 0)

    navigation.active = true
    navigation.paused = false
    navigation.following = true
    navigation.arrived = false
    navigation.transfer = false
    navigation.floor = currentRoute.fromFloor
    navigation.action = '直行'
    navigation.actionIcon = 'straight'
    navigation.remaining = currentRoute.distance
    navigation.progress = 0
    navigation.hint = ''
    navigation.hintType = 'primary'
    navigation.position = [...currentRoute.from.center]
    navigation.direction = [0, 0, 1]

    currentFloor.value = currentRoute.fromFloor
    syncFloorLabel()

    if (viewMode.value !== '3d') {
      viewMode.value = '3d'
    }

    updatePosition()
    updateUi()

    rafId = requestAnimationFrame(tick)
  }

  const togglePause = () => {
    if (!navigation.active || navigation.arrived) {
      return
    }
    navigation.paused = !navigation.paused
  }

  const exitFollow = () => {
    if (!navigation.active || navigation.arrived) {
      return
    }
    navigation.following = false
    navigation.paused = true
    showHint('已退出跟随 · 自由浏览', 'warning')
  }

  const resumeFollow = () => {
    if (!navigation.active || navigation.arrived) {
      return
    }
    navigation.following = true
    navigation.paused = false
    if (viewMode.value !== '3d') {
      viewMode.value = '3d'
    }
  }

  const handleCameraInteraction = () => {
    if (
      !navigation.active ||
      !navigation.following ||
      navigation.paused ||
      navigation.arrived
    ) {
      return
    }
    navigation.following = false
    navigation.paused = true
    showHint('已偏离跟随 · 当前为自由浏览', 'warning')
  }

  const end = () => {
    resetInternal()
    navigation.active = false
    navigation.paused = false
    navigation.following = true
    navigation.arrived = false
    navigation.transfer = false
    navigation.progress = 0
    navigation.remaining = 0
    navigation.action = '直行'
    navigation.actionIcon = 'straight'
    navigation.hint = ''
    navigation.hintType = 'primary'
    navigation.position = [0, 0, 0]
    navigation.direction = [0, 0, 1]
  }

  watch(viewMode, (mode) => {
    if (!navigation.active || navigation.arrived) {
      return
    }

    if (mode === '2d' && navigation.following) {
      navigation.following = false
      navigation.paused = true
      showHint('已暂停跟随 · 2D 俯视', 'primary')
    } else if (mode === '3d' && !navigation.following) {
      navigation.following = true
      navigation.paused = false
    }
  })

  onBeforeUnmount(() => {
    resetInternal()
  })

  return {
    navigation,
    start,
    togglePause,
    exitFollow,
    resumeFollow,
    end,
    handleCameraInteraction
  }
}
