<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { router } from '../router';

const user = ref({ username: "" });
const error = ref<string | null>(null);
const loading = ref(true);

onMounted(async () => {
    const controller = new AbortController();
    try {
        const response = await fetch("/api/v1/user", {
            credentials: "same-origin", // include if you rely on session cookies
            signal: controller.signal,
        });

        if (!response.ok) {
            throw new Error(`Server responded with ${response.status}`);
        }

        const contentType = response.headers.get("content-type");
        if (!contentType?.includes("application/json")) {
            throw new Error("Expected JSON but got something else — check your proxy/route config");
        }

        user.value = await response.json();
    } catch (err) {
        if (err instanceof Error) {
            console.error("Failed to load user:", err.message);
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
    </div>


</template>