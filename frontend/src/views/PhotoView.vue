<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import type { PhotoMetadata, PhotoMetadataDTO } from '../models/models';
import { router } from '../router';
import { formatByteSize } from '../utils';

const route = useRoute()

const id = route.params.id;
const photoMetadata = ref({} as PhotoMetadata);
const notFound = ref(false);

async function deletePhoto() {
    if (!confirm("Are you sure you want to delete this photo?")) {
        return
    }

    const response = await fetch(`/api/v1/photo/${id}`, {
        method: "DELETE"
    })
    if (response.status !== 200) {
        console.error(await response.text())
        return;
    }

    router.push("/")
}

onMounted(async () => {
    const response = await fetch(`/api/v1/photo/${id}`)
    if (response.status === 404) {
        notFound.value = true;
        return;
    } else if (response.status !== 200) {
        console.error("Got status code:", response.status)
        return
    }

    const jsonData = await response.json() as PhotoMetadataDTO;
    photoMetadata.value = {
        ...jsonData,
        uploaded: new Date(jsonData.uploaded),
        exifTakenAt: jsonData.exifTakenAt ? new Date(jsonData.exifTakenAt) : null,
    };
})
</script>

<template>
    <div v-if="notFound">
        <h1>404 Not Found</h1>
    </div>
    <div v-else>
        <p>Camera Model: {{ photoMetadata.exifCameraModel }}</p>
        <p>Taken At: {{ photoMetadata.exifTakenAt?.toDateString() }}</p>
        <p>Uploaded At: {{ photoMetadata.uploaded?.toDateString() }}</p>
        <p>File Size: {{ formatByteSize(photoMetadata.size) }}</p>
        <button @click="deletePhoto">Delete Photo</button>
        <img :src="photoMetadata.permalink" alt="">
    </div>

</template>