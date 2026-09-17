<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import type { PhotoMetadata } from '../models/models';

const route = useRoute()

const id = route.params.id;
const photoMetadata = ref({} as PhotoMetadata);

function formatByteSize(bytes: number, decimals: number = 2): string {
    if (!+bytes) return '0 Bytes'

    const k = 1024
    const dm = decimals < 0 ? 0 : decimals
    const sizes = ['Bytes', 'KiB', 'MiB', 'GiB', 'TiB', 'PiB', 'EiB', 'ZiB', 'YiB']

    const i = Math.floor(Math.log(bytes) / Math.log(k))

    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`
}

onMounted(async () => {
    const response = await fetch(`/api/v1/photo/${id}`)
    if (response.status !== 200) {
        console.error("Got status code:", response.status)
        return
    }

    const jsonData = await response.json();
    photoMetadata.value = jsonData;
})
</script>

<template>
    <p>Camera Model: {{ photoMetadata.exifCameraModel }}</p>
    <p>Taken At: {{ new Date(photoMetadata.exifTakenAt!).toDateString() }}</p>
    <p>Uploaded At: {{ new Date(photoMetadata.uploaded!).toDateString() }}</p>
    <p>File Size: {{ formatByteSize(photoMetadata.size) }}</p>
    <img :src="photoMetadata.permalink" alt="">
</template>