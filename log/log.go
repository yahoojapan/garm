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

package log

import (
	"io"

	"github.com/kpango/glg"
	webhook "github.com/yahoo/k8s-athenz-webhook"
)

// Logger represents a logger interface for logging.
type Logger interface {
	webhook.Logger
}

type logger struct {
	log *glg.Glg
}

// New returns a new Logger instance.
// The logger will automatically append a request ID as prefix, and write logs to the writer.
func New(w io.Writer, requestID string) Logger {
	return &logger{
		log: glg.New().
			SetPrefix(glg.PRINT, requestID).
			SetLevelWriter(glg.PRINT, w).
			SetLevelMode(glg.PRINT, glg.WRITER),
	}
}

// Printf prints the formatted string and the corresponding object values to the logger.
func (l *logger) Printf(format string, args ...interface{}) {
	err := l.log.Printf(format, args...)
	if err != nil {
		glg.Fatal(err)
	}
}

// Println prints all the object values to the logger.
func (l *logger) Println(args ...interface{}) {
	err := l.log.Println(args...)
	if err != nil {
		glg.Fatal(err)
	}
}
