import { createRouter, createWebHistory } from "vue-router";
import MainLayout from "../layouts/MainLayout.vue";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/login",
      name: "login",
      component: () => import("../views/Login.vue"),
    },
    {
      path: "/",
      component: MainLayout,
      children: [
        {
          path: "",
          name: "home",
          component: () => import("../views/Home.vue"),
        },
        {
          path: "nodes",
          name: "gameNodes",
          component: () => import("../views/node/index.vue"),
        },
        {
          path: "platforms",
          name: "gamePlatforms",
          component: () => import("../views/platform/index.vue"),
        },
        {
          path: "cards",
          name: "gameCards",
          component: () => import("../views/card/index.vue"),
        },
        {
          path: "instances",
          name: "gameInstances",
          component: () => import("../views/instance/index.vue"),
        },
        {
          path: "nodes/detail/:id",
          name: "gameNodeDetail",
          component: () => import("../views/node/detail.vue"),
          meta: {
            title: "节点详情",
          },
        },
        {
          path: "platforms/detail/:id",
          name: "gamePlatformDetail",
          component: () => import("../views/platform/detail.vue"),
          meta: {
            title: "平台详情",
          },
        },
        {
          path: "/cards/detail/:id",
          name: "GameCardDetail",
          component: () => import("@/views/card/detail.vue"),
          meta: {
            title: "游戏详情",
          },
        },
        {
          path: "instances/detail/:id",
          name: "gameInstanceDetail",
          component: () => import("../views/instance/detail.vue"),
          meta: {
            title: "实例详情",
          },
        },
      ],
    },
  ],
});

// 路由守卫
router.beforeEach((to, from, next) => {
  const isAuthenticated = localStorage.getItem("isAuthenticated");

  if (to.path !== "/login" && !isAuthenticated) {
    next("/login");
  } else {
    next();
  }
});

export default router;
