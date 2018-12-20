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
	"io"
	"os"
	"strings"

	"github.com/kpango/glg"
	"github.com/yahoo/k8s-athenz-webhook"
	"github.com/yahoojapan/garm/config"
	"github.com/yahoojapan/garm/log"
)

// Logger is an intermediate interface to create an actual logger.
type Logger interface {
	// GetProvider returns the LogProvider function that returns the actual logger instance.
	GetProvider() webhook.LogProvider
	// GetLogFlags returns the LogFlags for log filter inside the actual logger instance.
	GetLogFlags() webhook.LogFlags
	// Close closes the output resources used by the logger.
	Close() error
}

// logger implements Logger and holds required settings and runtime data.
type logger struct {
	// file is output destination of the logger.
	file *os.File
	// provider is a function that can return the actual logger object for logging.
	provider webhook.LogProvider
	// flgs is the logger's setting for log filtering.
	flgs webhook.LogFlags
}

// NewLogger creates a Logger based on the input configuration.
func NewLogger(cfg config.Logger) Logger {
	// the default log destination is standard error
	w := os.Stderr
	// change the log destination if a file can be used
	if cfg.LogPath != "" {
		f, err := os.OpenFile(cfg.LogPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err == nil {
			w = f
		}
	}
	return &logger{
		file:     w,
		provider: newLogProvider(w),
		// "server,athenz" => webhook.LogFlags
		flgs: newLogTraceFlag(strings.Split(strings.ToLower(cfg.LogTrace), ",")),
	}
}

// GetProvider returns the internal LogProvider.
func (l *logger) GetProvider() webhook.LogProvider {
	return l.provider
}

// GetLogFlags returns the internal LogFlags.
func (l *logger) GetLogFlags() webhook.LogFlags {
	return l.flgs
}

// newLogProvider creates a LogProvider that make use of the given Writer.
func newLogProvider(w io.Writer) webhook.LogProvider {
	return func(requestID string) webhook.Logger {
		return log.New(w, requestID)
	}
}

// newLogTraceFlag parses the input strings to corresponding LogFlags.
func newLogTraceFlag(traces []string) webhook.LogFlags {
	var flgs webhook.LogFlags
	for _, t := range traces {
		switch t {
		case "server":
			flgs |= webhook.LogTraceServer
		case "athenz":
			flgs |= webhook.LogTraceAthenz
		case "mapping":
			flgs |= webhook.LogVerboseMapping
		default:
			err := glg.Errorf("unsupported trace event, %v, ignored", t)
			if err != nil {
				glg.Fatal(err)
			}
		}
	}
	return flgs
}

// Close closes the file holding by the logger.
func (l *logger) Close() error {
	return l.file.Close()
}
