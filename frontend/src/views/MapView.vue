<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import LocationDetailSheet from '../components/detail/LocationDetailSheet.vue'
import MapIcon from '../components/MapIcon.vue'
import NavigationOverlay from '../components/navigation/NavigationOverlay.vue'
import SearchDrawer from '../components/search/SearchDrawer.vue'
import SettingsOverlay from '../components/settings/SettingsOverlay.vue'
import ThreeMap from '../components/map/ThreeMap.vue'
import { useNavigationSession } from '../composables/useNavigationSession'
import {
  allLocations,
  buildRoute,
  buildings,
  findLocation,
  floors,
  getNearestLocation
} from '../mock/indoorData'

const mapRef = ref(null)

const currentFloor = ref('F1')
const currentBuilding = ref('A')
const viewMode = ref('3d')
const overviewMode = ref(false)
const route = ref(null)
const legendCollapsed = ref(
  typeof window !== 'undefined' ? window.innerWidth < 1280 : false
)

const form = reactive({
  from: '',
  to: '',
  elevatorOnly: false
})

const errors = reactive({
  from: '',
  to: '',
  form: ''
})

const toastMessage = ref('')
let toastTimer = null

const searchOpen = ref(false)
const searchQuery = ref('')
const activeSearchCategory = ref('all')
const highlightedSearchId = ref('')
const selectedLocation = ref(null)
const p5Open = ref(false)
const settingsOpen = ref(false)
const searchOriginBuilding = ref(currentBuilding.value)
const searchOriginFloor = ref(currentFloor.value)
const mockCurrentLocation = ref(null)
const destinationInputRef = ref(null)
const searchEntryButtonRef = ref(null)
let p5Timer = null

const {
  navigation,
  start: startNavigationSession,
  togglePause: toggleNavigationPause,
  exitFollow: exitNavigationFollow,
  resumeFollow: resumeNavigationFollow,
  end: endNavigationSession,
  handleCameraInteraction
} = useNavigationSession({ route, viewMode, currentFloor })

const floorOptions = computed(() => [...floors].reverse())

const searchOriginBuildingName = computed(
  () =>
    buildings.find(
      (building) => building.id === searchOriginBuilding.value
    )?.name || ''
)

const searchOriginLevel = computed(
  () =>
    floors.find((floor) => floor.id === searchOriginFloor.value)?.level || 1
)

const buildingNameMap = new Map(
  buildings.map((building) => [building.id, building.name])
)

const sortSearchItems = (items) =>
  [...items].sort(
    (first, second) =>
      first.distance - second.distance ||
      first.floorLevel - second.floorLevel ||
      first.name.localeCompare(second.name)
  )

const searchItems = computed(() =>
  allLocations.map((location) => {
    const levelDelta = Math.abs(
      location.floorLevel - searchOriginLevel.value
    )
    const buildingDelta =
      location.building === searchOriginBuilding.value ? 0 : 120
    return {
      ...location,
      buildingName: buildingNameMap.get(location.building) || location.building,
      distance:
        Math.round(
          Math.hypot(location.center[0], location.center[2]) * 2 +
            levelDelta * 25 +
            buildingDelta
        )
    }
  })
)

const searchKeyword = computed(() => searchQuery.value.trim().toUpperCase())

const suggestionRank = (location) => {
  const keyword = searchKeyword.value
  const id = location.id.toUpperCase()
  const name = location.name.toUpperCase()
  if (id === keyword) {
    return 0
  }
  if (id.startsWith(keyword)) {
    return 1
  }
  if (name.startsWith(keyword)) {
    return 2
  }
  if (id.includes(keyword)) {
    return 3
  }
  if (name.includes(keyword)) {
    return 4
  }
  return 5
}

const searchFiltered = computed(() => {
  const legacyKeyword = searchKeyword.value.replace(/^A(?=\d)/, '')
  return searchItems.value.filter((location) => {
    if (
      activeSearchCategory.value !== 'all' &&
      location.type !== activeSearchCategory.value
    ) {
      return false
    }
    if (!searchKeyword.value) {
      return true
    }
    return (
      location.id.toUpperCase().includes(searchKeyword.value) ||
      location.name.toUpperCase().includes(searchKeyword.value) ||
      (location.building === 'A' &&
        legacyKeyword !== searchKeyword.value &&
        location.id.toUpperCase().includes(legacyKeyword))
    )
  })
})

const searchResults = computed(() => sortSearchItems(searchFiltered.value))

const searchSuggestions = computed(() => {
  if (!searchKeyword.value) {
    return []
  }
  return [...searchFiltered.value]
    .sort(
      (first, second) =>
        suggestionRank(first) - suggestionRank(second) ||
        first.distance - second.distance
    )
    .slice(0, 6)
})

const nearbyItems = [
  { type: 'restroom', label: '卫生间', icon: 'toilet' },
  { type: 'elevator', label: '电梯', icon: 'elevator' },
  { type: 'stairs', label: '楼梯', icon: 'stairs' },
  { type: 'exit', label: '安全出口', icon: 'exit' },
  { type: 'water', label: '开水间', icon: 'water' }
]

const legendItems = [
  { label: '教室', className: 'classroom' },
  { label: '楼梯', className: 'stairs' },
  { label: '电梯', className: 'elevator' },
  { label: '卫生间', className: 'restroom' }
]

const showToast = (message) => {
  toastMessage.value = message
  if (toastTimer) {
    clearTimeout(toastTimer)
  }
  toastTimer = setTimeout(() => {
    toastMessage.value = ''
  }, 2400)
}

const clearError = (field) => {
  if (field === 'from' || field === 'to') {
    errors[field] = ''
  } else {
    errors.form = ''
  }
}

const handleFromInput = () => {
  clearError('from')
  if (form.from.trim() !== '当前位置') {
    mockCurrentLocation.value = null
  }
}

const handleFromClear = () => {
  clearError('from')
  mockCurrentLocation.value = null
}

const clearAllErrors = () => {
  errors.from = ''
  errors.to = ''
  errors.form = ''
}

const clearRoute = () => {
  route.value = null
}

const handlePlan = () => {
  clearAllErrors()

  const fromValue = form.from.trim()
  const toValue = form.to.trim()

  if (!fromValue) {
    errors.from = '请选择出发地'
  }
  if (!toValue) {
    errors.to = '请选择目的地'
  }
  if (!fromValue || !toValue) {
    clearRoute()
    return
  }

  const from =
    fromValue === '当前位置' && mockCurrentLocation.value
      ? mockCurrentLocation.value
      : findLocation(fromValue, currentBuilding.value, currentFloor.value)
  const to = findLocation(toValue, currentBuilding.value, currentFloor.value)

  if (!from) {
    errors.from = '未找到该地点，请检查教室号'
  }
  if (!to) {
    errors.to = '未找到该教室，请检查教室号'
  }
  if (!from || !to) {
    clearRoute()
    return
  }

  if (from.building !== to.building) {
    errors.to = '当前版本仅支持同一教学楼内导航'
    clearRoute()
    return
  }

  if (from.id === to.id && from.floor === to.floor) {
    errors.to = '起点和终点不能相同'
    clearRoute()
    return
  }

  const nextRoute = buildRoute(from, to, form.elevatorOnly)
  if (nextRoute.error) {
    errors.to = nextRoute.error
    clearRoute()
    return
  }

  currentBuilding.value = from.building
  currentFloor.value = from.floor
  route.value = nextRoute
}

const selectFloor = (floorId) => {
  if (currentFloor.value === floorId) {
    return
  }
  currentFloor.value = floorId
}

const setViewMode = (mode) => {
  if (viewMode.value === mode) {
    return
  }
  viewMode.value = mode
}

const setOverview = (value) => {
  if (overviewMode.value === value) {
    return
  }
  overviewMode.value = value
}

const handleBuildingSelect = (buildingId) => {
  const building = buildings.find((item) => item.id === buildingId)
  if (!building) {
    return
  }
  currentBuilding.value = buildingId
  currentFloor.value = 'F1'
  overviewMode.value = false
  clearRoute()
  showToast(`已进入 ${building.name} 1F`)
}

const handleNearby = (type) => {
  const location = getNearestLocation(
    type,
    currentBuilding.value,
    currentFloor.value
  )
  if (!location) {
    showToast('当前楼层暂无该设施')
    return
  }
  form.to = location.name
  clearError('to')
  handlePlan()
}

const handleMapPick = () => {
  showToast('地图选点将在后续版本接入')
}

const startNavigation = () => {
  if (!route.value) {
    return
  }
  startNavigationSession(route.value)
}

const openSearch = () => {
  if (navigation.active) {
    return
  }
  if (p5Timer) {
    clearTimeout(p5Timer)
    p5Timer = null
  }
  searchOriginBuilding.value = currentBuilding.value
  searchOriginFloor.value = currentFloor.value
  p5Open.value = false
  highlightedSearchId.value = ''
  searchOpen.value = true
}

const closeSearch = () => {
  searchOpen.value = false
  highlightedSearchId.value = ''
}

const handleSearchQuery = (value) => {
  searchQuery.value = value
  highlightedSearchId.value = ''
}

const handleSearchCategory = (key) => {
  activeSearchCategory.value = key
  // 分类浏览时清空关键词，避免旧关键词与分类叠加导致 0 结果
  searchQuery.value = ''
  highlightedSearchId.value = ''
}

const handleSearchSuggestion = (location) => {
  searchQuery.value = location.id
  if (
    activeSearchCategory.value !== 'all' &&
    activeSearchCategory.value !== location.type
  ) {
    activeSearchCategory.value = 'all'
  }
  highlightedSearchId.value = location.id
}

const clearSearchFilters = () => {
  searchQuery.value = ''
  activeSearchCategory.value = 'all'
  highlightedSearchId.value = ''
}

const selectSearchLocation = (location) => {
  selectedLocation.value = location
  currentBuilding.value = location.building
  currentFloor.value = location.floor
  searchOpen.value = false
  highlightedSearchId.value = ''

  if (p5Timer) {
    clearTimeout(p5Timer)
  }
  p5Open.value = false
  p5Timer = setTimeout(() => {
    mapRef.value?.focusLocation(location)
    p5Open.value = true
    p5Timer = null
  }, 240)
}

const closeLocationPreview = () => {
  p5Open.value = false
  nextTick(() => {
    searchEntryButtonRef.value?.focus()
  })
}

const createMockCurrentLocation = (buildingId, floorId) => {
  const floor = floors.find((item) => item.id === floorId)
  return {
    id: `CURRENT-${buildingId}-${floorId}`,
    name: '当前位置',
    type: 'current',
    building: buildingId,
    floor: floorId,
    floorLevel: floor?.level || 1,
    center: [0, 0, 0],
    door: [0, 0, 0],
    size: [0, 0]
  }
}

const handleNavigateHere = () => {
  const location = selectedLocation.value
  if (!location) {
    return
  }

  const fromValue = form.from.trim()
  if (!fromValue || fromValue === '当前位置') {
    mockCurrentLocation.value = createMockCurrentLocation(
      searchOriginBuilding.value,
      searchOriginFloor.value
    )
    form.from = '当前位置'
  } else {
    mockCurrentLocation.value = null
  }

  form.to = location.name
  clearAllErrors()
  p5Open.value = false
  nextTick(() => {
    handlePlan()
  })
}

const handleStartHere = () => {
  const location = selectedLocation.value
  if (!location) {
    return
  }

  mockCurrentLocation.value = null
  form.from = location.name
  form.to = ''
  clearAllErrors()
  clearRoute()
  p5Open.value = false
  showToast('已设为出发点')
  nextTick(() => {
    destinationInputRef.value?.focus()
  })
}

const handleZoomIn = () => {
  mapRef.value?.zoomIn()
}

const handleZoomOut = () => {
  mapRef.value?.zoomOut()
}

const handleReset = () => {
  mapRef.value?.resetCamera()
}

const handleLocate = () => {
  mapRef.value?.locate()
}

const handleWindowResize = () => {
  if (window.innerWidth < 1280) {
    legendCollapsed.value = true
  }
}

onMounted(() => {
  window.addEventListener('resize', handleWindowResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleWindowResize)
  if (toastTimer) {
    clearTimeout(toastTimer)
  }
  if (p5Timer) {
    clearTimeout(p5Timer)
  }
})
</script>

<template>
  <main
    class="map-page"
    :class="{
      'is-navigating': navigation.active,
      'is-searching': searchOpen,
      'is-inspecting': p5Open
    }"
  >
    <ThreeMap
      ref="mapRef"
      :view-mode="viewMode"
      :current-floor="currentFloor"
      :building="currentBuilding"
      :route="route"
      :overview="overviewMode"
      :navigation="navigation"
      :selected-location="selectedLocation"
      @building-select="handleBuildingSelect"
      @camera-interaction="handleCameraInteraction"
    />

    <div class="overview-switch glass-panel" role="group" aria-label="视图范围">
      <button
        type="button"
        :class="{ active: overviewMode }"
        :aria-pressed="overviewMode"
        @click="setOverview(true)"
      >
        校园总览
      </button>
      <button
        type="button"
        :class="{ active: !overviewMode }"
        :aria-pressed="!overviewMode"
        @click="setOverview(false)"
      >
        楼内
      </button>
    </div>

    <Transition name="toast">
      <div
        v-if="toastMessage"
        class="map-toast glass-panel glass-panel--strong"
        role="status"
      >
        {{ toastMessage }}
      </div>
    </Transition>

    <section
      v-if="!overviewMode"
      class="search-panel glass-panel"
      aria-label="路线搜索"
    >
      <header class="search-panel__head">
        <h1 class="search-panel__title">
          路线
          <span class="search-panel__title-en">SEARCH</span>
        </h1>
        <div class="search-panel__actions">
          <button
            ref="searchEntryButtonRef"
            type="button"
            class="search-entry"
            @click="openSearch"
          >
            <MapIcon name="search" :size="14" />
            <span>分类查找</span>
          </button>
          <button
            type="button"
            class="search-settings"
            title="设置"
            aria-label="打开设置"
            @click="settingsOpen = true"
          >
            <MapIcon name="settings" :size="16" />
          </button>
        </div>
      </header>

      <form class="search-form" @submit.prevent="handlePlan">
        <div class="field" :class="{ 'is-error': errors.from }">
          <el-input
            v-model="form.from"
            clearable
            autocomplete="off"
            placeholder="出发地，如 A101"
            aria-label="出发地"
            @input="handleFromInput"
            @clear="handleFromClear"
          >
            <template #prefix>
              <span class="input-dot input-dot--start" aria-hidden="true"></span>
            </template>
            <template #suffix>
              <button
                type="button"
                class="pin-button"
                title="地图选点"
                aria-label="在地图上选择出发地"
                @click="handleMapPick"
              >
                <MapIcon name="pin" :size="16" />
              </button>
            </template>
          </el-input>
          <p v-if="errors.from" class="field-error" role="alert">
            {{ errors.from }}
          </p>
        </div>

        <div class="field" :class="{ 'is-error': errors.to }">
          <el-input
            ref="destinationInputRef"
            v-model="form.to"
            clearable
            autocomplete="off"
            placeholder="目的地，如 A305"
            aria-label="目的地"
            @input="clearError('to')"
            @clear="clearError('to')"
          >
            <template #prefix>
              <span class="input-dot input-dot--end" aria-hidden="true"></span>
            </template>
            <template #suffix>
              <button
                type="button"
                class="pin-button"
                title="地图选点"
                aria-label="在地图上选择目的地"
                @click="handleMapPick"
              >
                <MapIcon name="pin" :size="16" />
              </button>
            </template>
          </el-input>
          <p v-if="errors.to" class="field-error" role="alert">
            {{ errors.to }}
          </p>
        </div>

        <div class="switch-row">
          <div class="switch-row__copy">
            <span class="switch-row__label">仅电梯</span>
            <span class="switch-row__hint">
              {{ form.elevatorOnly ? '优先使用电梯' : '楼梯 + 电梯结合' }}
            </span>
          </div>
          <el-switch
            v-model="form.elevatorOnly"
            :width="36"
            aria-label="仅电梯"
          />
        </div>

        <el-button
          type="primary"
          class="plan-button"
          native-type="submit"
        >
          开始规划
        </el-button>

        <p v-if="errors.form" class="field-error field-error--form" role="alert">
          {{ errors.form }}
        </p>

        <div class="nearby">
          <p class="nearby__label">附近查找</p>
          <div class="nearby__grid">
            <button
              v-for="item in nearbyItems"
              :key="item.type"
              type="button"
              class="nearby__item"
              @click="handleNearby(item.type)"
            >
              <MapIcon :name="item.icon" :size="20" />
              <span>{{ item.label }}</span>
            </button>
          </div>
        </div>
      </form>
    </section>

    <div class="map-toolbar glass-panel" aria-label="地图工具">
      <div class="view-switch" role="group" aria-label="视角模式">
        <button
          type="button"
          :class="{ active: viewMode === '2d' }"
          :aria-pressed="viewMode === '2d'"
          @click="setViewMode('2d')"
        >
          2D
        </button>
        <button
          type="button"
          :class="{ active: viewMode === '3d' }"
          :aria-pressed="viewMode === '3d'"
          @click="setViewMode('3d')"
        >
          3D
        </button>
      </div>

      <span class="toolbar-divider" aria-hidden="true"></span>

      <button
        type="button"
        class="tool-button"
        title="放大"
        aria-label="放大"
        @click="handleZoomIn"
      >
        <MapIcon name="plus" :size="20" />
      </button>
      <button
        type="button"
        class="tool-button"
        title="缩小"
        aria-label="缩小"
        @click="handleZoomOut"
      >
        <MapIcon name="minus" :size="20" />
      </button>
      <button
        type="button"
        class="tool-button"
        title="复位"
        aria-label="复位视角"
        @click="handleReset"
      >
        <MapIcon name="reset" :size="20" />
      </button>
      <button
        type="button"
        class="tool-button"
        title="定位"
        aria-label="定位到起点"
        @click="handleLocate"
      >
        <MapIcon name="locate" :size="20" />
      </button>
    </div>

    <nav
      v-if="!overviewMode"
      class="floor-rail glass-panel"
      aria-label="楼层选择"
    >
      <button
        v-for="floor in floorOptions"
        :key="floor.id"
        type="button"
        class="floor-button"
        :class="{ active: currentFloor === floor.id }"
        :aria-pressed="currentFloor === floor.id"
        @click="selectFloor(floor.id)"
      >
        {{ floor.label }}
      </button>
    </nav>

    <section
      v-if="!overviewMode"
      class="legend glass-panel"
      :class="{ 'is-collapsed': legendCollapsed }"
    >
      <button
        type="button"
        class="legend__toggle"
        :aria-expanded="!legendCollapsed"
        @click="legendCollapsed = !legendCollapsed"
      >
        <span class="legend__title">图例</span>
        <MapIcon
          :name="legendCollapsed ? 'chevron-down' : 'chevron-up'"
          :size="14"
        />
      </button>

      <div v-if="!legendCollapsed" class="legend__list">
        <div
          v-for="item in legendItems"
          :key="item.className"
          class="legend__row"
        >
          <span
            class="legend__swatch"
            :class="`legend__swatch--${item.className}`"
          ></span>
          <span>{{ item.label }}</span>
        </div>
      </div>
    </section>

    <Transition name="route-bar">
      <section
        v-if="route && !overviewMode"
        class="route-bar glass-panel glass-panel--strong"
        aria-live="polite"
      >
        <div class="route-summary">
          <div class="route-path">
            <span class="endpoint-dot endpoint-dot--start" aria-hidden="true"></span>
            <span class="route-path__name">{{ route.fromName }}</span>
            <MapIcon name="arrow-right" :size="15" class="route-path__arrow" />
            <span class="endpoint-dot endpoint-dot--end" aria-hidden="true"></span>
            <span class="route-path__name">{{ route.toName }}</span>
          </div>
          <div class="route-meta">
            <MapIcon
              :name="
                route.via === 'elevator'
                  ? 'elevator'
                  : route.via === 'stairs'
                    ? 'stairs'
                    : 'arrow-right'
              "
              :size="14"
            />
            <span>途经{{ route.viaLabel }}</span>
            <span class="meta-separator" aria-hidden="true"></span>
            <span class="floor-chip">{{ route.floorLabel }}</span>
          </div>
        </div>

        <span class="route-divider" aria-hidden="true"></span>

        <div class="route-distance">
          <span class="route-distance__value">{{ route.distance }}</span>
          <span class="route-distance__unit">m</span>
        </div>

        <span class="route-divider" aria-hidden="true"></span>

        <div class="route-actions">
          <el-button
            type="primary"
            class="nav-button"
            @click="startNavigation"
          >
            开始导航
          </el-button>
          <el-button class="clear-button" @click="clearRoute">
            清除路线
          </el-button>
        </div>
      </section>
    </Transition>

    <Transition name="hint">
      <div v-if="overviewMode" class="overview-hint glass-panel">
        点击任意教学楼进入楼内视图
      </div>
    </Transition>

    <Transition name="search-drawer">
      <SearchDrawer
        v-if="searchOpen"
        :query="searchQuery"
        :category="activeSearchCategory"
        :suggestions="searchSuggestions"
        :results="searchResults"
        :highlighted-id="highlightedSearchId"
        :current-building-name="searchOriginBuildingName"
        :current-floor="searchOriginFloor"
        @update:query="handleSearchQuery"
        @update:category="handleSearchCategory"
        @suggestion="handleSearchSuggestion"
        @select="selectSearchLocation"
        @clear-filters="clearSearchFilters"
        @close="closeSearch"
      />
    </Transition>

    <Transition name="location-preview">
      <LocationDetailSheet
        v-if="p5Open && selectedLocation"
        :location="selectedLocation"
        @close="closeLocationPreview"
        @navigate-here="handleNavigateHere"
        @start-here="handleStartHere"
      />
    </Transition>

    <NavigationOverlay
      v-if="navigation.active"
      :navigation="navigation"
      @toggle-pause="toggleNavigationPause"
      @end="endNavigationSession"
      @exit-follow="exitNavigationFollow"
      @resume-follow="resumeNavigationFollow"
    />

    <SettingsOverlay
      v-if="settingsOpen"
      @close="settingsOpen = false"
    />
  </main>
</template>

<style scoped lang="less">
.glass-surface() {
  background: rgba(255, 255, 255, 0.82);
  border: 1px solid rgba(255, 255, 255, 0.72);
  box-shadow:
    0 0 0 1px rgba(15, 23, 42, 0.04),
    inset 0 1px 0 rgba(255, 255, 255, 0.85),
    var(--shadow-2);
  -webkit-backdrop-filter: blur(20px) saturate(1.2);
  backdrop-filter: blur(20px) saturate(1.2);
}

.map-page {
  position: fixed;
  inset: 0;
  overflow: hidden;
  background: var(--color-map-3d-bg);
  color: var(--color-neutral-900);
}

/* P3 导航态：只保留地图、路线与 2D/3D 开关，其余 P2 浮层隐藏 */
.map-page.is-navigating {
  .search-panel,
  .overview-switch,
  .legend,
  .route-bar,
  .floor-rail,
  .map-toast {
    display: none;
  }

  .map-toolbar .tool-button,
  .map-toolbar .toolbar-divider {
    display: none;
  }
}

/* P4 搜索态：抽屉替代左侧路线面板，隐藏会与之碰撞的浮层 */
.map-page.is-searching {
  .search-panel,
  .overview-switch,
  .legend,
  .route-bar {
    display: none;
  }
}

/* P5 占位卡打开时，避免与路线结果条重叠 */
.map-page.is-inspecting .route-bar {
  display: none;
}

.glass-panel {
  .glass-surface();

  border-radius: var(--radius-lg);
}

.glass-panel--strong {
  background: rgba(255, 255, 255, 0.94);
  box-shadow:
    0 0 0 1px rgba(15, 23, 42, 0.06),
    var(--shadow-3);
}

/* ---------- 校园总览 / 楼内切换 ---------- */
.overview-switch {
  position: absolute;
  top: 24px;
  left: 50%;
  z-index: 20;
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 3px;
  border-radius: var(--radius-full);
  transform: translateX(-50%);
}

.overview-switch button {
  height: 30px;
  padding: 0 14px;
  border: 0;
  border-radius: var(--radius-full);
  background: transparent;
  color: var(--color-neutral-500);
  font-size: 14px;
  font-weight: 500;
  line-height: 22px;
  white-space: nowrap;
  transition:
    background-color var(--duration-base) var(--ease-standard),
    color var(--duration-base) var(--ease-standard),
    box-shadow var(--duration-base) var(--ease-standard);
}

.overview-switch button:hover {
  color: var(--color-neutral-900);
  background: rgba(255, 255, 255, 0.5);
}

.overview-switch button.active {
  background: var(--color-primary-50);
  box-shadow:
    inset 0 0 0 1px rgba(43, 107, 229, 0.12),
    0 1px 2px rgba(15, 23, 42, 0.04);
  color: var(--color-primary-600);
}

.overview-switch button:focus-visible,
.view-switch button:focus-visible,
.tool-button:focus-visible,
.floor-button:focus-visible,
.nearby__item:focus-visible,
.pin-button:focus-visible,
.legend__toggle:focus-visible {
  outline: 2px solid var(--color-primary-500);
  outline-offset: 2px;
}

/* ---------- 路线搜索面板 ---------- */
.search-panel {
  position: absolute;
  top: 24px;
  left: 24px;
  z-index: 20;
  width: 340px;
  max-height: calc(100dvh - 48px);
  padding: 20px;
  overflow-y: auto;
  border-radius: var(--radius-lg);
  scrollbar-width: thin;
}

.search-panel__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
}

.search-panel__title {
  display: flex;
  align-items: baseline;
  gap: 8px;
  color: var(--color-neutral-900);
  font-size: 20px;
  font-weight: 600;
  line-height: 28px;
}

.search-panel__title-en {
  color: var(--color-neutral-400);
  font-size: 11px;
  font-weight: 500;
  line-height: 18px;
  letter-spacing: 0.14em;
}

.search-panel__actions {
  display: flex;
  align-items: center;
  gap: 4px;
  flex: none;
}

.search-entry {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  height: 28px;
  flex: none;
  padding: 0 8px;
  border: 0;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-primary-600);
  font-size: 12px;
  font-weight: 500;
  line-height: 18px;
  transition:
    background-color var(--duration-fast) var(--ease-standard),
    color var(--duration-fast) var(--ease-standard);
}

.search-entry:hover {
  background: var(--color-primary-50);
}

.search-entry:focus-visible {
  outline: 2px solid var(--color-primary-500);
  outline-offset: 2px;
}

.search-settings {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  flex: none;
  border: 0;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-neutral-500);
  transition:
    background-color var(--duration-fast) var(--ease-standard),
    color var(--duration-fast) var(--ease-standard);
}

.search-settings:hover {
  background: var(--color-primary-50);
  color: var(--color-primary-600);
}

.search-settings:focus-visible {
  outline: 2px solid var(--color-primary-500);
  outline-offset: 2px;
}

.field + .field {
  margin-top: 12px;
}

.search-panel :deep(.el-input) {
  --el-input-height: 44px;
  width: 100%;
}

.search-panel :deep(.el-input__wrapper) {
  min-height: 44px;
  padding: 0 12px;
  border-radius: var(--radius-md);
  background: var(--color-neutral-50);
  box-shadow: 0 0 0 1px var(--color-neutral-200) inset;
  transition:
    box-shadow var(--duration-base) var(--ease-standard),
    background-color var(--duration-base) var(--ease-standard);
}

.search-panel :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px var(--color-neutral-300) inset;
}

.search-panel :deep(.el-input__wrapper.is-focus) {
  background: var(--color-neutral-0);
  box-shadow:
    0 0 0 1px var(--color-primary-500) inset,
    0 0 0 3px rgba(43, 107, 229, 0.16);
}

.search-panel :deep(.el-input__inner) {
  height: 42px;
  color: var(--color-neutral-900);
  font-size: 14px;
  line-height: 22px;
}

.search-panel :deep(.el-input__inner::placeholder) {
  color: var(--color-neutral-500);
}

.search-panel :deep(.el-input__prefix) {
  display: flex;
  align-items: center;
  margin-right: 8px;
}

.search-panel :deep(.el-input__suffix) {
  display: flex;
  align-items: center;
}

.field.is-error :deep(.el-input__wrapper),
.field.is-error :deep(.el-input__wrapper.is-focus) {
  background: var(--color-neutral-0);
  box-shadow:
    0 0 0 1px var(--color-end-500) inset,
    0 0 0 3px rgba(229, 72, 77, 0.15);
}

.input-dot {
  display: block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.input-dot--start {
  background: var(--color-start-500);
}

.input-dot--end {
  background: var(--color-end-500);
}

.pin-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: 0;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-neutral-400);
  transition:
    background-color var(--duration-fast) var(--ease-standard),
    color var(--duration-fast) var(--ease-standard);
}

.pin-button:hover {
  background: var(--color-neutral-100);
  color: var(--color-primary-600);
}

.field-error {
  margin-top: 4px;
  color: var(--color-end-600);
  font-size: 12px;
  font-weight: 400;
  line-height: 18px;
}

.field-error--form {
  margin-top: 8px;
}

.switch-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-top: 16px;
  padding: 10px 12px;
  border-radius: var(--radius-md);
  background: var(--color-neutral-50);
  box-shadow: inset 0 0 0 1px var(--color-neutral-100);
}

.switch-row__copy {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.switch-row__label {
  color: var(--color-neutral-900);
  font-size: 14px;
  font-weight: 500;
  line-height: 22px;
}

.switch-row__hint {
  color: var(--color-neutral-500);
  font-size: 12px;
  font-weight: 400;
  line-height: 18px;
}

.switch-row :deep(.el-switch) {
  --el-switch-on-color: var(--color-primary-500);
  --el-switch-off-color: var(--color-neutral-300);
  --el-switch-button-size: 16px;
}

.switch-row :deep(.el-switch__core) {
  min-width: 36px;
  height: 20px;
  border-radius: var(--radius-full);
}

.switch-row :deep(.el-switch__action) {
  width: 16px;
  height: 16px;
}

.plan-button {
  width: 100%;
  height: 44px;
  margin-top: 20px;
  border-color: var(--color-primary-500);
  border-radius: var(--radius-md);
  background: var(--color-primary-500);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.18),
    0 6px 16px rgba(43, 107, 229, 0.24);
  color: var(--color-neutral-0);
  font-size: 14px;
  font-weight: 600;
  letter-spacing: 0.02em;
  transition:
    background-color var(--duration-fast) var(--ease-standard),
    border-color var(--duration-fast) var(--ease-standard),
    box-shadow var(--duration-fast) var(--ease-standard),
    transform var(--duration-fast) var(--ease-standard);
}

.plan-button:hover {
  background: var(--color-primary-600);
  border-color: var(--color-primary-600);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.16),
    0 8px 20px rgba(43, 107, 229, 0.28);
}

.plan-button:active {
  background: var(--color-primary-700);
  border-color: var(--color-primary-700);
  box-shadow:
    inset 0 1px 2px rgba(15, 23, 42, 0.12),
    0 2px 6px rgba(43, 107, 229, 0.2);
  transform: translateY(1px) scale(0.997);
}

.plan-button:focus-visible {
  outline: 2px solid var(--color-neutral-0);
  outline-offset: 2px;
}

.nearby {
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid var(--color-neutral-100);
}

.nearby__label {
  margin-bottom: 10px;
  color: var(--color-neutral-500);
  font-size: 12px;
  font-weight: 500;
  line-height: 18px;
}

.nearby__grid {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 6px;
}

.nearby__item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  min-width: 0;
  height: 56px;
  padding: 6px 0;
  border: 0;
  border-radius: var(--radius-md);
  background: var(--color-neutral-50);
  box-shadow: inset 0 0 0 1px rgba(15, 23, 42, 0.04);
  color: var(--color-neutral-500);
  font-size: 12px;
  font-weight: 400;
  line-height: 18px;
  white-space: nowrap;
  transition:
    background-color var(--duration-fast) var(--ease-standard),
    color var(--duration-fast) var(--ease-standard),
    box-shadow var(--duration-fast) var(--ease-standard),
    transform var(--duration-fast) var(--ease-standard);
}

.nearby__item:hover {
  background: var(--color-primary-50);
  box-shadow: inset 0 0 0 1px var(--color-primary-100);
  color: var(--color-primary-600);
}

.nearby__item:active {
  transform: translateY(1px);
}

/* ---------- 右上角工具组 ---------- */
.map-toolbar {
  position: absolute;
  top: 24px;
  right: 24px;
  z-index: 20;
  display: flex;
  align-items: center;
  gap: 4px;
  height: 44px;
  padding: 4px;
  border-radius: var(--radius-md);
}

.view-switch {
  display: flex;
  align-items: center;
  gap: 2px;
  height: 36px;
  padding: 2px;
  border-radius: var(--radius-md);
  background: var(--color-neutral-100);
  box-shadow: inset 0 1px 2px rgba(15, 23, 42, 0.04);
}

.view-switch button {
  height: 32px;
  padding: 0 12px;
  border: 0;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-neutral-500);
  font-size: 14px;
  font-weight: 500;
  line-height: 22px;
  transition:
    background-color var(--duration-base) var(--ease-standard),
    color var(--duration-base) var(--ease-standard),
    box-shadow var(--duration-base) var(--ease-standard);
}

.view-switch button:hover:not(.active) {
  color: var(--color-neutral-900);
  background: rgba(255, 255, 255, 0.55);
}

.view-switch button.active {
  background: var(--color-neutral-0);
  box-shadow: var(--shadow-1);
  color: var(--color-primary-600);
}

.toolbar-divider {
  width: 1px;
  height: 20px;
  margin: 0 4px;
  background: var(--color-neutral-200);
}

.tool-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: 0;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-neutral-700);
  transition:
    background-color var(--duration-fast) var(--ease-standard),
    color var(--duration-fast) var(--ease-standard),
    transform var(--duration-fast) var(--ease-standard);
}

.tool-button:hover {
  background: var(--color-primary-50);
  color: var(--color-primary-600);
}

.tool-button:active {
  background: var(--color-neutral-100);
  color: var(--color-primary-600);
  transform: translateY(1px);
}

/* ---------- 楼层栏 ---------- */
.floor-rail {
  position: absolute;
  top: 88px;
  right: 24px;
  z-index: 20;
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 52px;
  max-height: calc(100dvh - 136px);
  padding: 6px;
  overflow-y: auto;
  border-radius: var(--radius-md);
  scrollbar-width: none;
}

.floor-rail::-webkit-scrollbar {
  display: none;
}

.floor-button {
  width: 40px;
  height: 40px;
  border: 0;
  border-radius: var(--radius-md);
  background: transparent;
  color: var(--color-neutral-500);
  font-family: var(--font-family-number);
  font-size: 14px;
  font-weight: 600;
  line-height: 1;
  font-variant-numeric: tabular-nums;
  transition:
    background-color var(--duration-fast) var(--ease-standard),
    color var(--duration-fast) var(--ease-standard),
    box-shadow var(--duration-fast) var(--ease-standard);
}

.floor-button:hover {
  background: var(--color-neutral-50);
  color: var(--color-neutral-900);
}

.floor-button.active {
  background: var(--color-primary-500);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.18),
    0 4px 10px rgba(43, 107, 229, 0.28);
  color: var(--color-neutral-0);
}

/* ---------- 左下角图例 ---------- */
.legend {
  position: absolute;
  bottom: 24px;
  left: 24px;
  z-index: 20;
  width: 148px;
  padding: 12px;
  border-radius: var(--radius-md);
}

.legend.is-collapsed {
  width: auto;
  padding: 0;
}

.legend__toggle {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  width: 100%;
  min-height: 20px;
  padding: 0;
  border: 0;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-neutral-700);
  transition: color var(--duration-fast) var(--ease-standard);
}

.legend__toggle:hover {
  color: var(--color-neutral-900);
}

.legend.is-collapsed .legend__toggle {
  padding: 8px 12px;
}

.legend__title {
  font-size: 12px;
  font-weight: 600;
  line-height: 18px;
}

.legend__list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 10px;
}

.legend__row {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--color-neutral-500);
  font-size: 12px;
  font-weight: 400;
  line-height: 18px;
}

.legend__swatch {
  width: 10px;
  height: 10px;
  flex: none;
  border-radius: 3px;
  box-shadow: inset 0 0 0 1px rgba(15, 23, 42, 0.06);
}

.legend__swatch--classroom {
  background: var(--color-neutral-500);
}

.legend__swatch--stairs {
  background: var(--color-accent-500);
}

.legend__swatch--elevator {
  background: var(--color-primary-500);
}

.legend__swatch--restroom {
  background: var(--color-warning-500);
}

/* ---------- 底部路线结果条 ---------- */
.route-bar {
  position: absolute;
  bottom: 24px;
  left: 50%;
  z-index: 30;
  display: flex;
  align-items: center;
  gap: 16px;
  width: min(760px, calc(100vw - 48px));
  padding: 16px 20px;
  border-radius: var(--radius-lg);
  transform: translateX(-50%);
}

.route-summary {
  flex: 1;
  min-width: 0;
}

.route-path {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  color: var(--color-neutral-900);
  font-size: 14px;
  font-weight: 600;
  line-height: 22px;
}

.route-path__name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.route-path__arrow {
  color: var(--color-neutral-400);
}

.endpoint-dot {
  width: 8px;
  height: 8px;
  flex: none;
  border-radius: 50%;
}

.endpoint-dot--start {
  background: var(--color-start-500);
}

.endpoint-dot--end {
  background: var(--color-end-500);
}

.route-meta {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 6px;
  color: var(--color-neutral-500);
  font-size: 12px;
  font-weight: 400;
  line-height: 18px;
}

.meta-separator {
  width: 1px;
  height: 12px;
  margin: 0 2px;
  background: var(--color-neutral-200);
}

.floor-chip {
  padding: 1px 6px;
  border-radius: var(--radius-xs);
  background: var(--color-primary-50);
  color: var(--color-primary-700);
  font-size: 12px;
  font-weight: 500;
  line-height: 18px;
}

.route-divider {
  width: 1px;
  height: 36px;
  flex: none;
  background: var(--color-neutral-200);
}

.route-distance {
  display: flex;
  align-items: baseline;
  gap: 2px;
  flex: none;
}

.route-distance__value {
  color: var(--color-neutral-900);
  font-family: var(--font-family-number);
  font-size: 20px;
  font-weight: 600;
  line-height: 28px;
  font-variant-numeric: tabular-nums;
}

.route-distance__unit {
  color: var(--color-neutral-500);
  font-size: 12px;
  font-weight: 500;
  line-height: 18px;
}

.route-actions {
  display: flex;
  gap: 8px;
  flex: none;
}

.nav-button {
  height: 40px;
  padding: 0 16px;
  border-radius: var(--radius-sm);
  font-size: 14px;
  font-weight: 500;
}

.nav-button:hover {
  background: var(--color-primary-600);
  border-color: var(--color-primary-600);
}

.nav-button:active {
  background: var(--color-primary-700);
  border-color: var(--color-primary-700);
}

.clear-button {
  height: 40px;
  padding: 0 14px;
  border: 1px solid var(--color-neutral-200);
  border-radius: var(--radius-sm);
  background: var(--color-neutral-0);
  color: var(--color-neutral-700);
  font-size: 14px;
  font-weight: 500;
}

.clear-button:hover {
  border-color: var(--color-neutral-300);
  background: var(--color-neutral-50);
  color: var(--color-neutral-900);
}

/* ---------- 总览提示 / Toast ---------- */
.overview-hint {
  position: absolute;
  bottom: 24px;
  left: 50%;
  z-index: 20;
  padding: 10px 16px;
  border-radius: var(--radius-md);
  color: var(--color-neutral-700);
  font-size: 14px;
  font-weight: 400;
  line-height: 22px;
  white-space: nowrap;
  transform: translateX(-50%);
}

.map-toast {
  position: absolute;
  top: 76px;
  left: 50%;
  z-index: 40;
  max-width: calc(100vw - 32px);
  padding: 10px 16px;
  border-radius: var(--radius-md);
  color: var(--color-neutral-900);
  font-size: 14px;
  font-weight: 400;
  line-height: 22px;
  text-align: center;
  transform: translateX(-50%);
}

/* ---------- 过渡 ---------- */
.route-bar-enter-active,
.route-bar-leave-active {
  transition:
    opacity 0.24s var(--ease-standard),
    transform 0.24s var(--ease-standard);
}

.route-bar-enter-from,
.route-bar-leave-to {
  opacity: 0;
  transform: translate(-50%, 16px);
}

.toast-enter-active,
.toast-leave-active,
.hint-enter-active,
.hint-leave-active {
  transition:
    opacity 0.18s var(--ease-standard),
    transform 0.18s var(--ease-standard);
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translate(-50%, -6px);
}

.hint-enter-from,
.hint-leave-to {
  opacity: 0;
  transform: translate(-50%, 6px);
}

/* P4 搜索抽屉与 P5 占位卡过渡 */
.search-drawer-enter-active,
.search-drawer-leave-active,
.location-preview-enter-active,
.location-preview-leave-active {
  transition:
    opacity 0.24s var(--ease-standard),
    transform 0.24s var(--ease-standard);
}

.search-drawer-enter-from,
.search-drawer-leave-to {
  opacity: 0;
  transform: translateX(-16px);
}

.location-preview-enter-from,
.location-preview-leave-to {
  opacity: 0;
}

@media (max-width: 767px) {
  .search-drawer-enter-from,
  .search-drawer-leave-to {
    transform: translateY(24px);
  }
}

/* ---------- 平板 ---------- */
@media (min-width: 768px) and (max-width: 1279px) {
  .search-panel {
    width: 320px;
  }

  .legend {
    width: 140px;
  }
}

@media (max-width: 1023px) {
  .overview-switch button {
    padding: 0 10px;
    font-size: 12px;
  }

  .map-toolbar {
    top: 16px;
    right: 16px;
  }

  .floor-rail {
    top: 76px;
    right: 16px;
  }

  .search-panel {
    top: 16px;
    left: 16px;
    width: 300px;
    padding: 16px;
    max-height: calc(100dvh - 32px);
  }

  .legend {
    bottom: 16px;
    left: 16px;
  }

  .route-bar {
    bottom: 16px;
    width: calc(100vw - 32px);
    gap: 12px;
  }
}

/* ---------- 移动端：面板靠上，楼层栏留出右侧通道 ---------- */
@media (max-width: 767px) {
  .overview-switch {
    top: 16px;
    left: 16px;
    transform: none;
  }

  .map-toolbar {
    top: 16px;
    right: 16px;
  }

  .map-toolbar .tool-button:nth-of-type(1),
  .map-toolbar .tool-button:nth-of-type(2) {
    display: none;
  }

  .search-panel {
    top: 76px;
    right: 76px;
    left: 16px;
    width: auto;
    max-height: 52dvh;
    padding: 16px;
    border-radius: var(--radius-lg);
  }

  .search-panel__head {
    margin-bottom: 12px;
  }

  .floor-rail {
    top: 76px;
    right: 16px;
    max-height: calc(100dvh - 260px);
  }

  .legend {
    display: none;
  }

  .route-bar {
    bottom: 16px;
    left: 16px;
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 12px;
    width: calc(100vw - 32px);
    padding: 14px 16px;
    transform: none;
  }

  .route-divider {
    display: none;
  }

  .route-summary {
    grid-column: 1 / -1;
  }

  .route-distance {
    align-self: center;
  }

  .route-actions {
    justify-self: end;
  }

  .route-bar-enter-from,
  .route-bar-leave-to {
    opacity: 0;
    transform: translateY(16px);
  }

  .overview-hint {
    bottom: 16px;
    max-width: calc(100vw - 32px);
    font-size: 12px;
    line-height: 18px;
    white-space: normal;
    text-align: center;
  }

  .map-toast {
    top: 68px;
  }
}
</style>
