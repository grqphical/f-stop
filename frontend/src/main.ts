import { createApp } from 'vue'
import App from './App.vue'
import { createRouter, createWebHistory } from 'vue-router'

import HomeView from './Views/HomeView.vue'

const routes = [
    {
        path: '/',
        component: HomeView,
    }
]

export const router = createRouter({
    history: createWebHistory(),
    routes,
})

createApp(App)
    .use(router)
    .mount('#app')
