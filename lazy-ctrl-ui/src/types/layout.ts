import type { Layout } from 'react-grid-layout'

export interface CardConfig {
    id: string
    title: string
    action?: () => void
    commandId?: string
    icon?: string
    color?: string
}

export interface LayoutData {
    layout: Layout[]
    cards: CardConfig[]
    timestamp: number
}

export interface LayoutManagerHook {
    layout: Layout[]
    cards: CardConfig[]
    setLayout: (layout: Layout[]) => void
    setCards: (cards: CardConfig[]) => void
    createCard: (config: Omit<CardConfig, 'id'>) => string
    removeCard: (cardId: string) => void
    updateCard: (cardId: string, updates: Partial<CardConfig>) => void
    loadLayout: (newLayout: Layout[], newCards?: CardConfig[]) => void
    saveLayout: () => LayoutData
    loadLayoutFromStorage: () => boolean
    exportLayout: () => LayoutData
    importLayout: (data: LayoutData) => void
    executeCommand: (commandId: string) => Promise<string>
    handleCardClick: (cardId: string) => void
}