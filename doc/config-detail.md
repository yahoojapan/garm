# Configuration Detail

<!-- MarkdownTOC levels="1,2" -->

- [TLS](#tls)
- [Athenz n-token](#athenz-n-token)
- [Request filtering](#request-filtering)
- [Admin domain](#admin-domain)
- [Resource mapping](#resource-mapping)
- [Optional API group and resource name control](#optional-api-group-and-resource-name-control)
- [Mapping for non-resources or empty namespace](#mapping-for-non-resources-or-empty-namespace)
- [P.S.](#ps)

<!-- /MarkdownTOC -->

<a id="tls"></a>
## TLS

![TLS](./assets/tls.png)

<a id="related-configuration"></a>
### Related configuration
1. For garm, `config.yaml`
	```yaml
	server.tls.ca_key
	server.tls.cert_key
	server.tls.key_key

	athenz.root_ca
	```
1. For kube-apiserver, `authz.yaml`
	```yaml
	# https://github.com/kubernetes/apiserver/blob/master/plugin/pkg/authorizer/webhook/webhook.go#L69
	clusters.cluster.certificate-authority

	# https://github.com/kubernetes/apiserver/blob/master/plugin/pkg/authorizer/webhook/webhook.go#L76-L77
	users.user.client-certificate
	users.user.client-key
	```

<a id="note"></a>
#### Note
- Garm uses the same server certificate for /authn and /authz.
- If `server.tls.ca_key` is not set, garm will not verify the client certificate of kube-apiserver.

---

<a id="athenz-n-token"></a>
## Athenz n-token

<a id="related-configuration-1"></a>
### Related configuration
```yaml
athenz.auth_header

athenz.token.*
```

<a id="note-1"></a>
#### Note
- N-token is for identifying a service (i.e. garm) in Athenz. Athenz then use the pre-configurated policy to check whether the requested access is authenticated.
- N-token is sent to Athenz on every authentication request on the HTTP header with name `athenz.auth_header`.
- If `athenz.token.ntoken_path` is set ([Copper Argos](https://github.com/yahoo/athenz/blob/master/docs/copper_argos_dev.md)), garm will use the n-token in the file directly.
	- It is better to set `athenz.token.validate_token: true` in this case.
- If `athenz.token.ntoken_path` is NOT set, garm will handle the token generation and update automatically.
	- As the token is signed by `athenz.token.private_key_env_name`, please make sure that the corresponding public key is configurated in Athenz with the same `athenz.token.key_version`.

---

<a id="request-filtering"></a>
## Request filtering

<a id="related-configuration-2"></a>
### Related configuration
```yaml
map_rule.tld.platform.black_list
map_rule.tld.platform.white_list

map_rule.tld.service_athenz_domains
```

<a id="note-2"></a>
#### Note
- Garm can directly reject kube-apiserver requests without querying Athenz.
- `in black_list AND NOT in white_list` => directly reject
- Support wildcard `*` matching.

---

<a id="admin-domain"></a>
## Admin domain

<a id="related-configuration-3"></a>
### Related configuration
```yaml
map_rule.tld.platform.admin_access_list
map_rule.tld.platform.admin_athenz_domain
```

<a id="note-3"></a>
#### Note
- Garm can map kube-apiserver requests using a separate admin domain in Athenz.
- If the request matches any rules in `map_rule.tld.platform.admin_access_list`, garm will use `map_rule.tld.platform.admin_athenz_domain`.
- Garm will send 1 more request than the number of `map_rule.tld.service_athenz_domains` to Athenz. The kube-apiserver request is allowed if any 1 is allowed in Athenz (OR logic).
- If `service_domain_a` and `service_domain_b` are specified in `map_rule.tld.service_athenz_domains`, it is requested 3 times.
	1. Athenz resource **with** `service_domain_a` (One of those specified in `map_rule.tld.service_athenz_domains`)
	1. Athenz resource **with** `service_domain_b` (One of those specified in `map_rule.tld.service_athenz_domains`)
	1. Athenz resource **without** `map_rule.tld.service_athenz_domains`

---

<a id="service-domain"></a>
## Service domains

<a id="related-configuration-3"></a>
### Related configuration
```yaml
map_rule.tld.service_athenz_domains
```

<a id="note-3"></a>
#### Note
- If the request not matches any rules in `map_rule.tld.platform.admin_access_list`, garm will use `map_rule.tld.service_athenz_domains`.
- Garm will send request number of `map_rule.tld.service_athenz_domains` to Athenz. The kube-apiserver request is allowed if any 1 is allowed in Athenz (OR logic).
- If `service_domain_a` and `service_domain_b` are specified, garm will be requested twice.

---


<a id="resource-mapping"></a>
## Resource mapping

<a id="related-configuration-4"></a>
### Related configuration
```yaml
map_rule.tld.platform.resource_mappings
map_rule.tld.platform.verb_mappings
```

<a id="note-4"></a>
#### Note
- Garm can map k8s resource to Athenz resource.
- `spec.resourceAttributes.subresource` is appended to `spec.resourceAttributes.resource` before mapping as `spec.resourceAttributes.resource` with format `${resource}.${subresource}`.

---

<a id="optional-api-group-and-resource-name-control"></a>
## Optional API group and resource name control

<a id="related-configuration-5"></a>
### Related configuration
```yaml
map_rule.tld.platform.api_group_control
map_rule.tld.platform.api_group_mappings

map_rule.tld.platform.resource_name_control
map_rule.tld.platform.resource_name_mappings
```

<a id="note-5"></a>
#### Note
- Garm will only map `spec.resourceAttributes.group` and `spec.resourceAttributes.name` in kube-apiserver request body when `map_rule.tld.platform.*_control` is `true`. Else, they will be treated as `""` during mapping.

---

<a id="mapping-for-non-resources-or-empty-namespace"></a>
## Mapping for non-resources or empty namespace

<a id="related-configuration-6"></a>
### Related configuration
```yaml
map_rule.tld.platform.empty_namespace

map_rule.tld.platform.non_resource_api_group
map_rule.tld.platform.non_resource_namespace
```

<a id="note-6"></a>
#### Note
- Garm can substitute empty or missing value from kube-apiserver request with above configuration.
- In case of non-resource, resource is equal to `spec.non-resource-attributes.path`.

---

<a id="ps"></a>
## P.S.
- Above resources,
	- `k8s resource`: ([refer](https://github.com/kubernetes/apiserver/blob/master/plugin/pkg/authorizer/webhook/webhook.go#L165))
	- `resource`: a variable inside garm
	- `Athenz resource`: resource inside policy
