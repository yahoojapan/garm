module github.com/yahoojapan/garm

go 1.13

replace (
	k8s.io/api => k8s.io/api v0.0.0-20190918195907-bd6ac527cfd2
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190918201827-3de75813f604
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190918200256-06eb1244587a
)

require (
	github.com/kpango/glg v1.4.6
	github.com/kr/pretty v0.1.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/yahoo/athenz v1.8.33
	github.com/yahoo/k8s-athenz-webhook v0.0.0-20190725182459-949d9ed74720
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/api v0.0.0-20190918195907-bd6ac527cfd2
)
