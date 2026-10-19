<script setup>
import { computed } from 'vue'
import MapIcon from '../MapIcon.vue'

const props = defineProps({
  navigation: {
    type: Object,
    required: true
  }
})

const emit = defineEmits([
  'toggle-pause',
  'end',
  'exit-follow',
  'resume-follow'
])

const isFree = computed(
  () => !props.navigation.following && !props.navigation.arrived
)
const isPaused = computed(() => props.navigation.paused)
const isArrived = computed(() => props.navigation.arrived)
const progressWidth = computed(
  () => `${Math.round(props.navigation.progress * 100)}%`
)
</script>

<template>
  <div class="nav-overlay">
    <Transition name="nav-card">
      <section v-if="!isFree" class="nav-card" aria-live="polite">
        <div class="nav-card__main">
          <span class="nav-card__icon">
            <MapIcon :name="navigation.actionIcon" :size="24" />
          </span>

          <div class="nav-card__body">
            <div class="nav-card__action-row">
              <h2 class="nav-card__action">{{ navigation.action }}</h2>
              <span v-if="isPaused && !isArrived" class="nav-card__chip">
                已暂停
              </span>
              <span
                v-else-if="navigation.transfer"
                class="nav-card__chip nav-card__chip--transfer"
              >
                换乘
              </span>
            </div>
            <div class="nav-card__meta">
              <span class="nav-card__floor">{{ navigation.floorLabel }}</span>
            </div>
          </div>

          <div class="nav-card__distance">
            <span class="nav-card__distance-label">剩余</span>
            <div class="nav-card__distance-row">
              <span class="nav-card__distance-value">
                {{ navigation.remaining }}
              </span>
              <span class="nav-card__distance-unit">m</span>
            </div>
          </div>
        </div>

        <div class="nav-card__progress" aria-hidden="true">
          <span :style="{ width: progressWidth }"></span>
        </div>
      </section>
    </Transition>

    <Transition name="nav-free">
      <div v-if="isFree" class="nav-free">
        <span class="nav-free__text">自由浏览中 · 导航已暂停</span>
        <span class="nav-free__floor">{{ navigation.floorLabel }}</span>
      </div>
    </Transition>

    <Transition name="nav-hint">
      <div
        v-if="navigation.hint"
        class="nav-hint"
        :class="`nav-hint--${navigation.hintType}`"
      >
        <MapIcon
          :name="navigation.hintType === 'warning' ? 'alert' : 'info'"
          :size="14"
        />
        <span>{{ navigation.hint }}</span>
      </div>
    </Transition>

    <div class="nav-actions">
      <template v-if="isArrived">
        <el-button class="nav-end nav-end--full" @click="emit('end')">
          结束导航
        </el-button>
      </template>

      <template v-else-if="isFree">
        <el-button
          type="primary"
          class="nav-primary"
          @click="emit('resume-follow')"
        >
          <span class="nav-btn-content">
            <MapIcon name="nav" :size="16" />
            <span>恢复跟随</span>
          </span>
        </el-button>
        <el-button class="nav-end" @click="emit('end')">
          结束导航
        </el-button>
      </template>

      <template v-else>
        <el-button
          type="primary"
          class="nav-primary"
          @click="emit('toggle-pause')"
        >
          <span class="nav-btn-content">
            <MapIcon :name="isPaused ? 'play' : 'pause'" :size="16" />
            <span>{{ isPaused ? '继续' : '暂停' }}</span>
          </span>
        </el-button>
        <el-button class="nav-end" @click="emit('end')">
          结束导航
        </el-button>
        <el-button
          class="nav-exit"
          aria-label="退出跟随"
          @click="emit('exit-follow')"
        >
          <span class="nav-btn-content">
            <MapIcon name="eye-off" :size="16" />
            <span>退出跟随</span>
          </span>
        </el-button>
      </template>
    </div>
  </div>
</template>

<style scoped lang="less">
.glass-strong() {
  background: rgba(255, 255, 255, 0.94);
  border: 1px solid rgba(255, 255, 255, 0.75);
  box-shadow:
    0 0 0 1px rgba(15, 23, 42, 0.06),
    var(--shadow-3);
  -webkit-backdrop-filter: blur(16px) saturate(1.1);
  backdrop-filter: blur(16px) saturate(1.1);
}

.nav-overlay {
  position: absolute;
  inset: 0;
  z-index: 50;
  pointer-events: none;
}

/* ---------- 导航步骤卡 ---------- */
.nav-card {
  .glass-strong();

  position: absolute;
  top: 24px;
  left: 50%;
  width: min(480px, calc(100vw - 48px));
  padding: 16px 20px 18px;
  border-radius: var(--radius-lg);
  color: var(--color-neutral-900);
  pointer-events: auto;
  transform: translateX(-50%);
}

.nav-card__main {
  display: flex;
  align-items: center;
  gap: 12px;
}

.nav-card__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  flex: none;
  border-radius: 50%;
  background: var(--color-primary-50);
  color: var(--color-primary-600);
}

.nav-card__body {
  min-width: 0;
  flex: 1;
}

.nav-card__action-row {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.nav-card__action {
  overflow: hidden;
  color: var(--color-neutral-900);
  font-size: 24px;
  font-weight: 600;
  line-height: 32px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.nav-card__chip {
  flex: none;
  padding: 1px 6px;
  border-radius: var(--radius-xs);
  background: var(--color-neutral-100);
  color: var(--color-neutral-700);
  font-size: 12px;
  font-weight: 500;
  line-height: 18px;
}

.nav-card__chip--transfer {
  background: var(--color-warning-50);
  color: var(--color-warning-600);
}

.nav-card__meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 4px;
}

.nav-card__floor {
  padding: 1px 6px;
  border-radius: var(--radius-xs);
  background: var(--color-primary-50);
  color: var(--color-primary-700);
  font-size: 12px;
  font-weight: 500;
  line-height: 18px;
}

.nav-card__distance {
  flex: none;
  text-align: right;
}

.nav-card__distance-label {
  display: block;
  color: var(--color-neutral-500);
  font-size: 12px;
  font-weight: 500;
  line-height: 18px;
}

.nav-card__distance-row {
  display: flex;
  align-items: baseline;
  justify-content: flex-end;
  gap: 2px;
}

.nav-card__distance-value {
  color: var(--color-neutral-900);
  font-family: var(--font-family-number);
  font-size: 20px;
  font-weight: 600;
  line-height: 28px;
  font-variant-numeric: tabular-nums;
}

.nav-card__distance-unit {
  color: var(--color-neutral-500);
  font-size: 12px;
  font-weight: 500;
  line-height: 18px;
}

.nav-card__progress {
  position: absolute;
  right: 20px;
  bottom: 7px;
  left: 20px;
  height: 2px;
  overflow: hidden;
  border-radius: var(--radius-full);
  background: var(--color-neutral-100);
}

.nav-card__progress span {
  display: block;
  height: 100%;
  border-radius: var(--radius-full);
  background: var(--color-primary-500);
  transition: width 120ms linear;
}

/* ---------- 自由浏览状态 ---------- */
.nav-free {
  .glass-strong();

  position: absolute;
  top: 24px;
  left: 50%;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  border-radius: var(--radius-md);
  pointer-events: auto;
  transform: translateX(-50%);
}

.nav-free__text {
  color: var(--color-neutral-700);
  font-size: 14px;
  font-weight: 500;
  line-height: 22px;
  white-space: nowrap;
}

.nav-free__floor {
  padding: 1px 6px;
  border-radius: var(--radius-xs);
  background: var(--color-primary-50);
  color: var(--color-primary-700);
  font-size: 12px;
  font-weight: 500;
  line-height: 18px;
  white-space: nowrap;
}

/* ---------- 提示条 ---------- */
.nav-hint {
  position: absolute;
  top: 140px;
  left: 50%;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-1);
  font-size: 12px;
  font-weight: 500;
  line-height: 18px;
  white-space: nowrap;
  pointer-events: none;
  transform: translateX(-50%);
}

.nav-hint--warning {
  border: 1px solid rgba(232, 163, 61, 0.28);
  background: var(--color-warning-50);
  color: var(--color-warning-600);
}

.nav-hint--primary {
  border: 1px solid rgba(43, 107, 229, 0.16);
  background: var(--color-primary-50);
  color: var(--color-primary-700);
}

/* ---------- 底部操作条 ---------- */
.nav-actions {
  .glass-strong();

  position: absolute;
  bottom: 24px;
  left: 50%;
  display: flex;
  align-items: center;
  gap: 8px;
  width: min(440px, calc(100vw - 32px));
  padding: 8px;
  border-radius: var(--radius-lg);
  pointer-events: auto;
  transform: translateX(-50%);
}

.nav-actions :deep(.el-button) {
  height: 44px;
  margin: 0;
  border-radius: var(--radius-sm);
  font-size: 14px;
  font-weight: 500;
}

.nav-actions :deep(.el-button + .el-button) {
  margin-left: 0;
}

.nav-btn-content {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.nav-primary {
  flex: 1;
}

.nav-primary:hover {
  border-color: var(--color-primary-600);
  background: var(--color-primary-600);
}

.nav-primary:active {
  border-color: var(--color-primary-700);
  background: var(--color-primary-700);
}

.nav-end {
  flex: 1;
  border: 1px solid var(--color-neutral-200);
  background: var(--color-neutral-0);
  color: var(--color-neutral-700);
}

.nav-end:hover {
  border-color: var(--color-neutral-300);
  background: var(--color-neutral-50);
  color: var(--color-neutral-900);
}

.nav-end--full {
  width: 100%;
  flex: none;
}

.nav-exit {
  flex: none;
  border: 0;
  background: transparent;
  color: var(--color-neutral-500);
}

.nav-exit:hover {
  background: var(--color-neutral-100);
  color: var(--color-neutral-900);
}

.nav-primary:focus-visible,
.nav-end:focus-visible,
.nav-exit:focus-visible {
  outline: 2px solid var(--color-primary-500);
  outline-offset: 2px;
}

/* ---------- 过渡 ---------- */
.nav-card-enter-active,
.nav-card-leave-active,
.nav-free-enter-active,
.nav-free-leave-active {
  transition:
    opacity 0.2s var(--ease-standard),
    transform 0.2s var(--ease-standard);
}

.nav-card-enter-from,
.nav-card-leave-to,
.nav-free-enter-from,
.nav-free-leave-to {
  opacity: 0;
  transform: translate(-50%, -8px);
}

.nav-hint-enter-active,
.nav-hint-leave-active {
  transition:
    opacity 0.18s var(--ease-standard),
    transform 0.18s var(--ease-standard);
}

.nav-hint-enter-from,
.nav-hint-leave-to {
  opacity: 0;
  transform: translate(-50%, -6px);
}

/* ---------- 响应式 ---------- */
@media (min-width: 768px) and (max-width: 1279px) {
  .nav-card {
    width: 440px;
  }
}

@media (max-width: 767px) {
  .nav-card {
    top: 72px;
    width: calc(100vw - 32px);
    padding: 14px 16px 18px;
  }

  .nav-card__icon {
    width: 40px;
    height: 40px;
  }

  .nav-card__action {
    font-size: 20px;
    line-height: 28px;
  }

  .nav-card__distance-value {
    font-size: 18px;
    line-height: 26px;
  }

  .nav-card__progress {
    right: 16px;
    left: 16px;
  }

  .nav-free {
    top: 72px;
    max-width: calc(100vw - 32px);
  }

  .nav-hint {
    top: 184px;
    max-width: calc(100vw - 32px);
    white-space: normal;
    text-align: center;
  }

  .nav-actions {
    bottom: 16px;
    width: calc(100vw - 32px);
  }

  .nav-actions :deep(.el-button) {
    padding: 0 10px;
  }

  .nav-exit {
    width: 44px;
    padding: 0;
  }

  .nav-exit .nav-btn-content > span {
    display: none;
  }
}
</style>
