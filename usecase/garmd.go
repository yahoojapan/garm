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

package usecase

import (
	"context"

	"github.com/pkg/errors"
	"github.com/yahoojapan/garm/config"
	"github.com/yahoojapan/garm/handler"
	"github.com/yahoojapan/garm/router"
	"github.com/yahoojapan/garm/service"
)

// GarmDaemon represents Garm daemon behavior.
type GarmDaemon interface {
	Start(ctx context.Context) chan []error
}

type garm struct {
	cfg    config.Config
	token  service.TokenService
	athenz service.Athenz
	server service.Server
}

// New returns a Garm daemon, or error occurred.
// The daemon contains a token service authentication and authorization server.
// This function will also initialize the mapping rules for the authentication and authorization check.
func New(cfg config.Config) (GarmDaemon, error) {
	token, err := service.NewTokenService(cfg.Token)
	if err != nil {
		return nil, errors.Wrap(err, "token service instantiate failed")
	}

	resolver := service.NewResolver(cfg.Mapping)
	// set up mapper
	cfg.Athenz.AuthZ.Mapper = service.NewResourceMapper(resolver)
	cfg.Athenz.AuthN.Mapper = service.NewUserMapper(resolver)

	// set token source (function pointer)
	cfg.Athenz.AuthZ.Token = token.GetToken

	athenz, err := service.NewAthenz(cfg.Athenz, service.NewLogger(cfg.Logger))
	if err != nil {
		return nil, errors.Wrap(err, "athenz service instantiate failed")
	}

	return &garm{
		cfg:    cfg,
		token:  token,
		athenz: athenz,
		server: service.NewServer(cfg.Server, router.New(cfg.Server, handler.New(athenz))),
	}, nil
}

// Start returns an error slice channel. This error channel reports the errors inside Garm server.
func (g *garm) Start(ctx context.Context) chan []error {
	g.token.StartTokenUpdater(ctx)
	return g.server.ListenAndServe(ctx)
}
