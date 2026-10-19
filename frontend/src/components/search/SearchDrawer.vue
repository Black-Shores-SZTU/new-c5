<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import MapIcon from '../MapIcon.vue'

const props = defineProps({
  query: {
    type: String,
    default: ''
  },
  category: {
    type: String,
    default: 'all'
  },
  suggestions: {
    type: Array,
    default: () => []
  },
  results: {
    type: Array,
    default: () => []
  },
  highlightedId: {
    type: String,
    default: ''
  },
  currentBuildingName: {
    type: String,
    default: ''
  },
  currentFloor: {
    type: String,
    default: ''
  }
})

const emit = defineEmits([
  'update:query',
  'update:category',
  'suggestion',
  'select',
  'clear-filters',
  'close'
])

const categories = [
  { key: 'all', label: '全部' },
  { key: 'classroom', label: '教室' },
  { key: 'restroom', label: '卫生间' },
  { key: 'elevator', label: '电梯' },
  { key: 'stairs', label: '楼梯' },
  { key: 'exit', label: '安全出口' }
]

const typeMeta = {
  classroom: { icon: 'classroom', className: 'classroom' },
  teacher_rest: { icon: 'classroom', className: 'teacher' },
  restroom: { icon: 'toilet', className: 'restroom' },
  elevator: { icon: 'elevator', className: 'elevator' },
  stairs: { icon: 'stairs', className: 'stairs' },
  exit: { icon: 'exit', className: 'exit' },
  water: { icon: 'water', className: 'water' },
  auditorium: { icon: 'palette', className: 'auditorium' },
  entrance: { icon: 'exit', className: 'entrance' }
}

const getTypeMeta = (type) => typeMeta[type] || typeMeta.classroom

const searchInputRef = ref(null)
const suggestionsOpen = ref(false)
let focusTimer = null

const showSuggestions = computed(
  () =>
    suggestionsOpen.value &&
    Boolean(props.query.trim()) &&
    props.suggestions.length > 0
)

const handleInput = (value) => {
  emit('update:query', value)
  suggestionsOpen.value = Boolean(String(value).trim())
}

const handleFocus = () => {
  if (props.query.trim()) {
    suggestionsOpen.value = true
  }
}

const handleBlur = () => {
  window.setTimeout(() => {
    suggestionsOpen.value = false
  }, 120)
}

const handleClear = () => {
  suggestionsOpen.value = false
  emit('update:query', '')
}

const handleSuggestion = (item) => {
  suggestionsOpen.value = false
  emit('suggestion', item)
}

const handleCategory = (key) => {
  suggestionsOpen.value = false
  emit('update:category', key)
}

const handleEnter = () => {
  if (!props.results.length) {
    return
  }
  suggestionsOpen.value = false
  emit('select', props.results[0])
}

const handleEscape = () => {
  suggestionsOpen.value = false
}

watch(
  () => props.query,
  (value) => {
    if (!value) {
      suggestionsOpen.value = false
    }
  }
)

onMounted(() => {
  // 等抽屉入场与 Element Plus 焦点事件绑定完成后再聚焦，确保焦点环出现
  focusTimer = window.setTimeout(() => {
    searchInputRef.value?.focus()
    focusTimer = null
  }, 80)
})

onBeforeUnmount(() => {
  if (focusTimer) {
    clearTimeout(focusTimer)
  }
})
</script>

<template>
  <aside class="search-drawer" aria-label="搜索与分类">
    <div class="search-drawer__handle" aria-hidden="true"></div>

    <header class="search-drawer__head">
      <div>
        <p class="search-drawer__eyebrow">SEARCH / CATEGORY</p>
        <h2 class="search-drawer__title">查找地点</h2>
      </div>
      <div class="search-drawer__head-actions">
        <span class="search-drawer__context">
          {{ currentBuildingName }} · {{ currentFloor }}
        </span>
        <button
          type="button"
          class="search-drawer__close"
          aria-label="关闭搜索"
          @click="emit('close')"
        >
          <MapIcon name="close" :size="18" />
        </button>
      </div>
    </header>

    <div class="search-drawer__search">
      <el-input
        ref="searchInputRef"
        :model-value="query"
        clearable
        autocomplete="off"
        placeholder="搜索教室号 / 名称，如 A101"
        aria-label="搜索教室号或名称"
        @update:model-value="handleInput"
        @focus="handleFocus"
        @blur="handleBlur"
        @clear="handleClear"
        @keyup.enter.prevent="handleEnter"
        @keyup.esc="handleEscape"
      >
        <template #prefix>
          <MapIcon name="search" :size="18" />
        </template>
      </el-input>

      <Transition name="suggest">
        <div v-if="showSuggestions" class="suggestions">
          <button
            v-for="item in suggestions"
            :key="item.id"
            type="button"
            class="suggestion-row"
            @mousedown.prevent
            @click="handleSuggestion(item)"
          >
            <span
              class="type-icon type-icon--small"
              :class="`type-icon--${getTypeMeta(item.type).className}`"
            >
              <MapIcon :name="getTypeMeta(item.type).icon" :size="16" />
            </span>
            <span class="suggestion-row__body">
              <span class="suggestion-row__name">{{ item.name }}</span>
              <span class="suggestion-row__meta">
                {{ item.buildingName }} · {{ item.floor }}
              </span>
            </span>
          </button>
        </div>
      </Transition>
    </div>

    <div class="category-row" role="group" aria-label="分类筛选">
      <button
        v-for="item in categories"
        :key="item.key"
        type="button"
        class="category-chip"
        :class="{ active: category === item.key }"
        :aria-pressed="category === item.key"
        @click="handleCategory(item.key)"
      >
        {{ item.label }}
      </button>
    </div>

    <div class="search-drawer__divider" aria-hidden="true"></div>

    <div class="search-drawer__result-head">
      <span>搜索结果</span>
      <span>共 {{ results.length }} 个地点</span>
    </div>

    <div class="search-drawer__results">
      <template v-if="results.length">
        <button
          v-for="item in results"
          :key="item.id"
          type="button"
          class="result-row"
          :class="{ 'is-highlighted': item.id === highlightedId }"
          @click="emit('select', item)"
        >
          <span
            class="type-icon"
            :class="`type-icon--${getTypeMeta(item.type).className}`"
          >
            <MapIcon :name="getTypeMeta(item.type).icon" :size="18" />
          </span>

          <span class="result-row__body">
            <span class="result-row__name">{{ item.name }}</span>
            <span class="result-row__meta">
              {{ item.buildingName }} · {{ item.floor }}
            </span>
          </span>

          <span class="result-row__distance">
            {{ item.distance }}
            <span class="result-row__unit">m</span>
          </span>
        </button>
      </template>

      <div v-else class="empty-state">
        <span class="empty-state__icon">
          <MapIcon name="search" :size="28" />
        </span>
        <p class="empty-state__title">没有找到相关地点</p>
        <p class="empty-state__description">换个教室号或分类试试</p>
        <button
          type="button"
          class="empty-state__action"
          @click="emit('clear-filters')"
        >
          清除筛选
        </button>
      </div>
    </div>
  </aside>
</template>

<style scoped lang="less">
.search-drawer {
  position: absolute;
  top: 24px;
  bottom: 24px;
  left: 24px;
  z-index: 45;
  display: flex;
  flex-direction: column;
  width: 420px;
  padding: 20px;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.88);
  box-shadow:
    0 0 0 1px rgba(15, 23, 42, 0.04),
    inset 0 1px 0 rgba(255, 255, 255, 0.85),
    var(--shadow-2);
  color: var(--color-neutral-900);
  -webkit-backdrop-filter: blur(20px) saturate(1.2);
  backdrop-filter: blur(20px) saturate(1.2);
}

.search-drawer__handle {
  display: none;
}

/* ---------- 头部 ---------- */
.search-drawer__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.search-drawer__eyebrow {
  color: var(--color-neutral-500);
  font-size: 12px;
  font-weight: 500;
  line-height: 18px;
  letter-spacing: 0.12em;
}

.search-drawer__title {
  margin-top: 2px;
  color: var(--color-neutral-900);
  font-size: 20px;
  font-weight: 600;
  line-height: 28px;
}

.search-drawer__head-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: none;
}

.search-drawer__context {
  padding: 1px 6px;
  border-radius: var(--radius-xs);
  background: var(--color-neutral-100);
  color: var(--color-neutral-500);
  font-size: 12px;
  font-weight: 500;
  line-height: 18px;
  white-space: nowrap;
}

.search-drawer__close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: 0;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-neutral-500);
  transition:
    background-color var(--duration-fast) var(--ease-standard),
    color var(--duration-fast) var(--ease-standard);
}

.search-drawer__close:hover {
  background: var(--color-neutral-100);
  color: var(--color-neutral-900);
}

.search-drawer__close:focus-visible,
.category-chip:focus-visible,
.suggestion-row:focus-visible,
.result-row:focus-visible,
.empty-state__action:focus-visible {
  outline: 2px solid var(--color-primary-500);
  outline-offset: 2px;
}

/* ---------- 搜索框 ---------- */
.search-drawer__search {
  position: relative;
  z-index: 2;
  margin-top: 16px;
}

.search-drawer :deep(.el-input) {
  --el-input-height: 44px;
  width: 100%;
}

.search-drawer :deep(.el-input__wrapper) {
  min-height: 44px;
  padding: 0 12px;
  border-radius: var(--radius-md);
  background: var(--color-neutral-50);
  box-shadow: 0 0 0 1px var(--color-neutral-200) inset;
  transition:
    box-shadow var(--duration-base) var(--ease-standard),
    background-color var(--duration-base) var(--ease-standard);
}

.search-drawer :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px var(--color-neutral-300) inset;
}

.search-drawer :deep(.el-input__wrapper.is-focus) {
  background: var(--color-neutral-0);
  box-shadow:
    0 0 0 1px var(--color-primary-500) inset,
    0 0 0 3px rgba(43, 107, 229, 0.16);
}

.search-drawer :deep(.el-input__inner) {
  height: 42px;
  color: var(--color-neutral-900);
  font-size: 14px;
  line-height: 22px;
}

.search-drawer :deep(.el-input__inner::placeholder) {
  color: var(--color-neutral-500);
}

.search-drawer :deep(.el-input__prefix) {
  display: flex;
  align-items: center;
  margin-right: 8px;
  color: var(--color-neutral-500);
}

/* ---------- 实时建议 ---------- */
.suggestions {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  left: 0;
  z-index: 5;
  max-height: 320px;
  padding: 4px;
  overflow-y: auto;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.94);
  box-shadow:
    0 0 0 1px rgba(15, 23, 42, 0.04),
    inset 0 1px 0 rgba(255, 255, 255, 0.85),
    var(--shadow-2);
  -webkit-backdrop-filter: blur(20px) saturate(1.2);
  backdrop-filter: blur(20px) saturate(1.2);
}

.suggestion-row {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  min-height: 44px;
  padding: 6px 8px;
  border: 0;
  border-radius: var(--radius-sm);
  background: transparent;
  text-align: left;
  transition: background-color var(--duration-fast) var(--ease-standard);
}

.suggestion-row:hover {
  background: var(--color-primary-50);
}

.suggestion-row__body {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.suggestion-row__name {
  overflow: hidden;
  color: var(--color-neutral-900);
  font-size: 14px;
  font-weight: 500;
  line-height: 20px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.suggestion-row__meta {
  color: var(--color-neutral-500);
  font-size: 12px;
  font-weight: 400;
  line-height: 18px;
}

/* ---------- 分类 ---------- */
.category-row {
  display: flex;
  gap: 8px;
  margin-top: 16px;
  padding-bottom: 2px;
  overflow-x: auto;
  scrollbar-width: none;
}

.category-row::-webkit-scrollbar {
  display: none;
}

.category-chip {
  height: 32px;
  flex: none;
  padding: 0 12px;
  border: 1px solid transparent;
  border-radius: var(--radius-full);
  background: var(--color-neutral-50);
  color: var(--color-neutral-500);
  font-size: 12px;
  font-weight: 500;
  line-height: 18px;
  white-space: nowrap;
  transition:
    background-color var(--duration-fast) var(--ease-standard),
    border-color var(--duration-fast) var(--ease-standard),
    color var(--duration-fast) var(--ease-standard);
}

.category-chip:hover {
  background: var(--color-neutral-100);
  color: var(--color-neutral-700);
}

.category-chip.active {
  border-color: var(--color-primary-100);
  background: var(--color-primary-50);
  box-shadow: inset 0 0 0 1px var(--color-primary-100);
  color: var(--color-primary-600);
}

.search-drawer__divider {
  height: 1px;
  margin: 16px 0 12px;
  flex: none;
  background: var(--color-neutral-100);
}

.search-drawer__result-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 8px;
  flex: none;
  color: var(--color-neutral-500);
  font-size: 12px;
  font-weight: 400;
  line-height: 18px;
}

.search-drawer__result-head span:first-child {
  color: var(--color-neutral-700);
  font-weight: 600;
}

/* ---------- 结果列表 ---------- */
.search-drawer__results {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 8px;
  min-height: 0;
  padding-right: 2px;
  padding-bottom: 4px;
  overflow-y: auto;
  scrollbar-width: thin;
}

.result-row {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  min-height: 68px;
  padding: 10px 12px;
  flex: none;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  background: var(--color-neutral-50);
  text-align: left;
  transition:
    background-color var(--duration-fast) var(--ease-standard),
    border-color var(--duration-fast) var(--ease-standard),
    box-shadow var(--duration-fast) var(--ease-standard);
}

.result-row:hover {
  border-color: var(--color-primary-100);
  background: var(--color-primary-50);
  box-shadow: var(--shadow-1);
}

.result-row.is-highlighted {
  border-color: var(--color-primary-500);
  background: var(--color-primary-50);
  box-shadow:
    0 0 0 1px var(--color-primary-500),
    var(--shadow-1);
}

.type-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  flex: none;
  border-radius: var(--radius-md);
}

.type-icon--small {
  width: 28px;
  height: 28px;
  border-radius: var(--radius-sm);
}

.type-icon--classroom {
  background: var(--color-neutral-100);
  color: var(--color-neutral-700);
}

.type-icon--stairs,
.type-icon--water {
  background: var(--color-accent-50);
  color: var(--color-accent-600);
}

.type-icon--elevator {
  background: var(--color-primary-50);
  color: var(--color-primary-600);
}

.type-icon--restroom,
.type-icon--exit,
.type-icon--teacher {
  background: var(--color-warning-50);
  color: var(--color-warning-600);
}

.type-icon--auditorium {
  background: #F1E9FF;
  color: #7C5CFC;
}

.type-icon--entrance {
  background: var(--color-primary-50);
  color: var(--color-primary-600);
}

.result-row__body {
  display: flex;
  min-width: 0;
  flex: 1;
  flex-direction: column;
}

.result-row__name {
  overflow: hidden;
  color: var(--color-neutral-900);
  font-size: 14px;
  font-weight: 600;
  line-height: 22px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.result-row__meta {
  color: var(--color-neutral-500);
  font-size: 12px;
  font-weight: 400;
  line-height: 18px;
}

.result-row__distance {
  flex: none;
  color: var(--color-neutral-900);
  font-family: var(--font-family-number);
  font-size: 14px;
  font-weight: 600;
  line-height: 22px;
  font-variant-numeric: tabular-nums;
}

.result-row__unit {
  color: var(--color-neutral-500);
  font-size: 12px;
  font-weight: 500;
  line-height: 18px;
}

/* ---------- 空状态 ---------- */
.empty-state {
  display: flex;
  flex: 1;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 280px;
  padding: 32px 16px;
  text-align: center;
}

.empty-state__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 64px;
  height: 64px;
  margin-bottom: 12px;
  border-radius: 50%;
  background: var(--color-neutral-100);
  color: var(--color-neutral-400);
}

.empty-state__title {
  color: var(--color-neutral-900);
  font-size: 14px;
  font-weight: 600;
  line-height: 22px;
}

.empty-state__description {
  margin-top: 4px;
  color: var(--color-neutral-500);
  font-size: 12px;
  font-weight: 400;
  line-height: 18px;
}

.empty-state__action {
  height: 32px;
  margin-top: 16px;
  padding: 0 12px;
  border: 0;
  border-radius: var(--radius-sm);
  background: var(--color-primary-50);
  color: var(--color-primary-600);
  font-size: 12px;
  font-weight: 500;
  line-height: 18px;
  transition:
    background-color var(--duration-fast) var(--ease-standard),
    transform var(--duration-fast) var(--ease-standard);
}

.empty-state__action:hover {
  background: var(--color-primary-100);
}

.empty-state__action:active {
  transform: translateY(1px);
}

/* ---------- 建议过渡 ---------- */
.suggest-enter-active,
.suggest-leave-active {
  transition:
    opacity 0.15s var(--ease-standard),
    transform 0.15s var(--ease-standard);
}

.suggest-enter-from,
.suggest-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

/* ---------- 响应式 ---------- */
@media (min-width: 768px) and (max-width: 1279px) {
  .search-drawer {
    top: 16px;
    bottom: 16px;
    left: 16px;
    width: 380px;
    padding: 16px;
  }
}

@media (max-width: 767px) {
  .search-drawer {
    top: auto;
    right: 0;
    bottom: 0;
    left: 0;
    width: auto;
    height: 82dvh;
    padding: 16px;
    padding-bottom: calc(16px + env(safe-area-inset-bottom));
    border-radius: var(--radius-lg) var(--radius-lg) 0 0;
  }

  .search-drawer__handle {
    display: block;
    width: 36px;
    height: 4px;
    margin: 0 auto 12px;
    flex: none;
    border-radius: var(--radius-full);
    background: var(--color-neutral-300);
  }

  .search-drawer__context {
    display: none;
  }

  .suggestions {
    max-height: 280px;
  }

  .category-chip {
    height: 30px;
    padding: 0 10px;
  }

  .result-row {
    min-height: 64px;
    padding: 8px 10px;
  }

  .empty-state {
    min-height: 220px;
  }
}
</style>
