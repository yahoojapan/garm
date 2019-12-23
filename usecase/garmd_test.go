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

package usecase

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/kpango/glg"

	"github.com/yahoojapan/garm/config"
	"github.com/yahoojapan/garm/handler"
	"github.com/yahoojapan/garm/router"
	"github.com/yahoojapan/garm/service"
)

func TestNew(t *testing.T) {
	glg.Get().SetLevelMode(glg.ERR, glg.NONE)
	type args struct {
		cfg config.Config
	}
	type test struct {
		name       string
		args       args
		beforeFunc func()
		checkFunc  func(GarmDaemon, GarmDaemon) error
		afterFunc  func()
		want       GarmDaemon
		wantErr    error
	}
	tests := []test{
		{
			name: "Check error when new token service",
			args: args{
				cfg: config.Config{
					Token: config.Token{},
				},
			},
			wantErr: fmt.Errorf("token service instantiate failed: invalid token refresh duration : time: invalid duration "),
		},
		func() test {
			keyKey := "dummyKey"
			key := "../service/testdata/dummyServer.key"

			return test{
				name: "Check error when new server",
				args: args{
					cfg: config.Config{
						Token: config.Token{
							AthenzDomain:      keyKey,
							ServiceName:       keyKey,
							PrivateKeyEnvName: "_" + keyKey + "_",
							ValidateToken:     false,
							RefreshDuration:   "1m",
							KeyVersion:        "1",
							Expiration:        "1m",
						},
					},
				},
				beforeFunc: func() {
					os.Setenv(keyKey, key)
				},
				afterFunc: func() {
					os.Unsetenv(keyKey)
				},
				wantErr: fmt.Errorf("athenz service instantiate failed: athenz timeout parse failed: time: invalid duration "),
			}
		}(),
		func() test {
			key := "../service/testdata/dummyServer.key"
			cfg := config.Config{
				Token: config.Token{
					AthenzDomain:      "dummyDomain",
					ServiceName:       "dummyService",
					NTokenPath:        "",
					PrivateKeyEnvName: key,
					ValidateToken:     false,
					RefreshDuration:   "1m",
					KeyVersion:        "1",
					Expiration:        "1m",
				},
				Athenz: config.Athenz{
					Timeout: "1m",
					URL:     "/",
				},
				Server: config.Server{
					HealthzPath: "/",
				},
			}

			return test{
				name: "Check new garm daemon return correct",
				args: args{
					cfg: cfg,
				},
				want: func() GarmDaemon {
					token, err := service.NewTokenService(cfg.Token)
					if token == nil || err != nil {
						t.Errorf("%v", err)
						t.Fail()
					}

					resolver := service.NewResolver(cfg.Mapping)
					cfg.Athenz.AuthZ.Mapper = service.NewResourceMapper(resolver)
					cfg.Athenz.AuthN.Mapper = service.NewUserMapper(resolver)
					cfg.Athenz.AuthZ.Token = token.GetToken
					athenz, _ := service.NewAthenz(cfg.Athenz, service.NewLogger(cfg.Logger))

					server := service.NewServer(cfg.Server, router.New(cfg.Server, handler.New(athenz)))
					return &garm{
						cfg:    cfg,
						token:  token,
						athenz: athenz,
						server: server,
					}

				}(),
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeFunc != nil {
				tt.beforeFunc()
			}
			if tt.afterFunc != nil {
				defer tt.afterFunc()
			}

			got, err := New(tt.args.cfg)
			if tt.wantErr == nil && err != nil {
				t.Errorf("failed to instantiate, err: %v", err)
				return
			} else if tt.wantErr != nil {
				if tt.wantErr.Error() != err.Error() {
					t.Errorf("error not the same, want: %v, got: %v", tt.wantErr, err)
				}
			}

			if tt.checkFunc != nil {
				err = tt.checkFunc(got, tt.want)
				if tt.wantErr == nil && err != nil {
					t.Errorf("compare check failed, err: %v", err)
					return
				}
			}
		})
	}
}

func Test_garm_Start(t *testing.T) {
	glg.Get().SetLevelMode(glg.ERR, glg.NONE)
	type fields struct {
		cfg    config.Config
		token  service.TokenService
		athenz service.Athenz
		server service.Server
	}
	type args struct {
		ctx context.Context
	}
	type test struct {
		name       string
		fields     fields
		args       args
		beforeFunc func()
		checkFunc  func(chan []error, []error) error
		afterFunc  func()
		want       []error
	}
	tests := []test{
		func() test {
			keyKey := "dummy_key"
			key := "../service/testdata/dummyServer.key"
			certKey := "dummy_cert"
			cert := "../service/testdata/dummyServer.crt"

			cfg := config.Config{
				Token: config.Token{
					AthenzDomain:      "dummyDomain",
					ServiceName:       "dummyService",
					NTokenPath:        "",
					PrivateKeyEnvName: key,
					ValidateToken:     false,
					RefreshDuration:   "1m",
					KeyVersion:        "1",
					Expiration:        "1m",
				},
				Athenz: config.Athenz{
					Timeout: "1m",
					URL:     "/",
				},
				Server: config.Server{
					HealthzPath: "/",
					TLS: config.TLS{
						Enabled: true,
						CertKey: certKey,
						KeyKey:  keyKey,
					},
				},
			}
			ctx, cancelFunc := context.WithCancel(context.Background())

			return test{
				name: "Check success start garm daemon",
				args: args{
					ctx: ctx,
				},
				fields: func() fields {
					os.Setenv(certKey, cert)
					os.Setenv(keyKey, key)
					defer func() {
						os.Unsetenv(keyKey)
						os.Unsetenv(certKey)
					}()

					token, _ := service.NewTokenService(cfg.Token)

					resolver := service.NewResolver(cfg.Mapping)
					cfg.Athenz.AuthZ.Mapper = service.NewResourceMapper(resolver)
					cfg.Athenz.AuthN.Mapper = service.NewUserMapper(resolver)
					cfg.Athenz.AuthZ.Token = token.GetToken
					athenz, _ := service.NewAthenz(cfg.Athenz, service.NewLogger(cfg.Logger))

					server := service.NewServer(cfg.Server, router.New(cfg.Server, handler.New(athenz)))
					return fields{
						cfg:    cfg,
						token:  token,
						athenz: athenz,
						server: server,
					}
				}(),
				afterFunc: func() {
					cancelFunc()
				},
				checkFunc: func(got chan []error, want []error) error {
					cancelFunc()
					time.Sleep(time.Second * 3)
					gotErr := <-got

					if !reflect.DeepEqual(gotErr, want) {
						return fmt.Errorf("Got: %v, want: %v", gotErr, want)
					}
					return nil

				},
				want: []error{context.Canceled},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.afterFunc != nil {
				defer tt.afterFunc()
			}

			if tt.beforeFunc != nil {
				tt.beforeFunc()
			}

			fields := tt.fields
			g := &garm{
				cfg:    fields.cfg,
				token:  fields.token,
				athenz: fields.athenz,
				server: fields.server,
			}
			got := g.Start(tt.args.ctx)
			if err := tt.checkFunc(got, tt.want); err != nil {
				t.Errorf("Start function error: %v", err)
			}
		})
	}
}
