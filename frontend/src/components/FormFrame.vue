<script setup lang="ts">
defineProps<{ title: string }>()
</script>

<template>
  <div class="frame">
    <div class="frame__header">{{ title }}</div>
    <div class="frame__body">
      <slot />
    </div>
  </div>
</template>

<style scoped>
.frame {
  width: 100%;
  max-width: 760px;
  border: 1px solid #222;
  background: #fff;
}
.frame__header {
  background: #00b050;
  color: #fff;
  text-align: center;
  font-weight: 600;
  padding: 8px;
  letter-spacing: 0.5px;
}
.frame__body {
  padding: 48px 24px;
  min-height: 360px;
}
</style>
