<script setup lang="ts">
import { reactive } from 'vue';
import { router } from '../router';
import type { User } from '../models/models';
import { PhEnvelope, PhLock, PhUser } from '@phosphor-icons/vue';

type CreateAccountForm = Pick<User, 'username' | 'email'> & { password: string }

const form = reactive<CreateAccountForm>({
    username: "",
    email: "",
    password: ""
})

async function handleSubmit(event: SubmitEvent) {
    event.preventDefault()

    const formData = new FormData();
    formData.append("username", form.username)
    formData.append("email", form.email)
    formData.append("password", form.password)

    const response = await fetch("/api/v1/create-account", {
        method: "POST",
        body: formData,
    })

    if (response.status == 201) {
        router.push('/login')
    }
}

</script>

<template>
    <div class="min-h-screen flex items-center justify-center bg-violet-700 p-4">
        <div class="w-full max-w-md flex flex-col gap-6 rounded-2xl shadow-xl p-8 pt-8 bg-white">
            <div>
                <h2 class="text-3xl font-bold text-center text-violet-700">f-stop</h2>
                <p class="mt-2 text-center ">Create your account</p>
            </div>

            <form id="create-account-form" @submit="handleSubmit" class="flex flex-col gap-4">
                <div class="flex flex-col gap-1">
                    <label for="username" class="text-sm font-semibold text-violet-800 flex items-center gap-2">
                        <PhUser :size="18" /> Username
                    </label>
                    <input
                        type="text"
                        v-model="form.username"
                        name="username"
                        id="username"
                        required
                        placeholder="your-username"
                        class="w-full bg-white border border-violet-600 placeholder-violet-300 rounded-lg px-4 py-2"
                    >
                </div>

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
                    value="Create Account"
                    class="mt-2 bg-violet-800 text-white font-bold px-4 py-2 rounded-lg cursor-pointer hover:brightness-95"
                >
            </form>

            <button
                @click="router.push('/login')"
                class="text-center text-sm text-violet-500 hover:text-violet-400 underline underline-offset-4 cursor-pointer"
            >
                Already have an account? Log in
            </button>
        </div>
    </div>
</template>
