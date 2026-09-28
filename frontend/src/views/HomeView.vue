<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref } from 'vue';
import { router } from '../router';
import type { PhotoMetadata, PhotoMetadataDTO, Job, JobDTO, Tag } from '../models/models';
import Sidebar from '../components/Sidebar.vue';
import PhotoTags from '../components/PhotoTags.vue';
import TagAssignDialog from '../components/TagAssignDialog.vue';
import { formatByteSize } from '../utils.ts';
import { PhPlus } from '@phosphor-icons/vue';

const error = ref<string | null>(null);
const loading = ref<boolean>(true);

const photosMetadata = ref<PhotoMetadata[]>([])
const activeTagPhoto = ref<PhotoMetadata | null>(null);

function openTagDialog(photo: PhotoMetadata): void {
    activeTagPhoto.value = photo;
}

function closeTagDialog(): void {
    activeTagPhoto.value = null;
}

function handleTagsAssigned(photoId: string, tags: Tag[]): void {
    const photo = photosMetadata.value.find((p) => p.id === photoId);
    if (photo) {
        photo.tags = [...(photo.tags ?? []), ...tags];
    }
}

function handleTagRemoved(photoId: string, tagId: number): void {
    const photo = photosMetadata.value.find((p) => p.id === photoId);
    if (photo) {
        photo.tags = (photo.tags ?? []).filter((t) => t.id !== tagId);
    }
}

// --- thumbnail job polling state (hook for future spinner/notification UI) ---
// pendingThumbnailJobId holds the active job id while polling; isPollingThumbnail
// can be bound directly to a spinner/notification component.
const pendingThumbnailJobId = ref<number | null>(null);
const isPollingThumbnail = ref<boolean>(false);
const controller = new AbortController();

onBeforeUnmount(() => controller.abort());

async function fetchPhotos(signal?: AbortSignal): Promise<void> {
    const response = await fetch("/api/v1/photo/all", {
        credentials: "same-origin",
        signal,
    });
    if (!response.ok) {
        throw new Error(`Server responded with ${response.status}`);
    }
    const jsonData = await response.json() as { photos: PhotoMetadataDTO[] };
    const photos = jsonData.photos.map((photo): PhotoMetadata => ({
        ...photo,
        uploaded: new Date(photo.uploaded),
        exifTakenAt: photo.exifTakenAt ? new Date(photo.exifTakenAt) : null,
    })).sort((a, b) => {
        const aTime = a.exifTakenAt?.getTime() ?? Number.NEGATIVE_INFINITY;
        const bTime = b.exifTakenAt?.getTime() ?? Number.NEGATIVE_INFINITY;
        return bTime - aTime;
    });

    const photosWithTags = await Promise.all(photos.map(async (photo) => {
        const response = await fetch(`/api/v1/photo/${photo.id}/tags`, {
            credentials: "same-origin",
            signal,
        })
        if (!response.ok) {
            throw new Error(`Server responded with ${response.status}`);
        }
        const tagsResponse = await response.json() as { tags: Tag[] };
        return { ...photo, tags: tagsResponse.tags };
    }));

    photosMetadata.value = photosWithTags;
}

async function pollThumbnailJob(jobId: number, signal?: AbortSignal): Promise<Job> {
    pendingThumbnailJobId.value = jobId;
    isPollingThumbnail.value = true;
    try {
        while (true) {
            if (signal?.aborted) throw new DOMException("Aborted", "AbortError");

            const response = await fetch(`/api/v1/jobs/${jobId}`, {
                credentials: "same-origin",
                signal,
            });
            if (!response.ok) {
                throw new Error(`Job poll failed with status ${response.status}`);
            }
            const jobDTO = await response.json() as JobDTO;
            const job: Job = {
                ...jobDTO,
                createdAt: new Date(jobDTO.createdAt),
                updatedAt: new Date(jobDTO.updatedAt),
            };

            if (job.status === "done") {
                return job;
            }
            if (job.status === "failed") {
                throw new Error(`Thumbnail job ${jobId} failed`);
            }

            // pending / in_progress -> wait before next poll
            await new Promise<void>((resolve, reject) => {
                const timeout = setTimeout(resolve, 500);
                signal?.addEventListener("abort", () => {
                    clearTimeout(timeout);
                    reject(new DOMException("Aborted", "AbortError"));
                }, { once: true });
            });
        }
    } finally {
        isPollingThumbnail.value = false;
        pendingThumbnailJobId.value = null;
    }
}

async function uploadPhotoHandler() {
    const input = document.createElement("input");
    input.type = "file";
    input.accept = "image/*";
    input.onchange = async () => {
        const file = input.files?.[0];
        if (!file) return;

        const formData = new FormData();
        formData.append("file", file);

        try {
            const response = await fetch("/api/v1/photo", {
                method: "PUT",
                body: formData,
                credentials: "same-origin",
            });

            if (!response.ok) {
                throw new Error(`Upload failed with status ${response.status}`);
            }

            const result = await response.json() as { photoId: string; thumbnailJobId: number };
            const thumbnailJobId = result.thumbnailJobId;

            if (thumbnailJobId != null) {
                await pollThumbnailJob(thumbnailJobId);
            }

            await fetchPhotos();
        } catch (err) {
            if (err instanceof Error) {
                console.error("Failed to upload photo:", err.message);
                error.value = err.message;
            }
        }
    };
    input.click();
}

onMounted(async () => {
    try {
        let response = await fetch("/api/v1/user", {
            credentials: "same-origin", // include if you rely on session cookies
            signal: controller.signal,
        });

        if (!response.ok) {
            throw new Error(`Server responded with ${response.status}`);
        }

        await fetchPhotos(controller.signal);

    } catch (err) {
        if (err instanceof Error) {
            console.error("Failed to load user data:", err.message);
            error.value = err.message;
        }
    } finally {
        loading.value = false;
    }
});
</script>

<template>
    <p v-if="loading" class="text-slate-500">Loading...</p>
    <p v-else-if="error" class="text-red-600">Something went wrong: {{ error }}</p>
    <div v-else class="flex flex-row items-start min-h-screen bg-slate-100">
        <Sidebar />
        <div class="p-4 flex-1 min-w-0">
            <div>
                <button @click="uploadPhotoHandler"
                    class="px-3 py-2 bg-indigo-600 hover:bg-indigo-500 text-white cursor-pointer mb-4 rounded-md ease-in">Upload
                    Photo</button>
            </div>

            <div class="grid grid-cols-3 gap-2">
                <div class="shadow-md p-4 rounded-md bg-white" v-for="photoMetadata in photosMetadata">
                    <img :src="photoMetadata.thumbnailPermalink" @click="router.push(`/photos/${photoMetadata.id}`)"
                        class="cursor-pointer">
                    <p><strong>{{ photoMetadata.exifTakenAt?.toLocaleDateString("en-us", {
                        year: 'numeric',
                        month: 'long',
                        day: 'numeric'
                            }) }}</strong> {{ formatByteSize(photoMetadata.size) }}
                    </p>
                    <div class="flex flex-row gap-1">
                        <PhotoTags :photo-id="photoMetadata.id" :tags="photoMetadata.tags ?? []"
                            @removed="handleTagRemoved" />
                        <button @click="openTagDialog(photoMetadata)"
                            class="bg-gray-300 rounded-md px-1 py-1 aspect-square cursor-pointer hover:bg-gray-400"
                            aria-label="Add tag" title="Add tag">
                            <PhPlus size="12px" weight="bold" />
                        </button>
                    </div>
                </div>


            </div>
        </div>
        <TagAssignDialog :photo="activeTagPhoto" @close="closeTagDialog" @assigned="handleTagsAssigned" />
    </div>



</template>