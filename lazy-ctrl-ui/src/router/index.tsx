import {
  createRootRoute,
  createRoute,
  createRouter,
  RouterProvider,
  Outlet,
} from '@tanstack/react-router'
import { lazyRouteComponent } from '@tanstack/react-router'

const Root = () => <Outlet />

const rootRoute = createRootRoute({ component: Root })

const routeTree = rootRoute.addChildren([
  createRoute({ 
    path: '/', 
    getParentRoute: () => rootRoute,
    component: lazyRouteComponent(() => import('@/pages/Home')) }),
])

export const router = createRouter({ routeTree })

export function RouterApp() {
  return <RouterProvider router={router} />
}
