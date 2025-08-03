import commandAPI, { type CommandInfo } from '@/api/commandAPI'
import { useCommandStore } from '@/stores/commandStore'
import { commandTemplates } from '@/data/commandTemplates'
import { getExecutionErrorMessage } from '@/utils/errorHandler'

export class CommandService {
  /**
   * 获取所有命令
   */
  static async fetchCommands(): Promise<void> {
    const { setLoading, setError, setCommands } = useCommandStore.getState()
    
    try {
      setLoading(true)
      setError(null)
      
      const response = await commandAPI.getCommands()
      
      // 处理后端的统一响应格式
      let commands = []
      if (response.success && response.data) {
        commands = Array.isArray(response.data) ? response.data : []
      } else {
        // 兼容旧格式
        commands = Array.isArray(response.data) ? response.data : []
      }
      
      // 数据转换和兼容性处理
      const processedCommands = commands.map((cmd: any) => ({
        id: cmd.id,
        name: cmd.name || '',
        description: cmd.description || '',
        category: cmd.category || '',
        icon: cmd.icon || '',
        timeout: cmd.timeout || 10000,
        requiresPin: cmd.requiresPin || false,
        whitelisted: cmd.whitelisted !== false, // 默认为true
        available: cmd.available !== false, // 默认为true
        command: cmd.command || '',
        showOnHomepage: cmd.showOnHomepage !== false, // 默认为true
        homepagePosition: cmd.homepagePosition ? {
          x: cmd.homepagePosition.x || 0,
          y: cmd.homepagePosition.y || 0,
          width: cmd.homepagePosition.width || 1,
          height: cmd.homepagePosition.height || 1
        } : { x: 0, y: 0, width: 1, height: 1 },
        homepageColor: cmd.homepageColor || '',
        homepagePriority: cmd.homepagePriority || 0
      }))
      
      setCommands(processedCommands)
      
      console.log(`Loaded ${processedCommands.length} commands`)
      console.log('Available commands:', processedCommands.filter(cmd => cmd.available).length)
      console.log('Homepage commands:', processedCommands.filter(cmd => cmd.showOnHomepage).length)
      console.log('Commands data:', processedCommands)
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '获取命令列表失败'
      setError(errorMessage)
      console.error('Failed to load commands:', error)
    } finally {
      setLoading(false)
    }
  }

  /**
   * 执行命令
   */
  static async executeCommand(commandId: string, timeout?: number): Promise<void> {
    const { setExecutionState } = useCommandStore.getState()
    
    try {
      setExecutionState({ status: 'executing', commandId })
      
      console.log(`Executing command: ${commandId}`)
      const result = await commandAPI.executeCommand(commandId, timeout)
      
      setExecutionState({
        status: result.success ? 'success' : 'error',
        commandId,
        result,
        error: result.success ? undefined : getExecutionErrorMessage(result)
      })
      
      console.log(`Command execution ${result.success ? 'succeeded' : 'failed'}:`, result)
      
      // 自动清除执行状态
      setTimeout(() => {
        setExecutionState({ status: 'idle' })
      }, 3000)
      
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : '命令执行失败'
      setExecutionState({
        status: 'error',
        commandId,
        error: errorMessage
      })
      console.error('Command execution failed:', error)
      
      // 自动清除错误状态
      setTimeout(() => {
        setExecutionState({ status: 'idle' })
      }, 5000)
    }
  }

  /**
   * 设置 PIN
   */
  static setPin(pin: string): void {
    const { setPin } = useCommandStore.getState()
    commandAPI.setPin(pin)
    setPin(pin)
    
    // 保存到 localStorage
    localStorage.setItem('lazy-ctrl-pin', pin)
  }

  /**
   * 从 localStorage 加载 PIN
   */
  static loadPin(): void {
    try {
      const savedPin = localStorage.getItem('lazy-ctrl-pin')
      if (savedPin) {
        this.setPin(savedPin)
      }
    } catch (error) {
      console.error('Failed to load PIN:', error)
    }
  }

  /**
   * 获取可用的命令模板
   */
  static getCommandTemplates() {
    return commandTemplates
  }

  /**
   * 根据模板创建命令
   */
  static createCommandFromTemplate(templateId: string, params: Record<string, any> = {}) {
    const template = commandTemplates.find(t => t.templateId === templateId)
    if (!template) {
      throw new Error(`Template not found: ${templateId}`)
    }

    // 生成唯一ID
    const commandId = `${templateId}_${Date.now()}`
    
    // 替换参数占位符
    const processedPlatforms = Object.fromEntries(
      Object.entries(template.platforms).map(([platform, command]) => [
        platform,
        this.replaceCommandParams(command, params)
      ])
    )

    const newCommand: CommandInfo = {
      id: commandId,
      name: template.name,
      description: template.description,
      category: template.category,
      icon: template.icon,
      timeout: 5000,
      requiresPin: false,
      whitelisted: true,
      available: true,
      command: typeof processedPlatforms.all === 'string' ? processedPlatforms.all : undefined
    }

    return newCommand
  }

  /**
   * 替换命令中的参数占位符
   */
  private static replaceCommandParams(command: string | any[], params: Record<string, any>): string | any[] {
    if (Array.isArray(command)) {
      return command.map(step => {
        if (step.cmd) {
          return { ...step, cmd: this.replaceCommandParams(step.cmd, params) }
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
}