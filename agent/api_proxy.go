package agent

import (
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/buildkite/agent/api"
	"github.com/buildkite/agent/logger"
)

type APIProxy struct {
	Token    string
	Endpoint string

	file     *os.File
	listener net.Listener
}

func (p *APIProxy) Listen() error {
	var err error
	p.file, err = ioutil.TempFile("", "api-proxy")
	if err != nil {
		return err
	}

	// Servers should unlink the socket path name prior to binding it.
	// https://troydhanson.github.io/network/Unix_domain_sockets.html
	_ = os.Remove(p.file.Name())

	logger.Debug("[APIProxy] Created proxy socket file: %s", p.file.Name())

	// create a unix socket to do the listening
	p.listener, err = net.Listen("unix", p.file.Name())
	if err != nil {
		return err
	}

	if err = os.Chmod(p.file.Name(), 0600); err != nil {
		return err
	}

	endpoint, err := url.Parse(p.Endpoint)
	if err != nil {
		return err
	}

	go func() {
		proxy := httputil.NewSingleHostReverseProxy(endpoint)
		proxy.Transport = &api.AuthenticatedTransport{Token: p.Token}

		// set the host header whilst proxying
		director := proxy.Director
		proxy.Director = func(req *http.Request) {
			director(req)
			req.Host = req.URL.Host
		}

		err := http.Serve(p.listener, proxy)
		logger.Debug("[APIProxy] Socket proxy terminated with: %v", err)
	}()

	return nil
}

func (p *APIProxy) Destroy() error {
	logger.Debug("[APIProxy] Destroying %s", p.file.Name())
	_ = p.listener.Close()
	_ = os.Remove(p.file.Name())
	return nil
}

func (p *APIProxy) Socket() string {
	return p.listener.Addr().String()
}
