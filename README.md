# Freedom DDD Framework
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/8treenet/gotree/blob/master/LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/8treenet/freedom)](https://goreportcard.com/report/github.com/8treenet/freedom) [![Build Status](https://travis-ci.org/8treenet/gotree.svg?branch=master)](https://travis-ci.org/8treenet/gotree) [![GoDoc](https://godoc.org/github.com/8treenet/freedom?status.svg)](https://godoc.org/github.com/8treenet/freedom)
[![GitHub release](https://img.shields.io/github/v/release/8treenet/freedom.svg)](https://github.com/8treenet/freedom/releases)
<img align="right" width="200px" src="https://raw.githubusercontent.com/8treenet/blog/master/img/freedom.png">
###### Freedom是一个基于六边形架构的框架，可以支撑充血的领域模型范式。

## Overview
- 集成 Iris
- HTTP/H2C Server & Client
- 集成普罗米修斯
- AOP Worker & 无侵入 Context
- 可扩展组件 Infrastructure
- 依赖注入 & 依赖倒置 & 开闭原则
- DDD & 六边形架构
- 领域事件 & 消息队列组件
- CQS & 聚合根
- CRUD & PO Generate
- 一级缓存 & 二级缓存 & 防击穿

## 安装
```sh
$ go install github.com/8treenet/freedom/freedom@latest
$ freedom version
```

## 脚手架创建项目
```sh
$ freedom new-project [project-name]
$ cd [project-name]
$ go run server/main.go
```

## 脚手架生成增删查改和持久化对象
####
```sh
# freedom new-po -h 查看更多
$ cd [project-name]

# 数据库数据源方式
$ freedom new-po --dsn "root:123123@tcp(127.0.0.1:3306)/freedom?charset=utf8"

# JSON 数据源方式
$ freedom new-po --json ./domain/po/schema.json
```

## Example

#### [基础教程](https://github.com/8treenet/freedom/blob/master/example/base)
#### [http2监听和依赖倒置](https://github.com/8treenet/freedom/blob/master/example/http2)
#### [事务组件&自定义组件&Kafka&领域事件组件](https://github.com/8treenet/freedom/blob/master/example/infra-example)

#### [一个完整的电商demo,包含CQS、聚合、实体、领域事件、资源库、基础设施](https://github.com/8treenet/freedom/blob/master/example/fshop)

