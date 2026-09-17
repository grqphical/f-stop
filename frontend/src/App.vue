<script setup lang="ts">
import { RouterView } from 'vue-router';

fetch("/api/v1/health").then(async (response) => console.log(await response.json()))
</script>

<template>
  <RouterView />
</template>
