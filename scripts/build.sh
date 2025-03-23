#!/bin/bash

# 进入前端目录
cd frontend

# 清理之前的构建
echo "Cleaning previous build..."
rm -rf dist

# 检查依赖
echo "Checking dependencies..."
npm install

# 执行构建并记录日志
echo "Building frontend..."
npm run build 2>&1 | tee build.log

# 检查构建结果
if [ $? -eq 0 ]; then
    echo "Build completed successfully!"
else
    echo "Build failed! Check build.log for details."
    exit 1
fi 