<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import MapIcon from '../MapIcon.vue'

const props = defineProps({
  location: {
    type: Object,
    required: true
  }
})

const emit = defineEmits(['close', 'navigate-here', 'start-here'])

const typeMeta = {
  classroom: { icon: 'classroom', className: 'classroom', label: '教室' },
  teacher_rest: { icon: 'classroom', className: 'teacher', label: '教师休息室' },
  restroom: { icon: 'toilet', className: 'restroom', label: '卫生间' },
  elevator: { icon: 'elevator', className: 'elevator', label: '电梯' },
  stairs: { icon: 'stairs', className: 'stairs', label: '楼梯' },
  exit: { icon: 'exit', className: 'exit', label: '安全出口' },
  water: { icon: 'water', className: 'water', label: '开水间' },
  auditorium: { icon: 'palette', className: 'auditorium', label: '演播厅' },
  entrance: { icon: 'exit', className: 'entrance', label: '教学楼入口' },
  current: { icon: 'locate', className: 'current', label: '当前位置' }
}

const meta = computed(
  () =>
    typeMeta[props.location.type] || {
      icon: 'pin',
      className: 'classroom',
      label: '地点'
    }
)

const primaryButtonRef = ref(null)

const handleKeydown = (event) => {
  if (event.key === 'Escape') {
    emit('close')
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeydown)
  nextTick(() => {
    primaryButtonRef.value?.focus()
  })
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleKeydown)
})
</script>

<template>
  <div class="p5-overlay" @click.self="emit('close')">
    <section
      class="p5-sheet"
      role="dialog"
      aria-modal="true"
      aria-labelledby="p5-title"
      @click.stop
    >
      <div class="p5-sheet__handle" aria-hidden="true"></div>

      <header class="p5-sheet__head">
        <span
          class="p5-sheet__icon"
          :class="`p5-sheet__icon--${meta.className}`"
        >
          <MapIcon :name="meta.icon" :size="22" />
        </span>

        <div class="p5-sheet__body">
          <p class="p5-sheet__type">{{ meta.label }}</p>
          <h2 id="p5-title" class="p5-sheet__title">
            {{ location.name }}
          </h2>
          <p class="p5-sheet__meta">
            {{ location.buildingName }} · {{ location.floor }} · 距你
            {{ location.distance }} m
          </p>
        </div>

        <button
          type="button"
          class="p5-sheet__close"
          aria-label="关闭地点详情"
          @click="emit('close')"
        >
          <MapIcon name="close" :size="18" />
        </button>
      </header>

      <div class="p5-sheet__actions">
        <button
          ref="primaryButtonRef"
          type="button"
          class="p5-button p5-button--primary"
          @click="emit('navigate-here')"
        >
          <span class="p5-button__content">
            <MapIcon name="nav" :size="18" />
            <span>到这去</span>
          </span>
        </button>

        <button
          type="button"
          class="p5-button p5-button--secondary"
          @click="emit('start-here')"
        >
          <span class="p5-button__content">
            <MapIcon name="pin" :size="18" />
            <span>从这出发</span>
          </span>
        </button>
      </div>
    </section>
  </div>
</template>

<style scoped lang="less">
.p5-overlay {
  position: absolute;
  inset: 0;
  z-index: 60;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  padding: 24px;
  background: rgba(15, 23, 42, 0.35);
}

.p5-sheet {
  width: min(520px, 100%);
  max-height: calc(100dvh - 48px);
  padding: 20px;
  overflow-y: auto;
  border: 1px solid rgba(255, 255, 255, 0.72);
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.9);
  box-shadow:
    0 0 0 1px rgba(15, 23, 42, 0.04),
    inset 0 1px 0 rgba(255, 255, 255, 0.85),
    var(--shadow-3);
  color: var(--color-neutral-900);
  animation: p5-sheet-in 0.24s var(--ease-standard);
  -webkit-backdrop-filter: blur(20px) saturate(1.2);
  backdrop-filter: blur(20px) saturate(1.2);
}

.p5-sheet__handle {
  display: none;
}

.p5-sheet__head {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.p5-sheet__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  flex: none;
  border-radius: var(--radius-md);
}

.p5-sheet__icon--classroom {
  background: var(--color-neutral-100);
  color: var(--color-neutral-700);
}

.p5-sheet__icon--stairs,
.p5-sheet__icon--water,
.p5-sheet__icon--current {
  background: var(--color-accent-50);
  color: var(--color-accent-600);
}

.p5-sheet__icon--elevator {
  background: var(--color-primary-50);
  color: var(--color-primary-600);
}

.p5-sheet__icon--restroom,
.p5-sheet__icon--exit,
.p5-sheet__icon--teacher {
  background: var(--color-warning-50);
  color: var(--color-warning-600);
}

.p5-sheet__icon--auditorium {
  background: #F1E9FF;
  color: #7C5CFC;
}

.p5-sheet__icon--entrance {
  background: var(--color-primary-50);
  color: var(--color-primary-600);
}

.p5-sheet__body {
  min-width: 0;
  flex: 1;
}

.p5-sheet__type {
  color: var(--color-neutral-500);
  font-size: 12px;
  font-weight: 500;
  line-height: 18px;
}

.p5-sheet__title {
  margin-top: 2px;
  color: var(--color-neutral-900);
  font-size: 24px;
  font-weight: 600;
  line-height: 32px;
}

.p5-sheet__meta {
  margin-top: 4px;
  color: var(--color-neutral-500);
  font-size: 14px;
  font-weight: 400;
  line-height: 22px;
}

.p5-sheet__close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  flex: none;
  border: 0;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-neutral-500);
  transition:
    background-color var(--duration-fast) var(--ease-standard),
    color var(--duration-fast) var(--ease-standard);
}

.p5-sheet__close:hover {
  background: var(--color-neutral-100);
  color: var(--color-neutral-900);
}

.p5-sheet__close:focus-visible {
  outline: 2px solid var(--color-primary-500);
  outline-offset: 2px;
}

.p5-sheet__actions {
  display: flex;
  gap: 8px;
  margin-top: 20px;
}

.p5-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 48px;
  flex: 1;
  border-radius: var(--radius-md);
  font-family: var(--font-family-base);
  font-size: 14px;
  font-weight: 500;
  line-height: 22px;
  transition:
    background-color var(--duration-fast) var(--ease-standard),
    border-color var(--duration-fast) var(--ease-standard),
    box-shadow var(--duration-fast) var(--ease-standard),
    transform var(--duration-fast) var(--ease-standard);
}

.p5-button__content {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.p5-button--primary {
  border: 0;
  background: var(--color-primary-500);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.18),
    0 6px 16px rgba(43, 107, 229, 0.24);
  color: var(--color-neutral-0);
  font-weight: 600;
}

.p5-button--primary:hover {
  background: var(--color-primary-600);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.16),
    0 8px 20px rgba(43, 107, 229, 0.28);
}

.p5-button--primary:active {
  background: var(--color-primary-700);
  box-shadow:
    inset 0 1px 2px rgba(15, 23, 42, 0.12),
    0 2px 6px rgba(43, 107, 229, 0.2);
  transform: translateY(1px) scale(0.997);
}

.p5-button--secondary {
  border: 1px solid var(--color-neutral-200);
  background: var(--color-neutral-0);
  color: var(--color-neutral-700);
}

.p5-button--secondary:hover {
  border-color: var(--color-neutral-300);
  background: var(--color-neutral-50);
  color: var(--color-neutral-900);
}

.p5-button--secondary:active {
  background: var(--color-neutral-100);
  transform: translateY(1px);
}

.p5-button:focus-visible {
  outline: 2px solid var(--color-primary-500);
  outline-offset: 2px;
}

@keyframes p5-sheet-in {
  from {
    opacity: 0;
    transform: translateY(24px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (min-width: 768px) and (max-width: 1279px) {
  .p5-sheet {
    width: 480px;
  }
}

@media (max-width: 767px) {
  .p5-overlay {
    padding: 0;
  }

  .p5-sheet {
    width: 100%;
    max-height: 80dvh;
    padding: 20px;
    padding-bottom: calc(20px + env(safe-area-inset-bottom));
    border-radius: var(--radius-lg) var(--radius-lg) 0 0;
  }

  .p5-sheet__handle {
    display: block;
    width: 36px;
    height: 4px;
    margin: 0 auto 12px;
    border-radius: var(--radius-full);
    background: var(--color-neutral-300);
  }

  .p5-sheet__title {
    font-size: 22px;
    line-height: 30px;
  }
}
</style>
