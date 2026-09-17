<script setup lang="ts">
import { reactive } from 'vue';
import { router } from '../router';


const form = reactive({ email: "", password: "" })

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
    <form id="login-form" @submit="handleLoginSubmit">
        <label for="email">Email:</label>
        <input type="email" v-model="form.email" name="email" id="email" required>

        <label for="password">Password:</label>
        <input type="password" v-model="form.password" name="password" id="password" required>

        <input type="submit" value="Log In">
    </form>
    <button @click="router.push('/create-account')">Create Account</button>
</template>