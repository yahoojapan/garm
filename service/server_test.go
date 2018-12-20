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
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/yahoojapan/garm/config"
)

func TestNewServer(t *testing.T) {
	type args struct {
		cfg config.Server
		h   http.Handler
	}
	tests := []struct {
		name      string
		args      args
		want      Server
		checkFunc func(got, want Server) error
	}{
		{
			name: "Check health address",
			args: args{
				cfg: config.Server{
					HealthzPath: "/healthz",
					HealthzPort: 8080,
				},
				h: func() http.Handler {
					return nil
				}(),
			},
			want: &server{
				hcsrv: &http.Server{
					Addr: fmt.Sprintf(":%d", 8080),
				},
			},
			checkFunc: func(got, want Server) error {
				if got.(*server).hcsrv.Addr != want.(*server).hcsrv.Addr {
					return fmt.Errorf("Healthz Addr not equals\tgot: %s\twant: %s", got.(*server).hcsrv.Addr, want.(*server).hcsrv.Addr)
				}
				return nil
			},
		},
		{
			name: "Check server address",
			args: args{
				cfg: config.Server{
					Port:        8081,
					HealthzPath: "/healthz",
					HealthzPort: 8080,
				},
				h: func() http.Handler {
					return nil
				}(),
			},
			want: &server{
				srv: &http.Server{
					Addr: fmt.Sprintf(":%d", 8081),
				},
			},
			checkFunc: func(got, want Server) error {
				if got.(*server).srv.Addr != want.(*server).srv.Addr {
					return fmt.Errorf("Server Addr not equals\tgot: %s\twant: %s", got.(*server).srv.Addr, want.(*server).srv.Addr)
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewServer(tt.args.cfg, tt.args.h)
			if err := tt.checkFunc(got, tt.want); err != nil {
				t.Errorf("NewServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_server_ListenAndServe(t *testing.T) {
	type fields struct {
		srv   *http.Server
		hcsrv *http.Server
		cfg   config.Server
		mu    *sync.RWMutex
	}
	type args struct {
		ctx context.Context
	}
	type test struct {
		name       string
		fields     fields
		args       args
		beforeFunc func() error
		checkFunc  func(*server, chan []error, error) error
		afterFunc  func() error
		want       error
	}
	tests := []test{
		func() test {
			ctx, cancelFunc := context.WithCancel(context.Background())

			keyKey := "dummy_key"
			key := "./testdata/dummyServer.key"
			certKey := "dummy_cert"
			cert := "./testdata/dummyServer.crt"

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				fmt.Fprintln(w, "Hello, client")
			})

			apiSrvPort := 9998
			hcSrvPort := 9999
			apiSrvAddr := fmt.Sprintf("https://127.0.0.1:%v", apiSrvPort)
			hcSrvAddr := fmt.Sprintf("http://127.0.0.1:%v", hcSrvPort)

			return test{
				name: "Test servers can start and stop",
				fields: fields{
					srv: func() *http.Server {
						s := &http.Server{
							Addr:    fmt.Sprintf(":%d", apiSrvPort),
							Handler: handler,
						}
						s.SetKeepAlivesEnabled(true)
						return s
					}(),
					hcsrv: func() *http.Server {
						s := &http.Server{
							Addr:    fmt.Sprintf(":%d", hcSrvPort),
							Handler: handler,
						}
						s.SetKeepAlivesEnabled(true)
						return s
					}(),
					cfg: config.Server{
						Port: apiSrvPort,
						TLS: config.TLS{
							Enabled: false,
							CertKey: certKey,
							KeyKey:  keyKey,
						},
					},
					mu: &sync.RWMutex{},
				},
				args: args{
					ctx: ctx,
				},
				beforeFunc: func() error {
					if err := os.Setenv(keyKey, key); err != nil {
						return err
					}
					if err := os.Setenv(certKey, cert); err != nil {
						return err
					}
					return nil
				},
				checkFunc: func(s *server, got chan []error, want error) error {
					time.Sleep(time.Millisecond * 150)
					http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

					checkSrvRunning := func(addr string) error {
						res, err := http.DefaultClient.Get(addr)
						if err != nil {
							return err
						}
						if res.StatusCode != 200 {
							return fmt.Errorf("Response status code invalid, %v", res.StatusCode)
						}
						return nil
					}

					if err := checkSrvRunning(apiSrvAddr); err != nil {
						return fmt.Errorf("Server not running, err: %v", err)
					}

					if err := checkSrvRunning(hcSrvAddr); err != nil {
						return fmt.Errorf("Health Check server not running")
					}

					cancelFunc()
					time.Sleep(time.Millisecond * 250)

					if err := checkSrvRunning(apiSrvAddr); err == nil {
						return fmt.Errorf("Server running")
					}

					if err := checkSrvRunning(hcSrvAddr); err == nil {
						return fmt.Errorf("Health Check server running")
					}

					return nil
				},
				afterFunc: func() error {
					cancelFunc()
					if err := os.Unsetenv(keyKey); err != nil {
						return err
					}
					if err := os.Unsetenv(certKey); err != nil {
						return nil
					}
					return nil
				},
			}
		}(),
		func() test {
			keyKey := "dummy_key"
			key := "./testdata/dummyServer.key"
			certKey := "dummy_cert"
			cert := "./testdata/dummyServer.crt"

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				fmt.Fprintln(w, "Hello, client")
			})

			apiSrvPort := 9998
			hcSrvPort := 9999
			apiSrvAddr := fmt.Sprintf("https://127.0.0.1:%v", apiSrvPort)
			hcSrvAddr := fmt.Sprintf("http://127.0.0.1:%v", hcSrvPort)

			return test{
				name: "Test HC server stop when api server stop",
				fields: fields{
					srv: func() *http.Server {
						srv := &http.Server{
							Addr:    fmt.Sprintf(":%d", apiSrvPort),
							Handler: handler,
						}

						srv.SetKeepAlivesEnabled(true)
						return srv
					}(),
					hcsrv: func() *http.Server {
						srv := &http.Server{
							Addr:    fmt.Sprintf(":%d", hcSrvPort),
							Handler: handler,
						}

						srv.SetKeepAlivesEnabled(true)
						return srv
					}(),
					cfg: config.Server{
						Port: apiSrvPort,
						TLS: config.TLS{
							Enabled: true,
							CertKey: certKey,
							KeyKey:  keyKey,
						},
					},
					mu: &sync.RWMutex{},
				},
				args: args{
					ctx: context.Background(),
				},
				beforeFunc: func() error {
					if err := os.Setenv(keyKey, key); err != nil {
						return err
					}
					if err := os.Setenv(certKey, cert); err != nil {
						return err
					}
					return nil
				},
				checkFunc: func(s *server, got chan []error, want error) error {
					time.Sleep(time.Millisecond * 150)
					http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

					checkSrvRunning := func(addr string) error {
						res, err := http.DefaultClient.Get(addr)
						if err != nil {
							return err
						}
						if res.StatusCode != 200 {
							return fmt.Errorf("Response status code invalid, %v", res.StatusCode)
						}
						return nil
					}

					if err := checkSrvRunning(apiSrvAddr); err != nil {
						return fmt.Errorf("Server not running, err: %v", err)
					}

					if err := checkSrvRunning(hcSrvAddr); err != nil {
						return fmt.Errorf("Health Check server not running")
					}

					s.srv.Close()
					time.Sleep(time.Millisecond * 150)

					if err := checkSrvRunning(apiSrvAddr); err == nil {
						return fmt.Errorf("Server running")
					}

					if err := checkSrvRunning(hcSrvAddr); err == nil {
						return fmt.Errorf("Health Check server running")
					}

					return nil
				},
				afterFunc: func() error {
					if err := os.Unsetenv(keyKey); err != nil {
						return err
					}
					if err := os.Unsetenv(certKey); err != nil {
						return nil
					}
					return nil
				},
			}
		}(),

		func() test {
			keyKey := "dummy_key"
			key := "./testdata/dummyServer.key"
			certKey := "dummy_cert"
			cert := "./testdata/dummyServer.crt"

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				fmt.Fprintln(w, "Hello, client")
			})

			apiSrvPort := 9998
			hcSrvPort := 9999
			apiSrvAddr := fmt.Sprintf("https://127.0.0.1:%v", apiSrvPort)
			hcSrvAddr := fmt.Sprintf("http://127.0.0.1:%v", hcSrvPort)

			return test{
				name: "Test api server stop when HC server stop",
				fields: fields{
					srv: func() *http.Server {
						srv := &http.Server{
							Addr:    fmt.Sprintf(":%d", apiSrvPort),
							Handler: handler,
						}

						srv.SetKeepAlivesEnabled(true)
						return srv
					}(),
					hcsrv: func() *http.Server {
						srv := &http.Server{
							Addr:    fmt.Sprintf(":%d", hcSrvPort),
							Handler: handler,
						}

						srv.SetKeepAlivesEnabled(true)
						return srv
					}(),
					cfg: config.Server{
						Port: apiSrvPort,
						TLS: config.TLS{
							Enabled: true,
							CertKey: certKey,
							KeyKey:  keyKey,
						},
					},
					mu: &sync.RWMutex{},
				},
				args: args{
					ctx: context.Background(),
				},
				beforeFunc: func() error {
					if err := os.Setenv(keyKey, key); err != nil {
						return err
					}
					if err := os.Setenv(certKey, cert); err != nil {
						return err
					}
					return nil
				},
				checkFunc: func(s *server, got chan []error, want error) error {
					time.Sleep(time.Millisecond * 150)
					http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

					checkSrvRunning := func(addr string) error {
						res, err := http.DefaultClient.Get(addr)
						if err != nil {
							return err
						}
						if res.StatusCode != 200 {
							return fmt.Errorf("Response status code invalid, %v", res.StatusCode)
						}
						return nil
					}

					if err := checkSrvRunning(apiSrvAddr); err != nil {
						return fmt.Errorf("Server not running, err: %v", err)
					}

					if err := checkSrvRunning(hcSrvAddr); err != nil {
						return fmt.Errorf("Health Check server not running")
					}

					s.hcsrv.Close()
					time.Sleep(time.Millisecond * 150)

					if err := checkSrvRunning(apiSrvAddr); err == nil {
						return fmt.Errorf("Server running")
					}

					if err := checkSrvRunning(hcSrvAddr); err == nil {
						return fmt.Errorf("Health Check server running")
					}

					return nil
				},
				afterFunc: func() error {
					if err := os.Unsetenv(keyKey); err != nil {
						return err
					}
					if err := os.Unsetenv(certKey); err != nil {
						return nil
					}
					return nil
				},
			}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeFunc != nil {
				err := tt.beforeFunc()
				if err != nil {
					t.Error(err)
				}
			}
			if tt.afterFunc != nil {
				defer func() {
					err := tt.afterFunc()
					if err != nil {
						t.Error(err)
					}
				}()
			}

			s := &server{
				srv:   tt.fields.srv,
				hcsrv: tt.fields.hcsrv,
				cfg:   tt.fields.cfg,
				mu:    tt.fields.mu,
			}

			e := s.ListenAndServe(tt.args.ctx)
			if err := tt.checkFunc(s, e, tt.want); err != nil {
				t.Errorf("server.listenAndServe() Error = %v", err)
			}
		})
	}
}

func Test_server_hcShutdown(t *testing.T) {
	type fields struct {
		srv        *http.Server
		srvRunning bool
		hcsrv      *http.Server
		hcrunning  bool
		cfg        config.Server
		pwt        time.Duration
		sddur      time.Duration
		mu         *sync.RWMutex
	}
	type args struct {
		ctx context.Context
	}
	type test struct {
		name       string
		fields     fields
		args       args
		beforeFunc func() error
		checkFunc  func(*server, error, error) error
		afterFunc  func() error
		want       error
	}
	tests := []test{
		func() test {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
			})
			hcsrv := httptest.NewServer(handler)

			return test{
				name: "hcShutdown works",
				fields: fields{
					hcsrv: hcsrv.Config,
				},
				args: args{
					ctx: context.Background(),
				},
				checkFunc: func(s *server, got, want error) error {
					return got
				},
				afterFunc: func() error {
					hcsrv.Close()
					return nil
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeFunc != nil {
				err := tt.beforeFunc()
				if err != nil {
					t.Error(err)
				}
			}
			if tt.afterFunc != nil {
				defer func() {
					err := tt.afterFunc()
					if err != nil {
						t.Error(err)
					}
				}()
			}

			s := &server{
				srv:        tt.fields.srv,
				srvRunning: tt.fields.srvRunning,
				hcsrv:      tt.fields.hcsrv,
				hcrunning:  tt.fields.hcrunning,
				cfg:        tt.fields.cfg,
				pwt:        tt.fields.pwt,
				sddur:      tt.fields.sddur,
				mu:         tt.fields.mu,
			}
			e := s.hcShutdown(tt.args.ctx)
			if err := tt.checkFunc(s, e, tt.want); err != nil {
				t.Errorf("server.listenAndServe() Error = %v", err)
			}
		})
	}
}

func Test_server_apiShutdown(t *testing.T) {
	type fields struct {
		srv        *http.Server
		srvRunning bool
		hcsrv      *http.Server
		hcrunning  bool
		cfg        config.Server
		pwt        time.Duration
		sddur      time.Duration
		mu         *sync.RWMutex
	}
	type args struct {
		ctx context.Context
	}
	type test struct {
		name       string
		fields     fields
		args       args
		beforeFunc func() error
		checkFunc  func(*server, error, error) error
		afterFunc  func() error
		want       error
	}
	tests := []test{
		func() test {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
			})
			apisrv := httptest.NewServer(handler)

			return test{
				name: "apiShutdown works",
				fields: fields{
					srv: apisrv.Config,
				},
				args: args{
					ctx: context.Background(),
				},
				checkFunc: func(s *server, got, want error) error {
					return got
				},
				afterFunc: func() error {
					apisrv.Close()
					return nil
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeFunc != nil {
				err := tt.beforeFunc()
				if err != nil {
					t.Error(err)
				}
			}
			if tt.afterFunc != nil {
				defer func() {
					err := tt.afterFunc()
					if err != nil {
						t.Error(err)
					}
				}()
			}

			s := &server{
				srv:        tt.fields.srv,
				srvRunning: tt.fields.srvRunning,
				hcsrv:      tt.fields.hcsrv,
				hcrunning:  tt.fields.hcrunning,
				cfg:        tt.fields.cfg,
				pwt:        tt.fields.pwt,
				sddur:      tt.fields.sddur,
				mu:         tt.fields.mu,
			}
			e := s.apiShutdown(tt.args.ctx)
			if err := tt.checkFunc(s, e, tt.want); err != nil {
				t.Errorf("server.listenAndServe() Error = %v", err)
			}
		})
	}
}

func Test_server_createHealthCheckServiceMux(t *testing.T) {
	type args struct {
		pattern string
	}
	type test struct {
		name       string
		args       args
		beforeFunc func() error
		checkFunc  func(*http.ServeMux) error
		afterFunc  func() error
	}
	tests := []test{
		func() test {
			return test{
				name: "Test create server mux",
				args: args{
					pattern: ":8080",
				},
				checkFunc: func(got *http.ServeMux) error {
					if got == nil {
						return fmt.Errorf("serveMux is empty")
					}
					return nil
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.afterFunc != nil {
				defer func() {
					if err := tt.afterFunc(); err != nil {
						t.Errorf("%v", err)
						return
					}
				}()
			}

			if tt.beforeFunc != nil {
				err := tt.beforeFunc()
				if err != nil {
					t.Errorf("beforeFunc error, error: %v", err)
					return
				}
			}

			got := createHealthCheckServiceMux(tt.args.pattern)
			if err := tt.checkFunc(got); err != nil {
				t.Errorf("server.listenAndServeAPI() Error = %v", err)
			}
		})
	}
}

func Test_server_handleHealthCheckRequest(t *testing.T) {
	type args struct {
		rw http.ResponseWriter
		r  *http.Request
	}
	type test struct {
		name       string
		args       args
		beforeFunc func() error
		checkFunc  func() error
		afterFunc  func() error
	}
	tests := []test{
		func() test {
			rw := httptest.NewRecorder()

			return test{
				name: "Test handle HTTP GET request health check request",
				args: args{
					rw: rw,
					r:  httptest.NewRequest(http.MethodGet, "/", nil),
				},
				checkFunc: func() error {
					result := rw.Result()
					if header := result.StatusCode; header != http.StatusOK {
						return fmt.Errorf("Header is not correct, got: %v", header)
					}
					if contentType := rw.Header().Get("Content-Type"); contentType != "text/plain;charset=UTF-8" {
						return fmt.Errorf("Content type is not correct, got: %v", contentType)
					}
					return nil
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.afterFunc != nil {
				defer func() {
					if err := tt.afterFunc(); err != nil {
						t.Errorf("%v", err)
						return
					}
				}()
			}

			if tt.beforeFunc != nil {
				err := tt.beforeFunc()
				if err != nil {
					t.Errorf("beforeFunc error, error: %v", err)
					return
				}
			}

			handleHealthCheckRequest(tt.args.rw, tt.args.r)
			if err := tt.checkFunc(); err != nil {
				t.Errorf("error: %v", err)
			}
		})
	}
}

func Test_server_listenAndServeAPI(t *testing.T) {
	type fields struct {
		srv   *http.Server
		hcsrv *http.Server
		cfg   config.Server
	}
	type test struct {
		name       string
		fields     fields
		beforeFunc func() error
		checkFunc  func(*server, error) error
		afterFunc  func() error
		want       error
	}
	tests := []test{
		func() test {
			keyKey := "dummy_key"
			key := "./testdata/dummyServer.key"
			certKey := "dummy_cert"
			cert := "./testdata/dummyServer.crt"

			return test{
				name: "Test server startup",
				fields: fields{
					srv: &http.Server{
						Handler: func() http.Handler {
							return nil
						}(),
						Addr: fmt.Sprintf(":%d", 9999),
					},
					cfg: config.Server{
						Port: 9999,
						TLS: config.TLS{
							Enabled: true,
							CertKey: certKey,
							KeyKey:  keyKey,
						},
					},
				},
				beforeFunc: func() error {
					err := os.Setenv(keyKey, key)
					if err != nil {
						return err
					}
					err = os.Setenv(certKey, cert)
					if err != nil {
						return err
					}
					return nil
				},
				checkFunc: func(s *server, want error) error {
					// listenAndServeAPI function is blocking, so we need to set timer to shutdown the process
					go func() {
						time.Sleep(time.Second * 1)
						err := s.srv.Shutdown(context.Background())
						if err != nil {
							t.Error(err)
						}
					}()

					got := s.listenAndServeAPI()

					if got != want {
						return fmt.Errorf("got:\t%v\nwant:\t%v", got, want)
					}
					return nil
				},
				afterFunc: func() error {
					os.Unsetenv(keyKey)
					os.Unsetenv(certKey)
					return nil
				},
				want: http.ErrServerClosed,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.afterFunc != nil {
				defer func() {
					if err := tt.afterFunc(); err != nil {
						t.Errorf("%v", err)
						return
					}
				}()
			}

			if tt.beforeFunc != nil {
				err := tt.beforeFunc()
				if err != nil {
					t.Errorf("beforeFunc error, error: %v", err)
					return
				}
			}

			if err := tt.checkFunc(&server{
				srv:   tt.fields.srv,
				hcsrv: tt.fields.hcsrv,
				cfg:   tt.fields.cfg,
			}, tt.want); err != nil {
				t.Errorf("server.listenAndServeAPI() Error = %v", err)
			}
		})
	}
}
