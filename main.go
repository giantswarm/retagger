package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
)

var configFile = flag.String("f", "images.yaml", "retagger config (images.yaml) file to use.")

func main() {
	flag.Parse()

	if *configFile == "" {
		log.Fatalf("-f flag (config file) should not be empty")
	}

	err := InitImages(*configFile)
	if err != nil {
		log.Fatalf("could not read retagger config %v", err)
	}

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

			for _, customImage := range tag.CustomImages {
				ok, err := registry.CheckImageTagExists(imageName, tag.Tag)
				if ok {
					log.Printf("rebuilt image %q with tag %q already exists, skipping", imageName, fmt.Sprintf("%s-%s", tag.Tag, customImage.TagSuffix))
					continue
				} else if err != nil {
					log.Fatalf("could not check image %q and tag %q: %v", imageName, tag.Tag, err)
				} else {
					log.Printf("rebuilt image %q with tag %q does not exists", imageName, fmt.Sprintf("%s-%s", tag.Tag, customImage.TagSuffix))
				}
				rebuiltImageTag, err := registry.Rebuild(imageName, tag.Tag, customImage)
				if err != nil {
					log.Fatalf("could not rebuild image: %v", err)
				}

				log.Printf("pushing rebuilt custom image %s-%s", tag.Tag, customImage.TagSuffix)
				push := exec.Command("docker", "push", rebuiltImageTag)
				if err := Run(push); err != nil {
					log.Fatalf("could not push image: %v", err)
				}
			}

			ok, err := registry.CheckImageTagExists(imageName, tag.Tag)
			if ok {
				log.Printf("retagged image %q with tag %q already exists, skipping", imageName, tag.Tag)
				continue
			} else if err != nil {
				log.Fatalf("could not check image %q and tag %q: %v", imageName, tag.Tag, err)
			} else {
				log.Printf("retagged image %q with tag %q does not exist", imageName, tag.Tag)
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
		}
	}
}
