import { createRouter, createWebHistory } from 'vue-router'

import HomeView from './Views/HomeView.vue'
import LoginView from './Views/LoginView.vue'
import CreateAccountView from './Views/CreateAccountView.vue'

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
    },
    {
        path: '/create-account',
        component: CreateAccountView
    }
]

export const router = createRouter({
    history: createWebHistory(),
    routes,
})

router.beforeEach(async (to, _from) => {
    if (to.path !== '/login' && to.path !== '/create-account' && !await checkUserAuthentication()) {
        return '/login'
    }
})