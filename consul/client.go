package consul

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Service struct {
	ID             string            `json:"ID"`
	Node           string            `json:"Node"`
	Datacenter     string            `json:"Datacenter"`
	NodeMeta       map[string]string `json:"NodeMeta"`
	ServiceID      string            `json:"ServiceID"`
	ServiceName    string            `json:"ServiceName"`
	ServiceAddress string            `json:"ServiceAddress"`
	ServicePort    int               `json:"ServicePort"`
}

type ConsulClient struct {
	Address string
}

func NewConsulClient(address string) *ConsulClient {
	return &ConsulClient{Address: address}
}

func (c *ConsulClient) FetchServices(serviceName, filter string) ([]Service, error) {
	var services []Service

	url := fmt.Sprintf("%s/v1/catalog/service/%s?filter=%s", c.Address, serviceName, filter)
	body, err := c.fetchWithRetry(url, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch services: %v", err)
	}

	if err := json.Unmarshal(body, &services); err != nil {
		return nil, fmt.Errorf("error decoding services JSON: %v", err)
	}

	return services, nil
}

func (c *ConsulClient) fetchWithRetry(url string, maxRetries int) ([]byte, error) {
	var body []byte
	var err error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Attempt %d: error fetching data from Consul: %v", attempt, err)
			time.Sleep(time.Duration(attempt) * time.Second * 2)
			continue
		}
		body, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Printf("Attempt %d: error reading response body: %v", attempt, err)
			continue
		}
		return body, nil
	}
	return body, fmt.Errorf("failed to fetch data from Consul after %d attempts: %v", maxRetries, err)
}
