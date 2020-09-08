# freedom
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/8treenet/gotree/blob/master/LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/8treenet/freedom)](https://goreportcard.com/report/github.com/8treenet/freedom) [![Build Status](https://travis-ci.org/8treenet/gotree.svg?branch=master)](https://travis-ci.org/8treenet/gotree) [![GoDoc](https://godoc.org/github.com/8treenet/freedom?status.svg)](https://godoc.org/github.com/8treenet/freedom)
<img align="right" width="200px" src="https://raw.githubusercontent.com/8treenet/blog/master/img/freedom.png">
###### freedom是一个基于六边形架构的框架，可以支撑充血的领域模型范式。

## Overview
- 集成iris v12
- 集成普罗米修斯
- 无侵入式Trace
- 基于组件式的基础设施扩展
- http2 server & client
- 依赖注入 & 依赖倒置
- CRUD & PO 代码生成
- DDD & 六边形架构
- 消息事件 & 事件重试 & 领域事件
- 一级缓存 & 二级缓存 & 防击穿

## 安装
```sh
$ go get github.com/8treenet/freedom/freedom
```

## 脚手架创建项目
```sh
$ freedom new-project [project-name]
```

## 脚手架生成增删查改和持久化对象
####
```sh
# 编辑 [project-name]/cmd/conf/db.toml 填入数据库地址
# freedom new-po -h 查看更多
$ cd [project-name]
$ freedom new-po
```

## Example

#### [基础教程](https://github.com/8treenet/freedom/blob/master/example/base)
#### [http2监听和依赖倒置](https://github.com/8treenet/freedom/blob/master/example/http2)
#### [事务组件和自定义组件](https://github.com/8treenet/freedom/blob/master/example/infra-example)
#### [消息事件和领域事件](https://github.com/8treenet/freedom/blob/master/example/event-example)

#### [一个完整的电商demo,包含CQRS、聚合、实体、领域事件、资源库、基础设施](https://github.com/8treenet/freedom/blob/master/example/fshop)

