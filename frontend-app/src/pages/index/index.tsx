import { Component } from 'react'
import { View, Text, Button } from '@tarojs/components'
import { navigateTo } from '@tarojs/taro'
import './index.scss'

export default class Index extends Component {

  componentWillMount () { }

  componentDidMount () { }

  componentWillUnmount () { }

  componentDidShow () { }

  componentDidHide () { }

  goToCommands = () => {
    navigateTo({
      url: '/pages/commands/index'
    })
  }

  render () {
    return (
      <View className='index'>
        <View className='hero-section'>
          <Text className='title'>Lazy Control</Text>
          <Text className='subtitle'>远程控制你的电脑</Text>
        </View>
        
        <View className='quick-actions'>
          <View className='action-card'>
            <Text className='card-title'>快速控制</Text>
            <Text className='card-desc'>执行常用的系统命令</Text>
            <Button className='card-button' onClick={this.goToCommands}>
              查看命令
            </Button>
          </View>
          
          <View className='action-card'>
            <Text className='card-title'>连接状态</Text>
            <Text className='card-desc'>检查与电脑的连接状态</Text>
            <Button className='card-button'>
              检查连接
            </Button>
          </View>
        </View>
      </View>
    )
  }
}