import { useEffect, useRef, useState } from 'react'
import GridLayout from 'react-grid-layout'
import type { Layout } from 'react-grid-layout'
import 'react-grid-layout/css/styles.css'
import 'react-resizable/css/styles.css'

import { useCommandStore } from '@/stores/commandStore'
import { useLayoutStore } from '@/stores/layoutStore'
import { useAppStore } from '@/stores/appStore'
import { CommandService } from '@/services/commandService'
import { LayoutService } from '@/services/layoutService'
import { SettingsModal } from '@/components/SettingsModal'

export default function Home() {
    // 状态管理
    const { commands, isLoading, error, executionState } = useCommandStore()
    const { 
        layout, 
        cards, 
        editMode, 
        selectedCard,
        setLayout,
        setEditMode,
        setSelectedCard 
    } = useLayoutStore()
    const { showSettings, setShowSettings } = useAppStore()
    
    // 本地状态
    const [containerWidth, setContainerWidth] = useState(400)
    const [dragState, setDragState] = useState<{ isDragging: boolean; draggedItem: string | null }>({
        isDragging: false,
        draggedItem: null
    })
    const containerRef = useRef<HTMLDivElement>(null)

    // 初始化
    useEffect(() => {
        // 加载应用设置
        useAppStore.getState().loadSettings()
        
        // 加载 PIN
        CommandService.loadPin()
        
        // 获取命令列表
        CommandService.fetchCommands()
    }, [])

    // 初始化布局
    useEffect(() => {
        if (!isLoading && commands.length > 0) {
            LayoutService.initializeLayout()
        }
    }, [commands, isLoading])

    // 响应式容器宽度
    useEffect(() => {
        const updateWidth = () => {
            if (containerRef.current) {
                const width = containerRef.current.offsetWidth
                setContainerWidth(width - 16)
            }
        }

        updateWidth()
        window.addEventListener('resize', updateWidth)
        return () => window.removeEventListener('resize', updateWidth)
    }, [])

    // 退出编辑模式时清除选中状态
    useEffect(() => {
        if (!editMode) {
            setSelectedCard(null)
        }
    }, [editMode, setSelectedCard])

    // 拖拽处理
    const handleDragStart = (_layout: Layout[], oldItem: Layout) => {
        setDragState({ isDragging: true, draggedItem: oldItem.i })
    }

    const handleDragStop = (_layout: Layout[], oldItem: Layout, newItem: Layout) => {
        const wasDragging = dragState.isDragging
        const draggedItem = dragState.draggedItem
        
        setDragState({ isDragging: false, draggedItem: null })
        
        // 如果位置没有改变，认为是点击而不是拖拽
        if (wasDragging && draggedItem && 
            oldItem.x === newItem.x && oldItem.y === newItem.y && 
            oldItem.w === newItem.w && oldItem.h === newItem.h) {
            LayoutService.handleCardClick(draggedItem)
        }
    }

    // 工具函数
    const handleRefresh = () => {
        CommandService.fetchCommands()
    }

    const handleReset = () => {
        LayoutService.resetLayout()
    }

    const handleSizeChange = (size: 'small' | 'medium' | 'large') => {
        if (selectedCard) {
            LayoutService.changeCardSize(selectedCard, size)
        }
    }

    const handleRemoveCard = () => {
        if (selectedCard) {
            useLayoutStore.getState().removeCard(selectedCard)
            LayoutService.saveLayout()
        }
    }

    // 渲染卡片图标
    const renderCardIcon = (icon?: string) => {
        const iconMap: Record<string, string> = {
            'volume-mute': '🔇',
            'volume-up': '🔊',
            'lock': '🔒',
            'volume-plus': '🔊+',
            'volume-minus': '🔊-',
            'test': '🧪',
            'power': '⚡',
            'terminal': '💻',
            'sequence': '🔄'
        }
        return iconMap[icon || ''] || '📱'
    }

    return (
        <div className="p-2" ref={containerRef}>
            {/* 加载状态 */}
            {isLoading && (
                <div className="mb-4 p-3 bg-blue-50 border border-blue-200 rounded-md">
                    <div className="flex items-center">
                        <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600 mr-2"></div>
                        <span className="text-blue-700">加载命令列表中...</span>
                    </div>
                </div>
            )}
            
            {/* 错误状态 */}
            {error && (
                <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-md">
                    <div className="flex items-center justify-between">
                        <span className="text-red-700">错误: {error}</span>
                        <button 
                            onClick={handleRefresh}
                            className="text-red-600 hover:text-red-800 underline text-sm"
                        >
                            重试
                        </button>
                    </div>
                </div>
            )}
            
            {/* 执行状态提示 */}
            {executionState.status === 'executing' && (
                <div className="fixed top-4 right-4 p-3 bg-blue-50 border border-blue-200 rounded-md shadow-lg z-50">
                    <div className="flex items-center">
                        <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600 mr-2"></div>
                        <span className="text-blue-700">正在执行命令...</span>
                    </div>
                </div>
            )}
            
            {executionState.status === 'success' && (
                <div className="fixed top-4 right-4 p-3 bg-green-50 border border-green-200 rounded-md shadow-lg z-50">
                    <div className="text-green-700">
                        <div className="font-medium">命令执行成功</div>
                        {executionState.result?.output && (
                            <div className="text-sm mt-1 max-w-xs truncate">
                                {executionState.result.output}
                            </div>
                        )}
                    </div>
                </div>
            )}
            
            {executionState.status === 'error' && (
                <div className="fixed top-4 right-4 p-3 bg-red-50 border border-red-200 rounded-md shadow-lg z-50">
                    <div className="text-red-700">
                        <div className="font-medium">命令执行失败</div>
                        <div className="text-sm mt-1">{executionState.error}</div>
                    </div>
                </div>
            )}
            
            {/* 顶部工具栏 */}
            <div className="flex justify-between items-center mb-4">
                <div className="flex items-center gap-4">
                    <h1 className="text-xl font-semibold text-gray-800">Lazy Control</h1>
                    {commands.length > 0 && (
                        <div className="text-sm text-gray-500">
                            {commands.filter(cmd => cmd.available).length} 个可用命令
                        </div>
                    )}
                </div>
                
                <div className="flex items-center gap-2">
                    <button
                        onClick={() => setShowSettings(true)}
                        className="p-2 text-gray-600 hover:text-gray-800 hover:bg-gray-100 rounded-md transition-colors"
                        title="设置"
                    >
                        ⚙️
                    </button>
                    
                    <button
                        onClick={handleRefresh}
                        className="p-2 text-gray-600 hover:text-gray-800 hover:bg-gray-100 rounded-md transition-colors"
                        title="刷新命令"
                        disabled={isLoading}
                    >
                        🔄
                    </button>
                    
                    <button
                        onClick={handleReset}
                        className="px-2 py-1 text-xs text-red-600 hover:text-red-800 hover:bg-red-50 rounded border border-red-200 transition-colors"
                        title="重置布局"
                    >
                        重置
                    </button>
                    
                    <button
                        onClick={() => setEditMode(!editMode)}
                        className={`px-3 py-1 rounded-md text-sm transition-colors ${
                            editMode 
                                ? 'bg-blue-600 text-white hover:bg-blue-700'
                                : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
                        }`}
                    >
                        {editMode ? '完成编辑' : '编辑布局'}
                    </button>
                </div>
            </div>

            {/* 网格布局 */}
            <div className="grid-layout-container">
                <GridLayout
                    className={`layout select-none ${editMode ? 'touch-none' : ''}`}
                    layout={layout}
                    cols={4}
                    rowHeight={100}
                    width={containerWidth}
                    isDraggable={editMode}
                    isResizable={editMode}
                    onLayoutChange={(newLayout) => {
                        setLayout(newLayout)
                        // 自动保存布局变化
                        LayoutService.saveLayout()
                    }}
                    onDragStart={handleDragStart}
                    onDragStop={handleDragStop}
                    onResizeStop={() => {
                        // 调整大小后也要保存
                        LayoutService.saveLayout()
                    }}
                    style={editMode ? { WebkitTouchCallout: 'none' } : {}}
                >
                    {layout.map((item) => {
                        const card = cards.find(c => c.id === item.i)
                        
                        return (
                            <div
                                key={item.i}
                                className={`bg-white rounded-md border shadow-md flex flex-col relative transition-all ${
                                    card?.available === false ? 'opacity-50 ' : ''
                                }${
                                    editMode && selectedCard === item.i 
                                        ? 'border-blue-400 border-2 shadow-lg' 
                                        : 'border-gray-300'
                                }`}
                                onContextMenu={(e) => {
                                    e.preventDefault()
                                    setEditMode(true)
                                }}
                            >
                                {/* 卡片内容 */}
                                <div className="flex-1 p-2 text-center text-sm flex flex-col justify-center">
                                    {/* 卡片图标 */}
                                    {card?.icon && (
                                        <div className="text-lg mb-1">
                                            {renderCardIcon(card.icon)}
                                        </div>
                                    )}
                                    
                                    {/* 卡片标题 */}
                                    <div className="font-medium">
                                        {card ? card.title : `卡片 ${item.i}`}
                                    </div>
                                    
                                    {/* 卡片描述 */}
                                    {card?.description && !editMode && (
                                        <div className="text-xs text-gray-500 mt-1 line-clamp-2">
                                            {card.description}
                                        </div>
                                    )}
                                    
                                    {/* 类别标签 */}
                                    {card?.category && !editMode && (
                                        <div className="inline-block px-2 py-1 text-xs bg-gray-100 text-gray-600 rounded mt-1">
                                            {card.category}
                                        </div>
                                    )}
                                    
                                    {/* PIN 要求指示 */}
                                    {card?.requiresPin && !editMode && (
                                        <div className="text-xs text-orange-500 mt-1">
                                            🔐 需要PIN
                                        </div>
                                    )}
                                    
                                    {/* 状态提示 */}
                                    {editMode && selectedCard === item.i && (
                                        <div className="text-xs text-blue-500 mt-1">已选中</div>
                                    )}
                                    {editMode && selectedCard !== item.i && (
                                        <div className="text-xs text-gray-400 mt-1">
                                            {dragState.isDragging && dragState.draggedItem === item.i ? '拖拽中...' : '点击选择'}
                                        </div>
                                    )}
                                    {!editMode && (
                                        <div className="text-xs text-gray-400 mt-1">
                                            {card?.available !== false ? '点击执行' : '不可用'}
                                        </div>
                                    )}
                                </div>
                            </div>
                        )
                    })}
                </GridLayout>
            </div>

            {/* 选中卡片的工具栏 */}
            {editMode && selectedCard && (
                <div className="fixed bottom-16 left-0 right-0 bg-white border-t border-gray-200 p-3 shadow-lg">
                    <div className="text-center text-sm text-gray-600 mb-2">
                        编辑卡片 {selectedCard}
                    </div>
                    <div className="flex justify-around">
                        <button 
                            className="px-4 py-2 bg-blue-100 text-blue-600 rounded-md hover:bg-blue-200 transition-colors"
                            onClick={() => handleSizeChange('small')}
                        >
                            小尺寸
                        </button>
                        <button 
                            className="px-4 py-2 bg-blue-100 text-blue-600 rounded-md hover:bg-blue-200 transition-colors"
                            onClick={() => handleSizeChange('medium')}
                        >
                            中尺寸
                        </button>
                        <button 
                            className="px-4 py-2 bg-blue-100 text-blue-600 rounded-md hover:bg-blue-200 transition-colors"
                            onClick={() => handleSizeChange('large')}
                        >
                            大尺寸
                        </button>
                        <button
                            className="px-4 py-2 bg-red-100 text-red-600 rounded-md hover:bg-red-200 transition-colors"
                            onClick={handleRemoveCard}
                        >
                            删除
                        </button>
                    </div>
                </div>
            )}
            
            {/* 设置模态框 */}
            <SettingsModal 
                isOpen={showSettings} 
                onClose={() => setShowSettings(false)} 
            />
            
            <style dangerouslySetInnerHTML={{
                __html: `
                    .grid-layout-container .react-grid-placeholder {
                        background: rgba(59, 130, 246, 0.3) !important;
                        border: 2px dashed #3b82f6 !important;
                        border-radius: 8px !important;
                    }
                    .line-clamp-2 {
                        display: -webkit-box;
                        -webkit-line-clamp: 2;
                        -webkit-box-orient: vertical;
                        overflow: hidden;
                    }
                `
            }} />
        </div>
    )
}