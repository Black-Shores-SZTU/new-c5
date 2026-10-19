<script setup>
import { nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import MapIcon from '../MapIcon.vue'
import { floors } from '../../mock/indoorData'

const emit = defineEmits(['close'])

const STORAGE_KEY = 'campus-nav-settings-v1'
const floorOptions = floors.map((floor) => floor.id)

const preferences = reactive({
  preferElevator: false,
  avoidStairs: false,
  defaultFloor: 'F1'
})

const aboutOpen = ref(false)
const pageRef = ref(null)
const toastMessage = ref('')
let toastTimer = null

const showToast = (message) => {
  toastMessage.value = message
  if (toastTimer) {
    clearTimeout(toastTimer)
  }
  toastTimer = window.setTimeout(() => {
    toastMessage.value = ''
    toastTimer = null
  }, 2200)
}

const loadSettings = () => {
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY)
    if (!raw) {
      return
    }
    const data = JSON.parse(raw)
    preferences.preferElevator = Boolean(data.preferElevator)
    preferences.avoidStairs = Boolean(data.avoidStairs)
    if (floorOptions.includes(data.defaultFloor)) {
      preferences.defaultFloor = data.defaultFloor
    }
  } catch {
    // 本地存储不可用时使用默认值
  }
}

const saveSettings = () => {
  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(preferences))
  } catch {
    // 忽略本地存储异常，仅影响偏好持久化
  }
}

watch(preferences, saveSettings, { deep: true })

const handleClearCache = () => {
  try {
    window.localStorage.removeItem('campus-nav-map-cache')
    window.localStorage.removeItem('campus-nav-search-cache')
  } catch {
    // 演示环境无真实缓存，忽略异常
  }
  showToast('缓存已清除（本地演示）')
}

const handleThemePlaceholder = () => {
  showToast('主题切换将在后续版本开放')
}

const handleHistoryPlaceholder = () => {
  showToast('历史路线将在后续版本开放')
}

const handleEscape = () => {
  if (aboutOpen.value) {
    aboutOpen.value = false
    return
  }
  emit('close')
}

onMounted(() => {
  loadSettings()
  nextTick(() => {
    pageRef.value?.focus()
  })
})

onBeforeUnmount(() => {
  if (toastTimer) {
    clearTimeout(toastTimer)
  }
})
</script>

<template>
  <main
    ref="pageRef"
    class="settings-page"
    tabindex="-1"
    aria-label="设置"
    @keydown.esc.stop="handleEscape"
  >
    <header class="settings-header">
      <div class="settings-header__inner">
        <button
          type="button"
          class="settings-back"
          aria-label="返回地图"
          @click="emit('close')"
        >
          <MapIcon name="chevron-left" :size="20" />
        </button>
        <div class="settings-title-group">
          <h1 class="settings-title">设置</h1>
          <span class="settings-eyebrow">SETTINGS</span>
        </div>
      </div>
    </header>

    <div class="settings-content">
      <!-- 出行偏好 -->
      <section class="settings-group">
        <h2 class="settings-group__title">出行偏好</h2>
        <div class="settings-card">
          <div class="settings-row">
            <span class="settings-row__icon">
              <MapIcon name="elevator" :size="18" />
            </span>
            <div class="settings-row__body">
              <span class="settings-row__title">优先电梯</span>
              <span class="settings-row__desc">规划路线时优先选择电梯</span>
            </div>
            <div class="settings-row__control">
              <el-switch
                v-model="preferences.preferElevator"
                class="settings-switch"
                :width="36"
                aria-label="优先电梯"
              />
            </div>
          </div>

          <div class="settings-row">
            <span class="settings-row__icon">
              <MapIcon name="stairs" :size="18" />
            </span>
            <div class="settings-row__body">
              <span class="settings-row__title">少爬楼</span>
              <span class="settings-row__desc">尽量避开楼梯</span>
            </div>
            <div class="settings-row__control">
              <el-switch
                v-model="preferences.avoidStairs"
                class="settings-switch"
                :width="36"
                aria-label="少爬楼"
              />
            </div>
          </div>
        </div>
      </section>

      <!-- 默认设置 -->
      <section class="settings-group">
        <h2 class="settings-group__title">默认设置</h2>
        <div class="settings-card">
          <div class="settings-row settings-row--floor">
            <span class="settings-row__icon">
              <MapIcon name="building" :size="18" />
            </span>
            <div class="settings-row__body">
              <span class="settings-row__title">默认楼层</span>
              <span class="settings-row__desc">进入地图时默认显示的楼层</span>
            </div>
            <div class="settings-row__control">
              <div class="floor-select" role="group" aria-label="默认楼层">
                <button
                  v-for="floor in floorOptions"
                  :key="floor"
                  type="button"
                  class="floor-select__item"
                  :class="{ active: preferences.defaultFloor === floor }"
                  :aria-pressed="preferences.defaultFloor === floor"
                  @click="preferences.defaultFloor = floor"
                >
                  {{ floor }}
                </button>
              </div>
            </div>
          </div>

          <button
            type="button"
            class="settings-row settings-row--button settings-row--placeholder"
            @click="handleThemePlaceholder"
          >
            <span class="settings-row__icon">
              <MapIcon name="palette" :size="18" />
            </span>
            <div class="settings-row__body">
              <span class="settings-row__title">主题 / 色系</span>
              <span class="settings-row__desc">当前为深色 · 科技蓝</span>
            </div>
            <div class="settings-row__control">
              <span class="settings-tag">即将开放</span>
            </div>
          </button>
        </div>
      </section>

      <!-- 数据 -->
      <section class="settings-group">
        <h2 class="settings-group__title">数据</h2>
        <div class="settings-card">
          <button
            type="button"
            class="settings-row settings-row--button settings-row--placeholder"
            @click="handleHistoryPlaceholder"
          >
            <span class="settings-row__icon">
              <MapIcon name="history" :size="18" />
            </span>
            <div class="settings-row__body">
              <span class="settings-row__title">历史路线</span>
              <span class="settings-row__desc">暂无历史记录</span>
            </div>
            <div class="settings-row__control">
              <span class="settings-tag">即将开放</span>
            </div>
          </button>

          <div class="settings-row">
            <span class="settings-row__icon">
              <MapIcon name="trash" :size="18" />
            </span>
            <div class="settings-row__body">
              <span class="settings-row__title">清除缓存</span>
              <span class="settings-row__desc">清除本地地图与搜索缓存</span>
            </div>
            <div class="settings-row__control">
              <button
                type="button"
                class="settings-clear"
                @click="handleClearCache"
              >
                清除
              </button>
            </div>
          </div>
        </div>
      </section>

      <!-- 帮助 -->
      <section class="settings-group">
        <h2 class="settings-group__title">帮助</h2>
        <div class="settings-card">
          <button
            type="button"
            class="settings-row settings-row--button"
            @click="aboutOpen = true"
          >
            <span class="settings-row__icon">
              <MapIcon name="info" :size="18" />
            </span>
            <div class="settings-row__body">
              <span class="settings-row__title">帮助与关于</span>
              <span class="settings-row__desc">使用说明、版本与关于</span>
            </div>
            <div class="settings-row__control settings-row__control--link">
              <span class="settings-version">v0.1.0</span>
              <MapIcon name="chevron-right" :size="16" />
            </div>
          </button>
        </div>
      </section>

      <p class="settings-footnote">设置会自动保存在本机</p>
    </div>

    <Transition name="settings-toast">
      <div v-if="toastMessage" class="settings-toast" role="status">
        {{ toastMessage }}
      </div>
    </Transition>

    <Transition name="about">
      <div v-if="aboutOpen" class="about-overlay" @click.self="aboutOpen = false">
        <div
          class="about-card"
          role="dialog"
          aria-modal="true"
          aria-labelledby="settings-about-title"
          @click.stop
        >
          <p class="about-eyebrow">关于</p>
          <h2 id="settings-about-title" class="about-title">校园室内导航</h2>
          <p class="about-subtitle">3D视角 · 教学楼内快速找路</p>
          <p class="about-note">仅限已建模教学楼</p>
          <p class="about-version">v0.1.0</p>
          <button
            type="button"
            class="about-close"
            @click="aboutOpen = false"
          >
            知道了
          </button>
        </div>
      </div>
    </Transition>
  </main>
</template>

<style scoped lang="less">
.settings-page {
  position: absolute;
  inset: 0;
  z-index: 70;
  overflow-y: auto;
  background: var(--color-map-3d-bg);
  color: var(--color-neutral-0);
}

.settings-page:focus {
  outline: none;
}

/* ---------- 顶部标题栏 ---------- */
.settings-header {
  position: sticky;
  top: 0;
  z-index: 3;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(16, 27, 44, 0.86);
  -webkit-backdrop-filter: blur(20px) saturate(1.2);
  backdrop-filter: blur(20px) saturate(1.2);
}

.settings-header__inner {
  display: flex;
  align-items: center;
  gap: 12px;
  max-width: 720px;
  height: 64px;
  margin: 0 auto;
  padding: 0 24px;
}

.settings-back {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  flex: none;
  border: 0;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-neutral-300);
  transition:
    background-color var(--duration-fast) var(--ease-standard),
    color var(--duration-fast) var(--ease-standard);
}

.settings-back:hover {
  background: rgba(255, 255, 255, 0.06);
  color: var(--color-neutral-0);
}

.settings-back:focus-visible,
.floor-select__item:focus-visible,
.settings-clear:focus-visible,
.about-close:focus-visible {
  outline: 2px solid var(--color-primary-500);
  outline-offset: 2px;
}

.settings-title-group {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.settings-title {
  color: var(--color-neutral-0);
  font-size: 24px;
  font-weight: 600;
  line-height: 32px;
}

.settings-eyebrow {
  color: var(--color-neutral-400);
  font-size: 12px;
  font-weight: 500;
  line-height: 18px;
  letter-spacing: 0.12em;
}

/* ---------- 内容 ---------- */
.settings-content {
  max-width: 720px;
  margin: 0 auto;
  padding: 24px 24px 48px;
}

.settings-group + .settings-group {
  margin-top: 28px;
}

.settings-group__title {
  margin: 0 0 8px 4px;
  color: var(--color-neutral-400);
  font-size: 12px;
  font-weight: 500;
  line-height: 18px;
  letter-spacing: 0.08em;
}

.settings-card {
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.055);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.06),
    var(--shadow-2);
}

.settings-row {
  position: relative;
  display: flex;
  align-items: center;
  gap: 12px;
  min-height: 68px;
  padding: 0 16px;
  transition: background-color var(--duration-fast) var(--ease-standard);
}

.settings-row + .settings-row::before {
  position: absolute;
  top: 0;
  right: 16px;
  left: 64px;
  height: 1px;
  background: rgba(255, 255, 255, 0.06);
  content: '';
}

.settings-row--button {
  width: 100%;
  border: 0;
  background: transparent;
  color: inherit;
  text-align: left;
  font: inherit;
}

.settings-row--button:hover {
  background: rgba(255, 255, 255, 0.03);
}

.settings-row__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  flex: none;
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.06);
  color: var(--color-neutral-300);
}

.settings-row__body {
  display: flex;
  min-width: 0;
  flex: 1;
  flex-direction: column;
  gap: 2px;
}

.settings-row__title {
  color: var(--color-neutral-0);
  font-size: 14px;
  font-weight: 500;
  line-height: 22px;
}

.settings-row__desc {
  color: var(--color-neutral-400);
  font-size: 12px;
  font-weight: 400;
  line-height: 18px;
}

.settings-row__control {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: none;
}

.settings-row__control--link {
  color: var(--color-neutral-400);
}

.settings-row--placeholder .settings-row__title {
  color: var(--color-neutral-300);
}

.settings-row--placeholder:hover {
  background: rgba(255, 255, 255, 0.03);
}

/* ---------- 开关 ---------- */
.settings-switch :deep(.el-switch) {
  --el-switch-on-color: var(--color-primary-500);
  --el-switch-off-color: var(--color-neutral-300);
  --el-switch-button-size: 16px;
}

.settings-switch :deep(.el-switch__core) {
  min-width: 36px;
  height: 20px;
  border-radius: var(--radius-full);
}

.settings-switch :deep(.el-switch__action) {
  width: 16px;
  height: 16px;
}

/* ---------- 默认楼层 ---------- */
.settings-row--floor {
  flex-wrap: nowrap;
}

.floor-select {
  display: flex;
  gap: 2px;
  padding: 2px;
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.06);
  box-shadow: inset 0 1px 2px rgba(0, 0, 0, 0.12);
}

.floor-select__item {
  min-width: 38px;
  height: 28px;
  padding: 0 6px;
  border: 0;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-neutral-300);
  font-family: var(--font-family-number);
  font-size: 12px;
  font-weight: 500;
  line-height: 18px;
  transition:
    background-color var(--duration-fast) var(--ease-standard),
    color var(--duration-fast) var(--ease-standard),
    box-shadow var(--duration-fast) var(--ease-standard);
}

.floor-select__item:hover {
  background: rgba(255, 255, 255, 0.06);
  color: var(--color-neutral-0);
}

.floor-select__item.active {
  background: var(--color-neutral-0);
  box-shadow: var(--shadow-1);
  color: var(--color-primary-600);
}

/* ---------- 标签 / 清除 ---------- */
.settings-tag {
  padding: 1px 6px;
  border-radius: var(--radius-xs);
  background: rgba(255, 255, 255, 0.08);
  color: var(--color-neutral-300);
  font-size: 12px;
  font-weight: 500;
  line-height: 18px;
  white-space: nowrap;
}

.settings-clear {
  height: 32px;
  padding: 0 12px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-neutral-300);
  font-size: 12px;
  font-weight: 500;
  line-height: 18px;
  transition:
    background-color var(--duration-fast) var(--ease-standard),
    border-color var(--duration-fast) var(--ease-standard),
    color var(--duration-fast) var(--ease-standard);
}

.settings-clear:hover {
  border-color: rgba(255, 255, 255, 0.2);
  background: rgba(255, 255, 255, 0.06);
  color: var(--color-neutral-0);
}

.settings-version {
  color: var(--color-neutral-400);
  font-family: var(--font-family-number);
  font-size: 12px;
  font-weight: 500;
  line-height: 18px;
}

.settings-footnote {
  margin-top: 32px;
  color: var(--color-neutral-400);
  font-size: 12px;
  font-weight: 400;
  line-height: 18px;
  text-align: center;
}

/* ---------- Toast ---------- */
.settings-toast {
  position: fixed;
  bottom: 32px;
  left: 50%;
  z-index: 90;
  max-width: calc(100vw - 32px);
  padding: 10px 16px;
  border: 1px solid var(--color-neutral-700);
  border-radius: var(--radius-md);
  background: var(--color-neutral-900);
  box-shadow: var(--shadow-3);
  color: var(--color-neutral-0);
  font-size: 14px;
  font-weight: 400;
  line-height: 22px;
  text-align: center;
  transform: translateX(-50%);
}

/* ---------- 关于弹层 ---------- */
.about-overlay {
  position: fixed;
  inset: 0;
  z-index: 80;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background: rgba(15, 23, 42, 0.65);
}

.about-card {
  width: min(400px, calc(100vw - 32px));
  padding: 24px;
  border: 1px solid var(--color-neutral-700);
  border-radius: var(--radius-lg);
  background: var(--color-neutral-900);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.06),
    var(--shadow-3);
  color: var(--color-neutral-0);
}

.about-eyebrow {
  margin-bottom: 8px;
  color: var(--color-neutral-400);
  font-size: 12px;
  font-weight: 500;
  line-height: 18px;
  letter-spacing: 0.08em;
}

.about-title {
  color: var(--color-neutral-0);
  font-size: 20px;
  font-weight: 600;
  line-height: 28px;
}

.about-subtitle {
  margin-top: 8px;
  color: var(--color-neutral-300);
  font-size: 14px;
  font-weight: 400;
  line-height: 22px;
}

.about-note {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--color-neutral-700);
  color: var(--color-neutral-400);
  font-size: 12px;
  font-weight: 400;
  line-height: 18px;
}

.about-version {
  margin-top: 4px;
  color: var(--color-neutral-400);
  font-family: var(--font-family-number);
  font-size: 12px;
  line-height: 18px;
}

.about-close {
  width: 100%;
  height: 40px;
  margin-top: 24px;
  border: 1px solid var(--color-neutral-700);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-neutral-300);
  font-size: 14px;
  font-weight: 500;
  line-height: 22px;
  transition:
    background-color var(--duration-fast) var(--ease-standard),
    border-color var(--duration-fast) var(--ease-standard),
    color var(--duration-fast) var(--ease-standard),
    transform var(--duration-fast) var(--ease-standard);
}

.about-close:hover {
  border-color: var(--color-neutral-500);
  background: rgba(255, 255, 255, 0.06);
  color: var(--color-neutral-0);
}

.about-close:active {
  transform: translateY(1px);
}

/* ---------- 过渡 ---------- */
.settings-toast-enter-active,
.settings-toast-leave-active {
  transition:
    opacity 0.18s var(--ease-standard),
    transform 0.18s var(--ease-standard);
}

.settings-toast-enter-from,
.settings-toast-leave-to {
  opacity: 0;
  transform: translate(-50%, 8px);
}

.about-enter-active,
.about-leave-active {
  transition: opacity 0.18s var(--ease-standard);
}

.about-enter-from,
.about-leave-to {
  opacity: 0;
}

/* ---------- 响应式 ---------- */
@media (min-width: 1024px) and (max-width: 1279px) {
  .settings-content,
  .settings-header__inner {
    max-width: 680px;
  }
}

@media (max-width: 767px) {
  .settings-header__inner {
    height: 60px;
    padding: 0 16px;
  }

  .settings-title {
    font-size: 20px;
    line-height: 28px;
  }

  .settings-content {
    padding: 20px 16px 40px;
  }

  .settings-row {
    min-height: 64px;
  }

  .settings-row--floor {
    flex-wrap: wrap;
    padding-top: 12px;
    padding-bottom: 12px;
  }

  .settings-row--floor .settings-row__control {
    width: 100%;
    margin-top: 4px;
    padding-left: 48px;
  }

  .floor-select {
    width: 100%;
  }

  .floor-select__item {
    flex: 1;
    min-width: 0;
  }

  .settings-toast {
    bottom: 24px;
  }
}
</style>
