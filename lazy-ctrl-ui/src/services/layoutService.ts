import type { CardConfig } from '@/types/layout'
import { useLayoutStore } from '@/stores/layoutStore'
import { useCommandStore } from '@/stores/commandStore'
import { CommandService } from './commandService'

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
        const defaultLayout = availableCards.slice(0, 6).map((card, index) => ({
          i: card.id,
          x: index % 4,
          y: Math.floor(index / 4),
          w: 1,
          h: 1
        }))
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
      .filter(cmd => cmd.available)
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
  static addCommandToLayout(commandId: string): boolean {
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
    this.saveLayout()
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
   * 保存布局到 localStorage
   */
  static saveLayout(): void {
    const { saveToLocalStorage } = useLayoutStore.getState()
    saveToLocalStorage()
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
  static changeCardSize(cardId: string, size: 'small' | 'medium' | 'large'): void {
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
    this.saveLayout()
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
  static importLayout(data: any): void {
    const { loadLayout } = useLayoutStore.getState()
    if (data.layout && data.cards) {
      loadLayout(data.layout, data.cards)
      this.saveLayout()
    }
  }
}