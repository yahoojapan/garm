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
	"io/ioutil"
	"os"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/kpango/glg"
	"github.com/pkg/errors"
	"github.com/yahoo/athenz/libs/go/zmssvctoken"
	webhook "github.com/yahoo/k8s-athenz-webhook"
	"github.com/yahoojapan/garm/config"
)

// TokenService represents an interface for user to get the token, and automatically update the token.
type TokenService interface {
	StartTokenUpdater(context.Context) TokenService
	GetToken() (string, error)
	createTokenBuilder(string, string, string, []byte) (TokenService, error)
}

type token struct {
	tokenFilePath   string
	token           *atomic.Value
	validateToken   bool
	tokenExpiration time.Duration
	refreshDuration time.Duration
	builder         zmssvctoken.TokenBuilder
}

var (
	// ErrTokenNotFound represents the error that the token is not found
	ErrTokenNotFound = errors.New("Error:\ttoken not found")
)

// NewTokenService returns a TokenService.
// It initializes information that required to generate the token (for example, RefreshDuration, Expiration, PrivateKey, etc).
func NewTokenService(cfg config.Token) (TokenService, error) {
	dur, err := time.ParseDuration(cfg.RefreshDuration) // example: 1s, 1m
	if err != nil {
		return nil, errors.Wrapf(err, "invalid token refresh duration %s", cfg.RefreshDuration)
	}

	exp, err := time.ParseDuration(cfg.Expiration) // example: 1s, 1m
	if err != nil {
		return nil, errors.Wrapf(err, "invalid token expiration %s", cfg.Expiration)
	}

	keyData, err := ioutil.ReadFile(os.Getenv(cfg.PrivateKeyEnvName))
	if err != nil && keyData == nil {
		if cfg.NTokenPath == "" {
			return nil, errors.Wrap(err, "invalid token certificate")
		}
	}

	athenzDomain := config.GetActualValue(cfg.AthenzDomain)
	serviceName := config.GetActualValue(cfg.ServiceName)

	return (&token{
		token:           new(atomic.Value),
		tokenFilePath:   cfg.NTokenPath,
		validateToken:   cfg.ValidateToken,
		tokenExpiration: exp,
		refreshDuration: dur,
	}).createTokenBuilder(athenzDomain, serviceName, cfg.KeyVersion, keyData)
}

// StartTokenUpdater returns a TokenService.
// It starts a go routine to update the token periodically.
func (t *token) StartTokenUpdater(ctx context.Context) TokenService {
	go func() {
		err := t.update()
		if err != nil {
			err = glg.Error(errors.Wrap(err, "token first update failed"))
			if err != nil {
				glg.Fatal(err)
			}
		}

		ticker := time.NewTicker(t.refreshDuration)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				err = t.update()
				if err != nil {
					err = glg.Error(errors.Wrap(err, "token update failed"))
					if err != nil {
						glg.Fatal(err)
					}
				}
			}
		}
	}()
	return t
}

// GetToken returns a token string or error
// It is thread-safe and returns the token stored in an atomic variable, or returns the error when the token is not initialized or cannot be generated.
func (t *token) GetToken() (string, error) {
	tok := t.token.Load()
	if tok == nil {
		return "", ErrTokenNotFound
	}
	return tok.(string), nil
}

// createTokenBuilder returns a TokenService or error.
// It initializes a token builder with Athenz domain, service name, key version and the signature private key
// , then returns a TokenService containing the token builder.
func (t *token) createTokenBuilder(athenzDomain, serviceName, keyVersion string, keyData []byte) (TokenService, error) {
	builder, err := zmssvctoken.NewTokenBuilder(
		athenzDomain,
		serviceName,
		keyData,
		keyVersion)

	if err != nil {
		return nil, errors.Wrapf(err, "failed to create ZMS SVC Token Builder\nAthenzDomain:\t%s\nServiceName:\t%s\nKeyVersion:\t%s", athenzDomain, serviceName, keyVersion)
	}

	t.builder = builder

	return t, nil
}

// loadToken returns a n-token string, or error.
// It returns n-token, which is generated with the token builder. If the ntoken_path is set in the configuration YAML (Copper Argos),
// it directly returns the token file content.
// If ntoken_path is not set (k8s secret), the builder reads the private key file path from environment variable (private_key_env_name), and then generates and signs a new token.
// It can also validate the token generated or read. If validate_token flag is true, it verifies the token.
func (t *token) loadToken() (ntoken string, err error) {
	if t.tokenFilePath == "" {
		// k8s secret
		t.builder.SetExpiration(t.tokenExpiration)

		// generate new token from token builder
		ntoken, err = t.builder.Token().Value()
		if err != nil {
			return "", errors.Wrap(err, "token builder.Token().Value() load failed")
		}

	} else {
		// Copper Argos
		tok, err := ioutil.ReadFile(t.tokenFilePath)
		if err != nil {
			return "", errors.Wrap(err, "load token from filepath failed")
		}

		ntoken = strings.TrimRight(*(*string)(unsafe.Pointer(&tok)), "\r\n")
	}

	if t.validateToken {
		err = webhook.VerifyToken(ntoken, false)
		if err != nil {
			return "", errors.Wrap(err, "invalid server identity token")
		}
	}
	return ntoken, nil
}

// update returns error.
// It generates a token from loadToken() function, and stores into memory, and returns if any errors occurred.
func (t *token) update() error {
	token, err := t.loadToken()
	if err != nil {
		return errors.Wrap(err, "loadToken failed")
	}
	t.setToken(token)
	return nil
}

// setToken set the given token as internal token.
func (t *token) setToken(token string) {
	t.token.Store(token)
}
