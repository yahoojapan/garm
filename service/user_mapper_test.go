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
	"reflect"
	"testing"

	authn "k8s.io/api/authentication/v1beta1"
)

func TestNewUserMapper(t *testing.T) {
	type args struct {
		resolver Resolver
	}
	tests := []struct {
		name string
		args args
		want UserMapper
	}{
		{
			name: "Test UserMapper includes reslover",
			args: args{
				resolver: &resolve{},
			},
			want: &userMapper{
				res: &resolve{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewUserMapper(tt.args.resolver); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewUserMapper() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_userMapper_MapUser(t *testing.T) {
	type fields struct {
		res    Resolver
		groups []string
	}
	type args struct {
		ctx     context.Context
		domain  string
		service string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		checkFunc func(got, want authn.UserInfo) error
		want      authn.UserInfo
		wantErr   error
	}{
		{
			name: "Test UserInfo return",
			fields: fields{
				res:    nil,
				groups: nil,
			},
			args: args{
				ctx:     nil,
				domain:  "testdomain",
				service: "testservice",
			},
			want: authn.UserInfo{
				Username: "testdomain.testservice",
				UID:      "testdomain.testservice",
				Groups:   nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &userMapper{
				res:    tt.fields.res,
				groups: tt.fields.groups,
			}
			got, err := u.MapUser(tt.args.ctx, tt.args.domain, tt.args.service)
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
