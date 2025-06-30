import { create } from 'zustand'
import type { CommandInfo } from '@/api/commandAPI'

export interface CommandExecutionState {
  status: 'idle' | 'executing' | 'success' | 'error'
  commandId?: string
  result?: {
    success: boolean
    output: string
    error?: string
    exit_code: number
    execution_time: number
  }
  error?: string
}

interface CommandState {
  // 状态
  commands: CommandInfo[]
  isLoading: boolean
  error: string | null
  executionState: CommandExecutionState
  pin?: string

  // 操作
  setCommands: (commands: CommandInfo[]) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  setExecutionState: (state: CommandExecutionState) => void
  setPin: (pin: string) => void
  
  // 便捷方法
  getCommandById: (id: string) => CommandInfo | undefined
  getAvailableCommands: () => CommandInfo[]
}

export const useCommandStore = create<CommandState>((set, get) => ({
  // 初始状态
  commands: [],
  isLoading: false,
  error: null,
  executionState: { status: 'idle' },
  pin: undefined,

  // 操作
  setCommands: (commands) => set({ commands }),
  setLoading: (isLoading) => set({ isLoading }),
  setError: (error) => set({ error }),
  setExecutionState: (executionState) => set({ executionState }),
  setPin: (pin) => set({ pin }),

  // 便捷方法
  getCommandById: (id) => {
    const { commands } = get()
    return commands.find(cmd => cmd.id === id)
  },
  
  getAvailableCommands: () => {
    const { commands } = get()
    return commands.filter(cmd => cmd.available)
  },
}))