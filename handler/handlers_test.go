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

package handler

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/yahoojapan/garm/service"
)

func TestNew(t *testing.T) {
	type args struct {
		a service.Athenz
	}
	tests := []struct {
		name string
		args args
		want Handler
	}{
		{
			name: "Test new athenz handler",
			args: args{
				a: nil,
			},
			want: &handler{
				athenz: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handler_Authenticate(t *testing.T) {
	type fields struct {
		athenz service.Athenz
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "Test return error",
			fields: fields{
				athenz: &mockAthenz{
					AthenzAuthenticatorFunc: func(http.ResponseWriter, *http.Request) error {
						return fmt.Errorf("Test error")
					},
				},
			},
			args:    args{},
			wantErr: fmt.Errorf("Authenticate Webhook Handler failed: Test error"),
		},
		{
			name: "Test return success",
			fields: fields{
				athenz: &mockAthenz{
					AthenzAuthenticatorFunc: func(http.ResponseWriter, *http.Request) error {
						return nil
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &handler{
				athenz: tt.fields.athenz,
			}
			err := h.Authenticate(tt.args.w, tt.args.r)
			if tt.wantErr == nil && err != nil {
				t.Errorf("failed to instantiate, err: %v", err)
				return
			} else if tt.wantErr != nil {
				if tt.wantErr.Error() != err.Error() {
					t.Errorf("error not the same, want: %v, got: %v", tt.wantErr, err)
				}
			}
		})
	}
}

func Test_handler_Authorize(t *testing.T) {
	type fields struct {
		athenz service.Athenz
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "Test return error",
			fields: fields{
				athenz: &mockAthenz{
					AthenzAuthorizerFunc: func(http.ResponseWriter, *http.Request) error {
						return fmt.Errorf("Test")
					},
				},
			},
			args:    args{},
			wantErr: fmt.Errorf("Authorization Webhook Handler failed: Test"),
		},
		{
			name: "Test return success",
			fields: fields{
				athenz: &mockAthenz{
					AthenzAuthorizerFunc: func(http.ResponseWriter, *http.Request) error {
						return nil
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &handler{
				athenz: tt.fields.athenz,
			}
			err := h.Authorize(tt.args.w, tt.args.r)
			if tt.wantErr == nil && err != nil {
				t.Errorf("failed to instantiate, err: %v", err)
				return
			} else if tt.wantErr != nil {
				if tt.wantErr.Error() != err.Error() {
					t.Errorf("error not the same, want: %v, got: %v", tt.wantErr, err)
				}
			}
		})
	}
}
