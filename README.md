# freedom
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/8treenet/gotree/blob/master/LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/8treenet/tcp)](https://goreportcard.com/report/github.com/8treenet/tcp) [![Build Status](https://travis-ci.org/8treenet/gotree.svg?branch=master)](https://travis-ci.org/8treenet/gotree) [![GoDoc](https://godoc.org/github.com/8treenet/gotree?status.svg)](https://godoc.org/github.com/8treenet/gotree)
###### freedom是一个基于六边形架构的框架，可以支撑充血的领域模型范式。

## Overview
- 集成iris
- 集成普罗米修斯
- 集成gorm
- 链路追踪
- infra容器 基于组件的基础设施
- http2 server
- http2 client
- 依赖注入&依赖倒置
- CRUD 代码生成
- DDD&六边形架构
- 消息事件

## 进行中
- aggregate最终一致
- aggregate强一致

## 安装
```sh
$ go get github.com/8treenet/freedom/freedom
```

## 创建项目
```sh
$ freedom new-project [project-name]
```

## 生成crud 值对象
```sh
# 编辑 [project-name]/cmd/conf/db.toml 填入数据库地址
# 可指定配置地址和输出目录 freedom new-crud -h 查看更多
$ cd [project-name]
$ freedom new-crud
```

## Example

#### [基础教程](https://github.com/8treenet/freedom/blob/master/example/base)
#### [http2监听和依赖倒置](https://github.com/8treenet/freedom/blob/master/example/http2)
#### [repository和事务组件](https://github.com/8treenet/freedom/blob/master/example/infra-example)
#### [消息事件和领域事件](https://github.com/8treenet/freedom/blob/master/example/event-example)
#### [aggregate和entity](https://github.com/8treenet/freedom/blob/master/example/ddd-example)