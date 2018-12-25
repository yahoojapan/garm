/*
Copyright (C)  2018 Yahoo Japan Corporation Athenz team.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/kpango/glg"
	"github.com/pkg/errors"
	"github.com/yahoojapan/garm/config"
)

// Server represents a Garm server behaviour.
type Server interface {
	ListenAndServe(context.Context) chan []error
}

type server struct {
	// Webhook server
	srv        *http.Server
	srvRunning bool

	// Health Check server
	hcsrv     *http.Server
	hcrunning bool

	cfg config.Server

	// ProbeWaitTime
	pwt time.Duration

	// ShutdownDuration
	sddur time.Duration

	// mutex lock variable
	mu *sync.RWMutex
}

const (
	// ContentType represents a HTTP header name "Content-Type"
	ContentType = "Content-Type"

	// TextPlain represents a HTTP content type "text/plain"
	TextPlain = "text/plain"

	// CharsetUTF8 represents a UTF-8 charset for HTTP response "charset=UTF-8"
	CharsetUTF8 = "charset=UTF-8"
)

var (
	// ErrContextClosed represents the error that the context is closed
	ErrContextClosed = errors.New("context Closed")
)

// NewServer returns a Server interface, which includes Webhook server and health check server structs.
// The webhook server is a http.Server instance, which the port number is read from "config.Server.Port"
// , and make use of the given handler.
//
// The health check server is a http.Server instance, which the port number is read from "config.Server.HealthzPort"
// , and its handler always return HTTP Status OK (200) response on HTTP GET request.
func NewServer(cfg config.Server, h http.Handler) Server {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: h,
	}
	srv.SetKeepAlivesEnabled(true)

	hcsrv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HealthzPort),
		Handler: createHealthCheckServiceMux(cfg.HealthzPath),
	}
	hcsrv.SetKeepAlivesEnabled(true)

	dur, err := time.ParseDuration(cfg.ShutdownDuration)
	if err != nil {
		err = glg.Error(errors.Wrapf(err, "invalid shutdown duration %s", cfg.ShutdownDuration))
		if err != nil {
			glg.Fatal(errors.Wrap(err, "shutdown duration parse error log output failed"))
			return nil
		}
		dur = time.Second * 5
	}

	pwt, err := time.ParseDuration(cfg.ProbeWaitTime)
	if err != nil {
		err = glg.Error(errors.Wrapf(err, "invalid ProbeWaitTime duration %s", cfg.ProbeWaitTime))
		if err != nil {
			glg.Fatal(errors.Wrap(err, "ProbeWaitTime duration parse error log output failed"))
			return nil
		}
		pwt = time.Second * 3
	}

	return &server{
		srv:   srv,
		hcsrv: hcsrv,
		cfg:   cfg,
		pwt:   pwt,
		sddur: dur,
		mu:    &sync.RWMutex{},
	}
}

// ListenAndServe returns an error channel, which includes the errors returned from webhook server.
// It start both health check and webhook server, and both servers will close whenever the context receives a Done signal.
// Whenever the server closed, the webhook server will shutdown after a defined duration (cfg.ProbeWaitTime), while the health check server will shutdown immediately.
func (s *server) ListenAndServe(ctx context.Context) chan []error {
	echan := make(chan []error, 1)
	// error channels to keep track server status
	sech := make(chan error, 1)
	hech := make(chan error, 1)

	wg := new(sync.WaitGroup)
	wg.Add(2)

	// start both webhook server and health check server
	go func() {
		s.mu.Lock()
		err := glg.Info("garm api server starting")
		if err != nil {
			err = glg.Error(errors.Wrap(err, "garm api server start message output failed"))
			if err != nil {
				glg.Fatal(errors.Wrap(err, "error log output failed"))
			}
			s.mu.Unlock()
			sech <- err
			return
		}
		s.srvRunning = true
		s.mu.Unlock()
		wg.Done()

		sech <- s.listenAndServeAPI()
		close(sech)

		s.mu.Lock()
		s.srvRunning = false
		s.mu.Unlock()
		err = glg.Info("garm api server stopped")
		if err != nil {
			err = glg.Error(errors.Wrap(err, "garm api server stop message output failed"))
			if err != nil {
				glg.Fatal(errors.Wrap(err, "error log output failed"))
			}
		}
	}()

	go func() {
		s.mu.Lock()
		err := glg.Info("garm health check server starting")
		if err != nil {
			err = glg.Error(errors.Wrap(err, "garm health check server start message output failed"))
			if err != nil {
				glg.Fatal(errors.Wrap(err, "error log output failed"))
			}
			s.mu.Unlock()
			hech <- err
			return
		}
		s.hcrunning = true
		s.mu.Unlock()
		wg.Done()

		hech <- s.hcsrv.ListenAndServe()
		close(hech)

		s.mu.Lock()
		s.hcrunning = false
		s.mu.Unlock()
		err = glg.Info("garm health check server stopped")
		if err != nil {
			err = glg.Error(errors.Wrap(err, "garm health check server stop message output failed"))
			if err != nil {
				glg.Fatal(errors.Wrap(err, "error log output failed"))
			}
		}
	}()

	go func() {
		// wait for all server running
		wg.Wait()

		appendErr := func(errs []error, err error) []error {
			if err != nil {
				return append(errs, err)
			}
			return errs
		}

		errs := make([]error, 0, 3)
		for {
			select {
			case <-ctx.Done(): // when context receives Done signal, closes running servers and returns any errors
				s.mu.RLock()
				if s.hcrunning {
					err := glg.Info("garm health check server will shutdown")
					if err != nil {
						errs = appendErr(errs, errors.Wrap(err, "garm health check server shutdowm message output failed"))
					}
					err = s.hcShutdown(context.Background())
					if err != nil {
						errs = appendErr(errs, errors.Wrap(err, "garm health check server shutdown failed"))
					}
				}
				if s.srvRunning {
					err := glg.Info("garm api server will shutdown")
					if err != nil {
						errs = appendErr(errs, errors.Wrap(err, "garm api server shutdowm message output failed"))
					}
					err = s.apiShutdown(context.Background())
					if err != nil {
						errs = appendErr(errs, errors.Wrap(err, "garm api server shutdown failed"))
					}
				}
				s.mu.RUnlock()

				echan <- appendErr(errs, ctx.Err())
				return

			case err := <-sech: // when webhook server returns, closes running health check server and returns any errors
				if err != nil {
					errs = appendErr(errs, err)
				}

				s.mu.RLock()
				if s.hcrunning {
					err = glg.Info("garm health check server will shutdown")
					if err != nil {
						errs = appendErr(errs, errors.Wrap(err, "garm health check server shutdowm message output failed"))
					}
					err = s.hcShutdown(ctx)
					if err != nil {
						errs = appendErr(errs, errors.Wrap(err, "garm health check server shutdown failed"))
					}
				}
				s.mu.RUnlock()
				echan <- errs
				return

			case err := <-hech: // when health check server returns, closes running webhook server and returns any errors
				if err != nil {
					errs = append(errs, err)
				}

				s.mu.RLock()
				if s.srvRunning {
					err = glg.Info("garm api server will shutdown")
					if err != nil {
						errs = appendErr(errs, errors.Wrap(err, "garm api server shutdowm message output failed"))
					}
					err = s.apiShutdown(ctx)
					if err != nil {
						errs = appendErr(errs, errors.Wrap(err, "garm api server shutdown failed"))
					}
				}
				s.mu.RUnlock()
				echan <- errs
				return
			}
		}
	}()

	return echan
}

// hcShutdown returns any errors on shutting down the health check server.
func (s *server) hcShutdown(ctx context.Context) error {
	hctx, hcancel := context.WithTimeout(ctx, s.sddur)
	defer hcancel()
	return s.hcsrv.Shutdown(hctx)
}

// apiShutdown returns any errors on shutting down the webhook server.
// To prevent any issues from K8s, sleeps config.ProbeWaitTime before shutting down the webhook server.
func (s *server) apiShutdown(ctx context.Context) error {
	time.Sleep(s.pwt)
	sctx, scancel := context.WithTimeout(ctx, s.sddur)
	defer scancel()
	return s.srv.Shutdown(sctx)
}

// createHealthCheckServiceMux returns a *http.ServeMux object.
// It registers the health check server handler to given pattern.
func createHealthCheckServiceMux(pattern string) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc(pattern, handleHealthCheckRequest)
	return mux
}

// handleHealthCheckRequest is a handler function for health check requests, which always response HTTP Status OK (200).
func handleHealthCheckRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
		w.Header().Set(ContentType, fmt.Sprintf("%s;%s", TextPlain, CharsetUTF8))
		_, err := fmt.Fprint(w, http.StatusText(http.StatusOK))
		if err != nil {
			err = glg.Error(errors.Wrap(err, "health check response failed"))
			if err != nil {
				glg.Fatal(errors.Wrap(err, "error log output failed"))
			}
		}
	}
}

// listenAndServeAPI returns any errors on starting the HTTPS server, including any errors on loading TLS certificate.
func (s *server) listenAndServeAPI() error {
	if !s.cfg.TLS.Enabled {
		return s.srv.ListenAndServe()
	}

	cfg, err := NewTLSConfig(s.cfg.TLS)
	if err == nil && cfg != nil {
		s.srv.TLSConfig = cfg
	}
	if err != nil {
		err = glg.Error(errors.Wrap(err, "tls configuration failed"))
		if err != nil {
			return errors.Wrap(err, "tls config error output failed")
		}
	}
	return s.srv.ListenAndServeTLS("", "")
}
