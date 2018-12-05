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
	"net/http"

	"github.com/yahoojapan/garm/service"
)

type mockAthenz struct {
	AthenzAuthorizerFunc    Func
	AthenzAuthenticatorFunc Func
}

// NewMockAthenz creates a mock object of service.Athenz.
func NewMockAthenz() service.Athenz {
	return &mockAthenz{}
}

// AthenzAuthorizer returns a mock value of service.Athenz.AthenzAuthenticator() function.
// User should initialize the AthenzAuthorizerFunc function pointer and this function will trigger it.
// Example: mock := &mockAthenz{
//		AthenzAuthenticatorFunc: func(http.ResponseWriter, *http.Request) error {
//			// mock function logic
//		},
//	}
func (a *mockAthenz) AthenzAuthorizer(rw http.ResponseWriter, r *http.Request) error {
	return a.AthenzAuthorizerFunc(rw, r)
}

// AthenzAuthenticator returns a mock value of service.Athenz.AthenzAuthenticator() function.
// User should initialize the AthenzAuthenticatorFunc function pointer and this function will trigger it.
// Example:
// mock := &mockAthenz{
//		AthenzAuthenticatorFunc: func(http.ResponseWriter, *http.Request) error {
//			// mock function logic
//		},
//	}
func (a *mockAthenz) AthenzAuthenticator(rw http.ResponseWriter, r *http.Request) error {
	return a.AthenzAuthenticatorFunc(rw, r)
}
