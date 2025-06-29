// src/router/index.tsx
import {
  createRootRoute,
  createRoute,
  createRouter,
  RouterProvider,
  Outlet,
  lazyRouteComponent,
} from '@tanstack/react-router'
import MainLayout from '@/layouts/MainLayout'

// 根路由（只包一层 Outlet）
const rootRoute = createRootRoute({
  component: () => <Outlet />,
})

// Layout 路由，负责渲染 tab 栏
const layoutRoute = createRoute({
  getParentRoute: () => rootRoute,
  component: () => <MainLayout />,
  id: 'layout',
})

// 子页面：Home
const homeRoute = createRoute({
  path: '/',
  getParentRoute: () => layoutRoute,
  component: lazyRouteComponent(() => import('@/pages/Home')),
})

// 子页面：Commands
const commandRoute = createRoute({
  path: '/commands',
  getParentRoute: () => layoutRoute,
  component: lazyRouteComponent(() => import('@/pages/Commands')),
})

// // 子页面：About
// const aboutRoute = createRoute({
//   path: '/about',
//   getParentRoute: () => layoutRoute,
//   component: lazyRouteComponent(() => import('@/pages/About')),
// })

const routeTree = rootRoute.addChildren([
  layoutRoute.addChildren([
    homeRoute,
    commandRoute,
    // aboutRoute,
  ]),
])

export const router = createRouter({ routeTree })

export function RouterApp() {
  return <RouterProvider router={router} />
}
