package internal

import (
	"net/http"

	"github.com/kataras/iris/v12"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// NewRunner can be used as an argument for the `Run` method.
// It accepts a host address which is used to build a server
// and a listener which listens on that host and port.
//
// Addr should have the form of [host]:port, i.e localhost:8080 or :8080.
//
// Second argument is optional, it accepts one or more
// `func(*host.Configurator)` that are being executed
// on that specific host that this function will create to start the server.
// Via host configurators you can configure the back-end host supervisor,
// i.e to add events for shutdown, serve or error.
func NewRunner(addr string, configurators ...IrisHostConfigurator) IrisRunner {
	return iris.Addr(addr, configurators...)
}

// NewAutoTLSRunner can be used as an argument for the `Run` method.
// It will start the Application's secure server using
// certifications created on the fly by the "autocert" golang/x package,
// so localhost may not be working, use it at "production" machine.
//
// Addr should have the form of [host]:port, i.e mydomain.com:443.
//
// The whitelisted domains are separated by whitespace in "domain" argument,
// i.e "8tree.net", can be different than "addr".
// If empty, all hosts are currently allowed. This is not recommended,
// as it opens a potential attack where clients connect to a server
// by IP address and pretend to be asking for an incorrect host name.
// Manager will attempt to obtain a certificate for that host, incorrectly,
// eventually reaching the CA's rate limit for certificate requests
// and making it impossible to obtain actual certificates.
//
// For an "e-mail" use a non-public one, letsencrypt needs that for your own security.
//
// Note: `AutoTLS` will start a new server for you
// which will redirect all http versions to their https, including subdomains as well.
//
// Last argument is optional, it accepts one or more
// `func(*host.Configurator)` that are being executed
// on that specific host that this function will create to start the server.
// Via host configurators you can configure the back-end host supervisor,
// i.e to add events for shutdown, serve or error.
// Look at the `ConfigureHost` too.
func NewAutoTLSRunner(addr string, domain string, email string, configurators ...IrisHostConfigurator) IrisRunner {
	return func(irisApp *IrisApplication) error {
		return irisApp.NewHost(&http.Server{Addr: addr}).
			Configure(configurators...).
			ListenAndServeAutoTLS(domain, email, "letscache")
	}
}

// NewTLSRunner can be used as an argument for the `Run` method.
// It will start the Application's secure server.
//
// Use it like you used to use the http.ListenAndServeTLS function.
//
// Addr should have the form of [host]:port, i.e localhost:443 or :443.
// CertFile & KeyFile should be filenames with their extensions.
//
// Second argument is optional, it accepts one or more
// `func(*host.Configurator)` that are being executed
// on that specific host that this function will create to start the server.
// Via host configurators you can configure the back-end host supervisor,
// i.e to add events for shutdown, serve or error.
// An example of this use case can be found at:
func NewTLSRunner(addr string, certFile, keyFile string, configurators ...IrisHostConfigurator) IrisRunner {
	return func(irisApp *IrisApplication) error {
		return irisApp.NewHost(&http.Server{Addr: addr}).
			Configure(configurators...).
			ListenAndServeTLS(certFile, keyFile)
	}
}

// NewH2CRunner .
func NewH2CRunner(addr string, configurators ...IrisHostConfigurator) IrisRunner {
	return func(app *IrisApplication) error {
		h2cSer := &http2.Server{}
		ser := &http.Server{
			Addr:    addr,
			Handler: h2c.NewHandler(app, h2cSer),
		}
		return app.NewHost(ser).Configure(configurators...).ListenAndServe()
	}
}
