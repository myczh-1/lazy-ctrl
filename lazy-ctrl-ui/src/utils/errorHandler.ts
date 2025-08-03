import type { ExecutionResult } from '@/api/commandAPI'

/**
 * 错误处理函数 - 提供友好的错误提示
 * @param result 命令执行结果
 * @returns 友好的错误提示信息
 */
export const getExecutionErrorMessage = (result: ExecutionResult): string => {
  // 检查退出码和输出信息，提供具体的错误提示
  if (result.exit_code === 127) {
    // 命令未找到
    if (result.output?.includes('amixer: not found')) {
      return '音频控制工具 amixer 未安装。请运行以下命令安装：\nsudo apt-get install alsa-utils'
    }
    
    // 通用命令未找到错误
    const missingCommand = result.output?.match(/([^:]+): not found/)?.[1]
    if (missingCommand) {
      return `命令 "${missingCommand}" 未找到，请检查是否已安装相关工具`
    }
    
    return '命令未找到，请检查命令是否正确或相关工具是否已安装'
  }
  
  if (result.exit_code === 126) {
    return '命令无执行权限，请检查文件权限或使用 sudo 运行'
  }
  
  if (result.exit_code === 1) {
    return '命令执行失败，请检查命令参数是否正确'
  }
  
  // 检查是否是超时错误
  if (result.output?.includes('timeout') || result.error?.includes('timeout')) {
    return '命令执行超时，请检查命令是否需要更长的执行时间'
  }
  
  // 返回原始错误信息
  return result.error || result.output || '命令执行失败，未知错误'
}

/**
 * 常见错误码说明
 */
export const ERROR_CODE_DESCRIPTIONS = {
  0: '命令执行成功',
  1: '一般性错误',
  2: '误用命令',
  126: '命令不可执行',
  127: '命令未找到',
  128: '无效退出参数',
  129: '致命错误信号1',
  130: '致命错误信号2 (Ctrl+C)',
  255: '退出状态码超出范围'
} as const