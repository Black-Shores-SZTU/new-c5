<script setup>
defineProps({
  name: {
    type: String,
    required: true
  },
  size: {
    type: [Number, String],
    default: 20
  }
})
</script>

<template>
  <svg
    class="map-icon"
    :width="size"
    :height="size"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="1.7"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
    focusable="false"
  >
    <template v-if="name === 'toilet'">
      <rect x="4" y="3" width="16" height="18" rx="3" />
      <text
        x="12"
        y="16"
        fill="currentColor"
        stroke="none"
        font-size="7.5"
        font-weight="600"
        text-anchor="middle"
      >
        WC
      </text>
    </template>

    <template v-else-if="name === 'elevator'">
      <rect x="5" y="3" width="14" height="18" rx="2.5" />
      <path d="M9.5 10l2.5-2.5 2.5 2.5M9.5 14l2.5 2.5 2.5-2.5" />
    </template>

    <template v-else-if="name === 'stairs'">
      <path d="M4 20h4v-4h4v-4h4V8h4" />
      <path d="M4 20h16" opacity="0.35" />
    </template>

    <template v-else-if="name === 'exit'">
      <path d="M14 4h5v16h-5" />
      <path d="M4 12h11M11 8l4 4-4 4" />
    </template>

    <template v-else-if="name === 'water'">
      <path d="M6.5 9h11l-1.3 11h-8.4L6.5 9Z" />
      <path d="M9.5 5.5c0 1 .9 1 .9 2s-.9 1-.9 2M13.5 4.5c0 1 .9 1 .9 2s-.9 1-.9 2" />
    </template>

    <template v-else-if="name === 'plus'">
      <path d="M12 5v14M5 12h14" />
    </template>

    <template v-else-if="name === 'minus'">
      <path d="M5 12h14" />
    </template>

    <template v-else-if="name === 'reset'">
      <path d="M4 10a8 8 0 1 1 2.3 7.7" />
      <path d="M4 4v6h6" />
    </template>

    <template v-else-if="name === 'locate'">
      <circle cx="12" cy="12" r="3" />
      <path d="M12 2v3M12 19v3M2 12h3M19 12h3" />
    </template>

    <template v-else-if="name === 'pin'">
      <path d="M12 21s6-5.2 6-10a6 6 0 1 0-12 0c0 4.8 6 10 6 10Z" />
      <circle cx="12" cy="11" r="2" />
    </template>

    <template v-else-if="name === 'arrow-right'">
      <path d="M4 12h15M14 6l6 6-6 6" />
    </template>

    <template v-else-if="name === 'chevron-down'">
      <path d="m6 9 6 6 6-6" />
    </template>

    <template v-else-if="name === 'chevron-up'">
      <path d="m6 15 6-6 6 6" />
    </template>

    <template v-else-if="name === 'nav'">
      <path d="m4 11 16-7-5.5 16-3.2-6.2L4 11Z" />
      <path d="m11.3 13.8 3.2-4.2" />
    </template>

    <template v-else-if="name === 'building'">
      <path d="M4 21V5l8-2v18M12 21V9l8 2v10" />
      <path d="M7 8h2M7 12h2M7 16h2M15 13h2M15 17h2" />
    </template>

    <template v-else-if="name === 'straight'">
      <path d="M12 20V5M6 11l6-6 6 6" />
    </template>

    <template v-else-if="name === 'turn-left'">
      <path d="M18 20v-7a3 3 0 0 0-3-3H6" />
      <path d="m10 6-4 4 4 4" />
    </template>

    <template v-else-if="name === 'turn-right'">
      <path d="M6 20v-7a3 3 0 0 1 3-3h9" />
      <path d="m14 6 4 4-4 4" />
    </template>

    <template v-else-if="name === 'flag'">
      <path d="M6 21V4" />
      <path d="M6 5h11l-2.5 3.5L17 12H6" />
    </template>

    <template v-else-if="name === 'pause'">
      <path d="M9 5v14M15 5v14" stroke-width="2.4" />
    </template>

    <template v-else-if="name === 'play'">
      <path d="M8 5l11 7-11 7V5Z" />
    </template>

    <template v-else-if="name === 'eye-off'">
      <path d="M3 12s3.5-6 9-6 9 6 9 6-3.5 6-9 6-9-6-9-6Z" />
      <circle cx="12" cy="12" r="2.5" />
      <path d="m4 4 16 16" />
    </template>

    <template v-else-if="name === 'alert'">
      <path d="M12 3 2 21h20L12 3Z" />
      <path d="M12 9v5M12 17h.01" />
    </template>

    <template v-else-if="name === 'info'">
      <circle cx="12" cy="12" r="9" />
      <path d="M12 11v5M12 8h.01" />
    </template>

    <template v-else-if="name === 'search'">
      <circle cx="11" cy="11" r="6.5" />
      <path d="m16 16 4 4" />
    </template>

    <template v-else-if="name === 'close'">
      <path d="M6 6l12 12M18 6 6 18" />
    </template>

    <template v-else-if="name === 'classroom'">
      <rect x="4" y="4" width="16" height="16" rx="2" />
      <path d="M9 20v-6h6v6M8 8h2M14 8h2" />
    </template>

    <template v-else-if="name === 'settings'">
      <path
        d="M12 3.5l1.2 2.1 2.4-.4.8 2.3 2.3.8-.4 2.4 2.1 1.2-2.1 1.2.4 2.4-2.3.8-.8 2.3-2.4-.4L12 20.5l-1.2-2.1-2.4.4-.8-2.3-2.3-.8.4-2.4L3.6 12l2.1-1.2-.4-2.4 2.3-.8.8-2.3 2.4.4L12 3.5Z"
      />
      <circle cx="12" cy="12" r="3" />
    </template>

    <template v-else-if="name === 'chevron-left'">
      <path d="m15 6-6 6 6 6" />
    </template>

    <template v-else-if="name === 'chevron-right'">
      <path d="m9 6 6 6-6 6" />
    </template>

    <template v-else-if="name === 'palette'">
      <path
        d="M12 3a9 9 0 1 0 0 18h1.4a2.4 2.4 0 0 0 0-4.8H12a2 2 0 0 1 0-4h4a5 5 0 0 0 0-10h-4Z"
      />
      <circle cx="8" cy="9" r="1" />
      <circle cx="8" cy="15" r="1" />
      <circle cx="12" cy="6" r="1" />
    </template>

    <template v-else-if="name === 'history'">
      <path d="M4 12a8 8 0 1 0 2.3-5.7L4 8.6" />
      <path d="M4 4v4.6h4.6M12 8v4l3 2" />
    </template>

    <template v-else-if="name === 'trash'">
      <path d="M4 7h16M9 7V4h6v3M6 7l1 14h10l1-14M10 11v6M14 11v6" />
    </template>
  </svg>
</template>

<style scoped>
.map-icon {
  display: block;
  flex: none;
}
</style>
