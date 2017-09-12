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
		Name: "gcr.io/google_containers/kube-state-metrics",
		Tags: []Tag{
			Tag{
				Sha: "e913a24b0a0a89e23968d5e3fbf99501d17c04011fb54b24df0aca6bea232022",
				Tag: "v0.5.0",
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
				Sha: "51f988b1a86f528c2e40681175088b5312b96bba9bea0f05bdb7ab504425c52d",
				Tag: "1.8.3",
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
				Sha: "64d1130dd5f64fca6e7959a6991c64ca655ee9fb9c897ac25a43fef585f045a5",
				Tag: "v1.7.5_coreos.0",
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
