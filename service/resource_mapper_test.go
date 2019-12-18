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
	"reflect"
	"testing"

	webhook "github.com/yahoo/k8s-athenz-webhook"
	"github.com/yahoojapan/garm/config"
	authz "k8s.io/api/authorization/v1beta1"
)

func TestNewResourceMapper(t *testing.T) {
	type args struct {
		resolver Resolver
	}
	type testcase struct {
		name string
		args args
		want ResourceMapper
	}
	tests := []testcase{
		{
			name: "Check NewResourceMapper, nil resolver",
			args: args{
				resolver: nil,
			},
			want: &resourceMapper{},
		},
		func() testcase {
			resolver := &resolve{
				athenzDomains: []string{
					"athenzDomain-24",
					"athenzDomain-25",
				},
			}
			return testcase{
				name: "Check NewResourceMapper",
				args: args{
					resolver: resolver,
				},
				want: &resourceMapper{
					res: resolver,
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewResourceMapper(tt.args.resolver); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewResourceMapper() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resourceMapper_MapResource(t *testing.T) {
	type fields struct {
		res Resolver
	}
	type args struct {
		ctx  context.Context
		spec authz.SubjectAccessReviewSpec
	}
	tests := []struct {
		name                   string
		fields                 fields
		args                   args
		wantIdentity           string
		wantAthenzAccessChecks []webhook.AthenzAccessCheck
		wantError              error
	}{
		{
			name: "Check resourceMapper MapResource, all empty",
			fields: fields{
				res: &resolve{
					athenzDomains: []string{""},
				},
			},
			args: args{
				spec: authz.SubjectAccessReviewSpec{
					ResourceAttributes: &authz.ResourceAttributes{
						Name:        "",
						Namespace:   "",
						Verb:        "",
						Resource:    "",
						Subresource: "",
						Group:       "",
					},
					NonResourceAttributes: &authz.NonResourceAttributes{
						Path: "",
						Verb: "",
					},
					User: "",
				},
			},
			wantIdentity: "",
			wantAthenzAccessChecks: []webhook.AthenzAccessCheck{
				{
					Resource: "",
					Action:   "",
				},
			},
			wantError: nil,
		},
		{
			name: "Check resourceMapper MapResource, nil ResourceAttributes, use non-resources attributes",
			fields: fields{
				res: &resolve{
					athenzDomains: []string{"athenz-domain-106"},
					cfg: config.Platform{
						NonResourceAPIGroup:  "non-resource-api-group-108",
						NonResourceNamespace: "non-resource-namespace-109",
					},
				},
			},
			args: args{
				spec: authz.SubjectAccessReviewSpec{
					ResourceAttributes: nil,
					NonResourceAttributes: &authz.NonResourceAttributes{
						Path: "path-117",
						Verb: "verb-118",
					},
					User: "user-120",
				},
			},
			wantIdentity: "user-120",
			wantAthenzAccessChecks: []webhook.AthenzAccessCheck{
				{
					Resource: "athenz-domain-106:path-117",
					Action:   "verb-118",
				},
			},
			wantError: nil,
		},
		{
			name: "Check resourceMapper MapResource, nil ResourceAttributes, use non-resources attributes, multi domain",
			fields: fields{
				res: &resolve{
					athenzDomains: []string{
						"athenz-domain-158",
						"athenz-domain-159",
					},
					cfg: config.Platform{
						NonResourceAPIGroup:  "non-resource-api-group-162",
						NonResourceNamespace: "non-resource-namespace-163",
					},
				},
			},
			args: args{
				spec: authz.SubjectAccessReviewSpec{
					ResourceAttributes: nil,
					NonResourceAttributes: &authz.NonResourceAttributes{
						Path: "path-171",
						Verb: "verb-172",
					},
					User: "user-174",
				},
			},
			wantIdentity: "user-174",
			wantAthenzAccessChecks: []webhook.AthenzAccessCheck{
				{
					Resource: "athenz-domain-158:path-171",
					Action:   "verb-172",
				},
				{
					Resource: "athenz-domain-159:path-171",
					Action:   "verb-172",
				},
			},
			wantError: nil,
		},
		{
			name: "Check resourceMapper MapResource, ResourceAttributes with empty namespace & non-empty sub-resource",
			fields: fields{
				res: &resolve{
					athenzDomains: []string{"athenz-domain-138._namespace_"},
					cfg: config.Platform{
						EmptyNamespace:             "empty-namespace-140",
						APIGroupControlEnabled:     true,
						ResourceNameControlEnabled: true,
					},
				},
			},
			args: args{
				spec: authz.SubjectAccessReviewSpec{
					ResourceAttributes: &authz.ResourceAttributes{
						Name:        "name-147",
						Namespace:   "",
						Verb:        "verb-149",
						Resource:    "resource-150",
						Subresource: "sub-resource-151",
						Group:       "group-152",
					},
					User: "user-154",
				},
			},
			wantIdentity: "user-154",
			wantAthenzAccessChecks: []webhook.AthenzAccessCheck{
				{
					Resource: "athenz-domain-138.empty-namespace-140:group-152.resource-150.sub-resource-151.name-147",
					Action:   "verb-149",
				},
			},
			wantError: nil,
		},
		{
			name: "Check resourceMapper MapResource, mapping verb & resource & group & name OK",
			fields: fields{
				res: &resolve{
					athenzDomains: []string{"athenz-domain-171"},
					cfg: config.Platform{
						APIGroupControlEnabled:     true,
						ResourceNameControlEnabled: true,
						VerbMappings: map[string]string{
							"verb-176": "athenz-action-176",
						},
						ResourceMappings: map[string]string{
							"k8s-resource-179.sub-resource-179": "athenz-resource-179",
						},
						APIGroupMappings: map[string]string{
							"api-group-182": "mapped-group-182",
						},
						ResourceNameMappings: map[string]string{
							"resource-name-185": "mapped-resource-name-185",
						},
					},
				},
			},
			args: args{
				spec: authz.SubjectAccessReviewSpec{
					ResourceAttributes: &authz.ResourceAttributes{
						Name:        "resource-name-185",
						Verb:        "verb-176",
						Resource:    "k8s-resource-179",
						Group:       "api-group-182",
						Namespace:   "namespace-197",
						Subresource: "sub-resource-179",
					},
					User: "user-200",
				},
			},
			wantIdentity: "user-200",
			wantAthenzAccessChecks: []webhook.AthenzAccessCheck{
				{
					Resource: "athenz-domain-171:mapped-group-182.athenz-resource-179.mapped-resource-name-185",
					Action:   "athenz-action-176",
				},
			},
			wantError: nil,
		},
		{
			name: "Check resourceMapper MapResource, domain from namespace & principal from user",
			fields: fields{
				res: &resolve{
					athenzDomains: []string{"athenz-domain-216._namespace_"},
					cfg: config.Platform{
						APIGroupControlEnabled:     true,
						ResourceNameControlEnabled: true,
						ServiceAccountPrefixes:     []string{"user."},
						AthenzServiceAccountPrefix: "athenz-domain-278._namespace_.",
					},
				},
			},
			args: args{
				spec: authz.SubjectAccessReviewSpec{
					ResourceAttributes: &authz.ResourceAttributes{
						Name:        "name-227",
						Namespace:   "namespace-228",
						Verb:        "verb-229",
						Resource:    "resource-230",
						Subresource: "sub-resource-231",
						Group:       "group-232",
					},
					User: "user.namespace-234:sub-domain-234:user-234",
				},
			},
			wantIdentity: "athenz-domain-278.namespace-234.sub-domain-234.user-234",
			wantAthenzAccessChecks: []webhook.AthenzAccessCheck{
				{
					Resource: "athenz-domain-216.namespace-228:group-232.resource-230.sub-resource-231.name-227",
					Action:   "verb-229",
				},
			},
			wantError: nil,
		},
		{
			name: "Check resourceMapper MapResource, admin access",
			fields: fields{
				res: &resolve{
					athenzDomains: []string{"athenz-admin-domain-250"},
					cfg: config.Platform{
						APIGroupControlEnabled:     true,
						ResourceNameControlEnabled: true,
						AdminAccessList: []*config.RequestInfo{
							{
								Verb:      "verb-*",
								Namespace: "namespace-*",
								APIGroup:  "group-*",
								Resource:  "resource-*",
								Name:      "name-*",
							},
						},
					},
				},
			},
			args: args{
				spec: authz.SubjectAccessReviewSpec{
					ResourceAttributes: &authz.ResourceAttributes{
						Name:        "name-269",
						Namespace:   "namespace-270",
						Verb:        "verb-271",
						Resource:    "resource-272",
						Subresource: "sub-resource-273",
						Group:       "group-274",
					},
					User: "user-276",
				},
			},
			wantIdentity: "user-276",
			wantAthenzAccessChecks: []webhook.AthenzAccessCheck{
				{
					Resource: "group-274.athenz-admin-domain-250.resource-272.sub-resource-273.name-269",
					Action:   "verb-271",
				},
				{
					Resource: "group-274.resource-272.sub-resource-273.name-269",
					Action:   "verb-271",
				},
			},
			wantError: nil,
		},
		{
			name: "Check resourceMapper MapResource, not allowed, directly reject",
			fields: fields{
				res: &resolve{
					athenzDomains: []string{"athenz-domain-296"},
					cfg: config.Platform{
						APIGroupControlEnabled:     true,
						ResourceNameControlEnabled: true,
						BlackList: []*config.RequestInfo{
							{
								Verb:      "verb-*",
								Namespace: "namespace-*",
								APIGroup:  "group-*",
								Resource:  "resource-*",
								Name:      "name-*",
							},
						},
					},
				},
			},
			args: args{
				spec: authz.SubjectAccessReviewSpec{
					ResourceAttributes: &authz.ResourceAttributes{
						Name:        "name-315",
						Namespace:   "namespace-316",
						Verb:        "verb-317",
						Resource:    "resource-318",
						Subresource: "sub-resource-319",
						Group:       "group-320",
					},
					User: "user-322",
				},
			},
			wantIdentity:           "",
			wantAthenzAccessChecks: nil,
			wantError: fmt.Errorf(
				"----user-322's request is not allowed----\nVerb:\tverb-317\nNamespaceb:\tnamespace-316\nAPI Group:\tgroup-320\nResource:\tresource-318.sub-resource-319\nResource Name:\tname-315\n"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &resourceMapper{
				res: tt.fields.res,
			}
			gotIdentity, gotAthenzAccessChecks, gotError := m.MapResource(tt.args.ctx, tt.args.spec)

			if !reflect.DeepEqual(gotError, tt.wantError) {
				t.Errorf("resourceMapper.MapResource() error = %v, wantErr %v", gotError, tt.wantError)
				return
			}
			if gotIdentity != tt.wantIdentity {
				t.Errorf("resourceMapper.MapResource() gotIdentity = %v, want %v", gotIdentity, tt.wantIdentity)
				return
			}
			if !reflect.DeepEqual(gotAthenzAccessChecks, tt.wantAthenzAccessChecks) {
				t.Errorf("resourceMapper.MapResource() gotAthenzAccessChecks = %v, want %v", gotAthenzAccessChecks, tt.wantAthenzAccessChecks)
				return
			}
		})
	}
}
