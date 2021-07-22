package main

import (
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/yahoojapan/garm/config"
)

func TestParseParams(t *testing.T) {
	type test struct {
		name       string
		beforeFunc func()
		checkFunc  func(*params) error
		checkErr   bool
	}
	tests := []test{
		func() test {
			return test{
				name: "check parseParams set default value",
				beforeFunc: func() {
					os.Args = []string{""}
				},
				checkFunc: func(p *params) error {
					if p.configFilePath != "/etc/garm/config.yaml" {
						return errors.Errorf("unexpected file path. got: %s, want: /etc/garm/config.yaml", p.configFilePath)
					}
					if p.showVersion != false {
						return errors.Errorf("unexpected showVersion flag. got: %v, want : false", p.showVersion)
					}
					return nil
				},
				checkErr: false,
			}
		}(),
		func() test {
			return test{
				name: "check parse error",
				checkFunc: func(p *params) error {
					return nil
				},
				beforeFunc: func() {
					os.Args = []string{"", "-="}
				},
				checkErr: true,
			}
		}(),
		func() test {
			return test{
				name: "check parseParams set user flags",
				beforeFunc: func() {
					os.Args = []string{"", "-f", "/dummy/path", "-version", "true"}
				},
				checkFunc: func(p *params) error {
					if p.configFilePath != "/dummy/path" {
						return errors.Errorf("unexpected file path. got: %s, want: /dummy/path", p.configFilePath)
					}
					if p.showVersion != true {
						return errors.Errorf("unexpected showVersion flag. got: %v, want: true", p.showVersion)
					}

					return nil
				},
				checkErr: false,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeFunc != nil {
				tt.beforeFunc()
			}

			got, err := parseParams()
			if err != nil && !tt.checkErr {
				t.Errorf("unexpected error: %v", err)
			}
			if err := tt.checkFunc(got); err != nil {
				t.Errorf("checkFunc() error: %v", err)
			}
		})
	}
}

func Test_run(t *testing.T) {
	type args struct {
		cfg config.Config
	}
	type test struct {
		name      string
		args      args
		checkFunc func(config.Config) error
	}
	tests := []test{
		func() test {
			return test{
				name: "run error",
				args: args{
					cfg: config.Config{
						Token: config.Token{
							AthenzDomain:    "domain",
							ServiceName:     "service",
							RefreshDuration: "1h",
							KeyVersion:      "keyId",
							Expiration:      "1h",
							PrivateKey:      "./service/testdata/dummyServer.key",
						},
						Athenz: config.Athenz{
							Timeout: "dummy",
						},
					},
				},
				checkFunc: func(cfg config.Config) error {
					got := run(cfg)
					want := `failed to instantiate daemon: athenz service instantiate failed: athenz timeout parse failed: time: invalid duration "dummy"`
					if len(got) != 1 {
						return errors.New("len(got) != 1")
					}
					if got[0].Error() != want {
						return errors.Errorf("got: %v, want: %v", got[0], want)
					}
					return nil
				},
			}
		}(),
		func() test {
			return test{
				name: "daemon init error",
				args: args{
					cfg: config.Config{
						Token: config.Token{
							RefreshDuration: "dummy",
						},
					},
				},
				checkFunc: func(cfg config.Config) error {
					got := run(cfg)
					want := `failed to instantiate daemon: token service instantiate failed: invalid token refresh duration dummy: time: invalid duration "dummy"`
					if len(got) != 1 {
						return errors.New("len(got) != 1")
					}
					if got[0].Error() != want {
						return errors.Errorf("got: %v, want: %v", got[0], want)
					}
					return nil
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.checkFunc(tt.args.cfg); err != nil {
				t.Errorf("run() error = %v", err)
			}
		})
	}
}

func Test_getVersion(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "default",
			want: "development version",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getVersion(); got != tt.want {
				t.Errorf("getVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
