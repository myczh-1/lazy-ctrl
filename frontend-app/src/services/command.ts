import { request } from '@tarojs/taro'

const BASE_URL = 'http://localhost:7001/api'

export interface Command {
  id: string
  name: string
  description: string
  script_path: string
  args?: string[]
  env?: Record<string, string>
  work_dir?: string
}

export interface ExecutionResult {
  success: boolean
  output: string
  error: string
  exit_code: number
  executed_at: string
}

export class CommandService {
  async getCommands(): Promise<Command[]> {
    const response = await request({
      url: `${BASE_URL}/commands`,
      method: 'GET'
    })
    return response.data
  }

  async getCommand(id: string): Promise<Command> {
    const response = await request({
      url: `${BASE_URL}/commands/${id}`,
      method: 'GET'
    })
    return response.data
  }

  async executeCommand(id: string, args: string[] = []): Promise<ExecutionResult> {
    const response = await request({
      url: `${BASE_URL}/commands/${id}/execute`,
      method: 'POST',
      data: { args }
    })
    return response.data
  }

  async createCommand(command: Omit<Command, 'id'>): Promise<Command> {
    const response = await request({
      url: `${BASE_URL}/commands`,
      method: 'POST',
      data: command
    })
    return response.data
  }

  async updateCommand(id: string, command: Partial<Command>): Promise<Command> {
    const response = await request({
      url: `${BASE_URL}/commands/${id}`,
      method: 'PUT',
      data: command
    })
    return response.data
  }

  async deleteCommand(id: string): Promise<boolean> {
    const response = await request({
      url: `${BASE_URL}/commands/${id}`,
      method: 'DELETE'
    })
    return response.statusCode === 200
  }
}