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
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AthenZ/athenz/libs/go/zmssvctoken"
	"github.com/yahoojapan/garm/config"
)

func TestNewTokenService(t *testing.T) {
	type args struct {
		cfg config.Token
	}
	type test struct {
		name       string
		args       args
		want       TokenService
		beforeFunc func()
		checkFunc  func(TokenService, TokenService) error
		afterFunc  func()
		wantErr    error
	}
	tests := []test{
		{
			name: "Test error invalid refresh duration",
			args: args{
				cfg: config.Token{
					RefreshDuration: "test",
				},
			},
			wantErr: fmt.Errorf("invalid token refresh duration %s: %v", "test", "time: invalid duration test"),
		},
		{
			name: "Test error invalid expiration",
			args: args{
				cfg: config.Token{
					RefreshDuration: "1m",
					Expiration:      "test",
				},
			},
			wantErr: fmt.Errorf("invalid token expiration %s: %v", "test", "time: invalid duration test"),
		},
		func() test {
			keyEnvName := "dummyKey"
			key := "notexists"

			return test{
				name: "Test error private key not exist",
				args: func() args {
					return args{
						cfg: config.Token{
							RefreshDuration: "1m",
							Expiration:      "1m",
							PrivateKey:      "_" + keyEnvName + "_",
						},
					}
				}(),
				beforeFunc: func() {
					os.Setenv(keyEnvName, key)
				},
				afterFunc: func() {
					os.Unsetenv(keyEnvName)
				},
				wantErr: fmt.Errorf("invalid token certificate: open %v", "notexists: no such file or directory"),
			}
		}(),
		func() test {
			keyEnvName := "dummyKey"
			key := "notexists"

			return test{
				name: "Test error private key not valid",
				args: func() args {

					return args{
						cfg: config.Token{
							RefreshDuration: "1m",
							Expiration:      "1m",
							PrivateKey:      "_" + keyEnvName + "_",
							NTokenPath:      "",
						},
					}
				}(),
				beforeFunc: func() {
					os.Setenv(keyEnvName, key)
				},
				afterFunc: func() {
					os.Unsetenv(keyEnvName)
				},
				wantErr: fmt.Errorf("invalid token certificate: open %v", "notexists: no such file or directory"),
			}
		}(),
		func() test {
			keyEnvName := "dummyKey"
			key := "./testdata/dummyServer.key"
			cfg := config.Token{
				AthenzDomain:    keyEnvName,
				ServiceName:     keyEnvName,
				NTokenPath:      "",
				PrivateKey:      "_" + keyEnvName + "_",
				ValidateToken:   false,
				RefreshDuration: "1s",
				KeyVersion:      "1",
				Expiration:      "1s",
			}
			keyData, _ := ioutil.ReadFile(key)
			athenzDomain := config.GetActualValue(cfg.AthenzDomain)
			serviceName := config.GetActualValue(cfg.ServiceName)

			return test{
				name: "Check return value",
				args: args{
					cfg: cfg,
				},
				want: func() TokenService {
					tok, err := (&token{
						token:           new(atomic.Value),
						tokenFilePath:   cfg.NTokenPath,
						validateToken:   cfg.ValidateToken,
						tokenExpiration: time.Second,
						refreshDuration: time.Second,
					}).createTokenBuilder(athenzDomain, serviceName, cfg.KeyVersion, keyData)
					if err != nil {
						panic(err)
					}
					return tok
				}(),
				beforeFunc: func() {
					os.Setenv(keyEnvName, key)
				},
				checkFunc: func(got, want TokenService) error {
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					got.StartTokenUpdater(ctx)
					want.StartTokenUpdater(ctx)
					time.Sleep(time.Millisecond * 100)
					g, err := got.GetToken()
					if err != nil {
						return err
					}
					w, err := want.GetToken()
					if err != nil {
						return err
					}
					parse := func(str string) map[string]string {
						m := make(map[string]string)
						for _, pair := range strings.Split(str, ";") {
							kv := strings.SplitN(pair, "=", 2)
							if len(kv) < 2 {
								continue
							}
							m[kv[0]] = kv[1]
						}
						return m
					}

					gm := parse(g)
					wm := parse(w)

					check := func(key string) bool {
						return gm[key] != wm[key]
					}

					if check("v") || check("d") || check("n") || check("k") || check("h") || check("i") || check("t") || check("e") {
						return fmt.Errorf("invalid token, got: %s, want: %s", g, w)
					}

					return nil
				},
				afterFunc: func() {
					os.Unsetenv(keyEnvName)
				},
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

			got, err := NewTokenService(tt.args.cfg)
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

func Test_token_StartTokenUpdater(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	type test struct {
		name       string
		args       args
		cfg        config.Token
		beforeFunc func()
		checkFunc  func(TokenService) error
		afterFunc  func()
		wantErr    error
	}
	tests := []test{
		func() test {
			keyEnvName := "dummyKey"
			key := "./testdata/dummyServer.key"
			cfg := config.Token{
				AthenzDomain:    keyEnvName,
				ServiceName:     keyEnvName,
				NTokenPath:      "",
				PrivateKey:      "_" + keyEnvName + "_",
				ValidateToken:   false,
				RefreshDuration: "1s",
				KeyVersion:      "1",
				Expiration:      "1s",
			}

			return test{
				name: "Check return value",
				args: args{
					ctx: context.Background(),
				},
				cfg: cfg,
				beforeFunc: func() {
					os.Setenv(keyEnvName, key)
				},
				checkFunc: func(got TokenService) error {
					time.Sleep(time.Millisecond * 100)
					g, err := got.GetToken()
					if err != nil {
						return err
					}
					if len(g) == 0 {
						return fmt.Errorf("invalid token, got: %s", g)
					}

					return nil
				},
				afterFunc: func() {
					os.Unsetenv(keyEnvName)
				},
			}
		}(),
		func() test {
			keyEnvName := "dummyKey"
			key := "./testdata/dummyServer.key"
			cfg := config.Token{
				AthenzDomain:    keyEnvName,
				ServiceName:     keyEnvName,
				NTokenPath:      "",
				PrivateKey:      "_" + keyEnvName + "_",
				ValidateToken:   false,
				RefreshDuration: "2s",
				KeyVersion:      "1",
				Expiration:      "1s",
			}
			ctx, cancel := context.WithCancel(context.Background())

			return test{
				name: "Check context canceled",
				args: args{
					ctx: ctx,
				},
				cfg: cfg,
				beforeFunc: func() {
					os.Setenv(keyEnvName, key)
				},
				checkFunc: func(got TokenService) error {
					time.Sleep(time.Millisecond * 100)
					cancel()

					g, err := got.GetToken()
					if err != nil {
						return err
					}

					time.Sleep(time.Millisecond * 100)

					g2, err := got.GetToken()
					if err != nil {
						return err
					}

					if g != g2 {
						return fmt.Errorf("Context is canceled, but the token refreshed, g: %v\tg2: %v", g, g2)
					}
					return nil
				},
				afterFunc: func() {
					os.Unsetenv(keyEnvName)
					cancel()
				},
			}
		}(),
		func() test {
			keyEnvName := "dummyKey"
			key := "./testdata/dummyServer.key"
			cfg := config.Token{
				AthenzDomain:    keyEnvName,
				ServiceName:     keyEnvName,
				NTokenPath:      "",
				PrivateKey:      "_" + keyEnvName + "_",
				ValidateToken:   false,
				RefreshDuration: "100ms",
				KeyVersion:      "1",
				Expiration:      "1s",
			}

			return test{
				name: "Check token will update periodically ",
				args: args{
					ctx: context.Background(),
				},
				cfg: cfg,
				beforeFunc: func() {
					os.Setenv(keyEnvName, key)
				},
				checkFunc: func(got TokenService) error {
					// wait for the updater to update the token and get the first token
					time.Sleep(time.Millisecond * 100)
					g1, err := got.GetToken()
					if err != nil {
						return err
					}
					if len(g1) == 0 {
						return fmt.Errorf("invalid token, got: %s", g1)
					}

					// sleep again and get the second token
					time.Sleep(time.Millisecond * 120)
					g2, err := got.GetToken()
					if err != nil {
						return err
					}

					if g1 == g2 {
						return fmt.Errorf("Token did not refreshed, got1: %v, got2: %v", g1, g2)
					}

					return nil
				},
				afterFunc: func() {
					os.Unsetenv(keyEnvName)
				},
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

			tok, err := NewTokenService(tt.cfg)
			if err != nil {
				t.Errorf("failed to instantiate, err: %v", err)
				return
			}

			got := tok.StartTokenUpdater(tt.args.ctx)

			if tt.checkFunc != nil {
				err := tt.checkFunc(got)
				if tt.wantErr == nil && err != nil {
					t.Errorf("compare check failed, err: %v", err)
					return
				}
			}
		})
	}
}

func Test_token_GetToken(t *testing.T) {
	type test struct {
		name       string
		cfg        config.Token
		beforeFunc func()
		checkFunc  func(TokenService) error
		afterFunc  func()
		wantErr    error
	}
	tests := []test{
		func() test {
			keyEnvName := "dummyKey"
			key := "./testdata/dummyServer.key"
			cfg := config.Token{
				AthenzDomain:    keyEnvName,
				ServiceName:     keyEnvName,
				NTokenPath:      "",
				PrivateKey:      "_" + keyEnvName + "_",
				ValidateToken:   false,
				RefreshDuration: "1s",
				KeyVersion:      "1",
				Expiration:      "1s",
			}

			return test{
				name: "Check return value",
				cfg:  cfg,
				beforeFunc: func() {
					os.Setenv(keyEnvName, key)
				},
				checkFunc: func(tok TokenService) error {
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					tok.StartTokenUpdater(ctx)
					time.Sleep(time.Millisecond * 50)

					got, err := tok.GetToken()
					if err != nil {
						return err
					}
					if len(got) == 0 {
						return fmt.Errorf("invalid token, got: %s", got)
					}

					return nil
				},
				afterFunc: func() {
					os.Unsetenv(keyEnvName)
				},
			}
		}(),
		func() test {
			keyEnvName := "dummyKey"
			key := "./testdata/dummyServer.key"
			cfg := config.Token{
				AthenzDomain:    keyEnvName,
				ServiceName:     keyEnvName,
				NTokenPath:      "",
				PrivateKey:      "_" + keyEnvName + "_",
				ValidateToken:   false,
				RefreshDuration: "2s",
				KeyVersion:      "1",
				Expiration:      "1s",
			}
			return test{
				name: "Check error",
				cfg:  cfg,
				beforeFunc: func() {
					os.Setenv(keyEnvName, key)
				},
				checkFunc: func(tok TokenService) error {
					got, err := tok.GetToken()
					if got != "" || err == nil {
						return fmt.Errorf("Daemon is not started, but the GetToken didn't return error  got: %v", got)
					}
					return nil
				},
				afterFunc: func() {
					os.Unsetenv(keyEnvName)
				},
				wantErr: ErrTokenNotFound,
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

			tok, err := NewTokenService(tt.cfg)
			if err != nil {
				t.Errorf("failed to instantiate, err: %v", err)
				return
			}

			if tt.checkFunc != nil {
				err := tt.checkFunc(tok)
				if tt.wantErr == nil && err != nil {
					t.Errorf("compare check failed, err: %v", err)
					return
				}
			}
		})
	}
}

func Test_token_createTokenBuilder(t *testing.T) {
	type args struct {
		athenzDomain string
		serviceName  string
		keyVersion   string
		keyData      []byte
	}
	type fields struct {
		tokenFilePath   string
		token           *atomic.Value
		validateToken   bool
		tokenExpiration time.Duration
		refreshDuration time.Duration
	}
	type test struct {
		name       string
		fields     fields
		args       args
		beforeFunc func() error
		checkFunc  func(TokenService) error
		afterFunc  func() error
		wantErr    error
	}
	tests := []test{
		func() test {
			keyData, err := ioutil.ReadFile("./testdata/dummyServer.key")
			if err != nil {
				panic(err)
			}

			return test{
				name: "Check create token builder success",
				fields: fields{
					token:           new(atomic.Value),
					tokenFilePath:   "",
					validateToken:   false,
					tokenExpiration: time.Second,
					refreshDuration: time.Second,
				},
				args: args{
					athenzDomain: "athenz",
					serviceName:  "service",
					keyVersion:   "1",
					keyData:      keyData,
				},
				checkFunc: func(tv TokenService) error {
					if tv.(*token).builder == nil {
						return fmt.Errorf("Token builder is empty")
					}
					return nil
				},
			}
		}(),
		func() test {
			keyData, _ := ioutil.ReadFile("./testdata/emptyfile")

			return test{
				name: "Check error create token builder",
				fields: fields{
					token:           new(atomic.Value),
					tokenFilePath:   "",
					validateToken:   false,
					tokenExpiration: time.Second,
					refreshDuration: time.Second,
				},
				args: args{
					athenzDomain: "athenz",
					serviceName:  "service",
					keyVersion:   "1",
					keyData:      keyData,
				},
				wantErr: fmt.Errorf(`failed to create ZMS SVC Token Builder
AthenzDomain:	athenz
ServiceName:	service
KeyVersion:	1: Unable to create signer: Unable to load private key`),
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

			tok := &token{
				token:           tt.fields.token,
				tokenFilePath:   tt.fields.tokenFilePath,
				validateToken:   tt.fields.validateToken,
				tokenExpiration: tt.fields.tokenExpiration,
				refreshDuration: tt.fields.refreshDuration,
			}

			got, err := tok.createTokenBuilder(tt.args.athenzDomain, tt.args.serviceName, tt.args.keyVersion, tt.args.keyData)

			if tt.checkFunc != nil {
				if e := tt.checkFunc(got); e != nil {
					t.Errorf("createTokenBuilder error, error: %v", e)
					return
				}
			}

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("expected error, but got nil")
					return
				} else if !reflect.DeepEqual(tt.wantErr.Error(), err.Error()) {
					t.Errorf("error not expected, want: %v, got: %v", tt.wantErr, err)
					return
				}
			}
		})
	}
}

func Test_token_loadToken(t *testing.T) {
	type fields struct {
		tokenFilePath   string
		token           *atomic.Value
		validateToken   bool
		tokenExpiration time.Duration
		refreshDuration time.Duration
		builder         zmssvctoken.TokenBuilder
	}
	type test struct {
		name       string
		fields     fields
		beforeFunc func() error
		checkFunc  func(got, want string) error
		afterFunc  func() error
		want       string
		wantErr    error
	}
	tests := []test{
		{
			name: "Test error tokenFilePath is empty (k8s secret)",
			fields: fields{
				token:           new(atomic.Value),
				tokenFilePath:   "",
				validateToken:   false,
				tokenExpiration: time.Second,
				refreshDuration: time.Second,
				builder: func() zmssvctoken.TokenBuilder {
					tb := NewMockTokenBuilder()
					tb.(*mockTokenBuilder).valueFunc = func() (string, error) {
						return "", fmt.Errorf("Error")
					}
					tb.(*mockTokenBuilder).SetExpirationFunc = func(dur time.Duration) {}

					return tb
				}(),
			},
			wantErr: fmt.Errorf("token builder.Token().Value() load failed: Error"),
		},
		{
			name: "Test success tokenFilePath is empty (k8s secret)",
			fields: fields{
				token:           new(atomic.Value),
				tokenFilePath:   "",
				validateToken:   false,
				tokenExpiration: time.Second,
				refreshDuration: time.Second,
				builder: func() zmssvctoken.TokenBuilder {
					tb := NewMockTokenBuilder()
					tb.(*mockTokenBuilder).valueFunc = func() (string, error) {
						return "token", nil
					}
					tb.(*mockTokenBuilder).SetExpirationFunc = func(dur time.Duration) {}

					return tb
				}(),
			},
			checkFunc: func(got, want string) error {
				if got != want {
					return fmt.Errorf("Token mismatch, got: %v, want: %v", got, want)
				}
				return nil
			},
			want: "token",
		},
		{
			name: "Test tokenFilePath not exists error (Copper argos)",
			fields: fields{
				token:           new(atomic.Value),
				tokenFilePath:   "",
				validateToken:   false,
				tokenExpiration: time.Second,
				refreshDuration: time.Second,
				builder: func() zmssvctoken.TokenBuilder {
					tb := NewMockTokenBuilder()
					tb.(*mockTokenBuilder).valueFunc = func() (string, error) {
						return "", fmt.Errorf("open notexists: no such file or directory")
					}
					tb.(*mockTokenBuilder).SetExpirationFunc = func(dur time.Duration) {}

					return tb
				}(),
			},
			wantErr: fmt.Errorf("token builder.Token().Value() load failed: open notexists: no such file or directory"),
		},
		{
			name: "Test tokenFilePath exists (Copper argos)",
			fields: fields{
				token:           new(atomic.Value),
				tokenFilePath:   "./testdata/dummyToken",
				validateToken:   false,
				tokenExpiration: time.Second,
				refreshDuration: time.Second,
				builder: func() zmssvctoken.TokenBuilder {
					tb := NewMockTokenBuilder()
					tb.(*mockTokenBuilder).valueFunc = func() (string, error) {
						return "token", nil
					}
					tb.(*mockTokenBuilder).SetExpirationFunc = func(dur time.Duration) {}

					return tb
				}(),
			},
			checkFunc: func(got, want string) error {
				if got != want {
					return fmt.Errorf("Token mismatch, got: %v, want: %v", got, want)
				}
				return nil
			},
			want: "dummy token",
		},
		{
			name: "Test error validate token",
			fields: fields{
				token:           new(atomic.Value),
				tokenFilePath:   "./testdata/dummyToken",
				validateToken:   true,
				tokenExpiration: time.Second,
				refreshDuration: time.Second,
				builder: func() zmssvctoken.TokenBuilder {
					tb := NewMockTokenBuilder()
					tb.(*mockTokenBuilder).valueFunc = func() (string, error) {
						return "token", nil
					}
					tb.(*mockTokenBuilder).SetExpirationFunc = func(dur time.Duration) {}

					return tb
				}(),
			},
			wantErr: fmt.Errorf("invalid server identity token: bad field in token 'dummy token'"),
		},
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

			tok := &token{
				token:           tt.fields.token,
				tokenFilePath:   tt.fields.tokenFilePath,
				validateToken:   tt.fields.validateToken,
				tokenExpiration: tt.fields.tokenExpiration,
				refreshDuration: tt.fields.refreshDuration,
				builder:         tt.fields.builder,
			}

			got, err := tok.loadToken()

			if tt.checkFunc != nil {
				if e := tt.checkFunc(got, tt.want); e != nil {
					t.Errorf("loadToken error, error: %v", e)
					return
				}
			}

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("expected error, but got nil")
					return
				} else if !reflect.DeepEqual(tt.wantErr.Error(), err.Error()) {
					t.Errorf("error not expected, want: %v, got: %v", tt.wantErr, err)
					return
				}
			}
		})
	}
}

func Test_token_update(t *testing.T) {
	type fields struct {
		tokenFilePath   string
		token           *atomic.Value
		validateToken   bool
		tokenExpiration time.Duration
		refreshDuration time.Duration
		builder         zmssvctoken.TokenBuilder
	}
	tests := []struct {
		name       string
		fields     fields
		beforeFunc func() error
		checkFunc  func(TokenService) error
		afterFunc  func() error
		wantErr    error
	}{
		{
			name: "Test update success",
			fields: fields{
				token:           new(atomic.Value),
				tokenFilePath:   "./testdata/dummyToken",
				validateToken:   false,
				tokenExpiration: time.Second,
				refreshDuration: time.Second,
				builder:         NewMockTokenBuilder(),
			},
			checkFunc: func(tv TokenService) error {
				t, err := tv.GetToken()
				if err != nil {
					return fmt.Errorf("unexpected error when get token, err: %v", err)
				}
				if t == "" {
					return fmt.Errorf("token is empty")
				}
				return nil
			},
		},
		{
			name: "Test update fail",
			fields: fields{
				token:           new(atomic.Value),
				tokenFilePath:   "notexists",
				validateToken:   false,
				tokenExpiration: time.Second,
				refreshDuration: time.Second,
				builder:         NewMockTokenBuilder(),
			},
			wantErr: fmt.Errorf("loadToken failed: load token from filepath failed: open notexists: no such file or directory"),
		},
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

			tok := &token{
				token:           tt.fields.token,
				tokenFilePath:   tt.fields.tokenFilePath,
				validateToken:   tt.fields.validateToken,
				tokenExpiration: tt.fields.tokenExpiration,
				refreshDuration: tt.fields.refreshDuration,
				builder:         tt.fields.builder,
			}

			err := tok.update()

			if tt.checkFunc != nil {
				if e := tt.checkFunc(tok); e != nil {
					t.Errorf("createTokenBuilder error, error: %v", e)
					return
				}
			}

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("expected error, but got nil")
					return
				} else if !reflect.DeepEqual(tt.wantErr.Error(), err.Error()) {
					t.Errorf("error not expected, want: %v, got: %v", tt.wantErr, err)
					return
				}
			}
		})
	}
}

func Test_token_setToken(t *testing.T) {
	type fields struct {
		tokenFilePath   string
		token           *atomic.Value
		validateToken   bool
		tokenExpiration time.Duration
		refreshDuration time.Duration
		builder         zmssvctoken.TokenBuilder
	}
	type args struct {
		token string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		checkFunc func(TokenService, string) error
		want      string
	}{
		{
			name: "Test set token correct",
			fields: fields{
				token: new(atomic.Value),
			},
			args: args{
				token: "token",
			},
			checkFunc: func(tv TokenService, want string) error {
				t, err := tv.GetToken()
				if err != nil {
					return fmt.Errorf("unexpected error when get token, err: %v", err)
				}
				if t != want {
					return fmt.Errorf("Token is not the same, got: %v, want: %v", t, want)
				}
				return nil
			},
			want: "token",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tok := &token{
				tokenFilePath:   tt.fields.tokenFilePath,
				token:           tt.fields.token,
				validateToken:   tt.fields.validateToken,
				tokenExpiration: tt.fields.tokenExpiration,
				refreshDuration: tt.fields.refreshDuration,
				builder:         tt.fields.builder,
			}
			tok.setToken(tt.args.token)

			if tt.checkFunc != nil {
				if err := tt.checkFunc(tok, tt.want); err != nil {
					t.Errorf("setToken error: %v", err)
					return
				}
			}
		})
	}
}
