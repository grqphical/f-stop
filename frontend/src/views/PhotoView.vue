<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import type { PhotoMetadata, PhotoMetadataDTO, Tag } from '../models/models';
import { router } from '../router';
import { formatByteSize, getImageExtension } from '../utils';
import Sidebar from '../components/Sidebar.vue';
import PhotoTags from '../components/PhotoTags.vue';
import TagAssignDialog from '../components/TagAssignDialog.vue';
import { PhPlus } from '@phosphor-icons/vue';

const route = useRoute()

const id = route.params.id;
const photoMetadata = ref({} as PhotoMetadata);
const notFound = ref(false);
const loading = ref(true);

const activeTagPhoto = ref<PhotoMetadata | null>(null);

function openTagDialog(photo: PhotoMetadata): void {
    activeTagPhoto.value = photo;
}

function closeTagDialog(): void {
    activeTagPhoto.value = null;
}

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

async function downloadPhoto() {
    const a = document.createElement("a");
    a.href = photoMetadata.value.permalink;
    a.download = `image.${getImageExtension(photoMetadata.value.mimeType)}`

    document.body.appendChild(a);

    a.click();
    document.body.removeChild(a);
}

onMounted(async () => {
    try {
        let response = await fetch(`/api/v1/photo/${id}`)
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

        response = await fetch(`/api/v1/photo/${photoMetadata.value.id}/tags`, {
            credentials: "same-origin",
        })
        if (!response.ok) {
            throw new Error(`Server responded with ${response.status}`);
        }
        const tagsResponse = await response.json() as { tags: Tag[] };
        photoMetadata.value = { ...photoMetadata.value, tags: tagsResponse.tags };

    } finally {
        loading.value = false;
    }
})

function handleTagsAssigned(photoId: string, tags: Tag[]): void {
    if (photoMetadata.value?.id !== photoId) return;
    photoMetadata.value.tags = [...(photoMetadata.value.tags ?? []), ...tags];
}

function handleTagRemoved(photoId: string, tagId: number): void {
    if (photoMetadata.value?.id !== photoId) return;
    photoMetadata.value.tags = (photoMetadata.value.tags ?? []).filter((t) => t.id !== tagId);
}
</script>

<template>
    <div class="flex flex-row items-start h-dvh overflow-hidden">
        <Sidebar />
        <main class="p-6 flex-1 min-w-0 h-full overflow-y-auto flex justify-center bg-slate-100">
            <div v-if="loading" class="text-slate-500 mt-12">Loading...</div>
            <div v-else-if="notFound" class="text-center mt-12">
                <h1 class="text-3xl font-bold text-slate-900">404 Not Found</h1>
                <p class="text-slate-500 mt-2">This photo does not exist.</p>
                <button @click="router.push('/')"
                    class="mt-4 px-3 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-md cursor-pointer">
                    Back to Photos
                </button>
            </div>
            <div v-else class="w-full max-w-4xl flex flex-col gap-4 h-full min-h-0">
                <div class="shrink-0">
                    <button @click="router.push('/')"
                        class="text-sm text-indigo-600 hover:text-indigo-800 cursor-pointer">
                        &larr; Back to Photos
                    </button>
                </div>

                <div
                    class="bg-slate-950 rounded-lg shadow-md overflow-hidden flex flex-1 min-h-[240px] justify-center items-center">
                    <img :src="photoMetadata.permalink" alt="Photo"
                        class="block h-full max-h-full w-auto max-w-full object-cover" />
                </div>

                <section class="shadow-md rounded-md p-6 bg-white shrink-0">
                    <div class="flex items-start justify-between gap-4 mb-4">
                        <div>
                            <h1 class="text-xl font-bold text-slate-900">
                                {{ photoMetadata.exifTakenAt?.toLocaleDateString("en-us", {
                                    year: 'numeric', month: 'long', day: 'numeric'
                                }) ?? 'Untitled Photo' }}
                            </h1>
                            <div class="flex flex-row gap-1">
                                <PhotoTags :photo-id="photoMetadata.id" :tags="photoMetadata.tags ?? []"
                                    @removed="handleTagRemoved" />
                                <button @click="openTagDialog(photoMetadata)"
                                    class="bg-gray-300 rounded-md px-1 py-1 aspect-square cursor-pointer hover:bg-gray-400"
                                    aria-label="Add tag" title="Add tag">
                                    <PhPlus size="12px" weight="bold" />
                                </button>
                            </div>
                            <p class="text-sm text-slate-500">{{ photoMetadata.mimeType }}</p>

                        </div>
                        <div class="flex flex-row gap-1">
                            <button @click="downloadPhoto"
                                class="px-3 py-2 bg-indigo-600 hover:bg-indigo-500 text-white text-sm rounded-md cursor-pointer shrink-0">
                                Download Photo
                            </button>
                            <button @click="deletePhoto"
                                class="px-3 py-2 bg-red-600 hover:bg-red-500 text-white text-sm rounded-md cursor-pointer shrink-0">
                                Delete Photo
                            </button>
                        </div>
                    </div>

                    <dl class="grid grid-cols-1 sm:grid-cols-2 gap-x-8 gap-y-3 text-sm">
                        <div class="flex justify-between border-b border-slate-100 pb-2">
                            <dt class="text-slate-500">Camera</dt>
                            <dd class="font-medium text-slate-900">{{ photoMetadata.exifCameraModel ?? 'Unknown' }}</dd>
                        </div>
                        <div class="flex justify-between border-b border-slate-100 pb-2">
                            <dt class="text-slate-500">File Size</dt>
                            <dd class="font-medium text-slate-900">{{ formatByteSize(photoMetadata.size) }}</dd>
                        </div>
                        <div class="flex justify-between border-b border-slate-100 pb-2">
                            <dt class="text-slate-500">Taken At</dt>
                            <dd class="font-medium text-slate-900">{{ photoMetadata.exifTakenAt?.toLocaleString() ??
                                'Unknown' }}
                            </dd>
                        </div>
                        <div class="flex justify-between border-b border-slate-100 pb-2">
                            <dt class="text-slate-500">Uploaded At</dt>
                            <dd class="font-medium text-slate-900">{{ photoMetadata.uploaded?.toLocaleString() }}</dd>
                        </div>
                        <div v-if="photoMetadata.exifCoordinates?.latitude != null && photoMetadata.exifCoordinates?.longitude != null"
                            class="flex justify-between border-b border-slate-100 pb-2 sm:col-span-2">
                            <dt class="text-slate-500">Location</dt>
                            <dd class="font-medium text-slate-900">{{ photoMetadata.exifCoordinates.latitude }}, {{
                                photoMetadata.exifCoordinates.longitude }}</dd>
                        </div>
                    </dl>
                </section>
            </div>
        </main>
    </div>
    <TagAssignDialog :photo="activeTagPhoto" @close="closeTagDialog" @assigned="handleTagsAssigned" />
</template>