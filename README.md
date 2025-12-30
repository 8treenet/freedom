# Freedom DDD 框架

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/8treenet/freedom/blob/master/LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/8treenet/freedom)](https://goreportcard.com/report/github.com/8treenet/freedom)[![GoDoc](https://godoc.org/github.com/8treenet/freedom?status.svg)](https://godoc.org/github.com/8treenet/freedom)
[![GitHub release](https://img.shields.io/github/v/release/8treenet/freedom.svg)](https://github.com/8treenet/freedom/releases)
<img align="right" width="200px" src="https://raw.githubusercontent.com/8treenet/blog/master/img/freedom.png">
## 简介

Freedom 是一个基于六边形架构（Hexagonal Architecture）的 Go 语言框架，专注于支持领域驱动设计（DDD）开发范式。本框架提供了完整的基础设施和工具链，帮助开发者构建可维护、可扩展的企业级应用。

## 核心特性

### 架构支持
- 完整实现六边形架构（端口和适配器模式）
- 领域驱动设计（DDD）最佳实践支持
- 依赖注入（DI）和依赖倒置原则（DIP）
- 完全符合开闭原则的插件化设计

### 框架集成
- 无缝集成 Iris Web 框架
- 支持 HTTP/H2C 服务端和客户端
- 内置 Prometheus 监控集成
- 支持 AOP（面向切面编程）
- 基于 Worker 的无侵入 Context 设计

### 领域模型支持
- 聚合根（Aggregate Root）实现
- 领域事件（Domain Events）支持
- CQS（命令查询分离）模式
- 实体（Entity）和值对象（Value Object）支持

### 数据处理
- 自动化 CRUD 操作生成
- PO（持久化对象）代码生成器
- 多级缓存架构
  - 一级缓存（内存）
  - 二级缓存（分布式）
  - 缓存击穿防护

### 消息和事件
- 集成消息队列组件
- 领域事件发布订阅
- 事件驱动架构支持

## 快速开始

### 安装框架

```bash
# 安装 Freedom 命令行工具
$ go install github.com/8treenet/freedom/freedom@latest

# 验证安装
$ freedom version
```
### 创建新项目

```bash
# 创建项目
$ freedom new-project [项目名称]

# 进入项目目录
$ cd [项目名称]

# 安装依赖
$ go mod tidy

# 运行服务
$ go run main.go
```

### 代码生成工具

```bash
# 生成数据库相关代码（支持两种方式）

# 1. 通过数据库连接生成
$ freedom new-po --dsn "root:密码@tcp(127.0.0.1:3306)/数据库名?charset=utf8"

# 2. 通过 JSON Schema 生成
$ freedom new-po --json ./domain/po/schema.json

# 查看更多生成选项
$ freedom new-po -h
```

## 文档指南

### 核心文档
- **[路由指南](doc/route-guide.md)** - HTTP 路由配置与 API 设计规范
- **[服务指南](doc/service-guide.md)** - 服务层设计原则与业务逻辑实现
- **[持久化对象指南](doc/po-guide.md)** - PO 对象使用说明与数据库操作最佳实践
- **[HTTP 客户端指南](doc/http-client-guide.md)** - HTTP 客户端配置与请求处理
- **[DDD 指南](doc/ddd-guide.md)** - 领域驱动设计实践与架构设计原则
- **[Worker 指南](doc/worker-guide.md)** - Worker 机制详解与 Context 使用说明

## 学习资源

### 示例项目
- **[基础教程](https://github.com/8treenet/freedom/blob/master/example/base)** - DDD 基础概念实践与框架核心功能演示
- **[HTTP2 示例](https://github.com/8treenet/freedom/blob/master/example/http2)** - HTTP2 服务配置与依赖倒置实现
- **[基础设施示例](https://github.com/8treenet/freedom/blob/master/example/infra-example)** - 事务组件、自定义组件开发、Kafka 集成与领域事件
- **[电商系统示例](https://github.com/8treenet/freedom/blob/master/example/fshop)** - 完整电商领域实现，涵盖 CQS 模式、聚合根、实体值对象、领域事件、资源库模式及基础设施层集成

## 贡献指南

欢迎提交 Issue 和 Pull Request 来帮助改进 Freedom 框架。在提交代码前，请确保：

- 代码符合 Go 语言规范
- 添加了必要的测试用例
- 更新了相关文档

## 开源协议

本项目采用 Apache 2.0 开源协议。详见 [LICENSE](https://github.com/8treenet/freedom/blob/master/LICENSE) 文件。

