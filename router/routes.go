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
	"net/http"

	"github.com/yahoojapan/garm/handler"
)

type Route struct {
	Name        string
	Methods     []string
	Pattern     string
	HandlerFunc handler.Func
}

// NewRoutes returns routes defined for handling authenticate and authorize requests.
// The authenticate requests will accept for only HTTP POST requests, and the endpoint is /authn.
// The authorize requests will accept for only HTTP POST requests, and the endpoint is /authz.
func NewRoutes(h handler.Handler) []Route {
	return []Route{
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
	}
}
