export default defineAppConfig({
  pages: [
    'pages/index/index',
    'pages/commands/index',
    'pages/settings/index'
  ],
  window: {
    backgroundTextStyle: 'light',
    navigationBarBackgroundColor: '#fff',
    navigationBarTitleText: 'Lazy Control',
    navigationBarTextStyle: 'black'
  },
  tabBar: {
    color: '#666',
    selectedColor: '#b4282d',
    backgroundColor: '#fafafa',
    borderStyle: 'black',
    list: [
      {
        pagePath: 'pages/index/index',
        iconPath: 'assets/home.png',
        selectedIconPath: 'assets/home-active.png',
        text: '首页'
      },
      {
        pagePath: 'pages/commands/index',
        iconPath: 'assets/command.png',
        selectedIconPath: 'assets/command-active.png',
        text: '命令'
      },
      {
        pagePath: 'pages/settings/index',
        iconPath: 'assets/settings.png',
        selectedIconPath: 'assets/settings-active.png',
        text: '设置'
      }
    ]
  }
})