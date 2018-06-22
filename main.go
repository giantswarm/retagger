package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
)

func main() {
	registry := os.Getenv("REGISTRY")
	if registry == "" {
		log.Fatalf("could not get registry")
	}

	registryOrganisation := os.Getenv("REGISTRY_ORGANISATION")
	if registryOrganisation == "" {
		log.Fatalf("could not get registry organisation")
	}

	registryUsername := os.Getenv("REGISTRY_USERNAME")
	if registryUsername == "" {
		log.Fatalf("could not get registry username")
	}

	registryPassword := os.Getenv("REGISTRY_PASSWORD")
	if registryPassword == "" {
		log.Fatalf("could not get registry password")
	}

	quayToken := os.Getenv("QUAY_TOKEN")
	if quayToken == "" {
		log.Fatalf("could not get quay token")
	}

	login := exec.Command("docker", "login", "-u", registryUsername, "-p", registryPassword, registry)
	if err := Run(login); err != nil {
		log.Fatalf("could not login to registry: %v", err)
	}

	client := &http.Client{}

	for _, image := range Images {
		for _, tag := range image.Tags {
			log.Printf("managing: %v, %v, %v", image.Name, tag.Sha, tag.Tag)

			url := fmt.Sprintf("https://quay.io/api/v1/repository/%s/tag/%s/images", ImageName(registryOrganisation, image), tag.Tag)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				log.Fatalf("could not create request: %v", err)
			}
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", quayToken))

			res, err := client.Do(req)
			if err != nil {
				log.Fatalf("could not check if image retagged: %v", err)
			}
			switch res.StatusCode {
			case http.StatusOK:
				log.Printf("retagged image already exists, skipping")
				continue
			case http.StatusNotFound:
				log.Printf("retagged image does not exist")
			default:
				log.Printf("could not check retag status: %v", res.StatusCode)
			}

			shaName := ShaName(image.Name, tag.Sha)

			log.Printf("pulling original image")
			pullOriginal := exec.Command("docker", "pull", shaName)
			if err := Run(pullOriginal); err != nil {
				log.Fatalf("could not pull image: %v", err)
			}

			retaggedName := RetaggedName(registry, registryOrganisation, image)
			retaggedNameWithTag := ImageWithTag(retaggedName, tag.Tag)

			log.Printf("retagging image")
			retag := exec.Command("docker", "tag", shaName, retaggedNameWithTag)
			if err := Run(retag); err != nil {
				log.Fatalf("could not retag image: %v", err)
			}

			log.Printf("pushing image")
			push := exec.Command("docker", "push", retaggedNameWithTag)
			if err := Run(push); err != nil {
				log.Fatalf("could not push image: %v", err)
			}
		}
	}
}
