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

package router

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/yahoojapan/garm/handler"
)

func TestNewRoutes(t *testing.T) {
	type args struct {
		h handler.Handler
	}
	type test struct {
		name      string
		args      args
		checkFunc func(got, want []Route) error
		want      []Route
	}
	tests := []test{
		func() test {
			h := handler.New(nil)

			return test{
				name: "Test success NewRoutes",
				args: args{
					h: h,
				},
				checkFunc: func(got, want []Route) error {
					for i, g := range got {
						w := want[i]
						if g.Name != w.Name || !reflect.DeepEqual(g.Methods, w.Methods) || g.Pattern != w.Pattern ||
							reflect.ValueOf(g.HandlerFunc).Pointer() != reflect.ValueOf(w.HandlerFunc).Pointer() {
							return fmt.Errorf("got not equals want, got: %v, want: %v", g, w)
						}
					}

					return nil
				},
				want: []Route{
					{
						"Authenticate",
						[]string{
							http.MethodPost,
						},
						"/authn",
						h.Authenticate,
					},
					{
						"Authorize",
						[]string{
							http.MethodPost,
						},
						"/authz",
						h.Authorize,
					},
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRoutes(tt.args.h)
			if err := tt.checkFunc(got, tt.want); err != nil {
				t.Errorf("NewRoutes error: %v", err)
				return
			}
		})
	}
}
