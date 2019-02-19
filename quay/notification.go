package quay

import (
	"encoding/json"
	"fmt"
)

type NotificationConfig struct {
	URL string `json:"url"`
}
type Notification struct {
	EventConfig      map[string]string  `json:"event_config"`
	UUID             string             `json:"uuid,omitempty"`
	Title            interface{}        `json:"title"`
	NumberOfFailures int                `json:"number_of_failures,omitempty"`
	Method           string             `json:"method"`
	Config           NotificationConfig `json:"config"`
	Event            string             `json:"event"`
}

type NotificationInput struct {
	EventConfig map[string]string  `json:"eventConfig"`
	Title       interface{}        `json:"title"`
	Method      string             `json:"method"`
	Config      NotificationConfig `json:"config"`
	Event       string             `json:"event"`
}

type NotificationResponse struct {
	Notifications []Notification `json:"notifications"`
}

func getNotificationsPath(repositoryName string) (string, string) {
	return "GET", fmt.Sprintf("/api/v1/repository/%s/notification/", repositoryName)
}

func postNotificationPath(repositoryName string) (string, string) {
	return "POST", fmt.Sprintf("/api/v1/repository/%s/notification/", repositoryName)
}
func (c *Client) GetNotifications(repositoryName string) ([]Notification, error) {
	m, u := getNotificationsPath(repositoryName)
	statuscode, status, body, err := c.do(m, u, nil)
	if err != nil {
		return nil, err
	}

	if statuscode != 200 {
		if statuscode == 404 {
			return nil, nil
		}
		return nil, getAPIError(status, body)
	}

	var nr NotificationResponse
	err = json.Unmarshal(body, &nr)

	if err != nil {
		return nil, err
	}

	return nr.Notifications, nil

}

func (c *Client) CreateRepoPushNotification(repositoryName string, method string, hook string, title string) (*Notification, error) {
	n := NotificationInput{
		Event: "repo_push",
		Title: title,
		Config: NotificationConfig{
			URL: hook,
		},
		Method:      method,
		EventConfig: map[string]string{},
	}

	return c.CreateNotification(repositoryName, n)
}

func (c *Client) CreatePackageVulnerabilityFoundNotification(repositoryName string, method string, hook string, title string, level string) (*Notification, error) {
	n := NotificationInput{
		Event: "vulnerability_found",
		Title: title,
		Config: NotificationConfig{
			URL: hook,
		},
		Method: method,
		EventConfig: map[string]string{
			"level": level,
		},
	}

	return c.CreateNotification(repositoryName, n)
}

func (c *Client) CreateNotification(repositoryName string, n NotificationInput) (*Notification, error) {

	m, u := postNotificationPath(repositoryName)

	statuscode, status, body, err := c.do(m, u, mustReader(n))
	if err != nil {
		return nil, err
	}

	if statuscode != 201 {
		return nil, getAPIError(status, body)
	}

	notifications, err := c.GetNotifications(repositoryName)
	if err != nil {
		panic(err)
	}

	for _, no := range notifications {
		if no.Config.URL == n.Config.URL {
			return &no, nil
		}
	}

	return nil, fmt.Errorf("Notification not created")

}
