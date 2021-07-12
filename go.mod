module github.com/yahoojapan/garm

go 1.14

replace k8s.io/client-go => k8s.io/client-go v0.21.2

require (
	github.com/AthenZ/athenz v1.10.24
	github.com/kpango/glg v1.6.0
	github.com/pkg/errors v0.9.1
	github.com/yahoo/k8s-athenz-webhook v0.1.1
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.21.2
)
