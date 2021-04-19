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
	"os"
	"reflect"
	"testing"

	"github.com/yahoojapan/garm/config"
)

func TestNewResolver(t *testing.T) {
	type args struct {
		cfg config.Mapping
	}
	tests := []struct {
		name string
		args args
		want Resolver
	}{
		{
			name: "Check NewResolver, invalid platform name",
			args: args{
				cfg: config.Mapping{
					TLD: config.TLD{
						Platform: config.Platform{
							Name: "invalid",
						},
					},
				},
			},
			want: &resolve{},
		},
		{
			name: "Check NewResolver, platform = k8s",
			args: args{
				cfg: config.Mapping{
					TLD: config.TLD{
						Platform: config.Platform{
							Name: "k8s",
						},
					},
				},
			},
			want: &K8SResolve{
				resolve{},
			},
		},
		{
			name: "Check NewResolver, platform = aks",
			args: args{
				cfg: config.Mapping{
					TLD: config.TLD{
						Platform: config.Platform{
							Name: "aks",
						},
					},
				},
			},
			want: &AKSResolve{
				resolve{},
			},
		},
		{
			name: "Check NewResolver, platform = eks",
			args: args{
				cfg: config.Mapping{
					TLD: config.TLD{
						Platform: config.Platform{
							Name: "eks",
						},
					},
				},
			},
			want: &EKSResolve{
				resolve{},
			},
		},
		{
			name: "Check NewResolver, ",
			args: args{
				cfg: config.Mapping{
					TLD: config.TLD{
						Platform: config.Platform{
							ServiceAthenzDomains: []string{
								"test-domain1",
								"test-domain2",
							},
							AthenzServiceAccountPrefix: "test-prefix1.",
						},
					},
				},
			},
			want: &resolve{
				athenzDomains: []string{
					"test-domain1",
					"test-domain2",
				},
				athenzSAPrefix: "test-prefix1.",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewResolver(tt.args.cfg)

			switch v := got.(type) {
			case *K8SResolve:
				if !reflect.DeepEqual(v.resolve.cfg, tt.args.cfg.TLD.Platform) {
					t.Errorf("NewResolver() = %v, want %v", v.resolve.cfg, tt.args.cfg.TLD.Platform)
					return
				}
			case *AKSResolve:
				if !reflect.DeepEqual(v.resolve.cfg, tt.args.cfg.TLD.Platform) {
					t.Errorf("NewResolver() = %v, want %v", v.resolve.cfg, tt.args.cfg.TLD.Platform)
					return
				}
			case *EKSResolve:
				if !reflect.DeepEqual(v.resolve.cfg, tt.args.cfg.TLD.Platform) {
					t.Errorf("NewResolver() = %v, want %v", v.resolve.cfg, tt.args.cfg.TLD.Platform)
					return
				}
			}

			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("NewResolver() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resolve_MapVerbAction(t *testing.T) {
	type fields struct {
		cfg           config.Platform
		athenzDomains []string
	}
	type args struct {
		verb string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "Check resolve MapVerbAction, empty VerbMappings",
			fields: fields{
				cfg: config.Platform{},
			},
			args: args{
				verb: "verb-50",
			},
			want: "verb-50",
		},
		{
			name: "Check resolve MapVerbAction, VerbMappings no matches",
			fields: fields{
				cfg: config.Platform{
					VerbMappings: map[string]string{
						"verb-59": "mapped-verb-59",
						"verb-60": "mapped-verb-60",
					},
				},
			},
			args: args{
				verb: "verb-65",
			},
			want: "verb-65",
		},
		{
			name: "Check resolve MapVerbAction, VerbMappings matche",
			fields: fields{
				cfg: config.Platform{
					VerbMappings: map[string]string{
						"verb-74": "mapped-verb-74",
						"verb-80": "mapped-verb-75",
					},
				},
			},
			args: args{
				verb: "verb-80",
			},
			want: "mapped-verb-75",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resolve{
				cfg:           tt.fields.cfg,
				athenzDomains: tt.fields.athenzDomains,
			}
			if got := r.MapVerbAction(tt.args.verb); got != tt.want {
				t.Errorf("resolve.MapVerbAction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resolve_MapK8sResourceAthenzResource(t *testing.T) {
	type fields struct {
		cfg           config.Platform
		athenzDomains []string
	}
	type args struct {
		k8sRes string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "Check resolve MapK8sResourceAthenzResource, empty ResourceMappings",
			fields: fields{
				cfg: config.Platform{},
			},
			args: args{
				k8sRes: "k8sRes-79",
			},
			want: "k8sRes-79",
		},
		{
			name: "Check resolve MapK8sResourceAthenzResource, ResourceMappings no matches",
			fields: fields{
				cfg: config.Platform{
					ResourceMappings: map[string]string{
						"k8sRes-88": "athenzRes-88",
						"k8sRes-89": "athenzRes-89",
					},
				},
			},
			args: args{
				k8sRes: "k8sRes-91",
			},
			want: "k8sRes-91",
		},
		{
			name: "Check resolve MapK8sResourceAthenzResource, ResourceMappings matche",
			fields: fields{
				cfg: config.Platform{
					ResourceMappings: map[string]string{
						"k8sRes-103": "athenzRes-103",
						"k8sRes-109": "athenzRes-104",
					},
				},
			},
			args: args{
				k8sRes: "k8sRes-109",
			},
			want: "athenzRes-104",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resolve{
				cfg:           tt.fields.cfg,
				athenzDomains: tt.fields.athenzDomains,
			}
			if got := r.MapK8sResourceAthenzResource(tt.args.k8sRes); got != tt.want {
				t.Errorf("resolve.MapK8sResourceAthenzResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resolve_createAthenzDomains(t *testing.T) {
	type fields struct {
		cfg            config.Platform
		athenzDomains  []string
		athenzSAPrefix string
	}
	type args struct {
		athenzDomains []string
	}
	type testcase struct {
		name       string
		fields     fields
		args       args
		want       []string
		beforeFunc func() error
		afterFunc  func() error
	}
	tests := []testcase{
		{
			name: "Check resolve createAthenzDomains, empty serviceAthenzDomains",
			args: args{
				athenzDomains: []string{},
			},
			want: []string{},
		},
		{
			name: "Check resolve createAthenzDomains, serviceAthenzDomains with empty string",
			args: args{
				athenzDomains: []string{""},
			},
			want: []string{""},
		},
		{
			name: "Check resolve createAthenzDomains, serviceAthenzDomains no split, no replace",
			args: args{
				athenzDomains: []string{"service-athenz-domain-192"},
			},
			want: []string{"service-athenz-domain-192"},
		},
		{
			name: "Check resolve createAthenzDomains, multi serviceAthenzDomains",
			args: args{
				athenzDomains: []string{
					"service-athenz-domain-296",
					"service-athenz-domain-297._namespace_",
				},
			},
			want: []string{
				"service-athenz-domain-296",
				"service-athenz-domain-297._namespace_",
			},
		},
		{
			name: "Check resolve createAthenzDomains, multi serviceAthenzDomains end with dot",
			args: args{
				athenzDomains: []string{
					"athenz-sa-prefix-296.",
					"athenz-sa-prefix-297._namespace_.",
				},
			},
			want: []string{
				"athenz-sa-prefix-296.",
				"athenz-sa-prefix-297._namespace_.",
			},
		},
		func() testcase {
			env := map[string]string{
				"env-199": "evalue-199",
				"env-200": "evalue-200",
			}
			serviceAthenzDomains := []string{"_namespace_._env-200_._env-199_"}

			return testcase{
				name: "Check resolve createAthenzDomains, serviceAthenzDomains, multiple replace, skip _namespace_",
				args: args{
					athenzDomains: serviceAthenzDomains,
				},
				beforeFunc: func() error {
					for k, v := range env {
						err := os.Setenv(k, v)
						if err != nil {
							return err
						}
					}
					return nil
				},
				afterFunc: func() error {
					for k := range env {
						err := os.Unsetenv(k)
						if err != nil {
							return err
						}
					}
					return nil
				},
				want: []string{"_namespace_.evalue-200.evalue-199"},
			}
		}(),
		func() testcase {
			env := map[string]string{
				"env-235": "evalue-235",
				"env-236": "evalue-236",
			}
			serviceAthenzDomains := []string{"_env-236_.env-235."}

			return testcase{
				name: "Check resolve createAthenzDomains, serviceAthenzDomains, single replace",
				args: args{
					athenzDomains: serviceAthenzDomains,
				},
				beforeFunc: func() error {
					for k, v := range env {
						err := os.Setenv(k, v)
						if err != nil {
							return err
						}
					}
					return nil
				},
				afterFunc: func() error {
					for k := range env {
						err := os.Unsetenv(k)
						if err != nil {
							return err
						}
					}
					return nil
				},
				want: []string{"evalue-236.env-235."},
			}
		}(),
		func() testcase {
			env := map[string]string{
				"env-270": "evalue-270",
				"env-271": "evalue-271",
				"env-272": "evalue-272",
			}
			serviceAthenzDomains := []string{".env-270.env-271.env-272"}

			return testcase{
				name: "Check resolve createAthenzDomains, serviceAthenzDomains, split but no replace",
				args: args{
					athenzDomains: serviceAthenzDomains,
				},
				beforeFunc: func() error {
					for k, v := range env {
						err := os.Setenv(k, v)
						if err != nil {
							return err
						}
					}
					return nil
				},
				afterFunc: func() error {
					for k := range env {
						err := os.Unsetenv(k)
						if err != nil {
							return err
						}
					}
					return nil
				},
				want: []string{".env-270.env-271.env-272"},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.afterFunc != nil {
				defer func() {
					err := tt.afterFunc()
					if err != nil {
						t.Error(err)
					}
				}()
			}

			if tt.beforeFunc != nil {
				err := tt.beforeFunc()
				if err != nil {
					t.Error(err)
					return
				}
			}

			r := &resolve{
				cfg:           tt.fields.cfg,
				athenzDomains: tt.fields.athenzDomains,
			}
			if got := r.createAthenzDomains(tt.args.athenzDomains); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("resolve.createAthenzDomains() = %v, want %v", got, tt.want)
				return
			}
		})
	}
}

func Test_resolve_BuildDomainsFromNamespace(t *testing.T) {
	type fields struct {
		cfg           config.Platform
		athenzDomains []string
	}
	type args struct {
		namespace string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		{
			name: "Check resolve BuildDomainsFromNamespace, empty namespace, empty athenzDomains",
			fields: fields{
				athenzDomains: []string{""},
			},
			args: args{
				namespace: "",
			},
			want: []string{""},
		},
		{
			name: "Check resolve BuildDomainsFromNamespace, empty namespace, athenzDomains no replace & trim",
			fields: fields{
				athenzDomains: []string{"athenz-domain-140"},
			},
			args: args{
				namespace: "",
			},
			want: []string{"athenz-domain-140"},
		},
		{
			name: "Check resolve BuildDomainsFromNamespace, empty namespace, multi athenzDomains no replace & no trim ",
			fields: fields{
				athenzDomains: []string{
					"athenz-domain-482",
					"athenz-domain-483",
				},
			},
			args: args{
				namespace: "",
			},
			want: []string{
				"athenz-domain-482",
				"athenz-domain-483",
			},
		},
		{
			name: "Check resolve BuildDomainsFromNamespace, empty namespace, athenzDomains no replace, full trim",
			fields: fields{
				athenzDomains: []string{".-:athenz-domain-150:-."},
			},
			args: args{
				namespace: "",
			},
			want: []string{"athenz-domain-150"},
		},
		{
			name: "Check resolve BuildDomainsFromNamespace, empty namespace, athenzDomains no replace, partially trim",
			fields: fields{
				athenzDomains: []string{":-.athenz-domain-160.:-"},
			},
			args: args{
				namespace: "",
			},
			want: []string{"-.athenz-domain-160."},
		},
		{
			name: "Check resolve BuildDomainsFromNamespace, empty namespace, athenzDomains no trim, replace",
			fields: fields{
				athenzDomains: []string{"athenz-|._namespace_||._namespace_|-domain-170"},
			},
			args: args{
				namespace: "",
			},
			want: []string{"athenz-||||-domain-170"},
		},
		{
			name: "Check resolve BuildDomainsFromNamespace, athenzDomains no trim, no replace namespace",
			fields: fields{
				athenzDomains: []string{"athenz-|.namespace||.namespace|-domain-180"},
			},
			args: args{
				namespace: "namespace-183",
			},
			want: []string{"athenz-|.namespace||.namespace|-domain-180"},
		},
		{
			name: "Check resolve BuildDomainsFromNamespace, athenzDomains no trim, replace namespace",
			fields: fields{
				athenzDomains: []string{"athenz-|._namespace_||._namespace_|-domain-190"},
			},
			args: args{
				namespace: "namespace-193",
			},
			want: []string{"athenz-|.namespace-193||.namespace-193|-domain-190"},
		},
		{
			name: "Check resolve BuildDomainsFromNamespace, namspace replace, athenzDomains no trim, replace namespace",
			fields: fields{
				athenzDomains: []string{"athenz-<._namespace_>-domain-200"},
			},
			args: args{
				namespace: "namespace|//|/./|./.././../.|./n-s/.ns/../nn-ss//sss|-183",
			},
			want: []string{"athenz-<.namespace|-|-.|-----.|-n-s-ns--nn-ss-sss|-183>-domain-200"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resolve{
				cfg:           tt.fields.cfg,
				athenzDomains: tt.fields.athenzDomains,
			}
			if got := r.BuildDomainsFromNamespace(tt.args.namespace); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("resolve.BuildDomainsFromNamespace() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Test_resolve_BuildServiceAccountPrefixFromNamespace(t *testing.T) {
	type fields struct {
		athenzSAPrefix string
	}
	type args struct {
		namespace string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		{
			name: "Check resolve BuildServiceAccountPrefixFromNamespace, empty namespace, empty AthenzServiceAccountPrefix",
			fields: fields{
				athenzSAPrefix: "",
			},
			args: args{
				namespace: "",
			},
			want: []string{""},
		},
		{
			name: "Check resolve BuildServiceAccountPrefixFromNamespace, empty namespace, AthenzServiceAccountPrefix no replace & no trim",
			fields: fields{
				athenzSAPrefix: "athenz-domain-506",
			},
			args: args{
				namespace: "",
			},
			want: []string{"athenz-domain-506"},
		},
		{
			name: "Check resolve BuildServiceAccountPrefixFromNamespace, empty namespace, AthenzServiceAccountPrefix no replace, full trim",
			fields: fields{
				athenzSAPrefix: ".-:athenz-domain-608:-.",
			},
			args: args{
				namespace: "",
			},
			want: []string{"athenz-domain-608"},
		},
		{
			name: "Check resolve BuildServiceAccountPrefixFromNamespace, empty namespace, AthenzServiceAccountPrefix no replace, partially trim",
			fields: fields{
				athenzSAPrefix: ":-.athenz-domain-620.:-",
			},
			args: args{
				namespace: "",
			},
			want: []string{"-.athenz-domain-620."},
		},
		{
			name: "Check resolve BuildServiceAccountPrefixFromNamespace, empty namespace, AthenzServiceAccountPrefix no trim, replace",
			fields: fields{
				athenzSAPrefix: "athenz-|._namespace_||._namespace_|-domain-632",
			},
			args: args{
				namespace: "",
			},
			want: []string{"athenz-||||-domain-632"},
		},
		{
			name: "Check resolve BuildServiceAccountPrefixFromNamespace, AthenzServiceAccountPrefix no trim, no replace namespace",
			fields: fields{
				athenzSAPrefix: "athenz-|.namespace||.namespace|-domain-644",
			},
			args: args{
				namespace: "namespace-648",
			},
			want: []string{"athenz-|.namespace||.namespace|-domain-644"},
		},
		{
			name: "Check resolve BuildServiceAccountPrefixFromNamespace, AthenzServiceAccountPrefix no trim, replace namespace",
			fields: fields{
				athenzSAPrefix: "athenz-|._namespace_||._namespace_|-domain-656",
			},
			args: args{
				namespace: "namespace-660",
			},
			want: []string{"athenz-|.namespace-660||.namespace-660|-domain-656"},
		},
		{
			name: "Check resolve BuildServiceAccountPrefixFromNamespace, namspace replace, AthenzServiceAccountPrefix no trim, replace namespace",
			fields: fields{
				athenzSAPrefix: "athenz-<._namespace_>-domain-668",
			},
			args: args{
				namespace: "namespace|//|/./|./.././../.|./n-s/.ns/../nn-ss//sss|-672",
			},
			want: []string{"athenz-<.namespace|-|-.|-----.|-n-s-ns--nn-ss-sss|-672>-domain-668"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resolve{
				athenzSAPrefix: tt.fields.athenzSAPrefix,
			}
			if got := r.BuildServiceAccountPrefixFromNamespace(tt.args.namespace); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("resolve.BuildServiceAccountPrefixFromNamespace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resolve_MapAPIGroup(t *testing.T) {
	type fields struct {
		cfg           config.Platform
		athenzDomains []string
	}
	type args struct {
		group string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "Check resolve MapAPIGroup, not APIGroupControlEnabled",
			fields: fields{
				cfg: config.Platform{
					APIGroupControlEnabled: false,
				},
			},
			args: args{
				group: "api-group-164",
			},
			want: "",
		},
		{
			name: "Check resolve MapAPIGroup, empty map",
			fields: fields{
				cfg: config.Platform{
					APIGroupControlEnabled: true,
				},
			},
			args: args{
				group: "api-group-177",
			},
			want: "api-group-177",
		},
		{
			name: "Check resolve MapAPIGroup, APIGroupMappings no matches",
			fields: fields{
				cfg: config.Platform{
					APIGroupControlEnabled: true,
					APIGroupMappings: map[string]string{
						"api-group-187": "mapped-name-187",
						"api-group-188": "mapped-name-188",
					},
				},
			},
			args: args{
				group: "api-group-193",
			},
			want: "api-group-193",
		},
		{
			name: "Check resolve MapAPIGroup, APIGroupMappings match",
			fields: fields{
				cfg: config.Platform{
					APIGroupControlEnabled: true,
					APIGroupMappings: map[string]string{
						"api-group-203": "mapped-name-203",
						"api-group-209": "mapped-name-204",
					},
				},
			},
			args: args{
				group: "api-group-209",
			},
			want: "mapped-name-204",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resolve{
				cfg:           tt.fields.cfg,
				athenzDomains: tt.fields.athenzDomains,
			}
			if got := r.MapAPIGroup(tt.args.group); got != tt.want {
				t.Errorf("resolve.MapAPIGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resolve_MapResourceName(t *testing.T) {
	type fields struct {
		cfg           config.Platform
		athenzDomains []string
	}
	type args struct {
		name string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "Check resolve MapResourceName, not ResourceNameControlEnabled",
			fields: fields{
				cfg: config.Platform{
					ResourceNameControlEnabled: false,
				},
			},
			args: args{
				name: "resource-name-192",
			},
			want: "",
		},
		{
			name: "Check resolve MapResourceName, empty map",
			fields: fields{
				cfg: config.Platform{
					ResourceNameControlEnabled: true,
				},
			},
			args: args{
				name: "resource-name-203",
			},
			want: "resource-name-203",
		},
		{
			name: "Check resolve MapResourceName, ResourceNameMappings no matches",
			fields: fields{
				cfg: config.Platform{
					ResourceNameControlEnabled: true,
					ResourceNameMappings: map[string]string{
						"resource-name-212": "mapped-name-212",
						"resource-name-213": "mapped-name-213",
					},
				},
			},
			args: args{
				name: "resource-name-217",
			},
			want: "resource-name-217",
		},
		{
			name: "Check resolve MapResourceName, ResourceNameMappings match",
			fields: fields{
				cfg: config.Platform{
					ResourceNameControlEnabled: true,
					ResourceNameMappings: map[string]string{
						"resource-name-226": "mapped-name-226",
						"resource-name-231": "mapped-name-227",
					},
				},
			},
			args: args{
				name: "resource-name-231",
			},
			want: "mapped-name-227",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resolve{
				cfg:           tt.fields.cfg,
				athenzDomains: tt.fields.athenzDomains,
			}
			if got := r.MapResourceName(tt.args.name); got != tt.want {
				t.Errorf("resolve.MapResourceName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resolve_GetEmptyNamespace(t *testing.T) {
	type fields struct {
		cfg           config.Platform
		athenzDomains []string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Check resolve GetEmptyNamespace",
			fields: fields{
				cfg: config.Platform{
					EmptyNamespace: "empty-namespace-214",
				},
			},
			want: "empty-namespace-214",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resolve{
				cfg:           tt.fields.cfg,
				athenzDomains: tt.fields.athenzDomains,
			}
			if got := r.GetEmptyNamespace(); got != tt.want {
				t.Errorf("resolve.GetEmptyNamespace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resolve_GetNonResourceGroup(t *testing.T) {
	type fields struct {
		cfg           config.Platform
		athenzDomains []string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Check resolve GetNonResourceGroup",
			fields: fields{
				cfg: config.Platform{
					NonResourceAPIGroup: "non-resource-api-group-247",
				},
			},
			want: "non-resource-api-group-247",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resolve{
				cfg:           tt.fields.cfg,
				athenzDomains: tt.fields.athenzDomains,
			}
			if got := r.GetNonResourceGroup(); got != tt.want {
				t.Errorf("resolve.GetNonResourceGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resolve_GetNonResourceNamespace(t *testing.T) {
	type fields struct {
		cfg           config.Platform
		athenzDomains []string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Check resolve GetNonResourceNamespace",
			fields: fields{
				cfg: config.Platform{
					NonResourceNamespace: "non-resource-namespace-264",
				},
			},
			want: "non-resource-namespace-264",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resolve{
				cfg:           tt.fields.cfg,
				athenzDomains: tt.fields.athenzDomains,
			}
			if got := r.GetNonResourceNamespace(); got != tt.want {
				t.Errorf("resolve.GetNonResourceNamespace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resolve_PrincipalFromUser(t *testing.T) {
	type fields struct {
		athenzSAPrefix string
		cfg            config.Platform
		athenzDomains  []string
	}
	type args struct {
		user   string
		groups []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "Check resolve PrincipalFromUser empty ServiceAccountPrefixes",
			fields: fields{
				cfg: config.Platform{
					ServiceAccountPrefixes: []string{},
					AthenzUserPrefix:       "athenz-user-prefix-295",
				},
			},
			args: args{
				user: "user-299",
			},
			want: "athenz-user-prefix-295user-299",
		},
		{
			name: "Check resolve PrincipalFromUser user do not contains any ServiceAccountPrefixes",
			fields: fields{
				cfg: config.Platform{
					ServiceAccountPrefixes: []string{"prefix-308"},
					AthenzUserPrefix:       "athenz-user-prefix-309",
				},
			},
			args: args{
				user: "user-313",
			},
			want: "athenz-user-prefix-309user-313",
		},
		{
			name: "Check resolve PrincipalFromUser ServiceAccountPrefixes match user prefix, single part, no trim",
			fields: fields{
				cfg: config.Platform{
					ServiceAccountPrefixes: []string{"prefix-319:"},
				},
			},
			args: args{
				user:   "prefix-319:user-323",
				groups: []string{"system:serviceaccounts"},
			},
			want: "user-323",
		},
		{
			name: "Check resolve PrincipalFromUser ServiceAccountPrefixes match user prefix, single part, no groups",
			fields: fields{
				cfg: config.Platform{
					ServiceAccountPrefixes: []string{"prefix-319:"},
				},
			},
			args: args{
				user:   "prefix-319:user-323",
				groups: []string{},
			},
			want: "prefix-319:user-323",
		},

		{
			name: "Check resolve PrincipalFromUser ServiceAccountPrefixes match user prefix, single part, need trim",
			fields: fields{
				cfg: config.Platform{
					ServiceAccountPrefixes: []string{"prefix-331"},
				},
			},
			args: args{
				user:   "prefix-331:user-335:",
				groups: []string{"system:serviceaccounts"},
			},
			want: "user-335",
		},
		{
			name: "Check resolve PrincipalFromUser ServiceAccountPrefixes match user prefix, multiple parts, empty namespace",
			fields: fields{
				athenzSAPrefix: "athenz-|._namespace_||._namespace_|-domain-342",
				cfg: config.Platform{
					ServiceAccountPrefixes: []string{"prefix-not-match", "prefix-344"},
				},
			},
			args: args{
				user:   "prefix-344::part-1:user-349:",
				groups: []string{"system:serviceaccounts"},
			},
			want: "athenz-||||-domain-342.part-1.user-349",
		},
		{
			name: "Check resolve PrincipalFromUser ServiceAccountPrefixes match user prefix, multiple parts, non-empty namespace",
			fields: fields{
				athenzSAPrefix: "athenz-|._namespace_||._namespace_|-domain-356",
				cfg: config.Platform{
					ServiceAccountPrefixes: []string{"prefix-not-match", "prefix-358"},
				},
			},
			args: args{
				user:   "prefix-358:ns-361:part-1:user-361:",
				groups: []string{"system:serviceaccounts"},
			},
			want: "athenz-|.ns-361||.ns-361|-domain-356.part-1.user-361",
		},
		{
			name: "Check resolve PrincipalFromUser ServiceAccountPrefixes with empty ServiceAccountPrefixes",
			fields: fields{
				cfg: config.Platform{
					ServiceAccountPrefixes: []string{"prefix-not-match", ""},
				},
			},
			args: args{
				user:   ":user-373:",
				groups: []string{"system:serviceaccounts"},
			},
			want: ":user-373:",
		},
		{
			name: "Check resolve PrincipalFromUser ServiceAccountPrefixes with UserPrefix and empty ServiceAccountPrefixes",
			fields: fields{
				cfg: config.Platform{
					ServiceAccountPrefixes: []string{"prefix-not-match", ""},
					AthenzUserPrefix:       "user-prefix.",
				},
			},
			args: args{
				user:   ":user-373:",
				groups: []string{"system:serviceaccounts"},
			},
			want: "user-prefix.:user-373:",
		},
		{
			name: "Check resolve PrincipalFromUser with a request without service account and athenz user",
			fields: fields{
				cfg: config.Platform{
					ServiceAccountPrefixes: []string{"prefix-not-match", ""},
					AthenzUserPrefix:       "user-prefix.",
				},
			},
			args: args{
				user: "domain.service",
			},
			want: "domain.service",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resolve{
				cfg:            tt.fields.cfg,
				athenzDomains:  tt.fields.athenzDomains,
				athenzSAPrefix: tt.fields.athenzSAPrefix,
			}
			if got := r.PrincipalFromUser(tt.args.user, tt.args.groups); got != tt.want {
				t.Errorf("resolve.PrincipalFromUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resolve_TrimResource(t *testing.T) {
	type fields struct {
		cfg           config.Platform
		athenzDomains []string
	}
	type args struct {
		res string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name:   "Check resolve TrimResource trim unchanged",
			fields: fields{},
			args: args{
				res: "res-322",
			},
			want: "res-322",
		},
		{
			name:   "Check resolve TrimResource trim all",
			fields: fields{},
			args: args{
				res: "..-:res-330:-...",
			},
			want: "res-330",
		},
		{
			name:   "Check resolve TrimResource trim partially",
			fields: fields{},
			args: args{
				res: "..-:.-res-338.//:-",
			},
			want: "-res-338-",
		},
		{
			name:   "Check resolve TrimResource replace",
			fields: fields{},
			args: args{
				res: "|//.|.//|/.-:|:/.-|//.:.//|..://:--:.|ttt/.rrr|../:./:./.abc/|",
			},
			want: "|-.|-.|-:|:-|-:|:::|ttt-rrr|-::abc.|",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resolve{
				cfg:           tt.fields.cfg,
				athenzDomains: tt.fields.athenzDomains,
			}
			if got := r.TrimResource(tt.args.res); got != tt.want {
				t.Errorf("resolve.TrimResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resolve_IsAllowed(t *testing.T) {
	type fields struct {
		cfg           config.Platform
		athenzDomains []string
	}
	type args struct {
		verb      string
		namespace string
		apiGroup  string
		resource  string
		name      string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Check resolve IsAllowed not in both list",
			fields: fields{
				cfg: config.Platform{
					WhiteList: []*config.RequestInfo{},
					BlackList: []*config.RequestInfo{},
				},
			},
			args: args{
				verb:      "verb-360",
				namespace: "namespace-361",
				apiGroup:  "apiGroup-362",
				resource:  "resource-363",
				name:      "name-364",
			},
			want: true,
		},
		{
			name: "Check resolve IsAllowed in black but not in white",
			fields: fields{
				cfg: config.Platform{
					WhiteList: []*config.RequestInfo{},
					BlackList: []*config.RequestInfo{
						{
							Verb:      "verb*",
							Namespace: "namespace*",
							APIGroup:  "apiGroup*",
							Resource:  "resource*",
							Name:      "name*",
						},
					},
				},
			},
			args: args{
				verb:      "verb-385",
				namespace: "namespace-386",
				apiGroup:  "apiGroup-387",
				resource:  "resource-388",
				name:      "name-389",
			},
			want: false,
		},
		{
			name: "Check resolve IsAllowed in white but not in black",
			fields: fields{
				cfg: config.Platform{
					WhiteList: []*config.RequestInfo{
						{
							Verb:      "verb*",
							Namespace: "namespace*",
							APIGroup:  "apiGroup*",
							Resource:  "resource*",
							Name:      "name*",
						}},
					BlackList: []*config.RequestInfo{},
				},
			},
			args: args{
				verb:      "verb-409",
				namespace: "namespace-410",
				apiGroup:  "apiGroup-411",
				resource:  "resource-412",
				name:      "name-413",
			},
			want: true,
		},
		{
			name: "Check resolve IsAllowed in both list",
			fields: fields{
				cfg: config.Platform{
					WhiteList: []*config.RequestInfo{
						{
							Verb:      "verb*",
							Namespace: "namespace*",
							APIGroup:  "apiGroup*",
							Resource:  "resource*",
							Name:      "name*",
						}},
					BlackList: []*config.RequestInfo{
						{
							Verb:      "verb*",
							Namespace: "namespace*",
							APIGroup:  "apiGroup*",
							Resource:  "resource*",
							Name:      "name*",
						}},
				},
			},
			args: args{
				verb:      "verb-440",
				namespace: "namespace-441",
				apiGroup:  "apiGroup-442",
				resource:  "resource-443",
				name:      "name-444",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resolve{
				cfg:           tt.fields.cfg,
				athenzDomains: tt.fields.athenzDomains,
			}
			if got := r.IsAllowed(tt.args.verb, tt.args.namespace, tt.args.apiGroup, tt.args.resource, tt.args.name); got != tt.want {
				t.Errorf("resolve.IsAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resolve_IsAdminAccess(t *testing.T) {
	type fields struct {
		cfg           config.Platform
		athenzDomains []string
	}
	type args struct {
		verb      string
		namespace string
		apiGroup  string
		resource  string
		name      string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Check resolve IsAdminAccess empty list",
			fields: fields{
				cfg: config.Platform{
					AdminAccessList: []*config.RequestInfo{},
				},
			},
			args: args{
				verb:      "verb-392",
				namespace: "namespace-393",
				apiGroup:  "apiGroup-394",
				resource:  "resource-395",
				name:      "name-396",
			},
			want: false,
		},
		{
			name: "Check resolve IsAdminAccess no match",
			fields: fields{
				cfg: config.Platform{
					AdminAccessList: []*config.RequestInfo{
						{
							Verb:      "verb-406",
							Namespace: "namespace-407",
							APIGroup:  "apiGroup-408",
							Resource:  "resource-409",
							Name:      "name-410",
						},
					},
				},
			},
			args: args{
				verb:      "verb-416",
				namespace: "namespace-417",
				apiGroup:  "apiGroup-418",
				resource:  "resource-419",
				name:      "name-420",
			},
			want: false,
		},
		{
			name: "Check resolve IsAdminAccess has exact match",
			fields: fields{
				cfg: config.Platform{
					AdminAccessList: []*config.RequestInfo{
						{
							Verb:      "verb-430",
							Namespace: "namespace-431",
							APIGroup:  "apiGroup-432",
							Resource:  "resource-433",
							Name:      "name-434",
						},
						{
							Verb:      "verb-437",
							Namespace: "namespace-438",
							APIGroup:  "apiGroup-439",
							Resource:  "resource-440",
							Name:      "name-441",
						},
					},
				},
			},
			args: args{
				verb:      "verb-437",
				namespace: "namespace-438",
				apiGroup:  "apiGroup-439",
				resource:  "resource-440",
				name:      "name-441",
			},
			want: true,
		},
		{
			name: "Check resolve IsAdminAccess regex match",
			fields: fields{
				cfg: config.Platform{
					AdminAccessList: []*config.RequestInfo{
						{
							Verb:      "verb-461",
							Namespace: "namespace-462",
							APIGroup:  "apiGroup-463",
							Resource:  "resource-.*",
							Name:      "name-465",
						},
					},
				},
			},
			args: args{
				verb:      "verb-461",
				namespace: "namespace-462",
				apiGroup:  "apiGroup-463",
				resource:  "resource-474",
				name:      "name-465",
			},
			want: true,
		},
		{
			name: "Check resolve IsAdminAccess regex match fail after APIGroup replace",
			fields: fields{
				cfg: config.Platform{
					AdminAccessList: []*config.RequestInfo{
						{
							Verb:      "verb-484",
							Namespace: "namespace-485",
							APIGroup:  "apiGroup-.*",
							Resource:  "resource-488",
							Name:      "name-489",
						},
					},
				},
			},
			args: args{
				verb:      "verb-484",
				namespace: "namespace-485",
				apiGroup:  "apiGroup-497",
				resource:  "resource-488",
				name:      "name-489",
			},
			want: false,
		},
		{
			name: "Check resolve IsAdminAccess regex match success after APIGroup replace",
			fields: fields{
				cfg: config.Platform{
					AdminAccessList: []*config.RequestInfo{
						{
							Verb:      "verb-509",
							Namespace: "namespace-510",
							APIGroup:  "apiGroup-.*",
							Resource:  "resource-512",
							Name:      "name-513",
						},
					},
				},
			},
			args: args{
				verb:      "verb-509",
				namespace: "namespace-510",
				apiGroup:  "apiGroup-_______________",
				resource:  "resource-512",
				name:      "name-513",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resolve{
				cfg:           tt.fields.cfg,
				athenzDomains: tt.fields.athenzDomains,
			}
			if got := r.IsAdminAccess(tt.args.verb, tt.args.namespace, tt.args.apiGroup, tt.args.resource, tt.args.name); got != tt.want {
				t.Errorf("resolve.IsAdminAccess() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resolve_GetAdminDomain(t *testing.T) {
	type fields struct {
		cfg           config.Platform
		athenzDomains []string
	}
	type args struct {
		namespace string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "Check resolve GetAdminDomain trim unchanged",
			fields: fields{
				cfg: config.Platform{
					AdminAthenzDomain: "aaDomain-417",
				},
			},
			args: args{
				namespace: "namespace-421",
			},
			want: "aaDomain-417",
		},
		{
			name: "Check resolve GetAdminDomain trim all",
			fields: fields{
				cfg: config.Platform{
					AdminAthenzDomain: ".-:aaDomain-429:-.",
				},
			},
			args: args{
				namespace: "namespace-433",
			},
			want: "aaDomain-429",
		},
		{
			name: "Check resolve GetAdminDomain trim partially",
			fields: fields{
				cfg: config.Platform{
					AdminAthenzDomain: ":-!.aaDomain-441:-!.",
				},
			},
			args: args{
				namespace: "namespace-445",
			},
			want: "-!.aaDomain-441:-!",
		},
		{
			name: "Check resolve GetAdminDomain replace admin athenz domain by namespace",
			fields: fields{
				cfg: config.Platform{
					AdminAthenzDomain: "admin_athenz_domain_namespace_xxx_namespace_yyy-453",
				},
			},
			args: args{
				namespace: "(namespace-457)",
			},
			want: "admin_athenz_domain(namespace-457)xxx(namespace-457)yyy-453",
		},
		{
			name: "Check resolve GetAdminDomain replace namespace",
			fields: fields{
				cfg: config.Platform{
					AdminAthenzDomain: "admin_athenz_domain-_namespace_-465",
				},
			},
			args: args{
				namespace: "namespace|/|..|//|/.|./|/..|./.|../|469",
			},
			want: "admin_athenz_domain-namespace|.|-|-|-|-|-.|-.|-.|469-465",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &resolve{
				cfg:           tt.fields.cfg,
				athenzDomains: tt.fields.athenzDomains,
			}
			if got := r.GetAdminDomain(tt.args.namespace); got != tt.want {
				t.Errorf("resolve.GetAdminDomain() = %v, want %v", got, tt.want)
			}
		})
	}
}
