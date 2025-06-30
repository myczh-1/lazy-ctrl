import { useState, useEffect } from 'react'
import { useCommandStore } from '@/stores/commandStore'
import { useAppStore } from '@/stores/appStore'
import { CommandService } from '@/services/commandService'

interface SettingsModalProps {
  isOpen: boolean
  onClose: () => void
}

export function SettingsModal({ isOpen, onClose }: SettingsModalProps) {
  const { commands } = useCommandStore()
  const { apiBaseUrl, setApiBaseUrl, saveSettings } = useAppStore()
  const [pinValue, setPinValue] = useState('')
  const [showPin, setShowPin] = useState(false)

  // 从localStorage加载设置
  useEffect(() => {
    const savedPin = localStorage.getItem('lazy-ctrl-pin')
    if (savedPin) setPinValue(savedPin)
  }, [])

  const handleSave = () => {
    // 设置PIN
    if (pinValue) {
      CommandService.setPin(pinValue)
    }
    
    // 保存应用设置
    saveSettings()
    
    // 重新加载命令
    CommandService.fetchCommands()
    
    onClose()
  }

  const handleTestConnection = async () => {
    try {
      await CommandService.fetchCommands()
      alert('连接成功！')
    } catch (error) {
      alert(`连接失败: ${error instanceof Error ? error.message : '未知错误'}`)
    }
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-md w-full max-h-[90vh] overflow-y-auto">
        <div className="p-6">
          <div className="flex justify-between items-center mb-6">
            <h2 className="text-xl font-semibold">设置</h2>
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-600"
            >
              ✕
            </button>
          </div>

          <div className="space-y-6">
            {/* API URL 设置 */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                API 地址
              </label>
              <input
                type="text"
                value={apiBaseUrl}
                onChange={(e) => setApiBaseUrl(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="http://localhost:7070"
              />
              <button
                onClick={handleTestConnection}
                className="mt-2 text-sm text-blue-600 hover:text-blue-800"
              >
                测试连接
              </button>
            </div>

            {/* PIN 设置 */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                PIN 码 (可选)
              </label>
              <div className="relative">
                <input
                  type={showPin ? "text" : "password"}
                  value={pinValue}
                  onChange={(e) => setPinValue(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="输入PIN码"
                />
                <button
                  type="button"
                  onClick={() => setShowPin(!showPin)}
                  className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600"
                >
                  {showPin ? '🙈' : '👁️'}
                </button>
              </div>
              <p className="text-xs text-gray-500 mt-1">
                某些命令可能需要PIN验证
              </p>
            </div>

            {/* 状态信息 */}
            <div className="bg-gray-50 p-4 rounded-md">
              <h3 className="text-sm font-medium text-gray-700 mb-2">连接状态</h3>
              <div className="space-y-1 text-sm text-gray-600">
                <div>可用命令: {commands.filter(cmd => cmd.available).length} / {commands.length}</div>
                <div>API版本: 2.0</div>
              </div>
            </div>

            {/* 命令统计 */}
            {commands.length > 0 && (
              <div className="bg-gray-50 p-4 rounded-md">
                <h3 className="text-sm font-medium text-gray-700 mb-2">命令分类</h3>
                <div className="space-y-1 text-sm text-gray-600">
                  {Object.entries(
                    commands.reduce((acc, cmd) => {
                      const category = cmd.category || '其他'
                      acc[category] = (acc[category] || 0) + 1
                      return acc
                    }, {} as Record<string, number>)
                  ).map(([category, count]) => (
                    <div key={category} className="flex justify-between">
                      <span>{category}</span>
                      <span>{count}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>

          <div className="flex gap-3 mt-8">
            <button
              onClick={handleSave}
              className="flex-1 bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 transition-colors"
            >
              保存设置
            </button>
            <button
              onClick={onClose}
              className="flex-1 bg-gray-300 text-gray-700 py-2 px-4 rounded-md hover:bg-gray-400 transition-colors"
            >
              取消
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}