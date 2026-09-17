<script setup lang="ts">
import { reactive } from 'vue';
import { router } from '../router';

const form = reactive({
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
    <form id="login-form" @submit="handleSubmit">
        <label for="username">Username:</label>
        <input type="text" v-model="form.username" name="username" id="username" required>

        <label for="email">Email:</label>
        <input type="email" v-model="form.email" name="email" id="email" required>

        <label for="password">Password:</label>
        <input type="password" v-model="form.password" name="password" id="password" required>

        <input type="submit" value="Create Account">
    </form>
    <button @click="router.push('/login')">Log In To An Existing Account</button>
</template>