[![License: Apache](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://opensource.org/licenses/Apache-2.0) [![release](https://img.shields.io/github/release/yahoojapan/garm.svg?style=flat-square)](https://github.com/yahoojapan/garm/releases/latest) [![CircleCI](https://circleci.com/gh/yahoojapan/garm.svg)](https://circleci.com/gh/yahoojapan/garm) [![codecov](https://codecov.io/gh/yahoojapan/garm/branch/master/graph/badge.svg?token=2CzooNJtUu&style=flat-square)](https://codecov.io/gh/yahoojapan/garm) [![Go Report Card](https://goreportcard.com/badge/github.com/yahoojapan/garm)](https://goreportcard.com/report/github.com/yahoojapan/garm) [![GolangCI](https://golangci.com/badges/github.com/yahoojapan/garm.svg?style=flat-square)](https://golangci.com/r/github.com/yahoojapan/garm) [![Codacy Badge](https://api.codacy.com/project/badge/Grade/32397d339f6c450a82af72c8a0c15e5f)](https://www.codacy.com/app/i.can.feel.gravity/garm?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=yahoojapan/garm&amp;utm_campaign=Badge_Grade) [![GoDoc](http://godoc.org/github.com/yahoojapan/garm?status.svg)](http://godoc.org/github.com/yahoojapan/garm) [![DepShield Badge](https://depshield.sonatype.org/badges/yahoojapan/garm/depshield.svg)](https://depshield.github.io)

![logo](./images/logo.png)

---

## What is Garm
Garm is API for a Kubernetes authorization webhook that integrates with [Athenz](https://github.com/yahoo/athenz) 
for access checks. It allows flexible resource mapping from K8s resources to Athenz ones.

You can also use just the authorization hook without also using the authentication hook.
Use of the authentication hook requires Athenz to be able to sign tokens for users.

Requires go 1.11 or later.

## Use case
### Authorization
![Use case](./doc/assets/use-case.png)

 1. K8s webhook request (SubjectAccessReview) ([Webhook Mode - Kubernetes](https://kubernetes.io/docs/reference/access-authn-authz/webhook/))
    - the K8s API server wants to know if the user is allowed to do the requested action
 2. Athenz RBAC request ([Athenz](http://www.athenz.io/))
    - Athenz server contains the user authorization information for access control
    - ask Athenz server is the user action is allowed based on pre-configurated policy

Garm convert the K8s request to Athenz request based on the mapping rules in `config.yaml` ([example](./config/testdata/example_config.yaml)).
  - [Conversion logic](./doc/garm-functional-overview.md)
  - [Config details](./doc/config-detail.md)

P.S. It is just a sample deployment solution above. Garm can work on any environment as long as it can access both the API server and the Athenz server.

### Docker
```shell
$ docker pull yahoojapan/garm
```

### Usage
  - [install Garm](https://github.com/yahoojapan/garm/blob/master/doc/installation/02.%20install-garm.md)
  - [configure k8s webhook](https://github.com/yahoojapan/garm/blob/master/doc/installation/03.%20config-k8s-in-webhook-mode.md)
  - [configure Athenz & Garm yaml](./doc/config-detail.md)

## CI/CD
  - [CircleCI](https://circleci.com/gh/yahoojapan/garm)

## Futurework
 1. Authentication support for Garm
 2. Helm Support
 3. mTLS Support between Athenz and Garm
 4. multi Athenz domain support

## License
```markdown
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
```

## Contributor License Agreement

This project requires contributors to agree to a [Contributor License Agreement (CLA)](https://gist.github.com/ydnjp/3095832f100d5c3d2592).

Note that only for contributions to the garm repository on the [GitHub](https://github.com/yahoojapan/garm), the contributors of them shall be deemed to have agreed to the CLA without individual written agreements.

## Authors
  - [kpango](https://github.com/kpango)
  - [kevindiu](https://github.com/kevindiu)
  - [WindzCUHK](https://github.com/WindzCUHK)
  - [tatyano](https://github.com/tatyano)
  - [rinx](https://github.com/rinx)
