package main

import (
	"fmt"
	"go-nftables-manager/consul"
	"log"
)

const (
	consulAddress = "http://localhost:8500"
	serviceName   = "wireguard"
)

var (
	envValues   = []string{"metrics", "logs", "backups", "app"}
	stageValues = []string{"prod", "test"}
)

func main() {
	consulClient := consul.NewConsulClient(consulAddress)

	var services []consul.Service
	for _, env := range envValues {
		for _, stage := range stageValues {
			filter := fmt.Sprintf("NodeMeta.env==%s and NodeMeta.stage==%s", env, stage)
			fetchedServices, err := consulClient.FetchServices(serviceName, filter)
			if err != nil {
				log.Fatalf("Failed to fetch services: %v", err)
			}
			services = append(services, fetchedServices...)
		}
	}

	log.Println("Fetched services from consul successfully")
}
