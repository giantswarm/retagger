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
		Name: "coredns/coredns",
		Tags: []Tag{
			Tag{
				Sha: "399cc5b2e2f0d599ef22f43aab52492e88b4f0fd69da9b10545e95a4253c86ce",
				Tag: "1.1.1",
			},
		},
	},
	Image{
		Name: "quay.io/calico/node",
		Tags: []Tag{
			Tag{
				Sha: "f5c53992c20d15d5976f9ee9ac776d63de833c5abd26e127fc638c75e9e2f5d4",
				Tag: "v3.0.8",
			},
		},
	},
	Image{
		Name: "quay.io/calico/cni",
		Tags: []Tag{
			Tag{
				Sha: "91f3b7a4a1004269ed09f9d856395046a56e9984c782ca0b037ad88c1b90c11e",
				Tag: "v2.0.6",
			},
		},
	},
	Image{
		Name: "quay.io/calico/kube-controllers",
		Tags: []Tag{
			Tag{
				Sha: "378cc28e1b588b0b7e68bee4432c3fc76d6b718dfffbbfec01434fbffbf18188",
				Tag: "v2.0.5",
			},
		},
	},
	Image{
		Name: "quay.io/calico/typha",
		Tags: []Tag{
			Tag{
				Sha: "35334ae788a460f62b1668470a359e3affd7cbbb2c6b6782c560d691754d7686",
				Tag: "v0.6.6",
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
		Name: "quay.io/coreos/etcd",
		Tags: []Tag{
			Tag{
				Sha: "05c576849d6af2d30551a28e3dd7ba480a94b20dc48f5ac0a56ddf7e4f2c2269",
				Tag: "v3.3.8",
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
		Name: "alpine",
		Tags: []Tag{
			Tag{
				Sha: "7df6db5aa61ae9480f52f0b3a06a140ab98d427f86d8d5de0bedab9b8df6b1c0",
				Tag: "3.7",
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
				Sha: "ca6d734e11a63be10fb9d270a7f248cfb273bb02724b79acd7b1415a582ef290",
				Tag: "0.13.1",
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
		// see https://github.com/kubernetes/kubernetes/blob/master/cluster/addons/fluentd-elasticsearch/fluentd-es-ds.yaml
		Name: "gcr.io/google-containers/fluentd-elasticsearch",
		Tags: []Tag{
			// from k8s.gcr.io/fluentd-elasticsearch:v2.0.4
			Tag{
				Sha: "b8c94527b489fb61d3d81ce5ad7f3ddbb7be71e9620a3a36e2bede2f2e487d73",
				Tag: "v2.0.4",
			},
		},
	},
	Image{
		Name: "gcr.io/google_containers/defaultbackend",
		Tags: []Tag{
			Tag{
				Sha: "a64c8ed5df00c9f238ecdeb28eb4ed226faace573695e290a99d92d503593e87",
				Tag: "1.2",
			},
		},
	},
	Image{
		Name: "gcr.io/google_containers/hyperkube",
		Tags: []Tag{
			Tag{
				Sha: "9ef46393310efe17746df1349a31ae7b9dd998655253cf392aa18b9888a8756c",
				Tag: "v1.11.0",
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
		Name: "quay.io/coreos/kube-state-metrics",
		Tags: []Tag{
			Tag{
				Sha: "fa2e6d33183755f924f05744c282386f38e962160f66ad0b6a8a24a36884fb9a",
				Tag: "v1.3.1",
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
				Sha: "104f434d47c8830be44560edc012c31114a104301cdb81bad6e8abc52a2304f9",
				Tag: "5.2.1",
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
		Name: "quay.io/kubernetes-ingress-controller/nginx-ingress-controller",
		Tags: []Tag{
			Tag{
				Sha: "7b79e1bc6437e6376dadf558e012adde6395bb28dee4a38ce08c7e5c9f220178",
				Tag: "0.15.0",
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
		Name: "quay.io/jetstack/cert-manager-controller",
		Tags: []Tag{
			Tag{
				Sha: "61546385c284af5620ac1e861943e73164a6ac37ec76520ef43be2ec2bd769fb",
				Tag: "v0.2.5",
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
		Name: "redis",
		Tags: []Tag{
			Tag{
				Sha: "002a1870fa2ffd11dbd7438527a2c17f794f6962f5d3a4f048f848963ab954a8",
				Tag: "3.2.11-alpine",
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
		Name: "gcr.io/kubernetes-helm/tiller",
		Tags: []Tag{
			Tag{
				Sha: "9b373c71ea2dfdb7d42a6c6dada769cf93be682df7cfabb717748bdaef27d10a",
				Tag: "v2.8.2",
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
