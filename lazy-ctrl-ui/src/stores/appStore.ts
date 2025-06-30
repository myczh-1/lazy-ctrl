import { create } from 'zustand'

interface AppState {
  // UI 状态
  showSettings: boolean
  showAddCommandModal: boolean
  
  // 设置
  apiBaseUrl: string
  
  // 操作
  setShowSettings: (show: boolean) => void
  setShowAddCommandModal: (show: boolean) => void
  setApiBaseUrl: (url: string) => void
  
  // 初始化
  loadSettings: () => void
  saveSettings: () => void
}

export const useAppStore = create<AppState>((set, get) => ({
  // 初始状态
  showSettings: false,
  showAddCommandModal: false,
  apiBaseUrl: import.meta.env.DEV ? '/api' : 'http://localhost:7070',

  // 操作
  setShowSettings: (showSettings) => set({ showSettings }),
  setShowAddCommandModal: (showAddCommandModal) => set({ showAddCommandModal }),
  setApiBaseUrl: (apiBaseUrl) => set({ apiBaseUrl }),

  // 设置管理
  loadSettings: () => {
    try {
      const savedApiUrl = localStorage.getItem('lazy-ctrl-api-url')
      if (savedApiUrl) {
        set({ apiBaseUrl: savedApiUrl })
      }
    } catch (error) {
      console.error('Failed to load settings:', error)
    }
  },

  saveSettings: () => {
    try {
      const { apiBaseUrl } = get()
      localStorage.setItem('lazy-ctrl-api-url', apiBaseUrl)
    } catch (error) {
      console.error('Failed to save settings:', error)
    }
  },
}))