<script setup lang="ts">
import { reactive, ref } from 'vue';
import { router } from '../router';
import { PhEnvelope, PhLock } from '@phosphor-icons/vue';

interface LoginForm {
    email: string
    password: string
}

const form = reactive<LoginForm>({ email: "", password: "" })
const loginError = ref(false);

async function handleLoginSubmit(event: SubmitEvent) {
    event.preventDefault()
    loginError.value = false;

    const data = new FormData()
    data.append("email", form.email)
    data.append("password", form.password)

    const response = await fetch("/api/v1/login", {
        method: "POST",
        body: data,
    })

    if (response.status == 200) {
        router.push('/')
    } else {
        loginError.value = true;
    }
}

</script>

<template>
    <div
        class="min-h-screen flex items-center justify-center bg-gradient-to-br from-indigo-950 via-indigo-900 to-indigo-700 p-4">
        <div class="w-full max-w-md flex flex-col gap-6 rounded-2xl shadow-xl p-8 pt-8 bg-white">
            <div>
                <h2 class="text-3xl font-bold text-center text-indigo-950">f-stop</h2>
                <p class="mt-2 text-center text-slate-500">Log in to your account</p>
            </div>

            <form id="login-form" @submit="handleLoginSubmit" class="flex flex-col gap-4">
                <div class="flex flex-col gap-1">
                    <label for="email" class="text-sm font-semibold text-slate-700 flex items-center gap-2">
                        <PhEnvelope :size="18" /> Email
                    </label>
                    <input type="email" v-model="form.email" :aria-invalid="loginError" name="email" id="email" required
                        placeholder="you@example.com" class="w-full bg-white border border-slate-300 placeholder-slate-400 rounded-lg px-4 py-2 
                        focus:border-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-200
                        aria-invalid:border-red-500 focus:aria-invalid:ring-red-600">
                </div>

                <div class="flex flex-col gap-1">
                    <label for="password" class="text-sm font-semibold text-slate-700 flex items-center gap-2">
                        <PhLock :size="18" /> Password
                    </label>
                    <input type="password" :aria-invalid="loginError" v-model="form.password" name="password"
                        id="password" required placeholder="••••••••" class="w-full bg-white border 
                        border-slate-300 placeholder-slate-400 rounded-lg px-4 py-2 focus:border-indigo-500 
                        focus:outline-none focus:ring-2 focus:ring-indigo-200
                        aria-invalid:border-red-500  focus:aria-invalid:ring-red-600 peer">
                    <p class="hidden group-data-[invalid=true]:block text-sm text-red-600 peer-aria-invalid:block">
                        Incorrect email and/or password</p>
                </div>

                <input type="submit" value="Log In"
                    class="mt-2 bg-indigo-600 hover:bg-indigo-500 text-white font-bold px-4 py-2 rounded-lg cursor-pointer">
            </form>

            <button @click="router.push('/create-account')"
                class="text-center text-sm text-indigo-600 hover:text-indigo-500 underline underline-offset-4 cursor-pointer">
                Need an account? Create one
            </button>
        </div>
    </div>
</template>
