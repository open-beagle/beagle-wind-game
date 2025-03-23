import { createRouter, createWebHistory } from 'vue-router'
import MainLayout from '../layouts/MainLayout.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/Login.vue')
    },
    {
      path: '/',
      component: MainLayout,
      children: [
        {
          path: '',
          name: 'home',
          component: () => import('../views/Home.vue')
        },
        {
          path: 'nodes',
          name: 'nodes',
          component: () => import('../views/Nodes.vue')
        },
        {
          path: 'platforms',
          name: 'platforms',
          component: () => import('../views/Platforms.vue')
        },
        {
          path: 'cards',
          name: 'cards',
          component: () => import('../views/Cards.vue')
        },
        {
          path: 'instances',
          name: 'instances',
          component: () => import('../views/Instances.vue')
        }
      ]
    }
  ]
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const isAuthenticated = localStorage.getItem('isAuthenticated')
  
  if (to.path !== '/login' && !isAuthenticated) {
    next('/login')
  } else {
    next()
  }
})

export default router 