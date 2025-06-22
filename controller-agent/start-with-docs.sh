#!/bin/bash

echo "🚀 启动 Lazy-Ctrl Agent with Swagger 文档..."

# 编译应用
echo "📦 编译应用..."
go build -o lazy-ctrl-agent cmd/agent/main.go

if [ $? -eq 0 ]; then
    echo "✅ 编译成功"
    
    echo ""
    echo "🌐 服务将在以下地址启动:"
    echo "   主页面:      http://localhost:7070/web/"
    echo "   API文档:     http://localhost:7070/swagger/index.html"
    echo "   健康检查:    http://localhost:7070/api/v1/health"
    echo "   命令列表:    http://localhost:7070/api/v1/commands"
    echo ""
    
    # 启动服务
    echo "🎯 启动服务..."
    ./lazy-ctrl-agent
else
    echo "❌ 编译失败"
    exit 1
fi