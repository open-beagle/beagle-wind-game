#!/bin/bash

# 设置变量
BACKEND_PORT=8080
FRONTEND_PORT=5173

# 处理命令行参数
USE_MOCK=false
for arg in "$@"; do
  case $arg in
    --mock)
      USE_MOCK=true
      export VITE_USE_MOCK=true
      shift
      ;;
  esac
done

# 检查端口是否被占用并杀死进程
kill_process_on_port() {
  local port=$1
  local pid=$(lsof -t -i:$port)
  if [ ! -z "$pid" ]; then
    echo "Port $port is in use by PID $pid, killing process..."
    kill -9 $pid
  fi
}

# 清理端口
kill_process_on_port $BACKEND_PORT
kill_process_on_port $FRONTEND_PORT

# 构建后端
echo "Building backend..."
go build -o bin/server cmd/server/main.go

# 创建必要目录
mkdir -p data

# 打印当前模式信息
echo "====================================================="
echo "  启动开发环境"
echo "====================================================="
if [ "$USE_MOCK" = true ]; then
  echo "  • 数据源: Mock数据"
else
  echo "  • 数据源: API数据"
fi
echo "  • 后端服务将在 http://localhost:$BACKEND_PORT 启动"
echo "  • 前端开发服务器将在 http://localhost:$FRONTEND_PORT 启动"
echo "  • API将通过相对路径 /api/v1 访问"
echo "====================================================="
echo "  提示: 可以通过添加 --mock 参数启用Mock数据模式"
echo "  示例: bash ./scripts/dev.sh --mock"
echo "====================================================="

# 在后台启动后端服务器
echo "Starting backend server on port $BACKEND_PORT..."
./bin/server &
BACKEND_PID=$!

# 启动前端开发服务器
echo "Starting frontend development server on port $FRONTEND_PORT..."
cd frontend
npm run dev -- --host 0.0.0.0 &
FRONTEND_PID=$!

# 处理退出信号
trap "echo 'Stopping servers...'; kill $BACKEND_PID $FRONTEND_PID; exit" SIGINT SIGTERM

# 等待两个进程都结束
wait
