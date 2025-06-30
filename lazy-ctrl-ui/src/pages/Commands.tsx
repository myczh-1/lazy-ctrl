import { useState, useEffect } from 'react'
import type { Command, DisplayCommand } from '@/types/command'
import { commandTemplates, categoryInfo, templateToCommands, type CommandTemplate } from '@/data/commandTemplates'
import { LayoutService } from '@/services/layoutService'
import ParameterForm from '@/components/ParameterForm'

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
const ConfiguredCommandCard = ({ command, onExecute, onAddToLayout, onDelete, onEditConfig }: { 
  command: DisplayCommand; 
  onExecute: (id: string) => void;
  onAddToLayout: (command: DisplayCommand) => void;
  onDelete: (id: string) => void;
  onEditConfig?: (command: DisplayCommand) => void;
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
    const platform = getCurrentPlatform()
    const platformCommand = command.platforms[platform] || command.platforms['all']
    
    if (Array.isArray(platformCommand)) {
      return `多步骤命令 (${platformCommand.length} 步骤)`
    }
    return platformCommand || 'N/A'
  }
  
  const getCurrentPlatform = () => {
    const platform = navigator.platform.toLowerCase()
    if (platform.includes('win')) return 'windows'
    if (platform.includes('mac')) return 'darwin'
    return 'linux'
  }

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
            <p className="text-xs text-gray-500 mb-1">当前平台命令:</p>
            <code className="text-xs text-gray-700 font-mono break-all">
              {getCurrentPlatformCommand()}
            </code>
            {Object.keys(command.platforms).length > 1 && (
              <p className="text-xs text-gray-500 mt-1">
                支持平台: {Object.keys(command.platforms).join(', ')}
              </p>
            )}
          </div>
          
          {showActions ? (
            <div className="space-y-2">
              <div className="grid grid-cols-3 gap-2">
                <button
                  onClick={handleExecute}
                  disabled={isExecuting}
                  className={`px-3 py-2 rounded-lg font-medium text-sm transition-all ${
                    isExecuting
                      ? 'bg-gray-200 text-gray-500 cursor-not-allowed'
                      : 'bg-blue-500 hover:bg-blue-600 text-white shadow-sm'
                  }`}
                >
                  {isExecuting ? '执行中' : '执行'}
                </button>
                <button
                  onClick={() => onAddToLayout(command)}
                  className="px-3 py-2 bg-green-500 hover:bg-green-600 text-white rounded-lg font-medium text-sm transition-all shadow-sm"
                >
                  添加到主页
                </button>
                <button
                  onClick={() => onDelete(command.id)}
                  className="px-3 py-2 bg-red-500 hover:bg-red-500 text-white rounded-lg font-medium text-sm transition-all shadow-sm"
                >
                  删除
                </button>
              </div>
              {onEditConfig && (
                <button
                  onClick={() => onEditConfig(command)}
                  className="w-full px-3 py-2 bg-purple-500 hover:bg-purple-600 text-white rounded-lg font-medium text-sm transition-all shadow-sm"
                >
                  ⚙️ 重新配置参数
                </button>
              )}
            </div>
          ) : (
            <div className="grid grid-cols-2 gap-2">
              <button
                onClick={handleExecute}
                disabled={isExecuting}
                className={`px-4 py-2 rounded-lg font-medium text-sm transition-all ${
                  isExecuting
                    ? 'bg-gray-200 text-gray-500 cursor-not-allowed'
                    : 'bg-blue-500 hover:bg-blue-600 text-white shadow-sm'
                }`}
              >
                {isExecuting ? '执行中...' : '执行命令'}
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

  // 获取命令列表
  const fetchCommands = async () => {
    try {
      // 首先尝试从 controller agent 获取
      const response = await fetch('/api/v1/commands')
      if (response.ok) {
        const data = await response.json()
        const parsedCommands = parseCommandsFromAPI(data)
        setCommands(parsedCommands)
      } else {
        // 如果 API 不可用，从本地存储获取
        const savedCommands = localStorage.getItem('lazy-ctrl-commands')
        if (savedCommands) {
          const rawCommands = JSON.parse(savedCommands) as Command[]
          const displayCommands = groupCommandsForDisplay(rawCommands)
          setCommands(displayCommands)
        } else {
          setCommands([])
        }
      }
    } catch (error) {
      console.error('Failed to fetch commands:', error)
      // 从本地存储获取
      const savedCommands = localStorage.getItem('lazy-ctrl-commands')
      if (savedCommands) {
        const rawCommands = JSON.parse(savedCommands) as Command[]
        const displayCommands = groupCommandsForDisplay(rawCommands)
        setCommands(displayCommands)
      } else {
        setCommands([])
      }
    } finally {
      setLoading(false)
    }
  }

  // 解析从 API 获取的命令配置（兼容旧格式）
  const parseCommandsFromAPI = (rawCommands: any): DisplayCommand[] => {
    // 如果是新格式（数组）
    if (Array.isArray(rawCommands)) {
      return groupCommandsForDisplay(rawCommands as Command[])
    }
    
    // 如果是旧格式（对象），转换为新格式
    const commands: Command[] = []
    Object.entries(rawCommands).forEach(([id, config]: [string, any]) => {
      const categoryInfo = inferCategoryFromId(id)
      
      if (typeof config === 'string') {
        commands.push({
          id,
          name: formatCommandName(id),
          platform: 'all',
          command: config,
          category: categoryInfo.category,
          icon: getCategoryIcon(categoryInfo.category),
          description: categoryInfo.description
        })
      } else if (typeof config === 'object') {
        // 多平台命令
        Object.entries(config).forEach(([platform, cmd]) => {
          commands.push({
            id,
            name: formatCommandName(id),
            platform,
            command: cmd as string,
            category: categoryInfo.category,
            icon: getCategoryIcon(categoryInfo.category),
            description: categoryInfo.description
          })
        })
      }
    })
    
    return groupCommandsForDisplay(commands)
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

  // 根据ID推断分类
  const inferCategoryFromId = (id: string) => {
    if (id.includes('volume') || id.includes('mute')) {
      return { category: 'audio', description: '音频控制命令' }
    }
    if (id.includes('lock') || id.includes('shutdown')) {
      return { category: 'system', description: '系统控制命令' }
    }
    if (id.includes('power') || id.includes('sleep')) {
      return { category: 'power', description: '电源管理命令' }
    }
    if (id.includes('test')) {
      return { category: 'custom', description: '测试命令' }
    }
    return { category: 'custom', description: '自定义命令' }
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

  // 保存命令到本地存储
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
    
    const newCommands = templateToCommands(processedTemplate)
    const newDisplayCommand = groupCommandsForDisplay(newCommands)[0]
    
    const updatedCommands = [...commands, newDisplayCommand]
    setCommands(updatedCommands)
    saveCommands(updatedCommands)
    
    showToast(`命令 "${newDisplayCommand.name}" 已添加`, 'success')
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
  const handleTemplateConfiguration = (template: CommandTemplate, mode: 'add' | 'execute' | 'both') => {
    setConfigureTemplate(template)
    setConfigureMode(mode)
    setEditingCommand(null)
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
      const platform = getCurrentPlatform()
      const platformCommand = processedTemplate.platforms[platform] || processedTemplate.platforms.all
      
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
  
  // 获取当前平台
  const getCurrentPlatform = () => {
    const platform = navigator.platform.toLowerCase()
    if (platform.includes('win')) return 'windows'
    if (platform.includes('mac')) return 'darwin'
    return 'linux'
  }
  
  // 删除命令
  const deleteCommand = async (commandId: string) => {
    if (confirm('确定要删除这个命令吗？')) {
      const updatedCommands = commands.filter(cmd => cmd.id !== commandId)
      setCommands(updatedCommands)
      saveCommands(updatedCommands)
      showToast('命令已删除', 'success')
    }
  }
  
  // 添加命令到主页布局
  const addToLayout = (displayCommand: DisplayCommand) => {
    try {
      const success = LayoutService.addCommandToLayout(displayCommand.id)
      if (success) {
        showToast(`"${displayCommand.name}" 已添加到主页`, 'success')
      } else {
        showToast('添加失败，请确保主页已加载', 'error')
      }
    } catch (error) {
      console.error('Add to layout failed:', error)
      showToast('添加失败', 'error')
    }
  }

  // 执行命令
  const executeCommand = async (commandId: string) => {
    try {
      const response = await fetch(`/api/v1/execute?id=${commandId}`)
      const result = await response.text()
      console.log('Command result:', result)
      
      const command = commands.find(c => c.id === commandId)
      showToast(`命令 "${command?.name || commandId}" 执行成功`, 'success')
    } catch (error) {
      console.error('Command execution failed:', error)
      showToast(`命令执行失败: ${error}`, 'error')
    }
  }
  
  // 显示提示消息
  const showToast = (message: string, type: 'success' | 'error') => {
    const toast = document.createElement('div')
    toast.className = `fixed top-4 right-4 px-4 py-2 rounded-lg shadow-lg z-50 text-white ${
      type === 'success' ? 'bg-green-500' : 'bg-red-500'
    }`
    toast.textContent = message
    document.body.appendChild(toast)
    setTimeout(() => document.body.removeChild(toast), 3000)
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
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 mb-2">命令管理</h1>
          <p className="text-gray-600">管理系统命令，添加到主页或直接执行</p>
        </div>
        <button
          onClick={() => setShowAddModal(true)}
          className="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-lg font-medium shadow-sm hover:shadow-md active:scale-95 transition-all"
        >
          + 添加命令
        </button>
      </div>

      {/* 搜索和筛选 */}
      <div className="mb-6 space-y-4">
        <div className="relative">
          <input
            type="text"
            placeholder="搜索命令..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          />
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <span className="text-gray-400">🔍</span>
          </div>
        </div>
        
        <div className="flex flex-wrap gap-2">
          <button
            onClick={() => setSelectedCategory(null)}
            className={`px-3 py-1 rounded-full text-sm font-medium transition-all ${
              selectedCategory === null
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
              className={`px-3 py-1 rounded-full text-sm font-medium transition-all ${
                selectedCategory === category
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
          onClick={fetchCommands}
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