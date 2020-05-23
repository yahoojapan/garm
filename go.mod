module github.com/yahoojapan/garm

go 1.14

replace k8s.io/client-go => k8s.io/client-go v0.18.3

require (
	github.com/kpango/glg v1.5.1
	github.com/pkg/errors v0.9.1
	github.com/yahoo/athenz v1.9.4
	github.com/yahoo/k8s-athenz-webhook v0.1.0
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/api v0.18.3
)
