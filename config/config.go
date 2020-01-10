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

package config

import (
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/pkg/errors"
	webhook "github.com/yahoo/k8s-athenz-webhook"
	"gopkg.in/yaml.v2"
)

const (
	// currentVersion represents the configuration version.
	currentVersion = "v2.0.0"
)

// Config represents an application configuration content (config.yaml).
// In K8s environment, this configuration is stored in K8s ConfigMap.
type Config struct {
	// Version represents configuration file version.
	Version string `yaml:"version"`

	// EnableColorLogging represents if user want to enable colorized logging.
	EnableColorLogging bool `yaml:"enable_log_color"`

	// Logger represents logging configuration for Garm application.
	Logger Logger `yaml:"logger"`

	// Server represents webhook server and health check server configuration.
	Server Server `yaml:"server"`

	// Athenz represents Athenz configuration for Garm to connect to Athenz server.
	Athenz Athenz `yaml:"athenz"`

	// Token represents configuration to generate n-token for connecting to Athenz.
	Token Token `yaml:"token"`

	// Mapping represents the mapping rule for mapping K8s authentication and authorization requests to Athenz requests.
	Mapping Mapping `yaml:"map_rule"`
}

// Logger represents logging configuration for Garm.
type Logger struct {
	// LogPath represents log file path.
	LogPath string `yaml:"log_path"`

	// LogTrace represents the event to be logged.
	// LogTrace is a comma separated list, the value can be "server", "athenz" and "mapping".
	LogTrace string `yaml:"log_trace"`
}

// Server represents webhook server and health check server configuration.
type Server struct {
	// Port represents webhook server port.
	Port int `yaml:"port"`

	// HealthzPort represents health check server port.
	HealthzPort int `yaml:"health_check_port"`

	// HealthzPath represents the API path (pattern) for health check server.
	HealthzPath string `yaml:"health_check_path"`

	// Timeout represents the maximum webhook server request handling duration.
	Timeout string `yaml:"timeout"`

	// ShutdownDuration represents the maximum shutdown duration.
	ShutdownDuration string `yaml:"shutdown_duration"`

	// ProbeWaitTime represents the pause duration before shutting down webhook server after health check server shutdown.
	ProbeWaitTime string `yaml:"probe_wait_time"`

	// TLS represents the TLS configuration for webhook server.
	TLS TLS `yaml:"tls"`
}

// TLS represents the TLS configuration for webhook server.
type TLS struct {
	// Enable represents the webhook server enable TLS or not.
	Enabled bool `yaml:"enabled"`

	// Cert represents the certificate file path of webhook server.
	Cert string `yaml:"cert"`

	// Key represents the private key file path of webhook server certificate.
	Key string `yaml:"key"`

	// CA represents the CA certificates file path for verifying clients connecting to webhook server.
	CA string `yaml:"ca"`
}

// Athenz represents the configuration for webhook server to connect to Athenz.
type Athenz struct {
	// AuthHeader represents the HTTP request header key name to attach the n-token for authentication requests to Athenz.
	AuthHeader string `yaml:"auth_header"`

	// URL represents the Athenz (ZMS and ZTS) URL handle authentication and authorization request.
	URL string `yaml:"url"`

	// Timeout represents the request timeout duration to Athenz server.
	Timeout string `yaml:"timeout"`

	// AthenzRootCA is the Athenz root CA certificate file path for connecting to Athenz.
	AthenzRootCA string `yaml:"root_ca"`

	// AuthN represents the authentication configuration.
	AuthN webhook.AuthenticationConfig

	// AuthZ represents the authorization configuration.
	AuthZ webhook.AuthorizationConfig

	// Config is the common configuration for authentication and authorization server.
	Config webhook.Config
}

// Token represents the token generation details or the n-token file for Copper Argos.
type Token struct {
	// AthenzDomain represents the Athenz domain value to generate the n-token.
	AthenzDomain string `yaml:"athenz_domain"`

	// ServiceName represents the Athenz service name value to generate the n-token.
	ServiceName string `yaml:"service_name"`

	// NTokenPath represents the n-token path. Only for Copper Argos.
	NTokenPath string `yaml:"ntoken_path"`

	// PrivateKey represents the private key file path for signing the n-token.
	PrivateKey string `yaml:"private_key"`

	// ValidateToken represents should validate the token or not. Set true when NTokenPath is set.
	ValidateToken bool `yaml:"validate_token"`

	// RefreshDuration represents the token refresh duration, no matter it is generated, or it is get from the token file (Copper Argos).
	RefreshDuration string `yaml:"refresh_duration"`

	// KeyVersion represents the key version of the n-token.
	KeyVersion string `yaml:"key_version"`

	// Expiration represents the duration of the expiration.
	Expiration string `yaml:"expiration"`
}

// Mapping represents the mapping rules from K8s authentication and authorization requests to Athenz requests.
type Mapping struct {
	// TLD represents the mapping rules for each Top Level Domain.
	TLD TLD `yaml:"tld"`
}

// TLD represents the mapping rules for each Top Level Domain.
type TLD struct {
	// Name represents the top level domain name.
	Name string `yaml:"name"`

	// Platform represents the mapping rules for each K8s request.
	Platform Platform `yaml:"platform"`
}

// Platform represents the mapping rules for each K8s request.
type Platform struct {
	// Name represents the platform name. Currently, it supports "k8s", "aks" and "k8s".
	Name string `yaml:"name"`

	// ServiceAthenzDomains represents the Athenz domain name used for non-administrative K8s webhook requests.
	ServiceAthenzDomains []string `yaml:"service_athenz_domains"`

	// ResourceMappings maps the K8s webhook request "resource" to the "resource" part of Athenz resource name.
	ResourceMappings map[string]string `yaml:"resource_mappings,omitempty"`

	// VerbMappings maps the K8s webhook request "verb" to the "verb" part of Athenz resource name.
	VerbMappings map[string]string `yaml:"verb_mappings,omitempty"`

	// APIGroupControlEnable enables the use of API group in Athenz resource name.
	APIGroupControlEnabled bool `yaml:"api_group_control"`

	// APIGroupMappings maps the K8s webhook request "group" to the "group" part of Athenz resource name.
	APIGroupMappings map[string]string `yaml:"api_group_mappings"`

	// EmptyNamespace represents the value used when the K8s webhook request "namespace" is empty.
	EmptyNamespace string `yaml:"empty_namespace"`

	// ResourceNameControlEnable enables the use of resource name in Athenz resource name.
	ResourceNameControlEnabled bool `yaml:"resource_name_control"`

	// ResourceNameMappings maps the K8s webhook request "name" to the "name" part of Athenz resource name.
	ResourceNameMappings map[string]string `yaml:"resource_name_mappings"`

	// ResourceNameReplacer represents the replacer to replace ":" or some values to another string.
	ResourceNameReplacer map[string]string `yaml:"resource_name_replacer"`

	// NonResourceAPIGroup represents the API group value (Athenz resource name) used for K8s non-resource webhook requests.
	NonResourceAPIGroup string `yaml:"non_resource_api_group"`

	// NonResourceNamespace represents the namespace value (Athenz resource name) used for K8s non-resource webhook requests.
	NonResourceNamespace string `yaml:"non_resource_namespace"`

	// ServiceAccountPrefixes represents the service account prefix for identifying service accounts.
	ServiceAccountPrefixes []string `yaml:"service_account_prefixes"`

	// AthenzUserPrefix represents the Athenz user prefix used when the K8s webhook request is from a user account.
	AthenzUserPrefix string `yaml:"athenz_user_prefix"`

	// AthenzServiceAccountPrefix represents the Athenz service account prefix used when the K8s webhook request is from a service account.
	AthenzServiceAccountPrefix string `yaml:"athenz_service_account_prefix"`

	// AdminAthenzDomain represents the Athenz domain name used for administrative K8s webhook requests.
	AdminAthenzDomain string `yaml:"admin_athenz_domain"`

	// AdminAccessList represents the list of administrative K8s webhook request patterns, which should use the admin domain in Athenz.
	AdminAccessList []*RequestInfo `yaml:"admin_access_list"`

	// Whitelist represents the list of whitelist K8s webhook request patterns. These requests will always trigger requests to Athenz to do authorization check.
	WhiteList []*RequestInfo `yaml:"white_list"`

	// Blacklist represents the list of blacklist K8s webhook request patterns. These requests will always rejected by Garm directly.
	BlackList []*RequestInfo `yaml:"black_list"`
}

// RequestInfo represents the rule of the K8s webhook request.
type RequestInfo struct {
	// Verb represents the K8s verb field inside K8s webhook request.
	Verb string `yaml:"verb"`

	// Namespace represents the K8s namespace field inside K8s webhook request.
	Namespace string `yaml:"namespace"`

	// APIGroup represents the K8s API Group field inside K8s webhook request.
	APIGroup string `yaml:"api_group"`

	// Resource represents the K8s resource field inside K8s webhook request.
	Resource string `yaml:"resource"`

	// Name represents the K8s resource name field inside K8s webhook request.
	Name string `yaml:"name"`

	// reg represents the compiled regexp for matching another RequestInfo.
	reg *regexp.Regexp

	// once ensure that the reg is compiled only once.
	once *sync.Once
}

// Serialize returns RequestInfo in string format.
// 1. replacedAPIGroup = replace `. => _` in r.APIGroup
// 2. output format: `${r.Verb}-${r.Namespace}-${replacedAPIGroup}-${r.Resource}-${r.Name}`
func (r *RequestInfo) Serialize() string {
	return strings.Join([]string{r.Verb, r.Namespace, strings.Replace(r.APIGroup, ".", "_", -1), r.Resource, r.Name}, "-")
}

// Match checks if the given RequestInfo matches with the regular expression in this RequestInfo.
// 1. r.Serialize()
// 2. replace `* => .*`
// 3. replace `..* => .*`
// return is regexp match
func (r *RequestInfo) Match(req RequestInfo) bool {
	if r.once == nil {
		r.once = new(sync.Once)
	}
	r.once.Do(func() {
		r.reg = regexp.MustCompile(strings.Replace(strings.Replace(r.Serialize(), "*", ".*", -1), "..*", ".*", -1))
	})

	return r.reg.Copy().MatchString(req.Serialize())
}

// New returns the decoded configuration YAML file as *Config struct. Returns non-nil error if any.
func New(path string) (*Config, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0600)
	if err != nil {
		return nil, errors.Wrap(err, "config read failed")
	}
	cfg := new(Config)
	err = yaml.NewDecoder(f).Decode(&cfg)
	if err != nil {
		return nil, errors.Wrap(err, "yaml parse failed")
	}
	return cfg, nil
}

// GetVersion returns the current configuration version of Garm.
func GetVersion() string {
	return currentVersion
}

// GetActualValue returns the environment variable value if the given val has "_" prefix and suffix, otherwise returns val directly.
func GetActualValue(val string) string {
	if checkPrefixAndSuffix(val, "_", "_") {
		return os.Getenv(strings.TrimPrefix(strings.TrimSuffix(val, "_"), "_"))
	}
	return val
}

// checkPrefixAndSuffix checks if the given string has given prefix and suffix.
func checkPrefixAndSuffix(str, pref, suf string) bool {
	return strings.HasPrefix(str, pref) && strings.HasSuffix(str, suf)
}
