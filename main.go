package main

import (
	"log"
	"net/http"
	"os"
	"os/exec"
)

func main() {
	c := &RegistryConfig{
		Client: &http.Client{},

		Host:         os.Getenv("REGISTRY"),
		Organisation: os.Getenv("REGISTRY_ORGANISATION"),
		Password:     os.Getenv("REGISTRY_PASSWORD"),
		Username:     os.Getenv("REGISTRY_USERNAME"),
	}

	registry, err := NewRegistry(c)
	if err != nil {
		log.Fatalf("could not create registry %v", err)
	}

	err = registry.Login()
	if err != nil {
		log.Fatalf("could not login to registry %v", err)
	}

	for _, image := range Images {
		for _, tag := range image.Tags {
			imageName := image.Name
			if image.OverrideRepoName != "" {
				log.Printf("Override Name specified. Using %s as mirrored image name", image.OverrideRepoName)
				imageName = image.OverrideRepoName
			}
			log.Printf("managing: %v, %v, %v", imageName, tag.Sha, tag.Tag)

			ok, err := registry.CheckImageTagExists(imageName, tag.Tag)
			if ok {
				log.Printf("retagged image already exists, skipping")
				continue
			} else if err != nil {
				log.Fatalf("could not check image %q and tag %q: %v", imageName, tag.Tag, err)
			} else {
				log.Printf("retagged image does not exist")
			}

			shaName := ShaName(image.Name, tag.Sha)

			log.Printf("pulling original image")
			pullOriginal := exec.Command("docker", "pull", shaName)
			if err := Run(pullOriginal); err != nil {
				log.Fatalf("could not pull image: %v", err)
			}

			retaggedNameWithTag, err := registry.Retag(imageName, shaName, tag.Tag)
			if err != nil {
				log.Fatalf("could not retag image: %v", err)
			}

			log.Printf("pushing retagged image")
			push := exec.Command("docker", "push", retaggedNameWithTag)
			if err := Run(push); err != nil {
				log.Fatalf("could not push image: %v", err)
			}

			if len(tag.DockerfileOptions) != 0 {
				rebuildedImageTag, err := registry.Rebuild(imageName, tag.Tag, tag.DockerfileOptions)
				if err != nil {
					log.Fatalf("could not rebuild image: %v", err)
				}

				log.Printf("pushing rebuilded image")
				push := exec.Command("docker", "push", rebuildedImageTag)
				if err := Run(push); err != nil {
					log.Fatalf("could not push image: %v", err)
				}

			}
		}
	}
}
