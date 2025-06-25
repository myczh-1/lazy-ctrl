// 文件：src/pages/Home.tsx
import { useState, useEffect, useRef } from 'react'
import GridLayout, { Layout } from 'react-grid-layout'
import 'react-grid-layout/css/styles.css'
import 'react-resizable/css/styles.css'
import { useEditMode } from '../contexts/EditModeContext'

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
    const [layout, setLayout] = useState<Layout[]>(initialLayout)
    const { editMode, setEditMode } = useEditMode()
    const [containerWidth, setContainerWidth] = useState(400)
    const [selectedCard, setSelectedCard] = useState<string | null>(null)
    const [dragState, setDragState] = useState<{ isDragging: boolean; draggedItem: string | null }>({
        isDragging: false,
        draggedItem: null
    })
    const containerRef = useRef<HTMLDivElement>(null)

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
        setLayout(layout.filter((item) => item.i !== i))
        setSelectedCard(null)
    }

    // 拖拽开始
    const handleDragStart = (layout: Layout[], oldItem: Layout, newItem: Layout, placeholder: Layout, e: MouseEvent, element: HTMLElement) => {
        setDragState({ isDragging: true, draggedItem: oldItem.i })
    }

    // 拖拽结束
    const handleDragStop = (layout: Layout[], oldItem: Layout, newItem: Layout, placeholder: Layout, e: MouseEvent, element: HTMLElement) => {
        const wasDragging = dragState.isDragging
        const draggedItem = dragState.draggedItem
        
        setDragState({ isDragging: false, draggedItem: null })
        
        // 如果位置没有改变，认为是点击而不是拖拽
        if (wasDragging && draggedItem && oldItem.x === newItem.x && oldItem.y === newItem.y && oldItem.w === newItem.w && oldItem.h === newItem.h) {
            // 这是一个点击事件，切换选中状态
            setSelectedCard(selectedCard === draggedItem ? null : draggedItem)
        }
    }

    return (
        <div className="p-2" ref={containerRef}>


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
                            卡片 {item.i}
                            {editMode && selectedCard === item.i && (
                                <div className="text-xs text-blue-500 mt-1">已选中</div>
                            )}
                            {editMode && (
                                <div className="text-xs text-gray-400 mt-1">
                                    {dragState.isDragging && dragState.draggedItem === item.i ? '拖拽中...' : '点击选择'}
                                </div>
                            )}
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
            
            <style dangerouslySetInnerHTML={{
                __html: `
                    .grid-layout-container .react-grid-placeholder {
                        background: rgba(59, 130, 246, 0.3) !important;
                        border: 2px dashed #3b82f6 !important;
                        border-radius: 8px !important;
                    }
                `
            }} />
        </div>
    )
}
