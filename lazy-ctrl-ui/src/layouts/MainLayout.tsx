import { Outlet, Link, useRouterState } from '@tanstack/react-router'
import { EditModeProvider, useEditMode } from '../contexts/EditModeContext'

const tabs = [
  { path: '/', label: '主页' },
  { path: '/commands', label: '命令' },
  { path: '/settings', label: '设置' },
]

function TabBar() {
  const { location } = useRouterState()
  const { editMode, toggleEditMode } = useEditMode()

  return (
    <nav className="fixed bottom-0 left-0 right-0 h-14 bg-white border-t flex items-center shadow">
      {editMode ? (
        <div className="flex-1 flex justify-center">
          <button
            onClick={toggleEditMode}
            className="text-blue-500 font-bold text-sm px-4 py-2"
          >
            退出编辑模式
          </button>
        </div>
      ) : (
        <>
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
        </>
      )}
    </nav>
  )
}

export default function MainLayout() {
  return (
    <EditModeProvider>
      <div className="flex flex-col min-h-screen">
        <div className="flex-1 overflow-auto pb-14">
          <Outlet />
        </div>
        <TabBar />
      </div>
    </EditModeProvider>
  )
}
