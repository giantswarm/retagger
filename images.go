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
				Sha: "d291f8b87eab26845a0c4605df4194924806712c4f624b9a9ddfc9d382b3ddbd",
				Tag: "1.0.4",
			},
			Tag{
				Sha: "a01b8b7465f8ce5326e1589c7bbed1b99322804c472872a03edb60fbedaaa6f6",
				Tag: "1.0.5",
			},
			Tag{
				Sha: "ddcc984408e779e0baa92aab754b8988ba57d3ca3478837e186c90350624374b",
				Tag: "1.0.6",
			},
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
				Sha: "77f21d72021ed7ccb0fa6fac2ce0a466536ead17bfb7e3e28afcf93a31d6e896",
				Tag: "v3.0.5",
			},
		},
	},
	Image{
		Name: "quay.io/calico/cni",
		Tags: []Tag{
			Tag{
				Sha: "a5de754aeab76601fd7bbe0ff1622ab49060a13c25f5d43ae15ccbf1fe46fef7",
				Tag: "v2.0.4",
			},
		},
	},
	Image{
		Name: "quay.io/calico/kube-controllers",
		Tags: []Tag{
			Tag{
				Sha: "58ddfd9de2e91b160440c3ede51d3cb7c0250450d047d9ba34d874c59b710619",
				Tag: "v2.0.3",
			},
		},
	},
	Image{
		Name: "quay.io/calico/typha",
		Tags: []Tag{
			Tag{
				Sha: "e5f7143147254b9ddfa3620aeb0151ae10d29bcd2e62d83d20b25fe54a5fcbdc",
				Tag: "v0.6.3",
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
		Name: "quay.io/giantswarm/etcd",
		Tags: []Tag{
			Tag{
				Sha: "454e69370d87554dcb4272833b8f07ce1b5d457caa153bda4070b76d89a1cc97",
				Tag: "v3.3.1",
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
			Tag{
				Sha: "306af26503cd751440f55428c2d7c14d09105ed125e5a9fc0b8d29206042053e",
				Tag: "6.1.1",
			},
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
				Sha: "3101144a0ae219e60a944ee3c5939a95fd18ea18a8ebb1fdfd3126bc6513a1cd",
				Tag: "0.13",
			},
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
				Sha: "f41081bed4870c910df045e489c067fd05fd09fb06de6ceece39fba713fa185e",
				Tag: "0.16",
			},
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
			// from k8s.gcr.io/fluentd-elasticsearch:v2.0.3
			Tag{
				Sha: "0bee097b7f7f23c2fc79a1ad39beabe97832b6ceb8e03e12408f16e99ac56d3a",
				Tag: "v2.0.3",
			},
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
		Name: "gcr.io/google_containers/hyperkube",
		Tags: []Tag{
			Tag{
				Sha: "a31961a719a1d0ade89149a6a8db5181cbef461baa6ef049681c31c0e48d9f1e",
				Tag: "v1.9.5",
			},
			Tag{
				Sha: "ef1169d8ed08a69fb31a9aaafdceced4200dc9bea0ec03751f3638ae28646f51",
				Tag: "v1.9.6",
			},
			Tag{
				Sha: "badd2b1da29d4d530b10b920f64bf66a1b41150db46c3c99b49d56f3f18a82db",
				Tag: "v1.10.2",
			},
		},
	},
	Image{
		Name: "gcr.io/google_containers/kube-state-metrics",
		Tags: []Tag{
			Tag{
				Sha: "e913a24b0a0a89e23968d5e3fbf99501d17c04011fb54b24df0aca6bea232022",
				Tag: "v0.5.0",
			},
			Tag{
				Sha: "b8b536771d5c23a9344c90662b2ca9ba00421e050ae593264bc51803470a2526",
				Tag: "v1.0.1",
			},
			Tag{
				Sha: "53416b3d560a1b821b7e302460a387fef887ce72206c3ccbf82fd9e2d1f71fd9",
				Tag: "v1.1.0",
			},
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
				Sha: "f053462579a86ff5b14941635b659089dae31e207472b72551d5f7339b143a54",
				Tag: "v1.3.0",
			},
			Tag{
				Sha: "3054e18f76bb96de6faba2001c212c40766e517b56b424dacafe4a97eada5dda",
				Tag: "v1.3.1",
			},
		},
	},
	Image{
		Name: "gcr.io/google_containers/nginx-ingress-controller",
		Tags: []Tag{
			Tag{
				Sha: "995427304f514ac1b70b2c74ee3c6d4d4ea687fb2dc63a1816be15e41cf0e063",
				Tag: "0.9.0-beta.3",
			},
			Tag{
				Sha: "897b86cd624e3d5b6e69c3b0336f10726ac6314736bef96d6eedec6b6eb7712b",
				Tag: "0.9.0-beta.7",
			},
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
				Sha: "0bc4b605f127ebcda9c96d8b0411780e8dcc03ee695c5f9cdf6298f9977b8ca8",
				Tag: "1.9.0",
			},
			Tag{
				Sha: "f755ff87e4b7a5f597a4ed5f0a1013dd5550f21615ce71312936dc36988cb274",
				Tag: "1.9.1",
			},
			Tag{
				Sha: "cd78c0227f4fbc7fa820a2b11c1ef4b4880cc047687d63f0bd0e7e7e363589ca",
				Tag: "1.10.0",
			},
		},
	},
	Image{
		Name: "grafana/grafana",
		Tags: []Tag{
			Tag{
				Sha: "6397aafb899ef7a9ca61c2ef80863dbebce504620b044954d80203e0b8c1ada4",
				Tag: "4.6.3",
			},
		},
	},
	Image{
		Name: "nginx",
		Tags: []Tag{
			Tag{
				Sha: "5269659b61c4f19a3528a9c22f9fa8f4003e186d6cb528d21e411578d1e16bdb",
				Tag: "1.12.2",
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
				Sha: "33c41643b9f3504ff5381a306fad5ca90269cafdcc1495c43cade31f462f3933",
				Tag: "v1.6.2",
			},
			Tag{
				Sha: "7b4428a9658dd7f0ff826ecbd20eb2ea653852ef580c13b2087e5476a73d4b1f",
				Tag: "v1.6.3",
			},
			Tag{
				Sha: "7b987901dbc44d17a88e7bda42dbbbb743c161e3152662959acd9f35aeefb9a3",
				Tag: "v2.1.0",
			},
		},
	},
	Image{
		Name: "quay.io/coreos/etcd-operator",
		Tags: []Tag{
			Tag{
				Sha: "efa735007e3c989c99dc76a1c8adcd1ea492b02804669dc9d95bb59706d96c89",
				Tag: "v0.1.0",
			},
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
				Sha: "cee8b233374e489b324a415f169b674aedfd4c8e7f400d22dd96a08dabca4316",
				Tag: "0.10.0",
			},
			Tag{
				Sha: "4f0cabc2f810c7eaec9fe3002bef0666e15309e30156ef780efeaa5bae1a311f",
				Tag: "0.10.1",
			},
			Tag{
				Sha: "20fb21709d0fa52c5f873ba68d464e04981d0cedf07e900f8a9def6874cf4cee",
				Tag: "0.10.2",
			},
			Tag{
				Sha: "885b65cec9e58c4829be447af4b0b00ecc40c09e0b9e9f662374f308e536c217",
				Tag: "0.11.0",
			},
			Tag{
				Sha: "36523a0b8b35b082211caa2bebb95c43578f85a51c03a28599b39a13b27965cb",
				Tag: "0.12.0",
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
				Sha: "b376a1b4f6734ed610b448603bc0560106c2e601471b49f72dda5bd40da095dd",
				Tag: "v0.14.0",
			},
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
				Sha: "848b4fd76a5dacb56988af810a6e86719e313cf4e1186f3d3050384686dbc120",
				Tag: "3.2.10",
			},
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
				Sha: "87f3caef34d1da704e4ba3cfa1f0ce03469c5bd4bc2b18ea728ce76d5d494f08",
				Tag: "0.9.3",
			},
			Tag{
				Sha: "8d2813d4fbc145d867218b60e13b29941edf60d1c0929964db42879a1aacc889",
				Tag: "0.10.1",
			},
		},
	},
}

func RetaggedName(registry, organisation string, image Image) string {
	parts := strings.Split(image.Name, "/")

	return fmt.Sprintf("%v/%v/%v", registry, organisation, parts[len(parts)-1])
}

func ImageWithTag(image, tag string) string {
	return fmt.Sprintf("%v:%v", image, tag)
}

func ShaName(imageName, sha string) string {
	return fmt.Sprintf("%v@sha256:%v", imageName, sha)
}
