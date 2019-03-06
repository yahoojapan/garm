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
	"crypto/tls"
	"net/http"
	"time"

	"github.com/pkg/errors"
	webhook "github.com/yahoo/k8s-athenz-webhook"
	"github.com/yahoojapan/garm/config"
)

// Athenz interface is used to send HTTP requests to Athenz server.
type Athenz interface {
	// AthenzAuthorizer sends HTTP requests to Athenz server for authorization.
	AthenzAuthorizer(http.ResponseWriter, *http.Request) error
	// AthenzAuthenticator sends HTTP requests to Athenz server for authentication.
	AthenzAuthenticator(http.ResponseWriter, *http.Request) error
}

// Wrapper for Athenz HTTP request handlers
type athenz struct {
	// authConfig is the shared configuration for the HTTP handlers.
	authConfig config.Athenz
	// authn is Athenz authenticator.
	authn http.Handler
	// authz is Athenz authorizer.
	authz http.Handler
}

// NewAthenz creates a new Athenz object that can handle HTTP requests based on the given configuration.
// The HTTP handlers will use the given logger for logging.
func NewAthenz(cfg config.Athenz, log Logger) (Athenz, error) {
	athenzTimeout, err := time.ParseDuration(cfg.Timeout)
	if err != nil {
		return nil, errors.Wrap(err, "athenz timeout parse failed")
	}

	c := webhook.Config{
		ZMSEndpoint: cfg.URL,
		ZTSEndpoint: cfg.URL,
		AuthHeader:  cfg.AuthHeader,
		Timeout:     athenzTimeout,
		LogProvider: log.GetProvider(),
		LogFlags:    log.GetLogFlags(),
	}
	cfg.AuthN.Config = c
	cfg.AuthZ.Config = c
	cfg.AuthZ.AthenzX509 = func() (*tls.Config, error) {
		pool, err := NewX509CertPool(config.GetActualValue(cfg.AthenzRootCAKey))
		if err != nil {
			err = errors.Wrap(err, "authorization x509 certpool error")
		}
		return &tls.Config{RootCAs: pool}, err
	}

	return &athenz{
		authConfig: cfg,
		authn:      webhook.NewAuthenticator(cfg.AuthN),
		authz:      webhook.NewAuthorizer(cfg.AuthZ),
	}, nil
}

// AthenzAuthenticator passes the request to a.authn HTTP handler to handle.
func (a *athenz) AthenzAuthenticator(w http.ResponseWriter, r *http.Request) error {
	a.authn.ServeHTTP(w, r)
	return nil
}

// AthenzAuthorizer passes the request to a.authz HTTP handler to handle.
func (a *athenz) AthenzAuthorizer(w http.ResponseWriter, r *http.Request) error {
	a.authz.ServeHTTP(w, r)
	return nil
}
