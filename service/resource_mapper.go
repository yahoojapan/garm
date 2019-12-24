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

	webhook "github.com/yahoo/k8s-athenz-webhook"
	authz "k8s.io/api/authorization/v1beta1"
)

// ResourceMapper allows for mapping from an authorization request to Athenz principals.
// Wrapper of webhook.ResourceMapper.
type ResourceMapper interface {
	webhook.ResourceMapper
}

// resourceMapper is a ResourceMapper implementation using Resolver to map resources.
type resourceMapper struct {
	// res is a Resolver with mapping rules for resources mapping.
	res Resolver
}

// athenzAccessCheckParam is parameters for creating accesscheck.
// Each member is a mapped string.
type athenzAccessCheckParam struct {
	action      string
	group       string
	resource    string
	name        string
	adminDomain string
	domains     []string
}

// NewResourceMapper creates a new ResourceMapper for mapping K8s resources to Athenz principals.
func NewResourceMapper(resolver Resolver) ResourceMapper {
	return &resourceMapper{
		res: resolver,
	}
}

// MapResource maps K8s access request object to Athenz access request object.
// 1. check is non-resources group or not
// 2. replace the value based on internal resolver configuration according
// 3. value mapping using internal resolver
// 4. get Athenz domains
// 5. get Athenz user
// 6. create Athenz principal based on internal resolver configuration (directly reject, admin domain, user domain)
func (m *resourceMapper) MapResource(ctx context.Context, spec authz.SubjectAccessReviewSpec) (string, []webhook.AthenzAccessCheck, error) {
	var verb, namespace, group, resource, sub, name string

	if spec.ResourceAttributes != nil {
		name = spec.ResourceAttributes.Name
		namespace = spec.ResourceAttributes.Namespace
		if namespace == "" {
			namespace = m.res.GetEmptyNamespace()
		}
		verb = spec.ResourceAttributes.Verb
		resource = spec.ResourceAttributes.Resource
		sub = spec.ResourceAttributes.Subresource
		if sub != "" {
			resource = fmt.Sprintf("%s.%s", resource, sub)
		}
		group = spec.ResourceAttributes.Group
	} else {
		group = m.res.GetNonResourceGroup()
		verb = spec.NonResourceAttributes.Verb
		resource = spec.NonResourceAttributes.Path
		namespace = m.res.GetNonResourceNamespace()
	}

	accessCheckParam := athenzAccessCheckParam{
		action:      m.res.MapVerbAction(verb),
		group:       m.res.MapAPIGroup(group),
		resource:    m.res.MapK8sResourceAthenzResource(resource),
		name:        m.res.MapResourceName(name),
		adminDomain: m.res.GetAdminDomain(namespace),
		domains:     m.res.BuildDomainsFromNamespace(namespace),
	}

	identity := m.res.PrincipalFromUser(spec.User)

	switch {
	case !m.res.IsAllowed(verb, namespace, group, resource, name): // Not Allowed
		return "", nil,
			fmt.Errorf(
				"----%s's request is not allowed----\nVerb:\t%s\nNamespaceb:\t%s\nAPI Group:\t%s\nResource:\t%s\nResource Name:\t%s\n",
				identity, verb, namespace, group, resource, name)
	case m.res.IsAdminAccess(accessCheckParam.action, namespace, accessCheckParam.group, accessCheckParam.resource, accessCheckParam.name):
		return identity, m.createAdminAccessCheck(accessCheckParam), nil
	default:
		return identity, m.createAccessCheck(accessCheckParam), nil
	}
}

func (m *resourceMapper) createAdminAccessCheck(accessCheckParam athenzAccessCheckParam) []webhook.AthenzAccessCheck {
	accessChecks := make([]webhook.AthenzAccessCheck, 0, len(accessCheckParam.domains)*2)
	for _, domain := range accessCheckParam.domains {
		accessChecks = append(accessChecks,
			webhook.AthenzAccessCheck{
				Resource: m.res.TrimResource(fmt.Sprintf("%s:%s.%s.%s.%s", accessCheckParam.adminDomain, accessCheckParam.group, domain, accessCheckParam.resource, accessCheckParam.name)),
				Action:   accessCheckParam.action,
			},
			webhook.AthenzAccessCheck{
				Resource: m.res.TrimResource(fmt.Sprintf("%s:%s.%s.%s", accessCheckParam.adminDomain, accessCheckParam.group, accessCheckParam.resource, accessCheckParam.name)),
				Action:   accessCheckParam.action,
			})
	}
	return accessChecks
}

func (m *resourceMapper) createAccessCheck(accessCheckParam athenzAccessCheckParam) []webhook.AthenzAccessCheck {
	accessChecks := make([]webhook.AthenzAccessCheck, 0, len(accessCheckParam.domains))
	for _, domain := range accessCheckParam.domains {
		accessChecks = append(accessChecks,
			webhook.AthenzAccessCheck{
				Resource: m.res.TrimResource(fmt.Sprintf("%s:%s.%s.%s", domain, accessCheckParam.group, accessCheckParam.resource, accessCheckParam.name)),
				Action:   accessCheckParam.action,
			})
	}
	return accessChecks
}
