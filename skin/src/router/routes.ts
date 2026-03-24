import type { RouteRecordRaw } from "vue-router";

const routes: RouteRecordRaw[] = [
    {
      path: "/",
      component: () => import("@/layout/Layout.vue"),
    },
  ]
export default routes