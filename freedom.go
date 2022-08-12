package freedom

import (
	"github.com/8treenet/freedom/internal"
	"github.com/8treenet/iris/v12/hero"
	"github.com/8treenet/iris/v12/mvc"

	iris "github.com/8treenet/iris/v12"
	"github.com/kataras/golog"
)

func init() {
	initApp()
	initConfigurator()
}

const (
	//TransactionKey for transaction DB handles in the first-level cache
	TransactionKey = internal.TransactionKey
)

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
	// Worker Request a runtime object.
	Worker = internal.Worker

	// Initiator Initialize the binding relationship between the user's code and the associated code.
	Initiator = internal.Initiator

	// Repository The parent class of the repository, which can be inherited using the parent class's methods.
	Repository = internal.Repository

	// Infra The parent class of the infrastructure, which can be inherited using the parent class's methods.
	Infra = internal.Infra

	// Entity is the entity's father interface.
	Entity = internal.Entity

	// DomainEvent Interface definition of domain events for subscription and publishing of entities.
	DomainEvent = internal.DomainEvent

	//UnitTest is a unit test tool.
	UnitTest = internal.UnitTest

	// BootManager Can be used to launch in the application.
	BootManager = internal.BootManager

	// Bus Message bus, using http header to pass through data.
	Bus = internal.Bus

	// BusHandler is the bus message middleware type.
	BusHandler = internal.BusHandler

	// Result is the controller return type.
	Result = IrisResult

	// Context is the context type.
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

// Prepare A prepared function is passed in for initialization.
func Prepare(f func(Initiator)) {
	internal.Prepare(f)
}

// NewUnitTest Unit testing tools.
func NewUnitTest() UnitTest {
	return internal.NewUnitTest()
}

// DefaultConfiguration Proxy a call to iris.DefaultConfiguration
func DefaultConfiguration() iris.Configuration {
	return iris.DefaultConfiguration()
}
