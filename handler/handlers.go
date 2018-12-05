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

// Handler is an interface to handle authentication and authorization requests.
type Handler interface {
	Authenticate(http.ResponseWriter, *http.Request) error
	Authorize(http.ResponseWriter, *http.Request) error
}

// Func is HTTP request handler function with error return.
type Func func(http.ResponseWriter, *http.Request) error

type handler struct {
	athenz service.Athenz
}

const (
	ContentType     = "Content-Type"
	ApplicationJSON = "application/json"
	TextPlain       = "text/plain"
	CharsetUTF8     = "charset=UTF-8"
)

// New returns a Handler with the given Athenz service.
func New(a service.Athenz) Handler {
	return &handler{
		athenz: a,
	}
}

// Authenticate returns an error if any.
// The function will handle HTTP request, authenticate the N-token, and write the result into ResponseWriter.
func (h *handler) Authenticate(w http.ResponseWriter, r *http.Request) error {
	return h.athenz.AthenzAuthenticator(w, r)
}

// Authorize returns an error if any.
// The function will handle HTTP request, authorize the result, and write the result into ResponseWriter.
func (h *handler) Authorize(w http.ResponseWriter, r *http.Request) error {
	return h.athenz.AthenzAuthorizer(w, r)
}
