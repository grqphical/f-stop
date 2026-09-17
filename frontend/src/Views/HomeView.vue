<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { router } from '../router';
import type { User, PhotoMetadata } from '../models/models';

const user = ref<User>({ id: 0, username: "", email: "" });
const error = ref<string | null>(null);
const loading = ref<boolean>(true);

const photosMetadata = ref<PhotoMetadata[]>([])

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
        user.value = await response.json() as User;

        response = await fetch("/api/v1/photo/all", {
            credentials: "same-origin",
            signal: controller.signal
        })
        if (!response.ok) {
            throw new Error(`Server responded with ${response.status}`);
        }
        const jsonData = await response.json() as { photos: PhotoMetadata[] }
        photosMetadata.value = jsonData.photos;

    } catch (err) {
        if (err instanceof Error) {
            console.error("Failed to load user data:", err.message);
            error.value = err.message;
        }
    } finally {
        loading.value = false;
    }
});

async function logoutHandler() {
    await fetch("/api/v1/logout");
    router.push("/login")
}
</script>

<template>
    <p v-if="loading">Loading...</p>
    <p v-else-if="error">Something went wrong: {{ error }}</p>
    <div v-else>
        <h1>Hello, {{ user.username }}!</h1>
        <button @click="logoutHandler">Logout</button>

        <div>
            <img v-for="photoMetadata in photosMetadata" :src="photoMetadata.thumbnailPermalink">
        </div>
    </div>


</template>