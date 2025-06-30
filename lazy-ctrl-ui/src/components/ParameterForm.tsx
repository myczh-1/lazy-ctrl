import { useState, useEffect } from 'react'
import type { UIConfig, UIParam, CommandTemplate } from '@/types/command'
import type { Command } from '@/types/command'

interface ParameterFormProps {
  template: CommandTemplate
  onExecute?: (params: Record<string, any>) => void  // 执行命令
  onAddCommand?: (params: Record<string, any>) => void  // 添加到命令列表
  onCancel: () => void
  mode?: 'execute' | 'add' | 'both'  // 模式：仅执行、仅添加、两者都有
  initialParams?: Record<string, any>  // 初始参数（用于编辑模式）
}

// 单个参数输入组件
const ParameterInput = ({ param, value, onChange }: {
  param: UIParam
  value: any
  onChange: (value: any) => void
}) => {
  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const newValue = param.type === 'number' || param.type === 'range' 
      ? Number(e.target.value) 
      : e.target.value
    onChange(newValue)
  }

  const renderInput = () => {
    switch (param.type) {
      case 'range':
        const rangeValue = value !== undefined ? value : (param.default !== undefined ? param.default : param.min || 0)
        return (
          <div className="space-y-2">
            <div className="flex items-center space-x-4">
              <input
                type="range"
                min={param.min}
                max={param.max}
                step={param.step || 1}
                value={rangeValue}
                onChange={handleChange}
                className="flex-1 h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer slider"
              />
              <div className="flex items-center min-w-0">
                <span className="text-lg font-semibold text-blue-600">
                  {rangeValue}
                </span>
                {param.unit && <span className="text-sm text-gray-500 ml-1">{param.unit}</span>}
              </div>
            </div>
            <div className="flex justify-between text-xs text-gray-400">
              <span>{param.min}{param.unit}</span>
              <span>{param.max}{param.unit}</span>
            </div>
          </div>
        )

      case 'number':
        const numberValue = value !== undefined ? value : (param.default !== undefined ? param.default : '')
        return (
          <div className="flex items-center space-x-2">
            <input
              type="number"
              min={param.min}
              max={param.max}
              step={param.step || 1}
              value={numberValue}
              onChange={handleChange}
              placeholder={param.placeholder}
              className="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            />
            {param.unit && <span className="text-sm text-gray-500">{param.unit}</span>}
          </div>
        )

      case 'select':
        const selectValue = value !== undefined ? value : (param.default !== undefined ? param.default : '')
        return (
          <select
            value={selectValue}
            onChange={handleChange}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          >
            <option value="">请选择...</option>
            {param.options?.map(option => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        )

      case 'textarea':
        const textareaValue = value !== undefined ? value : (param.default !== undefined ? param.default : '')
        return (
          <textarea
            value={textareaValue}
            onChange={handleChange}
            placeholder={param.placeholder}
            rows={4}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 resize-vertical"
          />
        )

      case 'file':
        return (
          <input
            type="file"
            onChange={(e) => {
              const file = e.target.files?.[0]
              onChange(file ? file.path || file.name : '')
            }}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          />
        )

      default: // text
        const textValue = value !== undefined ? value : (param.default !== undefined ? param.default : '')
        return (
          <input
            type="text"
            value={textValue}
            onChange={handleChange}
            placeholder={param.placeholder}
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          />
        )
    }
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <label className="block text-sm font-medium text-gray-700">
          {param.label}
          {param.required && <span className="text-red-500 ml-1">*</span>}
        </label>
      </div>
      {renderInput()}
      {param.description && (
        <p className="text-xs text-gray-500">{param.description}</p>
      )}
    </div>
  )
}

// 命令预览组件
const CommandPreview = ({ template, params }: {
  template: CommandTemplate
  params: Record<string, any>
}) => {
  const [currentPlatform, setCurrentPlatform] = useState<string>('all')

  useEffect(() => {
    const platform = navigator.platform.toLowerCase()
    if (platform.includes('win')) setCurrentPlatform('windows')
    else if (platform.includes('mac')) setCurrentPlatform('darwin')
    else setCurrentPlatform('linux')
  }, [])

  const replaceParams = (command: string | any[], params: Record<string, any>): string => {
    if (Array.isArray(command)) {
      return JSON.stringify(command.map(step => {
        if (step.cmd) {
          return { ...step, cmd: replaceParams(step.cmd, params) }
        }
        if (step.duration && typeof step.duration === 'string') {
          return { ...step, duration: params[step.duration.replace(/[{}]/g, '')] || step.duration }
        }
        return step
      }), null, 2)
    }

    let result = command
    Object.entries(params).forEach(([key, value]) => {
      const regex = new RegExp(`{{${key}}}`, 'g')
      result = result.replace(regex, String(value))
    })
    return result
  }

  const getPreviewCommand = () => {
    const platformCommand = template.platforms[currentPlatform] || template.platforms.all
    if (!platformCommand) return 'N/A'
    
    return replaceParams(platformCommand, params)
  }

  return (
    <div className="bg-gray-50 rounded-lg p-4 border">
      <div className="flex items-center justify-between mb-2">
        <h4 className="text-sm font-medium text-gray-700">命令预览</h4>
        <select
          value={currentPlatform}
          onChange={(e) => setCurrentPlatform(e.target.value)}
          className="text-xs px-2 py-1 border border-gray-300 rounded"
        >
          {Object.keys(template.platforms).map(platform => (
            <option key={platform} value={platform}>
              {platform === 'all' ? '通用' : platform}
            </option>
          ))}
        </select>
      </div>
      <pre className="text-xs font-mono text-gray-800 whitespace-pre-wrap break-all bg-white p-3 rounded border">
        {getPreviewCommand()}
      </pre>
    </div>
  )
}

export default function ParameterForm({ 
  template, 
  onExecute, 
  onAddCommand, 
  onCancel, 
  mode = 'both',
  initialParams = {}
}: ParameterFormProps) {
  const [params, setParams] = useState<Record<string, any>>({})
  const [errors, setErrors] = useState<Record<string, string>>({})

  const uiConfig = template.ui!
  
  // 初始化默认值和初始参数
  useEffect(() => {
    const defaultParams: Record<string, any> = {}
    uiConfig.params.forEach(param => {
      if (initialParams[param.key] !== undefined) {
        defaultParams[param.key] = initialParams[param.key]
      } else if (param.default !== undefined) {
        defaultParams[param.key] = param.default
      } else if (param.type === 'range' && param.min !== undefined) {
        // 滑块类型如果没有默认值，使用最小值
        defaultParams[param.key] = param.min
      }
    })
    setParams(defaultParams)
    console.log('初始化参数:', defaultParams) // 调试日志
  }, [uiConfig, initialParams])

  const updateParam = (key: string, value: any) => {
    console.log('更新参数:', key, '=', value) // 调试日志
    setParams(prev => {
      const newParams = { ...prev, [key]: value }
      console.log('新参数状态:', newParams) // 调试日志
      return newParams
    })
    // 清除该字段的错误
    if (errors[key]) {
      setErrors(prev => ({ ...prev, [key]: '' }))
    }
  }

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {}
    let isValid = true

    uiConfig.params.forEach(param => {
      const value = params[param.key]
      
      if (param.required && (!value || value === '')) {
        newErrors[param.key] = `${param.label}是必填项`
        isValid = false
      }
      
      if (param.type === 'number' || param.type === 'range') {
        const numValue = Number(value)
        if (value !== '' && value !== undefined) {
          if (param.min !== undefined && numValue < param.min) {
            newErrors[param.key] = `${param.label}不能小于 ${param.min}`
            isValid = false
          }
          if (param.max !== undefined && numValue > param.max) {
            newErrors[param.key] = `${param.label}不能大于 ${param.max}`
            isValid = false
          }
        }
      }
    })

    setErrors(newErrors)
    return isValid
  }

  const handleExecute = () => {
    if (validateForm() && onExecute) {
      onExecute(params)
    }
  }
  
  const handleAddCommand = () => {
    if (validateForm() && onAddCommand) {
      onAddCommand(params)
    }
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
      <div className="bg-white rounded-xl max-w-2xl w-full max-h-[90vh] overflow-hidden">
        {/* 头部 */}
        <div className="p-6 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-3">
              <span className="text-2xl">{template.icon}</span>
              <div>
                <h2 className="text-xl font-bold text-gray-900">
                  {uiConfig.title || template.name}
                </h2>
                <p className="text-gray-600 text-sm">
                  {uiConfig.description || template.description}
                </p>
              </div>
            </div>
            <button
              onClick={onCancel}
              className="text-gray-400 hover:text-gray-600 text-xl"
            >
              ✕
            </button>
          </div>
        </div>

        {/* 表单内容 */}
        <div className="p-6 overflow-y-auto max-h-[60vh] space-y-6">
          {uiConfig.params.map(param => (
            <div key={param.key}>
              <ParameterInput
                param={param}
                value={params[param.key]}
                onChange={(value) => updateParam(param.key, value)}
              />
              {errors[param.key] && (
                <p className="text-red-500 text-xs mt-1">{errors[param.key]}</p>
              )}
            </div>
          ))}

          {/* 命令预览 */}
          {uiConfig.preview && (
            <CommandPreview template={template} params={params} />
          )}
        </div>

        {/* 底部操作按钮 */}
        <div className="p-6 border-t border-gray-200 flex justify-end space-x-3">
          <button
            onClick={onCancel}
            className="px-4 py-2 text-gray-600 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors"
          >
            取消
          </button>
          
          {/* 按模式显示不同按钮 */}
          {mode === 'execute' && onExecute && (
            <button
              onClick={handleExecute}
              className="px-6 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-lg font-medium transition-colors"
            >
              执行命令
            </button>
          )}
          
          {mode === 'add' && onAddCommand && (
            <button
              onClick={handleAddCommand}
              className="px-6 py-2 bg-green-500 hover:bg-green-600 text-white rounded-lg font-medium transition-colors"
            >
              添加到列表
            </button>
          )}
          
          {mode === 'both' && (
            <>
              {onAddCommand && (
                <button
                  onClick={handleAddCommand}
                  className="px-4 py-2 bg-green-500 hover:bg-green-600 text-white rounded-lg font-medium transition-colors"
                >
                  添加到列表
                </button>
              )}
              {onExecute && (
                <button
                  onClick={handleExecute}
                  className="px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-lg font-medium transition-colors"
                >
                  执行命令
                </button>
              )}
            </>
          )}
        </div>
      </div>

      {/* 滑块样式 */}
      <style dangerouslySetInnerHTML={{
        __html: `
          .slider::-webkit-slider-thumb {
            appearance: none;
            height: 20px;
            width: 20px;
            border-radius: 50%;
            background: #3b82f6;
            cursor: pointer;
            box-shadow: 0 2px 4px rgba(0,0,0,0.2);
          }
          .slider::-moz-range-thumb {
            height: 20px;
            width: 20px;
            border-radius: 50%;
            background: #3b82f6;
            cursor: pointer;
            border: none;
            box-shadow: 0 2px 4px rgba(0,0,0,0.2);
          }
        `
      }} />
    </div>
  )
}