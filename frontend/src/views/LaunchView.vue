<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()
const aboutDialogRef = ref(null)

// 进入 P2，默认楼层为 1 楼，通过 query 传递，避免 P2 尚未开发时丢失楼层语义
const enterNavigation = () => {
  router.push({ name: 'map', query: { floor: '1' } })
}

const openAbout = () => {
  const dialog = aboutDialogRef.value
  if (dialog && !dialog.open) {
    dialog.showModal()
  }
}

const closeAbout = () => {
  const dialog = aboutDialogRef.value
  if (dialog?.open) {
    dialog.close()
  }
}

// 点击 dialog 自身（即遮罩区域）时关闭，点击内容区不关闭
const handleDialogClick = (event) => {
  if (event.target === aboutDialogRef.value) {
    closeAbout()
  }
}
</script>

<template>
  <main class="launch-page">
    <!-- 背景装饰：网格 / 主色氛围光 / 等距楼宇线稿，均不参与交互 -->
    <div class="bg-grid" aria-hidden="true"></div>
    <div class="bg-glow" aria-hidden="true"></div>
    <svg
      class="bg-building"
      viewBox="0 0 960 760"
      fill="none"
      aria-hidden="true"
      focusable="false"
    >
      <g
        stroke="currentColor"
        stroke-width="1.2"
        stroke-linecap="round"
        stroke-linejoin="round"
        vector-effect="non-scaling-stroke"
      >
        <!-- 四条结构柱 -->
        <path d="M260 360V560M500 420V620M740 360V560M500 300V500" />

        <!-- 顶层楼板 -->
        <polygon points="260,360 500,420 740,360 500,300" />
        <polygon points="260,368 500,428 740,368 500,308" />
        <path d="M260 360V368M500 420V428M740 360V368M500 300V308" />
        <path d="M380 330L620 390M380 390L620 330" />

        <!-- 中层楼板 -->
        <polygon points="260,460 500,520 740,460 500,400" />
        <polygon points="260,468 500,528 740,468 500,408" />
        <path d="M260 460V468M500 520V528M740 460V468M500 400V408" />
        <path d="M380 430L620 490M380 490L620 430" />

        <!-- 底层楼板 -->
        <polygon points="260,560 500,620 740,560 500,500" />
        <polygon points="260,568 500,628 740,568 500,508" />
        <path d="M260 560V568M500 620V628M740 560V568M500 500V508" />
        <path d="M380 530L620 590M380 590L620 530" />
      </g>
    </svg>

    <header class="topbar">
      <button type="button" class="about-trigger" @click="openAbout">关于</button>
    </header>

    <section class="hero" aria-labelledby="launch-title">
      <h1 id="launch-title" class="hero__title">校园室内导航</h1>
      <p class="hero__subtitle">3D视角 · 教学楼内快速找路</p>
      <button type="button" class="enter-button" @click="enterNavigation">进入导航</button>
    </section>

    <p class="footnote">仅限已建模教学楼</p>

    <!-- 原生 dialog：支持 ESC 关闭与焦点管理，样式全部走全局令牌 -->
    <dialog
      ref="aboutDialogRef"
      class="about-dialog"
      aria-labelledby="about-title"
      @click="handleDialogClick"
    >
      <div class="about-dialog__inner">
        <p class="about-dialog__eyebrow">关于</p>
        <h2 id="about-title" class="about-dialog__title">校园室内导航</h2>
        <p class="about-dialog__subtitle">3D视角 · 教学楼内快速找路</p>
        <p class="about-dialog__note">仅限已建模教学楼</p>
        <button type="button" class="about-dialog__close" @click="closeAbout">知道了</button>
      </div>
    </dialog>
  </main>
</template>

<style scoped lang="less">
.launch-page {
  position: relative;
  isolation: isolate;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  min-height: 100dvh;
  padding: 88px 24px 96px;
  overflow: hidden;
  background: var(--color-map-3d-bg);
}

/* ---------- 背景：网格 ---------- */
.bg-grid {
  position: absolute;
  inset: 0;
  z-index: 0;
  background-image:
    linear-gradient(rgba(199, 212, 228, 0.05) 1px, transparent 1px),
    linear-gradient(90deg, rgba(199, 212, 228, 0.05) 1px, transparent 1px);
  background-size: 48px 48px;
  -webkit-mask-image: radial-gradient(ellipse at 50% 42%, #000 0%, transparent 72%);
  mask-image: radial-gradient(ellipse at 50% 42%, #000 0%, transparent 72%);
  pointer-events: none;
}

/* ---------- 背景：主色氛围光 ---------- */
.bg-glow {
  position: absolute;
  inset: 0;
  z-index: 0;
  background: radial-gradient(
    ellipse at 50% 42%,
    rgba(43, 107, 229, 0.1) 0%,
    rgba(43, 107, 229, 0) 62%
  );
  pointer-events: none;
}

/* ---------- 背景：等距楼宇线稿 ---------- */
.bg-building {
  position: absolute;
  right: -140px;
  bottom: -150px;
  z-index: 1;
  width: 780px;
  height: auto;
  color: var(--color-map-3d-edge);
  opacity: 0.1;
  pointer-events: none;
  user-select: none;
}

/* ---------- 右上角关于入口 ---------- */
.topbar {
  position: absolute;
  top: 24px;
  right: 24px;
  z-index: 3;
}

.about-trigger {
  min-width: 40px;
  min-height: 40px;
  padding: 8px 12px;
  border: 0;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-text-on-dark-tertiary);
  font-size: 14px;
  font-weight: 400;
  line-height: 22px;
  -webkit-tap-highlight-color: transparent;
}

.about-trigger:hover {
  color: var(--color-text-on-dark);
  background: rgba(255, 255, 255, 0.06);
}

.about-trigger:focus-visible {
  outline: 2px solid var(--color-neutral-0);
  outline-offset: 2px;
}

/* ---------- 主内容 ---------- */
.hero {
  position: relative;
  z-index: 2;
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
  max-width: 640px;
  text-align: center;
}

.hero__title {
  color: var(--color-text-on-dark);
  font-size: 32px;
  font-weight: 600;
  line-height: 40px;
}

.hero__subtitle {
  margin-top: 12px;
  color: var(--color-text-on-dark-secondary);
  font-size: 16px;
  font-weight: 400;
  line-height: 24px;
}

/* ---------- 主按钮：本页唯一动效 ---------- */
.enter-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 240px;
  height: 48px;
  margin-top: 32px;
  padding: 0 24px;
  border: 0;
  border-radius: var(--radius-sm);
  background: var(--color-primary-500);
  box-shadow: var(--shadow-button);
  color: var(--color-neutral-0);
  font-family: var(--font-family-base);
  font-size: 16px;
  font-weight: 500;
  line-height: 1;
  letter-spacing: 0.04em;
  user-select: none;
  -webkit-tap-highlight-color: transparent;
  transition:
    transform var(--duration-slow) var(--ease-standard),
    background-color var(--duration-base) var(--ease-standard),
    box-shadow var(--duration-slow) var(--ease-standard);
}

.enter-button:hover {
  background: var(--color-primary-600);
  box-shadow: var(--shadow-button-hover);
  transform: translateY(-2px);
}

.enter-button:active {
  background: var(--color-primary-700);
  box-shadow: var(--shadow-button-active);
  transform: translateY(0);
}

.enter-button:focus-visible {
  outline: 2px solid var(--color-neutral-0);
  outline-offset: 4px;
}

/* ---------- 底部说明 ---------- */
.footnote {
  position: absolute;
  right: 0;
  bottom: 32px;
  left: 0;
  z-index: 2;
  color: var(--color-text-on-dark-secondary);
  font-size: 12px;
  font-weight: 400;
  line-height: 18px;
  text-align: center;
  pointer-events: none;
}

/* ---------- 关于弹层 ---------- */
.about-dialog {
  position: fixed;
  inset: 0;
  width: min(400px, calc(100vw - 32px));
  max-height: calc(100dvh - 32px);
  margin: auto;
  padding: 0;
  overflow: auto;
  border: 1px solid var(--color-neutral-700);
  border-radius: var(--radius-lg);
  background: var(--color-neutral-900);
  box-shadow: var(--shadow-3);
  color: var(--color-text-on-dark);
}

.about-dialog::backdrop {
  background: rgba(15, 23, 42, 0.65);
}

.about-dialog__inner {
  padding: 24px;
}

.about-dialog__eyebrow {
  margin-bottom: 8px;
  color: var(--color-text-on-dark-secondary);
  font-size: 12px;
  font-weight: 500;
  line-height: 18px;
  letter-spacing: 0.08em;
}

.about-dialog__title {
  color: var(--color-text-on-dark);
  font-size: 20px;
  font-weight: 600;
  line-height: 28px;
}

.about-dialog__subtitle {
  margin-top: 8px;
  color: var(--color-text-on-dark-tertiary);
  font-size: 14px;
  font-weight: 400;
  line-height: 22px;
}

.about-dialog__note {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--color-neutral-700);
  color: var(--color-text-on-dark-secondary);
  font-size: 12px;
  font-weight: 400;
  line-height: 18px;
}

.about-dialog__close {
  width: 100%;
  height: 40px;
  margin-top: 24px;
  border: 1px solid var(--color-neutral-700);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--color-text-on-dark-tertiary);
  font-size: 14px;
  font-weight: 500;
  line-height: 22px;
  -webkit-tap-highlight-color: transparent;
}

.about-dialog__close:hover {
  border-color: var(--color-neutral-500);
  background: rgba(255, 255, 255, 0.06);
  color: var(--color-text-on-dark);
}

.about-dialog__close:focus-visible {
  outline: 2px solid var(--color-neutral-0);
  outline-offset: 2px;
}

/* ---------- 平板 ---------- */
@media (min-width: 768px) and (max-width: 1023px) {
  .bg-building {
    right: -170px;
    bottom: -80px;
    width: 560px;
  }
}

/* ---------- 移动端 ---------- */
@media (max-width: 767px) {
  .launch-page {
    padding: 80px 16px 88px;
  }

  .bg-grid {
    background-size: 40px 40px;
  }

  .bg-building {
    right: -170px;
    bottom: -40px;
    width: 440px;
    opacity: 0.07;
  }

  .topbar {
    top: 16px;
    right: 16px;
  }

  .hero__title {
    font-size: 24px;
    line-height: 32px;
  }

  .hero__subtitle {
    font-size: 14px;
    line-height: 22px;
  }

  .enter-button {
    width: min(320px, 100%);
    margin-top: 24px;
  }

  .footnote {
    bottom: 24px;
  }
}

/* ---------- 矮屏：压缩纵向间距，保持单屏 ---------- */
@media (max-height: 600px) {
  .launch-page {
    padding: 72px 24px 64px;
  }

  .hero__subtitle {
    margin-top: 8px;
  }

  .enter-button {
    margin-top: 24px;
  }

  .footnote {
    bottom: 16px;
  }
}
</style>
