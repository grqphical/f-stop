<script setup lang="ts">
import HelloWorld from './components/HelloWorld.vue'

fetch("/api/v1/health").then(async (response) => console.log(await response.json()))
</script>

<template>
  <HelloWorld />
</template>
