import { createRouter, createWebHistory } from 'vue-router'

import HomeView from './views/HomeView.vue'
import LoginView from './views/LoginView.vue'
import CreateAccountView from './views/CreateAccountView.vue'
import PhotoView from './views/PhotoView.vue'

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
    },
    {
        path: '/photos/:id',
        component: PhotoView
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