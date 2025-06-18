# Lazy Control System

一个用于远程控制本地计算机的分布式系统，支持本地和云端控制模式。

## 系统架构

### 1. 控制器 Agent (Go)
- **位置**: 本地电脑
- **功能**: 执行脚本、提供通信接口
- **端口**: 50051 (gRPC)

### 2. 配置服务 (Node.js + Midway)
- **位置**: 本地电脑
- **功能**: 管理脚本配置、提供 REST API
- **端口**: 7001 (HTTP)

### 3. 前端应用 (Taro)
- **位置**: 移动端/Web
- **功能**: 用户界面、控制操作
- **支持**: 微信小程序、H5

### 4. 云端服务 (Node.js, 可选)
- **位置**: 云服务器
- **功能**: 用户认证、消息转发
- **端口**: 3000 (HTTP + WebSocket)

## 快速开始

### 1. 启动控制器 Agent

```bash
cd controller-agent
go mod tidy
go run cmd/main.go --config configs/commands.json
```

### 2. 启动配置服务

```bash
cd config-service
npm install
npm run dev
```

### 3. 启动前端应用

```bash
cd frontend-app
npm install
# 微信小程序
npm run dev:weapp
# H5
npm run dev:h5
```

### 4. 启动云端服务 (可选)

```bash
cd cloud-service
npm install
npm run dev
```

## 项目结构

```
lazy-ctrl/
├── controller-agent/          # Go 控制器
│   ├── cmd/                  # 入口文件
│   ├── pkg/                  # 核心包
│   │   ├── config/          # 配置管理
│   │   ├── executor/        # 命令执行器
│   │   └── server/          # gRPC 服务器
│   ├── api/proto/           # Protocol Buffers
│   └── configs/             # 配置文件
│
├── config-service/           # Node.js 配置服务
│   ├── src/
│   │   ├── controller/      # 控制器
│   │   ├── service/         # 业务逻辑
│   │   └── config/          # 配置
│   └── proto/               # gRPC 协议文件
│
├── frontend-app/             # Taro 前端应用
│   ├── src/
│   │   ├── pages/           # 页面
│   │   ├── components/      # 组件
│   │   └── services/        # API 服务
│   └── config/              # 构建配置
│
├── cloud-service/            # 云端服务 (可选)
│   ├── src/
│   │   ├── controller/      # 路由控制器
│   │   ├── service/         # 业务服务
│   │   ├── model/           # 数据模型
│   │   └── middleware/      # 中间件
│   └── config/              # 配置文件
│
└── README.md
```

## 功能特性

### 本地控制模式
- 直接通过局域网控制本地设备
- 无需云端服务，更快响应
- 支持常用系统命令（静音、关机、休眠等）

### 云端控制模式 (可选)
- 通过互联网远程控制
- 支持多用户、多设备管理
- 用户认证和权限控制
- 实时状态同步

### 安全特性
- 脚本路径白名单限制
- 可配置的命令执行权限
- 可选的 PIN 码验证
- JWT 令牌认证 (云端模式)

## 配置说明

### 命令配置示例

```json
{
  "commands": [
    {
      "id": "mute",
      "name": "静音/取消静音",
      "description": "切换系统音频静音状态",
      "script_path": "/usr/bin/osascript",
      "args": ["-e", "set volume output muted not (output muted of (get volume settings))"]
    }
  ],
  "security": {
    "allowed_paths": ["/usr/bin/", "/sbin/"],
    "whitelist": ["mute", "shutdown", "sleep"],
    "require_auth": false
  }
}
```

## API 接口

### 配置服务 API (本地)
- `GET /api/commands` - 获取命令列表
- `POST /api/commands` - 创建命令
- `PUT /api/commands/:id` - 更新命令
- `DELETE /api/commands/:id` - 删除命令
- `POST /api/commands/:id/execute` - 执行命令

### 云端服务 API (可选)
- `POST /api/auth/login` - 用户登录
- `GET /api/devices` - 获取设备列表
- `POST /api/devices/:id/commands` - 执行设备命令

## 环境要求

- Go 1.21+
- Node.js 16+
- MongoDB (云端服务)

## 许可证

MIT License