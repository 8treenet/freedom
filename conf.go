package freedom

var (
	// configurator is a instance of Configurator. It is never nil.
	configurator Configurator
)

// Configurer .
type Configurer = Configurator

// Configurator .
type Configurator interface {
	Configure(obj interface{}, file string, metaData ...interface{}) error
}

func initConfigurator() {
	SetConfigurator(newFallbackConfigurator())
}

// SetConfigurator assigns a Configurator to global configurator
func SetConfigurator(c Configurator) {
	configurator = c
}

// SetConfigurer assigns a Configurator to global configurator
func SetConfigurer(c Configurer) {
	SetConfigurator(c)
}

// Configure .
func Configure(obj interface{}, file string, metaData ...interface{}) error {
	return configurator.Configure(obj, file)
}
