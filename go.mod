module github.com/yahoojapan/garm

go 1.16

replace k8s.io/client-go => k8s.io/client-go v0.18.3

require (
	github.com/falz-tino/k8s-athenz-webhook v0.1.1-0.20210210051049-e59f25e8bd97
	github.com/kpango/glg v1.5.1
	github.com/pkg/errors v0.9.1
	github.com/yahoo/athenz v1.9.6
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.18.3
)
