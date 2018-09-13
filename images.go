package main

import (
	"fmt"
	"strings"
)

type Image struct {
	Name string
	Tags []Tag
}

type Tag struct {
	Sha string
	Tag string
}

var Images = []Image{
	Image{
		Name: "alpine",
		Tags: []Tag{
			Tag{
				Sha: "7df6db5aa61ae9480f52f0b3a06a140ab98d427f86d8d5de0bedab9b8df6b1c0",
				Tag: "3.7",
			},
		},
	},
	Image{
		Name: "busybox",
		Tags: []Tag{
			Tag{
				Sha: "58ac43b2cc92c687a32c8be6278e50a063579655fe3090125dcb2af0ff9e1a64",
				Tag: "1.28.3",
			},
		},
	},
	Image{
		Name: "coredns/coredns",
		Tags: []Tag{
			Tag{
				Sha: "399cc5b2e2f0d599ef22f43aab52492e88b4f0fd69da9b10545e95a4253c86ce",
				Tag: "1.1.1",
			},
		},
	},
	Image{
		Name: "dduportal/bats",
		Tags: []Tag{
			Tag{
				Sha: "b2d533b27109f7c9ea1e270e23f212c47906346f9cffaa4da6da48ed9d8031da",
				Tag: "0.4.0",
			},
		},
	},
	Image{
		Name: "docker.elastic.co/elasticsearch/elasticsearch-oss",
		Tags: []Tag{
			Tag{
				Sha: "e86f0491edab3d0fd20a1aa0218fda795e12f20e7fe07a454101d7446b29522d",
				Tag: "6.1.1",
			},
		},
	},
	Image{
		Name: "docker.elastic.co/kibana/kibana-oss",
		Tags: []Tag{
			// via https://www.elastic.co/guide/en/kibana/6.1/_pulling_the_image.html
			Tag{
				Sha: "f9addd642b184a81daa77c4301a800009aa714296220549ad1c61a22ca9bb8d3",
				Tag: "6.1.4",
			},
		},
	},
	Image{
		Name: "fluent/fluent-bit",
		Tags: []Tag{
			Tag{
				Sha: "00ae5afa4e113352f6f7db1377780e6e23a09a157d287f7c09ea8605e0e4492f",
				Tag: "0.13.7",
			},
		},
	},
	Image{
		Name: "fluent/fluent-bit-0.13-dev",
		Tags: []Tag{
			Tag{
				Sha: "0ae482ad3e8d951a66090b968389f2a11b0b16b8a208b5edb4ccb0ca5800b90d",
				Tag: "0.18",
			},
		},
	},
	Image{
		Name: "gcr.io/google_containers/defaultbackend",
		Tags: []Tag{
			Tag{
				Sha: "ee3aa1187023d0197e3277833f19d9ef7df26cee805fef32663e06c7412239f9",
				Tag: "1.0",
			},
			Tag{
				Sha: "a64c8ed5df00c9f238ecdeb28eb4ed226faace573695e290a99d92d503593e87",
				Tag: "1.2",
			},
		},
	},
	Image{
		// see https://github.com/kubernetes/kubernetes/blob/master/cluster/addons/fluentd-elasticsearch/fluentd-es-ds.yaml
		Name: "gcr.io/google_containers/fluentd-elasticsearch",
		Tags: []Tag{
			// from k8s.gcr.io/fluentd-elasticsearch:v2.0.4
			Tag{
				Sha: "b8c94527b489fb61d3d81ce5ad7f3ddbb7be71e9620a3a36e2bede2f2e487d73",
				Tag: "v2.0.4",
			},
		},
	},
	Image{
		Name: "gcr.io/google_containers/hyperkube",
		Tags: []Tag{
			Tag{
				Sha: "ac62a2f03cb254ca31654fe49492bf21d2282f8c1cbffa7c91518af0fb5f3c73",
				Tag: "v1.11.3",
			},
		},
	},
	Image{
		Name: "gcr.io/google_containers/kube-state-metrics",
		Tags: []Tag{
			Tag{
				Sha: "953a3b6bf0046333c656fcfa2fc3a08f4055dc3fbd5b1dcdcdf865a2534db526",
				Tag: "v1.2.0",
			},
		},
	},
	Image{
		Name: "gcr.io/google_containers/metrics-server-amd64",
		Tags: []Tag{
			Tag{
				Sha: "6f4a027083d92fd0f28d1aca83364e376e440625ca9a403f1d2d50adaa298d88",
				Tag: "v0.3.0",
			},
		},
	},
	Image{
		Name: "gcr.io/google_containers/nginx-ingress-controller",
		Tags: []Tag{
			Tag{
				Sha: "03fd8fc46018d09b4050d4daaf50bff73c80936994b374319ed33cbb2c1684f4",
				Tag: "0.9.0-beta.11",
			},
		},
	},
	Image{
		Name: "gcr.io/google_containers/pause-amd64",
		Tags: []Tag{
			Tag{
				Sha: "59eec8837a4d942cc19a52b8c09ea75121acc38114a2c68b98983ce9356b8610",
				Tag: "3.1",
			},
		},
	},
	Image{
		Name: "gcr.io/heptio-images/kube-conformance",
		Tags: []Tag{
			Tag{
				Sha: "4b63b91265ed0e4a986db9ce4bab22f590d773108713f806180990bd0e0c0806",
				Tag: "v1.11",
			},
		},
	},
	Image{
		Name: "gcr.io/heptio-images/sonobuoy",
		Tags: []Tag{
			Tag{
				Sha: "9f2a352b44143c8c4dc72ea2df07d1b3c9d37e45a2ebcfa72c048cca17b9d6eb",
				Tag: "v0.10.0",
			},
		},
	},
	Image{
		Name: "gcr.io/kubernetes-helm/tiller",
		Tags: []Tag{
			Tag{
				Sha: "9b373c71ea2dfdb7d42a6c6dada769cf93be682df7cfabb717748bdaef27d10a",
				Tag: "v2.8.2",
			},
		},
	},
	Image{
		Name: "gfkse/oauth2_proxy",
		Tags: []Tag{
			Tag{
				Sha: "4bba1afcd3af85b550b42647e92b3fab36448c75e1af611a65644f77f4dde314",
				Tag: "kubernetes-3-ca-certs",
			},
		},
	},
	Image{
		Name: "golang",
		Tags: []Tag{
			Tag{
				Sha: "80fa52964480e2e32dc2a82a04dfd41797787383a547f2498baaa87584bae672",
				Tag: "1.10.3",
			},
		},
	},
	Image{
		Name: "grafana/grafana",
		Tags: []Tag{
			Tag{
				Sha: "86d8159672941cafb67f2d1a134d119ba7f7aa933e801e98bf6e18bc94394203",
				Tag: "5.2.3",
			},
		},
	},
	Image{
		Name: "jetstack/kube-lego",
		Tags: []Tag{
			Tag{
				Sha: "10e19105596be0ee03b2a38879dd0f1e72bff26230961c22796136defcc3c7cb",
				Tag: "0.1.5",
			},
		},
	},
	Image{
		Name: "nginx",
		Tags: []Tag{
			Tag{
				Sha: "0fb320e2a1b1620b4905facb3447e3d84ad36da0b2c8aa8fe3a5a81d1187b884",
				Tag: "1.13.12",
			},
		},
	},
	Image{
		Name: "prom/cloudwatch-exporter",
		Tags: []Tag{
			Tag{
				Sha: "7ce6d17ad3360ed5d5dddc57ebf13dc7d676900c40a22b3144a8a17af352c399",
				Tag: "0.4",
			},
		},
	},
	Image{
		Name: "prom/prometheus",
		Tags: []Tag{
			Tag{
				Sha: "129e16b08818a47259d972767fd834d84fb70ca11b423cc9976c9bce9b40c58f",
				Tag: "v2.2.1",
			},
		},
	},
	Image{
		Name: "quay.io/calico/cni",
		Tags: []Tag{
			Tag{
				Sha: "b45dab776f42c684fe378cea1482b5bc691fe500f29d00a50f749306a178bcdd",
				Tag: "v3.2.0",
			},
		},
	},
	Image{
		Name: "quay.io/calico/kube-controllers",
		Tags: []Tag{
			Tag{
				Sha: "6fc737c578da404199a891087c3bb03e9398905670d0a809dc21f8a1f11bafef",
				Tag: "v3.2.0",
			},
		},
	},
	Image{
		Name: "quay.io/calico/node",
		Tags: []Tag{
			Tag{
				Sha: "dd9ae045359dd9e22b66c2d2bedf8e173e3cf8ca8893a15d8543b96fa2cae072",
				Tag: "v3.2.0",
			},
		},
	},
	Image{
		Name: "quay.io/calico/typha",
		Tags: []Tag{
			Tag{
				Sha: "f7cb43c1a1e4398e1b43e8d09124a0c722502a06d0433a3b2693634c8d1b6300",
				Tag: "v3.2.0",
			},
		},
	},
	Image{
		Name: "quay.io/coreos/etcd",
		Tags: []Tag{
			Tag{
				Sha: "05c576849d6af2d30551a28e3dd7ba480a94b20dc48f5ac0a56ddf7e4f2c2269",
				Tag: "v3.3.8",
			},
		},
	},
	Image{
		Name: "quay.io/coreos/etcd-operator",
		Tags: []Tag{
			Tag{
				Sha: "2a1ff56062861e3eaf216899e6e73fdff311e5842d2446223924a9cc69f2cc69",
				Tag: "v0.3.2",
			},
		},
	},
	Image{
		Name: "quay.io/coreos/flannel",
		Tags: []Tag{
			Tag{
				Sha: "88f2b4d96fae34bfff3d46293f7f18d1f9f3ca026b4a4d288f28347fcb6580ac",
				Tag: "v0.10.0-amd64",
			},
		},
	},
	Image{
		Name: "quay.io/coreos/kube-state-metrics",
		Tags: []Tag{
			Tag{
				Sha: "fa2e6d33183755f924f05744c282386f38e962160f66ad0b6a8a24a36884fb9a",
				Tag: "v1.3.1",
			},
		},
	},
	Image{
		Name: "quay.io/giantswarm/docker-kubectl",
		Tags: []Tag{
			Tag{
				Sha: "995bd3fee6899569d5d9ab77948f25d6bc2e9a95efa988de37c6b8c3095ac819",
				Tag: "8cabd75bacbcdad7ac5d85efc3ca90c2fabf023b",
			},
		},
	},
	Image{
		Name: "quay.io/giantswarm/k8s-migrator",
		Tags: []Tag{
			Tag{
				Sha: "9259af85d5a7f395feab8162b4d47a79f6b53a897568bc3a9d9d4908a8ac0de2",
				Tag: "4a4c553280d99b28cb0114797ba59aa380e808b1",
			},
		},
	},
	Image{
		Name: "quay.io/giantswarm/k8s-setup-network-environment",
		Tags: []Tag{
			Tag{
				Sha: "e337d03e569e53b246f4dea91359efbabe7b3ddc78878e1875d0c7aaf0e17fd5",
				Tag: "1f4ffc52095ac368847ce3428ea99b257003d9b9",
			},
		},
	},
	Image{
		Name: "quay.io/jetstack/cert-manager-controller",
		Tags: []Tag{
			Tag{
				Sha: "61546385c284af5620ac1e861943e73164a6ac37ec76520ef43be2ec2bd769fb",
				Tag: "v0.2.5",
			},
			Tag{
				Sha: "3ff035464d9349f5b1d23a2fef9c9b8419026671506a5be5cbae6e958ac46802",
				Tag: "v0.4.0",
			},
		},
	},
	Image{
		Name: "quay.io/jetstack/cert-manager-ingress-shim",
		Tags: []Tag{
			Tag{
				Sha: "544b8602ee566d7ca22aa9e76a92dde4c2ca8dab642f75ea3a4b0a577193632a",
				Tag: "v0.2.5",
			},
		},
	},
	Image{
		Name: "quay.io/kubernetes-ingress-controller/nginx-ingress-controller",
		Tags: []Tag{
			Tag{
				Sha: "36523a0b8b35b082211caa2bebb95c43578f85a51c03a28599b39a13b27965cb",
				Tag: "0.12.0",
			},
			Tag{
				Sha: "7b79e1bc6437e6376dadf558e012adde6395bb28dee4a38ce08c7e5c9f220178",
				Tag: "0.15.0",
			},
			Tag{
				Sha: "84ed5290a91c53b4c224a724a251347dfc8cf2bca4be06e32f642c396eb02429",
				Tag: "0.16.2",
			},
		},
	},
	Image{
		Name: "quay.io/prometheus/alertmanager",
		Tags: []Tag{
			Tag{
				Sha: "2843872cb4cd20da5b75286a5a2ac25a17ec1ae81738ba5f75d5ee8794b82eaf",
				Tag: "v0.7.1",
			},
			Tag{
				Sha: "0ed4a8f776c5570b9e8152a670d3087a73164b20476a6a94768468759fbb5ad8",
				Tag: "v0.15.0",
			},
		},
	},
	Image{
		Name: "quay.io/prometheus/node-exporter",
		Tags: []Tag{
			Tag{
				Sha: "fc004c4a3d1096d5a0f144b1093daa9257a573ce1fde5a9b8511e59a7080a1bb",
				Tag: "v0.15.1",
			},
		},
	},
	Image{
		Name: "redis",
		Tags: []Tag{
			Tag{
				Sha: "002a1870fa2ffd11dbd7438527a2c17f794f6962f5d3a4f048f848963ab954a8",
				Tag: "3.2.11-alpine",
			},
		},
	},
	Image{
		Name: "registry.opensource.zalan.do/teapot/external-dns",
		Tags: []Tag{
			Tag{
				Sha: "0c6e7c59bd204db7dd13c98e8ec6b1af5b9b102f0badc091d50683458570c6c6",
				Tag: "v0.5.2-3-g6c05028",
			},
		},
	},
	Image{
		Name: "sysdig/falco",
		Tags: []Tag{
			Tag{
				Sha: "3533054a0543e4a2eb4f4c61d68d5d3a39e5886a65d07f011d45044a5c470ccd",
				Tag: "0.11.1",
			},
		},
	},
	Image{
		Name: "vault",
		Tags: []Tag{
			Tag{
				Sha: "8d2813d4fbc145d867218b60e13b29941edf60d1c0929964db42879a1aacc889",
				Tag: "0.10.1",
			},
			Tag{
				Sha: "366eddc65d233c7b43269ba80e27aeb1269837beadd011c8d7b3daa999cce70a",
				Tag: "0.10.3",
			},
		},
	},
}

func ImageName(organisation string, image string) string {
	parts := strings.Split(image, "/")

	return fmt.Sprintf("%v/%v", organisation, parts[len(parts)-1])
}

func RetaggedName(registry, organisation string, image string) string {
	parts := strings.Split(image, "/")

	return fmt.Sprintf("%v/%v/%v", registry, organisation, parts[len(parts)-1])
}

func ImageWithTag(image, tag string) string {
	return fmt.Sprintf("%v:%v", image, tag)
}

func ShaName(imageName, sha string) string {
	return fmt.Sprintf("%v@sha256:%v", imageName, sha)
}
