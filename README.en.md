# freedom
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/8treenet/gotree/blob/master/LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/8treenet/tcp)](https://goreportcard.com/report/github.com/8treenet/tcp) [![Build Status](https://travis-ci.org/8treenet/gotree.svg?branch=master)](https://travis-ci.org/8treenet/gotree) [![GoDoc](https://godoc.org/github.com/8treenet/gotree?status.svg)](https://godoc.org/github.com/8treenet/gotree)
###### Freedom is a DDD Web framework, which could support Hexagonal Architecture and paradigm of rich models.

## Overview
- Based on [Iris](https://iris-go.com/)
- [Prometheus](https://prometheus.io) Integration
- [GORM](https://gorm.io/)
- Tracing
- Component-based container of infrastructure
- HTT2 Server & Client
- Dependency Injection
- Dependency Inversion
- CRUD Template Code Generation
- Events of Message & Domain, Auto Retry...

## Install
```sh
$ go get github.com/8treenet/freedom/freedom
```

## Create Project
```sh
$ freedom new-project [project-name]
```

## Generate CRUD Objects & Values
```sh
# Edit [project-name]/cmd/conf/db.toml # Fill database connetion
# Configurable configure path & output path
# freedom new-crud -h # To get more help
$ cd [project-name]
$ freedom new-crud
```

## Example

#### [Tutorial](https://github.com/8treenet/freedom/blob/master/example/base)
#### [HTTP2 Listening & Dependency Inversion](https://github.com/8treenet/freedom/blob/master/example/http2)
#### [Repository & Transaction component](https://github.com/8treenet/freedom/blob/master/example/infra-example)
#### [Message events & Domain events](https://github.com/8treenet/freedom/blob/master/example/event-example)
#### [ddd-example](https://github.com/8treenet/freedom/blob/master/example/ddd-example)
###### 一个完整的电商项目,包含CQRS、聚合、实体、领域事件、仓库、基础设施。
###### A complete e-commerce project, including CQRS, aggregation, entities, domain events, warehouses, and infrastructure.


## Todo
- XA Transaction
