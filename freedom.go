package freedom

import (
	"github.com/8treenet/freedom/internal"
	"github.com/kataras/iris/v12/hero"
	"github.com/kataras/iris/v12/mvc"

	"github.com/kataras/golog"
	iris "github.com/kataras/iris/v12"
)

func init() {
	initApp()
	initConfigurator()
}

type (
	// IrisResult represents an type alias to hero.Result
	IrisResult = hero.Result

	// IrisContext represents an type alias to iris.Context
	IrisContext = iris.Context

	// IrisBeforeActivation represents an type alias to mvc.BeforeActivation
	IrisBeforeActivation = mvc.BeforeActivation

	// IrisConfiguration represents an type alias to iris.Configuration
	IrisConfiguration = iris.Configuration
)

type (
	// Worker .
	Worker = internal.Worker

	// Initiator .
	Initiator = internal.Initiator

	//Repository .
	Repository = internal.Repository

	//Infra .
	Infra = internal.Infra

	//SingleBoot .
	SingleBoot = internal.SingleBoot

	//Entity is the entity's father interface.
	Entity = internal.Entity

	// DomainEvent represents a domain event
	DomainEvent = internal.DomainEvent

	//UnitTest is a unit test tool.
	UnitTest = internal.UnitTest

	//Starter is the startup interface.
	Starter = internal.Starter

	// Bus is the bus message type.
	Bus = internal.Bus

	// BusHandler is the bus message middleware type.
	BusHandler = internal.BusHandler

	//Result is the controller return type.
	Result = IrisResult

	//Context is the context type.
	Context = IrisContext

	// Configuration is the configuration type of the app.
	Configuration = IrisConfiguration

	// BeforeActivation is Is the start-up pre-processing of the action.
	BeforeActivation = IrisBeforeActivation

	// LogFields is the column type of the log.
	LogFields = golog.Fields

	// LogRow is the log per line callback.
	LogRow = golog.Log
)

// Prepare .
func Prepare(f func(Initiator)) {
	internal.Prepare(f)
}

// NewUnitTest .
func NewUnitTest() UnitTest {
	return internal.NewUnitTest()
}

// DefaultConfiguration proxy a call to iris.DefaultConfiguration
func DefaultConfiguration() iris.Configuration {
	return iris.DefaultConfiguration()
}
