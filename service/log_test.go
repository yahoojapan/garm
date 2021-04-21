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
	"errors"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/kpango/glg"
	webhook "github.com/yahoo/k8s-athenz-webhook"
	"github.com/yahoojapan/garm/config"
	"github.com/yahoojapan/garm/log"
)

func TestNewLogger(t *testing.T) {
	type args struct {
		cfg config.Logger
	}

	testFilePath := "/tmp/garm_logging_test.txt"
	f, err := os.OpenFile(testFilePath, os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		t.Errorf("tmp file %s open error", testFilePath)
	}
	defer func() {
		err = f.Close()
		if err != nil {
			t.Error(err)
		}
	}()

	tests := []struct {
		name      string
		args      args
		want      Logger
		checkFunc func(Logger, Logger) error
	}{
		{
			name: "Check LogFlags Successful",
			args: args{
				cfg: config.Logger{
					LogPath:  "",
					LogTrace: "server",
				},
			},
			want: &logger{
				flgs: webhook.LogTraceServer,
			},
			checkFunc: func(got, want Logger) error {
				if !reflect.DeepEqual(got.GetLogFlags(), want.GetLogFlags()) {
					return errors.New("LogFlags not equal")
				}
				return nil
			},
		},
		{
			name: "Check Provider Successful",
			args: args{
				cfg: config.Logger{
					LogPath:  "",
					LogTrace: "server",
				},
			},
			want: &logger{
				provider: func(requestID string) webhook.Logger {
					return log.New(os.Stderr, requestID)
				},
			},
			checkFunc: func(got, want Logger) error {
				if reflect.TypeOf(got.GetProvider()) != reflect.TypeOf(want.GetProvider()) {
					return errors.New("LogProvider not equal")
				}
				return nil
			},
		},
		{
			name: "Check LogFile Successful",
			args: args{
				cfg: config.Logger{
					LogPath:  testFilePath,
					LogTrace: "server",
				},
			},
			want: &logger{
				file: f,
			},
			checkFunc: func(got, want Logger) error {
				dummyID := "dummy ID"
				dummyMsg := "dummy msg"
				expectedLogMsg := "[dummy ID]:\tdummy msg\n"
				got.GetProvider()(dummyID).Println(dummyMsg)

				body, err := ioutil.ReadAll(f)
				if err != nil {
					return err
				}

				if !strings.HasSuffix(string(body[:]), expectedLogMsg) {
					return errors.New("FilePath not equal")
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewLogger(tt.args.cfg)
			err := tt.checkFunc(got, tt.want)
			if err != nil {
				t.Errorf("NewLogger() = %v, want %v\nError: %v", got, tt.want, err)
			}
		})
	}
}

func Test_logger_GetProvider(t *testing.T) {

	dummyLogger := log.New(os.Stderr, "")
	dummyProvider := func(requestID string) webhook.Logger {
		return dummyLogger
	}

	type fields struct {
		file     *os.File
		provider webhook.LogProvider
		flgs     webhook.LogFlags
	}
	tests := []struct {
		name      string
		fields    fields
		want      webhook.LogProvider
		checkFunc func(webhook.LogProvider, webhook.LogProvider) error
	}{
		{
			name: "Check GetProvider Successful",
			fields: fields{
				file:     nil,
				provider: dummyProvider,
				flgs:     webhook.LogTraceAthenz,
			},
			want: dummyProvider,
			checkFunc: func(got, want webhook.LogProvider) error {
				gotLogger := got("")
				wantLogger := want("")

				if gotLogger != wantLogger {
					return errors.New("Provider not equal")
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &logger{
				file:     tt.fields.file,
				provider: tt.fields.provider,
				flgs:     tt.fields.flgs,
			}
			got := l.GetProvider()
			err := tt.checkFunc(got, tt.want)
			if err != nil {
				t.Errorf("logger.GetProvider() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_logger_GetLogFlags(t *testing.T) {

	type fields struct {
		file     *os.File
		provider webhook.LogProvider
		flgs     webhook.LogFlags
	}
	tests := []struct {
		name   string
		fields fields
		want   webhook.LogFlags
	}{
		{
			name: "Check GetLogFlags (single) Successful",
			fields: fields{
				file:     nil,
				provider: nil,
				flgs:     webhook.LogTraceServer,
			},
			want: webhook.LogTraceServer,
		},
		{
			name: "Check GetLogFlags (multiple) Successful",
			fields: fields{
				file:     nil,
				provider: nil,
				flgs:     webhook.LogTraceServer | webhook.LogTraceAthenz,
			},
			want: webhook.LogTraceServer | webhook.LogTraceAthenz,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &logger{
				file:     tt.fields.file,
				provider: tt.fields.provider,
				flgs:     tt.fields.flgs,
			}

			if got := l.GetLogFlags(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("logger.GetLogFlags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newLogProvider(t *testing.T) {
	tests := []struct {
		name          string
		writer        io.Writer
		providerParam string
		loggerParam   string
		want          string
	}{
		{
			name:          "Check GetLogFlags (single) Successful",
			writer:        &bytes.Buffer{},
			providerParam: "dummy request ID",
			loggerParam:   "dummy msg",
			want:          "[dummy request ID]:\tdummy msg\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := newLogProvider(tt.writer)
			logger := provider(tt.providerParam)
			logger.Println(tt.loggerParam)

			got := tt.writer.(*bytes.Buffer).String()
			if !strings.HasSuffix(got, tt.want) {
				t.Errorf("newLogProvider() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newLogTraceFlag(t *testing.T) {
	type args struct {
		traces []string
	}
	tests := []struct {
		name          string
		args          args
		want          webhook.LogFlags
		expectedError string
	}{
		{
			name: "Check empty flags",
			args: args{
				traces: []string{},
			},
			// want: 0,
		},
		{
			name: "Check single flag",
			args: args{
				traces: []string{"server"},
			},
			want: webhook.LogTraceServer,
		},
		{
			name: "Check all flags",
			args: args{
				traces: []string{"server", "athenz", "mapping"},
			},
			want: webhook.LogTraceServer | webhook.LogTraceAthenz | webhook.LogVerboseMapping,
		},
		{
			name: "Check invalid flags",
			args: args{
				traces: []string{"athenz", "invalid"},
			},
			want:          webhook.LogTraceAthenz,
			expectedError: "[ERR]:\tunsupported trace event, invalid, ignored\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorBuffer := &bytes.Buffer{}
			if tt.expectedError != "" {
				glg.Get().SetMode(glg.WRITER).SetLevelWriter(glg.ERR, errorBuffer)
			}
			if got := newLogTraceFlag(tt.args.traces); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newLogTraceFlag() = %v, want %v", got, tt.want)
			}
			if tt.expectedError != "" {
				gotError := errorBuffer.String()
				if !strings.HasSuffix(gotError, tt.expectedError) {
					t.Errorf("newLogTraceFlag() gotError = %v, expectedError %v", gotError, tt.expectedError)
				}
			}
		})
	}
}

func Test_logger_Close(t *testing.T) {
	type fields struct {
		filePath string
		provider webhook.LogProvider
		flgs     webhook.LogFlags
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{
			name: "Check file close",
			fields: fields{
				filePath: "/tmp/garm_close_test.txt",
			},
			wantErr: nil,
		},
		{
			name: "Check file close error",
			fields: fields{
				filePath: "",
			},
			wantErr: os.ErrInvalid,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var file *os.File
			if tt.fields.filePath != "" {
				var err error
				file, err = os.OpenFile(tt.fields.filePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
				if err != nil {
					t.Errorf("tmp file %s open error", tt.fields.filePath)
					return
				}
			}
			l := &logger{
				file:     file,
				provider: tt.fields.provider,
				flgs:     tt.fields.flgs,
			}
			if err := l.Close(); err != tt.wantErr {
				t.Errorf("logger.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
