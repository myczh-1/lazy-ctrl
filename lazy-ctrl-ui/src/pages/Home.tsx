// 文件：src/pages/Home.tsx
import { useState, useEffect, useRef } from 'react'
import GridLayout from 'react-grid-layout'
import type { Layout } from 'react-grid-layout'
import 'react-grid-layout/css/styles.css'
import 'react-resizable/css/styles.css'
import { useEditMode } from '../contexts/EditModeContext'
import { useLayoutManager } from '@/hooks/useLayoutManager'
import { useCommandAPI } from '@/hooks/useCommandAPI'
import type { CardConfig } from '@/types/layout'
import { CommandStatus } from '@/api/commandAPI'

import layoutAPI from '../api/layoutAPI'
import { SettingsModal } from '../components/SettingsModal'

// 默认卡片配置（仅在没有保存数据时使用）
const defaultCards: CardConfig[] = [
    { id: '1', title: '卡片 1', commandId: 'cmd1' },
    { id: '2', title: '卡片 2', commandId: 'cmd2' },
    { id: '3', title: '卡片 3', commandId: 'cmd3' },
]

const initialLayout: Layout[] = [
    { i: '1', x: 0, y: 0, w: 1, h: 2 },
    { i: '2', x: 1, y: 0, w: 1, h: 1 },
    { i: '3', x: 2, y: 0, w: 2, h: 2 },
]

const SIZES = {
    small: { w: 1, h: 1 },
    medium: { w: 2, h: 2 },
    large: { w: 3, h: 3 },
}

export default function Home() {
    // 后端 API 状态
    const {
        commands,
        isLoading,
        error,
        executionState,
        executeCommand,
        getAvailableCards,
        refreshCommands
    } = useCommandAPI()
    
    // 布局管理
    const {
        layout,
        cards,
        setLayout,
        handleCardClick,
        removeCard,
        createCard,
        loadLayout,
        saveLayout,
        loadLayoutFromStorage,
        updateCard
    } = useLayoutManager([], [])  // 空初始值，让hook从localStorage加载
    
    const { editMode, setEditMode } = useEditMode()
    const [containerWidth, setContainerWidth] = useState(400)
    const [selectedCard, setSelectedCard] = useState<string | null>(null)
    const [dragState, setDragState] = useState<{ isDragging: boolean; draggedItem: string | null }>({
        isDragging: false,
        draggedItem: null
    })
    const [showSettings, setShowSettings] = useState(false)
    const containerRef = useRef<HTMLDivElement>(null)

    // 注册 API
    useEffect(() => {
        layoutAPI.setLayoutManager({
            layout,
            cards,
            setLayout,
            handleCardClick,
            removeCard,
            createCard,
            loadLayout,
            saveLayout,
            exportLayout: () => ({ layout, cards, timestamp: Date.now() }),
            importLayout: (data: any) => loadLayout(data.layout, data.cards),
            loadLayoutFromStorage: () => {
                const savedData = localStorage.getItem('lazy-ctrl-layout')
                if (savedData) {
                    try {
                        const { layout: savedLayout, cards: savedCards } = JSON.parse(savedData)
                        if (savedLayout && savedCards) {
                            loadLayout(savedLayout, savedCards)
                            return true
                        }
                    } catch (error) {
                        console.error('Failed to load saved layout:', error)
                    }
                }
                return false
            },
            executeCommand: async (commandId: string) => {
                await executeCommand(commandId)
                return 'Command executed via API'
            },
            updateCard: updateCard
        })
    }, [layout, cards])

    // 初始化布局数据
    useEffect(() => {
        // 如果没有布局数据，尝试从localStorage加载
        if (layout.length === 0 && cards.length === 0) {
            const hasStoredData = loadLayoutFromStorage()
            if (!hasStoredData) {
                // 没有存储数据，检查是否有后端命令
                if (commands.length > 0) {
                    // 使用后端命令创建默认布局
                    const availableCards = getAvailableCards()
                    if (availableCards.length > 0) {
                        const defaultLayout = availableCards.slice(0, 6).map((card, index) => ({
                            i: card.id,
                            x: (index % 4),
                            y: Math.floor(index / 4),
                            w: 1,
                            h: 1
                        }))
                        loadLayout(defaultLayout, availableCards)
                        console.log('Initialized layout with backend commands:', availableCards.length)
                    }
                } else if (!isLoading) {
                    // 后端没有可用命令且不在加载中，使用静态默认配置
                    loadLayout(initialLayout, defaultCards)
                    console.log('Initialized layout with default cards')
                }
            }
        }
    }, [commands, layout.length, cards.length, isLoading])

    useEffect(() => {
        const updateWidth = () => {
            if (containerRef.current) {
                const width = containerRef.current.offsetWidth
                setContainerWidth(width - 16) // 减去 padding
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
    }, [editMode])

    const changeSize = (i: string, sizeKey: keyof typeof SIZES) => {
        const newLayout = layout.map((item) =>
            item.i === i ? { ...item, ...SIZES[sizeKey] } : item
        )
        setLayout(newLayout)
    }

    const removeItem = (i: string) => {
        removeCard(i)
        setSelectedCard(null)
    }

    // 拖拽开始

    const handleDragStart = (_layout: Layout[], oldItem: Layout, _newItem: Layout, _placeholder: Layout, _e: MouseEvent, _element: HTMLElement) => {
        setDragState({ isDragging: true, draggedItem: oldItem.i })
    }

    // 拖拽结束
    const handleDragStop = (_layout: Layout[], oldItem: Layout, newItem: Layout, _placeholder: Layout, _e: MouseEvent, _element: HTMLElement) => {
        const wasDragging = dragState.isDragging
        const draggedItem = dragState.draggedItem
        
        setDragState({ isDragging: false, draggedItem: null })
        
        // 如果位置没有改变，认为是点击而不是拖拽
        if (wasDragging && draggedItem && oldItem.x === newItem.x && oldItem.y === newItem.y && oldItem.w === newItem.w && oldItem.h === newItem.h) {
            if (editMode) {
                // 编辑模式下切换选中状态
                setSelectedCard(selectedCard === draggedItem ? null : draggedItem)
            } else {
                // 非编辑模式下执行卡片功能
                const card = cards.find(c => c.id === draggedItem)
                if (card && card.commandId) {
                    executeCommand(card.commandId)
                } else {
                    handleCardClick(draggedItem)
                }
            }
        }
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
                            onClick={refreshCommands}
                            className="text-red-600 hover:text-red-800 underline text-sm"
                        >
                            重试
                        </button>
                    </div>
                </div>
            )}
            
            {/* 执行状态 */}
            {executionState.status === CommandStatus.EXECUTING && (
                <div className="fixed top-4 right-4 p-3 bg-blue-50 border border-blue-200 rounded-md shadow-lg z-50">
                    <div className="flex items-center">
                        <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600 mr-2"></div>
                        <span className="text-blue-700">正在执行命令...</span>
                    </div>
                </div>
            )}
            
            {executionState.status === CommandStatus.SUCCESS && (
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
            
            {executionState.status === CommandStatus.ERROR && (
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
                        onClick={refreshCommands}
                        className="p-2 text-gray-600 hover:text-gray-800 hover:bg-gray-100 rounded-md transition-colors"
                        title="刷新命令"
                        disabled={isLoading}
                    >
                        {isLoading ? '🔄' : '🔄'}
                    </button>
                    
                    <button
                        onClick={() => {
                            localStorage.removeItem('lazy-ctrl-layout')
                            window.location.reload()
                        }}
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


            <div className="grid-layout-container">
                <GridLayout
                    className={`layout select-none ${editMode ? 'touch-none' : ''}`}
                    layout={layout}
                    cols={4}
                    rowHeight={100}
                    width={containerWidth}
                    isDraggable={editMode}
                    isResizable={editMode}
                    onLayoutChange={(l) => setLayout(l)}
                    onDragStart={handleDragStart}
                    onDragStop={handleDragStop}
                    style={editMode ? { WebkitTouchCallout: 'none' } : {}}
                >
                {layout.map((item) => (
                    <div
                        key={item.i}
                        className={`bg-white rounded-md border shadow-md flex flex-col relative transition-all ${
                            (() => {
                                const card = cards.find(c => c.id === item.i)
                                return card?.available === false ? 'opacity-50 ' : ''
                            })()
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
                            {(() => {
                                const card = cards.find(c => c.id === item.i)
                                return (
                                    <div>
                                        {/* 卡片图标 */}
                                        {card?.icon && (
                                            <div className="text-lg mb-1">
                                                {card.icon === 'volume-mute' && '🔇'}
                                                {card.icon === 'volume-up' && '🔊'}
                                                {card.icon === 'lock' && '🔒'}
                                                {card.icon === 'volume-plus' && '🔊+'}
                                                {card.icon === 'volume-minus' && '🔊-'}
                                                {card.icon === 'test' && '🧪'}
                                                {card.icon === 'power' && '⚡'}
                                                {card.icon === 'terminal' && '💻'}
                                                {card.icon === 'sequence' && '🔄'}
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
                                        
                                        {/* 编辑模式状态 */}
                                        {editMode && selectedCard === item.i && (
                                            <div className="text-xs text-blue-500 mt-1">已选中</div>
                                        )}
                                        {editMode && (
                                            <div className="text-xs text-gray-400 mt-1">
                                                {dragState.isDragging && dragState.draggedItem === item.i ? '拖拽中...' : '点击选择'}
                                            </div>
                                        )}
                                        
                                        {/* 非编辑模式状态 */}
                                        {!editMode && (
                                            <div className="text-xs text-gray-400 mt-1">
                                                {card?.available !== false ? '点击执行' : '不可用'}
                                            </div>
                                        )}
                                    </div>
                                )
                            })()}
                        </div>
                    </div>
                ))}
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
                            onClick={() => changeSize(selectedCard, 'small')}
                        >
                            小尺寸
                        </button>
                        <button 
                            className="px-4 py-2 bg-blue-100 text-blue-600 rounded-md hover:bg-blue-200 transition-colors"
                            onClick={() => changeSize(selectedCard, 'medium')}
                        >
                            中尺寸
                        </button>
                        <button 
                            className="px-4 py-2 bg-blue-100 text-blue-600 rounded-md hover:bg-blue-200 transition-colors"
                            onClick={() => changeSize(selectedCard, 'large')}
                        >
                            大尺寸
                        </button>
                        <button
                            className="px-4 py-2 bg-red-100 text-red-600 rounded-md hover:bg-red-200 transition-colors"
                            onClick={() => removeItem(selectedCard)}
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
