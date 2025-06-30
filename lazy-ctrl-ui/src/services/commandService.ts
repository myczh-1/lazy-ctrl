import commandAPI, { type CommandInfo } from '@/api/commandAPI'
import { useCommandStore } from '@/stores/commandStore'
import { commandTemplates } from '@/data/commandTemplates'

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
      setCommands(response.commands)
      
      console.log(`Loaded ${response.commands.length} commands`)
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
        error: result.error
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