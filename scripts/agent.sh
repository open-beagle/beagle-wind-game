#!/bin/bash

# 设置环境变量
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# 进入项目根目录
cd "$PROJECT_ROOT" || exit 1

# 解析命令行参数
SERVER_ADDR="localhost:50051"
LOG_LEVEL="INFO"
CUSTOM_NODE_ID=""

# 显示帮助信息
show_help() {
  echo "使用方法: $0 [选项]"
  echo "选项:"
  echo "  --help                显示帮助信息"
  echo "  --server-addr ADDR    设置服务器地址 (默认: localhost:50051)"
  echo "  --log-level LEVEL     设置日志级别: DEBUG, INFO, WARN, ERROR, FATAL (默认: INFO)"
  echo "  --node-id ID          设置自定义节点ID (默认: 自动生成)"
  exit 1
}

# 解析参数
while [[ $# -gt 0 ]]; do
  case "$1" in
  --help)
    show_help
    ;;
  --server-addr)
    SERVER_ADDR="$2"
    shift 2
    ;;
  --log-level)
    LOG_LEVEL="$2"
    shift 2
    ;;
  --node-id)
    CUSTOM_NODE_ID="$2"
    shift 2
    ;;
  *)
    echo "未知选项: $1"
    show_help
    ;;
  esac
done

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
if [ -z "$CUSTOM_NODE_ID" ]; then
  # 自动生成节点ID (使用主机名)
  NODE_ID="node-$(hostname)"
else
  # 使用自定义节点ID
  NODE_ID="$CUSTOM_NODE_ID"
fi

# 启动 agent
echo "========================================"
echo "启动 agent..."
echo "节点 ID: $NODE_ID"
echo "服务器地址: $SERVER_ADDR"
echo "日志级别: $LOG_LEVEL"
echo "========================================"

# 在前台运行 agent，同时输出到控制台和日志文件
./bin/agent \
  --node-id="$NODE_ID" \
  --server-addr="$SERVER_ADDR" \
  --log-level="$LOG_LEVEL" \
  --log-file="logs/agent.log" \
  --log-both=false
