# HyperRegistry
Hyperregistry is a private artifact (images, helmcharts, etc.) registry based on an open source [Harbor](https://github.com/goharbor/harbor). We use [Helm chart](https://github.com/tmax-cloud/harbor-helm) to deploy hyperregistry on k8s. You can push and pull images like Docker hub, and you can use [Notary](https://github.com/notaryproject/notary) to perform [DCT](https://docs.docker.com/engine/security/trust/)-based image signing. Also, because [Trivy](https://github.com/aquasecurity/trivy) open source is loaded, image vulnerability inspection is possible.

## Overview
* 목적: 이미지를 저장하고 관리(서명, 스캔)하기 위함
* 역할: Hypercloud 내 모든 이미지를 저장하고 관리하는 용도

# Harbor Documentation

All Harbor documentation is presented on [goharbor.io/docs](https://goharbor.io/docs).

To contribute to the documentation, please head over to the [website repository](https://github.com/goharbor/website).
