<script setup lang="ts">
import { reactive } from 'vue';
import { router } from '../router';
import { PhEnvelope, PhLock } from '@phosphor-icons/vue';

interface LoginForm {
  email: string
  password: string
}

const form = reactive<LoginForm>({ email: "", password: "" })

async function handleLoginSubmit(event: SubmitEvent) {
    event.preventDefault()

    const data = new FormData()
    data.append("email", form.email)
    data.append("password", form.password)

    const response = await fetch("/api/v1/login", {
        method: "POST",
        body: data,
    })

    if (response.status == 200) {
        router.push('/')
    }
}

</script>

<template>
    <div class="min-h-screen flex items-center justify-center bg-violet-700 p-">
        <div class="w-full max-w-md flex flex-col gap-6 rounded-2xl shadow-xl p-8 pt-8 bg-white">
            <div>
                <h2 class="text-3xl font-bold text-center text-violet-700">f-stop</h2>
                <p class="mt-2 text-center ">Log in to your account</p>
            </div>

            <form id="login-form" @submit="handleLoginSubmit" class="flex flex-col gap-4">
                <div class="flex flex-col gap-1">
                    <label for="email" class="text-sm font-semibold text-violet-800 flex items-center gap-2">
                        <PhEnvelope :size="18" /> Email
                    </label>
                    <input
                        type="email"
                        v-model="form.email"
                        name="email"
                        id="email"
                        required
                        placeholder="you@example.com"
                        class="w-full bg-white border border-violet-600 placeholder-violet-300 rounded-lg px-4 py-2"
                    >
                </div>

                <div class="flex flex-col gap-1">
                    <label for="password" class="text-sm font-semibold text-violet-800 flex items-center gap-2">
                        <PhLock :size="18" /> Password
                    </label>
                    <input
                        type="password"
                        v-model="form.password"
                        name="password"
                        id="password"
                        required
                        placeholder="••••••••"
                        class="w-full bg-white border border-violet-600 placeholder-violet-300 rounded-lg px-4 py-2"
                    >
                </div>

                <input
                    type="submit"
                    value="Log In"
                    class="mt-2 bg-violet-800 text-white font-bold px-4 py-2 rounded-lg cursor-pointer hover:brightness-95"
                >
            </form>

            <button
                @click="router.push('/create-account')"
                class="text-center text-sm text-violet-500 hover:text-violet-400 underline underline-offset-4 cursor-pointer"
            >
                Need an account? Create one
            </button>
        </div>
    </div>
</template>
