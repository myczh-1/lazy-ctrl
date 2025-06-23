import {
  createRootRoute,
  createRoute,
  createRouter,
  RouterProvider,
  Outlet,
} from '@tanstack/react-router'
import { lazy } from 'react'

const Root = () => <Outlet />

const rootRoute = createRootRoute({ component: Root })

const routeTree = rootRoute.addChildren([
  createRoute({ path: '/', component: lazy(() => import('@/pages/LoginIntro')) }),
  createRoute({ path: '/login', component: lazy(() => import('@/pages/LoginPassword')) }),
  createRoute({ path: '/commands', component: lazy(() => import('@/pages/CommandList')) }),
])

export const router = createRouter({ routeTree })

export function RouterApp() {
  return <RouterProvider router={router} />
}
