import { useState, useEffect, useRef } from 'react'
import type { Command, DisplayCommand } from '@/types/command'
import { commandTemplates, categoryInfo } from '@/data/commandTemplates'
import type { CommandTemplate } from '@/data/commandTemplates'
import { LayoutService } from '@/services/layoutService'
import platformService from '@/services/platformService'
import ParameterForm from '@/components/ParameterForm'
import commandAPI, { type CreateCommandRequest } from '@/api/commandAPI'
import { getExecutionErrorMessage } from '@/utils/errorHandler'

// 图标组件
const CommandIcon = ({ icon, category }: { icon?: string; category?: string }) => {
  const categoryColor = category && categoryInfo[category as keyof typeof categoryInfo]
    ? categoryInfo[category as keyof typeof categoryInfo].color
    : 'bg-gray-500'

  return (
    <div className={`w-12 h-12 rounded-full ${categoryColor} flex items-center justify-center text-white text-xl shadow-md`}>
      {icon || '📱'}
    </div>
  )
}

// 已配置命令卡片组件
const ConfiguredCommandCard = ({ command, onExecute, onAddToLayout, onDelete, onEditConfig, onDirectEdit }: {
  command: DisplayCommand;
  onExecute: (id: string) => void;
  onAddToLayout: (command: DisplayCommand) => void;
  onDelete: (id: string) => void;
  onEditConfig?: (command: DisplayCommand) => void;
  onDirectEdit?: (command: DisplayCommand) => void;
}) => {
  const [isExecuting, setIsExecuting] = useState(false)
  const [showActions, setShowActions] = useState(false)

  const handleExecute = async () => {
    setIsExecuting(true)
    try {
      await onExecute(command.id)
    } finally {
      setIsExecuting(false)
    }
  }

  const getCurrentPlatformCommand = () => {
    // 从新的数据结构中获取命令
    if (command.platforms && command.platforms['all']) {
      const platformCommand = command.platforms['all']
      if (Array.isArray(platformCommand)) {
        return `多步骤命令 (${platformCommand.length} 步骤)`
      }
      return platformCommand || 'N/A'
    }

    // 如果没有 platforms 数据，尝试从 commands 数组中获取
    if (command.commands && command.commands.length > 0) {
      const cmd = command.commands[0]
      if (Array.isArray(cmd.command)) {
        return `多步骤命令 (${cmd.command.length} 步骤)`
      }
      return cmd.command || 'N/A'
    }

    return 'N/A'
  }

  const [currentPlatform, setCurrentPlatform] = useState<string>('linux')

  useEffect(() => {
    platformService.getCurrentPlatform().then(setCurrentPlatform)
  }, [])

  const isCommandAvailable = () => {
    
    // 检查从后端返回的available状态
    if (command.commands && command.commands.length > 0) {
      const cmd = command.commands[0]
      // 如果后端明确返回了available状态
      if (cmd.hasOwnProperty('available')) {
        return (cmd as any).available
      }
    }
    
    // 检查平台兼容性 - 修复：使用hasOwnProperty而不是值的真假性
    if (command.platforms) {
      return command.platforms.hasOwnProperty(currentPlatform) || command.platforms.hasOwnProperty('all')
    }
    
    if (command.commands && command.commands.length > 0) {
      const cmd = command.commands[0]
      return cmd.platform === 'all' || cmd.platform === currentPlatform
    }
    
    return true // 默认可用
  }

  // 检查命令是否为空
  const isCommandEmpty = () => {
    // 检查 commands 数组中的命令内容
    if (command.commands && command.commands.length > 0) {
      const cmd = command.commands[0]
      return !cmd.command || (typeof cmd.command === 'string' && cmd.command.trim() === '')
    }
    
    // 如果有 platforms 结构，检查当前平台的命令
    if (command.platforms) {
      const currentPlatformCommand = command.platforms[currentPlatform] || command.platforms['all']
      return !currentPlatformCommand || (typeof currentPlatformCommand === 'string' && currentPlatformCommand.trim() === '')
    }
    
    // 如果都没有，说明命令为空
    return true
  }

  // 获取不可用的原因
  const getUnavailableReason = () => {
    if (isCommandEmpty()) {
      return 'empty_command'
    }
    if (!isCommandAvailable()) {
      return 'platform_incompatible'
    }
    return null
  }

  const getPlatformDisplayInfo = () => {
    const available = isCommandAvailable() && !isCommandEmpty()
    const unavailableReason = getUnavailableReason()
    
    let supportedPlatforms: string[] = []
    
    if (command.platforms) {
      supportedPlatforms = Object.keys(command.platforms)
    } else if (command.commands && command.commands.length > 0) {
      supportedPlatforms = [command.commands[0].platform]
    }
    
    return {
      available,
      unavailableReason,
      supportedPlatforms: supportedPlatforms.map(p => platformService.getPlatformDisplayName(p)),
      currentPlatform: platformService.getPlatformDisplayName(currentPlatform)
    }
  }

  const platformInfo = getPlatformDisplayInfo()

  return (
    <div className="bg-white rounded-xl shadow-md border border-gray-200 p-4 hover:shadow-lg transition-all duration-200">
      <div className="flex items-start space-x-4">
        <CommandIcon icon={command.icon} category={command.category} />
        <div className="flex-1 min-w-0">
          <div className="flex items-center justify-between mb-2">
            <h3 className="font-semibold text-gray-900 truncate">{command.name}</h3>
            <div className="flex items-center space-x-2">
              <span className="text-xs px-2 py-1 bg-gray-100 text-gray-600 rounded-full">
                {categoryInfo[command.category as keyof typeof categoryInfo]?.name || command.category || '通用'}
              </span>
              {!platformInfo.available && (
                <span className="text-xs px-2 py-1 bg-red-100 text-red-600 rounded-full">
                  不兼容
                </span>
              )}
              {platformInfo.available && platformInfo.supportedPlatforms.length > 0 && (
                <span className="text-xs px-2 py-1 bg-green-100 text-green-600 rounded-full">
                  {platformInfo.supportedPlatforms.join(', ')}
                </span>
              )}
              <button
                onClick={() => setShowActions(!showActions)}
                className="text-gray-400 hover:text-gray-600 p-1"
              >
                ⋮
              </button>
            </div>
          </div>
          {command.description && (
            <p className="text-sm text-gray-600 mb-3">{command.description}</p>
          )}
          <div className="bg-gray-50 rounded-lg p-2 mb-3">
            <p className="text-xs text-gray-500 mb-1">命令详情:</p>
            <code className="text-xs text-gray-700 font-mono break-all">
              {getCurrentPlatformCommand()}
            </code>
            {!platformInfo.available && (
              <div className="mt-2 p-2 bg-red-50 rounded border border-red-200">
                {platformInfo.unavailableReason === 'empty_command' ? (
                  <>
                    <p className="text-xs text-red-600">
                      ⚠️ 命令内容为空，需要配置
                    </p>
                    <p className="text-xs text-red-500 mt-1">
                      请编辑此命令添加具体的命令内容
                    </p>
                  </>
                ) : platformInfo.unavailableReason === 'platform_incompatible' ? (
                  <>
                    <p className="text-xs text-red-600">
                      ⚠️ 此命令在当前平台 ({platformInfo.currentPlatform}) 不可用
                    </p>
                    <p className="text-xs text-red-500 mt-1">
                      支持平台: {platformInfo.supportedPlatforms.join(', ')}
                    </p>
                  </>
                ) : (
                  <p className="text-xs text-red-600">
                    ⚠️ 此命令当前不可用
                  </p>
                )}
              </div>
            )}
            {command.platforms && Object.keys(command.platforms).length > 1 && (
              <p className="text-xs text-gray-500 mt-1">
                支持平台: {Object.keys(command.platforms).join(', ')}
              </p>
            )}
            {command.commands && command.commands.length > 0 && command.commands[0].platform && (
              <p className="text-xs text-gray-500 mt-1">
                平台: {command.commands[0].platform}
              </p>
            )}
          </div>

          {showActions ? (
            <div className="space-y-2">
              <div className="grid grid-cols-2 gap-2">
               <button
                  onClick={() => setShowActions(false)}
                  className="px-3 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-lg font-medium text-sm transition-all shadow-sm"
                >
                  完成
                </button>
                <button
                  onClick={() => onDelete(command.id)}
                  className="px-3 py-2 bg-red-500 hover:bg-red-600 text-white rounded-lg font-medium text-sm transition-all shadow-sm"
                >
                  删除
                </button>
              </div>
              <div className="space-y-2">
                {onEditConfig && (
                  <button
                    onClick={() => onEditConfig(command)}
                    className="w-full px-3 py-2 bg-purple-500 hover:bg-purple-600 text-white rounded-lg font-medium text-sm transition-all shadow-sm"
                  >
                    ⚙️ 重新配置参数
                  </button>
                )}
                {onDirectEdit && (
                  <button
                    onClick={() => onDirectEdit(command)}
                    className="w-full px-3 py-2 bg-yellow-500 hover:bg-yellow-600 text-white rounded-lg font-medium text-sm transition-all shadow-sm"
                  >
                    📝 编辑信息
                  </button>
                )}
              </div>
            </div>
          ) : (
            <div className="grid grid-cols-2 gap-2">
              <button
                onClick={handleExecute}
                disabled={isExecuting || !platformInfo.available}
                className={`px-4 py-2 rounded-lg font-medium text-sm transition-all ${
                  isExecuting || !platformInfo.available
                    ? 'bg-gray-200 text-gray-500 cursor-not-allowed'
                    : 'bg-blue-500 hover:bg-blue-600 text-white shadow-sm'
                  }`}
              >
                {isExecuting ? '执行中...' : !platformInfo.available ? '不可用' : '执行命令'}
              </button>
              <button
                onClick={() => onAddToLayout(command)}
                className="px-4 py-2 bg-green-500 hover:bg-green-600 text-white rounded-lg font-medium text-sm transition-all shadow-sm"
              >
                添加到主页
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

// 命令模板卡片组件
const TemplateCard = ({ template, onConfigure }: {
  template: CommandTemplate;
  onConfigure: (template: CommandTemplate, mode: 'add' | 'execute' | 'both') => void;
}) => {
  const hasUI = template.ui && template.ui.params.length > 0

  return (
    <div className="bg-gradient-to-br from-gray-50 to-gray-100 rounded-xl border-2 border-dashed border-gray-300 p-4 hover:border-blue-400 hover:from-blue-50 hover:to-blue-100 transition-all duration-200">
      <div className="flex items-start space-x-4">
        <div className="relative">
          <CommandIcon icon={template.icon} category={template.category} />
          {hasUI && (
            <div className="absolute -top-1 -right-1 w-4 h-4 bg-purple-500 rounded-full flex items-center justify-center">
              <span className="text-white text-xs">⚙️</span>
            </div>
          )}
        </div>
        <div className="flex-1 min-w-0">
          <div className="flex items-center justify-between mb-2">
            <h3 className="font-semibold text-gray-700">{template.name}</h3>
            <div className="flex items-center space-x-1">
              {hasUI && (
                <span className="text-xs px-2 py-1 bg-purple-100 text-purple-600 rounded-full">
                  需要配置
                </span>
              )}
              <span className="text-xs px-2 py-1 bg-white bg-opacity-80 text-gray-600 rounded-full">
                模板
              </span>
            </div>
          </div>
          <p className="text-sm text-gray-600 mb-3">{template.description}</p>

          {hasUI ? (
            // 可配置模板：必须先配置
            <button
              onClick={() => onConfigure(template, 'both')}
              className="w-full px-4 py-2 bg-purple-500 hover:bg-purple-600 text-white rounded-lg font-medium text-sm transition-all shadow-sm hover:shadow-md active:scale-95"
            >
              ⚙️ 配置参数
            </button>
          ) : (
            // 普通模板：直接添加
            <button
              onClick={() => onConfigure(template, 'add')}
              className="w-full px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-lg font-medium text-sm transition-all shadow-sm hover:shadow-md active:scale-95"
            >
              + 添加命令
            </button>
          )}
        </div>
      </div>
    </div>
  )
}

// 直接编辑命令模态框
const DirectEditModal = ({ isOpen, onClose, command, onSave }: {
  isOpen: boolean;
  onClose: () => void;
  command: DisplayCommand | null;
  onSave: (updatedCommand: CreateCommandRequest) => void;
}) => {
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    category: 'custom',
    icon: '',
    command: '',
    platform: 'all'
  })
  const [currentPlatform, setCurrentPlatform] = useState<string>('linux')

  // 获取当前平台信息
  useEffect(() => {
    platformService.getCurrentPlatform().then(setCurrentPlatform)
  }, [])

  // 判断是否为模板命令
  const isTemplateCommand = command?.templateId && command.templateId !== 'custom_command_builder'
  const template = commandTemplates.find(t => t.templateId === command?.templateId)

  // 从模板获取指定平台的命令
  const getTemplateCommand = (template: CommandTemplate | undefined, platform: string): string => {
    if (!template) return ''
    
    // 优先使用指定平台的命令，其次使用通用命令
    const platformCommand = template.platforms[platform as keyof typeof template.platforms] || 
                           template.platforms.all
    
    if (typeof platformCommand === 'string') {
      return platformCommand
    } else if (Array.isArray(platformCommand)) {
      return `多步骤命令 (${platformCommand.length} 步骤)`
    }
    
    return ''
  }

  useEffect(() => {
    if (command) {
      const currentPlatform = command.commands?.[0]?.platform || 'all'
      const currentTemplate = commandTemplates.find(t => t.templateId === command.templateId)
      
      // 对于模板命令，从模板获取命令内容；对于自定义命令，使用保存的命令内容
      const commandContent = isTemplateCommand 
        ? getTemplateCommand(currentTemplate, currentPlatform)
        : (typeof command.commands?.[0]?.command === 'string' ? command.commands[0].command : '')
      
      setFormData({
        name: command.name || '',
        description: command.description || '',
        category: command.category || 'custom',
        icon: command.icon || '',
        command: commandContent,
        platform: currentPlatform
      })
    }
  }, [command, isTemplateCommand])

  if (!isOpen || !command) return null

  // 处理平台选择变化
  const handlePlatformChange = (newPlatform: string) => {
    setFormData(prev => {
      const newCommand = isTemplateCommand 
        ? getTemplateCommand(template, newPlatform)
        : prev.command
      
      return {
        ...prev,
        platform: newPlatform,
        command: newCommand
      }
    })
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()

    const updatedCommand: CreateCommandRequest = {
      id: command.id,
      name: formData.name,
      description: formData.description,
      category: formData.category,
      icon: formData.icon,
      command: formData.command,
      platform: formData.platform,
      timeout: 10000,
      security: {
        requirePin: false,
        whitelist: true
      },
      updatedAt: new Date().toISOString()
    }

    onSave(updatedCommand)
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-40">
      <div className="bg-white rounded-xl max-w-2xl w-full max-h-[80vh] overflow-hidden">
        <div className="p-6 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-bold text-gray-900">编辑命令</h2>
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-600 text-xl"
            >
              ✕
            </button>
          </div>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">命令名称</label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              required
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">描述</label>
            <textarea
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              rows={2}
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">分类</label>
              <select
                value={formData.category}
                onChange={(e) => setFormData({ ...formData, category: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              >
                {Object.entries(categoryInfo).map(([key, info]) => (
                  <option key={key} value={key}>{info.name}</option>
                ))}
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">图标</label>
              <input
                type="text"
                value={formData.icon}
                onChange={(e) => setFormData({ ...formData, icon: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                placeholder="💻"
              />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">支持平台</label>
            <select
              value={formData.platform}
              onChange={(e) => handlePlatformChange(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            >
              <option value="all">全平台</option>
              <option value="windows">Windows</option>
              <option value="darwin">macOS</option>
              <option value="linux">Linux</option>
            </select>
            <p className="text-xs text-gray-500 mt-1">
              选择此命令支持的操作系统平台
            </p>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              命令内容
              {isTemplateCommand && (
                <span className="text-xs text-blue-600 ml-2">(模板命令，根据平台自动填充)</span>
              )}
            </label>
            <textarea
              value={formData.command}
              onChange={isTemplateCommand ? undefined : (e) => setFormData({ ...formData, command: e.target.value })}
              readOnly={isTemplateCommand}
              className={`w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono text-sm ${
                isTemplateCommand ? 'bg-gray-50 cursor-not-allowed' : ''
              }`}
              rows={4}
              required={!isTemplateCommand}
              placeholder={isTemplateCommand ? "模板命令将根据选择的平台自动填充" : "输入要执行的命令..."}
            />
            {isTemplateCommand && (
              <p className="text-xs text-gray-500 mt-1">
                此命令来自模板 "{template?.name}"，请选择平台查看对应命令
              </p>
            )}
          </div>

          <div className="flex justify-end space-x-3 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg font-medium transition-all"
            >
              取消
            </button>
            <button
              type="submit"
              className="px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-lg font-medium transition-all"
            >
              保存修改
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

// 添加命令模态框
const AddCommandModal = ({ isOpen, onClose, onConfigure }: {
  isOpen: boolean;
  onClose: () => void;
  onConfigure: (template: CommandTemplate, mode: 'add' | 'execute' | 'both') => void;
}) => {
  if (!isOpen) return null

  const categories = Object.entries(categoryInfo)

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-40">
      <div className="bg-white rounded-xl max-w-4xl w-full max-h-[80vh] overflow-hidden">
        <div className="p-6 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <h2 className="text-xl font-bold text-gray-900">选择命令模板</h2>
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-600 text-xl"
            >
              ✕
            </button>
          </div>
          <p className="text-gray-600 mt-2">
            从预设模板中选择一个命令，带 <span className="inline-flex items-center px-1 py-0.5 bg-purple-100 text-purple-600 rounded text-xs">⚙️ 可配置</span> 标签的支持参数调整
          </p>
        </div>

        <div className="p-6 overflow-y-auto max-h-[60vh]">
          {categories.map(([categoryKey, categoryData]) => {
            const templates = commandTemplates.filter(t => t.category === categoryKey)
            if (templates.length === 0) return null

            return (
              <div key={categoryKey} className="mb-8">
                <h3 className="flex items-center text-lg font-semibold text-gray-800 mb-4">
                  <span className="mr-2">{categoryData.icon}</span>
                  {categoryData.name}
                </h3>
                <div className="grid gap-4 md:grid-cols-2">
                  {templates.map(template => (
                    <TemplateCard
                      key={template.templateId}
                      template={template}
                      onConfigure={(template, mode) => {
                        onConfigure(template, mode)
                        onClose()
                      }}
                    />
                  ))}
                </div>
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}

export default function Commands() {
  const [commands, setCommands] = useState<DisplayCommand[]>([])
  const [loading, setLoading] = useState(true)
  const [searchQuery, setSearchQuery] = useState('')
  const [showAddModal, setShowAddModal] = useState(false)
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null)
  const [configureTemplate, setConfigureTemplate] = useState<CommandTemplate | null>(null)
  const [configureMode, setConfigureMode] = useState<'add' | 'execute' | 'both'>('both')
  const [editingCommand, setEditingCommand] = useState<DisplayCommand | null>(null)
  const [showDirectEdit, setShowDirectEdit] = useState(false)
  const [currentPlatform, setCurrentPlatform] = useState<string>('linux')
  const fetchingRef = useRef(false)

  // 获取当前平台信息
  useEffect(() => {
    platformService.getCurrentPlatform().then(setCurrentPlatform)
  }, [])

  // 获取命令列表
  const fetchCommands = async () => {
    // 防止重复请求（解决 React.StrictMode 双重调用问题）
    if (fetchingRef.current) {
      console.log('Fetch already in progress, skipping duplicate request')
      return
    }

    fetchingRef.current = true
    try {
      console.log('Fetching commands from API...')
      // 优先从后端 API 获取
      const data = await commandAPI.getCommands()
      console.log('API response:', data)
      const parsedCommands = parseCommandsFromAPI(data)
      console.log('Parsed commands:', parsedCommands)
      setCommands(parsedCommands)

      // 同步到本地存储作为备份
      saveCommands(parsedCommands)
    } catch (error) {
      console.error('Failed to fetch commands from API:', error)
      // 后端不可用时，从本地存储获取
      try {
        const savedCommands = localStorage.getItem('lazy-ctrl-commands')
        if (savedCommands) {
          const rawCommands = JSON.parse(savedCommands) as Command[]
          const displayCommands = groupCommandsForDisplay(rawCommands)
          setCommands(displayCommands)
          console.log('Loaded commands from localStorage as fallback')
        } else {
          setCommands([])
          console.log('No commands found in localStorage')
        }
      } catch (localError) {
        console.error('Failed to load from localStorage:', localError)
        setCommands([])
      }
    } finally {
      setLoading(false)
      fetchingRef.current = false
    }
  }

  // 解析从 API 获取的命令配置
  const parseCommandsFromAPI = (apiResponse: any): DisplayCommand[] => {
    console.log('Parsing API response:', apiResponse)

    // 检查响应格式：应该是 { version: string, commands: CommandInfo[] }
    if (!apiResponse || !apiResponse.data || !Array.isArray(apiResponse.data)) {
      console.error('Invalid API response format:', apiResponse)
      return []
    }

    // 直接将后端返回的命令信息转换为 DisplayCommand 格式
    const displayCommands: DisplayCommand[] = apiResponse.data.map((cmdInfo: any) => {
      return {
        id: cmdInfo.id,
        name: cmdInfo.name || formatCommandName(cmdInfo.id),
        description: cmdInfo.description || '',
        icon: cmdInfo.icon || getCategoryIcon(cmdInfo.category || 'custom'),
        category: cmdInfo.category || 'custom',
        templateId: cmdInfo.templateId, // 传递模板ID
        platforms: {
          // 保留后端真实的平台信息，而不是强制设为'all'
          [cmdInfo.platform || 'all']: cmdInfo.command || ''
        },
        commands: [{
          id: cmdInfo.id,
          name: cmdInfo.name || formatCommandName(cmdInfo.id),
          platform: cmdInfo.platform || 'all',
          command: cmdInfo.command || '',
          category: cmdInfo.category || 'custom',
          icon: cmdInfo.icon || getCategoryIcon(cmdInfo.category || 'custom'),
          description: cmdInfo.description || ''
        }]
      }
    })

    console.log('Converted to DisplayCommands:', displayCommands)
    return displayCommands
  }

  // 将命令列表按ID分组以供显示
  const groupCommandsForDisplay = (commands: Command[]): DisplayCommand[] => {
    const grouped = new Map<string, Command[]>()

    commands.forEach(cmd => {
      if (!grouped.has(cmd.id)) {
        grouped.set(cmd.id, [])
      }
      grouped.get(cmd.id)!.push(cmd)
    })

    const displayCommands: DisplayCommand[] = []
    grouped.forEach((cmdList, id) => {
      const firstCmd = cmdList[0]
      const platforms: Record<string, string | any[]> = {}

      cmdList.forEach(cmd => {
        platforms[cmd.platform] = cmd.command
      })

      displayCommands.push({
        id,
        name: firstCmd.name,
        description: firstCmd.description,
        icon: firstCmd.icon,
        category: firstCmd.category,
        templateId: (firstCmd as any).templateId, // 传递模板ID
        platforms,
        commands: cmdList
      })
    })

    return displayCommands
  }

  // 格式化命令名称
  const formatCommandName = (id: string): string => {
    return id
      .split('_')
      .map(word => word.charAt(0).toUpperCase() + word.slice(1))
      .join(' ')
  }


  const getCategoryIcon = (category: string): string => {
    const icons: Record<string, string> = {
      audio: '🔊',
      system: '⚙️',
      power: '⚡',
      media: '🎵',
      application: '📱',
      custom: '🛠️'
    }
    return icons[category] || '📋'
  }

  // 保存命令到本地存储（作为备份）
  const saveCommands = (displayCommands: DisplayCommand[]) => {
    // 将显示命令转换为原始命令列表保存
    const rawCommands: Command[] = []
    displayCommands.forEach(displayCmd => {
      rawCommands.push(...displayCmd.commands)
    })
    localStorage.setItem('lazy-ctrl-commands', JSON.stringify(rawCommands))
  }

  // 添加新命令
  const addCommand = async (template: CommandTemplate, params?: Record<string, any>) => {
    try {
      let processedTemplate = template

      // 如果有参数，需要替换命令中的占位符
      if (params && Object.keys(params).length > 0) {
        processedTemplate = {
          ...template,
          platforms: Object.fromEntries(
            Object.entries(template.platforms).map(([platform, command]) => [
              platform,
              replaceCommandParams(command, params)
            ])
          )
        }
      }

      // 获取当前平台的命令
      const platformCommand = processedTemplate.platforms[currentPlatform] || processedTemplate.platforms.all

      if (!platformCommand || typeof platformCommand !== 'string') {
        throw new Error('当前平台不支持该命令')
      }

      // 生成唯一ID
      const commandId = `${template.templateId}_${Date.now()}`

      // 构建命令请求数据
      const commandRequest: CreateCommandRequest = {
        id: commandId,
        name: template.name + (params ? ` (${Object.values(params).join(', ')})` : ''),
        description: template.description,
        category: template.category,
        icon: template.icon,
        command: platformCommand,
        platform: currentPlatform,
        templateId: template.templateId,
        templateParams: params,
        userId: 'local',
        deviceId: 'default',
        timeout: 10000,
        security: {
          requirePin: false,
          whitelist: true
        },
        homeLayout: {
          showOnHome: false
        },
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      }

      // 保存到后端
      await commandAPI.createCommand(commandRequest)

      // 重新获取命令列表
      await fetchCommands()

      showToast(`命令 "${commandRequest.name}" 已添加`, 'success')
    } catch (error) {
      console.error('Failed to add command:', error)
      showToast(`添加命令失败: ${error}`, 'error')
    }
  }

  // 替换命令中的参数占位符
  const replaceCommandParams = (command: string | any[], params: Record<string, any>): string | any[] => {
    if (Array.isArray(command)) {
      return command.map(step => {
        if (step.cmd) {
          return { ...step, cmd: replaceCommandParams(step.cmd, params) }
        }
        if (step.duration && typeof step.duration === 'string' && step.duration.includes('{{')) {
          const key = step.duration.replace(/[{}]/g, '')
          return { ...step, duration: params[key] || step.duration }
        }
        return step
      })
    }

    let result = command
    Object.entries(params).forEach(([key, value]) => {
      const regex = new RegExp(`{{${key}}}`, 'g')
      result = result.replace(regex, String(value))
    })
    return result
  }

  // 处理模板配置
  const handleTemplateConfiguration = async (template: CommandTemplate, mode: 'add' | 'execute' | 'both') => {
    // 检查模板是否有UI配置
    const hasUI = template.ui && template.ui.params && template.ui.params.length > 0

    if (hasUI) {
      // 有UI配置：显示参数表单
      setConfigureTemplate(template)
      setConfigureMode(mode)
      setEditingCommand(null)
    } else {
      // 没有UI配置：直接添加命令
      if (mode === 'add' || mode === 'both') {
        await addCommand(template)
      }
    }
  }

  // 处理已有命令的重新配置
  const handleCommandEdit = (command: DisplayCommand) => {
    // 尝试从原始模板中找到对应的模板
    const template = commandTemplates.find(t => command.id.startsWith(t.templateId))
    if (template && template.ui) {
      setConfigureTemplate(template)
      setConfigureMode('execute')
      setEditingCommand(command)
    }
  }

  // 直接编辑命令信息
  const handleDirectEdit = (command: DisplayCommand) => {
    setEditingCommand(command)
    setShowDirectEdit(true)
  }

  // 保存直接编辑的命令
  const handleDirectEditSave = async (updatedCommand: CreateCommandRequest) => {
    try {
      await commandAPI.updateCommand(updatedCommand.id, updatedCommand)
      await fetchCommands()
      setShowDirectEdit(false)
      setEditingCommand(null)
      showToast('命令修改成功', 'success')
    } catch (error) {
      console.error('Failed to update command:', error)
      showToast(`修改命令失败: ${error}`, 'error')
    }
  }

  // 处理参数表单提交 - 添加命令
  const handleParameterAddCommand = async (params: Record<string, any>) => {
    if (configureTemplate) {
      await addCommand(configureTemplate, params)
      setConfigureTemplate(null)
      setEditingCommand(null)
    }
  }

  // 处理参数表单提交 - 执行命令
  const handleParameterExecute = async (params: Record<string, any>) => {
    if (configureTemplate) {
      // 构建临时命令并执行
      const processedTemplate = {
        ...configureTemplate,
        platforms: Object.fromEntries(
          Object.entries(configureTemplate.platforms).map(([platform, command]) => [
            platform,
            replaceCommandParams(command, params)
          ])
        )
      }

      // 获取当前平台的命令
      const platformCommand = processedTemplate.platforms[currentPlatform] || processedTemplate.platforms.all

      if (platformCommand && typeof platformCommand === 'string') {
        try {
          const response = await fetch(`/api/v1/execute`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json'
            },
            body: JSON.stringify({ command: platformCommand })
          })
          const result = await response.text()
          console.log('Command result:', result)
          showToast(`命令执行成功`, 'success')
        } catch (error) {
          console.error('Command execution failed:', error)
          showToast(`命令执行失败: ${error}`, 'error')
        }
      }

      setConfigureTemplate(null)
      setEditingCommand(null)
    }
  }

  // 获取当前平台（现在从状态获取）
  const getCurrentPlatform = () => currentPlatform

  // 删除命令
  const deleteCommand = async (commandId: string) => {
    if (confirm('确定要删除这个命令吗？')) {
      try {
        await commandAPI.deleteCommand(commandId)
        await fetchCommands()
        showToast('命令已删除', 'success')
      } catch (error) {
        console.error('Failed to delete command:', error)
        showToast(`删除命令失败: ${error}`, 'error')
      }
    }
  }

  // 添加命令到主页布局
  const addToLayout = async (displayCommand: DisplayCommand) => {
    try {
      // 首先更新后端的homeLayout配置
      const commandRequest: CreateCommandRequest = {
        id: displayCommand.id,
        name: displayCommand.name,
        description: displayCommand.description || '',
        category: displayCommand.category,
        icon: displayCommand.icon,
        command: typeof displayCommand.commands[0]?.command === 'string' ? displayCommand.commands[0].command : '',
        platform: displayCommand.commands[0]?.platform || currentPlatform,
        templateId: (displayCommand.commands[0] as any)?.templateId,
        templateParams: (displayCommand.commands[0] as any)?.templateParams,
        userId: (displayCommand.commands[0] as any)?.userId || 'local',
        deviceId: (displayCommand.commands[0] as any)?.deviceId || 'default',
        timeout: 10000,
        security: {
          requirePin: false,
          whitelist: true
        },
        homeLayout: {
          showOnHome: true,
          defaultPosition: {
            x: 0,
            y: 0,
            w: 2,
            h: 1
          },
          color: '',
          priority: 0
        },
        updatedAt: new Date().toISOString()
      }

      await commandAPI.updateCommand(displayCommand.id, commandRequest)

      // 然后添加到本地布局管理
      const success = await LayoutService.addCommandToLayout(displayCommand.id)
      if (success) {
        await fetchCommands() // 重新获取数据以保持同步
        showToast(`"${displayCommand.name}" 已添加到主页`, 'success')
      } else {
        showToast('添加失败，请确保主页已加载', 'error')
      }
    } catch (error) {
      console.error('Add to layout failed:', error)
      showToast(`添加失败: ${error}`, 'error')
    }
  }

  // 执行命令
  const executeCommand = async (commandId: string) => {
    try {
      const result = await commandAPI.executeCommand(commandId)
      console.log('Command result:', result)

      const command = commands.find(c => c.id === commandId)
      if (result.success) {
        showToast(`命令 "${command?.name || commandId}" 执行成功`, 'success')
      } else {
        // 使用友好的错误提示
        const friendlyErrorMessage = getExecutionErrorMessage(result)
        showToast(`命令执行失败:\n${friendlyErrorMessage}`, 'error')
      }
    } catch (error) {
      console.error('Command execution failed:', error)
      showToast(`命令执行失败: ${error}`, 'error')
    }
  }

  // 显示提示消息
  const showToast = (message: string, type: 'success' | 'error') => {
    const toast = document.createElement('div')
    toast.className = `fixed top-4 right-4 px-4 py-2 rounded-lg shadow-lg z-50 text-white ${type === 'success' ? 'bg-green-500' : 'bg-red-500'
      }`
    // 处理多行文本
    toast.style.whiteSpace = 'pre-line'
    toast.style.maxWidth = '400px'
    toast.textContent = message
    document.body.appendChild(toast)
    // 错误消息显示时间稍长一些
    const displayTime = type === 'error' ? 5000 : 3000
    setTimeout(() => document.body.removeChild(toast), displayTime)
  }

  // 筛选命令
  const filteredCommands = commands.filter(command => {
    const matchesCategory = !selectedCategory || command.category === selectedCategory
    const matchesSearch = !searchQuery ||
      command.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      command.description?.toLowerCase().includes(searchQuery.toLowerCase())
    return matchesCategory && matchesSearch
  })

  // 获取所有分类
  const allCategories = Object.keys(categoryInfo)

  useEffect(() => {
    fetchCommands()
  }, [])

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-4"></div>
          <p className="text-gray-600">加载命令列表...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="max-w-6xl mx-auto p-4">

      {/* 搜索和筛选 */}
      <div className="mb-6 space-y-4">
        <div className="relative flex">
          <input
            type="text"
            placeholder="搜索命令..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-4 pr-4 py-2 border border-gray-300 rounded-l-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          />
          <button
            onClick={() => setShowAddModal(true)}
            className="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-r-lg font-medium shadow-sm hover:shadow-md active:scale-95 transition-all"
          >
            +
          </button>
        </div>

        <div className="flex flex-wrap gap-2">
          <button
            onClick={() => setSelectedCategory(null)}
            className={`px-3 py-1 rounded-full text-sm font-medium transition-all ${selectedCategory === null
                ? 'bg-blue-500 text-white shadow-md'
                : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
              }`}
          >
            全部
          </button>
          {allCategories.map(category => (
            <button
              key={category}
              onClick={() => setSelectedCategory(category)}
              className={`px-3 py-1 rounded-full text-sm font-medium transition-all ${selectedCategory === category
                  ? 'bg-blue-500 text-white shadow-md'
                  : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
                }`}
            >
              {categoryInfo[category as keyof typeof categoryInfo]?.icon} {categoryInfo[category as keyof typeof categoryInfo]?.name}
            </button>
          ))}
        </div>
      </div>

      {/* 统计信息 */}
      <div className="mb-6 flex items-center justify-between">
        <div className="text-sm text-gray-600">
          显示 {filteredCommands.length} / {commands.length} 个命令
        </div>
        <button
          onClick={() => {
            fetchingRef.current = false // 重置请求状态，允许手动刷新
            fetchCommands()
          }}
          className="text-sm text-blue-500 hover:text-blue-600 font-medium"
        >
          刷新列表
        </button>
      </div>

      {/* 命令列表 */}
      <div className="grid gap-4 md:grid-cols-2">
        {filteredCommands.map(command => {
          // 检查是否是可配置的命令
          const template = commandTemplates.find(t => command.id.startsWith(t.templateId))
          const isConfigurable = template && template.ui && template.ui.params.length > 0

          return (
            <ConfiguredCommandCard
              key={command.id}
              command={command}
              onExecute={executeCommand}
              onAddToLayout={addToLayout}
              onDelete={deleteCommand}
              onEditConfig={isConfigurable ? handleCommandEdit : undefined}
              onDirectEdit={handleDirectEdit}
            />
          )
        })}
      </div>

      {filteredCommands.length === 0 && (
        <div className="text-center py-12">
          <div className="text-4xl mb-4">{commands.length === 0 ? '📦' : '🔍'}</div>
          <h3 className="text-lg font-medium text-gray-900 mb-2">
            {commands.length === 0 ? '还没有配置命令' : '没有找到命令'}
          </h3>
          <p className="text-gray-600 mb-4">
            {commands.length === 0
              ? '点击上方的 "+ 添加命令" 按钮开始配置'
              : (searchQuery ? '请尝试其他搜索关键词' : '当前分类下没有可用命令')
            }
          </p>
          {commands.length === 0 ? (
            <button
              onClick={() => setShowAddModal(true)}
              className="bg-blue-500 hover:bg-blue-600 text-white px-6 py-2 rounded-lg font-medium"
            >
              添加第一个命令
            </button>
          ) : (
            (searchQuery || selectedCategory) && (
              <button
                onClick={() => {
                  setSearchQuery('')
                  setSelectedCategory(null)
                }}
                className="text-blue-500 hover:text-blue-600 font-medium"
              >
                清除筛选条件
              </button>
            )
          )}
        </div>
      )}

      {/* 添加命令模态框 */}
      <AddCommandModal
        isOpen={showAddModal}
        onClose={() => setShowAddModal(false)}
        onConfigure={handleTemplateConfiguration}
      />

      {/* 直接编辑模态框 */}
      <DirectEditModal
        isOpen={showDirectEdit}
        onClose={() => {
          setShowDirectEdit(false)
          setEditingCommand(null)
        }}
        command={editingCommand}
        onSave={handleDirectEditSave}
      />

      {/* 参数配置模态框 */}
      {configureTemplate && (
        <ParameterForm
          template={configureTemplate}
          mode={configureMode}
          onAddCommand={configureMode !== 'execute' ? handleParameterAddCommand : undefined}
          onExecute={configureMode !== 'add' ? handleParameterExecute : undefined}
          onCancel={() => {
            setConfigureTemplate(null)
            setEditingCommand(null)
          }}
          initialParams={editingCommand ? {} : undefined} // TODO: 从已有命令中提取参数
        />
      )}
    </div>
  )
}