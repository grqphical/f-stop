<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import type { PhotoMetadata, PhotoMetadataDTO } from '../models/models';
import { router } from '../router';
import { formatByteSize } from '../utils';
import Sidebar from '../components/Sidebar.vue';

const route = useRoute()

const id = route.params.id;
const photoMetadata = ref({} as PhotoMetadata);
const notFound = ref(false);
const loading = ref(true);

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
    try {
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
    } finally {
        loading.value = false;
    }
})
</script>

<template>
    <div class="flex flex-row items-start h-dvh overflow-hidden">
        <Sidebar />
        <main class="p-6 flex-1 min-w-0 h-full overflow-y-auto flex justify-center">
            <div v-if="loading" class="text-gray-500 mt-12">Loading...</div>
            <div v-else-if="notFound" class="text-center mt-12">
                <h1 class="text-3xl font-bold">404 Not Found</h1>
                <p class="text-gray-500 mt-2">This photo does not exist.</p>
                <button @click="router.push('/')"
                    class="mt-4 px-3 py-2 bg-violet-500 text-white rounded-md hover:bg-violet-600 cursor-pointer">
                    Back to Photos
                </button>
            </div>
            <div v-else class="w-full max-w-4xl flex flex-col gap-4 h-full min-h-0">
                <div class="shrink-0">
                    <button @click="router.push('/')"
                        class="text-sm text-violet-600 hover:text-violet-800 cursor-pointer">
                        &larr; Back to Photos
                    </button>
                </div>

                <div
                    class="bg-neutral-900 rounded-lg shadow-md overflow-hidden flex flex-1 min-h-[240px] justify-center items-center">
                    <img :src="photoMetadata.permalink" alt="Photo"
                        class="block h-full max-h-full w-auto max-w-full object-contain" />
                </div>

                <section class="shadow-md rounded-md p-6 bg-white shrink-0">
                    <div class="flex items-start justify-between gap-4 mb-4">
                        <div>
                            <h1 class="text-xl font-bold">
                                {{ photoMetadata.exifTakenAt?.toLocaleDateString("en-us", {
                                    year: 'numeric', month: 'long', day: 'numeric'
                                }) ?? 'Untitled Photo' }}
                            </h1>
                            <p class="text-sm text-gray-500">{{ photoMetadata.mimeType }}</p>
                        </div>
                        <button @click="deletePhoto"
                            class="px-3 py-2 bg-red-500 text-white text-sm rounded-md hover:bg-red-600 cursor-pointer shrink-0">
                            Delete Photo
                        </button>
                    </div>

                    <dl class="grid grid-cols-1 sm:grid-cols-2 gap-x-8 gap-y-3 text-sm">
                        <div class="flex justify-between border-b border-neutral-100 pb-2">
                            <dt class="text-gray-500">Camera</dt>
                            <dd class="font-medium">{{ photoMetadata.exifCameraModel ?? 'Unknown' }}</dd>
                        </div>
                        <div class="flex justify-between border-b border-neutral-100 pb-2">
                            <dt class="text-gray-500">File Size</dt>
                            <dd class="font-medium">{{ formatByteSize(photoMetadata.size) }}</dd>
                        </div>
                        <div class="flex justify-between border-b border-neutral-100 pb-2">
                            <dt class="text-gray-500">Taken At</dt>
                            <dd class="font-medium">{{ photoMetadata.exifTakenAt?.toLocaleString() ?? 'Unknown' }}
                            </dd>
                        </div>
                        <div class="flex justify-between border-b border-neutral-100 pb-2">
                            <dt class="text-gray-500">Uploaded At</dt>
                            <dd class="font-medium">{{ photoMetadata.uploaded?.toLocaleString() }}</dd>
                        </div>
                        <div v-if="photoMetadata.exifCoordinates?.latitude != null && photoMetadata.exifCoordinates?.longitude != null"
                            class="flex justify-between border-b border-neutral-100 pb-2 sm:col-span-2">
                            <dt class="text-gray-500">Location</dt>
                            <dd class="font-medium">{{ photoMetadata.exifCoordinates.latitude }}, {{
                                photoMetadata.exifCoordinates.longitude }}</dd>
                        </div>
                    </dl>
                </section>
            </div>
        </main>
    </div>
</template>