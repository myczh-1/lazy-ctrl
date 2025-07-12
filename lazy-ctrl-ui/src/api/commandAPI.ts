
export interface CommandInfo {
  id: string
  name: string
  description: string
  category: string
  icon: string
  timeout: number
  requiresPin: boolean
  whitelisted: boolean
  available: boolean
  command?: string
  showOnHomepage?: boolean
  homepagePosition?: {
    x: number
    y: number
    width: number
    height: number
  }
  homepageColor?: string
  homepagePriority?: number
}

// 创建命令的请求数据结构
export interface CreateCommandRequest {
  id: string
  name?: string
  description?: string
  category?: string
  icon?: string
  command: string
  platform: string
  commandType?: string
  security?: {
    requirePin: boolean
    whitelist: boolean
    adminOnly?: boolean
  }
  timeout?: number
  userId?: string
  deviceId?: string
  homeLayout?: {
    showOnHome: boolean
    defaultPosition?: {
      x: number
      y: number
      w: number
      h: number
    }
    color?: string
    priority?: number
  }
  templateId?: string
  templateParams?: Record<string, any>
  createdAt?: string
  updatedAt?: string
}

export interface CommandResponse {
  version: string
  commands: CommandInfo[]
}

export interface ExecutionResult {
  success: boolean
  output: string
  error?: string
  exit_code: number
  execution_time: number
}

// 开发环境使用代理，生产环境使用完整URL
const API_BASE_URL = import.meta.env.DEV ? '' : 'http://localhost:7070'

export class CommandAPI {
  private static instance: CommandAPI
  private baseURL: string
  private pin?: string

  private constructor(baseURL: string = API_BASE_URL) {
    this.baseURL = baseURL
  }

  static getInstance(baseURL?: string): CommandAPI {
    if (!CommandAPI.instance) {
      CommandAPI.instance = new CommandAPI(baseURL)
    }
    return CommandAPI.instance
  }

  setPin(pin: string) {
    this.pin = pin
  }

  private getHeaders(): HeadersInit {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    }
    
    if (this.pin) {
      headers['X-Pin'] = this.pin
    }
    
    return headers
  }

  private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const url = `${this.baseURL}${endpoint}`
    const config: RequestInit = {
      ...options,
      headers: {
        ...this.getHeaders(),
        ...options.headers,
      },
    }

    try {
      const response = await fetch(url, config)
      
      if (!response.ok) {
        const error = await response.text()
        throw new Error(`HTTP ${response.status}: ${error}`)
      }

      return await response.json()
    } catch (error) {
      console.error(`API request failed: ${endpoint}`, error)
      throw error
    }
  }

  /**
   * 获取所有可用命令列表
   */
  async getCommands(): Promise<CommandResponse> {
    return this.request<CommandResponse>('/api/v1/commands')
  }

  /**
   * 执行命令
   */
  async executeCommand(
    commandId: string, 
    timeout?: number
  ): Promise<ExecutionResult> {
    const params = new URLSearchParams({ id: commandId })
    if (timeout) {
      params.append('timeout', timeout.toString())
    }

    return this.request<ExecutionResult>(`/api/v1/execute?${params}`)
  }

  /**
   * 重新加载命令配置
   */
  async reloadCommands(): Promise<{ message: string }> {
    return this.request<{ message: string }>('/api/v1/reload', {
      method: 'POST',
    })
  }

  /**
   * 健康检查
   */
  async health(): Promise<{ status: string; timestamp: number }> {
    return this.request<{ status: string; timestamp: number }>('/api/v1/health')
  }

  /**
   * 创建新命令
   */
  async createCommand(command: CreateCommandRequest): Promise<{ message: string; id: string }> {
    return this.request<{ message: string; id: string }>('/api/v1/commands', {
      method: 'POST',
      body: JSON.stringify(command),
    })
  }

  /**
   * 更新命令
   */
  async updateCommand(id: string, command: CreateCommandRequest): Promise<{ message: string }> {
    return this.request<{ message: string }>(`/api/v1/commands/${id}`, {
      method: 'PUT',
      body: JSON.stringify(command),
    })
  }

  /**
   * 删除命令
   */
  async deleteCommand(id: string): Promise<{ message: string }> {
    return this.request<{ message: string }>(`/api/v1/commands/${id}`, {
      method: 'DELETE',
    })
  }

}

// 导出单例实例
const commandAPIInstance = CommandAPI.getInstance()
export default commandAPIInstance

// 全局错误处理器
export class APIError extends Error {
  constructor(
    message: string,
    public status?: number,
    public response?: any
  ) {
    super(message)
    this.name = 'APIError'
  }
}

// 命令执行状态类型
export const CommandStatus = {
  IDLE: 'idle',
  EXECUTING: 'executing',
  SUCCESS: 'success',
  ERROR: 'error',
} as const

export type CommandStatusType = typeof CommandStatus[keyof typeof CommandStatus]

// Hook for command execution with state management
export interface CommandExecutionState {
  status: CommandStatusType
  result?: ExecutionResult
  error?: string
}