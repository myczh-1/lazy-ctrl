import { useState, useEffect } from 'react'
import type { Layout } from 'react-grid-layout'
import type { CardConfig, LayoutData, LayoutManagerHook } from '@/types/layout'

export function useLayoutManager(initialLayout: Layout[] = [], initialCards: CardConfig[] = []): LayoutManagerHook {
    const [layout, setLayout] = useState<Layout[]>(initialLayout)
    const [cards, setCards] = useState<CardConfig[]>(initialCards)

    // 创建新卡片
    const createCard = (config: Omit<CardConfig, 'id'>) => {
        const newId = Date.now().toString()
        const newCard: CardConfig = { ...config, id: newId }
        setCards(prev => [...prev, newCard])
        
        // 添加到布局
        const newLayout: Layout = {
            i: newId,
            x: 0,
            y: Math.max(...layout.map(l => l.y + l.h), 0),
            w: 1,
            h: 1
        }
        setLayout(prev => [...prev, newLayout])
        return newId
    }

    // 删除卡片
    const removeCard = (cardId: string) => {
        setLayout(prev => prev.filter(item => item.i !== cardId))
        setCards(prev => prev.filter(card => card.id !== cardId))
    }

    // 更新卡片配置
    const updateCard = (cardId: string, updates: Partial<CardConfig>) => {
        setCards(prev => prev.map(card => 
            card.id === cardId ? { ...card, ...updates } : card
        ))
    }

    // 加载布局
    const loadLayout = (newLayout: Layout[], newCards?: CardConfig[]) => {
        setLayout(newLayout)
        if (newCards) {
            setCards(newCards)
        }
    }

    // 保存布局到 localStorage
    const saveLayout = (): LayoutData => {
        const layoutData: LayoutData = {
            layout,
            cards,
            timestamp: Date.now()
        }
        localStorage.setItem('lazy-ctrl-layout', JSON.stringify(layoutData))
        return layoutData
    }

    // 从 localStorage 加载布局
    const loadLayoutFromStorage = () => {
        const savedData = localStorage.getItem('lazy-ctrl-layout')
        if (savedData) {
            try {
                const { layout: savedLayout, cards: savedCards } = JSON.parse(savedData) as LayoutData
                if (savedLayout && savedCards) {
                    setLayout(savedLayout)
                    setCards(savedCards)
                    return true
                }
            } catch (error) {
                console.error('Failed to load saved layout:', error)
            }
        }
        return false
    }

    // 导出布局为 JSON
    const exportLayout = (): LayoutData => {
        return {
            layout,
            cards,
            timestamp: Date.now()
        }
    }

    // 导入布局
    const importLayout = (data: LayoutData) => {
        setLayout(data.layout)
        setCards(data.cards)
    }

    // 执行命令
    const executeCommand = async (commandId: string) => {
        try {
            const response = await fetch(`http://localhost:7070/execute?id=${commandId}`)
            const result = await response.text()
            console.log('Command result:', result)
            return result
        } catch (error) {
            console.error('Command execution failed:', error)
            throw error
        }
    }

    // 处理卡片点击
    const handleCardClick = (cardId: string) => {
        const card = cards.find(c => c.id === cardId)
        if (!card) return

        if (card.action) {
            card.action()
        } else if (card.commandId) {
            executeCommand(card.commandId)
        }
    }

    // 自动加载
    useEffect(() => {
        if (initialLayout.length === 0 && initialCards.length === 0) {
            loadLayoutFromStorage()
        }
    }, [])

    return {
        layout,
        cards,
        setLayout,
        setCards,
        createCard,
        removeCard,
        updateCard,
        loadLayout,
        saveLayout,
        loadLayoutFromStorage,
        exportLayout,
        importLayout,
        executeCommand,
        handleCardClick
    }
}