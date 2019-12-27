# Garm Functional Overview

![flowchart](./assets/garm-functional-flowchart.png)

<!-- MarkdownTOC levels="1,2" -->

- [Garm Functional Overview](#garm-functional-overview)
	- [Parse k8s resources](#parse-k8s-resources)
	- [Map resources](#map-resources)
	- [Subsitute Athenz domain & principal](#subsitute-athenz-domain--principal)
	- [Filter k8s request](#filter-k8s-request)
	- [Select Athenz domain](#select-athenz-domain)
	- [Create Athenz assertion](#create-athenz-assertion)
					- [P.S. during mapping](#ps-during-mapping)

<!-- /MarkdownTOC -->

<a id="parse-k8s-resources"></a>
## Parse k8s resources

![parse k8s resources](./assets/parse-k8s-resources.png)

1. [K8s authorization attributes](https://kubernetes.io/docs/reference/access-authn-authz/webhook/)
	1. ResourceAttributes
		- [ResourceAttributes webhook.go](https://github.com/kubernetes/apiserver/blob/master/plugin/pkg/authorizer/webhook/webhook.go#L160-L168)
		- [ResourceAttributes struct](https://github.com/stefanprodan/kubectl-kubesec/blob/master/vendor/k8s.io/api/authorization/v1beta1/types.go#L86-L112)
	1. NonResourceAttributes
		- [NonResourceAttributes webhook.go](https://github.com/kubernetes/apiserver/blob/master/plugin/pkg/authorizer/webhook/webhook.go#L170-L173)
		- [NonResourceAttributes struct](https://github.com/stefanprodan/kubectl-kubesec/blob/master/vendor/k8s.io/api/authorization/v1beta1/types.go#L114-L122)
1. garm resource attributes
	1. `var namespace, verb, group, resource, name string`

<a id="map-resources"></a>
## Map resources

- verb
	- `config.yaml`, `map_rule.tld.platform.verb_mappings`
	- key-value mapping
- resource
	- `config.yaml`, `map_rule.tld.platform.resource_mappings`
	- key-value mapping
- group
	- is `""` if `map_rule.tld.platform.api_group_control == false`
	- `config.yaml`, `map_rule.tld.platform.api_group_mappings`
	- key-value mapping
- name
	- is `""` if `map_rule.tld.platform.resource_name_control == false`
	- `config.yaml`, `map_rule.tld.platform.resource_name_mappings`
	- key-value mapping

<a id="subsitute-athenz-domain--principal"></a>
## Subsitute Athenz domain & principal

- Map env. variable in Athenz service domain
	- expectation
		1. split by `.`
		1. for each token matches `_.*_`, subsitute with env. variable (except `_namespace_`)
	- example
		- `_k8s_cluster_._namespace_.athenz.service.domain` => `SANDBOX._namespace_.athenz.service.domain`
			+ `config.GetActualValue("k8s_cluster") == "SANDBOX"`
- Map namespace in Athenz admain (both admin & service domain)
	- expectation
		1. subsitute `_namespace_` string in `map_rule.tld.platform.admin_athenz_domain` with garm resource attributes `namespace`
	- example
		- `athenz.domain._namespace_` => `athenz.domain.kaas_namespace`
			+ `namespace = kaas_namespace`
- Map k8s user to Athenz principal
	- expectation
		1. remove `service_account_prefixes`
		1. subsitute namespace
		1. subsitute `:`
		1. if service account, prepend `athenz_service_account_prefix`
		1. if not service account, prepend `athenz_user_prefix`
	- example
		- `service_a:_namespace_:k8s_user` => `domain_a.k8s.kaas_namespace.k8s_user`
			+ `service_account_prefixes = []string{"service_a"}`
			+ `athenz_service_account_prefix = "domain_a.k8s."`
			+ `namespace = kaas_namespace`
		- `service_b:service_c:k8s_user` => `domain_b.serviceaccount.service_c.k8s_user`
			+ `service_account_prefixes = []string{"service_b", "service_c"}`
			+ `athenz_service_account_prefix = "domain_b.k8s."`
		- `service_b:k8s_user` => `domain_c.k8s.k8s_user`
			+ `service_account_prefixes = []string{"service_a", "service_b"}`
			+ `athenz_service_account_prefix = "domain_c.k8s."`
		- `k8s_user` => `user.k8s_user`
			+ `athenz_user_prefix = "user."`

P.S. It may be easier to read the code directly. [createAthenzDomains()](../service/resolver.go#L110), [GetAdminDomain()](../service/resolver.go#280), [BuildDomainsFromNamespace()](../service/resolver.go#125), [PrincipalFromUser()](../service/resolver.go#L187)

<a id="filter-k8s-request"></a>
## Filter k8s request

- `in black_list AND NOT in white_list` => directly reject
	- `config.yaml`, `map_rule.tld.platform.black_list` & `map_rule.tld.platform.white_list`
- Matching logic
	-  create rule RegExp for matching
		- ![garm resource matching](./assets/garm-resource-matching.png)
	- Garm resource attribute is serialized before matching with the rule RegExp.
- Example
	- `RequestInfo{ Verb: "get", Namespace: "kube-system", APIGroup: "*", Resource: "secrets", Name: "alertmanager"}` => check with Athenz
		- black_list contains `RequestInfo{ Verb: "*", Namespace: "kube-system", APIGroup: "*", Resource: "*", Name: "*"}`.
		- white_list contains `RequestInfo{ Verb: "get", Namespace: "kube-system", APIGroup: "*", Resource: "secrets", Name: "alertmanager"}`.
	- `RequestInfo{ Verb: "get", Namespace: "kube-system", APIGroup: "*", Resource: "secrets", Name: "my-secret"}` => directly reject
		- black_list contains `RequestInfo{ Verb: "*", Namespace: "kube-system", APIGroup: "*", Resource: "*", Name: "*"}`.
		- white_list **ONLY** contains `RequestInfo{ Verb: "get", Namespace: "kube-system", APIGroup: "*", Resource: "secrets", Name: "alertmanager"}`.

<a id="select-athenz-domain"></a>
## Select Athenz domain
- `in admin_access_list` => use admin domain
	- `config.yaml`, `map_rule.tld.platform.admin_access_list`
- Matching logic
	- same as above

<a id="create-athenz-assertion"></a>
## Create Athenz assertion

###### P.S. during mapping
![optional api group and resource name](./assets/optional-api-group-and-resource-name.png)

- Athenz service domain
	- ![create athenz assertion on service domain](./assets/create-athenz-assertion-on-service-domain.png)
- Athenz admin domain (2 requests to Athenz, OR logic, any one is allowed implies the action is allowed.)
	- ![create athenz assertion on admin domain](./assets/create-athenz-assertion-on-admin-domain.png)

