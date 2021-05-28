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

package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/kpango/glg"
	"github.com/pkg/errors"
	"github.com/yahoojapan/garm/config"
	"github.com/yahoojapan/garm/usecase"
)

// Version is set by the build command via LDFLAGS
var Version string

// params is the data model for Garm command line arguments.
type params struct {
	configFilePath string
	showVersion    bool
}

// parseParams parses command line arguments to params object.
func parseParams() (*params, error) {
	p := new(params)
	f := flag.NewFlagSet(filepath.Base(os.Args[0]), flag.ContinueOnError)
	f.StringVar(&p.configFilePath,
		"f",
		"/etc/garm/config.yaml",
		"garm config yaml file path")
	f.BoolVar(&p.showVersion,
		"version",
		false,
		"show garm version")

	err := f.Parse(os.Args[1:])
	if err != nil {
		return nil, errors.Wrap(err, "Parse Failed")
	}

	return p, nil
}

// run starts the daemon and listens for OS signal.
func run(cfg config.Config) []error {
	if !cfg.EnableColorLogging {
		glg.Get().DisableColor()
	}

	daemon, err := usecase.New(cfg)
	if err != nil {
		return []error{errors.Wrap(err, "failed to instantiate daemon")}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ech := daemon.Start(ctx)
	sigCh := make(chan os.Signal, 1)

	defer func() {
		close(sigCh)
		close(ech)
	}()

	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	for {
		select {
		case <-sigCh:
			cancel()
			err = glg.Warn("garm server shutdown...")
			if err != nil {
				glg.Fatal(err)
			}
			return nil
		case errs := <-ech:
			return errs
		}
	}
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			if _, ok := err.(runtime.Error); ok {
				panic(err)
			}
			err = glg.Error(err)
			if err != nil {
				glg.Fatal(err)
			}
		}
	}()

	// no need for colorized output, disable colorized logging
	glg.Get().DisableColor()

	p, err := parseParams()
	if err != nil {
		glg.Fatal(err)
		return
	}

	if p.showVersion {
		err := glg.Infof("garm version -> %s", getVersion())
		if err != nil {
			glg.Fatal(err)
		}
		err = glg.Infof("garm config version -> %s", config.GetVersion())
		if err != nil {
			glg.Fatal(err)
		}
		return
	}

	cfg, err := config.New(p.configFilePath)
	if err != nil {
		glg.Fatal(err)
		return
	}

	// check versions between configuration file and config.go
	if cfg.Version != config.GetVersion() {
		glg.Fatal(errors.New("invalid garm config version"))
		return
	}

	errs := run(*cfg)
	if len(errs) > 0 {
		var emsg string
		for _, err = range errs {
			emsg += "\n" + err.Error()
		}
		glg.Fatal(emsg)
		return
	}
}

func getVersion() string {
	if Version == "" {
		return "development version"
	}
	return Version
}
