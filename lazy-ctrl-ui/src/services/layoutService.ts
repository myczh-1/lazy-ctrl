import type { CardConfig } from '@/types/layout'
import { useLayoutStore } from '@/stores/layoutStore'
import { useCommandStore } from '@/stores/commandStore'
import { CommandService } from './commandService'
import commandAPI, { type CreateCommandRequest } from '@/api/commandAPI'

export class LayoutService {
  /**
   * 初始化布局
   */
  static async initializeLayout(): Promise<void> {
    const { loadFromLocalStorage, loadLayout } = useLayoutStore.getState()
    const { commands } = useCommandStore.getState()
    
    // 先尝试从 localStorage 加载
    const hasStoredData = loadFromLocalStorage()
    
    if (!hasStoredData && commands.length > 0) {
      // 没有存储数据，使用后端命令创建默认布局
      const availableCards = this.createCardsFromCommands(commands)
      if (availableCards.length > 0) {
        // 使用后端返回的位置信息创建布局
        const defaultLayout = availableCards.map(card => {
          const command = commands.find(cmd => cmd.id === card.commandId)
          const position = command?.homepagePosition
          
          return {
            i: card.id,
            x: position?.x || 0,
            y: position?.y || 0,
            w: position?.width || 1,
            h: position?.height || 1
          }
        })
        loadLayout(defaultLayout, availableCards)
        console.log('Initialized layout with backend commands:', availableCards.length)
      }
    }
  }

  /**
   * 将命令转换为卡片
   */
  static createCardsFromCommands(commands: any[]): CardConfig[] {
    return commands
      .filter(cmd => cmd.available && cmd.showOnHomepage)
      .map(cmd => ({
        id: cmd.id,
        title: cmd.name || cmd.id,
        commandId: cmd.id,
        icon: cmd.icon,
        category: cmd.category,
        description: cmd.description,
        available: cmd.available,
        requiresPin: cmd.requiresPin,
        timeout: cmd.timeout,
      }))
  }

  /**
   * 添加命令到布局
   */
  static async addCommandToLayout(commandId: string): Promise<boolean> {
    const { getCommandById } = useCommandStore.getState()
    const { addCard } = useLayoutStore.getState()
    
    const command = getCommandById(commandId)
    if (!command) {
      console.error('Command not found:', commandId)
      return false
    }

    const card: CardConfig = {
      id: `card_${commandId}_${Date.now()}`,
      title: command.name,
      commandId: command.id,
      icon: command.icon,
      category: command.category,
      description: command.description,
      available: command.available,
      requiresPin: command.requiresPin,
      timeout: command.timeout,
    }

    addCard(card)
    await this.saveLayout()
    return true
  }

  /**
   * 执行卡片对应的命令
   */
  static async executeCardCommand(cardId: string): Promise<void> {
    const { getCardById } = useLayoutStore.getState()
    const card = getCardById(cardId)
    
    if (!card || !card.commandId) {
      console.error('Card or command not found:', cardId)
      return
    }

    if (!card.available) {
      console.warn('Command not available:', card.commandId)
      return
    }

    await CommandService.executeCommand(card.commandId, card.timeout)
  }

  /**
   * 保存布局到 localStorage 和后端
   */
  static async saveLayout(): Promise<void> {
    const { saveToLocalStorage, layout, cards } = useLayoutStore.getState()
    
    // 保存到本地存储
    saveToLocalStorage()
    
    // 同步到后端：更新命令的homeLayout配置
    try {
      const updatePromises = cards.map(async (card) => {
        if (!card.commandId) return
        
        // 查找对应的布局项
        const layoutItem = layout.find(item => item.i === card.id)
        if (!layoutItem) return
        
        // 获取命令详情
        const { getCommandById } = useCommandStore.getState()
        const command = getCommandById(card.commandId)
        if (!command) return
        
        // 构建更新请求
        const commandRequest: CreateCommandRequest = {
          id: command.id,
          name: command.name,
          description: command.description || '',
          category: command.category,
          icon: command.icon,
          command: command.command || '',
          platform: this.getCurrentPlatform(),
          timeout: command.timeout || 10000,
          security: {
            requirePin: command.requiresPin || false,
            whitelist: command.available !== false
          },
          homeLayout: {
            showOnHome: true,
            defaultPosition: {
              x: layoutItem.x,
              y: layoutItem.y,
              w: layoutItem.w,
              h: layoutItem.h
            },
            color: '',
            priority: 0
          },
          updatedAt: new Date().toISOString()
        }
        
        // 更新后端
        await commandAPI.updateCommand(command.id, commandRequest)
      })
      
      await Promise.all(updatePromises)
      console.log('Layout synced to backend successfully')
    } catch (error) {
      console.error('Failed to sync layout to backend:', error)
      // 不阻止本地保存，即使同步失败
    }
  }
  
  /**
   * 获取当前平台
   */
  private static getCurrentPlatform(): string {
    const platform = navigator.platform.toLowerCase()
    if (platform.includes('win')) return 'windows'
    if (platform.includes('mac')) return 'darwin'
    return 'linux'
  }

  /**
   * 重置布局
   */
  static resetLayout(): void {
    localStorage.removeItem('lazy-ctrl-layout')
    window.location.reload()
  }

  /**
   * 更改卡片尺寸
   */
  static async changeCardSize(cardId: string, size: 'small' | 'medium' | 'large'): Promise<void> {
    const { layout, setLayout } = useLayoutStore.getState()
    
    const sizeMap = {
      small: { w: 1, h: 1 },
      medium: { w: 2, h: 2 },
      large: { w: 3, h: 3 },
    }

    const newLayout = layout.map(item =>
      item.i === cardId ? { ...item, ...sizeMap[size] } : item
    )
    
    setLayout(newLayout)
    await this.saveLayout()
  }

  /**
   * 处理卡片点击
   */
  static async handleCardClick(cardId: string): Promise<void> {
    const { editMode } = useLayoutStore.getState()
    
    if (editMode) {
      // 编辑模式下切换选中状态
      const { selectedCard, setSelectedCard } = useLayoutStore.getState()
      setSelectedCard(selectedCard === cardId ? null : cardId)
    } else {
      // 非编辑模式下执行命令
      await this.executeCardCommand(cardId)
    }
  }

  /**
   * 导出布局配置
   */
  static exportLayout() {
    const { layout, cards } = useLayoutStore.getState()
    return {
      layout,
      cards,
      timestamp: Date.now()
    }
  }

  /**
   * 导入布局配置
   */
  static async importLayout(data: any): Promise<void> {
    const { loadLayout } = useLayoutStore.getState()
    if (data.layout && data.cards) {
      loadLayout(data.layout, data.cards)
      await this.saveLayout()
    }
  }
}