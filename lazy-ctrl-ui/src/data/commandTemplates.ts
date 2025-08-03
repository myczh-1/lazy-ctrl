import type { Command, CommandStep, UIConfig } from '@/types/command'

export interface CommandTemplate {
  templateId: string
  name: string
  description: string
  category: string
  icon: string
  // 为不同平台提供命令
  platforms: {
    windows?: string | CommandStep[]
    linux?: string | CommandStep[]
    darwin?: string | CommandStep[]
    all?: string | CommandStep[]  // 通用命令
  }
  // UI 配置（可选）
  ui?: UIConfig
  isTemplate: true
}

// 预设的命令模板
export const commandTemplates: CommandTemplate[] = [
  // 音频控制
  {
    templateId: 'audio_mute',
    name: '静音',
    description: '切换系统音频静音状态',
    category: 'audio',
    icon: '🔇',
    platforms: {
      windows: 'powershell -c "(New-Object -comObject WScript.Shell).SendKeys([char]173)"',
      linux: 'amixer set Master toggle',
      darwin: 'osascript -e "set volume output muted (not (output muted of (get volume settings)))"'
    },
    isTemplate: true
  },
  {
    templateId: 'audio_volume_up',
    name: '音量+',
    description: '增加系统音量',
    category: 'audio',
    icon: '🔊',
    platforms: {
      windows: 'powershell -c "(New-Object -comObject WScript.Shell).SendKeys([char]175)"',
      linux: 'amixer set Master 5%+',
      darwin: 'osascript -e "set volume output volume (output volume of (get volume settings) + 10)"'
    },
    isTemplate: true
  },
  {
    templateId: 'audio_volume_down',
    name: '音量-',
    description: '降低系统音量',
    category: 'audio',
    icon: '🔉',
    platforms: {
      windows: 'powershell -c "(New-Object -comObject WScript.Shell).SendKeys([char]174)"',
      linux: 'amixer set Master 5%-',
      darwin: 'osascript -e "set volume output volume (output volume of (get volume settings) - 10)"'
    },
    isTemplate: true
  },
  // 新增：音量滑块控制
  {
    templateId: 'audio_volume_set',
    name: '设置音量',
    description: '通过滑块设置系统音量到指定百分比',
    category: 'audio',
    icon: '🎵',
    platforms: {
      windows: 'powershell -c "$volume = {{volume}}; (New-Object -ComObject WScript.Shell).SendKeys([char]173); Start-Sleep -Milliseconds 100; [console]::beep(800, 200)"',
      linux: 'amixer set Master {{volume}}%',
      darwin: 'osascript -e "set volume output volume {{volume}}"'
    },
    ui: {
      type: 'form',
      title: '音量调节',
      description: '拖动滑块设置系统音量',
      preview: true,
      params: [
        {
          key: 'volume',
          label: '音量百分比',
          type: 'range',
          default: 50,
          min: 0,
          max: 100,
          step: 5,
          unit: '%',
          description: '设置系统音量从 0% 到 100%'
        }
      ]
    },
    isTemplate: true
  },

  // 电源管理
  {
    templateId: 'power_lock',
    name: '锁屏',
    description: '锁定工作站屏幕',
    category: 'power',
    icon: '🔒',
    platforms: {
      windows: 'rundll32.exe user32.dll,LockWorkStation',
      linux: 'xdg-screensaver lock || gnome-screensaver-command -l',
      darwin: 'pmset displaysleepnow'
    },
    isTemplate: true
  },
  {
    templateId: 'power_sleep',
    name: '休眠',
    description: '让系统进入休眠状态',
    category: 'power',
    icon: '💤',
    platforms: {
      windows: 'rundll32.exe powrprof.dll,SetSuspendState 0,1,0',
      linux: 'systemctl suspend',
      darwin: 'pmset sleepnow'
    },
    isTemplate: true
  },
  {
    templateId: 'power_shutdown',
    name: '关机',
    description: '安全关闭系统',
    category: 'power',
    icon: '⏻',
    platforms: {
      windows: 'shutdown /s /t 0',
      linux: 'sudo shutdown -h now',
      darwin: 'sudo shutdown -h now'
    },
    isTemplate: true
  },

  // 应用程序控制
  {
    templateId: 'app_notepad',
    name: '记事本',
    description: '打开记事本应用',
    category: 'application',
    icon: '📝',
    platforms: {
      windows: 'notepad.exe',
      linux: 'gedit || nano',
      darwin: 'open -a TextEdit'
    },
    isTemplate: true
  },
  {
    templateId: 'app_calculator',
    name: '计算器',
    description: '打开计算器应用',
    category: 'application',
    icon: '🧮',
    platforms: {
      windows: 'calc.exe',
      linux: 'gnome-calculator || kcalc',
      darwin: 'open -a Calculator'
    },
    isTemplate: true
  },
  {
    templateId: 'app_browser',
    name: '浏览器',
    description: '打开默认浏览器',
    category: 'application',
    icon: '🌐',
    platforms: {
      windows: 'start ""',
      linux: 'xdg-open http://',
      darwin: 'open http://'
    },
    isTemplate: true
  },
  // 新增：自定义打开应用
  {
    templateId: 'app_custom_open',
    name: '打开应用',
    description: '打开指定路径或名称的应用程序',
    category: 'application',
    icon: '🚀',
    platforms: {
      windows: 'start "" "{{appPath}}"',
      linux: '{{appPath}} &',
      darwin: 'open "{{appPath}}"'
    },
    ui: {
      type: 'form',
      title: '应用启动器',
      description: '输入应用程序的路径或名称',
      preview: true,
      params: [
        {
          key: 'appPath',
          label: '应用路径或名称',
          type: 'text',
          placeholder: '例如: notepad.exe, firefox, /Applications/Calculator.app',
          required: true,
          description: '输入应用程序的完整路径或可执行文件名'
        }
      ]
    },
    isTemplate: true
  },
  // 新增：打开网站
  {
    templateId: 'app_open_website',
    name: '打开网站',
    description: '在默认浏览器中打开指定网站',
    category: 'application',
    icon: '🌍',
    platforms: {
      windows: 'start "" "{{url}}"',
      linux: 'xdg-open "{{url}}"',
      darwin: 'open "{{url}}"'
    },
    ui: {
      type: 'form',
      title: '网站快速访问',
      description: '输入网址在浏览器中打开',
      preview: true,
      params: [
        {
          key: 'url',
          label: '网站地址',
          type: 'text',
          placeholder: '例如: https://www.google.com',
          default: 'https://',
          required: true,
          description: '输入完整的 URL 地址（包含 http:// 或 https://）'
        }
      ]
    },
    isTemplate: true
  },

  // 媒体控制
  {
    templateId: 'media_play_pause',
    name: '播放/暂停',
    description: '切换媒体播放状态',
    category: 'media',
    icon: '⏯️',
    platforms: {
      windows: 'powershell -c "(New-Object -comObject WScript.Shell).SendKeys([char]179)"',
      linux: 'playerctl play-pause',
      darwin: 'osascript -e "tell application \\"System Events\\" to key code 49"'
    },
    isTemplate: true
  },
  {
    templateId: 'media_next',
    name: '下一首',
    description: '播放下一首媒体',
    category: 'media',
    icon: '⏭️',
    platforms: {
      windows: 'powershell -c "(New-Object -comObject WScript.Shell).SendKeys([char]176)"',
      linux: 'playerctl next',
      darwin: 'osascript -e "tell application \\"System Events\\" to key code 42"'
    },
    isTemplate: true
  },
  {
    templateId: 'media_previous',
    name: '上一首',
    description: '播放上一首媒体',
    category: 'media',
    icon: '⏮️',
    platforms: {
      windows: 'powershell -c "(New-Object -comObject WScript.Shell).SendKeys([char]177)"',
      linux: 'playerctl previous',
      darwin: 'osascript -e "tell application \\"System Events\\" to key code 43"'
    },
    isTemplate: true
  },

  // 系统工具
  {
    templateId: 'system_screenshot',
    name: '截图',
    description: '截取屏幕截图',
    category: 'system',
    icon: '📸',
    platforms: {
      windows: 'powershell -c "Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.SendKeys]::SendWait(\'{PRTSC}\')"',
      linux: 'gnome-screenshot || scrot',
      darwin: 'screencapture -c'
    },
    isTemplate: true
  },
  {
    templateId: 'system_taskmanager',
    name: '任务管理器',
    description: '打开系统任务管理器',
    category: 'system',
    icon: '📊',
    platforms: {
      windows: 'taskmgr.exe',
      linux: 'gnome-system-monitor || htop',
      darwin: 'open -a "Activity Monitor"'
    },
    isTemplate: true
  },
  {
    templateId: 'system_terminal',
    name: '终端',
    description: '打开命令行终端',
    category: 'system',
    icon: '💻',
    platforms: {
      windows: 'cmd.exe',
      linux: 'gnome-terminal || xterm',
      darwin: 'open -a Terminal'
    },
    isTemplate: true
  },

  // 系统工具 - 延时关机
  {
    templateId: 'system_shutdown_delay',
    name: '延时关机',
    description: '设置延时时间后自动关机',
    category: 'system',
    icon: '⏰',
    platforms: {
      windows: 'shutdown /s /t {{seconds}}',
      linux: 'sudo shutdown -h +{{minutes}}',
      darwin: 'sudo shutdown -h +{{minutes}}'
    },
    ui: {
      type: 'form',
      title: '延时关机设置',
      description: '设置系统在指定时间后自动关机',
      preview: true,
      params: [
        {
          key: 'minutes',
          label: '延时时间',
          type: 'number',
          default: 10,
          min: 1,
          max: 480,
          unit: '分钟',
          description: '设置 1-480 分钟后关机'
        },
        {
          key: 'seconds',
          label: '延时秒数',
          type: 'number',
          default: 600,
          min: 60,
          max: 28800,
          unit: '秒',
          description: 'Windows 系统使用秒数计算'
        }
      ]
    },
    isTemplate: true
  },
  

  // 新增：自定义命令构建器
  {
    templateId: 'custom_command_builder',
    name: '命令构建器',
    description: '自定义创建任意系统命令',
    category: 'custom',
    icon: '🛠️',
    platforms: {
      all: '{{command}}'
    },
    ui: {
      type: 'form',
      title: '自定义命令构建器',
      description: '输入任意系统命令进行执行',
      preview: true,
      params: [
        {
          key: 'command',
          label: '命令内容',
          type: 'textarea',
          placeholder: '输入命令...',
          required: true,
          description: '请谨慎输入命令，确保安全性'
        }
      ]
    },
    isTemplate: true
  },
  
  // 新增：自定义多步骤命令
  {
    templateId: 'multi_step_custom',
    name: '多步骤执行器',
    description: '自定义多个步骤的复杂命令序列',
    category: 'custom',
    icon: '🧩',
    platforms: {
      all: [
        { type: 'shell', cmd: '{{step1}}' },
        { type: 'delay', duration: '{{delay}}' },
        { type: 'shell', cmd: '{{step2}}' }
      ]
    },
    ui: {
      type: 'wizard',
      title: '多步骤命令构建器',
      description: '配置多个步骤的命令序列',
      preview: true,
      params: [
        {
          key: 'step1',
          label: '第一步命令',
          type: 'text',
          placeholder: '输入第一个命令...',
          required: true,
          description: '首先执行的命令'
        },
        {
          key: 'delay',
          label: '间隔时间',
          type: 'number',
          default: 1000,
          min: 100,
          max: 10000,
          unit: '毫秒',
          description: '两个命令之间的等待时间'
        },
        {
          key: 'step2',
          label: '第二步命令',
          type: 'text',
          placeholder: '输入第二个命令...',
          required: true,
          description: '在延时后执行的命令'
        }
      ]
    },
    isTemplate: true
  }
]

// 按分类分组模板
export const getTemplatesByCategory = () => {
  const categories = new Map<string, CommandTemplate[]>()
  
  commandTemplates.forEach(template => {
    if (!categories.has(template.category)) {
      categories.set(template.category, [])
    }
    categories.get(template.category)!.push(template)
  })
  
  return categories
}

// 将模板转换为命令列表（为每个平台生成一个命令）
export const templateToCommands = (template: CommandTemplate, userId = 'local', deviceId = 'default'): Command[] => {
  const commands: Command[] = []
  const timestamp = Date.now()
  
  Object.entries(template.platforms).forEach(([platform, command]) => {
    commands.push({
      id: `${template.templateId}`,
      name: template.name,
      platform: platform === 'all' ? 'all' : platform,
      command,
      category: template.category,
      icon: template.icon,
      description: template.description,
      userId,
      deviceId
    })
  })
  
  return commands
}

// 获取分类信息
export const categoryInfo = {
  audio: { name: '音频控制', icon: '🔊', color: 'bg-blue-500' },
  power: { name: '电源管理', icon: '⚡', color: 'bg-red-500' },
  application: { name: '应用程序', icon: '📱', color: 'bg-green-500' },
  media: { name: '媒体控制', icon: '🎵', color: 'bg-purple-500' },
  system: { name: '系统工具', icon: '⚙️', color: 'bg-gray-500' },
  custom: { name: '自定义', icon: '🛠️', color: 'bg-yellow-500' }
}