import type { Layout } from 'react-grid-layout'
import type { CardConfig, LayoutData } from '@/types/layout'

export class LayoutAPI {
    private static instance: LayoutAPI
    private layoutManagerRef: any = null

    private constructor() {}

    static getInstance(): LayoutAPI {
        if (!LayoutAPI.instance) {
            LayoutAPI.instance = new LayoutAPI()
        }
        return LayoutAPI.instance
    }

    setLayoutManager(layoutManager: any) {
        this.layoutManagerRef = layoutManager
    }

    // 创建新卡片
    createCard(config: Omit<CardConfig, 'id'>): string | null {
        if (!this.layoutManagerRef) {
            console.error('Layout manager not initialized')
            return null
        }
        return this.layoutManagerRef.createCard(config)
    }

    // 删除卡片
    removeCard(cardId: string): boolean {
        if (!this.layoutManagerRef) {
            console.error('Layout manager not initialized')
            return false
        }
        this.layoutManagerRef.removeCard(cardId)
        return true
    }

    // 更新卡片配置
    updateCard(cardId: string, updates: Partial<CardConfig>): boolean {
        if (!this.layoutManagerRef) {
            console.error('Layout manager not initialized')
            return false
        }
        this.layoutManagerRef.updateCard(cardId, updates)
        return true
    }

    // 获取当前布局
    getCurrentLayout(): LayoutData | null {
        if (!this.layoutManagerRef) {
            console.error('Layout manager not initialized')
            return null
        }
        return this.layoutManagerRef.exportLayout()
    }

    // 加载布局
    loadLayout(layout: Layout[], cards?: CardConfig[]): boolean {
        if (!this.layoutManagerRef) {
            console.error('Layout manager not initialized')
            return false
        }
        this.layoutManagerRef.loadLayout(layout, cards)
        return true
    }

    // 保存布局
    saveLayout(): LayoutData | null {
        if (!this.layoutManagerRef) {
            console.error('Layout manager not initialized')
            return null
        }
        return this.layoutManagerRef.saveLayout()
    }

    // 从本地存储加载
    loadFromStorage(): boolean {
        if (!this.layoutManagerRef) {
            console.error('Layout manager not initialized')
            return false
        }
        return this.layoutManagerRef.loadLayoutFromStorage()
    }

    // 导出布局为 JSON
    exportLayout(): LayoutData | null {
        if (!this.layoutManagerRef) {
            console.error('Layout manager not initialized')
            return null
        }
        return this.layoutManagerRef.exportLayout()
    }

    // 导入布局
    importLayout(data: LayoutData): boolean {
        if (!this.layoutManagerRef) {
            console.error('Layout manager not initialized')
            return false
        }
        this.layoutManagerRef.importLayout(data)
        return true
    }

    // 执行命令
    async executeCommand(commandId: string): Promise<string | null> {
        if (!this.layoutManagerRef) {
            console.error('Layout manager not initialized')
            return null
        }
        try {
            return await this.layoutManagerRef.executeCommand(commandId)
        } catch (error) {
            console.error('Command execution failed:', error)
            return null
        }
    }

    // 获取所有卡片
    getAllCards(): CardConfig[] {
        if (!this.layoutManagerRef) {
            console.error('Layout manager not initialized')
            return []
        }
        return this.layoutManagerRef.cards
    }

    // 获取布局信息
    getLayout(): Layout[] {
        if (!this.layoutManagerRef) {
            console.error('Layout manager not initialized')
            return []
        }
        return this.layoutManagerRef.layout
    }
}

// 默认导出单例实例
export default LayoutAPI.getInstance()

// 全局方法暴露（可在浏览器控制台或其他地方调用）
declare global {
    interface Window {
        lazyCtrlAPI: LayoutAPI
    }
}

if (typeof window !== 'undefined') {
    window.lazyCtrlAPI = LayoutAPI.getInstance()
}