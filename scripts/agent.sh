#!/bin/bash

# 设置环境变量
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# 进入项目根目录
cd "$PROJECT_ROOT" || exit 1

# 检查是否已经安装依赖
if [ ! -d "vendor" ]; then
  echo "安装依赖..."
  go mod vendor
fi

# 编译 agent
echo "编译 agent..."
go build -o bin/agent cmd/agent/main.go

# 检查编译是否成功
if [ $? -ne 0 ]; then
  echo "编译失败"
  exit 1
fi

# 创建日志目录
mkdir -p logs

# 设置 agent 参数
NODE_ID="node-$(hostname)"
SERVER_ADDR="localhost:50051"

# 启动 agent
echo "启动 agent..."
echo "Node ID: $NODE_ID"
echo "Server Address: $SERVER_ADDR"

# 在前台运行 agent
sudo ./bin/agent --node-id="$NODE_ID" --server-addr="$SERVER_ADDR"
