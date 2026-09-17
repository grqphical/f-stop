import { createRouter, createWebHistory } from 'vue-router'

import HomeView from './Views/HomeView.vue'
import LoginView from './Views/LoginView.vue'

async function checkUserAuthentication(): Promise<boolean> {
    const response = await fetch("/api/v1/user")
    return response.status != 401

}

const routes = [
    {
        path: '/',
        component: HomeView,
    },
    {
        path: '/login',
        component: LoginView,
    }
]

export const router = createRouter({
    history: createWebHistory(),
    routes,
})

router.beforeEach(async (to, from) => {
    if (to.path !== '/login' && !await checkUserAuthentication()) {
        return '/login'
    }
})