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
		},
	},
	Image{
		Name: "fluent/fluentd-kubernetes-daemonset",
		Tags: []Tag{
			Tag{
				Sha: "721d539e5edc941566c2422c7b6f3838fc0c96543b1853eb4d9acb3c70d5bc6b",
				Tag: "v0.12-alpine-elasticsearch",
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
				Sha: "39b49865ac5faa4787af92a5a7d00161e467005047976ce8ab9da284b5acb9da",
				Tag: "v1.9.0-beta.1",
			},
			Tag{
				Sha: "3e7e58f6bc4f28c3ba36e88ed2220695922c0304db802e72b20b987f1c89b126",
				Tag: "v1.9.0",
			},
			Tag{
				Sha: "cd3430a6fbbffa6f52f7233c08237477ef374424baf45df8e33c5dc9b1f87177",
				Tag: "v1.9.2",
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
		},
	},
	Image{
		Name: "grafana/grafana",
		Tags: []Tag{
			Tag{
				Sha: "2b08adb787f0b6c30a6cb13c46fdbae90e8f98d8570bdf468efd9d5ea4974b1a",
				Tag: "4.4.1",
			},
			Tag{
				Sha: "03f29c009d30101a1be5500212788099b0b5fc33cbcf20d8c0017b0d8c613683",
				Tag: "4.6.2",
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
				Sha: "1b401bf0c30bada9a539389c3be652b58fe38463361edf488e6543c8761d4970",
				Tag: "v0.9.0-amd64",
			},
		},
	},
	Image{
		Name: "quay.io/coreos/hyperkube",
		Tags: []Tag{
			Tag{
				Sha: "297f45919160ea076831cd067833ad3b64c789fcb3491016822e6f867d16dcd5",
				Tag: "v1.6.4_coreos.0",
			},
			Tag{
				Sha: "b26b7a598382e9db0623dd999f5506746a076b866f12188010f6c07fef26e4da",
				Tag: "v1.6.7_coreos.0",
			},
			Tag{
				Sha: "cc57bd170e562ab699b6dd2244f2683018214c8228591dc641cffce0f0f92037",
				Tag: "v1.7.1_coreos.0",
			},
			Tag{
				Sha: "44472a474d3e150bb9516dcd1380275d10ba5e57065347dd8f8aabfe64db9457",
				Tag: "v1.7.3_coreos.0",
			},
			Tag{
				Sha: "95d545f94e011a01517b220367c4c02a4559a72acc951737af22e81b98f78377",
				Tag: "v1.7.5_coreos.0",
			},
			Tag{
				Sha: "3b9d27edc492a5b37cb415d172a6df5885a17d8d68b75a893531a76b1431adfa",
				Tag: "v1.7.11_coreos.0",
			},
			Tag{
				Sha: "8755aefadd070df7b26e49ce2998209547eca7bd4054e5dbb434615407374753",
				Tag: "v1.8.0_coreos.0",
			},
			Tag{
				Sha: "131eb9d7665d3cc2f909a71e58ec53f27ef88e4fd6f2dda843a7bf4a043078e4",
				Tag: "v1.8.1_coreos.0",
			},
			Tag{
				Sha: "0d659da66e80396f7fbe1352878d1013b30f15a728924ee3eaff1cc79a38b825",
				Tag: "v1.8.4_coreos.0",
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
