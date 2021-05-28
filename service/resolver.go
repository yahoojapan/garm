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
	"strings"

	"github.com/yahoojapan/garm/config"
)

// Resolver is used to map K8s webhook requests to Athenz requests. (Athenz cannot use ":", hence, needs mapping.)
type Resolver interface {
	// MapVerbAction maps K8s verb to Athenz action.
	MapVerbAction(string) string
	// MapK8sResourceAthenzResource maps K8s resources to resources in Athenz resource.
	MapK8sResourceAthenzResource(string) string
	// BuildDomainsFromNamespace creates Athenz domains with namespace.
	BuildDomainsFromNamespace(string) []string
	// PrincipalFromUser creates principal name from user.
	PrincipalFromUser(user string, groups []string) string
	// GetAdminDomain creates Athenz admin domain with namespace.
	GetAdminDomain(string) string
	// MapAPIGroup maps K8s API group to API group in Athenz resource.
	MapAPIGroup(group string) string
	// MapResourceName maps K8s resources name to resources name in Athenz resource.

	MapResourceName(name string) string
	// GetEmptyNamespace returns the mapped value for empty K8s namespace.
	GetEmptyNamespace() string
	// GetNonResourceGroup returns the mapped value for K8s non-resource API group.
	GetNonResourceGroup() string
	// GetNonResourceNamespace returns the mapped value for K8s non-resource namespace.
	GetNonResourceNamespace() string
	// TrimResource sterilizes resources to match Athenz resource naming convention.
	TrimResource(string) string
	// IsAllowed returns true if the K8s request should to Athenz, else returns false if directly reject.
	IsAllowed(verb, namespace, apiGroup, resource, name string) bool
	// IsAdminAccess returns true if the K8s request should use Athenz admin domain.
	IsAdminAccess(verb, namespace, apiGroup, resource, name string) bool
}

// resolve implements Resolver. It contains the configuration information for a K8s platform.
type resolve struct {
	// cfg specifies the mapping rules and platform specific information.
	cfg config.Platform
	// athenzDomains specifies the Athenz domains for request to Athenz.
	athenzDomains []string
	// athenzSAPrefix specifies the Athenz domain for service account.
	athenzSAPrefix string
}

// K8SResolve implementation for K8S platform.
type K8SResolve struct {
	resolve
}

// AKSResolve implementation for Azure AKS platform.
type AKSResolve struct {
	resolve
}

// EKSResolve implementation for Amazon EKS platform.
type EKSResolve struct {
	resolve
}

// NewResolver returns a new resolver using cfg.TLD.Platform.
// The actual return type depends on cfg.TLD.Platform.Name.
// k8s => K8SResolve
// aks => AKSResolve
// eks => EKSResolve
func NewResolver(cfg config.Mapping) Resolver {
	pfConfig := cfg.TLD.Platform
	res := resolve{
		cfg: pfConfig,
	}

	res.athenzDomains = res.createAthenzDomains(pfConfig.ServiceAthenzDomains)
	res.athenzSAPrefix = res.createAthenzDomains([]string{pfConfig.AthenzServiceAccountPrefix})[0]

	switch res.cfg.Name {
	case "k8s":
		return &K8SResolve{res}
	case "aks":
		return &AKSResolve{res}
	case "eks":
		return &EKSResolve{res}
	}
	return &res
}

// MapVerbAction returns mapped value in cfg.VerbMappings,
// else returns the same value.
func (r *resolve) MapVerbAction(verb string) string {
	action, ok := r.cfg.VerbMappings[verb]
	if !ok {
		return verb
	}
	return action
}

// MapK8sResourceAthenzResource returns mapped value in cfg.ResourceMappings,
// else returns the same value.
func (r *resolve) MapK8sResourceAthenzResource(k8sRes string) string {
	athenzRes, ok := r.cfg.ResourceMappings[k8sRes]
	if !ok {
		return k8sRes
	}
	return athenzRes
}

// createAthenzDomains use athenzDomains;
// do the following for each athenzDomains
// split it with ".";
// for each token, if it match /^_.*_$/ but not "_namespace_", replace the token with config.GetActualValue(token);
// and then return the processed value
func (r *resolve) createAthenzDomains(athenzDomains []string) []string {
	if len(athenzDomains) == 0 {
		return athenzDomains
	}
	domains := make([]string, 0, len(athenzDomains))
	// reps stores information for replacement.
	// cap uses the average length of athenzDomains separated by dots
	reps := make([]string, 0, (strings.Count(strings.Join(athenzDomains, "-"), ".")/len(athenzDomains))+1)
	for _, domain := range athenzDomains {
		reps = reps[:0]
		for _, v := range strings.Split(domain, ".") {
			if v != "_namespace_" && strings.HasPrefix(v, "_") && strings.HasSuffix(v, "_") {
				// Note: If deploying in a different namespace than the kube-public namespace, change it to get information from kube api
				reps = append(reps, v, config.GetActualValue(v))
			}
		}
		domains = append(domains, strings.NewReplacer(reps...).Replace(domain))
	}
	return domains
}

// BuildDomainsFromNamespace returns domains by processing athenzDomains.
// if namespace != "", replace `/ = .`, then `.. => -`, then replace "_namespace_" in athenzDomains with namespace;
// else replace "._namespace_" in athenzDomains with namespace;
// trim ".", then "-", then ":"
func (r *resolve) BuildDomainsFromNamespace(namespace string) []string {
	return r.buildAthenzDomain(r.athenzDomains, namespace)
}

//  BuildServiceAccountPrefixFromNamespace returns domains by processing AthenzServiceAccountPrefix.
// if namespace != "", replace `/ = .`, then `.. => -`, then replace "_namespace_" in AthenzServiceAccountPrefix with namespace;
// else replace "._namespace_" in AthenzServiceAccountPrefix with namespace;
// trim ".", then "-", then ":"
func (r *resolve) BuildServiceAccountPrefixFromNamespace(namespace string) []string {
	return r.buildAthenzDomain([]string{r.athenzSAPrefix}, namespace)
}

// buildAthenzDomain returns builtDomains by processing domains.
// if namespace != "", replace `/ = .`, then `.. => -`, then replace "_namespace_" in domain with namespace;
// else replace "._namespace_" in domain with namespace;
// trim ".", then "-", then ":"
func (r *resolve) buildAthenzDomain(domains []string, namespace string) []string {
	builtDomains := make([]string, len(domains))
	for i, domain := range domains {
		if namespace == "" {
			builtDomains[i] = r.trimToValidAsDomain(strings.Replace(domain, "._namespace_", namespace, -1))
			continue
		}

		builtDomains[i] = r.trimToValidAsDomain(strings.Replace(domain, "_namespace_", r.replacePunctuationInNamespace(namespace), -1))
	}
	return builtDomains
}

func (r *resolve) replacePunctuationInNamespace(namespace string) string {
	return strings.Replace(strings.Replace(namespace, "/", ".", -1), "..", "-", -1)
}

func (r *resolve) trimToValidAsDomain(target string) string {
	return strings.TrimPrefix(strings.TrimSuffix(strings.TrimPrefix(strings.TrimSuffix(strings.TrimPrefix(strings.TrimSuffix(target, "."), "."), "-"), "-"), ":"), ":")
}

// MapAPIGroup returns "" if cfg.APIGroupControlEnabled == false;
// else returns cfg.APIGroupMappings mapped value if found, else return original name;
func (r *resolve) MapAPIGroup(group string) string {
	if !r.cfg.APIGroupControlEnabled {
		return ""
	}

	mgroup, ok := r.cfg.APIGroupMappings[group]
	if !ok {
		return group
	}
	return mgroup
}

// MapResourceName returns "" if cfg.ResourceNameControlEnabled == false;
// else returns cfg.ResourceNameMappings mapped value if found, else return original name;
func (r *resolve) MapResourceName(name string) string {
	if !r.cfg.ResourceNameControlEnabled {
		return ""
	}

	mname, ok := r.cfg.ResourceNameMappings[name]
	if ok {
		name = mname
	}

	for k, v := range r.cfg.ResourceNameReplacer {
		name = strings.Replace(name, k, v, -1)
	}

	return name
}

// GetEmptyNamespace returns cfg.EmptyNamespace
func (r *resolve) GetEmptyNamespace() string {
	return r.cfg.EmptyNamespace
}

// GetNonResourceGroup returns cfg.NonResourceAPIGroup
func (r *resolve) GetNonResourceGroup() string {
	return r.cfg.NonResourceAPIGroup
}

// GetNonResourceNamespace returns cfg.NonResourceNamespace
func (r *resolve) GetNonResourceNamespace() string {
	return r.cfg.NonResourceNamespace
}

// PrincipalFromUser maps K8s user to Athenz principal.
// 1. service account: if has ServiceAccountPrefixes, remove prefix, map to AthenzServiceAccountPrefix
// 1.1. if contains namespace, create domain by namespace and AthenzServiceAccountPrefix
// 1.2. if no namespaces, prepend AthenzServiceAccountPrefix
// 2. athenz user: if has AthenzUserPrefix, OR not contains ".", map to AthenzUserPrefix
// 3. certificate: if not service account and athenz user, no mapping
func (r *resolve) PrincipalFromUser(user string, groups []string) string {

	// functions
	hasSaGroup := func(groups []string) bool {
		for _, g := range groups {
			if g == "system:serviceaccounts" {
				return true
			}
		}
		return false
	}
	getSaPrefix := func(user string) string {
		for _, prefix := range r.cfg.ServiceAccountPrefixes {
			if strings.HasPrefix(user, prefix) {
				return prefix
			}
		}
		return ""
	}

	// service account
	if prefix := getSaPrefix(user); prefix != "" && hasSaGroup(groups) {
		parts := strings.Split(strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(user, prefix), ":"), ":"), ":")
		if len(parts) >= 2 {
			return strings.TrimPrefix(strings.TrimSuffix(strings.Join(
				append(r.BuildServiceAccountPrefixFromNamespace(parts[0]), parts[1:]...), "."), ":"), ":")
		}
		return r.cfg.AthenzServiceAccountPrefix + strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(user, prefix), ":"), ":")
	}

	// athenz user
	if !strings.Contains(user, ".") {
		return r.cfg.AthenzUserPrefix + user
	}

	// no mapping
	return user
}

// TrimResource processes res by
// 1. `/ => .`
// 2. `.. => -`
// 3. `-:, :-, .:, :. => :`
// 4. trim ".", then trim "-", then trim ":"
// example:
// TrimResource(fmt.Sprintf("%s:%s.%s.%s", "domain", "group", "resource", "name")) => "domain:group.resource.name"
// TrimResource(fmt.Sprintf("%s:%s.%s.%s", ".-:domain", "group", "resource", "name:-.")) => "domain:group.resource.name"
// TrimResource(fmt.Sprintf("%s:%s.%s.%s", "do/ma//in", "gr..oup", "re-source", "n-:a:-m.:e:.")) => "do.ma-in:gr-oup.re-source.n:a:m:e"
func (r *resolve) TrimResource(res string) string {
	return strings.TrimPrefix(strings.TrimSuffix(strings.TrimPrefix(strings.TrimSuffix(strings.TrimPrefix(strings.TrimSuffix(
		strings.Replace(strings.Replace(strings.Replace(strings.Replace(strings.Replace(strings.Replace(res,
			"/", ".", -1),
			"..", "-", -1),
			"-:", ":", -1),
			":-", ":", -1),
			".:", ":", -1),
			":.", ":", -1),
		"."), "."), "-"), "-"), ":"), ":")
}

// IsAllowed returns true, if inside whitelist or not in both list
// returns false, only if inside blacklist
// i.e. return (in whitelist || not in blacklist)
func (r *resolve) IsAllowed(verb, namespace, apiGroup, resource, name string) bool {
	var ok bool
	for _, white := range r.cfg.WhiteList {
		ok = white.Match(config.RequestInfo{
			Verb:      verb,
			Namespace: namespace,
			APIGroup:  apiGroup,
			Resource:  resource,
			Name:      name,
		})
		if ok {
			return true
		}
	}

	for _, black := range r.cfg.BlackList {
		ok = black.Match(config.RequestInfo{
			Verb:      verb,
			Namespace: namespace,
			APIGroup:  apiGroup,
			Resource:  resource,
			Name:      name,
		})
		if ok {
			return false
		}
	}

	return true
}

// IsAdminAccess returns true, if any admin access in config match
func (r *resolve) IsAdminAccess(verb, namespace, apiGroup, resource, name string) bool {
	var ok bool
	for _, admin := range r.cfg.AdminAccessList {
		ok = admin.Match(config.RequestInfo{
			Verb:      verb,
			Namespace: namespace,
			APIGroup:  apiGroup,
			Resource:  resource,
			Name:      name,
		})
		if ok {
			return true
		}
	}
	return false
}

// GetAdminDomain process cfg.AdminAthenzDomain by
// 1. replace `/ => .`, and then `.. => -` in namespace
// 2. replace "_namespace_" to replaced namespace in cfg.AdminAthenzDomain
// 3. trim ".", then trim "-", then trim ":"
func (r *resolve) GetAdminDomain(namespace string) string {
	return strings.TrimPrefix(strings.TrimSuffix(strings.TrimPrefix(strings.TrimSuffix(strings.TrimPrefix(strings.TrimSuffix(
		strings.Replace(r.cfg.AdminAthenzDomain, "_namespace_", strings.Replace(strings.Replace(namespace,
			"/", ".", -1),
			"..", "-", -1),
			-1),
		"."), "."), "-"), "-"), ":"), ":")
}
