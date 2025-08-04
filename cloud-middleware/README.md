# Lazy-Ctrl Cloud Middleware

云端中间件服务，为 lazy-ctrl 系统提供用户认证、设备管理和命令转发功能。

## 架构设计

```
移动端/Web客户端 ← HTTP/WebSocket → 云端中间件 ← gRPC → 家庭设备
                                      ↕
                                   PostgreSQL
                                      ↕
                                     Redis
```

## 功能特性

### 用户管理
- 用户注册/登录/注销
- JWT Token 认证
- 用户配置管理
- 密码修改

### 设备管理
- 设备注册和绑定
- 设备状态监控
- 多用户设备共享
- 设备权限控制

### 命令转发
- HTTP API → gRPC 转换
- 实时命令执行
- 命令历史记录
- 执行权限验证

### 系统管理
- 健康检查
- 系统监控
- 日志记录
- 配置管理

## API 接口

### 用户认证 API
- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录
- `POST /api/v1/auth/refresh` - 刷新令牌
- `POST /api/v1/auth/logout` - 用户注销

### 用户管理 API
- `GET /api/v1/user/profile` - 获取用户信息
- `PUT /api/v1/user/profile` - 更新用户信息
- `POST /api/v1/user/change-password` - 修改密码

### 设备管理 API
- `POST /api/v1/device/bind` - 绑定设备
- `DELETE /api/v1/device/:device_id` - 解绑设备
- `GET /api/v1/device/list` - 获取设备列表
- `PUT /api/v1/device/:device_id` - 更新设备信息

### 网关转发 API
- `POST /api/v1/gateway/commands` - 创建命令
- `PUT /api/v1/gateway/commands/:command_id` - 更新命令
- `DELETE /api/v1/gateway/commands/:command_id` - 删除命令
- `GET /api/v1/gateway/commands/:command_id` - 获取命令详情
- `GET /api/v1/gateway/commands` - 获取命令列表
- `GET /api/v1/gateway/commands/homepage` - 获取首页命令
- `POST /api/v1/gateway/execute` - 执行命令
- `GET /api/v1/gateway/health/:device_id` - 设备健康检查

## 配置文件

### 配置文件示例 (configs/config.yaml)
```yaml
server:
  port: 8080
  read_timeout: 30
  write_timeout: 30
  idle_timeout: 120

grpc:
  port: 8081

database:
  host: localhost
  port: 5432
  user: postgres
  password: your_password
  dbname: lazy_ctrl_cloud
  sslmode: disable
  timezone: Asia/Shanghai

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

jwt:
  secret_key: your-secret-key-change-in-production
  access_token_duration: 15   # minutes
  refresh_token_duration: 7   # days

log:
  level: info
  format: json
  output: stdout
```

## 数据库设计

### 主要表结构
- `users` - 用户表
- `devices` - 设备表
- `user_devices` - 用户设备关联表
- `device_commands` - 设备命令表
- `execution_logs` - 执行日志表

## gRPC 服务

### GatewayService
设备命令管理和执行服务，映射本地 Controller Agent 的所有 HTTP API。

### UserService  
用户管理服务，提供完整的用户生命周期管理。

## 部署说明

### 环境要求
- Go 1.21+
- PostgreSQL 13+
- Redis 6+

### 编译运行
```bash
# 安装依赖
go mod tidy

# 编译
go build -o bin/cloud-server cmd/server/main.go

# 运行
./bin/cloud-server -config configs/config.yaml
```

### Docker 部署
```bash
# 构建镜像
docker build -t lazy-ctrl-cloud .

# 运行容器
docker run -d \
  -p 8080:8080 \
  -p 8081:8081 \
  -v $(pwd)/configs:/app/configs \
  --name lazy-ctrl-cloud \
  lazy-ctrl-cloud
```

## 开发指南

### 项目结构
```
cloud-middleware/
├── cmd/server/          # 服务入口
├── configs/            # 配置文件
├── internal/
│   ├── app/            # 应用初始化
│   ├── config/         # 配置管理
│   ├── database/       # 数据库连接
│   ├── handler/        # HTTP/gRPC 处理器
│   ├── middleware/     # 中间件
│   ├── model/          # 数据模型
│   ├── repository/     # 数据访问层
│   └── service/        # 业务逻辑层
├── proto/              # Protocol Buffer 定义
└── migrations/         # 数据库迁移文件
```

### 添加新功能
1. 在 `proto/` 目录定义 gRPC 接口
2. 在 `model/` 目录定义数据模型
3. 在 `repository/` 目录实现数据访问
4. 在 `service/` 目录实现业务逻辑
5. 在 `handler/` 目录实现 HTTP/gRPC 处理器
6. 在 `app/app.go` 中注册路由和服务

## 安全考虑

- JWT Token 使用 HMAC-SHA256 签名
- 数据库连接使用参数化查询防止 SQL 注入
- API 接口进行权限验证
- 敏感信息不记录到日志
- 支持 HTTPS 和 gRPC TLS 加密