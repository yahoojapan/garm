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
	"time"

	"github.com/yahoo/athenz/libs/go/zmssvctoken"
)

type mockTokenBuilder struct {
	valueFunc         valueFunc // mock function of zmssvctoken.TokenBuilder.Value()
	SetExpirationFunc SetExpirationFunc
	SetHostnameFunc   SetHostnameFunc
	SetIPAddressFunc  SetIPAddressFunc
	SetKeyServiceFunc SetKeyServiceFunc
}

type valueFunc func() (string, error)
type SetExpirationFunc func(t time.Duration)
type SetHostnameFunc func(h string)
type SetIPAddressFunc func(ip string)
type SetKeyServiceFunc func(keyService string)

type mockToken struct {
	valueFunc valueFunc
}

// NewMockTokenBuilder creates a mock object of zmssvctoken.TokenBuilder.
func NewMockTokenBuilder() zmssvctoken.TokenBuilder {
	return &mockTokenBuilder{}
}

// NewMockToken creates a mock object of zmssvctoken.Token.
func NewMockToken() zmssvctoken.Token {
	return &mockToken{}
}

// SetExpiration returns a mock value of zmssvctoken.TokenBuilder.SetExpiration() function.
func (mt *mockTokenBuilder) SetExpiration(t time.Duration) {
	mt.SetExpirationFunc(t)
}

// SetHostname returns a mock value of zmssvctoken.TokenBuilder.SetHostname() function.
func (mt *mockTokenBuilder) SetHostname(h string) {
	mt.SetHostnameFunc(h)
}

// SetIPAddress returns a mock value of zmssvctoken.TokenBuilder.SetIPAddress() function.
func (mt *mockTokenBuilder) SetIPAddress(ip string) {
	mt.SetIPAddressFunc(ip)
}

// SetKeyService returns a mock value of zmssvctoken.TokenBuilder.SetKeyService() function.
func (mt *mockTokenBuilder) SetKeyService(keyService string) {
	mt.SetKeyServiceFunc(keyService)
}

// Token returns a mock object of zmssvctoken.Token.
func (mt *mockTokenBuilder) Token() zmssvctoken.Token {
	t := &mockToken{}
	t.valueFunc = mt.valueFunc
	return t
}

// Value returns a mock value of zmssvctoken.TokenBuilder.Value() function.
// Example:
// tb := NewMockTokenBuilder()
// tb.(*mockTokenBuilder).valueFunc = func() (string, error) {
//	 // mock function logic
// }
func (mt *mockToken) Value() (string, error) {
	return mt.valueFunc()
}
