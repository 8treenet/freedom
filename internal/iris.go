package internal

import (
	"github.com/8treenet/iris/v12"
	"github.com/8treenet/iris/v12/context"
	"github.com/8treenet/iris/v12/core/host"
	"github.com/8treenet/iris/v12/mvc"
)

type (
	// IrisApplication is a type alias to iris.Application
	IrisApplication = iris.Application

	// IrisContext is a type alias to iris.Context
	IrisContext = iris.Context

	// IrisRunner is a type alias to iris.Runner
	IrisRunner = iris.Runner

	// IrisConfiguration is a type alias to iris.Configuration.
	IrisConfiguration = iris.Configuration

	// IrisHostConfigurator is a type alias to host.Configurator
	IrisHostConfigurator = host.Configurator

	// IrisParty is a type alias to iris.Party.
	IrisParty = iris.Party

	// IrisHandler is a type alias to context.Handler.
	IrisHandler = context.Handler

	// IrisMVCApplication is a type alias to mvc.Application.
	IrisMVCApplication = mvc.Application

	// IrisController represents a standard iris controller.
	IrisController = interface{}
)
