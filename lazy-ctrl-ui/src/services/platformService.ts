import commandAPI from '@/api/commandAPI'

class PlatformService {
  private static instance: PlatformService
  private platformCache: string | null = null
  private fetchPromise: Promise<string> | null = null

  private constructor() {}

  static getInstance(): PlatformService {
    if (!PlatformService.instance) {
      PlatformService.instance = new PlatformService()
    }
    return PlatformService.instance
  }

  /**
   * 获取服务器执行器的平台信息
   * 使用缓存避免重复请求
   */
  async getCurrentPlatform(): Promise<string> {
    // 如果已有缓存，直接返回
    if (this.platformCache) {
      return this.platformCache
    }

    // 如果已有请求在进行中，等待该请求完成
    if (this.fetchPromise) {
      return this.fetchPromise
    }

    // 发起新的请求
    this.fetchPromise = this.fetchPlatformFromAPI()
    
    try {
      this.platformCache = await this.fetchPromise
      return this.platformCache
    } finally {
      this.fetchPromise = null
    }
  }

  /**
   * 从API获取平台信息
   */
  private async fetchPlatformFromAPI(): Promise<string> {
    try {
      return await commandAPI.getPlatform()
    } catch (error) {
      console.error('Failed to fetch platform from API:', error)
      // 如果API失败，返回一个默认值，避免应用崩溃
      return 'linux'
    }
  }

  /**
   * 清除缓存（用于重新获取平台信息）
   */
  clearCache(): void {
    this.platformCache = null
    this.fetchPromise = null
  }

  /**
   * 获取用户友好的平台显示名称
   */
  getPlatformDisplayName(platform: string): string {
    const platformNames: Record<string, string> = {
      'windows': 'Windows',
      'darwin': 'macOS',
      'linux': 'Linux',
      'all': '全平台'
    }
    return platformNames[platform] || platform
  }
}

// 导出单例实例
export const platformService = PlatformService.getInstance()
export default platformService