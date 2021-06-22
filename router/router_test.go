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

package router

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kpango/glg"
	"github.com/yahoojapan/garm/config"
	"github.com/yahoojapan/garm/handler"
)

// dummyHandler is a mock implement for handler.Handler
type dummyHandler struct {
	responseValue string
}

// Authenticate is a mock implement for handler.Handler. It writes h.responseValue to the http.ResponseWriter.
func (h *dummyHandler) Authenticate(w http.ResponseWriter, r *http.Request) error {
	_, err := w.Write([]byte(h.responseValue))
	return err
}

// Authorize is a mock implement for handler.Handler. It writes h.responseValue to the http.ResponseWriter.
func (h *dummyHandler) Authorize(w http.ResponseWriter, r *http.Request) error {
	_, err := w.Write([]byte(h.responseValue))
	return err
}

// readCloserMock is the adapter implementation of io.ReadCloser interface for mocking.
type readCloserMock struct {
	readMock  func(p []byte) (n int, err error)
	closeMock func() error
}

// Read is just an adapter.
func (r *readCloserMock) Read(p []byte) (n int, err error) {
	return r.readMock(p)
}

// Close is just an adapter.
func (r *readCloserMock) Close() error {
	return r.closeMock()
}

func TestNewServeMux(t *testing.T) {
	type args struct {
		cfg config.Server
		h   handler.Handler
	}
	type testcase struct {
		name      string
		args      args
		checkFunc func(*http.ServeMux) error
	}
	tests := []testcase{
		{
			name: "Check New, MaxIdleConnsPerHost set correctly",
			args: args{
				cfg: config.Server{},
				h:   &dummyHandler{"dummy-handler-46"},
			},
			checkFunc: func(serveMux *http.ServeMux) error {
				got := http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost
				want := 32
				if got != want {
					return fmt.Errorf("New() MaxIdleConnsPerHost = %v, want %v", got, want)
				}
				return nil
			},
		},
		func() testcase {
			handler := &dummyHandler{"dummy-handler-57"}
			return testcase{
				name: "Check New, can handle request correctly",
				args: args{
					cfg: config.Server{
						Timeout: "1s",
					},
					h: handler,
				},
				checkFunc: func(serveMux *http.ServeMux) (err error) {
					var recorder *httptest.ResponseRecorder
					var request *http.Request
					var response *http.Response
					var want string
					var gotByte []byte

					// Authenticate request
					want = "Authenticate response"
					handler.responseValue = want
					recorder = httptest.NewRecorder()
					request, err = http.NewRequest(http.MethodPost, "/authn", nil)
					if err != nil {
						return
					}
					serveMux.ServeHTTP(recorder, request)
					response = recorder.Result()
					defer response.Body.Close()
					gotByte, err = ioutil.ReadAll(response.Body)
					if err != nil {
						return
					}
					if string(gotByte) != want {
						return fmt.Errorf("New() ServeMux on request %v, response body = %v, want %v", request, string(gotByte), want)
					}

					// Authorize request
					want = "Authorize response"
					handler.responseValue = want
					recorder = httptest.NewRecorder()
					request, err = http.NewRequest(http.MethodPost, "/authz", nil)
					if err != nil {
						return
					}
					serveMux.ServeHTTP(recorder, request)
					response = recorder.Result()
					defer response.Body.Close()
					gotByte, err = ioutil.ReadAll(response.Body)
					if err != nil {
						return
					}
					if string(gotByte) != want {
						return fmt.Errorf("New() ServeMux on request %v, response body = %v, want %v", request, string(gotByte), want)
					}

					return nil
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.cfg, tt.args.h)
			if err := tt.checkFunc(got); err != nil {
				t.Error(err)
			}
		})
	}
}

func Test_parseTimeout(t *testing.T) {
	type args struct {
		timeout string
	}
	tests := []struct {
		name string
		args args
		want time.Duration
	}{
		{
			name: "Check parseTimeout, invalid duration, return default value",
			args: args{
				timeout: "",
			},
			want: time.Second * 3,
		},
		{
			name: "Check parseTimeout, parse duration success",
			args: args{
				timeout: "140s",
			},
			want: time.Second * 140,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseTimeout(tt.args.timeout); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseTimeout() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_routing(t *testing.T) {

	var glgMutex = &sync.Mutex{}

	type args struct {
		m []string
		t time.Duration
		h handler.Func
	}
	type testcase struct {
		name      string
		args      args
		checkFunc func(http.Handler) error
	}
	tests := []testcase{
		func() testcase {
			handlerFunc := func(rw http.ResponseWriter, r *http.Request) error {
				_, err := rw.Write([]byte("response-body-174"))
				return err
			}
			want := "response-body-174"

			return testcase{
				name: "Check routing, returned Handler can handle request correctly",
				args: args{
					m: []string{
						http.MethodGet,
					},
					t: time.Second * 3,
					h: handlerFunc,
				},
				checkFunc: func(server http.Handler) error {
					request, err := http.NewRequest(http.MethodGet, "/", nil)
					if err != nil {
						return err
					}
					recorder := httptest.NewRecorder()
					server.ServeHTTP(recorder, request)

					response := recorder.Result()
					defer response.Body.Close()
					gotByte, err := ioutil.ReadAll(response.Body)
					if err != nil {
						return err
					}

					got := string(gotByte)
					if got != want {
						return fmt.Errorf("routing() http.Handler on request %v, response body = %v, want %v", request, got, want)
					}

					return nil
				},
			}
		}(),
		func() testcase {

			handlerFunc := func(rw http.ResponseWriter, r *http.Request) error {
				return fmt.Errorf("error-response-body-%d", 215)
			}
			want := "Error: handler error occurred: error-response-body-215\t" + http.StatusText(http.StatusInternalServerError) + "\n"

			return testcase{
				name: "Check routing, returned Handler can handle handler.Func error correctly",
				args: args{
					m: []string{
						http.MethodGet,
					},
					t: time.Second * 3,
					h: handlerFunc,
				},
				checkFunc: func(server http.Handler) error {
					// disable logger
					glgMutex.Lock()
					glg.Get().SetMode(glg.NONE)

					request, err := http.NewRequest(http.MethodGet, "/", nil)
					if err != nil {
						return err
					}
					recorder := httptest.NewRecorder()
					server.ServeHTTP(recorder, request)
					glgMutex.Unlock()

					response := recorder.Result()
					defer response.Body.Close()
					gotByte, err := ioutil.ReadAll(response.Body)
					if err != nil {
						return err
					}

					got := string(gotByte)
					if got != want {
						return fmt.Errorf("routing() http.Handler on request %v, response body = %v, want %v", request, got, want)
					}

					return nil
				},
			}
		}(),
		func() testcase {
			handlerFunc := func(rw http.ResponseWriter, r *http.Request) error {
				return fmt.Errorf("error-response-body-%d", 251)
			}
			want := "Method: " + http.MethodGet + "\t" + http.StatusText(http.StatusMethodNotAllowed) + "\n"

			return testcase{
				name: "Check routing, returned Handler can handle unexpected HTTP request method correctly (empty method list)",
				args: args{
					t: time.Second * 3,
					h: handlerFunc,
				},
				checkFunc: func(server http.Handler) error {
					request, err := http.NewRequest(http.MethodGet, "/", nil)
					if err != nil {
						return err
					}
					recorder := httptest.NewRecorder()
					server.ServeHTTP(recorder, request)

					response := recorder.Result()
					defer response.Body.Close()
					gotByte, err := ioutil.ReadAll(response.Body)
					if err != nil {
						return err
					}

					got := string(gotByte)
					if got != want {
						return fmt.Errorf("routing() http.Handler on request %v, response body = %v, want %v", request, got, want)
					}

					return nil
				},
			}
		}(),
		func() testcase {
			handlerFunc := func(rw http.ResponseWriter, r *http.Request) error {
				return fmt.Errorf("error-response-body-%d", 284)
			}
			want := "Method: " + http.MethodOptions + "\t" + http.StatusText(http.StatusMethodNotAllowed) + "\n"

			return testcase{
				name: "Check routing, returned Handler can handle unexpected HTTP request method correctly (no matches in method list)",
				args: args{
					m: []string{
						http.MethodGet,
						http.MethodPost,
					},
					t: time.Second * 3,
					h: handlerFunc,
				},
				checkFunc: func(server http.Handler) error {
					request, err := http.NewRequest(http.MethodOptions, "/", nil)
					if err != nil {
						return err
					}
					recorder := httptest.NewRecorder()
					server.ServeHTTP(recorder, request)

					response := recorder.Result()
					defer response.Body.Close()
					gotByte, err := ioutil.ReadAll(response.Body)
					if err != nil {
						return err
					}

					got := string(gotByte)
					if got != want {
						return fmt.Errorf("routing() http.Handler on request %v, response body = %v, want %v", request, got, want)
					}

					return nil
				},
			}
		}(),
		func() testcase {
			handlerFunc := func(rw http.ResponseWriter, r *http.Request) error {
				_, err := rw.Write([]byte("response-body-321"))
				return err
			}
			want := "response-body-321"

			return testcase{
				name: "Check routing, returned Handler can handle request correctly (multiple methods)",
				args: args{
					m: []string{
						http.MethodPost,
						http.MethodGet,
					},
					t: time.Second * 3,
					h: handlerFunc,
				},
				checkFunc: func(server http.Handler) error {
					for _, method := range []string{http.MethodGet, http.MethodPost} {
						request, err := http.NewRequest(method, "/", nil)
						if err != nil {
							return err
						}
						recorder := httptest.NewRecorder()
						server.ServeHTTP(recorder, request)

						response := recorder.Result()
						defer response.Body.Close()
						gotByte, err := ioutil.ReadAll(response.Body)
						if err != nil {
							return err
						}

						got := string(gotByte)
						if got != want {
							return fmt.Errorf("routing() http.Handler on request %v, response body = %v, want %v", request, got, want)
						}
					}

					return nil
				},
			}
		}(),
		func() testcase {

			normalTime := 9 * time.Millisecond
			veryLongTime := 9 * time.Second
			handlerFunc := func(rw http.ResponseWriter, r *http.Request) error {
				time.Sleep(veryLongTime)
				return nil
			}
			wantPrefix := "Handler Time Out: "

			return testcase{
				name: "Check routing, returned Handler can handle timeout",
				args: args{
					m: []string{
						http.MethodGet,
					},
					t: normalTime,
					h: handlerFunc,
				},
				checkFunc: func(server http.Handler) error {
					// overwrite log destination
					glgMutex.Lock()
					errorBuffer := new(bytes.Buffer)
					glg.Get().SetMode(glg.WRITER).SetWriter(errorBuffer)

					request, err := http.NewRequest(http.MethodGet, "/", nil)
					if err != nil {
						return err
					}
					recorder := httptest.NewRecorder()
					server.ServeHTTP(recorder, request)
					glgMutex.Unlock()

					// check error message to logger
					got := errorBuffer.String()
					if !strings.Contains(got, wantPrefix) {
						return fmt.Errorf("routing() http.Handler will have error log message = %v, want prefix %v", got, wantPrefix)
					}

					return nil
				},
			}
		}(),
		func() testcase {

			normalTime := 9 * time.Millisecond
			veryLongTime := 9 * time.Second
			handlerFunc := func(rw http.ResponseWriter, r *http.Request) error {
				time.Sleep(veryLongTime)
				return nil
			}
			wantPrefix := "Handler Time Out: "

			return testcase{
				name: "Check routing, returned Handler can handle parent context closed",
				args: args{
					m: []string{
						http.MethodGet,
					},
					t: veryLongTime,
					h: handlerFunc,
				},
				checkFunc: func(server http.Handler) error {
					// overwrite log destination
					glgMutex.Lock()
					errorBuffer := new(bytes.Buffer)
					glg.Get().SetMode(glg.WRITER).SetWriter(errorBuffer)

					request, err := http.NewRequest(http.MethodGet, "/", nil)
					if err != nil {
						return err
					}
					recorder := httptest.NewRecorder()

					// create and use parent context
					ctx, cancel := context.WithCancel(context.Background())
					go func() {
						time.Sleep(normalTime)
						cancel()
					}()
					server.ServeHTTP(recorder, request.WithContext(ctx))
					glgMutex.Unlock()

					// check error message to logger
					got := errorBuffer.String()
					if !strings.Contains(got, wantPrefix) {
						return fmt.Errorf("routing() http.Handler will have error log message = %v, want prefix %v", got, wantPrefix)
					}

					return nil
				},
			}
		}(),
		func() testcase {
			// glg.Fatalln() exit directly.
			// need to a closed writer (will have error on write) to panic() instead of exit()
			// however, the actual error message is lost and cannot be used for verifying

			wantError := io.ErrClosedPipe

			return testcase{
				name: "Check routing, returned Handler on unexpected HTTP, close request error",
				args: args{
					m: []string{},
					t: time.Second * 3,
					h: func(rw http.ResponseWriter, r *http.Request) error {
						return nil
					},
				},
				checkFunc: func(server http.Handler) (testError error) {
					// overwrite log destination
					glgMutex.Lock()
					pr, pw := io.Pipe()
					glg.Get().SetMode(glg.WRITER).SetWriter(pw)
					pw.Close()
					pr.Close()

					// prepare closed request
					rpr, rpw := io.Pipe()
					rpw.Close()
					rpr.Close()
					request, err := http.NewRequest(http.MethodGet, "/", rpr)
					if err != nil {
						return err
					}
					recorder := httptest.NewRecorder()

					// to catch the panic from glg below
					defer func() {
						glgMutex.Unlock()

						gotError := recover()
						if gotError != wantError {
							testError = fmt.Errorf("flushAndClose() unexpected exit with error = %v, want = %v", gotError, wantError)
						}
					}()

					// panic from glg
					server.ServeHTTP(recorder, request)

					return nil
				},
			}
		}(),
		func() testcase {
			handlerFunc := func(rw http.ResponseWriter, r *http.Request) error {
				_, err := rw.Write([]byte("response-body-565"))
				if err != nil {
					return err
				}
				panic("panic-566")
			}
			want := "response-body-565"
			wantError := "recover panic from athenz webhook:"

			return testcase{
				name: "Check routing, panic in handlerFunc",
				args: args{
					m: []string{
						http.MethodGet,
					},
					t: time.Second * 3,
					h: handlerFunc,
				},
				checkFunc: func(server http.Handler) error {
					// overwrite log destination
					glgMutex.Lock()
					errorBuffer := new(bytes.Buffer)
					glg.Get().SetMode(glg.WRITER).SetWriter(errorBuffer)

					request, err := http.NewRequest(http.MethodGet, "/", nil)
					if err != nil {
						return err
					}
					recorder := httptest.NewRecorder()
					server.ServeHTTP(recorder, request)
					glgMutex.Unlock()

					response := recorder.Result()
					defer response.Body.Close()
					gotByte, err := ioutil.ReadAll(response.Body)
					if err != nil {
						return err
					}

					got := string(gotByte)
					if got != want {
						return fmt.Errorf("routing() http.Handler on request %v, response body = %v, want %v", request, got, want)
					}

					gotError := errorBuffer.String()
					if !strings.Contains(gotError, wantError) {
						return fmt.Errorf("routing() http.Handler will have error log message = %v, want error %v", gotError, wantError)
					}

					return nil
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := routing(tt.args.m, tt.args.t, tt.args.h)
			if err := tt.checkFunc(got); err != nil {
				t.Error(err)
			}
		})
	}
}

func Test_flushAndClose(t *testing.T) {
	type args struct {
		readCloser io.ReadCloser
	}
	type testcase struct {
		name      string
		args      args
		wantError error
	}
	tests := []testcase{
		{
			name: "Check flushAndClose, readCloser is nil",
			args: args{
				readCloser: nil,
			},
			wantError: nil,
		},
		{
			name: "Check flushAndClose, flush & close success",
			args: args{
				readCloser: &readCloserMock{
					readMock: func(p []byte) (n int, err error) {
						return 0, io.EOF
					},
					closeMock: func() error {
						return nil
					},
				},
			},
			wantError: nil,
		},
		{
			name: "Check flushAndClose, flush fail",
			args: args{
				readCloser: &readCloserMock{
					readMock: func(p []byte) (n int, err error) {
						return 0, fmt.Errorf("read-error-579")
					},
					closeMock: func() error {
						return nil
					},
				},
			},
			wantError: fmt.Errorf("request body flush failed: read-error-579"),
		},
		{
			name: "Check flushAndClose, close fail",
			args: args{
				readCloser: &readCloserMock{
					readMock: func(p []byte) (n int, err error) {
						return 0, io.EOF
					},
					closeMock: func() error {
						return fmt.Errorf("close-error-596")
					},
				},
			},
			wantError: fmt.Errorf("request body close failed: close-error-596"),
		},
		{
			name: "Check flushAndClose, flush & close fail",
			args: args{
				readCloser: &readCloserMock{
					readMock: func(p []byte) (n int, err error) {
						return 0, fmt.Errorf("read-error-607")
					},
					closeMock: func() error {
						return fmt.Errorf("close-error-610")
					},
				},
			},
			wantError: fmt.Errorf("request body flush failed: read-error-607"),
		},
	}

	errToStr := func(err error) string {
		if err != nil {
			return err.Error()
		}
		return ""
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotError := flushAndClose(tt.args.readCloser)
			if !reflect.DeepEqual(errToStr(gotError), errToStr(tt.wantError)) {
				t.Errorf("flushAndClose() error = %v, want %v", gotError, tt.wantError)
			}
		})
	}
}
