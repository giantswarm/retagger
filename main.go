package main

import (
	"log"
	"os"
	"os/exec"
)

func main() {
	registry := os.Getenv("REGISTRY")
	if registry == "" {
		log.Fatalf("could not get registry\n")
	}

	registryOrganisation := os.Getenv("REGISTRY_ORGANISATION")
	if registryOrganisation == "" {
		log.Fatalf("could not get registry organisation\n")
	}

	registryUsername := os.Getenv("REGISTRY_USERNAME")
	if registryUsername == "" {
		log.Fatalf("could not get registry username\n")
	}

	registryPassword := os.Getenv("REGISTRY_PASSWORD")
	if registryPassword == "" {
		log.Fatalf("could not get registry password\n")
	}

	login := exec.Command("docker", "login", "--email=\" \"", "-u", registryUsername, "-p", registryPassword, registry)
	if err := Run(login); err != nil {
		log.Fatalf("could not login to registry: %v\n", err)
	}

	for _, image := range Images {
		for _, tag := range image.Tags {
			log.Printf("managing: %v, %v, %v\n", image.Name, tag.Sha, tag.Tag)

			retaggedName := RetaggedName(registry, registryOrganisation, image)
			shaName := ShaName(image.Name, tag.Sha)

			retaggedNameWithTag := ImageWithTag(retaggedName, tag.Tag)

			log.Printf("checking if image has already been retagged")
			pullRetag := exec.Command("docker", "pull", retaggedNameWithTag)
			if err := Run(pullRetag); err == nil {
				log.Printf("retagged image already exists, skipping\n")
				continue
			} else {
				log.Printf("retagged image probably does not exist: %v\n", err)
			}

			log.Printf("pulling original image")
			pullOriginal := exec.Command("docker", "pull", shaName)
			if err := Run(pullOriginal); err != nil {
				log.Fatalf("could not pull image: %v\n", err)
			}

			log.Printf("retagging image")
			retag := exec.Command("docker", "tag", shaName, retaggedNameWithTag)
			if err := Run(retag); err != nil {
				log.Fatalf("could not retag image: %v\n", err)
			}

			log.Printf("pushing image")
			push := exec.Command("docker", "push", retaggedNameWithTag)
			if err := Run(push); err != nil {
				log.Fatalf("could not push image: %v\n", err)
			}
		}
	}
}
