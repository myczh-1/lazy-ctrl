import { useState, useCallback, useEffect } from 'react'
import commandAPI, { 
  CommandAPI,
  type CommandInfo, 
  CommandStatus, 
  type CommandExecutionState 
} from '@/api/commandAPI'
import type { CardConfig } from '@/types/layout'

export interface UseCommandAPIResult {
  // 命令列表状态
  commands: CommandInfo[]
  isLoading: boolean
  error: string | null
  
  // 执行状态
  executionState: CommandExecutionState
  
  // 方法
  refreshCommands: () => Promise<void>
  executeCommand: (commandId: string, timeout?: number) => Promise<void>
  setPin: (pin: string) => void
  
  // 便捷方法
  getAvailableCards: () => CardConfig[]
  getCommandById: (id: string) => CommandInfo | undefined
}

export function useCommandAPI(): UseCommandAPIResult {
  const [commands, setCommands] = useState<CommandInfo[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [executionState, setExecutionState] = useState<CommandExecutionState>({
    status: CommandStatus.IDLE
  })

  // 获取命令列表
  const refreshCommands = useCallback(async () => {
    try {
      setIsLoading(true)
      setError(null)
      
      const response = await commandAPI.getCommands()
      setCommands(response.commands)
      
      console.log(`Loaded ${response.commands.length} commands (version ${response.version})`)
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : '获取命令列表失败'
      setError(errorMessage)
      console.error('Failed to load commands:', err)
    } finally {
      setIsLoading(false)
    }
  }, [])

  // 执行命令
  const executeCommand = useCallback(async (commandId: string, timeout?: number) => {
    try {
      setExecutionState({ status: CommandStatus.EXECUTING })
      
      console.log(`Executing command: ${commandId}`)
      const result = await commandAPI.executeCommand(commandId, timeout)
      
      setExecutionState({
        status: result.success ? CommandStatus.SUCCESS : CommandStatus.ERROR,
        result,
        error: result.error
      })
      
      console.log(`Command execution ${result.success ? 'succeeded' : 'failed'}:`, result)
      
      // 自动清除执行状态
      setTimeout(() => {
        setExecutionState({ status: CommandStatus.IDLE })
      }, 3000)
      
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : '命令执行失败'
      setExecutionState({
        status: CommandStatus.ERROR,
        error: errorMessage
      })
      console.error('Command execution failed:', err)
      
      // 自动清除错误状态
      setTimeout(() => {
        setExecutionState({ status: CommandStatus.IDLE })
      }, 5000)
    }
  }, [])

  // 设置PIN
  const setPin = useCallback((pin: string) => {
    commandAPI.setPin(pin)
  }, [])

  // 获取可用的卡片配置
  const getAvailableCards = useCallback((): CardConfig[] => {
    return CommandAPI.commandsToCards(commands)
  }, [commands])

  // 根据ID获取命令
  const getCommandById = useCallback((id: string): CommandInfo | undefined => {
    return commands.find(cmd => cmd.id === id)
  }, [commands])

  // 初始加载
  useEffect(() => {
    refreshCommands()
  }, [refreshCommands])

  return {
    commands,
    isLoading,
    error,
    executionState,
    refreshCommands,
    executeCommand,
    setPin,
    getAvailableCards,
    getCommandById,
  }
}

// 用于单个命令执行的简化Hook
export function useCommandExecution() {
  const [state, setState] = useState<CommandExecutionState>({
    status: CommandStatus.IDLE
  })

  const execute = useCallback(async (commandId: string, timeout?: number) => {
    try {
      setState({ status: CommandStatus.EXECUTING })
      
      const result = await commandAPI.executeCommand(commandId, timeout)
      
      setState({
        status: result.success ? CommandStatus.SUCCESS : CommandStatus.ERROR,
        result,
        error: result.error
      })
      
      return result
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : '命令执行失败'
      setState({
        status: CommandStatus.ERROR,
        error: errorMessage
      })
      throw err
    }
  }, [])

  const reset = useCallback(() => {
    setState({ status: CommandStatus.IDLE })
  }, [])

  return {
    ...state,
    execute,
    reset,
    isExecuting: state.status === CommandStatus.EXECUTING,
    isSuccess: state.status === CommandStatus.SUCCESS,
    isError: state.status === CommandStatus.ERROR,
  }
}