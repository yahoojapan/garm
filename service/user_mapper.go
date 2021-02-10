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
	"fmt"

	webhook "github.com/falz-tino/k8s-athenz-webhook"
	authn "k8s.io/api/authentication/v1beta1"
)

// UserMapper allows for mapping from Athenz principals to k8s objects.
type UserMapper interface {
	webhook.UserMapper
}

type userMapper struct {
	res    Resolver
	groups []string
}

// NewUserMapper returns a UserMapper instance with given Resolver.
func NewUserMapper(resolver Resolver) UserMapper {
	return &userMapper{
		res: resolver,
	}
}

// MapUser returns UserInfo.
// UserInfo contains the username, UID and groups of the user.
func (u *userMapper) MapUser(ctx context.Context, domain, service string) (authn.UserInfo, error) {
	principal := fmt.Sprintf("%s.%s", domain, service)
	return authn.UserInfo{
		Username: principal,
		UID:      principal,
		Groups:   u.groups,
	}, nil
}
