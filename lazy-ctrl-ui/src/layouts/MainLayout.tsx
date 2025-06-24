import { Outlet, Link, useRouterState } from '@tanstack/react-router'

const tabs = [
  { path: '/', label: '主页' },
  { path: '/commands', label: '命令' },
  { path: '/settings', label: '设置' },
]

export default function MainLayout() {
  const { location } = useRouterState()

  return (
    <div className="flex flex-col h-screen">
      <div className="flex-1 overflow-auto">
        <Outlet />
      </div>

      <nav className="fixed bottom-0 left-0 right-0 h-14 bg-white border-t flex justify-around items-center shadow">
        {tabs.map(tab => {
          const isActive = location.pathname === tab.path
          return (
            <Link
              key={tab.path}
              to={tab.path}
              className={`flex-1 text-center py-2 text-sm ${
                isActive ? 'text-blue-500 font-bold' : 'text-gray-500'
              }`}
            >
              {tab.label}
            </Link>
          )
        })}
      </nav>
    </div>
  )
}
