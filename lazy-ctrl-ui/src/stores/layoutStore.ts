import { create } from 'zustand'
import type { Layout } from 'react-grid-layout'
import type { CardConfig } from '@/types/layout'

interface LayoutState {
  // 状态
  layout: Layout[]
  cards: CardConfig[]
  editMode: boolean
  selectedCard: string | null

  // 操作
  setLayout: (layout: Layout[]) => void
  setCards: (cards: CardConfig[]) => void
  setEditMode: (editMode: boolean) => void
  setSelectedCard: (cardId: string | null) => void
  
  // 卡片管理
  addCard: (card: CardConfig) => void
  removeCard: (cardId: string) => void
  updateCard: (cardId: string, updates: Partial<CardConfig>) => void
  
  // 布局管理
  loadLayout: (layout: Layout[], cards: CardConfig[]) => void
  saveToLocalStorage: () => void
  loadFromLocalStorage: () => boolean
  
  // 便捷方法
  getCardById: (id: string) => CardConfig | undefined
}

const STORAGE_KEY = 'lazy-ctrl-layout'

export const useLayoutStore = create<LayoutState>((set, get) => ({
  // 初始状态
  layout: [],
  cards: [],
  editMode: false,
  selectedCard: null,

  // 基础操作
  setLayout: (layout) => set({ layout }),
  setCards: (cards) => set({ cards }),
  setEditMode: (editMode) => set({ editMode, selectedCard: editMode ? get().selectedCard : null }),
  setSelectedCard: (selectedCard) => set({ selectedCard }),

  // 卡片管理
  addCard: (card) => {
    const { cards, layout } = get()
    const newCards = [...cards, card]
    
    // 自动添加到布局
    const newLayoutItem: Layout = {
      i: card.id,
      x: layout.length % 4,
      y: Math.floor(layout.length / 4),
      w: 1,
      h: 1
    }
    const newLayout = [...layout, newLayoutItem]
    
    set({ cards: newCards, layout: newLayout })
  },

  removeCard: (cardId) => {
    const { cards, layout } = get()
    const newCards = cards.filter(card => card.id !== cardId)
    const newLayout = layout.filter(item => item.i !== cardId)
    
    set({ cards: newCards, layout: newLayout, selectedCard: null })
  },

  updateCard: (cardId, updates) => {
    const { cards } = get()
    const newCards = cards.map(card => 
      card.id === cardId ? { ...card, ...updates } : card
    )
    set({ cards: newCards })
  },

  // 布局管理
  loadLayout: (layout, cards) => {
    set({ layout, cards })
  },

  saveToLocalStorage: () => {
    const { layout, cards } = get()
    const data = {
      layout,
      cards,
      timestamp: Date.now()
    }
    localStorage.setItem(STORAGE_KEY, JSON.stringify(data))
  },

  loadFromLocalStorage: () => {
    try {
      const saved = localStorage.getItem(STORAGE_KEY)
      if (!saved) return false
      
      const data = JSON.parse(saved)
      if (data.layout && data.cards) {
        set({ layout: data.layout, cards: data.cards })
        return true
      }
    } catch (error) {
      console.error('Failed to load layout from localStorage:', error)
    }
    return false
  },

  // 便捷方法
  getCardById: (id) => {
    const { cards } = get()
    return cards.find(card => card.id === id)
  },
}))