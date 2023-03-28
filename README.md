[![CircleCI](https://circleci.com/gh/giantswarm/retagger.svg?style=shield)](https://circleci.com/gh/giantswarm/retagger)

# retagger

> A tool to handle the retagging of third party docker images and make them
  available in own registries.

## CAVEAT

⚠️ Although development branch has been merged to `main`, the refactoring of
`retagger` is still a work-in-progress ⚠️

Why? Mainly due to an increase in the number of images that need retagging. The
new version handles multi-architecture images, which means instead of copying
one image per tag, we copy between 3 and 10 images for multiple architectures.
Also, connection to Aliyun is slow and unreliable, so we need a bunch of
retries.

At this point individual job fails are still expected. This does not mean your
image won't be retagged at all - even if copying a single tag fails, the job
carries on with all the other tags in the list.

In the following weeks, Team Honeybadger will be introducing ownership of jobs
submitted for retagger and the files in `images/` will be split by team.

<details>

<summary>Table: number of tags to synchronize per image</summary>

Bear in mind, the numbers are **not multiplied** by number of available
architecture variants. The actual numbers are 3x to 10x higher.

- Total: 4080 images
- The percentages **in bold** add up to 50.39% of all images. These are **Top 21
  Positions** taking up over **half** of the retagging time.


| Name                                                             |   Count    | % of total |
|------------------------------------------------------------------|------------|------------|
|grafana/grafana                                                   |     189    | **4.63%**  |
|registry.k8s.io/kube-controller-manager                           |     154    | **3.77%**  |
|registry.k8s.io/kube-proxy                                        |     154    | **3.77%**  |
|bitnami/kubectl                                                   |     152    | **3.73%**  |
|fluent/fluent-bit                                                 |     109    | **2.67%**  |
|registry.k8s.io/kube-apiserver                                    |     107    | **2.62%**  |
|vault                                                             |     103    | **2.52%**  |
|quay.io/argoproj/argocd                                           |     102    | **2.50%**  |
|quay.io/calico/node                                               |      96    | **2.35%**  |
|quay.io/calico/cni                                                |      94    | **2.30%**  |
|quay.io/calico/kube-controllers                                   |      94    | **2.30%**  |
|quay.io/calico/pod2daemon-flexvol                                 |      94    | **2.30%**  |
|alpine                                                            |      93    | **2.28%**  |
|quay.io/calico/typha                                              |      90    | **2.21%**  |
|mcr.microsoft.com/oss/kubernetes/azure-cloud-controller-manager   |      81    | **1.99%**  |
|mcr.microsoft.com/oss/kubernetes/azure-cloud-node-manager         |      81    | **1.99%**  |
|prom/prometheus                                                   |      73    | **1.79%**  |
|registry.k8s.io/hyperkube                                         |      68    | **1.67%**  |
|quay.io/coreos/etcd                                               |      62    | **1.52%**  |
|quay.io/ceph/ceph                                                 |      60    | **1.47%**  |
|quay.io/jetstack/cert-manager-controller                          |      60    |   1.47%    |
|quay.io/jetstack/cert-manager-webhook                             |      60    |   1.47%    |
|quay.io/jetstack/cert-manager-cainjector                          |      57    |   1.40%    |
|kong                                                              |      52    |   1.27%    |
|aquasec/trivy                                                     |      51    |   1.25%    |
|envoyproxy/envoy                                                  |      44    |   1.08%    |
|quay.io/fairwinds/polaris                                         |      42    |   1.03%    |
|registry.k8s.io/cluster-api/cluster-api-controller                |      40    |   0.98%    |
|registry.k8s.io/cluster-api/kubeadm-bootstrap-controller          |      40    |   0.98%    |
|registry.k8s.io/cluster-api/kubeadm-control-plane-controller      |      40    |   0.98%    |
|ghcr.io/prymitive/karma                                           |      39    |   0.96%    |
|quay.io/prometheus-operator/prometheus-config-reloader            |      38    |   0.93%    |
|quay.io/prometheus-operator/prometheus-operator                   |      38    |   0.93%    |
|amazon/aws-efs-csi-driver                                         |      36    |   0.88%    |
|registry.k8s.io/cluster-api-azure/cluster-api-azure-controller    |      36    |   0.88%    |
|quay.io/fairwinds/goldilocks                                      |      31    |   0.76%    |
|aquasec/trivy-operator                                            |      29    |   0.71%    |
|quay.io/jetstack/cert-manager-acmesolver                          |      29    |   0.71%    |
|aquasec/kube-bench                                                |      27    |   0.66%    |
|curlimages/curl                                                   |      26    |   0.64%    |
|k8scloudprovider/openstack-cloud-controller-manager               |      26    |   0.64%    |
|k8scloudprovider/cinder-csi-plugin                                |      25    |   0.61%    |
|registry.k8s.io/dns/k8s-dns-node-cache                            |      25    |   0.61%    |
|grafana/loki                                                      |      24    |   0.59%    |
|grafana/promtail                                                  |      24    |   0.59%    |
|k8scloudprovider/octavia-ingress-controller                       |      24    |   0.59%    |
|openpolicyagent/conftest                                          |      24    |   0.59%    |
|quay.io/thanos/thanos                                             |      24    |   0.59%    |
|velero/velero                                                     |      24    |   0.59%    |
|bash                                                              |      23    |   0.56%    |
|quay.io/cilium/cilium                                             |      22    |   0.54%    |
|quay.io/cilium/hubble-relay                                       |      22    |   0.54%    |
|registry.k8s.io/sig-storage/csi-snapshotter                       |      22    |   0.54%    |
|mcr.microsoft.com/oss/azure/aad-pod-identity/mic                  |      21    |   0.51%    |
|mcr.microsoft.com/oss/azure/aad-pod-identity/nmi                  |      21    |   0.51%    |
|quay.io/jetstack/cert-manager-ctl                                 |      21    |   0.51%    |
|registry.k8s.io/sig-storage/csi-provisioner                       |      21    |   0.51%    |
|kong/kong-gateway                                                 |      20    |   0.49%    |
|aquasec/starboard-operator                                        |      19    |   0.47%    |
|grafana/grafana-image-renderer                                    |      19    |   0.47%    |
|registry.k8s.io/cluster-api-aws/cluster-api-aws-controller        |      19    |   0.47%    |
|cytopia/yamllint                                                  |      18    |   0.44%    |
|eu.gcr.io/k8s-artifacts-prod/autoscaling/cluster-autoscaler       |      18    |   0.44%    |
|mcr.microsoft.com/oss/kubernetes-csi/azuredisk-csi                |      18    |   0.44%    |
|mcr.microsoft.com/oss/kubernetes-csi/azurefile-csi                |      17    |   0.42%    |
|quay.io/cephcsi/cephcsi                                           |      17    |   0.42%    |
|registry.k8s.io/sig-storage/snapshot-controller                   |      17    |   0.42%    |
|busybox                                                           |      16    |   0.39%    |
|grafana/loki-canary                                               |      15    |   0.37%    |
|registry.k8s.io/kube-state-metrics/kube-state-metrics             |      15    |   0.37%    |
|registry.k8s.io/sig-storage/csi-attacher                          |      15    |   0.37%    |
|registry.k8s.io/sig-storage/csi-node-driver-registrar             |      15    |   0.37%    |
|bats/bats                                                         |      14    |   0.34%    |
|quay.io/jacksontj/promxy                                          |      14    |   0.34%    |
|amazon/opendistro-for-elasticsearch                               |      13    |   0.32%    |
|docker.elastic.co/kibana/kibana-oss                               |      13    |   0.32%    |
|ghcr.io/kyverno/policy-reporter-kyverno-plugin                    |      13    |   0.32%    |
|golang                                                            |      13    |   0.32%    |
|registry.k8s.io/descheduler/descheduler                           |      13    |   0.32%    |
|amazon/opendistro-for-elasticsearch-kibana                        |      12    |   0.29%    |
|gcr.io/k8s-staging-cloud-provider-gcp/gcp-compute-persistent-dis  |      12    |   0.29%    |
|prom/pushgateway                                                  |      12    |   0.29%    |
|registry.k8s.io/capi-openstack/capi-openstack-controller          |      12    |   0.29%    |
|registry.k8s.io/sig-storage/csi-resizer                           |      12    |   0.29%    |
|jimmidyson/configmap-reload                                       |      11    |   0.27%    |
|opensearchproject/opensearch                                      |      11    |   0.27%    |
|opensearchproject/opensearch-dashboards                           |      11    |   0.27%    |
|gcr.io/kubebuilder/kube-rbac-proxy                                |      10    |   0.25%    |
|registry.k8s.io/metrics-server/metrics-server                     |      10    |   0.25%    |
|registry.k8s.io/sig-storage/livenessprobe                         |      10    |   0.25%    |
|golang                                                            |       9    |   0.22%    |
|quay.io/prometheus/alertmanager                                   |       9    |   0.22%    |
|registry.k8s.io/coredns/coredns                                   |       9    |   0.22%    |
|registry.k8s.io/pause                                             |       9    |   0.22%    |
|amazon/aws-alb-ingress-controller                                 |       8    |   0.20%    |
|falcosecurity/falcosidekick                                       |       8    |   0.20%    |
|jettech/kube-webhook-certgen                                      |       8    |   0.20%    |
|quay.io/coreos/prometheus-config-reloader                         |       8    |   0.20%    |
|quay.io/coreos/prometheus-operator                                |       8    |   0.20%    |
|registry.k8s.io/addon-resizer                                     |       8    |   0.20%    |
|registry.k8s.io/autoscaling/vpa-admission-controller              |       8    |   0.20%    |
|registry.k8s.io/autoscaling/vpa-recommender                       |       8    |   0.20%    |
|registry.k8s.io/autoscaling/vpa-updater                           |       8    |   0.20%    |
|directxman12/k8s-prometheus-adapter-amd64                         |       7    |   0.17%    |
|falcosecurity/falco-driver-loader                                 |       7    |   0.17%    |
|registry.k8s.io/external-dns/external-dns                         |       7    |   0.17%    |
|crossplane/crossplane                                             |       6    |   0.15%    |
|falcosecurity/falco-exporter                                      |       6    |   0.15%    |
|quay.io/cilium/hubble-ui                                          |       6    |   0.15%    |
|quay.io/cilium/hubble-ui-backend                                  |       6    |   0.15%    |
|quay.io/prometheus/node-exporter                                  |       6    |   0.15%    |
|registry.k8s.io/cluster-api-gcp/cluster-api-gcp-controller        |       6    |   0.15%    |
|registry.k8s.io/etcd                                              |       6    |   0.15%    |
|ealen/echo-server                                                 |       5    |   0.12%    |
|gcr.io/cadvisor/cadvisor                                          |       5    |   0.12%    |
|quay.io/dexidp/dex                                                |       5    |   0.12%    |
|quay.io/oauth2-proxy/oauth2-proxy                                 |       5    |   0.12%    |
|quay.io/open-policy-agent/gatekeeper                              |       5    |   0.12%    |
|quay.io/prometheus/node-exporter                                  |       5    |   0.12%    |
|registry.k8s.io/cluster-proportional-autoscaler-amd64             |       5    |   0.12%    |
|spvest/azure-keyvault-controller                                  |       5    |   0.12%    |
|coredns/coredns                                                   |       4    |   0.10%    |
|falcosecurity/falco                                               |       4    |   0.10%    |
|falcosecurity/falco-no-driver                                     |       4    |   0.10%    |
|omnition/opencensus-collector                                     |       4    |   0.10%    |
|prom/prometheus                                                   |       4    |   0.10%    |
|redis                                                             |       4    |   0.10%    |
|registry.k8s.io/autoscaling/cluster-autoscaler                    |       4    |   0.10%    |
|spvest/azure-keyvault-webhook                                     |       4    |   0.10%    |
|aquasec/scanner                                                   |       3    |   0.07%    |
|ghcr.io/external-secrets/external-secrets                         |       3    |   0.07%    |
|public.ecr.aws/aws-ec2/aws-node-termination-handler               |       3    |   0.07%    |
|python                                                            |       3    |   0.07%    |
|quay.io/uswitch/kiam                                              |       3    |   0.07%    |
|registry.k8s.io/cluster-api-aws/eks-bootstrap-controller          |       3    |   0.07%    |
|registry.k8s.io/cluster-api-aws/eks-controlplane-controller       |       3    |   0.07%    |
|registry.k8s.io/pause-amd64                                       |       3    |   0.07%    |
|alpine                                                            |       2    |   0.05%    |
|amazon/amazon-eks-pod-identity-webhook                            |       2    |   0.05%    |
|gcr.io/google_containers/defaultbackend                           |       2    |   0.05%    |
|ghcr.io/k8snetworkplumbingwg/multus-cni                           |       2    |   0.05%    |
|ghcr.io/kyverno/kyverno                                           |       2    |   0.05%    |
|ghcr.io/kyverno/kyvernopre                                        |       2    |   0.05%    |
|ghcr.io/opsgenie/kubernetes-event-exporter                        |       2    |   0.05%    |
|instrumenta/conftest                                              |       2    |   0.05%    |
|quay.io/giantswarm/amazon-k8s-cni                                 |       2    |   0.05%    |
|zricethezav/gitleaks                                              |       2    |   0.05%    |
|centos                                                            |       1    |   0.02%    |
|cibuilds/github                                                   |       1    |   0.02%    |
|docker                                                            |       1    |   0.02%    |
|docker.elastic.co/elasticsearch/elasticsearch-oss                 |       1    |   0.02%    |
|elasticsearch                                                     |       1    |   0.02%    |
|fluxcd/flux-cli                                                   |       1    |   0.02%    |
|gcr.io/google-containers/startup-script                           |       1    |   0.02%    |
|gcr.io/heptio-images/gangway                                      |       1    |   0.02%    |
|gcr.io/heptio-images/kube-conformance                             |       1    |   0.02%    |
|gcr.io/spark-operator/spark-operator                              |       1    |   0.02%    |
|ghcr.io/inlets/inlets-operator                                    |       1    |   0.02%    |
|ghcr.io/inlets/inlets-pro                                         |       1    |   0.02%    |
|ghcr.io/kyverno/cleanup-controller                                |       1    |   0.02%    |
|ghcr.io/kyverno/policy-reporter                                   |       1    |   0.02%    |
|ghcr.io/kyverno/policy-reporter-ui                                |       1    |   0.02%    |
|goharbor/chartmuseum-photon                                       |       1    |   0.02%    |
|goharbor/clair-photon                                             |       1    |   0.02%    |
|goharbor/harbor-adminserver                                       |       1    |   0.02%    |
|goharbor/harbor-db                                                |       1    |   0.02%    |
|goharbor/harbor-jobservice                                        |       1    |   0.02%    |
|goharbor/harbor-ui                                                |       1    |   0.02%    |
|goharbor/notary-server-photon                                     |       1    |   0.02%    |
|goharbor/notary-signer-photon                                     |       1    |   0.02%    |
|goharbor/registry-photon                                          |       1    |   0.02%    |
|golangci/golangci-lint                                            |       1    |   0.02%    |
|jgsqware/fluentd-loki-plugin                                      |       1    |   0.02%    |
|jimschubert/swagger-codegen-cli                                   |       1    |   0.02%    |
|jollinshead/journald-cloudwatch-logs                              |       1    |   0.02%    |
|justwatch/elasticsearch_exporter                                  |       1    |   0.02%    |
|k8spin/loki-multi-tenant-proxy                                    |       1    |   0.02%    |
|koalaman/shellcheck-alpine                                        |       1    |   0.02%    |
|looztra/kubesplit                                                 |       1    |   0.02%    |
|madnight/alpine-wkhtmltopdf-builder                               |       1    |   0.02%    |
|mcr.microsoft.com/azuremonitor/containerinsights/ciprod           |       1    |   0.02%    |
|mintel/dex-k8s-authenticator                                      |       1    |   0.02%    |
|nginx                                                             |       1    |   0.02%    |
|nginxinc/nginx-unprivileged                                       |       1    |   0.02%    |
|ns1labs/flame                                                     |       1    |   0.02%    |
|peaceiris/hugo                                                    |       1    |   0.02%    |
|prom/cloudwatch-exporter                                          |       1    |   0.02%    |
|quay.io/cilium/cilium-etcd-operator                               |       1    |   0.02%    |
|quay.io/coreos/configmap-reload                                   |       1    |   0.02%    |
|quay.io/coreos/etcd-operator                                      |       1    |   0.02%    |
|quay.io/coreos/flannel                                            |       1    |   0.02%    |
|quay.io/coreos/prometheus-operator                                |       1    |   0.02%    |
|quay.io/giantswarm/docker-strongswan                              |       1    |   0.02%    |
|quay.io/giantswarm/k8s-api-healthz                                |       1    |   0.02%    |
|quay.io/giantswarm/k8s-setup-network-environment                  |       1    |   0.02%    |
|quay.io/google-cloud-tools/kube-eagle                             |       1    |   0.02%    |
|quay.io/jetstack/cert-manager-ingress-shim                        |       1    |   0.02%    |
|quay.io/prometheus/haproxy-exporter                               |       1    |   0.02%    |
|quay.io/pusher/oauth2_proxy                                       |       1    |   0.02%    |
|squareup/ghostunnel                                               |       1    |   0.02%    |
|toniblyx/prowler                                                  |       1    |   0.02%    |
|weaveworks/watch                                                  |       1    |   0.02%    |

</details>


## What does retagger do, exactly?

`retagger` is first and foremost a CircleCI worfklow that runs every day at 21:30
UTC and on every merge to master branch. It utilizes [skopeo][skopeo] and
[custom golang code](main.go) to take upstream docker images, customize them if
necessary, and push them to Giant Swarm's container registries: `quay.io` and
`giantswarm-registry.cn-shanghai.cr.aliyuncs.com`. It is capable of working
with `v1`, `v2`, and `OCI` registries, as well as retagging multi-architecture
images.

> 💡Please note it **is not responsible** for pushing images to neither
`docker.io/giantswarm`, nor `azurecr.io/giantswarm` container registries.

## How to add your image to the job

You've come to the right place. Pick one of the below methods. For both methods first make sure
the repository exists in the desired container registry.

### Plain copy

You do **not** need any customizations. Great!
1. Find a `skopeo-*.yaml` file in [images](images/) matching your upstream
   container registry's name. Create a new one, if necessary.
2. Add a tag, SHA, or a semantic version constraint for your image. Refer to
   [Skopeo](#skopeo) section or existing files for format definition.
3. If you haven't created a new file, that's it. You're set. Otherwise continue
   following the steps.
4. Open [CircleCI config][ciconf] and add your file to both `retag-registry`
   steps under `matrix.parameters.images_file`.

### Modify upstream Dockerfile

You need to modify the upstream image before it's pushed to Giant Swarm registries.
1. Find the [customized-images.yaml][custom] and add your image to the list.
   Please refer to [Customized Images](#customized-images) section to see the
   options.

### Manual copy

You need to copy a few tags, it's a one-off situation. You can use `docker
pull/docker tag/docker push` combination or the below `skopeo` snippet:

```bash
$ skopeo sync --src docker --dest docker --all --keep-going crossplane/crossplane:v1.11.0 docker.io/giantswarm/
```

## Image list formats

### Skopeo

The basic file format looks as follows:
`images/skopeo-registry-example-com.yaml`
```yaml
registry.example.com:
    images:
        redis:
            - "1.0"
            - "2.0"
            - "sha256:0000000000000000000000000000000011111111111111111111111111111111"
    images-by-semver:
        alpine:
            - "3.12 - 3.13"
            - ">= 3.17"
```

The full specification is available in [upstream skopeo-sync docs][skopeo-sync
docs]. Semantic version constraint documentation is available in
[Masterminds/semver docs][masterminds docs].

### Customized images

Customized images are represented as an array of `CustomImage` objects. Please
see the below definition:

```golang
type CustomImage struct {
	// Image is the full name of the image to pull.
	// Example: "alpine", "docker.io/giantswarm/app-operator", or
	// "ghcr.io/fluxcd/kustomize-controller"
	Image string `yaml:"image"`
	// TagOrPattern is used to filter image tags. All tags matching the pattern
	// will be retagged. Required if SHA is specified.
	// Example: "v1.[234].*" or ".*-stable"
	TagOrPattern string `yaml:"tag_or_pattern,omitempty"`
	// SHA is used to filter image tags. If SHA is specified, it will take
	// precedence over TagOrPattern. However TagOrPattern is still required!
	// Example: 234cb88d3020898631af0ccbbcca9a66ae7306ecd30c9720690858c1b007d2a0
	SHA string `yaml:"sha,omitempty"`
	// Semver is used to filter image tags by semantic version constraints. All
	// tags satisfying the constraint will be retagged.
	Semver string `yaml:"semver,omitempty"`
	// Filter is a regexp pattern used to extract a part of the tag for Semver
	// comparison. First matched group will be supplied for semver comparison.
	// Example:
	//   Filter: "(.+)-alpine"  ->  Image tag: "3.12-alpine" -> Comparison: "3.12>=3.10"
	//   Semver: ">= 3.10"          Extracted group: "3.12"
	Filter string `yaml:"filter,omitempty"`
	// DockerfileExtras is a list of additional Dockerfile statements you want to
	// append to the upstream Dockerfile. (optional)
	// Example: ["RUN apk add -y bash"]
	DockerfileExtras []string `yaml:"dockerfile_extras,omitempty"`
	// AddTagSuffix is an extra string to append to the tag.
	// Example: "giantswarm", the tag would become "<tag>-giantswarm"
	AddTagSuffix string `yaml:"add_tag_suffix,omitempty"`
	// OverrideRepoName allows user to rewrite the name of the image entirely.
	// Example: "alpinegit", so "alpine" would become
	// "quay.io/giantswarm/alpinegit"
	OverrideRepoName string `yaml:"override_repo_name,omitempty"`
	// StripSemverPrefix removes the initial 'v' in 'v1.2.3' if enabled. Works
	// only when Semver is defined.
	StripSemverPrefix bool `yaml:"strip_semver_prefix,omitempty"`
}
```

## Contributing

Please refer to [CONTRIBUTING.md](CONTRIBUTING.md).

[skopeo]: https://github.com/containers/skopeo
[skopeo-sync docs]: https://github.com/kubasobon/skopeo/blob/semver/docs/skopeo-sync.1.md#yaml-file-content-used-source-for---src-yaml
[masterminds docs]: https://github.com/Masterminds/semver/tree/v3.2.0#basic-comparisons

[ciconf]: .circleci/config.yml
[custom]: images/customized-images.yaml
