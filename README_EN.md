# Freedom DDD Framework
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/8treenet/gotree/blob/master/LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/8treenet/freedom)](https://goreportcard.com/report/github.com/8treenet/freedom) [![Build Status](https://travis-ci.org/8treenet/gotree.svg?branch=master)](https://travis-ci.org/8treenet/gotree) [![GoDoc](https://godoc.org/github.com/8treenet/freedom?status.svg)](https://godoc.org/github.com/8treenet/freedom)
[![GitHub release](https://img.shields.io/github/v/release/8treenet/freedom.svg)](https://github.com/8treenet/freedom/releases)
<img align="right" width="200px" src="https://raw.githubusercontent.com/8treenet/blog/master/img/freedom.png">
###### Freedom is a framework based on a hexagonal architecture that supports the congestion domain model paradigm.

## Overview
- Integrated Iris v12
- Integrated Prometheus
- Link Tracing
- Infra Container, Component-based Infrastructure
- Http2 Server & Client
- Dependency Injection & Dependency Inversion
- CRUD Automatic Code Generation
- DDD & Hexagonal Architecture
- Domain Events & MQ Infrastructure
- CQS & Aggregation
- Message Events & Event Retries & Domain Events
- Primary Cache & Secondary Cache & Prevent Breakdown

## Install
```sh
$ go get github.com/8treenet/freedom/freedom
```

## Create Project
```sh
$ freedom new-project [project-name]
```

## Build Persistent Objects(PO)
```sh
# Configurable address and output directory, using 'freedom new-po -h' to see more
$ cd [project-name]

# DB shcema
$ freedom new-po --dsn "root:123123@tcp(127.0.0.1:3306)/freedom?charset=utf8"

# JSON shcema
$ freedom new-po --json ./domain/po/shcema.json
```

## Example

#### [Basic Tutorial](https://github.com/8treenet/freedom/blob/master/example/base)
#### [Http2 Listening And Dependency Inversion](https://github.com/8treenet/freedom/blob/master/example/http2)
#### [Transaction Components And Custom Components](https://github.com/8treenet/freedom/blob/master/example/infra-example)
#### [Message Events And Domain Events](https://github.com/8treenet/freedom/blob/master/example/event-example)
#### [Electronic Demo(Contains CQS、Aggregation、entity、Domain Events、Repository、Infrastructure)](https://github.com/8treenet/freedom/blob/master/example/fshop)

