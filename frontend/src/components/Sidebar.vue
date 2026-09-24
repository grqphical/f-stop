<script setup lang="ts">
import { onMounted, ref } from 'vue';
import type { User } from '../models/models';
import { router } from '../router';
import { PhImagesSquare, PhTag, PhUser } from '@phosphor-icons/vue';

const user = ref<User>({ id: 0, username: "", email: "" });
const error = ref<string | null>(null);
const loading = ref<boolean>(true);

async function logoutHandler() {
    await fetch("/api/v1/logout");
    router.push("/login")
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
        user.value = await response.json() as User;

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
    <aside class="p-4 pt-8 w-1/6 shrink-0 self-start sticky top-0 flex flex-col h-dvh bg-indigo-950 text-indigo-100">
        <div class="mb-24">
            <h2 class="text-3xl font-bold text-center">f-stop</h2>
        </div>
        <div>
            <ul>
                <li class="py-2 px-3 text-2xl rounded-lg bg-indigo-900 text-white cursor-pointer flex flex-row gap-2"><PhImagesSquare :size="32" /> Photos</li>
                <li class="py-2 px-3 text-2xl rounded-lg cursor-pointer flex flex-row gap-2 hover:bg-white/5 text-indigo-200"><PhTag :size="32"/>  Tags</li>
            </ul>
        </div>
        <div class="mt-auto flex flex-col gap-4">
            <p class="font-bold flex flex-row gap-2 items-center"><PhUser :size="32"/> {{ user.username }}</p>
            <button @click="logoutHandler" class="bg-white/10 hover:bg-white/20 px-4 py-2 rounded-lg cursor-pointer">Logout</button>
        </div>
    </aside>
</template>