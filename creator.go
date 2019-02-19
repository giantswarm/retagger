package main

import (
	"fmt"
	"log"

	"github.com/giantswarm/retagger/aliyun/cr"
	"github.com/giantswarm/retagger/quay"
)

const pushToRepositoryHook = "https://hooks.slack.com/services/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"
const packageVulnerabilityFoundHook = "https://hooks.slack.com/services/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"

type Creator struct {
	AliyunAccessKey    string
	AliyunAccessSecret string
	AliyunRegion       string
	QuayAccessToken    string
}

func (c Creator) CreateAliyunRepository(repoName string) error {
	crCli, err := cr.DefaultClient(c.AliyunAccessKey, c.AliyunAccessSecret, c.AliyunRegion)

	if err != nil {
		return err
	}

	repo, err := crCli.GetRepository(repoName)

	if err != nil {
		return err
	}

	if repo != nil {
		log.Printf("Repository already exists. Skipping\n")
	} else {
		repo, err := crCli.CreateRepository(repoName, true)
		if err != nil {
			return err
		}

		log.Println("Repository", repo.RepoNamespace+"/"+repo.RepoName, "created and configured on Aliyun")

	}

	return nil

}

func (c Creator) CreateQuayRepository(repoName string) error {

	quaycli := quay.DefaultClient()

	quaycli.AuthorizationHeader("Bearer " + c.QuayAccessToken)

	repo, err := quaycli.GetRepository(repoName)

	if err != nil {
		return err
	}

	if repo != nil {
		log.Printf("Repository already exists. Skipping\n")
	} else {

		repo, err = quaycli.CreateRepository(repoName, "public")

		if err != nil {
			return err
		}

		if repo == nil {
			return fmt.Errorf("Repository not created")
		}
	}
	notifications, err := quaycli.GetNotifications(repoName)
	if err != nil {
		return err
	}

	ptrhExists := false
	pvfhExists := false

	for _, n := range notifications {
		if n.Config.URL == pushToRepositoryHook {
			ptrhExists = true
			log.Println("Push To Repository Hook already exists. Skipping")
			continue
		}

		if n.Config.URL == packageVulnerabilityFoundHook {
			pvfhExists = true
			log.Println("Package Vulnerability Found Hook already exists. Skipping")
			continue
		}
	}

	if !ptrhExists {
		log.Println("Creating Push To Repository Hook")
		_, err := quaycli.CreateRepoPushNotification(repoName, "slack", pushToRepositoryHook, "Push To Repository Hook")
		if err != nil {
			log.Printf("Error creating Push To Repository Hook: %s\n", err)
		}
	}

	if !pvfhExists {
		log.Println("Creating Package Vulnerability Found Hook")
		_, err := quaycli.CreatePackageVulnerabilityFoundNotification(repoName, "slack", packageVulnerabilityFoundHook, "Package Vulnerability Found Hook", "4")
		if err != nil {
			log.Printf("Error creating Package Vulnerability Found: %s\n", err)
		}
	}

	log.Println("Repository", repoName, "created and configured on Quay")
	return nil
}
