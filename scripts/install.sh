#!/bin/bash

# 安装前端依赖
echo "Installing frontend dependencies..."
cd frontend
npm install
cd ..

# 安装后端依赖
echo "Installing backend dependencies..."
cd backend
go mod tidy
cd ..

echo "Dependencies installation completed!" 