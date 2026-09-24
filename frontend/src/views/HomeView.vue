<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { router } from '../router';
import type { PhotoMetadata, Job } from '../models/models';
import Sidebar from '../components/Sidebar.vue';

const error = ref<string | null>(null);
const loading = ref<boolean>(true);

const photosMetadata = ref<PhotoMetadata[]>([])

// --- thumbnail job polling state (hook for future spinner/notification UI) ---
// pendingThumbnailJobId holds the active job id while polling; isPollingThumbnail
// can be bound directly to a spinner/notification component.
const pendingThumbnailJobId = ref<number | null>(null);
const isPollingThumbnail = ref<boolean>(false);

async function fetchPhotos(signal?: AbortSignal): Promise<void> {
    const response = await fetch("/api/v1/photo/all", {
        credentials: "same-origin",
        signal,
    });
    if (!response.ok) {
        throw new Error(`Server responded with ${response.status}`);
    }
    const jsonData = await response.json() as { photos: PhotoMetadata[] };
    photosMetadata.value = jsonData.photos;
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
            const job = await response.json() as Job;

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
    const controller = new AbortController();
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
    <p v-if="loading">Loading...</p>
    <p v-else-if="error">Something went wrong: {{ error }}</p>
    <div v-else class="flex flex-row items-start min-h-screen">
        <Sidebar />
        <div class="p-4 flex-1 min-w-0">
            <button @click="uploadPhotoHandler">Upload Photo</button>
            <div>
                <img v-for="photoMetadata in photosMetadata" :src="photoMetadata.thumbnailPermalink"
                    @click="router.push(`/photos/${photoMetadata.id}`)">
            </div>
        </div>
    </div>



</template>