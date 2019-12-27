# freedom
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/8treenet/gotree/blob/master/LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/8treenet/tcp)](https://goreportcard.com/report/github.com/8treenet/tcp) [![Build Status](https://travis-ci.org/8treenet/gotree.svg?branch=master)](https://travis-ci.org/8treenet/gotree) [![GoDoc](https://godoc.org/github.com/8treenet/gotree?status.svg)](https://godoc.org/github.com/8treenet/gotree)
###### freedom-微服务框架。

## Overview
- 集成iris
- 集成普罗米修斯
- 集成gorm
- 集成gcache
- 链路追踪
- infra容器 基于组件的基础设施
- http2 server
- http2 client
- 依赖注入
- CRUD 代码生成
- 消息总线

## 进行中
- XA强一致分布事务
- 最终一致分布事务

## 安装
```sh
$ go get github.com/8treenet/freedom/freedom
```

## 创建项目
```sh
$ freedom new-project [project-name]
```

## 生成crud
```sh
# 编辑 [project-name]/cmd/conf/db.toml 填入数据库地址
# 可指定配置地址和输出目录 freedom new-crud -h 查看更多
$ cd [project-name]
$ freedom new-crud
```

## Example

#### [基础示例](https://github.com/8treenet/freedom/blob/master/example/base)
#### [http2监听和依赖倒置](https://github.com/8treenet/freedom/blob/master/example/http2)
#### [repository和事务组件](https://github.com/8treenet/freedom/blob/master/example/infra-example)
#### [消息总线](https://github.com/8treenet/freedom/blob/master/example/msg-example)
#### [DDD领域模型示例](https://github.com/8treenet/freedom/blob/master/example/ddd-example)