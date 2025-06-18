import { Component } from 'react'
import { View, Text, Button } from '@tarojs/components'
import { showToast, showLoading, hideLoading } from '@tarojs/taro'
import { CommandService } from '../../services/command'
import './index.scss'

interface Command {
  id: string
  name: string
  description: string
}

interface State {
  commands: Command[]
  executionResult: any
  loading: boolean
}

export default class Commands extends Component<{}, State> {
  private commandService: CommandService

  constructor(props) {
    super(props)
    this.state = {
      commands: [],
      executionResult: null,
      loading: false
    }
    this.commandService = new CommandService()
  }

  componentDidMount() {
    this.loadCommands()
  }

  loadCommands = async () => {
    try {
      this.setState({ loading: true })
      const commands = await this.commandService.getCommands()
      this.setState({ commands })
    } catch (error) {
      showToast({
        title: '加载命令失败',
        icon: 'error'
      })
    } finally {
      this.setState({ loading: false })
    }
  }

  executeCommand = async (commandId: string) => {
    try {
      showLoading({ title: '执行中...' })
      const result = await this.commandService.executeCommand(commandId)
      this.setState({ executionResult: result })
      
      if (result.success) {
        showToast({
          title: '执行成功',
          icon: 'success'
        })
      } else {
        showToast({
          title: '执行失败',
          icon: 'error'
        })
      }
    } catch (error) {
      showToast({
        title: '执行失败',
        icon: 'error'
      })
    } finally {
      hideLoading()
    }
  }

  render() {
    const { commands, executionResult, loading } = this.state

    return (
      <View className='commands'>
        <View className='header'>
          <Text className='title'>可用命令</Text>
          <Button className='refresh-btn' onClick={this.loadCommands} loading={loading}>
            刷新
          </Button>
        </View>

        {executionResult && (
          <View className='result-card'>
            <Text className='result-title'>执行结果</Text>
            <View className={`result-status ${executionResult.success ? 'success' : 'error'}`}>
              {executionResult.success ? '成功' : '失败'}
            </View>
            {executionResult.output && (
              <View className='result-output'>
                <Text className='output-label'>输出:</Text>
                <Text className='output-content'>{executionResult.output}</Text>
              </View>
            )}
            {executionResult.error && (
              <View className='result-error'>
                <Text className='error-label'>错误:</Text>
                <Text className='error-content'>{executionResult.error}</Text>
              </View>
            )}
          </View>
        )}

        <View className='command-list'>
          {commands.map(command => (
            <View key={command.id} className='command-item'>
              <Text className='command-name'>{command.name}</Text>
              <Text className='command-desc'>{command.description}</Text>
              <Button 
                className='command-button' 
                onClick={() => this.executeCommand(command.id)}
              >
                执行
              </Button>
            </View>
          ))}
        </View>

        {commands.length === 0 && !loading && (
          <View className='empty-state'>
            <Text>暂无可用命令</Text>
          </View>
        )}
      </View>
    )
  }
}