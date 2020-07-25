# freedom
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/8treenet/gotree/blob/master/LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/8treenet/tcp)](https://goreportcard.com/report/github.com/8treenet/tcp) [![Build Status](https://travis-ci.org/8treenet/gotree.svg?branch=master)](https://travis-ci.org/8treenet/gotree) [![GoDoc](https://godoc.org/github.com/8treenet/gotree?status.svg)](https://godoc.org/github.com/8treenet/gotree)
<img align="right" width="200px" src="https://raw.githubusercontent.com/8treenet/blog/master/img/freedom.png">
###### Freedom is a framework based on a hexagonal architecture that supports the congestion domain model paradigm.

## Overview
- Integrated Iris v12
- Integrated Prometheus
- Integrated Gorm
- Link Tracing
- Infra Container, Component-based Infrastructure
- Http2 Server & Client
- Dependency Injection & Dependency Inversion
- CRUD Automatic Code Generation
- DDD & Hexagonal Architecture
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
# Vim [project-name]/cmd/conf/db.toml -- Fill in database address
# Configurable address and output directory, using 'freedom new-po -h' to see more
$ cd [project-name]
$ freedom new-po
```

## Example

#### [Basic Tutorial](https://github.com/8treenet/freedom/blob/master/example/base)
#### [Http2 Listening And Dependency Inversion](https://github.com/8treenet/freedom/blob/master/example/http2)
#### [Transaction Components And Custom Components](https://github.com/8treenet/freedom/blob/master/example/infra-example)
#### [Message Events And Domain Events](https://github.com/8treenet/freedom/blob/master/example/event-example)
#### [Electronic Demo(Contains CQRS、Aggregation、entity、Domain Events、Repository、Infrastructure)](https://github.com/8treenet/freedom/blob/master/example/fshop)

