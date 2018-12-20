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
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	webhook "github.com/yahoo/k8s-athenz-webhook"
	"github.com/yahoojapan/garm/config"

	authn "k8s.io/api/authentication/v1beta1"
	authz "k8s.io/api/authorization/v1beta1"
)

// dummyLogger is a mock implementation for webhook.Logger
type dummyLogger string

// Println is mock method for webhook.Logger interface
func (dummyLogger) Println(args ...interface{}) {}

// Printf is mock method for webhook.Logger interface
func (dummyLogger) Printf(format string, args ...interface{}) {}

// dummyLogger is a mock implementation for webhook.ResourceMapper and webhook.UserMapper
type dummyMapper string

// MapResource is mock method for webhook.ResourceMapper interface
func (dummyMapper) MapResource(ctx context.Context, spec authz.SubjectAccessReviewSpec) (principal string, checks []webhook.AthenzAccessCheck, err error) {
	principal = "principal"
	checks = []webhook.AthenzAccessCheck{
		webhook.AthenzAccessCheck{
			Action:   "action_1",
			Resource: "resource_1",
		},
		webhook.AthenzAccessCheck{
			Action:   "action_2",
			Resource: "resource_2",
		},
	}
	err = nil
	return
}

// MapUser is mock method for webhook.UserMapper interface
func (dummyMapper) MapUser(ctx context.Context, domain, service string) (authn.UserInfo, error) {
	return authn.UserInfo{
		Username: "username",
		UID:      "uid",
		Groups: []string{
			"group_1",
			"group_2",
		},
	}, nil
}

func TestNewAthenz(t *testing.T) {

	type args struct {
		cfg config.Athenz
		log Logger
	}
	type testcase struct {
		name      string
		args      args
		want      Athenz
		wantError error
		checkFunc func(Athenz, Athenz) error
	}
	tests := []testcase{
		func() testcase {
			expectedTimeout, err := time.ParseDuration("3.3s")
			return testcase{
				name: "Check NewAthenz success",
				args: args{
					cfg: config.Athenz{
						AuthHeader:      "auth-header-31",
						URL:             "url-32",
						Timeout:         "3.3s",
						AthenzRootCAKey: "",
						AuthN: webhook.AuthenticationConfig{
							Config: webhook.Config{},
							Mapper: dummyMapper("dummy-mapper-61"),
						},
						AuthZ: webhook.AuthorizationConfig{
							Config:     webhook.Config{},
							Mapper:     dummyMapper("dummy-mapper-67"),
							AthenzX509: nil,
						},
						Config: webhook.Config{},
					},
					log: &logger{
						flgs: -1,
						provider: func(requestID string) webhook.Logger {
							return dummyLogger("dummy-logger-89")
						},
					},
				},
				want: &athenz{
					authConfig: config.Athenz{
						AuthHeader:      "auth-header-31",
						URL:             "url-32",
						Timeout:         "3.3s",
						AthenzRootCAKey: "",
						AuthN: webhook.AuthenticationConfig{
							Config: webhook.Config{
								ZMSEndpoint: "url-32",
								ZTSEndpoint: "url-32",
								AuthHeader:  "auth-header-31",
								Timeout:     expectedTimeout,
								LogProvider: func(requestID string) webhook.Logger {
									return dummyLogger("dummy-logger-89")
								},
								LogFlags: -1,
							},
							Mapper: dummyMapper("dummy-mapper-61"),
						},
						AuthZ: webhook.AuthorizationConfig{
							Config: webhook.Config{
								ZMSEndpoint: "url-32",
								ZTSEndpoint: "url-32",
								AuthHeader:  "auth-header-31",
								Timeout:     expectedTimeout,
								LogProvider: func(requestID string) webhook.Logger {
									return dummyLogger("dummy-logger-89")
								},
								LogFlags: -1,
							},
							Mapper:     dummyMapper("dummy-mapper-67"),
							AthenzX509: nil,
						},
						Config: webhook.Config{},
					},
				},
				wantError: err,
				checkFunc: func(got, want Athenz) error {
					gotConfig := got.(*athenz).authConfig
					wantConfig := want.(*athenz).authConfig
					requestID := "requestID"

					// skip authn checking (external library)
					// skip authz checking (external library)

					// skip authConfig.Authz.AthenzX509 checking (internal function)
					if reflect.TypeOf(gotConfig.AuthZ.AthenzX509).String() != "webhook.IdentityAthenzX509" {
						return fmt.Errorf("NewAthenz() AuthZ.AthenzX509 = %v, want %v", reflect.TypeOf(gotConfig.AuthZ.AthenzX509).String(), "webhook.IdentityAthenzX509")
					}
					gotConfig.AuthZ.AthenzX509 = nil

					// check AuthN.Config.LogProvider
					gotAuthNLogger := gotConfig.AuthN.Config.LogProvider(requestID)
					wantAuthNLogger := wantConfig.AuthN.Config.LogProvider(requestID)
					if gotAuthNLogger != wantAuthNLogger {
						return fmt.Errorf("NewAthenz() AuthN.Config.LogProvider = %v, want %v", gotAuthNLogger, wantAuthNLogger)
					}
					gotConfig.AuthN.Config.LogProvider = nil
					wantConfig.AuthN.Config.LogProvider = nil

					// check AuthZ.Config.LogProvider
					gotAuthZLogger := gotConfig.AuthZ.Config.LogProvider(requestID)
					wantAuthZLogger := wantConfig.AuthZ.Config.LogProvider(requestID)
					if gotAuthZLogger != wantAuthZLogger {
						return fmt.Errorf("NewAthenz() AuthZ.Config.LogProvider = %v, want %v", gotAuthZLogger, wantAuthZLogger)
					}
					gotConfig.AuthZ.Config.LogProvider = nil
					wantConfig.AuthZ.Config.LogProvider = nil

					// check the rest
					if !reflect.DeepEqual(gotConfig, wantConfig) {
						return fmt.Errorf("NewAthenz() authConfig = %v, want %v", gotConfig, wantConfig)
					}
					return nil
				},
			}
		}(),
		{
			name:      "Check NewAthenz fail with nil cfg",
			args:      args{},
			want:      nil,
			wantError: fmt.Errorf("time: invalid duration "),
		},
		{
			name: "Check NewAthenz fail with invalid timeout duration",
			args: args{
				cfg: config.Athenz{
					Timeout: "xxxtimeout",
				},
			},
			want:      nil,
			wantError: fmt.Errorf("time: invalid duration %s", "xxxtimeout"),
		},
		{
			name: "Check NewAthenz fail with unknown timeout unit",
			args: args{
				cfg: config.Athenz{
					Timeout: "10ss",
				},
			},
			want:      nil,
			wantError: fmt.Errorf("time: unknown unit %s in duration %s", "ss", "10ss"),
		},
		{
			name: "Check NewAthenz fail with timeout having no units",
			args: args{
				cfg: config.Athenz{
					Timeout: "99",
				},
			},
			want:      nil,
			wantError: fmt.Errorf("time: missing unit in duration %s", "99"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAthenz(tt.args.cfg, tt.args.log)
			if !reflect.DeepEqual(err, tt.wantError) {
				t.Errorf("NewAthenz() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if tt.checkFunc != nil {
				if err := tt.checkFunc(got, tt.want); err != nil {
					t.Error(err)
				}
			} else {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("NewAthenz() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func Test_athenz_AthenzAuthenticator(t *testing.T) {
	type fields struct {
		authConfig config.Athenz
		authn      http.Handler
		authz      http.Handler
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	cmpResponse := func(got, want *httptest.ResponseRecorder) error {
		if got.Code != want.Code {
			return fmt.Errorf("athenz.AthenzAuthenticator() code = %v, wanted code %v", got.Code, want.Code)
		}
		if !bytes.Equal(got.Body.Bytes(), want.Body.Bytes()) {
			return fmt.Errorf("athenz.AthenzAuthenticator() body = %v, wanted body %v", got.Body.String(), want.Body.String())
		}
		// need to test header?
		return nil
	}

	tests := []struct {
		name      string
		fields    fields
		args      args
		wantError error
		want      *httptest.ResponseRecorder
		checkFunc func(*httptest.ResponseRecorder, *httptest.ResponseRecorder) error
	}{
		{
			name: "Check AthenzAuthenticator success",
			fields: fields{
				authn: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					body, err := ioutil.ReadAll(r.Body)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}

					w.WriteHeader(http.StatusOK)
					w.Header().Set("Content-Type", "text/plain; charset=utf-8")
					_, err = io.WriteString(w, r.URL.String()+" - "+string(body[:]))
					if err != nil {
						t.Error(err)
					}
				}),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("POST", "http://dummy.url", bytes.NewBufferString("dummy athenz authenticator request body")),
			},
			wantError: nil,
			want: &httptest.ResponseRecorder{
				Code: 200,
				Body: bytes.NewBufferString("http://dummy.url - dummy athenz authenticator request body"),
			},
			checkFunc: cmpResponse,
		},
		{
			name: "Check AthenzAuthenticator fail with HTTP error",
			fields: fields{
				authn: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "expected authenticator error", http.StatusInternalServerError)
				}),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("GET", "http://dummy.url", nil),
			},
			wantError: nil,
			want: &httptest.ResponseRecorder{
				Code: 500,
				Body: bytes.NewBufferString("expected authenticator error\n"),
			},
			checkFunc: cmpResponse,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &athenz{
				authConfig: tt.fields.authConfig,
				authn:      tt.fields.authn,
				authz:      tt.fields.authz,
			}
			if err := a.AthenzAuthenticator(tt.args.w, tt.args.r); !reflect.DeepEqual(err, tt.wantError) {
				t.Errorf("athenz.AthenzAuthenticator() error = %v, wantError %v", err, tt.wantError)
			}
			got := tt.args.w.(*httptest.ResponseRecorder)
			if err := tt.checkFunc(got, tt.want); err != nil {
				t.Error(err)
			}
		})
	}
}

func Test_athenz_AthenzAuthorizer(t *testing.T) {
	type fields struct {
		authConfig config.Athenz
		authn      http.Handler
		authz      http.Handler
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	cmpResponse := func(got, want *httptest.ResponseRecorder) error {
		if got.Code != want.Code {
			return fmt.Errorf("athenz.AthenzAuthorizer() code = %v, wanted code %v", got.Code, want.Code)
		}
		if !bytes.Equal(got.Body.Bytes(), want.Body.Bytes()) {
			return fmt.Errorf("athenz.AthenzAuthorizer() body = %v, wanted body %v", got.Body.String(), want.Body.String())
		}
		// need to test header?
		return nil
	}

	tests := []struct {
		name      string
		fields    fields
		args      args
		wantError error
		want      *httptest.ResponseRecorder
		checkFunc func(*httptest.ResponseRecorder, *httptest.ResponseRecorder) error
	}{
		{
			name: "Check AthenzAuthorizer success",
			fields: fields{
				authz: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					body, err := ioutil.ReadAll(r.Body)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}

					w.WriteHeader(http.StatusOK)
					w.Header().Set("Content-Type", "text/plain; charset=utf-8")
					_, err = io.WriteString(w, r.URL.String()+" - "+string(body[:]))
					if err != nil {
						t.Error(err)
					}
				}),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("POST", "http://dummy.url", bytes.NewBufferString("dummy athenz authorizer request body")),
			},
			wantError: nil,
			want: &httptest.ResponseRecorder{
				Code: 200,
				Body: bytes.NewBufferString("http://dummy.url - dummy athenz authorizer request body"),
			},
			checkFunc: cmpResponse,
		},
		{
			name: "Check AthenzAuthorizer fail with HTTP error",
			fields: fields{
				authz: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "expected authorizer error", http.StatusInternalServerError)
				}),
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("GET", "http://dummy.url", nil),
			},
			wantError: nil,
			want: &httptest.ResponseRecorder{
				Code: 500,
				Body: bytes.NewBufferString("expected authorizer error\n"),
			},
			checkFunc: cmpResponse,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &athenz{
				authConfig: tt.fields.authConfig,
				authn:      tt.fields.authn,
				authz:      tt.fields.authz,
			}
			if err := a.AthenzAuthorizer(tt.args.w, tt.args.r); !reflect.DeepEqual(err, tt.wantError) {
				t.Errorf("athenz.AthenzAuthorizer() error = %v, wantError %v", err, tt.wantError)
			}
			got := tt.args.w.(*httptest.ResponseRecorder)
			if err := tt.checkFunc(got, tt.want); err != nil {
				t.Error(err)
			}
		})
	}
}
