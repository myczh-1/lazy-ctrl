// 新的命令配置格式
export interface Command {
  id: string
  name: string
  platform: string  // 'windows' | 'linux' | 'darwin' | 'all'
  command: string | CommandStep[]  // 单个命令或多步骤命令
  category: string
  icon: string
  description?: string
  userId?: string
  deviceId?: string
}

// 多步骤命令支持
export interface CommandStep {
  type: 'shell' | 'delay' | 'script'
  cmd?: string
  duration?: number
  script?: string
}

// 用于展示的聚合命令（多平台聚合）
export interface DisplayCommand {
  id: string
  name: string
  description?: string
  icon: string
  category: string
  platforms: Record<string, string | CommandStep[]>  // 平台 -> 命令映射
  commands: Command[]  // 原始命令列表
  templateId?: string  // 模板ID（如果来自模板）
}

export interface CommandCategory {
  id: string
  name: string
  icon?: string
  color?: string
}

// UI 参数配置
export interface UIParam {
  key: string        // 参数键名，用于替换命令中的占位符
  label: string      // UI显示的标签
  type: 'number' | 'text' | 'range' | 'file' | 'select' | 'textarea'
  default?: any      // 默认值
  min?: number       // 最小值（数字/滑块）
  max?: number       // 最大值（数字/滑块）
  step?: number      // 步长（滑块）
  options?: Array<{label: string, value: string}> // 选择项（下拉框）
  placeholder?: string // 输入提示
  required?: boolean  // 是否必填
  unit?: string      // 单位显示
  description?: string // 参数说明
}

// UI 配置
export interface UIConfig {
  type: 'form' | 'wizard' // 表单类型
  title?: string          // 自定义标题
  description?: string    // 自定义描述
  params: UIParam[]       // 参数列表
  preview?: boolean       // 是否显示命令预览
}